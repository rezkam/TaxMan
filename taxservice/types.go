package taxservice

import "github.com/rezkam/TaxMan/model"

// AddOrUpdateTaxRecordRequest is the request type for adding or updating a tax record.
type AddOrUpdateTaxRecordRequest struct {
	Municipality string           `json:"municipality"`
	TaxRate      float64          `json:"tax_rate"`
	StartDate    string           `json:"start_date"`
	EndDate      string           `json:"end_date"`
	PeriodType   model.PeriodType `json:"period_type"`
}

// AddOrUpdateTaxRecordResponse is the response type for adding or updating a tax record.
type AddOrUpdateTaxRecordResponse struct {
	Success bool `json:"success"`
}

// GetTaxRateResponse is the response type for retrieving the tax rate for a municipality on a given date.
type GetTaxRateResponse struct {
	Municipality  string  `json:"municipality"`
	Date          string  `json:"date"`
	TaxRate       float64 `json:"tax_rate"`
	IsDefaultRate bool    `json:"is_default_rate"`
}
