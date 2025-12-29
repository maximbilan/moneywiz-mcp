package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleListAccounts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Println("💳 [list_accounts] Handler called - fetching all accounts from database...")
	
	accounts, err := s.db.GetAccounts()
	if err != nil {
		log.Printf("❌ [list_accounts] Database query failed: %v", err)
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

	log.Printf("✅ [list_accounts] Successfully retrieved %d accounts", len(accounts))

	jsonData, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		log.Printf("❌ [list_accounts] JSON marshaling failed: %v", err)
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

	log.Println("✅ [list_accounts] Request completed successfully")
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
		log.Printf("❌ [get_account_balance] Invalid request parameter: %v", err)
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
	log.Printf("💰 [get_account_balance] Handler called - fetching balance for account ID: %d", accountID)

	account, err := s.db.GetAccountBalance(accountID)
	if err != nil {
		log.Printf("❌ [get_account_balance] Database query failed for account %d: %v", accountID, err)
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

	log.Printf("✅ [get_account_balance] Successfully retrieved balance for account %d", accountID)

	jsonData, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		log.Printf("❌ [get_account_balance] JSON marshaling failed: %v", err)
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

	log.Println("✅ [get_account_balance] Request completed successfully")
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
