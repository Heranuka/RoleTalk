// Package main provides the entry point for the Backforge server.
//
// It initializes the application, sets up signal handling for graceful shutdown,
// and runs the HTTP server.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"go-backend/internal/app"
	"go-backend/internal/logger"
)

func main() {
	// Create a context that is canceled on SIGINT or SIGTERM signals.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize the application with all dependencies.
	a, err := app.New(ctx)
	if err != nil {
		log.Printf("failed to create app: %v", err)
		return
	}

	// Flush any buffered log entries before the application exits.
	defer logger.Sync(a.Logger)

	// Run the application (HTTP server) and block until it exits.
	if err := a.Run(ctx); err != nil {
		a.Logger.Fatalf("app stopped with error: %v", err)
	}
}
