package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/moneywiz-mcp/internal/database"
	"github.com/moneywiz-mcp/internal/server"
)

const (
	defaultSQLiteName = "ipadMoneyWiz.sqlite"
	latestSentinel    = "latest"
)

type candidateDB struct {
	path    string
	modTime time.Time
}

func main() {
	// Parse command line arguments
	dbPath := flag.String("db", "", "Path to MoneyWiz DB (sqlite file or export folder). Use 'latest' to auto-pick newest export.")
	flag.Parse()

	resolvedDBPath, err := resolveDBPath(*dbPath)
	if err != nil {
		log.Fatalf("Failed to resolve database path: %v", err)
	}
	log.Printf("Using database: %s", resolvedDBPath)

	// Initialize database connection
	db, err := database.NewDB(resolvedDBPath)
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

func resolveDBPath(arg string) (string, error) {
	// Highest priority: explicit CLI argument.
	if strings.TrimSpace(arg) != "" {
		if arg == latestSentinel {
			return findLatestExportDBPath()
		}
		return normalizeDBPath(arg)
	}

	// Next priority: environment variable.
	if env := strings.TrimSpace(os.Getenv("MONEYWIZ_DB_PATH")); env != "" {
		if env == latestSentinel {
			return findLatestExportDBPath()
		}
		return normalizeDBPath(env)
	}

	// Next priority: canonical local managed path.
	if home, err := os.UserHomeDir(); err == nil {
		canonical := filepath.Join(home, ".moneywiz-mcp", defaultSQLiteName)
		if fileExists(canonical) {
			return canonical, nil
		}
	}

	// Fallback: best-effort latest export auto-discovery.
	latestPath, err := findLatestExportDBPath()
	if err == nil {
		return latestPath, nil
	}

	return "", errors.New(
		"database not found. Provide --db <path>, set MONEYWIZ_DB_PATH, or run ./scripts/import_db.sh /path/to/iMoneyWiz-Data-Backup-*",
	)
}

func normalizeDBPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %q: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %s", absPath)
	}
	if info.IsDir() {
		absPath = filepath.Join(absPath, defaultSQLiteName)
	}
	if !fileExists(absPath) {
		return "", fmt.Errorf("sqlite file not found: %s", absPath)
	}
	return absPath, nil
}

func findLatestExportDBPath() (string, error) {
	roots := discoveryRoots()
	candidates, err := discoverExportCandidates(roots)
	if err != nil {
		return "", err
	}
	if len(candidates) == 0 {
		return "", errors.New("no MoneyWiz export databases found in common locations")
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].modTime.After(candidates[j].modTime)
	})
	return candidates[0].path, nil
}

func discoveryRoots() []string {
	var roots []string
	wd, err := os.Getwd()
	if err == nil {
		roots = append(roots, wd, filepath.Dir(wd))
	}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, home)
	}
	return uniqCleanPaths(roots)
}

func discoverExportCandidates(roots []string) ([]candidateDB, error) {
	var candidates []candidateDB

	for _, root := range roots {
		pattern := filepath.Join(root, "iMoneyWiz-Data-Backup-*", defaultSQLiteName)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid search pattern %q: %w", pattern, err)
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			candidates = append(candidates, candidateDB{path: match, modTime: info.ModTime()})
		}
	}

	return candidates, nil
}

func uniqCleanPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	var out []string
	for _, p := range paths {
		clean := filepath.Clean(p)
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
