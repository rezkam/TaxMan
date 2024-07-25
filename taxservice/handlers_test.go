package taxservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rezkam/TaxMan/model"
	"github.com/stretchr/testify/require"
)

type MockStore struct {
	addOrUpdateTaxRecordFunc func(ctx context.Context, record model.TaxRecord) error
	getTaxRateFunc           func(ctx context.Context, query model.TaxQuery) (float64, error)
}

func (m *MockStore) AddOrUpdateTaxRecord(ctx context.Context, record model.TaxRecord) error {
	if m.addOrUpdateTaxRecordFunc != nil {
		return m.addOrUpdateTaxRecordFunc(ctx, record)
	}
	return nil
}

func (m *MockStore) GetTaxRate(ctx context.Context, query model.TaxQuery) (float64, error) {
	if m.getTaxRateFunc != nil {
		return m.getTaxRateFunc(ctx, query)
	}
	return 0, nil
}

func TestAddOrUpdateTaxRecordHandler(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        AddOrUpdateTaxRecordRequest
		expectedStatusCode int
		mockReturnError    error
	}{
		{
			name:               "success",
			requestBody:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 10, StartDate: "2020-12-31", EndDate: "2021-12-31"},
			expectedStatusCode: http.StatusOK,
			mockReturnError:    nil,
		},
		{
			name:               "invalid municipality",
			requestBody:        AddOrUpdateTaxRecordRequest{Municipality: "", TaxRate: 10, StartDate: "2020-12-31", EndDate: "2021-12-31"},
			expectedStatusCode: http.StatusBadRequest,
			mockReturnError:    nil,
		},
		{
			name:               "negative tax rate",
			requestBody:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: -10, StartDate: "2020-12-31", EndDate: "2021-12-31"},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid start date",
			requestBody:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 10, StartDate: "invalid-date", EndDate: "2021-12-31"},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid end date",
			requestBody:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 10, StartDate: "2020-12-31", EndDate: "invalid-date"},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "store error",
			requestBody:        AddOrUpdateTaxRecordRequest{Municipality: "Valid Name", TaxRate: 10, StartDate: "2020-12-31", EndDate: "2021-12-31"},
			expectedStatusCode: http.StatusInternalServerError,
			mockReturnError:    errors.New("store error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStore{
				addOrUpdateTaxRecordFunc: func(ctx context.Context, record model.TaxRecord) error {
					return tt.mockReturnError
				},
			}
			svc, err := New(mockStore, Config{
				MaxMunicipalityNameLength: 20,
			})
			require.NoError(t, err)

			reqBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/addOrUpdateTaxRecord", bytes.NewReader(reqBody))
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(svc.AddOrUpdateTaxRecordHandler)
			handler.ServeHTTP(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatusCode, resp.StatusCode)
		})
	}
}

func TestGetTaxRateHandler(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        GetTaxRateRequest
		expectedStatusCode int
		mockReturnTaxRate  float64
		mockReturnError    error
	}{
		{
			name:               "success",
			requestBody:        GetTaxRateRequest{Municipality: "Valid Name", Date: "2020-12-31"},
			expectedStatusCode: http.StatusOK,
			mockReturnTaxRate:  5.5,
			mockReturnError:    nil,
		},
		{
			name:               "invalid municipality",
			requestBody:        GetTaxRateRequest{Municipality: "", Date: "2020-12-31"},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid date format",
			requestBody:        GetTaxRateRequest{Municipality: "Valid Name", Date: "invalid-date"},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "store error",
			requestBody:        GetTaxRateRequest{Municipality: "Valid Name", Date: "2020-12-31"},
			expectedStatusCode: http.StatusInternalServerError,
			mockReturnTaxRate:  0,
			mockReturnError:    errors.New("store error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStore{
				getTaxRateFunc: func(ctx context.Context, query model.TaxQuery) (float64, error) {
					return tt.mockReturnTaxRate, tt.mockReturnError
				},
			}
			svc, err := New(mockStore, Config{
				MaxMunicipalityNameLength: 20,
			})
			require.NoError(t, err)

			reqBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/getTaxRate", bytes.NewReader(reqBody))
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(svc.GetTaxRateHandler)
			handler.ServeHTTP(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatusCode, resp.StatusCode)
		})
	}
}
