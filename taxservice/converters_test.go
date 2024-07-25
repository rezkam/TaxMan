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

type MockTaxService struct {
	Config Config
}

func (m *MockTaxService) AddOrUpdateTaxRecordRequestToModel(req AddOrUpdateTaxRecordRequest) (model.TaxRecord, error) {
	if req.Municipality == "" {
		return model.TaxRecord{}, errors.New("municipality is required")
	}
	if req.TaxRate < 0 {
		return model.TaxRecord{}, errors.New("tax rate cannot be negative")
	}
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return model.TaxRecord{}, errors.New("invalid start date format")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return model.TaxRecord{}, errors.New("invalid end date format")
	}
	return model.TaxRecord{
		Municipality: req.Municipality,
		TaxRate:      req.TaxRate,
		StartDate:    startDate,
		EndDate:      endDate,
	}, nil
}

func (m *MockTaxService) GetTaxRateRequestToModel(req GetTaxRateRequest) (model.TaxQuery, error) {
	if req.Municipality == "" {
		return model.TaxQuery{}, errors.New("municipality is required")
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return model.TaxQuery{}, errors.New("invalid date format")
	}
	return model.TaxQuery{
		Municipality: req.Municipality,
		Date:         date,
	}, nil
}

func TestAddOrUpdateTaxRecordRequestToModel(t *testing.T) {
	mockTaxService := &MockTaxService{
		Config: Config{MaxMunicipalityNameLength: 20},
	}

	tests := []struct {
		name           string
		request        AddOrUpdateTaxRecordRequest
		expectedRecord model.TaxRecord
		expectedErr    error
	}{
		{
			name:           "Invalid Municipality",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "", TaxRate: 10, StartDate: "2020-12-31", EndDate: "2021-12-31"},
			expectedRecord: model.TaxRecord{},
			expectedErr:    errors.New("municipality is required"),
		},
		{
			name:           "Negative Tax Rate",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: -10, StartDate: "2020-12-31", EndDate: "2021-12-31"},
			expectedRecord: model.TaxRecord{},
			expectedErr:    errors.New("tax rate cannot be negative"),
		},
		{
			name:           "Invalid Start Date",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 10, StartDate: "invalid-date", EndDate: "2021-12-31"},
			expectedRecord: model.TaxRecord{},
			expectedErr:    errors.New("invalid start date format"),
		},
		{
			name:           "Valid Request",
			request:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 10, StartDate: "2020-12-31", EndDate: "2021-12-31"},
			expectedRecord: model.TaxRecord{Municipality: "Valid Name", TaxRate: 10, StartDate: time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC)},
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := mockTaxService.AddOrUpdateTaxRecordRequestToModel(tt.request)
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
	mockTaxService := &MockTaxService{
		Config: Config{MaxMunicipalityNameLength: 20},
	}

	tests := []struct {
		name          string
		request       GetTaxRateRequest
		expectedQuery model.TaxQuery
		expectedErr   error
	}{
		{
			name:          "Invalid Municipality",
			request:       GetTaxRateRequest{Municipality: "", Date: "2020-12-31"},
			expectedQuery: model.TaxQuery{},
			expectedErr:   errors.New("municipality is required"),
		},
		{
			name:          "Invalid Date",
			request:       GetTaxRateRequest{Municipality: "Valid Name", Date: "invalid-date"},
			expectedQuery: model.TaxQuery{},
			expectedErr:   errors.New("invalid date format"),
		},
		{
			name:          "Valid Request",
			request:       GetTaxRateRequest{Municipality: "Valid Name", Date: "2020-12-31"},
			expectedQuery: model.TaxQuery{Municipality: "Valid Name", Date: time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)},
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := mockTaxService.GetTaxRateRequestToModel(tt.request)
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
