package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

// DB wraps the database connection
type DB struct {
	*sql.DB
	logger *zap.Logger
}

// Config holds database configuration
type Config struct {
	DBPath         string
	MigrationsPath string
}

// New creates a new database connection and runs migrations
func New(cfg Config, logger *zap.Logger) (*DB, error) {
	// Ensure data directory exists
	dataDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database connection
	sqlDB, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(1) // SQLite supports only one writer
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Enable foreign keys and WAL mode for better concurrency
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, pragma := range pragmas {
		if _, err := sqlDB.ExecContext(ctx, pragma); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}

	// Verify connection
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{DB: sqlDB, logger: logger}

	// Run migrations
	if cfg.MigrationsPath != "" {
		if err := db.runMigrations(ctx, cfg.MigrationsPath); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	logger.Info("Database initialized", zap.String("path", cfg.DBPath))
	return db, nil
}

// runMigrations executes all pending SQL migration files
func (db *DB) runMigrations(ctx context.Context, migrationsPath string) error {
	// Create migrations table if it doesn't exist
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedMigrations := make(map[string]bool)
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating migrations: %w", err)
	}

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		db.logger.Warn("Migrations directory does not exist, skipping migrations",
			zap.String("path", migrationsPath))
		return nil
	}

	// Read migration files
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort .sql files
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply pending migrations
	for _, filename := range migrationFiles {
		version := strings.TrimSuffix(filename, ".sql")

		if appliedMigrations[version] {
			continue
		}

		db.logger.Info("Applying migration", zap.String("file", filename))

		// Read migration file
		content, err := os.ReadFile(filepath.Join(migrationsPath, filename))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute migration in a transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
		}

		// Execute migration SQL
		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Record migration
		_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		db.logger.Info("Migration applied successfully", zap.String("file", filename))
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// HealthCheck verifies database connectivity
func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}
