package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"jayasri-stack/Greening-India-Assignment-jayasri/internal/db"
)

// runMigrations executes all pending migrations
func runMigrations(database *db.Database) error {
	slog.Info("checking and running database migrations...")

	// Get migration files directory
	migrationsDir := "./migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try alternate path for Docker
		migrationsDir = "/app/migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			slog.Warn("migrations directory not found, skipping migrations")
			return nil
		}
	}

	// Read migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Create migrations table if it doesn't exist
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := database.Exec(createTableSQL); err != nil {
		slog.Warn("could not create schema_migrations table", "error", err)
	}

	// Execute all .up.sql migrations that haven't been run
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		version := strings.TrimSuffix(file.Name(), ".up.sql")

		// Check if migration already ran
		var exists bool
		row := database.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version)
		if err := row.Scan(&exists); err != nil {
			slog.Warn("could not check migration status", "version", version, "error", err)
			continue
		}

		if exists {
			slog.Info("migration already applied", "version", version)
			continue
		}

		// Read and execute migration file
		filePath := filepath.Join(migrationsDir, file.Name())
		sql, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
		}

		slog.Info("executing migration", "version", version)
		if _, err := database.Exec(string(sql)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		// Record migration
		if _, err := database.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			slog.Warn("failed to record migration in schema_migrations table", "version", version, "error", err)
		}

		slog.Info("migration completed", "version", version)
	}

	slog.Info("all migrations completed successfully")
	return nil
}

// maskDSN masks sensitive information in DSN for logging
func maskDSN(dsn string) string {
	if strings.Contains(dsn, "@") {
		parts := strings.Split(dsn, "@")
		return "****@" + parts[1]
	}
	return dsn
}
