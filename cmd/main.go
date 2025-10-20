package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/freekieb7/smauth/internal"
	"github.com/freekieb7/smauth/internal/account"
	"github.com/freekieb7/smauth/internal/audit"
	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/http"
	"github.com/freekieb7/smauth/internal/oauth"
	"github.com/freekieb7/smauth/internal/openehr"
	"github.com/freekieb7/smauth/internal/openehr/aql"
	"github.com/freekieb7/smauth/internal/session"
	"github.com/freekieb7/smauth/internal/telemetry"
	"github.com/freekieb7/smauth/internal/web"

	_ "go.uber.org/automaxprocs"
)

// var (
// 	tracer = otel.Tracer(name)
// 	meter  = otel.Meter(name)
// 	logg   = otelslog.NewLogger(name)
// )

// func init() {
// 	os.Setenv("OTEL_SERVICE_NAME", "smauth")
// 	// os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.namespace=mynamespace,deployment.environment=development")
// 	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:4317")
// 	os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
// }

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		migrate(os.Args[1:])
		return
	}

	if err := run(context.Background()); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	// Create a context that can be cancelled for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	cfg := internal.NewConfig()

	// Initialize logger
	loggerProvider, err := telemetry.NewLoggerProvider(ctx, cfg.Server)
	if err != nil {
		return fmt.Errorf("failed to create logger provider: %w", err)
	}
	defer func() {
		if err := loggerProvider.Shutdown(context.Background()); err != nil {
			slog.Error("Failed to shutdown logger provider", slog.String("error", err.Error()))
		}
	}()

	logger := telemetry.NewLogger(cfg.Server.Name, loggerProvider)

	// Initialize tracer
	tracerProvider, err := telemetry.NewTracerProvider(ctx, cfg.Server)
	if err != nil {
		return fmt.Errorf("failed to create tracer provider: %w", err)
	}
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			slog.Error("Failed to shutdown tracer provider", slog.String("error", err.Error()))
		}
	}()

	tracer := telemetry.NewTracer(cfg.Server.Name, tracerProvider)

	// Initialize meter
	meterProvider, err := telemetry.NewMeterProvider(ctx, cfg.Server)
	if err != nil {
		return fmt.Errorf("failed to create meter provider: %w", err)
	}
	defer func() {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			slog.Error("Failed to shutdown meter provider", slog.String("error", err.Error()))
		}
	}()

	meter := telemetry.NewMeter(cfg.Server.Name, meterProvider)

	// Initialize database
	db := database.New()
	if err := db.ConnectWithRetry(cfg.Database.GetURL(), 5, 2*time.Second); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize stores and services
	aqlBuilder := aql.NewBuilder()
	sessionStore := session.NewStore(&logger, &db)
	accountStore := account.NewStore(&logger, &db)
	oauthStore := oauth.NewStore(cfg.OAuth, &logger, &db)
	oauthService := oauth.NewService(&logger, &db, &oauthStore, &accountStore)
	openehrStore := openehr.NewStore(&db)
	aqlService := aql.NewService(&logger, &db, &aqlBuilder)
	openehrValidator := openehr.NewValidator()
	auditStore := audit.NewStore(&logger, &db)
	auditService := audit.NewService(&logger, &db, &auditStore)
	clientService := account.NewClientService(&logger, &db)
	userService := account.NewUserService(&logger, &db)
	resourceServerService := oauth.NewResourceServerService(&logger, &db)

	// Initialize handler
	healthHandler := web.NewHealthHandler(&logger, &tracer, &meter)
	authHandler := web.NewAuthHandler(&logger, &sessionStore, &accountStore, &oauthStore, &oauthService, cfg)
	apiHandler := web.NewAPIHandler(cfg.Server, &logger, &sessionStore, &accountStore, &clientService, &oauthService)
	openehrHandler := web.NewOpenEHRHandler(&logger, &openehrStore, &aqlService, &openehrValidator)
	appHandler := web.NewAppHandler(&logger, &sessionStore, &accountStore, &auditStore, &auditService, &oauthStore, &clientService, &userService, &resourceServerService)

	// Ensure admin user exists
	if err := EnsureAdminUser(ctx, &logger, &accountStore, &cfg); err != nil {
		return fmt.Errorf("failed to ensure admin user: %w", err)
	}

	// Initialize router
	router := http.NewRouter()

	// Routes
	healthHandler.RegisterRoutes(&router)
	apiHandler.RegisterRoutes(&router)
	openehrHandler.RegisterRoutes(&router)
	authHandler.RegisterRoutes(&router)
	appHandler.RegisterRoutes(&router)

	// Start server
	server := http.NewServer(&logger, router)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting server", slog.String("addr", addr))

	// Channel to receive server errors
	serverErrChan := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(addr); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for either a signal or server error
	select {
	case <-stopChan:
		logger.Info("Received shutdown signal")
	case err := <-serverErrChan:
		logger.Error("Server failed to start", slog.String("error", err.Error()))
		return err
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	// Create a timeout context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logger.Info("Shutting down server...")

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown server gracefully", slog.String("error", err.Error()))
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	logger.Info("Server gracefully stopped")
	return nil
}

func EnsureAdminUser(ctx context.Context, logger *telemetry.Logger, accountStore *account.Store, config *internal.Config) error {
	exists, err := accountStore.AdminExists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for existing admin user: %w", err)
	}

	if exists {
		logger.Info("Admin user already exists, skipping creation")
		return nil
	}

	adminUser, err := accountStore.NewUser(config.Setup.AdminEmail, config.Setup.AdminPassword, account.RoleAdmin)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	passwordHash, err := account.HashPassword(config.Setup.AdminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}
	adminUser.PasswordHash = passwordHash

	if _, err := accountStore.SaveUser(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	logger.Info("Admin user created", slog.String("email", adminUser.Email))
	return nil
}

func migrate(args []string) {
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[1] // This should be the actual command (up/down/status/etc)

	// Migrations directory
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		// Default to migrations directory relative to project root
		migrationsDir = "internal/database/migrations"
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
		if len(args) < 4 { // Changed from os.Args to args
			fmt.Println("Error: migration name is required")
			fmt.Println("Usage: migrator create <migration_name>")
			os.Exit(1)
		}
		migrationName := args[3] // Changed from os.Args[2] to args[3]
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
		if err := db.ConnectWithRetry(connString, 5, 2*time.Second); err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Initialize migrator
		migrator := database.NewMigrator(&db, migrationsDir)
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
			fmt.Println("Migrations completed successfully")

		case "down":
			if err := migrator.Down(ctx); err != nil {
				fmt.Printf("Error rolling back migration: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Migration rolled back successfully")

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
