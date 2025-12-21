package server

import (
	"context"
	"encoding/json"
	"fmt"

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
	// List accounts tool
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_accounts",
		Description: "List all accounts in MoneyWiz with their balances and currencies",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, s.handleListAccounts)

	// Get account balance tool
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
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_transactions",
		Description: "List recent transactions, optionally filtered by account ID",
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
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_categories",
		Description: "List all categories in MoneyWiz",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, s.handleListCategories)

}

func (s *Server) handleListAccounts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	accounts, err := s.db.GetAccounts()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling accounts: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		StructuredContent: map[string]interface{}{
			"accounts": accounts,
		},
	}, nil
}

func (s *Server) handleGetAccountBalance(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	accountIDFloat, err := request.RequireFloat("account_id")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
			IsError: true,
		}, nil
	}
	accountID := int64(accountIDFloat)

	account, err := s.db.GetAccountBalance(accountID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling account: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		StructuredContent: account,
	}, nil
}

func (s *Server) handleListTransactions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	accountIDFloat := request.GetFloat("account_id", 0)
	accountID := int64(accountIDFloat)
	limit := request.GetInt("limit", 50)

	if limit == 0 {
		limit = 50
	}

	transactions, err := s.db.GetTransactions(accountID, limit)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling transactions: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		StructuredContent: map[string]interface{}{
			"transactions": transactions,
		},
	}, nil
}

func (s *Server) handleListCategories(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	categories, err := s.db.GetCategories()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling categories: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		StructuredContent: map[string]interface{}{
			"categories": categories,
		},
	}, nil
}

