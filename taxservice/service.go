package taxservice

import (
	"context"
	"errors"

	"github.com/rezkam/TaxMan/model"
)

// Service handles the business logic for managing municipality tax records.
type Service struct {
	store  taxStore
	config Config
}

type Config struct {
	// MaxMunicipalityNameLength is the maximum length allowed for a municipality's name.
	MaxMunicipalityNameLength int
	// MunicipalityNamePattern is the pattern used to extract the municipality name from a URL.
	MunicipalityURLPattern string
	// DateURLPattern is the pattern used to extract the date from a URL.
	DateURLPattern string
	// DefaultTaxRate is the default tax rate to use if no specific rate is found for a municipality.
	// This value is optional and can be nil.
	DefaultTaxRate *float64
}

type taxStore interface {
	// AddOrUpdateTaxRecord adds a new tax record or updates an existing one.
	AddOrUpdateTaxRecord(ctx context.Context, record model.TaxRecord) error
	// GetTaxRate retrieves the tax rate for a municipality on a given date.
	// multiple tax rates apply to a specific date, the more specific rate which has the smallest period should be returned.
	// takes precedence (daily > weekly > monthly > yearly).
	// if we have two records with the same length of the period we return the one with the highest tax rate.
	GetTaxRate(ctx context.Context, query model.TaxQuery) (float64, error)
}

// New creates a new Service with the provided store and configuration.
func New(store taxStore, config Config) (*Service, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return &Service{store: store, config: config}, nil
}

// validateConfig checks if the provided Config values are valid.
func validateConfig(config Config) error {
	if config.MaxMunicipalityNameLength <= 0 {
		return errors.New("MaxMunicipalityNameLength must be greater than 0")
	}
	if config.DefaultTaxRate != nil && (*config.DefaultTaxRate < 0 || *config.DefaultTaxRate > 1) {
		return errors.New("DefaultTaxRate must be between 0.0 and 1.0")
	}
	if config.MunicipalityURLPattern == "" {
		return errors.New("MunicipalityNamePattern cannot be empty")
	}
	if config.DateURLPattern == "" {
		return errors.New("DatePattern cannot be empty")
	}
	return nil
}
