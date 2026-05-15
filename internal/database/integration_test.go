package database

import (
	"database/sql"
	"math"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestAnalyzeSavingsWithFixtureDB(t *testing.T) {
	db := newFixtureDB(t)
	defer db.Close()

	got, err := db.AnalyzeSavings(0)
	if err != nil {
		t.Fatalf("AnalyzeSavings: %v", err)
	}

	if got.Period != "All data (2 months)" {
		t.Fatalf("period = %q, want %q", got.Period, "All data (2 months)")
	}
	assertFloatClose(t, "total income", got.TotalIncome, 5500, 0.001)
	assertFloatClose(t, "total spending", got.TotalSpending, 1500, 0.001)
	assertFloatClose(t, "net savings", got.NetSavings, 4000, 0.001)
	assertFloatClose(t, "average monthly income", got.AverageMonthlyIncome, 2750, 0.001)
	assertFloatClose(t, "average monthly spending", got.AverageMonthlySpending, 750, 0.001)
	assertFloatClose(t, "savings rate", got.SavingsRate, 72.7272727, 0.001)
	if got.MixedCurrencies {
		t.Fatal("mixed currencies = true, want false")
	}
	if got.PrimaryCurrency != "USD" {
		t.Fatalf("primary currency = %q, want %q", got.PrimaryCurrency, "USD")
	}
	if len(got.Currencies) != 1 || got.Currencies[0] != "USD" {
		t.Fatalf("currencies = %#v, want [USD]", got.Currencies)
	}
	usd, ok := got.ByCurrency["USD"]
	if !ok {
		t.Fatal("missing USD by_currency summary")
	}
	assertFloatClose(t, "usd total income", usd.TotalIncome, 5500, 0.001)
	assertFloatClose(t, "usd total spending", usd.TotalSpending, 1500, 0.001)
	assertFloatClose(t, "usd net savings", usd.NetSavings, 4000, 0.001)

	if len(got.TopSpendingCategories) != 2 {
		t.Fatalf("top spending categories len = %d, want 2", len(got.TopSpendingCategories))
	}
	if got.TopSpendingCategories[0].CategoryName != "Rent" {
		t.Fatalf("top category[0] = %q, want %q", got.TopSpendingCategories[0].CategoryName, "Rent")
	}
	assertFloatClose(t, "rent total", got.TopSpendingCategories[0].TotalAmount, 1200, 0.001)
	assertFloatClose(t, "rent percentage", got.TopSpendingCategories[0].Percentage, 80, 0.001)
	if got.TopSpendingCategories[1].CategoryName != "Groceries" {
		t.Fatalf("top category[1] = %q, want %q", got.TopSpendingCategories[1].CategoryName, "Groceries")
	}

	assertRecommendationPresent(t, got.Recommendations, "Excellent Savings Rate")
	assertRecommendationPresent(t, got.Recommendations, "Review Spending on Rent")
	if len(got.Recommendations) != 2 {
		t.Fatalf("recommendations len = %d, want 2", len(got.Recommendations))
	}
}

func TestGetAccountsAndAccountBalanceWithFixtureDB(t *testing.T) {
	db := newFixtureDB(t)
	defer db.Close()

	accounts, err := db.GetAccounts()
	if err != nil {
		t.Fatalf("GetAccounts: %v", err)
	}

	if len(accounts) != 1 {
		t.Fatalf("accounts len = %d, want 1", len(accounts))
	}

	account := accounts[0]
	if account.ID != 1 {
		t.Fatalf("account id = %d, want 1", account.ID)
	}
	if account.Name != "Checking" {
		t.Fatalf("account name = %q, want %q", account.Name, "Checking")
	}
	if account.Currency != "USD" {
		t.Fatalf("account currency = %q, want %q", account.Currency, "USD")
	}
	if account.AccountType != "bank" {
		t.Fatalf("account type = %q, want %q", account.AccountType, "bank")
	}
	assertFloatClose(t, "account balance", account.Balance, 5000, 0.001)

	single, err := db.GetAccountBalance(1)
	if err != nil {
		t.Fatalf("GetAccountBalance: %v", err)
	}
	assertFloatClose(t, "single account balance", single.Balance, 5000, 0.001)

	_, err = db.GetAccountBalance(999)
	if err == nil {
		t.Fatal("GetAccountBalance for missing account unexpectedly succeeded")
	}
}

func TestGetTransactionsAndCategoriesWithFixtureDB(t *testing.T) {
	db := newFixtureDB(t)
	defer db.Close()

	transactions, err := db.GetTransactions(1, 2)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(transactions) != 2 {
		t.Fatalf("transactions len = %d, want 2", len(transactions))
	}
	if transactions[0].ID != 1003 || transactions[1].ID != 1002 {
		t.Fatalf("transaction order = [%d %d], want [1003 1002]", transactions[0].ID, transactions[1].ID)
	}
	if transactions[0].Date != "2024-02-10 00:00:00" {
		t.Fatalf("latest transaction date = %q", transactions[0].Date)
	}
	if transactions[0].AccountName != "Checking" {
		t.Fatalf("transaction account name = %q, want %q", transactions[0].AccountName, "Checking")
	}
	if transactions[0].Currency != "USD" {
		t.Fatalf("transaction currency = %q, want %q", transactions[0].Currency, "USD")
	}
	if transactions[0].CategoryName != "Groceries" {
		t.Fatalf("transaction category = %q, want %q", transactions[0].CategoryName, "Groceries")
	}
	if transactions[0].MovementType != movementTypeRegular {
		t.Fatalf("transaction movement_type = %q, want %q", transactions[0].MovementType, movementTypeRegular)
	}

	categories, err := db.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}
	if len(categories) != 3 {
		t.Fatalf("categories len = %d, want 3", len(categories))
	}
	if categories[0].Name != "Groceries" || categories[1].Name != "Rent" || categories[2].Name != "Salary" {
		t.Fatalf("category order = [%s %s %s], want [Groceries Rent Salary]", categories[0].Name, categories[1].Name, categories[2].Name)
	}
}

