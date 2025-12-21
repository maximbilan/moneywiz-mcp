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

// SpendingData represents spending data for trend analysis
type SpendingData struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Month        string  `json:"month"` // YYYY-MM format
	Year         string  `json:"year"`  // YYYY format
}

// GetSpendingData retrieves spending transactions with category information
// Returns expenses (negative amounts) grouped by category and date
// months: number of months to look back (0 = all data)
func (db *DB) GetSpendingData(months int) ([]SpendingData, error) {
	// Calculate date range: months back from now
	// Core Data timestamp: seconds since 2001-01-01
	// Get the latest transaction date to calculate the cutoff
	
	var query string
	if months > 0 {
		// Calculate cutoff timestamp: months * average seconds per month (30.44 days)
		// We'll use a subquery to get the max date and calculate backwards
		query = `
			SELECT 
				COALESCE(c.Z_PK, 0) as category_id,
				COALESCE(c.ZNAME2, 'Uncategorized') as category_name,
				ABS(t.ZAMOUNT1) as amount,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds') ELSE NULL END as transaction_date,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN strftime('%Y-%m', datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds')) ELSE NULL END as month,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN strftime('%Y', datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds')) ELSE NULL END as year
			FROM ZSYNCOBJECT t
			LEFT JOIN ZCATEGORYASSIGMENT ca ON ca.ZTRANSACTION = t.Z_PK
			LEFT JOIN ZSYNCOBJECT c ON c.Z_PK = ca.ZCATEGORY AND c.Z_ENT = 19
			WHERE t.Z_ENT IN (37, 45, 46, 47, 43)
			AND t.ZAMOUNT1 < 0
			AND t.ZDATE1 IS NOT NULL
			AND t.ZDATE1 >= (SELECT MAX(ZDATE1) FROM ZSYNCOBJECT WHERE Z_ENT IN (37, 45, 46, 47, 43) AND ZDATE1 IS NOT NULL) - (? * 2629746)
			ORDER BY t.ZDATE1 DESC
		`
	} else {
		query = `
			SELECT 
				COALESCE(c.Z_PK, 0) as category_id,
				COALESCE(c.ZNAME2, 'Uncategorized') as category_name,
				ABS(t.ZAMOUNT1) as amount,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds') ELSE NULL END as transaction_date,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN strftime('%Y-%m', datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds')) ELSE NULL END as month,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN strftime('%Y', datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds')) ELSE NULL END as year
			FROM ZSYNCOBJECT t
			LEFT JOIN ZCATEGORYASSIGMENT ca ON ca.ZTRANSACTION = t.Z_PK
			LEFT JOIN ZSYNCOBJECT c ON c.Z_PK = ca.ZCATEGORY AND c.Z_ENT = 19
			WHERE t.Z_ENT IN (37, 45, 46, 47, 43)
			AND t.ZAMOUNT1 < 0
			AND t.ZDATE1 IS NOT NULL
			ORDER BY t.ZDATE1 DESC
		`
	}

	var rows *sql.Rows
	var err error
	if months > 0 {
		rows, err = db.conn.Query(query, months)
	} else {
		rows, err = db.conn.Query(query)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query spending data: %w", err)
	}
	defer rows.Close()

	var spending []SpendingData
	for rows.Next() {
		var sd SpendingData
		var categoryID sql.NullInt64
		var categoryName sql.NullString
		var date sql.NullString
		var month sql.NullString
		var year sql.NullString
		
		err := rows.Scan(&categoryID, &categoryName, &sd.Amount, &date, &month, &year)
		if err != nil {
			return nil, fmt.Errorf("failed to scan spending data: %w", err)
		}
		
		if categoryID.Valid {
			sd.CategoryID = categoryID.Int64
		}
		if categoryName.Valid {
			sd.CategoryName = categoryName.String
		}
		if date.Valid {
			sd.Date = date.String
		}
		if month.Valid {
			sd.Month = month.String
		}
		if year.Valid {
			sd.Year = year.String
		}
		
		spending = append(spending, sd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating spending data: %w", err)
	}

	return spending, nil
}

// SpendingTrend represents aggregated spending trend data
type SpendingTrend struct {
	Period         string             `json:"period"`          // "YYYY-MM" or "YYYY"
	TotalSpending  float64            `json:"total_spending"`
	TransactionCount int               `json:"transaction_count"`
	ByCategory     map[string]float64 `json:"by_category"`     // Category name -> total
}

// AnalyzeSpendingTrends analyzes spending trends grouped by time period and category
// groupBy: "month" or "year"
// months: number of months to analyze (default: 6)
func (db *DB) AnalyzeSpendingTrends(groupBy string, months int) ([]SpendingTrend, error) {
	if months <= 0 {
		months = 6
	}
	if groupBy != "month" && groupBy != "year" {
		groupBy = "month"
	}

	spending, err := db.GetSpendingData(months)
	if err != nil {
		return nil, err
	}

	// Group by period
	trendsMap := make(map[string]*SpendingTrend)
	
	for _, s := range spending {
		var period string
		if groupBy == "year" {
			period = s.Year
		} else {
			period = s.Month
		}
		
		if period == "" {
			continue
		}
		
		if trendsMap[period] == nil {
			trendsMap[period] = &SpendingTrend{
				Period:        period,
				ByCategory:    make(map[string]float64),
			}
		}
		
		trend := trendsMap[period]
		trend.TotalSpending += s.Amount
		trend.TransactionCount++
		trend.ByCategory[s.CategoryName] += s.Amount
	}

	// Convert to slice and sort by period
	var trends []SpendingTrend
	for _, trend := range trendsMap {
		trends = append(trends, *trend)
	}

	// Simple sort by period string (works for YYYY-MM and YYYY)
	for i := 0; i < len(trends)-1; i++ {
		for j := i + 1; j < len(trends); j++ {
			if trends[i].Period > trends[j].Period {
				trends[i], trends[j] = trends[j], trends[i]
			}
		}
	}

	return trends, nil
}

// IncomeData represents income data for trend analysis
type IncomeData struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Month        string  `json:"month"` // YYYY-MM format
	Year         string  `json:"year"`  // YYYY format
}

