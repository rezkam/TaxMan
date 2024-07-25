package taxservice

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/rezkam/TaxMan/internal/jsonutils"
)

func (tx *TaxService) AddOrUpdateTaxRecordHandler(w http.ResponseWriter, r *http.Request) {
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

func (tx *TaxService) GetTaxRateHandler(w http.ResponseWriter, r *http.Request) {
	var req GetTaxRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutils.JsonError(w, "invalid json input", http.StatusBadRequest)
		return
	}

	taxQuery, err := tx.GetTaxRateRequestToModel(req)
	if err != nil {
		jsonutils.JsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	taxRate, err := tx.store.GetTaxRate(r.Context(), taxQuery)
	if err != nil {
		slog.Error("failed to get tax rate", "error", err)
		jsonutils.JsonError(w, "failed to get tax rate", http.StatusInternalServerError)
		return
	}

	resp := GetTaxRateResponse{
		Municipality: req.Municipality,
		Date:         req.Date,
		TaxRate:      taxRate,
	}
	jsonutils.JsonResponse(w, resp, http.StatusOK)
}
