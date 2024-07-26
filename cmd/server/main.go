package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rezkam/TaxMan/internal/constants"
	"github.com/rezkam/TaxMan/internal/routes"
	"github.com/rezkam/TaxMan/store"
	"github.com/rezkam/TaxMan/taxservice"
)

const (
	// defaultHTTPPort is the default port for the HTTP server.
	defaultHTTPPort = "8080"
	// httpReadTimeout is the default read timeout for the HTTP server.
	httpReadTimeout = 15 * time.Second
	// httpWriteTimeout is the default write timeout for the HTTP server.
	httpWriteTimeout = 15 * time.Second
	// maxMunicipalityNameLength is the maximum length of a municipality name.
	maxMunicipalityNameLength = 100
)

func main() {
	// connectionString := os.Getenv("DATABASE_URL")
	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		slog.Error("failed to create database connection DATABASE_URL is not set")
		os.Exit(1)
	}
	postgresStore, err := store.NewPostgresStore(connectionString)
	if err != nil {
		slog.Error("failed to create postgres store", "error", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultHTTPPort
	}

	svc, err := taxservice.New(postgresStore, taxservice.Config{
		MaxMunicipalityNameLength: maxMunicipalityNameLength,
		MunicipalityURLPattern:    constants.MunicipalityURLPattern,
		DateURLPattern:            constants.DateURLPattern,
	})
	if err != nil {
		slog.Error("failed to create service", "error", err)
		os.Exit(1)
	}

	// Setup routes and start the server
	mux := http.NewServeMux()
	routes.SetupTaxRoutes(svc, mux)

	// Create an HTTP server with read and write timeouts
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
	}

	shutdownHandler := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown server", "error", err)
		}

		if err := postgresStore.Close(); err != nil {
			slog.Error("failed to close store", "error", err)
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		slog.Info("received shutdown signal")
		shutdownHandler()
		os.Exit(0)
	}()

	slog.Info("starting server", "port", port)
	if err := httpServer.ListenAndServe(); err != nil {
		slog.Error("server error", "error", err, "port", port)
		shutdownHandler()
		os.Exit(1)
	}
}
