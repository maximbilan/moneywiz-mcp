package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleAnalyzeSpendingTrends(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	groupBy := request.GetString("group_by", "month")
	months := request.GetInt("months", 0)

	if groupBy != "month" && groupBy != "year" {
		groupBy = "month"
	}

	trends, err := s.db.AnalyzeSpendingTrends(groupBy, months)
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

	jsonData, err := json.MarshalIndent(trends, "", "  ")
	if err != nil {
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

	trends, err := s.db.AnalyzeIncomeTrends(groupBy, months)
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

	jsonData, err := json.MarshalIndent(trends, "", "  ")
	if err != nil {
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

	analysis, err := s.db.AnalyzeSavings(months)
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

	jsonData, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
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
