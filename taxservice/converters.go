package taxservice

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/rezkam/TaxMan/model"
)

// validateMunicipality checks common municipality validation rules.
func validateMunicipality(municipality string, maxLength int) error {
	if municipality == "" {
		return errors.New("municipality is required")
	}
	if utf8.RuneCountInString(municipality) > maxLength {
		return errors.New("municipality name exceeds maximum length")
	}
	return nil
}

// validateDate parses and validates the date string.
func validateDate(dateStr, fieldName string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, errors.New(fieldName + " is required")
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, errors.New("invalid " + fieldName + " format")
	}
	return date, nil
}

// AddOrUpdateTaxRecordRequestToModel converts and validates the request for adding or updating a tax record.
func (tx *TaxService) AddOrUpdateTaxRecordRequestToModel(req AddOrUpdateTaxRecordRequest) (model.TaxRecord, error) {
	if err := validateMunicipality(req.Municipality, tx.config.MaxMunicipalityNameLength); err != nil {
		return model.TaxRecord{}, err
	}
	if req.TaxRate < 0 {
		return model.TaxRecord{}, errors.New("tax rate cannot be negative")
	}
	startDate, err := validateDate(req.StartDate, "start date")
	if err != nil {
		return model.TaxRecord{}, err
	}
	endDate, err := validateDate(req.EndDate, "end date")
	if err != nil {
		return model.TaxRecord{}, err
	}

	taxRecord := model.TaxRecord{
		Municipality: req.Municipality,
		TaxRate:      req.TaxRate,
		StartDate:    startDate,
		EndDate:      endDate,
	}
	return taxRecord, nil
}

// GetTaxRateRequestToModel converts and validates the request for retrieving the tax rate.
func (tx *TaxService) GetTaxRateRequestToModel(req GetTaxRateRequest) (model.TaxQuery, error) {
	if err := validateMunicipality(req.Municipality, tx.config.MaxMunicipalityNameLength); err != nil {
		return model.TaxQuery{}, err
	}
	date, err := validateDate(req.Date, "date")
	if err != nil {
		return model.TaxQuery{}, err
	}

	taxQuery := model.TaxQuery{
		Municipality: req.Municipality,
		Date:         date,
	}
	return taxQuery, nil
}
