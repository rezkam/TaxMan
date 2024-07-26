package taxservice

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rezkam/TaxMan/internal/jsonutils"
	"github.com/rezkam/TaxMan/model"
)

func (tx *Service) AddOrUpdateTaxRecordHandler(w http.ResponseWriter, r *http.Request) {
	var req AddOrUpdateTaxRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutils.JsonError(w, "invalid json input", http.StatusBadRequest)
		return
	}

	taxRecord, err := tx.AddOrUpdateTaxRecordRequestToModel(req)
	if err != nil {
		jsonutils.JsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = tx.store.AddOrUpdateTaxRecord(r.Context(), taxRecord)
	if err != nil {
		slog.Error("failed to add or update tax record", "error", err)
		jsonutils.JsonError(w, "failed to add or update tax record", http.StatusInternalServerError)
		return
	}

	resp := AddOrUpdateTaxRecordResponse{Success: true}
	jsonutils.JsonResponse(w, resp, http.StatusOK)
}

func (tx *Service) GetTaxRateHandler(w http.ResponseWriter, r *http.Request) {
	municipality := r.PathValue(tx.config.MunicipalityURLPattern)
	date := r.PathValue(tx.config.DateURLPattern)

	taxQuery, err := tx.GetTaxRateRequestToModel(municipality, date)
	if err != nil {
		jsonutils.JsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	taxRate, err := tx.store.GetTaxRate(r.Context(), taxQuery)
	if err != nil {
		slog.Error("failed to get tax rate", "error", err)
		if errors.Is(err, model.ErrNotFound) {
			if tx.config.DefaultTaxRate != nil {
				resp := GetTaxRateResponse{
					Municipality:  municipality,
					Date:          date,
					TaxRate:       *tx.config.DefaultTaxRate,
					IsDefaultRate: true,
				}
				jsonutils.JsonResponse(w, resp, http.StatusOK)
				return
			}
			jsonutils.JsonError(w, "tax rate not found", http.StatusNotFound)
			return
		}
		jsonutils.JsonError(w, "failed to get tax rate", http.StatusInternalServerError)
		return
	}

	resp := GetTaxRateResponse{
		Municipality:  municipality,
		Date:          date,
		TaxRate:       taxRate,
		IsDefaultRate: false,
	}
	jsonutils.JsonResponse(w, resp, http.StatusOK)
}
