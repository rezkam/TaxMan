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

// Configuration for retry logic
const (
	maxPingRetries = 5
	pingInterval   = 1 * time.Second
)

type PostgresStore struct {
	db  *sql.DB
	log *slog.Logger

	// Prepared statements
	stmtInsertOrUpdateTaxRecord *sql.Stmt
	stmtSelectTaxRate           *sql.Stmt
}

// NewPostgresStore initializes and returns a new PostgresStore after ensuring the database is ready.
func NewPostgresStore(connStr string, log *slog.Logger) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ensure the database is ready
	if !isDBReady(db, log) {
		return nil, fmt.Errorf("database is not ready")
	}

	// Create tables if they do not exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Prepare statements
	store := &PostgresStore{db: db, log: log}
	if err := store.prepareStatements(); err != nil {
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	return store, nil
}

// Close closes the database connection.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// isDBReady checks if the database is ready by pinging it multiple times.
func isDBReady(db *sql.DB, log *slog.Logger) bool {
	for i := 0; i < maxPingRetries; i++ {
		err := db.Ping()
		if err == nil {
			log.Info("Database is ready")
			return true
		}
		log.Info("Database not ready", "retrying in", pingInterval, "attempt", i+1, "max attempts", maxPingRetries)
		time.Sleep(pingInterval)
	}
	log.Error("Database is not ready after", "attempts", maxPingRetries)
	return false
}

// createTables creates the necessary tables if they do not exist.
func createTables(db *sql.DB) error {
	queries := []string{
		sqlCreateMunicipalityTaxesTable,
		sqlCreateIndexes,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
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

// prepareStatements prepares all the necessary SQL statements for the store.
func (s *PostgresStore) prepareStatements() error {
	var err error
	s.stmtInsertOrUpdateTaxRecord, err = s.db.Prepare(sqlInsertOrUpdateTaxRecord)
	if err != nil {
		return err
	}
	s.stmtSelectTaxRate, err = s.db.Prepare(sqlSelectTaxRate)
	if err != nil {
		return err
	}
	return nil
}

// AddOrUpdateTaxRecord adds a new tax record or updates an existing one.
func (s *PostgresStore) AddOrUpdateTaxRecord(ctx context.Context, record model.TaxRecord) error {
	period := fmt.Sprintf("[%s, %s)", record.StartDate.Format("2006-01-02"), record.EndDate.AddDate(0, 0, 1).Format("2006-01-02"))
	_, err := s.stmtInsertOrUpdateTaxRecord.ExecContext(ctx, record.Municipality, record.TaxRate, period, record.PeriodType)
	return err
}

// GetTaxRate retrieves the tax rate for a municipality on a given date.
func (s *PostgresStore) GetTaxRate(ctx context.Context, query model.TaxQuery) (float64, error) {
	var taxRate float64

	// Construct the daterange for the query date
	dateRange := fmt.Sprintf("[%s,%s)", query.Date.Format("2006-01-02"), query.Date.AddDate(0, 0, 1).Format("2006-01-02"))

	err := s.stmtSelectTaxRate.QueryRowContext(ctx, query.Municipality, dateRange).Scan(&taxRate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.ErrNotFound
		}
		return 0, err
	}
	return taxRate, nil
}
