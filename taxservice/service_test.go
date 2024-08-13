package taxservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rezkam/TaxMan/internal/utils"
	"github.com/rezkam/TaxMan/model"
	"github.com/stretchr/testify/require"
)

func TestGetTaxRate(t *testing.T) {
	defaultTaxRate := 0.1

	t.Run("success with best record selected", func(t *testing.T) {
		mockStore := &mockStore{
			getTaxRecordsFunc: func(ctx context.Context, query model.TaxQuery) ([]model.TaxRecord, error) {
				return []model.TaxRecord{
					{Municipality: "Copenhagen", TaxRate: 0.2, PeriodType: model.Yearly},
					{Municipality: "Copenhagen", TaxRate: 0.5, PeriodType: model.Monthly},
				}, nil
			},
		}
		svc, err := New(mockStore, Config{
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
		})
		require.NoError(t, err)

		query := model.TaxQuery{Municipality: "Copenhagen", Date: utils.DateOnly(2024, time.March, 1)}
		resp, err := svc.GetTaxRate(context.Background(), query)
		require.NoError(t, err)
		require.False(t, resp.IsDefaultRate)
		require.Equal(t, 0.5, resp.TaxRate)
	})

	t.Run("fallback to default rate when no records found", func(t *testing.T) {
		mockStore := &mockStore{
			getTaxRecordsFunc: func(ctx context.Context, query model.TaxQuery) ([]model.TaxRecord, error) {
				return nil, model.ErrNotFound
			},
		}
		svc, err := New(mockStore, Config{
			DefaultTaxRate:            &defaultTaxRate,
			MaxMunicipalityNameLength: 20,
			MunicipalityURLPattern:    "municipality",
			DateURLPattern:            "date",
		})
		require.NoError(t, err)

		query := model.TaxQuery{Municipality: "NonExistent", Date: utils.DateOnly(2024, time.March, 1)}
		resp, err := svc.GetTaxRate(context.Background(), query)
		require.NoError(t, err)
		require.True(t, resp.IsDefaultRate)
		require.Equal(t, defaultTaxRate, resp.TaxRate)
	})

	t.Run("error when no records found and no default rate", func(t *testing.T) {
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

		query := model.TaxQuery{Municipality: "NonExistent", Date: utils.DateOnly(2024, time.March, 1)}
		resp, err := svc.GetTaxRate(context.Background(), query)
		require.Error(t, err)
		require.Equal(t, model.ErrNotFound, err)
		require.Equal(t, 0.0, resp.TaxRate)
	})

	t.Run("error from store", func(t *testing.T) {
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

		query := model.TaxQuery{Municipality: "Copenhagen", Date: utils.DateOnly(2024, time.March, 1)}
		resp, err := svc.GetTaxRate(context.Background(), query)
		require.Error(t, err)
		require.Equal(t, "store error", err.Error())
		require.Equal(t, 0.0, resp.TaxRate)
	})
}

func TestSelectBestTaxRecord(t *testing.T) {
	svc, _ := New(nil, Config{
		MaxMunicipalityNameLength: 20,
		MunicipalityURLPattern:    "municipality",
		DateURLPattern:            "date",
	})

	t.Run("select best record by period type priority", func(t *testing.T) {
		records := []model.TaxRecord{
			{Municipality: "Copenhagen", TaxRate: 0.2, PeriodType: model.Yearly},
			{Municipality: "Copenhagen", TaxRate: 0.5, PeriodType: model.Monthly},
			{Municipality: "Copenhagen", TaxRate: 0.1, PeriodType: model.Daily},
		}

		bestRecord, err := svc.selectBestTaxRecord(records)
		require.NoError(t, err)
		require.Equal(t, 0.1, bestRecord.TaxRate)
		require.Equal(t, model.Daily, bestRecord.PeriodType)
	})

	t.Run("select highest tax rate if same period type", func(t *testing.T) {
		records := []model.TaxRecord{
			{Municipality: "Copenhagen", TaxRate: 0.2, PeriodType: model.Yearly},
			{Municipality: "Copenhagen", TaxRate: 0.3, PeriodType: model.Yearly},
		}

		bestRecord, err := svc.selectBestTaxRecord(records)
		require.NoError(t, err)
		require.Equal(t, 0.3, bestRecord.TaxRate)
		require.Equal(t, model.Yearly, bestRecord.PeriodType)
	})

	t.Run("error if no records provided", func(t *testing.T) {
		var records []model.TaxRecord
		_, err := svc.selectBestTaxRecord(records)
		require.Error(t, err)
		require.Equal(t, "no tax records found", err.Error())
	})
}