func TestAnalyzeIncomeAndSpendingTrendsWithFixtureDB(t *testing.T) {
	db := newFixtureDB(t)
	defer db.Close()

	incomeMonthly, err := db.AnalyzeIncomeTrends("month", 0)
	if err != nil {
		t.Fatalf("AnalyzeIncomeTrends month: %v", err)
	}
	if len(incomeMonthly) != 2 {
		t.Fatalf("monthly income trends len = %d, want 2", len(incomeMonthly))
	}
	if incomeMonthly[0].Period != "2024-01" || incomeMonthly[1].Period != "2024-02" {
		t.Fatalf("monthly income periods = [%s %s]", incomeMonthly[0].Period, incomeMonthly[1].Period)
	}
	assertFloatClose(t, "jan income", incomeMonthly[0].TotalIncome, 3000, 0.001)
	assertFloatClose(t, "feb income", incomeMonthly[1].TotalIncome, 2500, 0.001)
	assertFloatClose(t, "salary jan breakdown", incomeMonthly[0].ByCategory["Salary"], 3000, 0.001)
	assertFloatClose(t, "jan income usd breakdown", incomeMonthly[0].ByCurrency["USD"], 3000, 0.001)

	spendingMonthly, err := db.AnalyzeSpendingTrends("month", 0)
	if err != nil {
		t.Fatalf("AnalyzeSpendingTrends month: %v", err)
	}
	if len(spendingMonthly) != 2 {
		t.Fatalf("monthly spending trends len = %d, want 2", len(spendingMonthly))
	}
	assertFloatClose(t, "jan spending", spendingMonthly[0].TotalSpending, 1200, 0.001)
	assertFloatClose(t, "feb spending", spendingMonthly[1].TotalSpending, 300, 0.001)
	assertFloatClose(t, "rent jan breakdown", spendingMonthly[0].ByCategory["Rent"], 1200, 0.001)
	assertFloatClose(t, "groceries feb breakdown", spendingMonthly[1].ByCategory["Groceries"], 300, 0.001)
	assertFloatClose(t, "jan spending usd breakdown", spendingMonthly[0].ByCurrency["USD"], 1200, 0.001)

	incomeYearly, err := db.AnalyzeIncomeTrends("year", 0)
	if err != nil {
		t.Fatalf("AnalyzeIncomeTrends year: %v", err)
	}
	if len(incomeYearly) != 1 {
		t.Fatalf("yearly income trends len = %d, want 1", len(incomeYearly))
	}
	assertFloatClose(t, "2024 yearly income", incomeYearly[0].TotalIncome, 5500, 0.001)
	assertFloatClose(t, "2024 yearly salary breakdown", incomeYearly[0].ByCategory["Salary"], 5500, 0.001)

	spendingYearly, err := db.AnalyzeSpendingTrends("invalid", 0)
	if err != nil {
		t.Fatalf("AnalyzeSpendingTrends invalid groupBy: %v", err)
	}
	if len(spendingYearly) != 2 {
		t.Fatalf("invalid groupBy should fall back to month; len = %d, want 2", len(spendingYearly))
	}
}

