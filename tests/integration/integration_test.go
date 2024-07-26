package integration

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rezkam/TaxMan/internal/routes"
	"github.com/rezkam/TaxMan/store"
	"github.com/rezkam/TaxMan/taxservice"
	"github.com/stretchr/testify/require"
)

const municipalityWildcardName = "municipality"
const dateWildcardName = "date"

var postgresStore *store.PostgresStore

func TestMain(m *testing.M) {
	connectString := os.Getenv("TEST_DB_URL")
	if connectString == "" {
		slog.Error("TEST_DB_URL is not set. Database connection is needed for integration tests.")
		os.Exit(1)
	}

	var err error
	postgresStore, err = store.NewPostgresStore(connectString)
	if err != nil {
		slog.Error("failed to create postgres store", "error", err)
		os.Exit(1)
	}

	// Run tests
	exitCode := m.Run()

	// Cleanup
	postgresStore.Close()

	os.Exit(exitCode)
}

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	svc, err := taxservice.New(postgresStore, taxservice.Config{
		MunicipalityURLPattern:    municipalityWildcardName,
		DateURLPattern:            dateWildcardName,
		MaxMunicipalityNameLength: 100,
	})
	require.NoError(t, err)

	mux := http.NewServeMux()
	routes.SetupTaxRoutes(svc, mux)

	return httptest.NewServer(mux)
}

func cleanupDatabase(t *testing.T) {
	t.Helper()
	err := postgresStore.CleanupDB()
	require.NoError(t, err, "failed to clean up database")
}

func TestAddOrUpdateTaxRecord(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	cleanupDatabase(t)

	t.Run("success", func(t *testing.T) {
		reqBody, err := json.Marshal(taxservice.AddOrUpdateTaxRecordRequest{
			Municipality: "Copenhagen",
			TaxRate:      0.2,
			StartDate:    "2024-01-01",
			EndDate:      "2024-12-31",
			PeriodType:   "yearly",
		})
		require.NoError(t, err)
		resp, err := http.Post(ts.URL+"/tax", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody taxservice.AddOrUpdateTaxRecordResponse
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)
		require.True(t, respBody.Success)
	})

	t.Run("invalid municipality", func(t *testing.T) {
		reqBody, err := json.Marshal(taxservice.AddOrUpdateTaxRecordRequest{
			Municipality: "",
			TaxRate:      0.1,
			StartDate:    "2024-01-01",
			EndDate:      "2024-12-31",
			PeriodType:   "yearly",
		})
		require.NoError(t, err)
		resp, err := http.Post(ts.URL+"/tax", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid date format", func(t *testing.T) {
		reqBody, err := json.Marshal(taxservice.AddOrUpdateTaxRecordRequest{
			Municipality: "Copenhagen",
			TaxRate:      0.1,
			StartDate:    "invalid-date",
			EndDate:      "2024-12-31",
			PeriodType:   "yearly",
		})
		require.NoError(t, err)
		resp, err := http.Post(ts.URL+"/tax", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestGetTaxRate(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	cleanupDatabase(t)

	// Add tax records to the database for testing
	records := []taxservice.AddOrUpdateTaxRecordRequest{
		{
			Municipality: "Copenhagen",
			TaxRate:      0.2,
			StartDate:    "2024-01-01",
			EndDate:      "2024-12-31",
			PeriodType:   "yearly",
		},
		{
			Municipality: "Copenhagen",
			TaxRate:      0.4,
			StartDate:    "2024-05-01",
			EndDate:      "2024-05-31",
			PeriodType:   "monthly",
		},
		{
			Municipality: "Copenhagen",
			TaxRate:      0.1,
			StartDate:    "2024-01-01",
			EndDate:      "2024-01-01",
			PeriodType:   "daily",
		},
		{
			Municipality: "Copenhagen",
			TaxRate:      0.1,
			StartDate:    "2024-12-25",
			EndDate:      "2024-12-25",
			PeriodType:   "daily",
		},
	}

	for _, record := range records {
		reqBody, err := json.Marshal(record)
		require.NoError(t, err)
		resp, err := http.Post(ts.URL+"/tax", "application/json", bytes.NewReader(reqBody))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	t.Run("get tax rate success", func(t *testing.T) {
		testCases := []struct {
			municipality string
			date         string
			expectedRate float64
		}{
			{"Copenhagen", "2024-01-01", 0.1},
			{"Copenhagen", "2024-03-16", 0.2},
			{"Copenhagen", "2024-05-02", 0.4},
			{"Copenhagen", "2024-07-10", 0.2},
		}

		for _, tc := range testCases {
			t.Run(tc.date, func(t *testing.T) {
				resp, err := http.Get(ts.URL + "/tax/" + tc.municipality + "/" + tc.date)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var respBody taxservice.GetTaxRateResponse
				err = json.NewDecoder(resp.Body).Decode(&respBody)
				require.NoError(t, err)
				require.Equal(t, tc.expectedRate, respBody.TaxRate)
			})
		}
	})

	t.Run("municipality not found", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/tax/NonExistent/2024-01-01")
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid date format", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/tax/Copenhagen/invalid-date")
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestUpdateExistingTaxRecord(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	cleanupDatabase(t)

	// Add an initial tax record to the database
	initialRecord := taxservice.AddOrUpdateTaxRecordRequest{
		Municipality: "Copenhagen",
		TaxRate:      0.1,
		StartDate:    "2024-01-01",
		EndDate:      "2024-12-31",
		PeriodType:   "yearly",
	}
	reqBody, err := json.Marshal(initialRecord)
	require.NoError(t, err)
	resp, err := http.Post(ts.URL+"/tax", "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Update the tax record with a new tax rate
	updatedRecord := taxservice.AddOrUpdateTaxRecordRequest{
		Municipality: "Copenhagen",
		TaxRate:      0.15,
		StartDate:    "2024-01-01",
		EndDate:      "2024-12-31",
		PeriodType:   "yearly",
	}
	reqBody, err = json.Marshal(updatedRecord)
	require.NoError(t, err)
	resp, err = http.Post(ts.URL+"/tax", "application/json", bytes.NewReader(reqBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Fetch the tax rate to verify it has been updated
	resp, err = http.Get(ts.URL + "/tax/Copenhagen/2024-06-01")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody taxservice.GetTaxRateResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)
	require.Equal(t, 0.15, respBody.TaxRate)
}
