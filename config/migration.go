package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// RunMigrations executes all .sql files in the migrations directory in order
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Create migrations tracking table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename    VARCHAR(255) PRIMARY KEY,
			executed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		filename := filepath.Base(file)

		// Check if already executed
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)", filename).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", filename, err)
		}
		if exists {
			continue
		}

		// Read and execute migration
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", filename); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		log.Printf("Migration applied: %s\n", filename)
	}

	return nil
}
