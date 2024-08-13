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

func TestAddOrUpdateTaxRecordHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockStore := &mockStore{
			addOrUpdateTaxRecordFunc: func(ctx context.Context, record model.TaxRecord) error {
				return nil
			},
		}
		svc, err := New(mockStore, Config{
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
		})
		require.NoError(t, err)

		reqBody, err := json.Marshal(AddOrUpdateTaxRecordRequest{
			Municipality: "Valid Name",
			TaxRate:      0.1,
			StartDate:    "2020-12-31",
			EndDate:      "2021-12-31",
			PeriodType:   model.Yearly,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(svc.AddOrUpdateTaxRecordHandler)
		handler.ServeHTTP(rr, req)

		resp := rr.Result()
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid municipality", func(t *testing.T) {
		mockStore := &mockStore{}
		svc, err := New(mockStore, Config{
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
		})
		require.NoError(t, err)

		reqBody, err := json.Marshal(AddOrUpdateTaxRecordRequest{
			Municipality: "",
			TaxRate:      0.1,
			StartDate:    "2020-12-31",
			EndDate:      "2021-12-31",
			PeriodType:   model.Yearly,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(svc.AddOrUpdateTaxRecordHandler)
		handler.ServeHTTP(rr, req)

		resp := rr.Result()
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("store error", func(t *testing.T) {
		mockStore := &mockStore{
			addOrUpdateTaxRecordFunc: func(ctx context.Context, record model.TaxRecord) error {
				return errors.New("store error")
			},
		}
		svc, err := New(mockStore, Config{
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
		})
		require.NoError(t, err)

		reqBody, err := json.Marshal(AddOrUpdateTaxRecordRequest{
			Municipality: "Valid Name",
			TaxRate:      0.1,
			StartDate:    "2020-12-31",
			EndDate:      "2021-12-31",
			PeriodType:   model.Yearly,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(svc.AddOrUpdateTaxRecordHandler)
		handler.ServeHTTP(rr, req)

		resp := rr.Result()
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestGetTaxRateHandler(t *testing.T) {

	defaultTaxRate := 0.9

	t.Run("success", func(t *testing.T) {
		mockStore := &mockStore{
			getTaxRecordsFunc: func(ctx context.Context, query model.TaxQuery) ([]model.TaxRecord, error) {
				return []model.TaxRecord{
					{TaxRate: 5.5, PeriodType: model.Yearly},
				}, nil
			},
		}
		svc, err := New(mockStore, Config{
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		req.SetPathValue(svc.config.MunicipalityURLPattern, "Valid Name")
		req.SetPathValue(svc.config.DateURLPattern, "2020-12-31")

		handler := http.HandlerFunc(svc.GetTaxRateHandler)
		handler.ServeHTTP(rr, req)

		resp := rr.Result()
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody GetTaxRateResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)
		require.Equal(t, 5.5, respBody.TaxRate)
	})

	t.Run("invalid municipality", func(t *testing.T) {
		mockStore := &mockStore{}
		svc, err := New(mockStore, Config{
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		req.SetPathValue(svc.config.MunicipalityURLPattern, "")
		req.SetPathValue(svc.config.DateURLPattern, "2020-12-31")

		handler := http.HandlerFunc(svc.GetTaxRateHandler)
		handler.ServeHTTP(rr, req)

		resp := rr.Result()
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("store error", func(t *testing.T) {
		mockStore := &mockStore{
			getTaxRecordsFunc: func(ctx context.Context, query model.TaxQuery) ([]model.TaxRecord, error) {
				return nil, errors.New("store error")
			},
		}
		svc, err := New(mockStore, Config{
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		req.SetPathValue(svc.config.MunicipalityURLPattern, "Valid Name")
		req.SetPathValue(svc.config.DateURLPattern, "2020-12-31")

		handler := http.HandlerFunc(svc.GetTaxRateHandler)
		handler.ServeHTTP(rr, req)

		resp := rr.Result()
		defer resp.Body.Close()

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
	t.Run("not found error with default value", func(t *testing.T) {
		mockStore := &mockStore{
			getTaxRecordsFunc: func(ctx context.Context, query model.TaxQuery) ([]model.TaxRecord, error) {
				return nil, model.ErrNotFound
			},
		}
		svc, err := New(mockStore, Config{
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
			DefaultTaxRate:            &defaultTaxRate,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		req.SetPathValue(svc.config.MunicipalityURLPattern, "NonExistent")
		req.SetPathValue(svc.config.DateURLPattern, "2020-12-31")

		handler := http.HandlerFunc(svc.GetTaxRateHandler)
		handler.ServeHTTP(rr, req)

		resp := rr.Result()
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody GetTaxRateResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)
		require.Equal(t, defaultTaxRate, respBody.TaxRate)
		require.True(t, respBody.IsDefaultRate)
	})

	t.Run("not found error without default value", func(t *testing.T) {
		mockStore := &mockStore{
			getTaxRecordsFunc: func(ctx context.Context, query model.TaxQuery) ([]model.TaxRecord, error) {
				return nil, model.ErrNotFound
			},
		}
		svc, err := New(mockStore, Config{
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		req.SetPathValue(svc.config.MunicipalityURLPattern, "NonExistent")
		req.SetPathValue(svc.config.DateURLPattern, "2020-12-31")

		handler := http.HandlerFunc(svc.GetTaxRateHandler)
		handler.ServeHTTP(rr, req)

		resp := rr.Result()
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

}
