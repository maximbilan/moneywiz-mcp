package database

import (
	"fmt"
)

// FinancialStats represents comprehensive financial statistics
type FinancialStats struct {
	TotalTransactions    int                  `json:"total_transactions"`
	TotalIncome          float64              `json:"total_income"`
	TotalSpending        float64              `json:"total_spending"`
	NetSavings           float64              `json:"net_savings"`
	AverageTransaction   float64              `json:"average_transaction"`
	LargestIncome        float64              `json:"largest_income"`
	LargestExpense       float64              `json:"largest_expense"`
	AccountCount         int                  `json:"account_count"`
	CategoryCount        int                  `json:"category_count"`
	FirstTransactionDate string               `json:"first_transaction_date"`
	LastTransactionDate  string               `json:"last_transaction_date"`
	DateRange            string               `json:"date_range"`
	IncomeTransactions   int                  `json:"income_transactions"`
	ExpenseTransactions  int                  `json:"expense_transactions"`
	ByYear               map[string]YearStats `json:"by_year"`
}

// YearStats represents statistics for a specific year
type YearStats struct {
	Year             string  `json:"year"`
	Income           float64 `json:"income"`
	Spending         float64 `json:"spending"`
	NetSavings       float64 `json:"net_savings"`
	TransactionCount int     `json:"transaction_count"`
}

// GetFinancialStats calculates comprehensive financial statistics from all historical data
func (db *DB) GetFinancialStats() (*FinancialStats, error) {
	// Get all transactions (no date limit)
	incomeData, err := db.GetIncomeData(0) // 0 = all data
	if err != nil {
		return nil, fmt.Errorf("failed to get income data: %w", err)
	}

	spendingData, err := db.GetSpendingData(0) // 0 = all data
	if err != nil {
		return nil, fmt.Errorf("failed to get spending data: %w", err)
	}

	// Get accounts and categories count
	accounts, err := db.GetAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	categories, err := db.GetCategories()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// Calculate totals
	var totalIncome float64
	var totalSpending float64
	var largestIncome float64
	var largestExpense float64
	var firstDate string
	var lastDate string
	byYear := make(map[string]*YearStats)

	// Process income transactions
	for _, i := range incomeData {
		totalIncome += i.Amount
		if i.Amount > largestIncome {
			largestIncome = i.Amount
		}

		// Track dates
		if i.Date != "" {
			if firstDate == "" || i.Date < firstDate {
				firstDate = i.Date
			}
			if lastDate == "" || i.Date > lastDate {
				lastDate = i.Date
			}
		}

		// Group by year
		if i.Year != "" {
			if byYear[i.Year] == nil {
				byYear[i.Year] = &YearStats{Year: i.Year}
			}
			byYear[i.Year].Income += i.Amount
			byYear[i.Year].TransactionCount++
		}
	}

	// Process spending transactions
	for _, s := range spendingData {
		totalSpending += s.Amount
		if s.Amount > largestExpense {
			largestExpense = s.Amount
		}

		// Track dates
		if s.Date != "" {
			if firstDate == "" || s.Date < firstDate {
				firstDate = s.Date
			}
			if lastDate == "" || s.Date > lastDate {
				lastDate = s.Date
			}
		}

		// Group by year
		if s.Year != "" {
			if byYear[s.Year] == nil {
				byYear[s.Year] = &YearStats{Year: s.Year}
			}
			byYear[s.Year].Spending += s.Amount
			byYear[s.Year].TransactionCount++
		}
	}

	// Calculate net savings and finalize year stats
	netSavings := totalIncome - totalSpending
	totalTransactions := len(incomeData) + len(spendingData)
	averageTransaction := 0.0
	if totalTransactions > 0 {
		averageTransaction = (totalIncome + totalSpending) / float64(totalTransactions)
	}

	// Finalize year stats
	yearStatsMap := make(map[string]YearStats)
	for year, stats := range byYear {
		stats.NetSavings = stats.Income - stats.Spending
		yearStatsMap[year] = *stats
	}

	// Format date range
	dateRange := ""
	if firstDate != "" && lastDate != "" {
		dateRange = fmt.Sprintf("%s to %s", firstDate, lastDate)
	} else if firstDate != "" {
		dateRange = fmt.Sprintf("Since %s", firstDate)
	}

	return &FinancialStats{
		TotalTransactions:    totalTransactions,
		TotalIncome:          totalIncome,
		TotalSpending:        totalSpending,
		NetSavings:           netSavings,
		AverageTransaction:   averageTransaction,
		LargestIncome:        largestIncome,
		LargestExpense:       largestExpense,
		AccountCount:         len(accounts),
		CategoryCount:        len(categories),
		FirstTransactionDate: firstDate,
		LastTransactionDate:  lastDate,
		DateRange:            dateRange,
		IncomeTransactions:   len(incomeData),
		ExpenseTransactions:  len(spendingData),
		ByYear:               yearStatsMap,
	}, nil
}
