package taxservice

// AddOrUpdateTaxRecordRequest is the request type for adding or updating a tax record.
type AddOrUpdateTaxRecordRequest struct {
	Municipality string  `json:"municipality"`
	TaxRate      float64 `json:"tax_rate"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
}

// AddOrUpdateTaxRecordResponse is the response type for adding or updating a tax record.
type AddOrUpdateTaxRecordResponse struct {
	Success bool `json:"success"`
}

// GetTaxRateRequest is the request type for retrieving the tax rate for a municipality on a given date.
type GetTaxRateRequest struct {
	Municipality string `json:"municipality"`
	Date         string `json:"date"`
}

// GetTaxRateResponse is the response type for retrieving the tax rate for a municipality on a given date.
type GetTaxRateResponse struct {
	Municipality string  `json:"municipality"`
	Date         string  `json:"date"`
	TaxRate      float64 `json:"tax_rate"`
}
