package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleListTransactions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	accountID, limit := normalizeTransactionParams(
		request.GetFloat("account_id", 0),
		request.GetInt("limit", defaultTransactionLimit),
	)

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
