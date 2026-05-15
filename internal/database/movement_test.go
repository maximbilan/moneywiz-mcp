package database

import "testing"

func TestDetectMovementType(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{name: "regular", description: "Grocery store", want: movementTypeRegular},
		{name: "transfer to", description: "Transfer to Monobank USD", want: movementTypeTransfer},
		{name: "transfer from", description: "Transfer from Wise EUR", want: movementTypeTransfer},
		{name: "atm withdrawal", description: "ATM Withdrawal", want: movementTypeCashWithdrawal},
		{name: "uncategorized", description: "   ", want: movementTypeUncategorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectMovementType(tc.description); got != tc.want {
				t.Fatalf("detectMovementType(%q) = %q, want %q", tc.description, got, tc.want)
			}
		})
	}
}

func TestFallbackCategoryName(t *testing.T) {
	tests := []struct {
		name         string
		categoryName string
		description  string
		want         string
	}{
		{name: "preserve existing", categoryName: "Shop", description: "Transfer to Cash", want: "Shop"},
		{name: "transfer fallback", description: "Transfer to Revolut USD", want: "Internal Transfer"},
		{name: "cash withdrawal fallback", description: "ATM Withdrawal", want: "Cash Withdrawal"},
		{name: "uncategorized fallback", description: "Merchant without category", want: "Uncategorized"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := fallbackCategoryName(tc.categoryName, tc.description); got != tc.want {
				t.Fatalf("fallbackCategoryName(%q, %q) = %q, want %q", tc.categoryName, tc.description, got, tc.want)
			}
		})
	}
}
