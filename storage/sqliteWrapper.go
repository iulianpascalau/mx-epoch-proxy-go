package storage

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
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
		request_count INTEGER DEFAULT 0
	);`
	_, err := wrapper.db.Exec(usersTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

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
func (wrapper *sqliteWrapper) AddUser(username string, password string, isAdmin bool, maxRequests uint64) error {
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
	INSERT INTO users (username, hashed_password, is_admin, max_requests, request_count) 
	VALUES (?, ?, ?, ?, 0)
	`

	_, err = tx.Exec(query, username, hex.EncodeToString(hash), isAdmin, maxRequests)
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
func (wrapper *sqliteWrapper) UpdateUser(username string, password string, isAdmin bool, maxRequests uint64) error {
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

		query := `UPDATE users SET hashed_password = ?, is_admin = ?, max_requests = ? WHERE username = ?`
		_, err = tx.Exec(query, hex.EncodeToString(hash), isAdmin, maxRequests, username)
		if err != nil {
			return fmt.Errorf("failed to update user with password: %w", err)
		}
	} else {
		query := `UPDATE users SET is_admin = ?, max_requests = ? WHERE username = ?`
		_, err = tx.Exec(query, isAdmin, maxRequests, username)
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
func (wrapper *sqliteWrapper) IsKeyAllowed(key string) error {
	key, err := processKey(key)
	if err != nil {
		return err
	}

	tx, err := wrapper.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Get User limits via Key
	query := `
		SELECT u.max_requests, u.request_count, u.username 
		FROM users u
		JOIN access_keys k ON u.username = k.username
		WHERE k.key = ?
	`
	var maxRequests, requestCount uint64
	var username string

	err = tx.QueryRow(query, key).Scan(&maxRequests, &requestCount, &username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("the provided key is not allowed (no rows)")
		}

		return fmt.Errorf("error querying if is allowed: %w", err)
	}

	// Check limit
	if maxRequests > 0 && requestCount >= maxRequests {
		return fmt.Errorf("the provided key is not allowed, max_requests: %d, request_count: %d", maxRequests, requestCount)
	}

	// Increment counter on users
	query = `UPDATE users SET request_count = request_count + 1 WHERE username = ?`
	_, err = tx.Exec(query, username)
	if err != nil {
		return fmt.Errorf("error updating the request counter (update in users): %w", err)
	}

	// Increment counter on access_keys
	query = `UPDATE access_keys SET request_count = request_count + 1 WHERE key = ?`
	_, err = tx.Exec(query, key)
	if err != nil {
		return fmt.Errorf("error updating the request counter (update in access_keys): %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error updating the request counter (commit): %w", err)
	}

	return nil
}

// CheckUserCredentials checks if the user with the given username and password exists and returns details
func (wrapper *sqliteWrapper) CheckUserCredentials(username string, password string) (*common.UsersDetails, error) {
	query := `SELECT max_requests, request_count, username, hashed_password, is_admin FROM users WHERE username = ?`
	var details common.UsersDetails
	err := wrapper.db.QueryRow(query, username).Scan(&details.MaxRequests, &details.GlobalCounter, &details.Username, &details.HashedPassword, &details.IsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error querying user: %w", err)
	}

	err = checkPassword(password, details.HashedPassword)
	if err != nil {
		return nil, err
	}

	return &details, nil
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
		SELECT max_requests, request_count, username, hashed_password, is_admin 
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
		err = rows.Scan(&details.MaxRequests, &details.GlobalCounter, &details.Username, &details.HashedPassword, &details.IsAdmin)
		if err != nil {
			return nil, err
		}
		result[strings.ToLower(details.Username)] = details
	}
	return result, rows.Err()
}

// Close closes the database connection
func (wrapper *sqliteWrapper) Close() error {
	return wrapper.db.Close()
}

// IsInterfaceNil returns true if the value under the interface is nil
func (wrapper *sqliteWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
