package database

import (
	"database/sql"
	"fmt"
)

// Transaction represents a MoneyWiz transaction
type Transaction struct {
	ID          int64   `json:"id"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	AccountID   int64   `json:"account_id"`
}

// GetTransactions retrieves transactions for an account (or all transactions if accountID is 0)
// Transactions are entity types 37, 45, 46, 47, 43 (transfers), linked via ZACCOUNT2, using ZAMOUNT1
// Dates are Core Data timestamps (seconds since 2001-01-01), converted to ISO format
func (db *DB) GetTransactions(accountID int64, limit int) ([]Transaction, error) {
	var query string
	var args []interface{}

	if accountID > 0 {
		query = `
			SELECT Z_PK, ZAMOUNT1, 
				CASE WHEN ZDATE1 IS NOT NULL THEN datetime('2001-01-01', '+' || CAST(ZDATE1 AS INTEGER) || ' seconds') ELSE NULL END as transaction_date, 
				ZDESC2, ZACCOUNT2
			FROM ZSYNCOBJECT
			WHERE Z_ENT IN (37, 45, 46, 47, 43) AND ZAMOUNT1 IS NOT NULL AND (ZACCOUNT2 = ? OR ZACCOUNT = ?)
			ORDER BY ZDATE1 DESC
			LIMIT ?
		`
		args = []interface{}{accountID, accountID, limit}
	} else {
		query = `
			SELECT Z_PK, ZAMOUNT1, 
				CASE WHEN ZDATE1 IS NOT NULL THEN datetime('2001-01-01', '+' || CAST(ZDATE1 AS INTEGER) || ' seconds') ELSE NULL END as transaction_date, 
				ZDESC2, ZACCOUNT2
			FROM ZSYNCOBJECT
			WHERE Z_ENT IN (37, 45, 46, 47, 43) AND ZAMOUNT1 IS NOT NULL
			ORDER BY ZDATE1 DESC
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
