package taxservice

import (
	"context"
	"errors"

	"github.com/rezkam/TaxMan/model"
)

// TaxService handles the business logic for managing municipality tax records.
type TaxService struct {
	store  taxStore
	config Config
}

type Config struct {
	// MaxMunicipalityNameLength is the maximum length allowed for a municipality's name.
	MaxMunicipalityNameLength int
}

type taxStore interface {
	// AddOrUpdateTaxRecord adds a new tax record or updates an existing one.
	AddOrUpdateTaxRecord(ctx context.Context, record model.TaxRecord) error
	// GetTaxRate retrieves the tax rate for a municipality on a given date.
	GetTaxRate(ctx context.Context, query model.TaxQuery) (float64, error)
}

// New creates a new TaxService with the provided store and configuration.
func New(store taxStore, config Config) (*TaxService, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return &TaxService{store: store, config: config}, nil
}

// validateConfig checks if the provided Config values are valid.
func validateConfig(config Config) error {
	if config.MaxMunicipalityNameLength <= 0 {
		return errors.New("MaxMunicipalityNameLength must be greater than 0")
	}
	return nil
}
