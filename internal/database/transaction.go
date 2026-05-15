package database

import (
	"database/sql"
	"fmt"
)

// Transaction represents a MoneyWiz transaction
type Transaction struct {
	ID           int64   `json:"id"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Description  string  `json:"description"`
	AccountID    int64   `json:"account_id"`
	AccountName  string  `json:"account_name"`
	Currency     string  `json:"currency"`
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	MovementType string  `json:"movement_type"`
}

// GetTransactions retrieves transactions for an account (or all transactions if accountID is 0)
// Transactions are entity types 37, 45, 46, 47, 43 (transfers), linked via ZACCOUNT2, using ZAMOUNT1
// Dates are Core Data timestamps (seconds since 2001-01-01), converted to ISO format
func (db *DB) GetTransactions(accountID int64, limit int) ([]Transaction, error) {
	var query string
	var args []interface{}

	if accountID > 0 {
		query = `
			SELECT t.Z_PK, t.ZAMOUNT1, 
				CASE WHEN t.ZDATE1 IS NOT NULL THEN datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds') ELSE NULL END as transaction_date, 
				t.ZDESC2, t.ZACCOUNT2, a.ZNAME, a.ZCURRENCYNAME, c.Z_PK, c.ZNAME2
			FROM ZSYNCOBJECT t
			LEFT JOIN ZSYNCOBJECT a ON a.Z_PK = t.ZACCOUNT2 AND a.Z_ENT IN (10, 11, 12, 13, 15, 16)
			LEFT JOIN ZCATEGORYASSIGMENT ca ON ca.ZTRANSACTION = t.Z_PK
			LEFT JOIN ZSYNCOBJECT c ON c.Z_PK = ca.ZCATEGORY AND c.Z_ENT = 19
			WHERE t.Z_ENT IN (37, 45, 46, 47, 43) AND t.ZAMOUNT1 IS NOT NULL AND (t.ZACCOUNT2 = ? OR t.ZACCOUNT = ?)
			ORDER BY t.ZDATE1 DESC
			LIMIT ?
		`
		args = []interface{}{accountID, accountID, limit}
	} else {
		query = `
			SELECT t.Z_PK, t.ZAMOUNT1, 
				CASE WHEN t.ZDATE1 IS NOT NULL THEN datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds') ELSE NULL END as transaction_date, 
				t.ZDESC2, t.ZACCOUNT2, a.ZNAME, a.ZCURRENCYNAME, c.Z_PK, c.ZNAME2
			FROM ZSYNCOBJECT t
			LEFT JOIN ZSYNCOBJECT a ON a.Z_PK = t.ZACCOUNT2 AND a.Z_ENT IN (10, 11, 12, 13, 15, 16)
			LEFT JOIN ZCATEGORYASSIGMENT ca ON ca.ZTRANSACTION = t.Z_PK
			LEFT JOIN ZSYNCOBJECT c ON c.Z_PK = ca.ZCATEGORY AND c.Z_ENT = 19
			WHERE t.Z_ENT IN (37, 45, 46, 47, 43) AND t.ZAMOUNT1 IS NOT NULL
			ORDER BY t.ZDATE1 DESC
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
		var accountName sql.NullString
		var currency sql.NullString
		var categoryID sql.NullInt64
		var categoryName sql.NullString
		err := rows.Scan(&txn.ID, &txn.Amount, &date, &desc, &txn.AccountID, &accountName, &currency, &categoryID, &categoryName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		if date.Valid {
			txn.Date = date.String
		}
		if desc.Valid {
			txn.Description = desc.String
		}
		if accountName.Valid {
			txn.AccountName = accountName.String
		}
		if currency.Valid {
			txn.Currency = currency.String
		}
		if categoryID.Valid {
			txn.CategoryID = categoryID.Int64
		}
		if categoryName.Valid {
			txn.CategoryName = categoryName.String
		}
		txn.MovementType = detectMovementType(txn.Description)
		txn.CategoryName = fallbackCategoryName(txn.CategoryName, txn.Description)
		transactions = append(transactions, txn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}
