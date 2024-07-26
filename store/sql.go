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

	sqlSelectTaxRate = `
	SELECT tax_rate
	FROM municipality_taxes
	WHERE municipality_name = $1
	AND $2 <@ period
	ORDER BY
		CASE
			WHEN period_type = 'daily' THEN 1
			WHEN period_type = 'weekly' THEN 2
			WHEN period_type = 'monthly' THEN 3
			WHEN period_type = 'yearly' THEN 4
		END,
		tax_rate DESC
	LIMIT 1;`

	sqlTruncateMunicipalityTaxesTable = `TRUNCATE TABLE municipality_taxes;`
)
