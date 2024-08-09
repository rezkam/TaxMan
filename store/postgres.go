package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
	"github.com/rezkam/TaxMan/model"
)

// Configuration for retry logic and timeouts
const (
	maxPingRetries     = 10
	pingInterval       = 2 * time.Second
	connectTimeout     = 30 * time.Second
	statementTimeout   = 10 * time.Second
	transactionTimeout = 60 * time.Second
	maxOpenConnections = 10
	maxIdleConnections = 5
)

type PostgresStore struct {
	db *sql.DB

	preparedStatements map[string]*sql.Stmt
}

// NewPostgresStore initializes and returns a new PostgresStore after ensuring the database is ready.
func NewPostgresStore(connStr string) (*PostgresStore, error) {
	// handle connection to the database
	connectionCtx, cancelConnection := context.WithTimeout(context.Background(), connectTimeout)
	defer cancelConnection()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool options
	db.SetMaxOpenConns(maxOpenConnections)
	db.SetMaxIdleConns(maxIdleConnections)

	// Ensure the database is ready
	if err := pingDBWithRetry(connectionCtx, db); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	createTableCtx, cancelCreateTable := context.WithTimeout(context.Background(), transactionTimeout) // Use transaction timeout for creating tables
	defer cancelCreateTable()

	// Create tables if they do not exist
	if err := createTables(createTableCtx, db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	prepareStmtCtx, cancelPrepareStmt := context.WithTimeout(context.Background(), statementTimeout)
	defer cancelPrepareStmt()

	// Prepare statements
	store := &PostgresStore{
		db:                 db,
		preparedStatements: make(map[string]*sql.Stmt),
	}
	if err := store.prepareStatements(prepareStmtCtx); err != nil {
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return store, nil
}

// Close closes the database connection.
func (s *PostgresStore) Close() error {
	for _, stmt := range s.preparedStatements {
		if err := stmt.Close(); err != nil {
			slog.Error("Error closing prepared statement", "error", err)
		}
	}
	return s.db.Close()
}

// pingDBWithRetry attempts to ping the database with retries and a timeout.
func pingDBWithRetry(ctx context.Context, db *sql.DB) error {
	for i := range maxPingRetries {
		slog.Info("Pinging database...", "attempt", i+1, "maxAttempts", maxPingRetries)
		err := db.PingContext(ctx)
		if err == nil {
			slog.Info("Database is ready")
			return nil
		}
		slog.Warn("Database not ready, retrying...", "error", err, "retryIn", pingInterval)
		select {
		case <-time.After(pingInterval):
		case <-ctx.Done():
			return fmt.Errorf("database ping timed out: %w", ctx.Err())
		}
	}
	return fmt.Errorf("database is not ready after %d attempts", maxPingRetries)
}

// createTables creates the necessary tables if they do not exist.
func createTables(ctx context.Context, db *sql.DB) error {
	queries := []string{
		sqlCreateMunicipalityTaxesTable,
		sqlCreateIndexes,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

// prepareStatements prepares all the necessary SQL statements for the store.
func (s *PostgresStore) prepareStatements(ctx context.Context) error {
	statementsToPrepare := map[string]string{
		"insertOrUpdateTaxRecord": sqlInsertOrUpdateTaxRecord,
		"selectTaxRate":           sqlSelectTaxRate,
	}
	for name, query := range statementsToPrepare {
		stmt, err := s.db.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement %s: %w", name, err)
		}
		s.preparedStatements[name] = stmt
	}
	return nil
}

// CleanupDB removes all data from the database, used for testing purposes.
// Warning: This will remove all data from the database.
func (s *PostgresStore) CleanupDB() error {
	queries := []string{
		sqlTruncateMunicipalityTaxesTable,
	}
	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// AddOrUpdateTaxRecord adds a new tax record or updates an existing one.
func (s *PostgresStore) AddOrUpdateTaxRecord(ctx context.Context, record model.TaxRecord) error {
	ctx, cancel := context.WithTimeout(ctx, transactionTimeout)
	defer cancel()

	period := fmt.Sprintf("[%s, %s)", record.StartDate.Format("2006-01-02"), record.EndDate.AddDate(0, 0, 1).Format("2006-01-02"))
	stmt, ok := s.preparedStatements["insertOrUpdateTaxRecord"]
	if !ok {
		return fmt.Errorf("statement 'stmtInsertOrUpdateTaxRecord' not prepared")
	}

	_, err := stmt.ExecContext(ctx, record.Municipality, record.TaxRate, period, record.PeriodType)
	if err != nil {
		return fmt.Errorf("failed to execute stmtInsertOrUpdateTaxRecord: %w", err)
	}
	return err
}

// GetTaxRate retrieves the tax rate for a municipality on a given date.
func (s *PostgresStore) GetTaxRate(ctx context.Context, query model.TaxQuery) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, statementTimeout)
	defer cancel()

	var taxRate float64

	// Construct the daterange for the query date
	dateRange := fmt.Sprintf("[%s,%s)", query.Date.Format("2006-01-02"), query.Date.AddDate(0, 0, 1).Format("2006-01-02"))

	stmt, ok := s.preparedStatements["selectTaxRate"]
	if !ok {
		return 0, fmt.Errorf("statement 'stmtSelectTaxRate' not prepared")
	}

	err := stmt.QueryRowContext(ctx, query.Municipality, dateRange).Scan(&taxRate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.ErrNotFound
		}
		return 0, fmt.Errorf("failed to execute stmtSelectTaxRate: %w", err)
	}
	return taxRate, nil
}
