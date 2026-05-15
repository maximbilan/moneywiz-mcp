package server

import "testing"

func TestNormalizeTransactionParamsDefaultsLimitWhenMissingOrInvalid(t *testing.T) {
	tests := []struct {
		name        string
		accountID   float64
		limit       int
		wantAccount int64
		wantLimit   int
	}{
		{name: "explicit values", accountID: 249, limit: 20, wantAccount: 249, wantLimit: 20},
		{name: "zero limit", accountID: 12, limit: 0, wantAccount: 12, wantLimit: defaultTransactionLimit},
		{name: "negative limit", accountID: 0, limit: -5, wantAccount: 0, wantLimit: defaultTransactionLimit},
		{name: "fractional account id truncates", accountID: 42.9, limit: 10, wantAccount: 42, wantLimit: 10},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotAccount, gotLimit := normalizeTransactionParams(tc.accountID, tc.limit)
			if gotAccount != tc.wantAccount {
				t.Fatalf("account id = %d, want %d", gotAccount, tc.wantAccount)
			}
			if gotLimit != tc.wantLimit {
				t.Fatalf("limit = %d, want %d", gotLimit, tc.wantLimit)
			}
		})
	}
}

func TestNormalizeGroupBy(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "month", want: "month"},
		{input: "year", want: "year"},
		{input: "", want: "month"},
		{input: "weekly", want: "month"},
	}

	for _, tc := range tests {
		if got := normalizeGroupBy(tc.input); got != tc.want {
			t.Fatalf("normalizeGroupBy(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
