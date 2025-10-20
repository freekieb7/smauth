package main

import (
	"context"
	"log/slog"
	"os"

	_ "github.com/a-h/templ"
	"github.com/freekieb7/smauth/internal/database"
	"github.com/freekieb7/smauth/internal/web"
)

func main() {
	if err := run(context.Background()); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	db := database.New()
	if err := db.Connect("postgres://postgres:password@localhost:5432/smauth"); err != nil {
		return err
	}
	defer db.Close()

	// Ping database
	if err := db.Ping(ctx); err != nil {
		return err
	}

	// Initialize handler
	handler := web.NewHandler(logger, db)

	// Initialize router
	router := web.NewRouter()

	// Routes
	router.GET("/users", handler.ShowUsers)
	router.POST("/users/create-user", handler.CreateUser)
	router.POST("/users/update-user", handler.UpdateUser)
	router.POST("/users/delete-user", handler.DeleteUser)
	router.GET("/clients", handler.ShowClients)
	router.POST("/clients/create-client", handler.CreateClient)
	router.POST("/clients/update-client", handler.UpdateClient)
	router.POST("/clients/delete-client", handler.DeleteClient)
	router.GET("/auth/authorize", web.HandleAuthorize(db))
	router.POST("/auth/token", web.HandleToken(db))

	// Start server
	server := web.NewServer(logger, router)
	if err := server.ListenAndServe(":8080"); err != nil {
		return err
	}

	return nil
}
