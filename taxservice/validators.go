package taxservice

import (
	"errors"
	"time"
	"unicode/utf8"
)

// validateAddOrUpdateTaxRecordRequest validates the AddOrUpdateTaxRecordRequest fields
func validateAddOrUpdateTaxRecordRequest(req AddOrUpdateTaxRecordRequest, maxMunicipalityNameLength int) error {
	if req.Municipality == "" {
		return errors.New("municipality is required")
	}
	if utf8.RuneCountInString(req.Municipality) > maxMunicipalityNameLength {
		return errors.New("municipality name exceeds maximum length")
	}
	if req.TaxRate < 0 {
		return errors.New("tax rate cannot be negative")
	}
	if req.StartDate == "" {
		return errors.New("start date is required")
	}
	if req.EndDate == "" {
		return errors.New("end date is required")
	}
	if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
		return errors.New("invalid start date format")
	}
	if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
		return errors.New("invalid end date format")
	}
	return nil
}

// validateGetTaxRateRequest validates the GetTaxRateRequest fields
func validateGetTaxRateRequest(req GetTaxRateRequest, maxMunicipalityNameLength int) error {
	if req.Municipality == "" {
		return errors.New("municipality is required")
	}
	if utf8.RuneCountInString(req.Municipality) > maxMunicipalityNameLength {
		return errors.New("municipality name exceeds maximum length")
	}
	if req.Date == "" {
		return errors.New("date is required")
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return errors.New("invalid date format")
	}
	return nil
}
