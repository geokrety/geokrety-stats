package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func (s *Store) FetchCountryList(ctx context.Context, filters CountryListFilters, limit, offset int) ([]CountryDetails, error) {
	rows := []CountryDetails{}
	query := `
WITH names AS (
	SELECT
		UPPER(original) AS code,
		MIN(country) AS name
	FROM geokrety.gk_waypoints_country
	GROUP BY UPPER(original)
)
SELECT
	UPPER(cds.country_code) AS code,
	COALESCE(MIN(n.name), UPPER(cds.country_code)) AS name,
	MAX(cr.continent_code) AS continent_code,
	MAX(cr.continent_name) AS continent_name
FROM stats.country_daily_stats AS cds
LEFT JOIN names AS n ON n.code = UPPER(cds.country_code)
LEFT JOIN stats.continent_reference AS cr ON cr.country_alpha2 = UPPER(cds.country_code)::bpchar
GROUP BY UPPER(cds.country_code)
ORDER BY ` + countryOrderBy(filters.Sort) + `
LIMIT $1 OFFSET $2
`
	if err := s.db.SelectContext(ctx, &rows, query, limit, offset); err != nil {
		return nil, fmt.Errorf("query country list: %w", err)
	}
	for i := range rows {
		rows[i].Code = strings.ToUpper(rows[i].Code)
		rows[i].Flag = countryFlag(rows[i].Code)
	}
	return rows, nil
}

func (s *Store) FetchCountryListByCodes(ctx context.Context, codes []string) ([]CountryDetails, error) {
	if len(codes) == 0 {
		return []CountryDetails{}, nil
	}
	rows := []CountryDetails{}
	query, args, err := sqlx.In(`
WITH names AS (
	SELECT
		UPPER(original) AS code,
		MIN(country) AS name
	FROM geokrety.gk_waypoints_country
	GROUP BY UPPER(original)
)
SELECT
	UPPER(cds.country_code) AS code,
	COALESCE(MIN(n.name), UPPER(cds.country_code)) AS name,
	MAX(cr.continent_code) AS continent_code,
	MAX(cr.continent_name) AS continent_name
FROM stats.country_daily_stats AS cds
LEFT JOIN names AS n ON n.code = UPPER(cds.country_code)
LEFT JOIN stats.continent_reference AS cr ON cr.country_alpha2 = UPPER(cds.country_code)::bpchar
WHERE UPPER(cds.country_code) IN (?)
GROUP BY UPPER(cds.country_code)
`, codes)
	if err != nil {
		return nil, fmt.Errorf("build country ids query: %w", err)
	}
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("query countries by ids: %w", err)
	}
	for i := range rows {
		rows[i].Code = strings.ToUpper(rows[i].Code)
		rows[i].Flag = countryFlag(rows[i].Code)
	}
	return reorderCountriesByCode(rows, codes), nil
}

func (s *Store) FetchCountryDetails(ctx context.Context, countryCode string) (CountryDetails, error) {
	row := CountryDetails{}
	if err := s.db.GetContext(ctx, &row, `
WITH base AS (
	SELECT
		UPPER($1) AS code,
		MIN(wc.country) AS name,
		MAX(cr.continent_code) AS continent_code,
		MAX(cr.continent_name) AS continent_name
	FROM stats.country_daily_stats AS cds
	LEFT JOIN geokrety.gk_waypoints_country AS wc ON UPPER(wc.original) = UPPER($1)
	LEFT JOIN stats.continent_reference AS cr ON cr.country_alpha2 = UPPER($1)::bpchar
	WHERE UPPER(cds.country_code) = UPPER($1)
)
SELECT
	code,
	COALESCE(name, code) AS name,
	continent_code,
	continent_name
FROM base
`, countryCode); err != nil {
		return CountryDetails{}, fmt.Errorf("query country details: %w", err)
	}
	row.Code = strings.ToUpper(row.Code)
	row.Flag = countryFlag(row.Code)
	return row, nil
}

func (s *Store) FetchCountryGeokrety(ctx context.Context, countryCode string, sort Sort, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	query := geokretSelectColumns() + geokretBaseFromClause() + `
WHERE UPPER(g.country) = UPPER($1)
ORDER BY ` + geokretOrderBy(sort) + `
LIMIT $2 OFFSET $3
`
	if err := s.db.SelectContext(ctx, &rows, query, countryCode, limit, offset); err != nil {
		return nil, fmt.Errorf("query country geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func reorderCountriesByCode(rows []CountryDetails, codes []string) []CountryDetails {
	byCode := make(map[string]CountryDetails, len(rows))
	for _, row := range rows {
		byCode[strings.ToUpper(row.Code)] = row
	}
	ordered := make([]CountryDetails, 0, len(rows))
	for _, code := range codes {
		if row, ok := byCode[strings.ToUpper(code)]; ok {
			ordered = append(ordered, row)
		}
	}
	return ordered
}
