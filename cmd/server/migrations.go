package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jayasri-stack/Greening-India-Assignment-jayasri/internal/db"
)

func runMigrations(database *db.Database) error {
	slog.Info("checking and running database migrations...")

	migrationsDir := "./migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "/app/migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			slog.Warn("migrations directory not found, skipping migrations")
			return nil
		}
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := database.Exec(createTableSQL); err != nil {
		slog.Warn("could not create schema_migrations table", "error", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		version := strings.TrimSuffix(file.Name(), ".up.sql")

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

		filePath := filepath.Join(migrationsDir, file.Name())
		sql, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
		}

		slog.Info("executing migration", "version", version)
		if _, err := database.Exec(string(sql)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		if _, err := database.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			slog.Warn("failed to record migration", "version", version, "error", err)
		}

		slog.Info("migration completed", "version", version)
	}

	slog.Info("all migrations completed successfully")
	return nil
}