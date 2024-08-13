package store

const (
	sqlCreateMunicipalityTaxesTable = `
	CREATE TABLE IF NOT EXISTS municipality_taxes (
		id SERIAL PRIMARY KEY,
		municipality_name TEXT NOT NULL,
		tax_rate FLOAT NOT NULL,
		period DATERANGE NOT NULL,
		period_type TEXT NOT NULL CHECK (period_type IN ('yearly', 'monthly', 'weekly', 'daily')),
		UNIQUE (municipality_name, period, period_type)
	)`
	sqlCreateIndexes = `
	CREATE INDEX IF NOT EXISTS idx_municipality_name ON municipality_taxes(municipality_name);
	CREATE INDEX IF NOT EXISTS idx_period ON municipality_taxes USING GIST (period);`

	sqlInsertOrUpdateTaxRecord = `
	INSERT INTO municipality_taxes (municipality_name, tax_rate, period, period_type)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (municipality_name, period, period_type)
	DO UPDATE SET tax_rate = EXCLUDED.tax_rate, period_type = EXCLUDED.period_type`

	sqlSelectTaxRecords = `
	SELECT municipality_name, tax_rate, period, period_type
	FROM municipality_taxes
	WHERE municipality_name = $1
	AND $2 <@ period;
	`

	sqlTruncateMunicipalityTaxesTable = `TRUNCATE TABLE municipality_taxes;`
)
