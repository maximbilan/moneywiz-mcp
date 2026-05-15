package database

import (
	"sort"
	"strings"
)

const (
	movementTypeRegular        = "regular"
	movementTypeTransfer       = "transfer"
	movementTypeCashWithdrawal = "cash_withdrawal"
	movementTypeUncategorized  = "uncategorized"
)

func detectMovementType(description string) string {
	trimmed := strings.TrimSpace(strings.ToLower(description))

	switch {
	case strings.HasPrefix(trimmed, "transfer to "), strings.HasPrefix(trimmed, "transfer from "):
		return movementTypeTransfer
	case trimmed == "atm withdrawal", trimmed == "снятие наличных в банкоматe":
		return movementTypeCashWithdrawal
	case trimmed == "":
		return movementTypeUncategorized
	default:
		return movementTypeRegular
	}
}

func isInternalMovement(movementType string) bool {
	switch movementType {
	case movementTypeTransfer, movementTypeCashWithdrawal:
		return true
	default:
		return false
	}
}

func fallbackCategoryName(categoryName, description string) string {
	if strings.TrimSpace(categoryName) != "" {
		return categoryName
	}

	switch detectMovementType(description) {
	case movementTypeTransfer:
		return "Internal Transfer"
	case movementTypeCashWithdrawal:
		return "Cash Withdrawal"
	default:
		return "Uncategorized"
	}
}

func sortedCurrencyKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
