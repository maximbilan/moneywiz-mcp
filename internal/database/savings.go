package database

import (
	"fmt"
	"math"
)

// SavingsRecommendation represents a savings recommendation
type SavingsRecommendation struct {
	Type        string  `json:"type"`         // "warning", "suggestion", "positive"
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`     // "high", "medium", "low"
	Impact      float64 `json:"impact"`       // Potential savings amount
}

// SavingsAnalysis represents comprehensive savings analysis
type SavingsAnalysis struct {
	Period              string                  `json:"period"`
	TotalIncome         float64                 `json:"total_income"`
	TotalSpending       float64                 `json:"total_spending"`
	NetSavings          float64                 `json:"net_savings"`
	SavingsRate         float64                 `json:"savings_rate"`         // Percentage
	AverageMonthlyIncome float64                 `json:"average_monthly_income"`
	AverageMonthlySpending float64                `json:"average_monthly_spending"`
	TopSpendingCategories []CategorySpending     `json:"top_spending_categories"`
	Recommendations     []SavingsRecommendation  `json:"recommendations"`
}

// CategorySpending represents spending by category
type CategorySpending struct {
	CategoryName string  `json:"category_name"`
	TotalAmount  float64 `json:"total_amount"`
	Percentage   float64 `json:"percentage"` // Percentage of total spending
	TransactionCount int `json:"transaction_count"`
}

// AnalyzeSavings analyzes income vs spending and provides recommendations
// months: number of months to analyze (default: 6)
func (db *DB) AnalyzeSavings(months int) (*SavingsAnalysis, error) {
	if months <= 0 {
		months = 6
	}

	// Get income and spending data
	incomeData, err := db.GetIncomeData(months)
	if err != nil {
		return nil, fmt.Errorf("failed to get income data: %w", err)
	}

	spendingData, err := db.GetSpendingData(months)
	if err != nil {
		return nil, fmt.Errorf("failed to get spending data: %w", err)
	}

	// Calculate totals
	var totalIncome float64
	var totalSpending float64
	spendingByCategory := make(map[string]int) // count for transaction tracking
	spendingAmountByCategory := make(map[string]float64)

	for _, i := range incomeData {
		totalIncome += i.Amount
	}

	for _, s := range spendingData {
		totalSpending += s.Amount
		spendingByCategory[s.CategoryName]++
		spendingAmountByCategory[s.CategoryName] += s.Amount
	}

	netSavings := totalIncome - totalSpending
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = (netSavings / totalIncome) * 100
	}

	// Calculate averages
	monthCount := float64(months)
	averageMonthlyIncome := totalIncome / monthCount
	averageMonthlySpending := totalSpending / monthCount

	// Get top spending categories
	type catSpend struct {
		name  string
		amount float64
		count int
	}
	var topCategories []catSpend
	for name, amount := range spendingAmountByCategory {
		topCategories = append(topCategories, catSpend{
			name:  name,
			amount: amount,
			count: spendingByCategory[name],
		})
	}

	// Sort by amount descending
	for i := 0; i < len(topCategories)-1; i++ {
		for j := i + 1; j < len(topCategories); j++ {
			if topCategories[i].amount < topCategories[j].amount {
				topCategories[i], topCategories[j] = topCategories[j], topCategories[i]
			}
		}
	}

	// Take top 5
	topN := 5
	if len(topCategories) < topN {
		topN = len(topCategories)
	}
	var topSpendingCategories []CategorySpending
	for i := 0; i < topN; i++ {
		percentage := 0.0
		if totalSpending > 0 {
			percentage = (topCategories[i].amount / totalSpending) * 100
		}
		topSpendingCategories = append(topSpendingCategories, CategorySpending{
			CategoryName:     topCategories[i].name,
			TotalAmount:      topCategories[i].amount,
			Percentage:       percentage,
			TransactionCount: topCategories[i].count,
		})
	}

	// Generate recommendations
	recommendations := db.generateSavingsRecommendations(
		savingsRate,
		totalIncome,
		totalSpending,
		averageMonthlyIncome,
		averageMonthlySpending,
		topSpendingCategories,
		months,
	)

	return &SavingsAnalysis{
		Period:                fmt.Sprintf("Last %d months", months),
		TotalIncome:           totalIncome,
		TotalSpending:         totalSpending,
		NetSavings:            netSavings,
		SavingsRate:           savingsRate,
		AverageMonthlyIncome:  averageMonthlyIncome,
		AverageMonthlySpending: averageMonthlySpending,
		TopSpendingCategories: topSpendingCategories,
		Recommendations:      recommendations,
	}, nil
}

// generateSavingsRecommendations generates recommendations based on financial data
func (db *DB) generateSavingsRecommendations(
	savingsRate float64,
	totalIncome float64,
	totalSpending float64,
	avgMonthlyIncome float64,
	avgMonthlySpending float64,
	topCategories []CategorySpending,
	months int,
) []SavingsRecommendation {
	var recommendations []SavingsRecommendation

	// Savings rate recommendations
	if savingsRate < 0 {
		recommendations = append(recommendations, SavingsRecommendation{
			Type:        "warning",
			Title:       "Negative Savings Rate",
			Description: fmt.Sprintf("You're spending more than you earn (%.1f%% savings rate). Consider reducing expenses or increasing income.", savingsRate),
			Priority:    "high",
			Impact:      math.Abs(totalSpending - totalIncome),
		})
	} else if savingsRate < 10 {
		recommendations = append(recommendations, SavingsRecommendation{
			Type:        "warning",
			Title:       "Low Savings Rate",
			Description: fmt.Sprintf("Your savings rate is %.1f%%. Financial experts recommend saving at least 20%% of income. Consider reducing discretionary spending.", savingsRate),
			Priority:    "high",
			Impact:      (totalIncome * 0.20) - (totalIncome - totalSpending),
		})
	} else if savingsRate < 20 {
		recommendations = append(recommendations, SavingsRecommendation{
			Type:        "suggestion",
			Title:       "Moderate Savings Rate",
			Description: fmt.Sprintf("Your savings rate is %.1f%%. You're on the right track! Aim for 20%%+ for better financial security.", savingsRate),
			Priority:    "medium",
			Impact:      (totalIncome * 0.20) - (totalIncome - totalSpending),
		})
	} else {
		recommendations = append(recommendations, SavingsRecommendation{
			Type:        "positive",
			Title:       "Excellent Savings Rate",
			Description: fmt.Sprintf("Great job! Your savings rate is %.1f%%, which exceeds the recommended 20%%. Keep up the good work!", savingsRate),
			Priority:    "low",
			Impact:      0,
		})
	}

	// Top spending category recommendations
	if len(topCategories) > 0 {
		topCategory := topCategories[0]
		if topCategory.Percentage > 30 {
			potentialSavings := topCategory.TotalAmount * 0.10 // 10% reduction
			recommendations = append(recommendations, SavingsRecommendation{
				Type:        "suggestion",
				Title:       fmt.Sprintf("Review Spending on %s", topCategory.CategoryName),
				Description: fmt.Sprintf("%s accounts for %.1f%% of your spending. A 10%% reduction could save you %.2f per month.", topCategory.CategoryName, topCategory.Percentage, potentialSavings/float64(months)),
				Priority:    "medium",
				Impact:      potentialSavings,
			})
		}

		// Multiple high-spending categories
		highSpendingCount := 0
		for _, cat := range topCategories {
			if cat.Percentage > 15 {
				highSpendingCount++
			}
		}
		if highSpendingCount >= 3 {
			recommendations = append(recommendations, SavingsRecommendation{
				Type:        "suggestion",
				Title:       "Multiple High-Spending Categories",
				Description: fmt.Sprintf("You have %d categories each accounting for over 15%% of spending. Consider reviewing your budget priorities.", highSpendingCount),
				Priority:    "medium",
				Impact:      avgMonthlySpending * 0.05, // 5% overall reduction potential
			})
		}
	}

	// Spending vs income ratio
	spendingRatio := 0.0
	if avgMonthlyIncome > 0 {
		spendingRatio = (avgMonthlySpending / avgMonthlyIncome) * 100
	}
	if spendingRatio > 90 {
		recommendations = append(recommendations, SavingsRecommendation{
			Type:        "warning",
			Title:       "High Spending Ratio",
			Description: fmt.Sprintf("You're spending %.1f%% of your income. This leaves little room for savings and unexpected expenses.", spendingRatio),
			Priority:    "high",
			Impact:      avgMonthlySpending * 0.10, // 10% reduction potential
		})
	}

	// Income stability recommendation
	if avgMonthlyIncome > 0 && avgMonthlySpending > 0 {
		monthsOfExpenses := (totalIncome - totalSpending) / avgMonthlySpending
		if monthsOfExpenses < 3 {
			recommendations = append(recommendations, SavingsRecommendation{
				Type:        "suggestion",
				Title:       "Build Emergency Fund",
				Description: fmt.Sprintf("Aim to save 3-6 months of expenses (%.2f per month) as an emergency fund. You currently have about %.1f months saved.", avgMonthlySpending, monthsOfExpenses),
				Priority:    "high",
				Impact:      avgMonthlySpending * 3, // 3 months target
			})
		}
	}

	return recommendations
}