func TestGetFinancialStatsWithFixtureDB(t *testing.T) {
	db := newFixtureDB(t)
	defer db.Close()

	got, err := db.GetFinancialStats()
	if err != nil {
		t.Fatalf("GetFinancialStats: %v", err)
	}

	if got.TotalTransactions != 4 {
		t.Fatalf("total transactions = %d, want 4", got.TotalTransactions)
	}
	if got.IncomeTransactions != 2 {
		t.Fatalf("income transactions = %d, want 2", got.IncomeTransactions)
	}
	if got.ExpenseTransactions != 2 {
		t.Fatalf("expense transactions = %d, want 2", got.ExpenseTransactions)
	}
	if got.AccountCount != 1 {
		t.Fatalf("account count = %d, want 1", got.AccountCount)
	}
	if got.CategoryCount != 3 {
		t.Fatalf("category count = %d, want 3", got.CategoryCount)
	}

	assertFloatClose(t, "total income", got.TotalIncome, 5500, 0.001)
	assertFloatClose(t, "total spending", got.TotalSpending, 1500, 0.001)
	assertFloatClose(t, "net savings", got.NetSavings, 4000, 0.001)
	assertFloatClose(t, "average transaction", got.AverageTransaction, 1750, 0.001)
	assertFloatClose(t, "largest income", got.LargestIncome, 3000, 0.001)
	assertFloatClose(t, "largest expense", got.LargestExpense, 1200, 0.001)
	if got.MixedCurrencies {
		t.Fatal("mixed currencies = true, want false")
	}
	if got.PrimaryCurrency != "USD" {
		t.Fatalf("primary currency = %q, want %q", got.PrimaryCurrency, "USD")
	}
	if len(got.Currencies) != 1 || got.Currencies[0] != "USD" {
		t.Fatalf("currencies = %#v, want [USD]", got.Currencies)
	}
	usd, ok := got.ByCurrency["USD"]
	if !ok {
		t.Fatal("missing USD by_currency stats")
	}
	assertFloatClose(t, "usd average transaction", usd.AverageTransaction, 1750, 0.001)
	assertFloatClose(t, "usd total income", usd.TotalIncome, 5500, 0.001)
	assertFloatClose(t, "usd total spending", usd.TotalSpending, 1500, 0.001)

	if got.FirstTransactionDate != "2024-01-15 00:00:00" {
		t.Fatalf("first transaction date = %q", got.FirstTransactionDate)
	}
	if got.LastTransactionDate != "2024-02-10 00:00:00" {
		t.Fatalf("last transaction date = %q", got.LastTransactionDate)
	}
	if got.DateRange != "2024-01-15 00:00:00 to 2024-02-10 00:00:00" {
		t.Fatalf("date range = %q", got.DateRange)
	}

	year2024, ok := got.ByYear["2024"]
	if !ok {
		t.Fatalf("missing by_year entry for 2024")
	}
	assertFloatClose(t, "2024 income", year2024.Income, 5500, 0.001)
	assertFloatClose(t, "2024 spending", year2024.Spending, 1500, 0.001)
	assertFloatClose(t, "2024 net savings", year2024.NetSavings, 4000, 0.001)
	if year2024.TransactionCount != 4 {
		t.Fatalf("2024 transaction count = %d, want 4", year2024.TransactionCount)
	}
}

