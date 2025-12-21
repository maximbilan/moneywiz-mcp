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

// Account represents a MoneyWiz account
type Account struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Balance      float64 `json:"balance"`
	Currency     string  `json:"currency"`
	AccountType  string  `json:"account_type"`
}

// GetAccounts retrieves all accounts from the database
func (db *DB) GetAccounts() ([]Account, error) {
	query := `
		SELECT Z_PK, ZNAME, ZBALANCE, ZCURRENCYNAME, ZTYPE
		FROM ZSYNCOBJECT
		WHERE Z_ENT = 16 AND ZNAME IS NOT NULL
		ORDER BY ZNAME
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		var accountType sql.NullString
		err := rows.Scan(&acc.ID, &acc.Name, &acc.Balance, &acc.Currency, &accountType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		if accountType.Valid {
			acc.AccountType = accountType.String
		}
		accounts = append(accounts, acc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return accounts, nil
}

// GetAccountBalance retrieves the balance for a specific account
func (db *DB) GetAccountBalance(accountID int64) (*Account, error) {
	query := `
		SELECT Z_PK, ZNAME, ZBALANCE, ZCURRENCYNAME, ZTYPE
		FROM ZSYNCOBJECT
		WHERE Z_ENT = 16 AND Z_PK = ?
	`

	var acc Account
	var accountType sql.NullString
	err := db.conn.QueryRow(query, accountID).Scan(&acc.ID, &acc.Name, &acc.Balance, &acc.Currency, &accountType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account with ID %d not found", accountID)
		}
		return nil, fmt.Errorf("failed to query account: %w", err)
	}

	if accountType.Valid {
		acc.AccountType = accountType.String
	}

	return &acc, nil
}

// Transaction represents a MoneyWiz transaction
type Transaction struct {
	ID          int64   `json:"id"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	AccountID   int64   `json:"account_id"`
}

// GetTransactions retrieves transactions for an account (or all transactions if accountID is 0)
func (db *DB) GetTransactions(accountID int64, limit int) ([]Transaction, error) {
	var query string
	var args []interface{}

	if accountID > 0 {
		query = `
			SELECT Z_PK, ZAMOUNT, ZDATE, ZDESC, ZACCOUNT
			FROM ZSYNCOBJECT
			WHERE Z_ENT = 34 AND ZAMOUNT IS NOT NULL AND ZACCOUNT = ?
			ORDER BY ZDATE DESC
			LIMIT ?
		`
		args = []interface{}{accountID, limit}
	} else {
		query = `
			SELECT Z_PK, ZAMOUNT, ZDATE, ZDESC, ZACCOUNT
			FROM ZSYNCOBJECT
			WHERE Z_ENT = 34 AND ZAMOUNT IS NOT NULL
			ORDER BY ZDATE DESC
			LIMIT ?
		`
		args = []interface{}{limit}
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var txn Transaction
		var date sql.NullString
		var desc sql.NullString
		err := rows.Scan(&txn.ID, &txn.Amount, &date, &desc, &txn.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		if date.Valid {
			txn.Date = date.String
		}
		if desc.Valid {
			txn.Description = desc.String
		}
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// Category represents a MoneyWiz category
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GetCategories retrieves all categories from the database
func (db *DB) GetCategories() ([]Category, error) {
	query := `
		SELECT Z_PK, ZNAME2
		FROM ZSYNCOBJECT
		WHERE Z_ENT = 19 AND ZNAME2 IS NOT NULL
		ORDER BY ZNAME2
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		err := rows.Scan(&cat.ID, &cat.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// Budget represents a MoneyWiz budget
type Budget struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GetBudgets retrieves all budgets from the database
func (db *DB) GetBudgets() ([]Budget, error) {
	query := `
		SELECT Z_PK, ZNAME
		FROM ZSYNCOBJECT
		WHERE Z_ENT = 18 AND ZNAME IS NOT NULL
		ORDER BY ZNAME
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query budgets: %w", err)
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var budget Budget
		err := rows.Scan(&budget.ID, &budget.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}
		budgets = append(budgets, budget)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating budgets: %w", err)
	}

	return budgets, nil
}
