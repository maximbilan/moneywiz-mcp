package server

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/moneywiz-mcp/internal/database"

	_ "github.com/mattn/go-sqlite3"
)

func TestHandleGetAccountBalanceMissingAccountIDReturnsError(t *testing.T) {
	srv := &Server{}

	result, err := srv.handleGetAccountBalance(context.Background(), newCallToolRequest("get_account_balance", map[string]any{}))
	if err != nil {
		t.Fatalf("handleGetAccountBalance returned protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error result")
	}
	assertSingleTextContains(t, result, "account_id")
}

func TestHandleListAccountsReturnsStructuredAccounts(t *testing.T) {
	srv := newTestServer(t)

	result, err := srv.handleListAccounts(context.Background(), newCallToolRequest("list_accounts", map[string]any{}))
	if err != nil {
		t.Fatalf("handleListAccounts returned protocol error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected successful result")
	}

	structured, ok := result.StructuredContent.(map[string]interface{})
	if !ok {
		t.Fatalf("structured content type = %T, want map[string]interface{}", result.StructuredContent)
	}
	accounts, ok := structured["accounts"].([]database.Account)
	if !ok {
		t.Fatalf("accounts type = %T, want []database.Account", structured["accounts"])
	}
	if len(accounts) != 1 {
		t.Fatalf("accounts len = %d, want 1", len(accounts))
	}
	if accounts[0].Name != "Checking" {
		t.Fatalf("account name = %q, want %q", accounts[0].Name, "Checking")
	}
	assertSingleTextContains(t, result, "Checking")
}

func TestHandleListTransactionsReturnsStructuredTransactions(t *testing.T) {
	srv := newTestServer(t)

	result, err := srv.handleListTransactions(context.Background(), newCallToolRequest("list_transactions", map[string]any{
		"account_id": 1,
		"limit":      2,
	}))
	if err != nil {
		t.Fatalf("handleListTransactions returned protocol error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected successful result")
	}

	structured, ok := result.StructuredContent.(map[string]interface{})
	if !ok {
		t.Fatalf("structured content type = %T, want map[string]interface{}", result.StructuredContent)
	}
	transactions, ok := structured["transactions"].([]database.Transaction)
	if !ok {
		t.Fatalf("transactions type = %T, want []database.Transaction", structured["transactions"])
	}
	if len(transactions) != 2 {
		t.Fatalf("transactions len = %d, want 2", len(transactions))
	}
	if transactions[0].ID != 1003 || transactions[1].ID != 1002 {
		t.Fatalf("transaction order = [%d %d], want [1003 1002]", transactions[0].ID, transactions[1].ID)
	}
	assertSingleTextContains(t, result, "Groceries")
}

func TestHandleAnalyzeSpendingTrendsInvalidGroupByFallsBackToMonth(t *testing.T) {
	srv := newTestServer(t)

	result, err := srv.handleAnalyzeSpendingTrends(context.Background(), newCallToolRequest("analyze_spending_trends", map[string]any{
		"group_by": "weekly",
		"months":   0,
	}))
	if err != nil {
		t.Fatalf("handleAnalyzeSpendingTrends returned protocol error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected successful result")
	}

	structured, ok := result.StructuredContent.(map[string]interface{})
	if !ok {
		t.Fatalf("structured content type = %T, want map[string]interface{}", result.StructuredContent)
	}
	groupBy, ok := structured["group_by"].(string)
	if !ok {
		t.Fatalf("group_by type = %T, want string", structured["group_by"])
	}
	if groupBy != "month" {
		t.Fatalf("group_by = %q, want %q", groupBy, "month")
	}
	trends, ok := structured["trends"].([]database.SpendingTrend)
	if !ok {
		t.Fatalf("trends type = %T, want []database.SpendingTrend", structured["trends"])
	}
	if len(trends) != 2 {
		t.Fatalf("trends len = %d, want 2", len(trends))
	}
	if trends[0].Period != "2024-01" || trends[1].Period != "2024-02" {
		t.Fatalf("trend periods = [%s %s], want [2024-01 2024-02]", trends[0].Period, trends[1].Period)
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	db := newServerFixtureDB(t)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})

	return NewServer(db)
}

