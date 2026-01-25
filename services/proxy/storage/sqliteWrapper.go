package storage

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	_ "github.com/mattn/go-sqlite3"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"golang.org/x/crypto/bcrypt"
)

const maxPassLen = 72

var log = logger.GetOrCreate("storage")

// sqliteWrapper handles the connection to the SQLite database
type sqliteWrapper struct {
	db                     *sql.DB
	pendingWritesWaitGroup *sync.WaitGroup
	counters               CountersCache
}

// NewSQLiteWrapper creates a new instance of SQLiteWrapper
func NewSQLiteWrapper(dbPath string, counters CountersCache) (*sqliteWrapper, error) {
	if check.IfNil(counters) {
		return nil, errNilCountersCache
	}

	err := prepareDirectories(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create initial empty DB file: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	// Enable WAL mode for better performance
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	wrapper := &sqliteWrapper{
		db:                     db,
		counters:               counters,
		pendingWritesWaitGroup: &sync.WaitGroup{},
	}
	err = wrapper.initializeTables()
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return wrapper, nil
}

func prepareDirectories(dbPath string) error {
	return os.MkdirAll(filepath.Dir(dbPath), os.ModePerm)
}

func (wrapper *sqliteWrapper) initializeTables() error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		hashed_password TEXT,
		is_admin BOOLEAN DEFAULT FALSE,
		max_requests INTEGER DEFAULT 0,
		request_count INTEGER DEFAULT 0,

		is_premium BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT TRUE,
		activation_token TEXT DEFAULT '',
		pending_email TEXT DEFAULT '',
		change_email_token TEXT DEFAULT '',
		crypto_payment_id INTEGER DEFAULT NULL
	);`
	_, err := wrapper.db.Exec(usersTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Migration: Attempt to add columns for existing databases
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN is_premium BOOLEAN DEFAULT FALSE;")
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT TRUE;")
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN activation_token TEXT DEFAULT '';")
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN pending_email TEXT DEFAULT '';")
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN change_email_token TEXT DEFAULT '';")
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN crypto_payment_id INTEGER DEFAULT NULL;")

	keysTable := `
	CREATE TABLE IF NOT EXISTS access_keys (
		key TEXT PRIMARY KEY,
		username TEXT,
		request_count INTEGER DEFAULT 0,
		FOREIGN KEY(username) REFERENCES users(username)
	);`
	_, err = wrapper.db.Exec(keysTable)
	if err != nil {
		return fmt.Errorf("failed to create access_keys table: %w", err)
	}

	indexQuery := `CREATE INDEX IF NOT EXISTS idx_access_keys_username ON access_keys(username);`
	_, err = wrapper.db.Exec(indexQuery)
	if err != nil {
		return fmt.Errorf("failed to create index on access_keys: %w", err)
	}

	indexTokenQuery := `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_activation_token ON users(activation_token) WHERE activation_token != '';`
	_, err = wrapper.db.Exec(indexTokenQuery)
	if err != nil {
		return fmt.Errorf("failed to create index on users activation_token: %w", err)
	}

	indexEmailChangeTokenQuery := `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_change_email_token ON users(change_email_token) WHERE change_email_token != '';`
	_, err = wrapper.db.Exec(indexEmailChangeTokenQuery)
	if err != nil {
		return fmt.Errorf("failed to create index on users change_email_token: %w", err)
	}

	performanceTable := `
	CREATE TABLE IF NOT EXISTS performance (
		label TEXT PRIMARY KEY,
		counter INTEGER DEFAULT 0
	);`
	_, err = wrapper.db.Exec(performanceTable)
	if err != nil {
		return fmt.Errorf("failed to create performance table: %w", err)
	}

	return nil
}

func processKey(key string) (string, error) {
	key = strings.ToLower(key)
	key = strings.Trim(key, " \t\r\n")
	if len(key) == 0 {
		return "", errKeyIsEmpty
	}

	return key, nil
}

// AddUser creates the associated user
func (wrapper *sqliteWrapper) AddUser(username string, password string, isAdmin bool, maxRequests uint64, isPremium bool, isActive bool, activationToken string) error {
	if len(password) > maxPassLen {
		return fmt.Errorf("password is too long (maximum %d characters allowed)", maxPassLen)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Upsert User
	query := `
	INSERT INTO users (username, hashed_password, is_admin, max_requests, request_count, is_premium, is_active, activation_token) 
	VALUES (?, ?, ?, ?, 0, ?, ?, ?)
	`

	_, err = tx.Exec(query, username, hex.EncodeToString(hash), isAdmin, maxRequests, isPremium, isActive, activationToken)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return tx.Commit()
}

// RemoveUser removes the provided user and all associated keys
func (wrapper *sqliteWrapper) RemoveUser(username string) error {
	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete associated keys first
	queryDeleteKeys := `DELETE FROM access_keys WHERE username = ?`
	_, err = tx.Exec(queryDeleteKeys, username)
	if err != nil {
		return fmt.Errorf("failed to remove associated keys: %w", err)
	}

	// Delete user
	queryDeleteUser := `DELETE FROM users WHERE username = ?`
	_, err = tx.Exec(queryDeleteUser, username)
	if err != nil {
		return fmt.Errorf("failed to remove user: %w", err)
	}

	wrapper.counters.Remove(username)

	return tx.Commit()
}

// UpdateUser updates the user's details
func (wrapper *sqliteWrapper) UpdateUser(username string, password string, isAdmin bool, maxRequests uint64, isPremium bool) error {
	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if password != "" {
		if len(password) > maxPassLen {
			return fmt.Errorf("password is too long (maximum %d characters allowed)", maxPassLen)
		}
		hash, errGenerate := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if errGenerate != nil {
			return errGenerate
		}

		query := `UPDATE users SET hashed_password = ?, is_admin = ?, max_requests = ?, is_premium = ? WHERE username = ?`
		_, err = tx.Exec(query, hex.EncodeToString(hash), isAdmin, maxRequests, isPremium, username)
		if err != nil {
			return fmt.Errorf("failed to update user with password: %w", err)
		}
	} else {
		query := `UPDATE users SET is_admin = ?, max_requests = ?, is_premium = ? WHERE username = ?`
		_, err = tx.Exec(query, isAdmin, maxRequests, isPremium, username)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	}

	return tx.Commit()
}

// AddKey adds a new access key without checking user's credentials (trusted caller)
func (wrapper *sqliteWrapper) AddKey(username string, key string) error {
	key, err := processKey(key)
	if err != nil {
		return err
	}

	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `INSERT INTO access_keys (key, username) VALUES (?, ?)`
	_, err = tx.Exec(query, key, username)
	if err != nil {
		return fmt.Errorf("failed to insert key: %w", err)
	}

	return tx.Commit()
}

// RemoveKey removes the provided access key without checking user's credentials (trusted caller)
func (wrapper *sqliteWrapper) RemoveKey(username string, key string) error {
	key, err := processKey(key)
	if err != nil {
		return err
	}

	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `DELETE FROM access_keys WHERE key = ? and username = ?`
	_, err = tx.Exec(query, strings.ToLower(key), username)
	if err != nil {
		return fmt.Errorf("failed to remove key: %w", err)
	}

	return tx.Commit()
}

// IsKeyAllowed returns true if the key is allowed to do requests and false otherwise
func (wrapper *sqliteWrapper) IsKeyAllowed(key string) (string, common.AccountType, error) {
	key, err := processKey(key)
	if err != nil {
		return "", "", err
	}

	// Get User limits via Key
	query := `
		SELECT u.max_requests, u.request_count, u.username, u.is_premium
		FROM users u
		JOIN access_keys k ON u.username = k.username
		WHERE k.key = ?
	`
	var maxRequests, requestCount uint64
	var username string
	var isPremium bool

	err = wrapper.db.QueryRow(query, key).Scan(&maxRequests, &requestCount, &username, &isPremium)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("the provided key is not allowed (no rows)")
		}

		return "", "", fmt.Errorf("error querying if is allowed: %w", err)
	}

	userCounter := wrapper.counters.Get(username)

	// Determine account type return
	userDetails := &common.UsersDetails{
		IsPremium:     isPremium,
		MaxRequests:   maxRequests,
		GlobalCounter: max(userCounter, requestCount),
	}
	common.ProcessUserDetails(userDetails)
	wrapper.counters.Set(username, userDetails.GlobalCounter+1)

	wrapper.pendingWritesWaitGroup.Add(2)
	go wrapper.incrementCountersOnUsers(username)
	go wrapper.incrementCountersOnKeys(key)

	return username, userDetails.ProcessedAccountType, nil
}

func (wrapper *sqliteWrapper) incrementCountersOnUsers(username string) {
	tx, err := wrapper.db.Begin()
	if err != nil {
		log.Error("error creating transaction (update in users)", "username", username, "error", err)
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `UPDATE users SET request_count = request_count + 1 WHERE username = ?`
	_, err = tx.Exec(query, username)
	if err != nil {
		log.Error("error updating the request counter (update in users)", "username", username, "error", err)
	}

	query = `UPDATE users SET max_requests = request_count WHERE username = ? and max_requests < request_count`
	_, err = tx.Exec(query, username)
	if err != nil {
		log.Error("error updating the request counter (update in users)", "username", username, "error", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Error("error commiting transaction (update in users)", "username", username, "error", err)
	}

	wrapper.pendingWritesWaitGroup.Done()
}

func (wrapper *sqliteWrapper) incrementCountersOnKeys(key string) {
	query := `UPDATE access_keys SET request_count = request_count + 1 WHERE key = ?`
	_, err := wrapper.db.Exec(query, key)
	if err != nil {
		log.Error("error updating the request counter (update in keys)", "key", common.AnonymizeKey(key), "error", err)
	}

	wrapper.pendingWritesWaitGroup.Done()
}

// CheckUserCredentials checks if the user with the given username and password exists and returns details
func (wrapper *sqliteWrapper) CheckUserCredentials(username string, password string) (*common.UsersDetails, error) {
	details, err := wrapper.getUserDetails(username)
	if err != nil {
		return nil, err
	}

	err = checkPassword(password, details.HashedPassword)
	if err != nil {
		return nil, err
	}

	return details, nil
}

func checkPassword(passwordPlain string, hexHashedPass string) error {
	hashedPassBytes, err := hex.DecodeString(hexHashedPass)
	if err != nil {
		return fmt.Errorf("saved password is invalid")
	}

	err = bcrypt.CompareHashAndPassword(hashedPassBytes, []byte(passwordPlain))
	if err != nil {
		return fmt.Errorf("invalid password:  %w", err)
	}

	return nil
}

// GetUser returns the user details for the given username
func (wrapper *sqliteWrapper) GetUser(username string) (*common.UsersDetails, error) {
	return wrapper.getUserDetails(username)
}

func (wrapper *sqliteWrapper) getUserDetails(username string) (*common.UsersDetails, error) {
	query := `SELECT max_requests, request_count, username, hashed_password, is_admin, is_premium, is_active, crypto_payment_id FROM users WHERE username = ?`
	var details common.UsersDetails
	var cryptoPaymentID sql.NullInt64
	err := wrapper.db.QueryRow(query, username).Scan(&details.MaxRequests, &details.GlobalCounter, &details.Username, &details.HashedPassword, &details.IsAdmin, &details.IsPremium, &details.IsActive, &cryptoPaymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error querying user: %w", err)
	}

	if cryptoPaymentID.Valid {
		details.CryptoPaymentID = uint64(cryptoPaymentID.Int64)
	}
	details.GlobalCounter = max(details.GlobalCounter, wrapper.counters.Get(details.Username))
	common.ProcessUserDetails(&details)

	return &details, nil
}

// GetAllKeys returns all access keys and their details
func (wrapper *sqliteWrapper) GetAllKeys(username string) (map[string]common.AccessKeyDetails, error) {
	var rows *sql.Rows
	var err error
	if username == "" {
		query := `
		SELECT k.key, u.max_requests, u.request_count AS global_counter, k.request_count as key_counter, u.username, u.hashed_password, u.is_admin 
		FROM access_keys k
		JOIN users u ON k.username = u.username
	`
		rows, err = wrapper.db.Query(query)
	} else {
		query := `
		SELECT k.key, u.max_requests, u.request_count AS global_counter, k.request_count as key_counter, u.username, u.hashed_password, u.is_admin 
		FROM access_keys k
		JOIN users u ON k.username = u.username
		WHERE u.username = ?
	`
		rows, err = wrapper.db.Query(query, username)
	}

	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	result := make(map[string]common.AccessKeyDetails)
	for rows.Next() {
		var key string
		var details common.AccessKeyDetails
		err = rows.Scan(&key, &details.MaxRequests, &details.GlobalCounter, &details.KeyCounter, &details.Username, &details.HashedPassword, &details.IsAdmin)
		if err != nil {
			return nil, err
		}

		details.GlobalCounter = max(details.GlobalCounter, wrapper.counters.Get(details.Username))

		result[strings.ToLower(key)] = details
	}
	return result, rows.Err()
}

// GetAllUsers returns all access keys and their details
func (wrapper *sqliteWrapper) GetAllUsers() (map[string]common.UsersDetails, error) {
	query := `
		SELECT max_requests, request_count, username, hashed_password, is_admin, is_premium, is_active, crypto_payment_id
		FROM users
	`
	rows, err := wrapper.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	result := make(map[string]common.UsersDetails)
	for rows.Next() {
		var details common.UsersDetails
		var cryptoPaymentID sql.NullInt64
		err = rows.Scan(&details.MaxRequests, &details.GlobalCounter, &details.Username, &details.HashedPassword, &details.IsAdmin, &details.IsPremium, &details.IsActive, &cryptoPaymentID)
		if err != nil {
			return nil, err
		}
		if cryptoPaymentID.Valid {
			details.CryptoPaymentID = uint64(cryptoPaymentID.Int64)
		}

		details.GlobalCounter = max(details.GlobalCounter, wrapper.counters.Get(details.Username))
		common.ProcessUserDetails(&details)

		result[strings.ToLower(details.Username)] = details
	}
	return result, rows.Err()
}

// ActivateUser activates the user with the given token
func (wrapper *sqliteWrapper) ActivateUser(token string) error {
	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := "UPDATE users SET is_active = TRUE, activation_token = '' WHERE activation_token = ? AND activation_token != ''"
	result, err := tx.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invalid or expired activation token")
	}

	return tx.Commit()
}

// AddPerformanceMetricAsync increments the counter for the given label in an async manner
func (wrapper *sqliteWrapper) AddPerformanceMetricAsync(label string) {
	wrapper.pendingWritesWaitGroup.Add(1)

	go func() {
		defer wrapper.pendingWritesWaitGroup.Done()

		tx, err := wrapper.db.Begin()
		if err != nil {
			log.Error("failed to begin transaction for performance metric", "error", err)
			return
		}
		defer func() {
			_ = tx.Rollback()
		}()

		// Try update
		query := "UPDATE performance SET counter = counter + 1 WHERE label = ?"
		res, err := tx.Exec(query, label)
		if err != nil {
			log.Error("failed to update performance metric", "error", err)
			return
		}
		rows, err := res.RowsAffected()
		if err != nil {
			log.Error("failed to get rows affected for performance metric", "error", err)
			return
		}

		if rows == 0 {
			// Insert
			query = "INSERT INTO performance (label, counter) VALUES (?, 1)"
			_, err = tx.Exec(query, label)
			if err != nil {
				log.Error("failed to insert performance metric", "error", err)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Error("failed to commit performance metric transaction", "error", err)
		}
	}()
}

// GetPerformanceMetrics returns the performance metrics
func (wrapper *sqliteWrapper) GetPerformanceMetrics() (map[string]uint64, error) {
	query := "SELECT label, counter FROM performance"
	rows, err := wrapper.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query performance metrics: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	metrics := make(map[string]uint64)
	for rows.Next() {
		var label string
		var counter uint64
		err = rows.Scan(&label, &counter)
		if err != nil {
			return nil, fmt.Errorf("failed to scan performance metric: %w", err)
		}
		metrics[label] = counter
	}

	return metrics, nil
}

// UpdatePassword updates the user's password
func (wrapper *sqliteWrapper) UpdatePassword(username string, password string) error {
	if len(password) > maxPassLen {
		return fmt.Errorf("password is too long (maximum %d characters allowed)", maxPassLen)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `UPDATE users SET hashed_password = ? WHERE username = ?`
	result, err := tx.Exec(query, hex.EncodeToString(hash), username)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// RequestEmailChange initiates the email change process
func (wrapper *sqliteWrapper) RequestEmailChange(username string, newEmail string, token string) error {

	// Check if new email is already taken
	var exists bool
	queryCheck := "SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)"
	err := wrapper.db.QueryRow(queryCheck, newEmail).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("email already registered")
	}

	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `UPDATE users SET pending_email = ?, change_email_token = ? WHERE username = ?`
	result, err := tx.Exec(query, newEmail, token, username)
	if err != nil {
		return fmt.Errorf("failed to request email change: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// ConfirmEmailChange finalizes the email change process
func (wrapper *sqliteWrapper) ConfirmEmailChange(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("invalid token")
	}

	tx, err := wrapper.db.Begin()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. Find user with this token
	querySelect := `SELECT username, pending_email, hashed_password, is_admin, max_requests, request_count, is_premium, is_active, crypto_payment_id FROM users WHERE change_email_token = ?`
	var oldUsername, newEmail, hashedPassword string
	var isAdmin, isActive, isPremium bool
	var maxRequests, requestCount uint64
	var cryptoPaymentID sql.NullInt64

	err = tx.QueryRow(querySelect, token).Scan(&oldUsername, &newEmail, &hashedPassword, &isAdmin, &maxRequests, &requestCount, &isPremium, &isActive, &cryptoPaymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("invalid or expired token")
		}
		return "", fmt.Errorf("failed to find user by token: %w", err)
	}

	if newEmail == "" {
		return "", fmt.Errorf("no pending email found")
	}

	// 2. Create new user with new email (username)
	// Check if new email free again (double check)
	var exists bool
	queryCheck := "SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)"
	err = tx.QueryRow(queryCheck, newEmail).Scan(&exists)
	if err != nil {
		return "", err
	}
	if exists {
		return "", fmt.Errorf("email %s is already in use", newEmail)
	}

	insertQueryFull := `
	INSERT INTO users (username, hashed_password, is_admin, max_requests, request_count, is_premium, is_active, activation_token, pending_email, change_email_token, crypto_payment_id) 
	VALUES (?, ?, ?, ?, ?, ?, ?, '', '', '', ?)
	`
	_, err = tx.Exec(insertQueryFull, newEmail, hashedPassword, isAdmin, maxRequests, requestCount, isPremium, isActive, cryptoPaymentID)
	if err != nil {
		return "", fmt.Errorf("failed to create new user entry: %w", err)
	}

	// 3. Update Access Keys to point to new user
	updateKeysQuery := `UPDATE access_keys SET username = ? WHERE username = ?`
	_, err = tx.Exec(updateKeysQuery, newEmail, oldUsername)
	if err != nil {
		return "", fmt.Errorf("failed to migrat access keys: %w", err)
	}

	// 4. Delete old user
	deleteUserQuery := `DELETE FROM users WHERE username = ?`
	_, err = tx.Exec(deleteUserQuery, oldUsername)
	if err != nil {
		return "", fmt.Errorf("failed to delete old user: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return newEmail, nil
}

// SetCryptoPaymentID updates the user's crypto payment ID
func (wrapper *sqliteWrapper) SetCryptoPaymentID(username string, paymentID uint64) error {
	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `UPDATE users SET crypto_payment_id = ? WHERE username = ?`
	result, err := tx.Exec(query, paymentID, username)
	if err != nil {
		return fmt.Errorf("failed to set crypto payment ID: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// Close closes the database connection
func (wrapper *sqliteWrapper) Close() error {
	// allow all pending updates to finish before closing the db connection
	wrapper.pendingWritesWaitGroup.Wait()

	return wrapper.db.Close()
}

// UpdateMaxRequests updates the user's max requests
func (wrapper *sqliteWrapper) UpdateMaxRequests(username string, maxRequests uint64) error {
	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `UPDATE users SET max_requests = ? WHERE username = ?`
	result, err := tx.Exec(query, maxRequests, username)
	if err != nil {
		return fmt.Errorf("failed to update max requests: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// IsInterfaceNil returns true if the value under the interface is nil
func (wrapper *sqliteWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
