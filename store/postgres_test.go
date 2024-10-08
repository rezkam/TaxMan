package store_test

import (
	"context"
	"fmt"
	"github.com/rezkam/TaxMan/internal/utils"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq" // Required for PostgresSQL driver
	"github.com/rezkam/TaxMan/model"
	"github.com/rezkam/TaxMan/store"
	"github.com/stretchr/testify/require"
)

var testStore *store.PostgresStore

// TestMain handles setup and teardown for all tests.
func TestMain(m *testing.M) {
	// Setup code
	var testDBURL = os.Getenv("TEST_DB_URL")
	if testDBURL == "" {
		fmt.Println("TEST_DB_URL environment variable is not set.")
		os.Exit(0) // Skip tests if TEST_DB_URL is not set
	}

	var err error
	testStore, err = store.NewPostgresStore(testDBURL)
	if err != nil {
		fmt.Printf("Failed to create new PostgresStore: %v\n", err)
		os.Exit(1) // Exit if database setup fails
	}

	cleanupDB(nil, testStore) // Clean up the database before starting tests

	// Run tests
	exitCode := m.Run()

	// Teardown code
	if err := testStore.Close(); err != nil {
		fmt.Printf("Failed to close PostgresStore: %v\n", err)
	}

	// Exit with the proper code
	os.Exit(exitCode)
}

// cleanupDB cleans up the test database by removing all data from the tables.
func cleanupDB(t *testing.T, store *store.PostgresStore) {
	err := store.CleanupDB()
	if t != nil {
		t.Helper()
		require.NoError(t, err, "failed to clean up database")
	} else if err != nil {
		panic(err)
	}
}

func TestAddOrUpdateTaxRecord(t *testing.T) {
	cleanupDB(t, testStore)

	const municipality = "Copenhagen"

	t.Run("success", func(t *testing.T) {
		record := model.TaxRecord{
			Municipality: municipality,
			TaxRate:      0.1,
			StartDate:    utils.DateOnly(2024, time.January, 1),
			EndDate:      utils.DateOnly(2024, time.January, 1),
			PeriodType:   "daily",
		}

		err := testStore.AddOrUpdateTaxRecord(context.Background(), record)
		require.NoError(t, err)
	})

	t.Run("update existing record", func(t *testing.T) {
		record := model.TaxRecord{
			Municipality: municipality,
			TaxRate:      0.2,
			StartDate:    utils.DateOnly(2024, time.March, 16),
			EndDate:      utils.DateOnly(2024, time.March, 16),
			PeriodType:   "daily",
		}

		err := testStore.AddOrUpdateTaxRecord(context.Background(), record)
		require.NoError(t, err)
	})
}

func TestGetTaxRate(t *testing.T) {
	cleanupDB(t, testStore)

	const municipality = "Copenhagen"

	// sample records to insert for testing
	records := []model.TaxRecord{
		{
			Municipality: municipality,
			TaxRate:      0.2,
			StartDate:    utils.DateOnly(2024, time.January, 1),
			EndDate:      utils.DateOnly(2024, time.December, 31),
			PeriodType:   "yearly",
		},
		{
			Municipality: municipality,
			TaxRate:      0.4,
			StartDate:    utils.DateOnly(2024, time.May, 1),
			EndDate:      utils.DateOnly(2024, time.May, 31),
			PeriodType:   "monthly",
		},
		{
			Municipality: municipality,
			TaxRate:      0.1,
			StartDate:    utils.DateOnly(2024, time.January, 1),
			EndDate:      utils.DateOnly(2024, time.January, 1),
			PeriodType:   "daily",
		},
		{
			Municipality: municipality,
			TaxRate:      0.1,
			StartDate:    utils.DateOnly(2024, time.December, 25),
			EndDate:      utils.DateOnly(2024, time.December, 25),
			PeriodType:   "daily",
		},
	}

	for _, record := range records {
		err := testStore.AddOrUpdateTaxRecord(context.Background(), record)
		require.NoError(t, err)
	}

	t.Run("success", func(t *testing.T) {
		testCases := []struct {
			Municipality  string
			Date          time.Time
			ExpectedRates []float64
		}{
			{municipality, utils.DateOnly(2024, time.January, 1), []float64{0.1, 0.2}},
			{municipality, utils.DateOnly(2024, time.March, 16), []float64{0.2}},
			{municipality, utils.DateOnly(2024, time.May, 2), []float64{0.2, 0.4}},
			{municipality, utils.DateOnly(2024, time.July, 10), []float64{0.2}},
		}

		for _, tc := range testCases {
			t.Run(tc.Date.String(), func(t *testing.T) {
				query := model.TaxQuery{
					Municipality: tc.Municipality,
					Date:         tc.Date,
				}

				records, err := testStore.GetTaxRecords(context.Background(), query)
				require.NoError(t, err)
				require.Len(t, records, len(tc.ExpectedRates))

				for _, record := range records {
					require.Contains(t, tc.ExpectedRates, record.TaxRate)
				}
			})
		}
	})
}
