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

// validatePeriodType checks if the period type is valid.
func validatePeriodType(periodType model.PeriodType) error {
	for _, validType := range model.ValidPeriodTypes {
		if periodType == validType {
			return nil
		}
	}
	return errors.New("invalid period type")
}

// AddOrUpdateTaxRecordRequestToModel converts and validates the request for adding or updating a tax record.
// Assumption: Tax rates are expressed as decimal values representing percentages (e.g., 0.1 for 10%).
func (tx *Service) AddOrUpdateTaxRecordRequestToModel(req AddOrUpdateTaxRecordRequest) (model.TaxRecord, error) {
	if err := validateMunicipality(req.Municipality, tx.config.MaxMunicipalityNameLength); err != nil {
		return model.TaxRecord{}, err
	}
	if req.TaxRate < 0.0 || req.TaxRate > 1.0 {
		return model.TaxRecord{}, errors.New("tax rate must be between 0.0 and 1.0")
	}
	startDate, err := validateDate(req.StartDate, "start date")
	if err != nil {
		return model.TaxRecord{}, err
	}
	endDate, err := validateDate(req.EndDate, "end date")
	if err != nil {
		return model.TaxRecord{}, err
	}
	if err := validatePeriodType(req.PeriodType); err != nil {
		return model.TaxRecord{}, err
	}

	taxRecord := model.TaxRecord{
		Municipality: req.Municipality,
		TaxRate:      req.TaxRate,
		StartDate:    startDate,
		EndDate:      endDate,
		PeriodType:   req.PeriodType,
	}
	return taxRecord, nil
}

// GetTaxRateRequestToModel converts and validates the request for retrieving the tax rate.
func (tx *Service) GetTaxRateRequestToModel(municipality, date string) (model.TaxQuery, error) {
	if err := validateMunicipality(municipality, tx.config.MaxMunicipalityNameLength); err != nil {
		return model.TaxQuery{}, err
	}
	parsedDate, err := validateDate(date, "date")
	if err != nil {
		return model.TaxQuery{}, err
	}

	taxQuery := model.TaxQuery{
		Municipality: municipality,
		Date:         parsedDate,
	}
	return taxQuery, nil
}