// GetIncomeData retrieves income transactions with category information
// Returns income (positive amounts) grouped by category and date
// months: number of months to look back (0 = all data)
func (db *DB) GetIncomeData(months int) ([]IncomeData, error) {
	// Calculate date range: months back from now
	// Core Data timestamp: seconds since 2001-01-01
	// Get the latest transaction date to calculate the cutoff
	
	var query string
	if months > 0 {
		// Calculate cutoff timestamp: months * average seconds per month (30.44 days)
		// We'll use a subquery to get the max date and calculate backwards
		query = `
			SELECT 
				COALESCE(c.Z_PK, 0) as category_id,
				COALESCE(c.ZNAME2, 'Uncategorized') as category_name,
				t.ZAMOUNT1 as amount,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds') ELSE NULL END as transaction_date,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN strftime('%Y-%m', datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds')) ELSE NULL END as month,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN strftime('%Y', datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds')) ELSE NULL END as year
			FROM ZSYNCOBJECT t
			LEFT JOIN ZCATEGORYASSIGMENT ca ON ca.ZTRANSACTION = t.Z_PK
			LEFT JOIN ZSYNCOBJECT c ON c.Z_PK = ca.ZCATEGORY AND c.Z_ENT = 19
			WHERE t.Z_ENT IN (37, 45, 46, 47, 43)
			AND t.ZAMOUNT1 > 0
			AND t.ZDATE1 IS NOT NULL
			AND t.ZDATE1 >= (SELECT MAX(ZDATE1) FROM ZSYNCOBJECT WHERE Z_ENT IN (37, 45, 46, 47, 43) AND ZDATE1 IS NOT NULL) - (? * 2629746)
			ORDER BY t.ZDATE1 DESC
		`
	} else {
		query = `
			SELECT 
				COALESCE(c.Z_PK, 0) as category_id,
				COALESCE(c.ZNAME2, 'Uncategorized') as category_name,
				t.ZAMOUNT1 as amount,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds') ELSE NULL END as transaction_date,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN strftime('%Y-%m', datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds')) ELSE NULL END as month,
				CASE WHEN t.ZDATE1 IS NOT NULL THEN strftime('%Y', datetime('2001-01-01', '+' || CAST(t.ZDATE1 AS INTEGER) || ' seconds')) ELSE NULL END as year
			FROM ZSYNCOBJECT t
			LEFT JOIN ZCATEGORYASSIGMENT ca ON ca.ZTRANSACTION = t.Z_PK
			LEFT JOIN ZSYNCOBJECT c ON c.Z_PK = ca.ZCATEGORY AND c.Z_ENT = 19
			WHERE t.Z_ENT IN (37, 45, 46, 47, 43)
			AND t.ZAMOUNT1 > 0
			AND t.ZDATE1 IS NOT NULL
			ORDER BY t.ZDATE1 DESC
		`
	}

	var rows *sql.Rows
	var err error
	if months > 0 {
		rows, err = db.conn.Query(query, months)
	} else {
		rows, err = db.conn.Query(query)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query income data: %w", err)
	}
	defer rows.Close()

	var income []IncomeData
	for rows.Next() {
		var id IncomeData
		var categoryID sql.NullInt64
		var categoryName sql.NullString
		var date sql.NullString
		var month sql.NullString
		var year sql.NullString
		
		err := rows.Scan(&categoryID, &categoryName, &id.Amount, &date, &month, &year)
		if err != nil {
			return nil, fmt.Errorf("failed to scan income data: %w", err)
		}
		
		if categoryID.Valid {
			id.CategoryID = categoryID.Int64
		}
		if categoryName.Valid {
			id.CategoryName = categoryName.String
		}
		if date.Valid {
			id.Date = date.String
		}
		if month.Valid {
			id.Month = month.String
		}
		if year.Valid {
			id.Year = year.String
		}
		
		income = append(income, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating income data: %w", err)
	}

	return income, nil
}

// IncomeTrend represents aggregated income trend data
type IncomeTrend struct {
	Period         string             `json:"period"`          // "YYYY-MM" or "YYYY"
	TotalIncome    float64            `json:"total_income"`
	TransactionCount int               `json:"transaction_count"`
	ByCategory     map[string]float64 `json:"by_category"`     // Category name -> total
}

// AnalyzeIncomeTrends analyzes income trends grouped by time period and category
// groupBy: "month" or "year"
// months: number of months to analyze (default: 6)
func (db *DB) AnalyzeIncomeTrends(groupBy string, months int) ([]IncomeTrend, error) {
	if months <= 0 {
		months = 6
	}
	if groupBy != "month" && groupBy != "year" {
		groupBy = "month"
	}

	income, err := db.GetIncomeData(months)
	if err != nil {
		return nil, err
	}

	// Group by period
	trendsMap := make(map[string]*IncomeTrend)
	
	for _, i := range income {
		var period string
		if groupBy == "year" {
			period = i.Year
		} else {
			period = i.Month
		}
		
		if period == "" {
			continue
		}
		
		if trendsMap[period] == nil {
			trendsMap[period] = &IncomeTrend{
				Period:        period,
				ByCategory:    make(map[string]float64),
			}
		}
		
		trend := trendsMap[period]
		trend.TotalIncome += i.Amount
		trend.TransactionCount++
		trend.ByCategory[i.CategoryName] += i.Amount
	}

	// Convert to slice and sort by period
	var trends []IncomeTrend
	for _, trend := range trendsMap {
		trends = append(trends, *trend)
	}

	// Simple sort by period string (works for YYYY-MM and YYYY)
	for i := 0; i < len(trends)-1; i++ {
		for j := i + 1; j < len(trends); j++ {
			if trends[i].Period > trends[j].Period {
				trends[i], trends[j] = trends[j], trends[i]
			}
		}
	}

	return trends, nil
}
