package database

import (
	"database/sql"
	"fmt"
)

// IncomeData represents income data for trend analysis
type IncomeData struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Month        string  `json:"month"` // YYYY-MM format
	Year         string  `json:"year"`  // YYYY format
}

// IncomeTrend represents aggregated income trend data
type IncomeTrend struct {
	Period           string             `json:"period"` // "YYYY-MM" or "YYYY"
	TotalIncome      float64            `json:"total_income"`
	TransactionCount int                `json:"transaction_count"`
	ByCategory       map[string]float64 `json:"by_category"` // Category name -> total
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

// AnalyzeIncomeTrends analyzes income trends grouped by time period and category
// groupBy: "month" or "year"
// months: number of months to analyze (0 = all historical data)
func (db *DB) AnalyzeIncomeTrends(groupBy string, months int) ([]IncomeTrend, error) {
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
				Period:     period,
				ByCategory: make(map[string]float64),
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
