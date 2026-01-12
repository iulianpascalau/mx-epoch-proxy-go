package storage

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const maxPassLen = 72

// sqliteWrapper handles the connection to the SQLite database
type sqliteWrapper struct {
	db *sql.DB
}

// NewSQLiteWrapper creates a new instance of SQLiteWrapper
func NewSQLiteWrapper(dbPath string) (*sqliteWrapper, error) {
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

	wrapper := &sqliteWrapper{db: db}
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
		account_type TEXT DEFAULT 'free',
		is_active BOOLEAN DEFAULT TRUE,
		activation_token TEXT DEFAULT '',
		pending_email TEXT DEFAULT '',
		change_email_token TEXT DEFAULT ''
	);`
	_, err := wrapper.db.Exec(usersTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Migration: Attempt to add columns for existing databases
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN account_type TEXT DEFAULT 'free';")
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT TRUE;")
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN activation_token TEXT DEFAULT '';")
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN pending_email TEXT DEFAULT '';")
	_, _ = wrapper.db.Exec("ALTER TABLE users ADD COLUMN change_email_token TEXT DEFAULT '';")

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
func (wrapper *sqliteWrapper) AddUser(username string, password string, isAdmin bool, maxRequests uint64, accountTypeStr string, isActive bool, activationToken string) error {
	if len(password) > maxPassLen {
		return fmt.Errorf("password is too long (maximum %d characters allowed)", maxPassLen)
	}

	accountType := formatAccountType(accountTypeStr)

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
	INSERT INTO users (username, hashed_password, is_admin, max_requests, request_count, account_type, is_active, activation_token) 
	VALUES (?, ?, ?, ?, 0, ?, ?, ?)
	`

	_, err = tx.Exec(query, username, hex.EncodeToString(hash), isAdmin, maxRequests, accountType, isActive, activationToken)
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

	return tx.Commit()
}

// UpdateUser updates the user's details
func (wrapper *sqliteWrapper) UpdateUser(username string, password string, isAdmin bool, maxRequests uint64, accountTypeStr string) error {
	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	accountType := formatAccountType(accountTypeStr)

	if password != "" {
		if len(password) > maxPassLen {
			return fmt.Errorf("password is too long (maximum %d characters allowed)", maxPassLen)
		}
		hash, errGenerate := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if errGenerate != nil {
			return errGenerate
		}

		query := `UPDATE users SET hashed_password = ?, is_admin = ?, max_requests = ?, account_type = ? WHERE username = ?`
		_, err = tx.Exec(query, hex.EncodeToString(hash), isAdmin, maxRequests, accountType, username)
		if err != nil {
			return fmt.Errorf("failed to update user with password: %w", err)
		}
	} else {
		query := `UPDATE users SET is_admin = ?, max_requests = ?, account_type = ? WHERE username = ?`
		_, err = tx.Exec(query, isAdmin, maxRequests, accountType, username)
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

	tx, err := wrapper.db.Begin()
	if err != nil {
		return "", "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Get User limits via Key
	query := `
		SELECT u.max_requests, u.request_count, u.username, u.account_type
		FROM users u
		JOIN access_keys k ON u.username = k.username
		WHERE k.key = ?
	`
	var maxRequests, requestCount uint64
	var username string
	var accountType string

	err = tx.QueryRow(query, key).Scan(&maxRequests, &requestCount, &username, &accountType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("the provided key is not allowed (no rows)")
		}

		return "", "", fmt.Errorf("error querying if is allowed: %w", err)
	}

	// Check limit
	if maxRequests > 0 && requestCount >= maxRequests {
		return "", "", fmt.Errorf("the provided key is not allowed, max_requests: %d, request_count: %d", maxRequests, requestCount)
	}

	// Increment counter on users
	query = `UPDATE users SET request_count = request_count + 1 WHERE username = ?`
	_, err = tx.Exec(query, username)
	if err != nil {
		return "", "", fmt.Errorf("error updating the request counter (update in users): %w", err)
	}

	// Increment counter on access_keys
	query = `UPDATE access_keys SET request_count = request_count + 1 WHERE key = ?`
	_, err = tx.Exec(query, key)
	if err != nil {
		return "", "", fmt.Errorf("error updating the request counter (update in access_keys): %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", "", fmt.Errorf("error updating the request counter (commit): %w", err)
	}

	return username, formatAccountType(accountType), nil
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
	query := `SELECT max_requests, request_count, username, hashed_password, is_admin, account_type, is_active FROM users WHERE username = ?`
	var details common.UsersDetails
	err := wrapper.db.QueryRow(query, username).Scan(&details.MaxRequests, &details.GlobalCounter, &details.Username, &details.HashedPassword, &details.IsAdmin, &details.AccountType, &details.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error querying user: %w", err)
	}

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
		result[strings.ToLower(key)] = details
	}
	return result, rows.Err()
}

// GetAllUsers returns all access keys and their details
func (wrapper *sqliteWrapper) GetAllUsers() (map[string]common.UsersDetails, error) {
	query := `
		SELECT max_requests, request_count, username, hashed_password, is_admin, account_type, is_active
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
		err = rows.Scan(&details.MaxRequests, &details.GlobalCounter, &details.Username, &details.HashedPassword, &details.IsAdmin, &details.AccountType, &details.IsActive)
		if err != nil {
			return nil, err
		}
		result[strings.ToLower(details.Username)] = details
	}
	return result, rows.Err()
}

func formatAccountType(accountTypeStr string) common.AccountType {
	accountTypeStr = strings.ToLower(accountTypeStr)
	if accountTypeStr == string(common.PremiumAccountType) {
		return common.PremiumAccountType
	}

	return common.FreeAccountType
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

// AddPerformanceMetric increments the counter for the given label
func (wrapper *sqliteWrapper) AddPerformanceMetric(label string) error {
	tx, err := wrapper.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Try update
	query := "UPDATE performance SET counter = counter + 1 WHERE label = ?"
	res, err := tx.Exec(query, label)
	if err != nil {
		return fmt.Errorf("failed to update performance metric: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		// Insert
		query = "INSERT INTO performance (label, counter) VALUES (?, 1)"
		_, err = tx.Exec(query, label)
		if err != nil {
			return fmt.Errorf("failed to insert performance metric: %w", err)
		}
	}

	return tx.Commit()
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

// Close closes the database connection
func (wrapper *sqliteWrapper) Close() error {
	return wrapper.db.Close()
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
	querySelect := `SELECT username, pending_email, hashed_password, is_admin, max_requests, request_count, account_type, is_active FROM users WHERE change_email_token = ?`
	var oldUsername, newEmail, hashedPassword, accountType string
	var isAdmin, isActive bool
	var maxRequests, requestCount uint64

	err = tx.QueryRow(querySelect, token).Scan(&oldUsername, &newEmail, &hashedPassword, &isAdmin, &maxRequests, &requestCount, &accountType, &isActive)
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
	INSERT INTO users (username, hashed_password, is_admin, max_requests, request_count, account_type, is_active, activation_token, pending_email, change_email_token) 
	VALUES (?, ?, ?, ?, ?, ?, ?, '', '', '')
	`
	_, err = tx.Exec(insertQueryFull, newEmail, hashedPassword, isAdmin, maxRequests, requestCount, accountType, isActive)
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

// IsInterfaceNil returns true if the value under the interface is nil
func (wrapper *sqliteWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
