package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestNew_InMemory(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: "",
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer database.Close()

	// Verify connection works
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}

func TestNew_WithMigrations(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: "./migrations",
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create database with migrations: %v", err)
	}
	defer database.Close()

	// Verify migrations table exists
	ctx := context.Background()
	rows, err := database.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("Failed to query schema_migrations: %v", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("Failed to scan version: %v", err)
		}
		versions = append(versions, v)
	}

	if len(versions) == 0 {
		t.Fatal("Expected at least one migration to be applied")
	}

	// Verify users table was created (from 001_init.sql)
	var count int
	err = database.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check for users table: %v", err)
	}
	if count != 1 {
		t.Fatal("Expected users table to exist after migrations")
	}
}

func TestNew_WithFileDB(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data", "test.db")

	cfg := Config{
		DBPath:         dbPath,
		MigrationsPath: "",
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create file database: %v", err)
	}
	defer database.Close()

	// Verify the file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Expected database file to be created")
	}
}

func TestNew_MigrationsIdempotent(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: "./migrations",
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("First New failed: %v", err)
	}

	// Count migrations after first run
	ctx := context.Background()
	var count1 int
	err = database.QueryRowContext(ctx, "SELECT count(*) FROM schema_migrations").Scan(&count1)
	if err != nil {
		t.Fatalf("Failed to count migrations: %v", err)
	}

	// Run migrations again on same DB
	err = database.runMigrations(ctx, cfg.MigrationsPath)
	if err != nil {
		t.Fatalf("Second migration run failed: %v", err)
	}

	// Count should be the same
	var count2 int
	err = database.QueryRowContext(ctx, "SELECT count(*) FROM schema_migrations").Scan(&count2)
	if err != nil {
		t.Fatalf("Failed to count migrations after second run: %v", err)
	}

	if count1 != count2 {
		t.Errorf("Migration count changed: %d -> %d (expected idempotent)", count1, count2)
	}

	database.Close()
}

func TestNew_NonExistentMigrationsDir(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: "/nonexistent/path/migrations",
	}

	// Should succeed with a warning (non-existent dir is skipped)
	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Expected success with non-existent migrations dir: %v", err)
	}
	defer database.Close()
}

func TestClose(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: "",
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = database.Close()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}

	// Verify connection is closed by trying to ping
	ctx := context.Background()
	err = database.PingContext(ctx)
	if err == nil {
		t.Fatal("Expected error after closing database")
	}
}

func TestHealthCheck(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: "",
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	err = database.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck failed on open database: %v", err)
	}
}

func TestHealthCheck_ClosedDB(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: "",
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	database.Close()

	ctx := context.Background()
	err = database.HealthCheck(ctx)
	if err == nil {
		t.Fatal("Expected HealthCheck to fail on closed database")
	}
}

func TestNew_PragmasApplied(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: "",
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// Check foreign keys are enabled
	var fkEnabled int
	err = database.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatalf("Failed to check foreign_keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("Expected foreign_keys=1, got %d", fkEnabled)
	}

	// Check journal mode (WAL for file-based, memory for :memory:)
	var journalMode string
	err = database.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to check journal_mode: %v", err)
	}
	if journalMode != "wal" && journalMode != "memory" {
		t.Errorf("Expected journal_mode=wal or memory, got %s", journalMode)
	}
}

func TestRunMigrations_AppliesInOrder(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: "./migrations",
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// Get all applied migrations in order
	rows, err := database.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("Failed to query migrations: %v", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("Failed to scan: %v", err)
		}
		versions = append(versions, v)
	}

	// Verify migrations are in sorted order
	for i := 1; i < len(versions); i++ {
		if versions[i] < versions[i-1] {
			t.Errorf("Migrations not in order: %s came after %s", versions[i], versions[i-1])
		}
	}

	// Verify first migration starts with 001
	if len(versions) > 0 && versions[0][:3] != "001" {
		t.Errorf("Expected first migration to start with '001', got %s", versions[0])
	}
}

func TestRunMigrations_WithTempDir(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create a temp dir with a single migration
	tmpDir := t.TempDir()
	migration := `CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);`
	err := os.WriteFile(filepath.Join(tmpDir, "001_test.sql"), []byte(migration), 0644)
	if err != nil {
		t.Fatalf("Failed to write migration: %v", err)
	}

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: tmpDir,
	}

	database, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer database.Close()

	// Verify the test_table was created
	ctx := context.Background()
	var count int
	err = database.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check for test_table: %v", err)
	}
	if count != 1 {
		t.Fatal("Expected test_table to exist")
	}

	// Verify migration was recorded
	var version string
	err = database.QueryRowContext(ctx, "SELECT version FROM schema_migrations WHERE version = '001_test'").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to find migration record: %v", err)
	}
	if version != "001_test" {
		t.Errorf("Expected version '001_test', got %s", version)
	}
}

func TestRunMigrations_InvalidSQL(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "001_bad.sql"), []byte("THIS IS NOT VALID SQL;"), 0644)
	if err != nil {
		t.Fatalf("Failed to write migration: %v", err)
	}

	cfg := Config{
		DBPath:         ":memory:",
		MigrationsPath: tmpDir,
	}

	_, err = New(cfg, logger)
	if err == nil {
		t.Fatal("Expected error from invalid migration SQL")
	}
}
