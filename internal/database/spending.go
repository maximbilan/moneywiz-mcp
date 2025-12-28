package database

import (
	"database/sql"
	"fmt"
)

// SpendingData represents spending data for trend analysis
type SpendingData struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Month        string  `json:"month"` // YYYY-MM format
	Year         string  `json:"year"`  // YYYY format
}

// SpendingTrend represents aggregated spending trend data
type SpendingTrend struct {
	Period           string             `json:"period"` // "YYYY-MM" or "YYYY"
	TotalSpending    float64            `json:"total_spending"`
	TransactionCount int                `json:"transaction_count"`
	ByCategory       map[string]float64 `json:"by_category"` // Category name -> total
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

// AnalyzeSpendingTrends analyzes spending trends grouped by time period and category
// groupBy: "month" or "year"
// months: number of months to analyze (0 = all historical data)
func (db *DB) AnalyzeSpendingTrends(groupBy string, months int) ([]SpendingTrend, error) {
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
				Period:     period,
				ByCategory: make(map[string]float64),
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