func TestMixedCurrencyStatsAndInternalMovementsWithFixtureDB(t *testing.T) {
	db := newFixtureDBWithExtraRows(t, func(conn *sql.DB) {
		mustExecSQL(t, conn, `
			INSERT INTO ZSYNCOBJECT (Z_PK, Z_ENT, ZNAME, ZBALLANCE, ZOPENINGBALANCE, ZCURRENCYNAME, ZTYPE)
			VALUES (2, 10, 'EUR Checking', 0, 500, 'EUR', 'bank');
		`)

		insertTransaction(t, conn, 2000, 37, 2000, "2024-02-15", "Consulting invoice", 2, 0, 100)
		insertTransaction(t, conn, 2001, 37, -500, "2024-02-16", "Groceries EU", 2, 0, 102)
		insertUncategorizedTransaction(t, conn, 2002, 43, -100, "2024-02-17", "Transfer to Cash", 1, 0)
		insertUncategorizedTransaction(t, conn, 2003, 43, 100, "2024-02-17", "Transfer from Checking", 2, 0)
		insertUncategorizedTransaction(t, conn, 2004, 46, -50, "2024-02-18", "ATM Withdrawal", 1, 0)
		insertUncategorizedTransaction(t, conn, 2005, 46, 50, "2024-02-18", "ATM Withdrawal", 2, 0)
	})
	defer db.Close()

	savings, err := db.AnalyzeSavings(0)
	if err != nil {
		t.Fatalf("AnalyzeSavings: %v", err)
	}
	if !savings.MixedCurrencies {
		t.Fatal("mixed currencies = false, want true")
	}
	if savings.CurrencyWarning == "" {
		t.Fatal("currency warning is empty")
	}
	if savings.PrimaryCurrency != "" {
		t.Fatalf("primary currency = %q, want empty", savings.PrimaryCurrency)
	}
	if len(savings.Currencies) != 2 || savings.Currencies[0] != "EUR" || savings.Currencies[1] != "USD" {
		t.Fatalf("currencies = %#v, want [EUR USD]", savings.Currencies)
	}
	assertFloatClose(t, "mixed total income", savings.TotalIncome, 7500, 0.001)
	assertFloatClose(t, "mixed total spending", savings.TotalSpending, 2000, 0.001)
	assertFloatClose(t, "usd income", savings.ByCurrency["USD"].TotalIncome, 5500, 0.001)
	assertFloatClose(t, "usd spending", savings.ByCurrency["USD"].TotalSpending, 1500, 0.001)
	assertFloatClose(t, "eur income", savings.ByCurrency["EUR"].TotalIncome, 2000, 0.001)
	assertFloatClose(t, "eur spending", savings.ByCurrency["EUR"].TotalSpending, 500, 0.001)

	stats, err := db.GetFinancialStats()
	if err != nil {
		t.Fatalf("GetFinancialStats: %v", err)
	}
	if !stats.MixedCurrencies {
		t.Fatal("mixed currencies = false, want true")
	}
	if stats.CurrencyWarning == "" {
		t.Fatal("currency warning is empty")
	}
	if stats.TotalTransactions != 6 {
		t.Fatalf("total transactions = %d, want 6", stats.TotalTransactions)
	}
	assertFloatClose(t, "stats total income", stats.TotalIncome, 7500, 0.001)
	assertFloatClose(t, "stats total spending", stats.TotalSpending, 2000, 0.001)
	assertFloatClose(t, "stats eur income", stats.ByCurrency["EUR"].TotalIncome, 2000, 0.001)
	assertFloatClose(t, "stats eur spending", stats.ByCurrency["EUR"].TotalSpending, 500, 0.001)

	transactions, err := db.GetTransactions(0, 10)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	foundTransfer := false
	foundATM := false
	for _, txn := range transactions {
		if txn.Description == "Transfer to Cash" {
			foundTransfer = true
			if txn.MovementType != movementTypeTransfer {
				t.Fatalf("transfer movement type = %q, want %q", txn.MovementType, movementTypeTransfer)
			}
			if txn.CategoryName != "Internal Transfer" {
				t.Fatalf("transfer category = %q, want %q", txn.CategoryName, "Internal Transfer")
			}
		}
		if txn.Description == "ATM Withdrawal" && txn.Amount < 0 {
			foundATM = true
			if txn.MovementType != movementTypeCashWithdrawal {
				t.Fatalf("atm movement type = %q, want %q", txn.MovementType, movementTypeCashWithdrawal)
			}
			if txn.CategoryName != "Cash Withdrawal" {
				t.Fatalf("atm category = %q, want %q", txn.CategoryName, "Cash Withdrawal")
			}
		}
	}
	if !foundTransfer {
		t.Fatal("did not find transfer transaction")
	}
	if !foundATM {
		t.Fatal("did not find atm withdrawal transaction")
	}
}

func newFixtureDB(t *testing.T) *DB {
	t.Helper()

	return newFixtureDBWithExtraRows(t, nil)
}

