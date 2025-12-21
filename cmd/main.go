package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/moneywiz-mcp/internal/database"
	"github.com/moneywiz-mcp/internal/server"
)

func main() {
	// Parse command line arguments
	dbPath := flag.String("db", "", "Path to MoneyWiz database folder (e.g., iMoneyWiz-Data-Backup-2025_12_21-17_23)")
	flag.Parse()

	// Default database path if not provided
	if *dbPath == "" {
		// Try to find the database in the current directory or parent directories
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}

		// Check common locations
		possiblePaths := []string{
			filepath.Join(wd, "iMoneyWiz-Data-Backup-2025_12_21-17_23", "ipadMoneyWiz.sqlite"),
			filepath.Join(wd, "..", "iMoneyWiz-Data-Backup-2025_12_21-17_23", "ipadMoneyWiz.sqlite"),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				*dbPath = path
				break
			}
		}

		if *dbPath == "" {
			log.Fatalf("Database path not provided and not found in common locations. Use -db flag to specify the path to ipadMoneyWiz.sqlite")
		}
	} else {
		// If a folder path is provided, append the database filename
		if info, err := os.Stat(*dbPath); err == nil && info.IsDir() {
			*dbPath = filepath.Join(*dbPath, "ipadMoneyWiz.sqlite")
		}
	}

	// Initialize database connection
	db, err := database.NewDB(*dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create MCP server
	mcpServer := mcpserver.NewMCPServer("moneywiz-mcp", "1.0.0")

	// Create our server instance and register handlers
	srv := server.NewServer(db)
	srv.RegisterHandlers(mcpServer)

	// Start the stdio server
	log.Println("Starting MoneyWiz MCP server...")
	if err := mcpserver.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
