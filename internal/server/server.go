package server

import (
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/moneywiz-mcp/internal/database"
)

type Server struct {
	db *database.DB
}

func NewServer(db *database.DB) *Server {
	return &Server{db: db}
}

func (s *Server) RegisterHandlers(mcpServer *mcpserver.MCPServer) {
	log.Println("🔧 Registering MCP tools...")

	// List accounts tool
	log.Println("  ✓ Registering tool: list_accounts")
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_accounts",
		Description: "List all MoneyWiz accounts with balances and explicit account currencies",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, s.handleListAccounts)

	// Get account balance tool
	log.Println("  ✓ Registering tool: get_account_balance")
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_account_balance",
		Description: "Get the balance for a specific account by ID",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account_id": map[string]any{
					"type":        "integer",
					"description": "The ID of the account",
				},
			},
			Required: []string{"account_id"},
		},
	}, s.handleGetAccountBalance)

	// List transactions tool
	log.Println("  ✓ Registering tool: list_transactions")
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_transactions",
		Description: "List recent transactions with account name, currency, category, and movement type; transfer-like rows are labeled explicitly",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account_id": map[string]any{
					"type":        "integer",
					"description": "Optional account ID to filter transactions. If not provided, returns all transactions",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of transactions to return (default: 50)",
					"default":     50,
				},
			},
		},
	}, s.handleListTransactions)

	// List categories tool
	log.Println("  ✓ Registering tool: list_categories")
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_categories",
		Description: "List all categories in MoneyWiz",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, s.handleListCategories)

	// Analyze spending trends tool
	log.Println("  ✓ Registering tool: analyze_spending_trends")
	mcpServer.AddTool(mcp.Tool{
		Name:        "analyze_spending_trends",
		Description: "Analyze spending trends by category and time period (month or year), including by_currency totals and excluding internal transfers/cash withdrawals",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"group_by": map[string]any{
					"type":        "string",
					"description": "Group by 'month' or 'year' (default: 'month')",
					"enum":        []string{"month", "year"},
					"default":     "month",
				},
				"months": map[string]any{
					"type":        "integer",
					"description": "Number of months to analyze (0 or omitted = all historical data)",
					"default":     0,
				},
			},
		},
	}, s.handleAnalyzeSpendingTrends)

	// Analyze income trends tool
	log.Println("  ✓ Registering tool: analyze_income_trends")
	mcpServer.AddTool(mcp.Tool{
		Name:        "analyze_income_trends",
		Description: "Analyze income trends by category and time period (month or year), including by_currency totals and excluding internal transfers/cash withdrawals",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"group_by": map[string]any{
					"type":        "string",
					"description": "Group by 'month' or 'year' (default: 'month')",
					"enum":        []string{"month", "year"},
					"default":     "month",
				},
				"months": map[string]any{
					"type":        "integer",
					"description": "Number of months to analyze (0 or omitted = all historical data)",
					"default":     0,
				},
			},
		},
	}, s.handleAnalyzeIncomeTrends)

	// Savings recommendations tool
	log.Println("  ✓ Registering tool: get_savings_recommendations")
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_savings_recommendations",
		Description: "Analyze income vs spending with per-currency breakdowns and mixed-currency warnings, then return savings recommendations",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"months": map[string]any{
					"type":        "integer",
					"description": "Number of months to analyze (0 or omitted = all historical data)",
					"default":     0,
				},
			},
		},
	}, s.handleGetSavingsRecommendations)

	// Calculate net worth tool
	log.Println("  ✓ Registering tool: calculate_net_worth")
	mcpServer.AddTool(mcp.Tool{
		Name:        "calculate_net_worth",
		Description: "Calculate total net worth from all accounts (assets minus liabilities)",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, s.handleCalculateNetWorth)

	// Get financial stats tool
	log.Println("  ✓ Registering tool: get_financial_stats")
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_financial_stats",
		Description: "Get comprehensive financial statistics with explicit currency context, per-currency breakdowns, and totals excluding internal transfers/cash withdrawals",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, s.handleGetFinancialStats)

	log.Println("✅ All 9 MCP tools registered successfully!")
}
