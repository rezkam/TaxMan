package taxservice

import (
	"errors"
	"testing"
	"time"

	"github.com/rezkam/TaxMan/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMunicipality(t *testing.T) {
	tests := []struct {
		name         string
		municipality string
		maxLength    int
		expectedErr  error
	}{
		{"Empty Municipality", "", 10, errors.New("municipality is required")},
		{"Exceeds Max Length", "A very long municipality name", 10, errors.New("municipality name exceeds maximum length")},
		{"Valid Municipality", "Valid Name", 20, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMunicipality(tt.municipality, tt.maxLength)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name         string
		dateStr      string
		fieldName    string
		expectedDate time.Time
		expectedErr  error
	}{
		{"Empty Date", "", "start date", time.Time{}, errors.New("start date is required")},
		{"Invalid Format", "2020-13-40", "start date", time.Time{}, errors.New("invalid start date format")},
		{"Valid Date", "2020-12-31", "start date", time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := validateDate(tt.dateStr, tt.fieldName)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDate, date)
			}
		})
	}
}

func TestAddOrUpdateTaxRecordRequestToModel(t *testing.T) {
	config := Config{
		MaxMunicipalityNameLength: 20,
		MunicipalityURLPattern:    "municipality",
		DateURLPattern:            "date",
	}
	mockStore := &mockStore{}
	svc, err := New(mockStore, config)
	require.NoError(t, err)

	tests := []struct {
		name           string
		request        AddOrUpdateTaxRecordRequest
		expectedRecord model.TaxRecord
		expectedErr    error
	}{
		{
			name:           "Invalid Municipality",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "", TaxRate: 0.1, StartDate: "2020-12-31", EndDate: "2021-12-31", PeriodType: model.Yearly},
			expectedRecord: model.TaxRecord{},
			expectedErr:    errors.New("municipality is required"),
		},
		{
			name:           "Negative Tax Rate",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: -0.1, StartDate: "2020-12-31", EndDate: "2021-12-31", PeriodType: model.Yearly},
			expectedRecord: model.TaxRecord{},
			expectedErr:    errors.New("tax rate must be between 0.0 and 1.0"),
		},
		{
			name:           "Tax Rate Exceeds Maximum",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 1.1, StartDate: "2020-12-31", EndDate: "2021-12-31", PeriodType: model.Yearly},
			expectedRecord: model.TaxRecord{},
			expectedErr:    errors.New("tax rate must be between 0.0 and 1.0"),
		},
		{
			name:           "Invalid Start Date",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 0.1, StartDate: "invalid-date", EndDate: "2021-12-31", PeriodType: model.Yearly},
			expectedRecord: model.TaxRecord{},
			expectedErr:    errors.New("invalid start date format"),
		},
		{
			name:           "Invalid End Date",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 0.1, StartDate: "2020-12-31", EndDate: "invalid-date", PeriodType: model.Yearly},
			expectedRecord: model.TaxRecord{},
			expectedErr:    errors.New("invalid end date format"),
		},
		{
			name:           "Invalid Period Type",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 0.1, StartDate: "2020-12-31", EndDate: "2021-12-31", PeriodType: "invalid"},
			expectedRecord: model.TaxRecord{},
			expectedErr:    errors.New("invalid period type"),
		},
		{
			name:           "Valid Request",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 0.1, StartDate: "2020-12-31", EndDate: "2021-12-31", PeriodType: model.Yearly},
			expectedRecord: model.TaxRecord{Municipality: "Valid Name", TaxRate: 0.1, StartDate: time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC), PeriodType: model.Yearly},
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := svc.AddOrUpdateTaxRecordRequestToModel(tt.request)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRecord, record)
			}
		})
	}
}

func TestGetTaxRateRequestToModel(t *testing.T) {
	config := Config{
		MaxMunicipalityNameLength: 20,
		MunicipalityURLPattern:    "municipality",
		DateURLPattern:            "date",
	}
	svc, _ := New(nil, config)

	tests := []struct {
		name          string
		municipality  string
		date          string
		expectedQuery model.TaxQuery
		expectedErr   error
	}{
		{
			name:          "Invalid Municipality",
			municipality:  "",
			date:          "2020-12-31",
			expectedQuery: model.TaxQuery{},
			expectedErr:   errors.New("municipality is required"),
		},
		{
			name:          "Invalid Date",
			municipality:  "Valid Name",
			date:          "invalid-date",
			expectedQuery: model.TaxQuery{},
			expectedErr:   errors.New("invalid date format"),
		},
		{
			name:          "Valid Request",
			municipality:  "Valid Name",
			date:          "2020-12-31",
			expectedQuery: model.TaxQuery{Municipality: "Valid Name", Date: time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)},
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := svc.GetTaxRateRequestToModel(tt.municipality, tt.date)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedQuery, query)
			}
		})
	}
}
