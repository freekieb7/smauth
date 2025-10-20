package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	ID        string
	Name      string
	UpSQL     string
	DownSQL   string
	AppliedAt *time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db            *Database
	migrationsDir string
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *Database, migrationsDir string) *Migrator {
	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// InitMigrationTable creates the migrations tracking table
func (m *Migrator) InitMigrationTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS tbl_migration (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_migration_applied_at ON tbl_migration(applied_at);
	`

	_, err := m.db.Conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	return nil
}

// CreateMigration generates up and down migration files
func (m *Migrator) CreateMigration(name string) error {
	// Generate timestamp for unique migration ID
	timestamp := time.Now().Format("20060102150405")
	migrationID := fmt.Sprintf("%s_%s", timestamp, name)

	// Ensure migrations directory exists
	if err := os.MkdirAll(m.migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Create up migration file
	upFile := filepath.Join(m.migrationsDir, fmt.Sprintf("%s_up.sql", migrationID))
	upContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: %s

-- Add your up migration SQL here

`, migrationID, time.Now().Format(time.RFC3339), name)

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Create down migration file
	downFile := filepath.Join(m.migrationsDir, fmt.Sprintf("%s_down.sql", migrationID))
	downContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: %s (rollback)

-- Add your down migration SQL here

`, migrationID, time.Now().Format(time.RFC3339), name)

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("Created migration files:\n")
	fmt.Printf("  Up:   %s\n", upFile)
	fmt.Printf("  Down: %s\n", downFile)

	return nil
}

// GetPendingMigrations returns migrations that haven't been applied
func (m *Migrator) GetPendingMigrations(ctx context.Context) ([]Migration, error) {
	// Get applied migrations from database
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	// Get all migration files
	allMigrations, err := m.getAllMigrations()
	if err != nil {
		return nil, err
	}

	// Filter out applied migrations
	var pending []Migration
	for _, migration := range allMigrations {
		if _, applied := appliedMigrations[migration.ID]; !applied {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// GetAppliedMigrations returns migrations that have been applied (for rollback)
func (m *Migrator) GetAppliedMigrations(ctx context.Context) ([]Migration, error) {
	appliedMap, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	allMigrations, err := m.getAllMigrations()
	if err != nil {
		return nil, err
	}

	var applied []Migration
	for _, migration := range allMigrations {
		if appliedTime, exists := appliedMap[migration.ID]; exists {
			migration.AppliedAt = &appliedTime
			applied = append(applied, migration)
		}
	}

	// Sort by applied time (most recent first for rollback)
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].AppliedAt.After(*applied[j].AppliedAt)
	})

	return applied, nil
}

// Up runs pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	pending, err := m.GetPendingMigrations(ctx)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		fmt.Println("No pending migrations to run")
		return nil
	}

	fmt.Printf("Running %d pending migration(s):\n", len(pending))

	for _, migration := range pending {
		fmt.Printf("Applying migration: %s\n", migration.ID)

		// Start transaction
		tx, err := m.db.Conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// Execute up migration
		if _, err := tx.Exec(ctx, migration.UpSQL); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", migration.ID, err)
		}

		// Record migration as applied
		if _, err := tx.Exec(ctx,
			"INSERT INTO tbl_migration (id, name) VALUES ($1, $2)",
			migration.ID, migration.Name); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.ID, err)
		}

		fmt.Printf("✓ Applied migration: %s\n", migration.ID)
	}

	fmt.Println("All migrations applied successfully")
	return nil
}

// Down rolls back the most recent migration
func (m *Migrator) Down(ctx context.Context) error {
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		fmt.Println("No migrations to rollback")
		return nil
	}

	// Get the most recent migration
	migration := applied[0]

	fmt.Printf("Rolling back migration: %s\n", migration.ID)

	// Start transaction
	tx, err := m.db.Conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Execute down migration
	if _, err := tx.Exec(ctx, migration.DownSQL); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to execute rollback for %s: %w", migration.ID, err)
	}

	// Remove migration record
	if _, err := tx.Exec(ctx,
		"DELETE FROM tbl_migration WHERE id = $1", migration.ID); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to remove migration record %s: %w", migration.ID, err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit rollback %s: %w", migration.ID, err)
	}

	fmt.Printf("✓ Rolled back migration: %s\n", migration.ID)
	return nil
}

// getAppliedMigrations returns a map of applied migration IDs to their applied times
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[string]time.Time, error) {
	query := "SELECT id, applied_at FROM tbl_migration ORDER BY applied_at"
	rows, err := m.db.Conn.Query(ctx, query)
	if err != nil {
		// If table doesn't exist, return empty map
		if strings.Contains(err.Error(), "does not exist") {
			return make(map[string]time.Time), nil
		}
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]time.Time)
	for rows.Next() {
		var id string
		var appliedAt time.Time
		if err := rows.Scan(&id, &appliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}
		applied[id] = appliedAt
	}

	return applied, nil
}

// getAllMigrations reads all migration files from the migrations directory
func (m *Migrator) getAllMigrations() ([]Migration, error) {
	files, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Migration{}, nil
		}
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrationMap := make(map[string]*Migration)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if !strings.HasSuffix(filename, ".sql") {
			continue
		}

		// Parse filename: {timestamp}_{name}_{up|down}.sql
		parts := strings.Split(filename, "_")
		if len(parts) < 3 {
			continue
		}

		migrationID := strings.Join(parts[:len(parts)-1], "_")
		direction := strings.TrimSuffix(parts[len(parts)-1], ".sql")

		if direction != "up" && direction != "down" {
			continue
		}

		// Read file content
		content, err := os.ReadFile(filepath.Join(m.migrationsDir, filename))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Get or create migration entry
		if migrationMap[migrationID] == nil {
			// Extract name from migration ID (remove timestamp prefix)
			nameParts := strings.Split(migrationID, "_")
			var name string
			if len(nameParts) > 1 {
				name = strings.Join(nameParts[1:], "_")
			} else {
				name = migrationID
			}

			migrationMap[migrationID] = &Migration{
				ID:   migrationID,
				Name: name,
			}
		}

		// Set SQL content based on direction
		if direction == "up" {
			migrationMap[migrationID].UpSQL = string(content)
		} else {
			migrationMap[migrationID].DownSQL = string(content)
		}
	}

	// Convert map to slice and sort by ID (timestamp)
	var migrations []Migration
	for _, migration := range migrationMap {
		// Only include migrations that have both up and down files
		if migration.UpSQL != "" && migration.DownSQL != "" {
			migrations = append(migrations, *migration)
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}
