package database

import (
	"database/sql"
	"fmt"
)

// Account represents a MoneyWiz account
type Account struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Balance     float64 `json:"balance"`
	Currency    string  `json:"currency"`
	AccountType string  `json:"account_type"`
}

// GetAccounts retrieves all accounts from the database
// Accounts can be stored in multiple entity types:
// - Entity 10: Regular bank accounts
// - Entity 11: Deposit accounts
// - Entity 12: Cash accounts
// - Entity 13: Other account types
// - Entity 15: Investment accounts
// - Entity 16: Regular accounts
// Note: Balance is stored in ZBALLANCE (double L), not ZBALANCE
// If balance is 0 or NULL, we calculate it from transactions + opening balance
func (db *DB) GetAccounts() ([]Account, error) {
	query := `
		SELECT Z_PK, ZNAME, ZBALLANCE, ZOPENINGBALANCE, ZCURRENCYNAME, ZTYPE
		FROM ZSYNCOBJECT
		WHERE Z_ENT IN (10, 11, 12, 13, 15, 16) AND ZNAME IS NOT NULL
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
		var name sql.NullString
		var accountType sql.NullString
		var balance sql.NullFloat64
		var openingBalance sql.NullFloat64
		var currency sql.NullString
		err := rows.Scan(&acc.ID, &name, &balance, &openingBalance, &currency, &accountType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		if name.Valid {
			acc.Name = name.String
		}

		// Calculate balance from opening balance + transactions (exactly as Python implementation)
		// Python code: current_balance = opening_balance + transaction_total
		calculatedBalance, err := db.calculateAccountBalance(acc.ID, openingBalance)
		if err == nil {
			acc.Balance = calculatedBalance
		} else {
			// Fallback to opening balance or stored balance
			if openingBalance.Valid {
				acc.Balance = openingBalance.Float64
			} else if balance.Valid {
				acc.Balance = balance.Float64
			} else {
				acc.Balance = 0.0
			}
		}

		if currency.Valid {
			acc.Currency = currency.String
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

// calculateAccountBalance calculates the account balance from opening balance + transactions
// Transactions are entity types 37, 45, 46, 47 (regular transactions) and 43 (transfers)
// They link to accounts via ZACCOUNT2 (and ZACCOUNT for transfers) and use ZAMOUNT1 for the amount
func (db *DB) calculateAccountBalance(accountID int64, openingBalance sql.NullFloat64) (float64, error) {
	var opening float64
	if openingBalance.Valid {
		opening = openingBalance.Float64
	}

	// Include entity 43 (transfers) and check both ZACCOUNT2 and ZACCOUNT
	query := `
		SELECT COALESCE(SUM(ZAMOUNT1), 0)
		FROM ZSYNCOBJECT
		WHERE Z_ENT IN (37, 45, 46, 47, 43) 
		AND (ZACCOUNT2 = ? OR ZACCOUNT = ?)
		AND ZAMOUNT1 IS NOT NULL
	`

	var transactionSum sql.NullFloat64
	err := db.conn.QueryRow(query, accountID, accountID).Scan(&transactionSum)
	if err != nil {
		return opening, err
	}

	var sum float64
	if transactionSum.Valid {
		sum = transactionSum.Float64
	}

	return opening + sum, nil
}

// GetAccountBalance retrieves the balance for a specific account
// Note: Balance is stored in ZBALLANCE (double L), not ZBALANCE
// If balance is 0 or NULL, we calculate it from transactions + opening balance
func (db *DB) GetAccountBalance(accountID int64) (*Account, error) {
	query := `
		SELECT Z_PK, ZNAME, ZBALLANCE, ZOPENINGBALANCE, ZCURRENCYNAME, ZTYPE
		FROM ZSYNCOBJECT
		WHERE Z_ENT IN (10, 11, 12, 13, 15, 16) AND Z_PK = ?
	`

	var acc Account
	var name sql.NullString
	var accountType sql.NullString
	var balance sql.NullFloat64
	var openingBalance sql.NullFloat64
	var currency sql.NullString
	err := db.conn.QueryRow(query, accountID).Scan(&acc.ID, &name, &balance, &openingBalance, &currency, &accountType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account with ID %d not found", accountID)
		}
		return nil, fmt.Errorf("failed to query account: %w", err)
	}

	if name.Valid {
		acc.Name = name.String
	}

	// Calculate balance from opening balance + transactions (exactly as Python implementation)
	// Python code: current_balance = opening_balance + transaction_total
	calculatedBalance, err := db.calculateAccountBalance(accountID, openingBalance)
	if err == nil {
		acc.Balance = calculatedBalance
	} else {
		// Fallback to opening balance or stored balance
		if openingBalance.Valid {
			acc.Balance = openingBalance.Float64
		} else if balance.Valid {
			acc.Balance = balance.Float64
		} else {
			acc.Balance = 0.0
		}
	}

	if currency.Valid {
		acc.Currency = currency.String
	}
	if accountType.Valid {
		acc.AccountType = accountType.String
	}

	return &acc, nil
}
