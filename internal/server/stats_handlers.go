package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleCalculateNetWorth(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Println("💎 [calculate_net_worth] Handler called - calculating net worth from all accounts...")
	
	netWorth, err := s.db.CalculateNetWorth()
	if err != nil {
		log.Printf("❌ [calculate_net_worth] Database query failed: %v", err)
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

	log.Println("✅ [calculate_net_worth] Successfully calculated net worth")

	jsonData, err := json.MarshalIndent(netWorth, "", "  ")
	if err != nil {
		log.Printf("❌ [calculate_net_worth] JSON marshaling failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling net worth: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	log.Println("✅ [calculate_net_worth] Request completed successfully")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		StructuredContent: netWorth,
	}, nil
}

func (s *Server) handleGetFinancialStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Println("📈 [get_financial_stats] Handler called - fetching comprehensive financial statistics...")
	
	stats, err := s.db.GetFinancialStats()
	if err != nil {
		log.Printf("❌ [get_financial_stats] Database query failed: %v", err)
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

	log.Println("✅ [get_financial_stats] Successfully retrieved financial statistics")

	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Printf("❌ [get_financial_stats] JSON marshaling failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling stats: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	log.Println("✅ [get_financial_stats] Request completed successfully")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		StructuredContent: stats,
	}, nil
}
