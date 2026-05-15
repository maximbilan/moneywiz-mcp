package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNormalizeDBPathAcceptsFolderAndFile(t *testing.T) {
	base := t.TempDir()
	exportDir := filepath.Join(base, "iMoneyWiz-Data-Backup-2026_05_15-10_26")
	dbPath := filepath.Join(exportDir, defaultSQLiteName)

	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		t.Fatalf("mkdir export dir: %v", err)
	}
	if err := os.WriteFile(dbPath, []byte("db"), 0o644); err != nil {
		t.Fatalf("write db file: %v", err)
	}

	gotFromDir, err := normalizeDBPath(exportDir)
	if err != nil {
		t.Fatalf("normalize folder path: %v", err)
	}
	if gotFromDir != dbPath {
		t.Fatalf("normalize folder path = %q, want %q", gotFromDir, dbPath)
	}

	gotFromFile, err := normalizeDBPath(dbPath)
	if err != nil {
		t.Fatalf("normalize file path: %v", err)
	}
	if gotFromFile != dbPath {
		t.Fatalf("normalize file path = %q, want %q", gotFromFile, dbPath)
	}
}

func TestResolveDBPathPrefersExplicitArgument(t *testing.T) {
	env := setupResolutionEnv(t)

	argExport := mustCreateExportDB(t, env.baseDir, "iMoneyWiz-Data-Backup-2026_05_15-10_26", time.Now())
	envExport := mustCreateExportDB(t, env.homeDir, "iMoneyWiz-Data-Backup-2026_05_16-10_26", time.Now().Add(time.Hour))
	t.Setenv("MONEYWIZ_DB_PATH", envExport)

	got, err := resolveDBPath(argExport)
	if err != nil {
		t.Fatalf("resolve explicit arg: %v", err)
	}
	if got != argExport {
		t.Fatalf("resolve explicit arg = %q, want %q", got, argExport)
	}
}

func TestResolveDBPathUsesEnvWhenArgumentMissing(t *testing.T) {
	env := setupResolutionEnv(t)

	envExport := mustCreateExportDB(t, env.homeDir, "iMoneyWiz-Data-Backup-2026_05_15-10_26", time.Now())
	t.Setenv("MONEYWIZ_DB_PATH", envExport)

	got, err := resolveDBPath("")
	if err != nil {
		t.Fatalf("resolve env path: %v", err)
	}
	if got != envExport {
		t.Fatalf("resolve env path = %q, want %q", got, envExport)
	}
}

func TestResolveDBPathUsesCanonicalBeforeExportDiscovery(t *testing.T) {
	env := setupResolutionEnv(t)

	canonicalDir := filepath.Join(env.homeDir, ".moneywiz-mcp")
	canonicalPath := filepath.Join(canonicalDir, defaultSQLiteName)
	if err := os.MkdirAll(canonicalDir, 0o755); err != nil {
		t.Fatalf("mkdir canonical dir: %v", err)
	}
	if err := os.WriteFile(canonicalPath, []byte("canonical"), 0o644); err != nil {
		t.Fatalf("write canonical db: %v", err)
	}

	mustCreateExportDB(t, env.baseDir, "iMoneyWiz-Data-Backup-2026_05_15-10_26", time.Now().Add(time.Hour))

	got, err := resolveDBPath("")
	if err != nil {
		t.Fatalf("resolve canonical path: %v", err)
	}
	if got != canonicalPath {
		t.Fatalf("resolve canonical path = %q, want %q", got, canonicalPath)
	}
}

func TestResolveDBPathLatestSelectsNewestExport(t *testing.T) {
	env := setupResolutionEnv(t)

	older := mustCreateExportDB(t, env.baseDir, "iMoneyWiz-Data-Backup-2026_05_15-10_26", time.Now().Add(-time.Hour))
	newer := mustCreateExportDB(t, env.baseDir, "iMoneyWiz-Data-Backup-2026_05_16-10_26", time.Now())

	got, err := resolveDBPath(latestSentinel)
	if err != nil {
		t.Fatalf("resolve latest sentinel: %v", err)
	}
	if canonicalTestPath(t, got) != canonicalTestPath(t, newer) {
		t.Fatalf("resolve latest sentinel = %q, want newest %q (older was %q)", got, newer, older)
	}
}

func TestResolveDBPathReturnsHelpfulErrorWhenNothingExists(t *testing.T) {
	_ = setupResolutionEnv(t)

	_, err := resolveDBPath("")
	if err == nil {
		t.Fatal("resolve empty path unexpectedly succeeded")
	}
	want := "database not found"
	if got := err.Error(); len(got) < len(want) || got[:len(want)] != want {
		t.Fatalf("resolve empty path error = %q, want prefix %q", got, want)
	}
}

type resolutionEnv struct {
	baseDir string
	homeDir string
}

func setupResolutionEnv(t *testing.T) resolutionEnv {
	t.Helper()

	baseDir := t.TempDir()
	workspaceDir := filepath.Join(baseDir, "workspace")
	homeDir := filepath.Join(baseDir, "home")

	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}

	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workspaceDir); err != nil {
		t.Fatalf("chdir workspace: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prevWD); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})

	t.Setenv("HOME", homeDir)
	t.Setenv("MONEYWIZ_DB_PATH", "")

	return resolutionEnv{
		baseDir: baseDir,
		homeDir: homeDir,
	}
}

func mustCreateExportDB(t *testing.T, parentDir, exportName string, modTime time.Time) string {
	t.Helper()

	exportDir := filepath.Join(parentDir, exportName)
	dbPath := filepath.Join(exportDir, defaultSQLiteName)

	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		t.Fatalf("mkdir export dir: %v", err)
	}
	if err := os.WriteFile(dbPath, []byte(exportName), 0o644); err != nil {
		t.Fatalf("write export db: %v", err)
	}
	if err := os.Chtimes(dbPath, modTime, modTime); err != nil {
		t.Fatalf("chtimes db: %v", err)
	}

	return dbPath
}

func canonicalTestPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved
	}
	return filepath.Clean(path)
}
