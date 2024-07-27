package main

import (
	"context"
	"errors"
	"go.uber.org/fx/fxevent"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/rezkam/TaxMan/internal/constants"
	"github.com/rezkam/TaxMan/internal/routes"
	"github.com/rezkam/TaxMan/store"
	"github.com/rezkam/TaxMan/taxservice"
	"go.uber.org/fx"
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
	// databaseURLKey is the key for the DATABASE_URL environment variable.
	databaseURLKey = "DATABASE_URL"
)

var (
	// defaultTaxRate is the default tax rate to use if no specific rate is found for a municipality.
	defaultTaxRate = 0.5
)

func main() {
	// Create a new application
	app := fx.New(
		fx.WithLogger(WithSlogLogger),
		fx.Provide(
			NewJSONLogger,
			NewPostgresStore,
			NewTaxService,
			NewHTTPServer,
			NewServeMux,
		),
		fx.Invoke(func(s *http.Server) {}),
	)

	// run the application
	app.Run()

}

func NewJSONLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, nil))
}

func WithSlogLogger(log *slog.Logger) fxevent.Logger {
	return &fxevent.SlogLogger{Logger: log}
}

func NewPostgresStore(lc fx.Lifecycle, log *slog.Logger) (*store.PostgresStore, error) {
	connectionString := os.Getenv(databaseURLKey)
	if connectionString == "" {
		log.Error("Database URL not set", "databaseURLKey", databaseURLKey)
		return nil, errors.New("database URL not set")
	}
	postgresStore, err := store.NewPostgresStore(connectionString, log)
	if err != nil {
		log.Error("failed to create postgres store", "error", err)
		return nil, err
	}
	// Add a hook to schedule the closing of the store when the application is stopped
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			log.Info("closing postgres store")
			return postgresStore.Close()
		},
	})

	return postgresStore, nil
}

func NewTaxService(store *store.PostgresStore, log *slog.Logger) (*taxservice.Service, error) {
	svc, err := taxservice.New(store, taxservice.Config{
		MaxMunicipalityNameLength: maxMunicipalityNameLength,
		MunicipalityURLPattern:    constants.MunicipalityURLPattern,
		DateURLPattern:            constants.DateURLPattern,
		DefaultTaxRate:            &defaultTaxRate,
	})
	if err != nil {
		log.Error("failed to create tax service", "error", err)
		return nil, err
	}
	return svc, nil
}

func NewHTTPServer(lc fx.Lifecycle, mux *http.ServeMux, log *slog.Logger) *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultHTTPPort
	}

	// Create an HTTP server with read and write timeouts
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
	}

	lc.Append(fx.Hook{
		// Add a hook to schedule the start the server after dependencies are available
		OnStart: func(context.Context) error {
			log.Info("Starting HTTP server", "addr", httpServer.Addr)
			go func() {
				if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
					log.Error("HTTP server stopped", "error", err)
				}
			}()
			return nil
		},
		// Add a hook to schedule the shutdown of the server when the application is stopped
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping HTTP server")
			return httpServer.Shutdown(ctx)
		},
	})
	return httpServer
}

func NewServeMux(taxService *taxservice.Service) *http.ServeMux {
	mux := http.NewServeMux()
	routes.SetupTaxRoutes(taxService, mux)
	return mux
}
