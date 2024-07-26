package model

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

// PeriodType defines the type of period for a tax record
type PeriodType string

const (
	Yearly  PeriodType = "yearly"
	Monthly PeriodType = "monthly"
	Weekly  PeriodType = "weekly"
	Daily   PeriodType = "daily"
)

// ValidPeriodTypes contains all valid period types
var ValidPeriodTypes = []PeriodType{Yearly, Monthly, Weekly, Daily}

// TaxRecord represents a tax record with appropriate types.
type TaxRecord struct {
	Municipality string
	TaxRate      float64
	StartDate    time.Time
	EndDate      time.Time
	PeriodType   PeriodType
}

// TaxQuery represents a query for a tax rate.
type TaxQuery struct {
	Municipality string
	Date         time.Time
}
