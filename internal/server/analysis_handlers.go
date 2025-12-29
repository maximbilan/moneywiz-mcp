package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleAnalyzeSpendingTrends(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupBy := request.GetString("group_by", "month")
	months := request.GetInt("months", 0)

	if groupBy != "month" && groupBy != "year" {
		groupBy = "month"
	}

	if months > 0 {
		log.Printf("📊 [analyze_spending_trends] Handler called - analyzing spending trends (group_by: %s, months: %d)", groupBy, months)
	} else {
		log.Printf("📊 [analyze_spending_trends] Handler called - analyzing spending trends (group_by: %s, all historical data)", groupBy)
	}

	trends, err := s.db.AnalyzeSpendingTrends(groupBy, months)
	if err != nil {
		log.Printf("❌ [analyze_spending_trends] Database query failed: %v", err)
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

	log.Printf("✅ [analyze_spending_trends] Successfully analyzed spending trends (%d periods)", len(trends))

	jsonData, err := json.MarshalIndent(trends, "", "  ")
	if err != nil {
		log.Printf("❌ [analyze_spending_trends] JSON marshaling failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling trends: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	log.Println("✅ [analyze_spending_trends] Request completed successfully")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		StructuredContent: map[string]interface{}{
			"trends":   trends,
			"group_by": groupBy,
			"months":   months,
		},
	}, nil
}

func (s *Server) handleAnalyzeIncomeTrends(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupBy := request.GetString("group_by", "month")
	months := request.GetInt("months", 0)

	if groupBy != "month" && groupBy != "year" {
		groupBy = "month"
	}

	if months > 0 {
		log.Printf("📈 [analyze_income_trends] Handler called - analyzing income trends (group_by: %s, months: %d)", groupBy, months)
	} else {
		log.Printf("📈 [analyze_income_trends] Handler called - analyzing income trends (group_by: %s, all historical data)", groupBy)
	}

	trends, err := s.db.AnalyzeIncomeTrends(groupBy, months)
	if err != nil {
		log.Printf("❌ [analyze_income_trends] Database query failed: %v", err)
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

	log.Printf("✅ [analyze_income_trends] Successfully analyzed income trends (%d periods)", len(trends))

	jsonData, err := json.MarshalIndent(trends, "", "  ")
	if err != nil {
		log.Printf("❌ [analyze_income_trends] JSON marshaling failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling trends: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	log.Println("✅ [analyze_income_trends] Request completed successfully")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		StructuredContent: map[string]interface{}{
			"trends":   trends,
			"group_by": groupBy,
			"months":   months,
		},
	}, nil
}

func (s *Server) handleGetSavingsRecommendations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	months := request.GetInt("months", 0)

	if months > 0 {
		log.Printf("💡 [get_savings_recommendations] Handler called - analyzing savings recommendations (months: %d)", months)
	} else {
		log.Println("💡 [get_savings_recommendations] Handler called - analyzing savings recommendations (all historical data)")
	}

	analysis, err := s.db.AnalyzeSavings(months)
	if err != nil {
		log.Printf("❌ [get_savings_recommendations] Database query failed: %v", err)
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

	log.Println("✅ [get_savings_recommendations] Successfully generated savings analysis")

	jsonData, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		log.Printf("❌ [get_savings_recommendations] JSON marshaling failed: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling analysis: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	log.Println("✅ [get_savings_recommendations] Request completed successfully")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
		StructuredContent: analysis,
	}, nil
}
