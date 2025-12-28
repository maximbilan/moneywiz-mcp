package database

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection
func NewDB(dbPath string) (*DB, error) {
	// Resolve the database path
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}

	conn, err := sql.Open("sqlite3", absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}
