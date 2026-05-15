package server

const defaultTransactionLimit = 50

func normalizeTransactionParams(accountID float64, limit int) (int64, int) {
	if limit <= 0 {
		limit = defaultTransactionLimit
	}
	return int64(accountID), limit
}

func normalizeGroupBy(groupBy string) string {
	if groupBy != "month" && groupBy != "year" {
		return "month"
	}
	return groupBy
}
