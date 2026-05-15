package database

import "testing"

func TestGenerateSavingsRecommendationsNegativeSavingsRate(t *testing.T) {
	db := &DB{}

	got := db.generateSavingsRecommendations(
		-10,
		1000,
		1100,
		1000,
		0,
		nil,
		1,
	)

	assertRecommendationTitles(t, got, []string{
		"Negative Savings Rate",
	})

	if len(got) != 1 {
		t.Fatalf("recommendation count = %d, want 1", len(got))
	}
	if got[0].Impact != 100 {
		t.Fatalf("negative savings impact = %v, want 100", got[0].Impact)
	}
}

func TestGenerateSavingsRecommendationsExcellentSavingsRate(t *testing.T) {
	db := &DB{}

	got := db.generateSavingsRecommendations(
		30,
		10000,
		7000,
		1000,
		200,
		nil,
		1,
	)

	assertRecommendationTitles(t, got, []string{
		"Excellent Savings Rate",
	})

	if len(got) != 1 {
		t.Fatalf("recommendation count = %d, want 1", len(got))
	}
	if got[0].Type != "positive" {
		t.Fatalf("recommendation type = %q, want %q", got[0].Type, "positive")
	}
}

func TestGenerateSavingsRecommendationsCategoryAndRatioSignals(t *testing.T) {
	db := &DB{}

	topCategories := []CategorySpending{
		{CategoryName: "Dining", TotalAmount: 3500, Percentage: 35, TransactionCount: 12},
		{CategoryName: "Travel", TotalAmount: 2000, Percentage: 20, TransactionCount: 4},
		{CategoryName: "Shopping", TotalAmount: 1600, Percentage: 16, TransactionCount: 8},
	}

	got := db.generateSavingsRecommendations(
		15,
		10000,
		8500,
		1000,
		950,
		topCategories,
		2,
	)

	assertRecommendationTitles(t, got, []string{
		"Moderate Savings Rate",
		"Review Spending on Dining",
		"Multiple High-Spending Categories",
		"High Spending Ratio",
		"Build Emergency Fund",
	})

	if len(got) != 5 {
		t.Fatalf("recommendation count = %d, want 5", len(got))
	}

	review := findRecommendationByTitle(t, got, "Review Spending on Dining")
	if review.Impact != 350 {
		t.Fatalf("review spending impact = %v, want 350", review.Impact)
	}

	highRatio := findRecommendationByTitle(t, got, "High Spending Ratio")
	if highRatio.Priority != "high" {
		t.Fatalf("high spending ratio priority = %q, want %q", highRatio.Priority, "high")
	}
}

func assertRecommendationTitles(t *testing.T, got []SavingsRecommendation, want []string) {
	t.Helper()

	for _, title := range want {
		if !hasRecommendationTitle(got, title) {
			t.Fatalf("missing recommendation title %q in %+v", title, recommendationTitles(got))
		}
	}
}

func hasRecommendationTitle(recommendations []SavingsRecommendation, title string) bool {
	for _, recommendation := range recommendations {
		if recommendation.Title == title {
			return true
		}
	}
	return false
}

func findRecommendationByTitle(t *testing.T, recommendations []SavingsRecommendation, title string) SavingsRecommendation {
	t.Helper()

	for _, recommendation := range recommendations {
		if recommendation.Title == title {
			return recommendation
		}
	}

	t.Fatalf("missing recommendation title %q", title)
	return SavingsRecommendation{}
}

func recommendationTitles(recommendations []SavingsRecommendation) []string {
	titles := make([]string, 0, len(recommendations))
	for _, recommendation := range recommendations {
		titles = append(titles, recommendation.Title)
	}
	return titles
}
