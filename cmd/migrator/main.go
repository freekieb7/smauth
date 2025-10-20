package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/freekieb7/smauth/internal/database"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Migrations directory
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		// Default to migrations directory relative to project root
		migrationsDir = "database/migrations"
	}

	// Make migrations directory absolute
	if !filepath.IsAbs(migrationsDir) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		migrationsDir = filepath.Join(cwd, migrationsDir)
	}

	// Execute command
	switch command {
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Error: migration name is required")
			fmt.Println("Usage: migrator create <migration_name>")
			os.Exit(1)
		}
		migrationName := os.Args[2]
		// For create command, we don't need database connection
		migrator := database.NewMigrator(nil, migrationsDir)
		if err := migrator.CreateMigration(migrationName); err != nil {
			fmt.Printf("Error creating migration: %v\n", err)
			os.Exit(1)
		}

	case "up", "down", "status", "init":
		// For other commands, we need database connection
		// Database connection string from environment or default
		connString := os.Getenv("DATABASE_URL")
		if connString == "" {
			connString = "postgres://username:password@localhost:5432/dbname?sslmode=disable"
			fmt.Println("Warning: Using default connection string. Set DATABASE_URL environment variable.")
		}

		// Initialize database connection
		db := database.New()
		if err := db.Connect(connString); err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Initialize migrator
		migrator := database.NewMigrator(db, migrationsDir)
		ctx := context.Background()

		// Ensure migration table exists for all commands except create
		if err := migrator.InitMigrationTable(ctx); err != nil {
			fmt.Printf("Error initializing migration table: %v\n", err)
			os.Exit(1)
		}

		switch command {
		case "up":
			if err := migrator.Up(ctx); err != nil {
				fmt.Printf("Error running migrations: %v\n", err)
				os.Exit(1)
			}

		case "down":
			if err := migrator.Down(ctx); err != nil {
				fmt.Printf("Error rolling back migration: %v\n", err)
				os.Exit(1)
			}

		case "status":
			if err := showStatus(ctx, migrator); err != nil {
				fmt.Printf("Error showing status: %v\n", err)
				os.Exit(1)
			}

		case "init":
			fmt.Println("Migration table initialized successfully")
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Migration CLI for PostgreSQL")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrator <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create <name>    Create new up and down migration files")
	fmt.Println("  up               Run all pending migrations")
	fmt.Println("  down             Rollback the most recent migration")
	fmt.Println("  status           Show migration status")
	fmt.Println("  init             Initialize the migration tracking table")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  DATABASE_URL     PostgreSQL connection string")
	fmt.Println("  MIGRATIONS_DIR   Directory containing migration files (default: database/migrations)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  migrator create add_users_table")
	fmt.Println("  migrator up")
	fmt.Println("  migrator down")
	fmt.Println("  migrator status")
}

func showStatus(ctx context.Context, migrator *database.Migrator) error {
	pending, err := migrator.GetPendingMigrations(ctx)
	if err != nil {
		return err
	}

	applied, err := migrator.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Migration Status:")
	fmt.Println("================")
	fmt.Printf("Applied migrations: %d\n", len(applied))
	fmt.Printf("Pending migrations: %d\n", len(pending))
	fmt.Println()

	if len(applied) > 0 {
		fmt.Println("Applied migrations:")
		for _, migration := range applied {
			fmt.Printf("  ✓ %s (%s)\n", migration.ID, migration.AppliedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}

	if len(pending) > 0 {
		fmt.Println("Pending migrations:")
		for _, migration := range pending {
			fmt.Printf("  ○ %s\n", migration.ID)
		}
	} else {
		fmt.Println("No pending migrations")
	}

	return nil
}
