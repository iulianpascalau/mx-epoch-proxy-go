package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// BalanceEntry represents a row in the balance-management table
type BalanceEntry struct {
	ID             int
	Address        string
	LastBalance    float64
	CurrentBalance float64
	TotalRequests  int
}

// sqliteWrapper handles the connection to the SQLite database
type sqliteWrapper struct {
	db             *sql.DB
	addressHandler AddressHandler
}

// NewSQLiteWrapper creates a new instance of SQLiteWrapper
func NewSQLiteWrapper(dbPath string, addressHandler AddressHandler) (*sqliteWrapper, error) {
	if check.IfNil(addressHandler) {
		return nil, errNilAddressHandler
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
	_, _ = db.Exec("PRAGMA journal_mode=WAL;")
	_, _ = db.Exec("PRAGMA synchronous=NORMAL;")

	wrapper := &sqliteWrapper{
		db:             db,
		addressHandler: addressHandler,
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
	query := `
	CREATE TABLE IF NOT EXISTS balance_management (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		address TEXT,
		last_balance REAL DEFAULT 0,
		current_balance REAL DEFAULT 0,
		total_requests INTEGER DEFAULT 0
	);`
	_, err := wrapper.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create balance_management table: %w", err)
	}
	return nil
}

// Get returns the row based on the ID
func (wrapper *sqliteWrapper) Get(id int) (*BalanceEntry, error) {
	query := `SELECT id, address, last_balance, current_balance, total_requests FROM balance_management WHERE id = ?`
	row := wrapper.db.QueryRow(query, id)

	var entry BalanceEntry
	err := row.Scan(&entry.ID, &entry.Address, &entry.LastBalance, &entry.CurrentBalance, &entry.TotalRequests)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entry with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	return &entry, nil
}

// Add creates a new entry and returns the created id and the address string
func (wrapper *sqliteWrapper) Add() (int, string, error) {
	tx, err := wrapper.db.Begin()
	if err != nil {
		return 0, "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `INSERT INTO balance_management (address, last_balance, current_balance, total_requests) VALUES ("", 0, 0, 0)`
	result, err := tx.Exec(query)
	if err != nil {
		return 0, "", fmt.Errorf("failed to add entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, "", fmt.Errorf("failed to get last insert id: %w", err)
	}

	address, err := wrapper.addressHandler.GetAddressAtIndex(uint32(id))
	if err != nil {
		return 0, "", fmt.Errorf("failed to generate address: %w", err)
	}

	updateQuery := `UPDATE balance_management SET address = ? WHERE id = ?`
	_, err = tx.Exec(updateQuery, address, id)
	if err != nil {
		return 0, "", fmt.Errorf("failed to update address: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return int(id), address, nil
}

// GetAll provides all rows
func (wrapper *sqliteWrapper) GetAll() ([]*BalanceEntry, error) {
	query := `SELECT id, address, last_balance, current_balance, total_requests FROM balance_management`
	rows, err := wrapper.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all entries: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var entries []*BalanceEntry
	for rows.Next() {
		var entry BalanceEntry
		err = rows.Scan(&entry.ID, &entry.Address, &entry.LastBalance, &entry.CurrentBalance, &entry.TotalRequests)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

// Close closes the database connection
func (wrapper *sqliteWrapper) Close() error {
	return wrapper.db.Close()
}

// IsInterfaceNil returns true if the value under the interface is nil
func (wrapper *sqliteWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
