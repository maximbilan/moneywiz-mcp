package database

import (
	"fmt"
	"math"
)

// NetWorth represents net worth calculation
type NetWorth struct {
	TotalAssets      float64            `json:"total_assets"`
	TotalLiabilities float64            `json:"total_liabilities"`
	NetWorth         float64            `json:"net_worth"`
	AccountCount     int                `json:"account_count"`
	ByCurrency       map[string]float64 `json:"by_currency"` // Net worth by currency
	Accounts         []AccountSummary   `json:"accounts"`    // Summary of all accounts
}

// AccountSummary represents a summary of an account for net worth calculation
type AccountSummary struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
	Type     string  `json:"type"`
}

// CalculateNetWorth calculates the total net worth from all accounts
func (db *DB) CalculateNetWorth() (*NetWorth, error) {
	accounts, err := db.GetAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	var totalAssets float64
	var totalLiabilities float64
	byCurrency := make(map[string]float64)
	var accountSummaries []AccountSummary

	for _, acc := range accounts {
		accountSummary := AccountSummary{
			ID:       acc.ID,
			Name:     acc.Name,
			Balance:  acc.Balance,
			Currency: acc.Currency,
			Type:     acc.AccountType,
		}
		accountSummaries = append(accountSummaries, accountSummary)

		// Categorize as asset or liability
		// In MoneyWiz, positive balances are typically assets
		// Negative balances or specific account types might be liabilities
		// For simplicity, we'll treat all balances as assets (net worth = sum of all balances)
		// If balance is negative, it reduces net worth
		if acc.Balance >= 0 {
			totalAssets += acc.Balance
		} else {
			totalLiabilities += math.Abs(acc.Balance)
		}

		// Track by currency
		if acc.Currency != "" {
			byCurrency[acc.Currency] += acc.Balance
		}
	}

	netWorth := totalAssets - totalLiabilities

	return &NetWorth{
		TotalAssets:      totalAssets,
		TotalLiabilities: totalLiabilities,
		NetWorth:         netWorth,
		AccountCount:     len(accounts),
		ByCurrency:       byCurrency,
		Accounts:         accountSummaries,
	}, nil
}
