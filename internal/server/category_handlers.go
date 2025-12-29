package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleListCategories(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Println("🏷️  [list_categories] Handler called - fetching all categories from database...")
	
	categories, err := s.db.GetCategories()
	if err != nil {
		log.Printf("❌ [list_categories] Database query failed: %v", err)
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

	log.Printf("✅ [list_categories] Successfully retrieved %d categories", len(categories))

	jsonData, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		log.Printf("❌ [list_categories] JSON marshaling failed: %v", err)
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

	log.Println("✅ [list_categories] Request completed successfully")
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
