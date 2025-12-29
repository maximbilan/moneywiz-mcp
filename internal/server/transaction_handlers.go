package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleListTransactions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	accountIDFloat := request.GetFloat("account_id", 0)
	accountID := int64(accountIDFloat)
	limit := request.GetInt("limit", 50)

	if limit == 0 {
		limit = 50
	}

	if accountID > 0 {
		log.Printf("📝 [list_transactions] Handler called - fetching transactions for account ID: %d (limit: %d)", accountID, limit)
	} else {
		log.Printf("📝 [list_transactions] Handler called - fetching all transactions (limit: %d)", limit)
	}

	transactions, err := s.db.GetTransactions(accountID, limit)
	if err != nil {
		log.Printf("❌ [list_transactions] Database query failed: %v", err)
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

	log.Printf("✅ [list_transactions] Successfully retrieved %d transactions", len(transactions))

	jsonData, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		log.Printf("❌ [list_transactions] JSON marshaling failed: %v", err)
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

	log.Println("✅ [list_transactions] Request completed successfully")
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
