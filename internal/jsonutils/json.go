package jsonutils

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// JsonError formats an error message as JSON and writes it to the response.
func JsonError(w http.ResponseWriter, msg string, code int) {
	errorMessage := ErrorResponse{Error: msg}
	JsonResponse(w, errorMessage, code)
}

// JsonResponse formats data as JSON and writes it to the response.
func JsonResponse(w http.ResponseWriter, data any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
