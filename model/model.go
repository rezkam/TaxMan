package model

import "time"

// TaxRecord represents a tax record with appropriate types.
type TaxRecord struct {
	Municipality string
	TaxRate      float64
	StartDate    time.Time
	EndDate      time.Time
}

// TaxQuery represents a query for a tax rate.
type TaxQuery struct {
	Municipality string
	Date         time.Time
}
