package routes

import (
	"fmt"
	"net/http"

	"github.com/rezkam/TaxMan/internal/constants"

	"github.com/rezkam/TaxMan/taxservice"
)

// SetupTaxRoutes sets up the routes for the tax service.
func SetupTaxRoutes(svc *taxservice.Service, mux *http.ServeMux) {

	const (
		municipalityNameWildcard = constants.MunicipalityURLPattern
		dateWildcard             = constants.DateURLPattern
	)

	mux.HandleFunc("POST /tax", svc.AddOrUpdateTaxRecordHandler)
	mux.HandleFunc(fmt.Sprintf("GET /tax/{%s}/{%s}", municipalityNameWildcard, dateWildcard), svc.GetTaxRateHandler)
}
