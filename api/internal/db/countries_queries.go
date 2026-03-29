package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func (s *Store) FetchCountryListByCodes(ctx context.Context, codes []string) ([]CountryDetails, error) {
	if len(codes) == 0 {
		return []CountryDetails{}, nil
	}
	rows := []CountryDetails{}
	normalized := make([]string, 0, len(codes))
	for _, code := range codes {
		normalized = append(normalized, strings.ToUpper(code))
	}
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
	MAX(cr.continent_name) AS continent_name,
	COALESCE(SUM(cds.moves_count), 0)::bigint AS moves_count,
	COALESCE(SUM(cds.unique_users), 0)::bigint AS unique_users,
	COALESCE(SUM(cds.unique_gks), 0)::bigint AS unique_gks,
	COALESCE(SUM(cds.km_contributed), 0)::double precision AS km_contributed,
	COALESCE(SUM(cds.points_contributed), 0)::double precision AS points_contributed,
	(
		SELECT COUNT(*)::bigint
		FROM geokrety.gk_geokrety_with_details AS g
		WHERE UPPER(g.country) = UPPER(cds.country_code)
	) AS current_geokrety,
	MAX(cds.stats_date)::timestamp AS last_stats_date
FROM stats.country_daily_stats AS cds
LEFT JOIN names AS n ON n.code = UPPER(cds.country_code)
LEFT JOIN stats.continent_reference AS cr ON cr.country_alpha2 = UPPER(cds.country_code)::bpchar
WHERE UPPER(cds.country_code) IN (?)
GROUP BY UPPER(cds.country_code)
`, normalized)
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
	return reorderCountriesByCode(rows, normalized), nil
}

func reorderCountriesByCode(rows []CountryDetails, codes []string) []CountryDetails {
	byCode := make(map[string]CountryDetails, len(rows))
	for _, row := range rows {
		byCode[row.Code] = row
	}
	ordered := make([]CountryDetails, 0, len(rows))
	for _, code := range codes {
		if row, ok := byCode[code]; ok {
			ordered = append(ordered, row)
		}
	}
	return ordered
}