func newCallToolRequest(name string, arguments map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: arguments,
		},
	}
}

func assertSingleTextContains(t *testing.T, result *mcp.CallToolResult, needle string) {
	t.Helper()

	if len(result.Content) != 1 {
		t.Fatalf("content len = %d, want 1", len(result.Content))
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want mcp.TextContent", result.Content[0])
	}
	if !contains(text.Text, needle) {
		t.Fatalf("content text %q does not contain %q", text.Text, needle)
	}
}

func contains(text, needle string) bool {
	return len(needle) == 0 || (len(text) >= len(needle) && index(text, needle) >= 0)
}

func index(text, needle string) int {
	for i := 0; i+len(needle) <= len(text); i++ {
		if text[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

func newServerFixtureDB(t *testing.T) *database.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "server-fixture.sqlite")
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open fixture sqlite: %v", err)
	}
	defer conn.Close()

	mustExecServerSQL(t, conn, `
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
	mustExecServerSQL(t, conn, `
		CREATE TABLE ZCATEGORYASSIGMENT (
			ZTRANSACTION INTEGER,
			ZCATEGORY INTEGER
		);
	`)

	mustExecServerSQL(t, conn, `
		INSERT INTO ZSYNCOBJECT (Z_PK, Z_ENT, ZNAME, ZBALLANCE, ZOPENINGBALANCE, ZCURRENCYNAME, ZTYPE)
		VALUES (1, 10, 'Checking', 0, 1000, 'USD', 'bank');
	`)
	mustExecServerSQL(t, conn, `
		INSERT INTO ZSYNCOBJECT (Z_PK, Z_ENT, ZNAME2) VALUES
			(100, 19, 'Salary'),
			(101, 19, 'Rent'),
			(102, 19, 'Groceries');
	`)

	insertServerFixtureTransaction(t, conn, 1000, 37, 3000, "2024-01-15", "January salary", 1, 0, 100)
	insertServerFixtureTransaction(t, conn, 1001, 37, -1200, "2024-01-20", "Rent payment", 1, 0, 101)
	insertServerFixtureTransaction(t, conn, 1002, 37, 2500, "2024-02-05", "February salary", 1, 0, 100)
	insertServerFixtureTransaction(t, conn, 1003, 37, -300, "2024-02-10", "Groceries", 1, 0, 102)

	db, err := database.NewDB(path)
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	return db
}

func insertServerFixtureTransaction(t *testing.T, conn *sql.DB, id, ent int64, amount float64, date, description string, account2, account int64, category int64) {
	t.Helper()

	mustExecServerSQL(t, conn, `
		INSERT INTO ZSYNCOBJECT (Z_PK, Z_ENT, ZAMOUNT1, ZDATE1, ZDESC2, ZACCOUNT2, ZACCOUNT)
		VALUES (?, ?, ?, ?, ?, ?, ?);
	`, id, ent, amount, serverCoreDataSeconds(t, date), description, account2, account)

	mustExecServerSQL(t, conn, `
		INSERT INTO ZCATEGORYASSIGMENT (ZTRANSACTION, ZCATEGORY)
		VALUES (?, ?);
	`, id, category)
}

func serverCoreDataSeconds(t *testing.T, date string) float64 {
	t.Helper()

	ts, err := time.Parse("2006-01-02", date)
	if err != nil {
		t.Fatalf("parse date %q: %v", date, err)
	}
	coreDataEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	return ts.Sub(coreDataEpoch).Seconds()
}

func mustExecServerSQL(t *testing.T, conn *sql.DB, query string, args ...any) {
	t.Helper()

	if _, err := conn.Exec(query, args...); err != nil {
		t.Fatalf("exec query failed: %v\nquery:\n%s", err, query)
	}
}
