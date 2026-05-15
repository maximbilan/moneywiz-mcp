package server

import (
	"sort"

	"github.com/moneywiz-mcp/internal/database"
)

func currencyMetaFromAccounts(accounts []database.Account) ([]string, bool, string) {
	set := make(map[string]struct{})
	for _, account := range accounts {
		if account.Currency != "" {
			set[account.Currency] = struct{}{}
		}
	}
	return currencyMetaFromSet(set)
}

func currencyMetaFromTransactions(transactions []database.Transaction) ([]string, bool, string) {
	set := make(map[string]struct{})
	for _, transaction := range transactions {
		if transaction.Currency != "" {
			set[transaction.Currency] = struct{}{}
		}
	}
	return currencyMetaFromSet(set)
}

func currencyMetaFromIncomeTrends(trends []database.IncomeTrend) ([]string, bool, string) {
	set := make(map[string]struct{})
	for _, trend := range trends {
		for currency := range trend.ByCurrency {
			if currency != "" {
				set[currency] = struct{}{}
			}
		}
	}
	return currencyMetaFromSet(set)
}

func currencyMetaFromSpendingTrends(trends []database.SpendingTrend) ([]string, bool, string) {
	set := make(map[string]struct{})
	for _, trend := range trends {
		for currency := range trend.ByCurrency {
			if currency != "" {
				set[currency] = struct{}{}
			}
		}
	}
	return currencyMetaFromSet(set)
}

func currencyMetaFromSet(set map[string]struct{}) ([]string, bool, string) {
	currencies := make([]string, 0, len(set))
	for currency := range set {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	if len(currencies) > 1 {
		return currencies, true, "Response includes multiple currencies. Prefer by_currency totals or per-transaction currency fields when interpreting amounts."
	}
	return currencies, false, ""
}