func newFixtureDBWithExtraRows(t *testing.T, extraRows func(conn *sql.DB)) *DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "moneywiz-fixture.sqlite")
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open fixture sqlite: %v", err)
	}
	defer conn.Close()

	mustExecSQL(t, conn, `
		CREATE TABLE ZSYNCOBJECT (
			Z_PK INTEGER PRIMARY KEY,
			Z_ENT INTEGER,
			ZNAME TEXT,
			ZDESC2 TEXT,
			ZBALLANCE REAL,
			ZOPENINGBALANCE REAL,
			ZCURRENCYNAME TEXT,
			ZTYPE TEXT,
			ZNAME2 TEXT,
			ZAMOUNT1 REAL,
			ZDATE1 REAL,
			ZACCOUNT2 INTEGER,
			ZACCOUNT INTEGER
		);
	`)
	mustExecSQL(t, conn, `
		CREATE TABLE ZCATEGORYASSIGMENT (
			ZTRANSACTION INTEGER,
			ZCATEGORY INTEGER
		);
	`)

	insertFixtureRows(t, conn)
	if extraRows != nil {
		extraRows(conn)
	}

	db, err := NewDB(path)
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	return db
}

func insertFixtureRows(t *testing.T, conn *sql.DB) {
	t.Helper()

	mustExecSQL(t, conn, `
		INSERT INTO ZSYNCOBJECT (Z_PK, Z_ENT, ZNAME, ZBALLANCE, ZOPENINGBALANCE, ZCURRENCYNAME, ZTYPE)
		VALUES (1, 10, 'Checking', 0, 1000, 'USD', 'bank');
	`)

	mustExecSQL(t, conn, `
		INSERT INTO ZSYNCOBJECT (Z_PK, Z_ENT, ZNAME2) VALUES
			(100, 19, 'Salary'),
			(101, 19, 'Rent'),
			(102, 19, 'Groceries');
	`)

	insertTransaction(t, conn, 1000, 37, 3000, "2024-01-15", "January salary", 1, 0, 100)
	insertTransaction(t, conn, 1001, 37, -1200, "2024-01-20", "Rent payment", 1, 0, 101)
	insertTransaction(t, conn, 1002, 37, 2500, "2024-02-05", "February salary", 1, 0, 100)
	insertTransaction(t, conn, 1003, 37, -300, "2024-02-10", "Groceries", 1, 0, 102)
}

func insertTransaction(t *testing.T, conn *sql.DB, id, ent int64, amount float64, date, description string, account2, account int64, category int64) {
	t.Helper()

	mustExecSQL(t, conn, `
		INSERT INTO ZSYNCOBJECT (Z_PK, Z_ENT, ZAMOUNT1, ZDATE1, ZDESC2, ZACCOUNT2, ZACCOUNT)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`, id, ent, amount, coreDataSeconds(t, date), description, account2, account)

	mustExecSQL(t, conn, `
		INSERT INTO ZCATEGORYASSIGMENT (ZTRANSACTION, ZCATEGORY)
		VALUES (?, ?);
	`, id, category)
}

func insertUncategorizedTransaction(t *testing.T, conn *sql.DB, id, ent int64, amount float64, date, description string, account2, account int64) {
	t.Helper()

	mustExecSQL(t, conn, `
		INSERT INTO ZSYNCOBJECT (Z_PK, Z_ENT, ZAMOUNT1, ZDATE1, ZDESC2, ZACCOUNT2, ZACCOUNT)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`, id, ent, amount, coreDataSeconds(t, date), description, account2, account)
}

func coreDataSeconds(t *testing.T, date string) float64 {
	t.Helper()

	ts, err := time.Parse("2006-01-02", date)
	if err != nil {
		t.Fatalf("parse date %q: %v", date, err)
	}
	coreDataEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	return ts.Sub(coreDataEpoch).Seconds()
}

func mustExecSQL(t *testing.T, conn *sql.DB, query string, args ...any) {
	t.Helper()

	if _, err := conn.Exec(query, args...); err != nil {
		t.Fatalf("exec query failed: %v\nquery:\n%s", err, query)
	}
}

func assertFloatClose(t *testing.T, label string, got, want, tolerance float64) {
	t.Helper()

	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func assertRecommendationPresent(t *testing.T, recommendations []SavingsRecommendation, title string) {
	t.Helper()

	for _, recommendation := range recommendations {
		if recommendation.Title == title {
			return
		}
	}

	t.Fatalf("missing recommendation %q", title)
}
