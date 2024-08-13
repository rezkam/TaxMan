package taxservice

import (
	"context"
	"errors"
	"fmt"

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

	// GetTaxRecords retrieves all tax records for a municipality that match a specific date.
	// The service layer will be responsible for selecting the most appropriate record.
	GetTaxRecords(ctx context.Context, query model.TaxQuery) ([]model.TaxRecord, error)
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

// TaxRateResponse represents the response containing the tax rate and whether it is the default rate.
type TaxRateResponse struct {
	TaxRate       float64
	IsDefaultRate bool
}

// GetTaxRate retrieves the tax rate for a municipality on a specific date.
func (tx *Service) GetTaxRate(ctx context.Context, query model.TaxQuery) (TaxRateResponse, error) {
	records, err := tx.store.GetTaxRecords(ctx, query)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) && tx.config.DefaultTaxRate != nil {
			return TaxRateResponse{TaxRate: *tx.config.DefaultTaxRate, IsDefaultRate: true}, nil
		}
		return TaxRateResponse{}, err
	}

	if len(records) == 0 {
		if tx.config.DefaultTaxRate != nil {
			return TaxRateResponse{TaxRate: *tx.config.DefaultTaxRate, IsDefaultRate: true}, nil
		}
		return TaxRateResponse{}, model.ErrNotFound
	}

	// Select the best record based on business logic
	bestRecord, err := tx.selectBestTaxRecord(records)
	if err == nil {
		return TaxRateResponse{TaxRate: bestRecord.TaxRate, IsDefaultRate: false}, nil
	}

	// If no suitable record was found, fall back to the default tax rate if it exists
	if tx.config.DefaultTaxRate != nil {
		return TaxRateResponse{TaxRate: *tx.config.DefaultTaxRate, IsDefaultRate: true}, nil
	}

	// If no records and no default tax rate, return an error
	return TaxRateResponse{}, model.ErrNotFound

}

// selectBestTaxRecord selects the most appropriate tax record from a list of records.
// if multiple tax rates apply to a specific date, the record with the highest priority period type is selected
// if multiple records have the same period type, the record with the highest tax rate is selected
func (tx *Service) selectBestTaxRecord(records []model.TaxRecord) (*model.TaxRecord, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no tax records found")
	}

	bestRecord := &records[0] // Start with the first record as the best candidate

	for _, record := range records[1:] {
		currentPriority, err := model.GetPeriodTypePriority(record.PeriodType)
		if err != nil {
			return nil, err
		}
		bestPriority, err := model.GetPeriodTypePriority(bestRecord.PeriodType)
		if err != nil {
			return nil, err
		}
		// Compare each record to determine if it's better than the current best
		if currentPriority < bestPriority {
			bestRecord = &record
		} else if currentPriority == bestPriority && record.TaxRate > bestRecord.TaxRate {
			bestRecord = &record
		}
	}

	return bestRecord, nil
}
