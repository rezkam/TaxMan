package model

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidPeriod = errors.New("invalid period type")
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

// periodTypePriority defines the priority of period types
var periodTypePriority = map[PeriodType]int{
	Daily:   1,
	Weekly:  2,
	Monthly: 3,
	Yearly:  4,
}

// GetPeriodTypePriority retrieves the priority of a period type.
func GetPeriodTypePriority(pt PeriodType) (int, error) {
	priority, exists := periodTypePriority[pt]
	if !exists {
		return 0, ErrInvalidPeriod
	}
	return priority, nil
}

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
