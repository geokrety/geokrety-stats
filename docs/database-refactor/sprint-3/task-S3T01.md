---
title: "Task S3T01: Create stats.country_daily_stats Table"
version: 1.0
date_created: 2026-03-08
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 3
tags:
  - country
  - database
  - geography
  - migration
  - postgresql
  - schema
  - sprint-3
  - stats
  - task-merge
  - traversal
depends_on: [1, 2]
task: S3T01
step: 3.1
migration: 20260310300000_create_country_daily_stats.php
blocks: [5, 6]
changelog:
  - 2026-03-10: created by merge of 03-sprint-3-country-geography.md step 3.1
---

# Task S3T01: Create stats.country_daily_stats Table

## Master-Spec Alignment

The normative contract for this sprint is [00-SPEC-DRAFT-v1.obsolete.md](../00-SPEC-DRAFT-v1.obsolete.md), Sections 5.3, 5.4, 8.4, 9.2, and 11.

- `stats.country_daily_stats.unique_users` and `unique_gks` are exact online-maintained values, not approximate placeholders.
- `INSERT`, `UPDATE`, and `DELETE` handling for `stats.gk_countries_visited`, `stats.user_countries`, and `stats.gk_country_history` must maintain exact state. When earliest/latest rows are invalidated, affected rows must be recomputed from remaining qualifying moves.
- Snapshot functions seed and verify canonical state; they do not compensate for knowingly inexact live maintenance.
- Any lower text that still describes `unique_users` or `unique_gks` as approximate is obsolete and superseded by this alignment block.

## Source

- Generated from sprint document step `3.1` in `03-sprint-3-country-geography.md`.

### Step 3.1: Create `stats.country_daily_stats` Table

**What this step does:** Creates the `stats.country_daily_stats` table that stores daily per-country aggregate statistics. This table is the primary source for country choropleth maps, country leaderboards, and country-level time series charts. The composite primary key `(stats_date, country_code)` supports efficient date-range queries partitioned by country.

**Migration file name:** `20260310300000_create_country_daily_stats.php`

#### Full SQL DDL

```sql
CREATE TABLE stats.country_daily_stats (
  stats_date DATE NOT NULL,
  country_code CHAR(2) NOT NULL,
  moves_count BIGINT NOT NULL DEFAULT 0,
  drops BIGINT NOT NULL DEFAULT 0,
  grabs BIGINT NOT NULL DEFAULT 0,
  comments BIGINT NOT NULL DEFAULT 0,
  sees BIGINT NOT NULL DEFAULT 0,
  archives BIGINT NOT NULL DEFAULT 0,
  dips BIGINT NOT NULL DEFAULT 0,
  unique_users BIGINT NOT NULL DEFAULT 0,
  unique_gks BIGINT NOT NULL DEFAULT 0,
  km_contributed NUMERIC(14,3) NOT NULL DEFAULT 0,
  points_contributed NUMERIC(16,4) NOT NULL DEFAULT 0,
  loves_count BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_total BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_avatar BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_move BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_user BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (stats_date, country_code)
);

COMMENT ON TABLE stats.country_daily_stats IS 'Daily per-country aggregate statistics for moves, distance, users, GKs, and content';
COMMENT ON COLUMN stats.country_daily_stats.unique_users IS 'Exact distinct user count maintained online for the date/country bucket';
COMMENT ON COLUMN stats.country_daily_stats.unique_gks IS 'Exact distinct GK count maintained online for the date/country bucket';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateCountryDailyStats extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.country_daily_stats (
  stats_date DATE NOT NULL,
  country_code CHAR(2) NOT NULL,
  moves_count BIGINT NOT NULL DEFAULT 0,
  drops BIGINT NOT NULL DEFAULT 0,
  grabs BIGINT NOT NULL DEFAULT 0,
  comments BIGINT NOT NULL DEFAULT 0,
  sees BIGINT NOT NULL DEFAULT 0,
  archives BIGINT NOT NULL DEFAULT 0,
  dips BIGINT NOT NULL DEFAULT 0,
  unique_users BIGINT NOT NULL DEFAULT 0,
  unique_gks BIGINT NOT NULL DEFAULT 0,
  km_contributed NUMERIC(14,3) NOT NULL DEFAULT 0,
  points_contributed NUMERIC(16,4) NOT NULL DEFAULT 0,
  loves_count BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_total BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_avatar BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_move BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_user BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (stats_date, country_code)
);

COMMENT ON TABLE stats.country_daily_stats IS 'Daily per-country aggregate statistics for moves, distance, users, GKs, and content';
COMMENT ON COLUMN stats.country_daily_stats.unique_users IS 'Exact distinct user count maintained online for the date/country bucket';
COMMENT ON COLUMN stats.country_daily_stats.unique_gks IS 'Exact distinct GK count maintained online for the date/country bucket';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.country_daily_stats;');
    }
}
```

#### SQL Usage Examples

```sql
-- Top 10 countries by total moves (all time)
SELECT country_code, SUM(moves_count) AS total_moves
FROM stats.country_daily_stats
GROUP BY country_code
ORDER BY total_moves DESC
LIMIT 10;

-- Daily move trend for Poland in 2025
SELECT stats_date, moves_count, km_contributed
FROM stats.country_daily_stats
WHERE country_code = 'PL'
  AND stats_date BETWEEN '2025-01-01' AND '2025-12-31'
ORDER BY stats_date;

-- Country leaderboard by km for current month
SELECT country_code, SUM(km_contributed) AS total_km, SUM(moves_count) AS total_moves
FROM stats.country_daily_stats
WHERE stats_date >= date_trunc('month', CURRENT_DATE)
GROUP BY country_code
ORDER BY total_km DESC
LIMIT 20;

-- Move-type breakdown for Germany last 30 days
SELECT SUM(drops) AS drops, SUM(grabs) AS grabs, SUM(sees) AS sees,
       SUM(dips) AS dips, SUM(comments) AS comments, SUM(archives) AS archives
FROM stats.country_daily_stats
WHERE country_code = 'DE'
  AND stats_date >= CURRENT_DATE - INTERVAL '30 days';
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Choropleth world map — heatmap of `moves_count` or `km_contributed` by country
- **Data source:** `SELECT country_code, SUM(moves_count) FROM stats.country_daily_stats GROUP BY country_code`
- **Color scale:** Sequential green (D3 `interpolateGreens`) with `scaleSequentialLog`

- **Chart type:** Stacked area chart — daily moves by type per country
- **X-axis:** `stats_date`
- **Y-axis:** `drops`, `grabs`, `sees`, `dips`, `comments`, `archives`
- **Filter:** Country code selector

```
ASCII Sample (Top Countries by Moves):
PL  |████████████████████████████████████████| 3.2M
DE  |████████████████████                    | 1.5M
CZ  |██████████████                          | 1.1M
FR  |████████                                | 0.6M
SK  |██████                                  | 0.5M
```

#### TimescaleDB Assessment

**RECOMMENDED (conditional).** Rationale:

- `country_daily_stats` is an append-heavy time-series table keyed by `(stats_date, country_code)`. With ~250 active countries and 15+ years of data, this table will grow to ~1.3M rows.
- TimescaleDB hypertable on `stats_date` would enable automatic chunk-based partitioning, transparent time-range pruning, and built-in retention policies.
- **However**, at ~1.3M rows, standard PostgreSQL with appropriate indexes handles this volume efficiently. Hypertable conversion is beneficial only if the table grows significantly larger (>10M rows) or if automated retention/compression policies are needed.
- **Recommendation:** Use standard PostgreSQL table now. Revisit hypertable conversion if TimescaleDB is installed and row count exceeds 5M.

#### pgTAP Unit Tests

| Test ID   | Test Name                                | Assertion                                                                                                                  |
| --------- | ---------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| T-3.1.001 | country_daily_stats table exists         | `has_table('stats', 'country_daily_stats')`                                                                                |
| T-3.1.002 | PK is (stats_date, country_code)         | `col_is_pk('stats', 'country_daily_stats', ARRAY['stats_date', 'country_code'])`                                           |
| T-3.1.003 | stats_date column type is date           | `col_type_is('stats', 'country_daily_stats', 'stats_date', 'date')`                                                        |
| T-3.1.004 | country_code column type is char(2)      | `col_type_is('stats', 'country_daily_stats', 'country_code', 'character(2)')`                                              |
| T-3.1.005 | moves_count default is 0                 | `col_default_is('stats', 'country_daily_stats', 'moves_count', '0')`                                                       |
| T-3.1.006 | km_contributed type is numeric(14,3)     | `col_type_is('stats', 'country_daily_stats', 'km_contributed', 'numeric(14,3)')`                                           |
| T-3.1.007 | points_contributed type is numeric(16,4) | `col_type_is('stats', 'country_daily_stats', 'points_contributed', 'numeric(16,4)')`                                       |
| T-3.1.008 | All 18 columns exist                     | `SELECT COUNT(*) = 18 FROM information_schema.columns WHERE table_schema = 'stats' AND table_name = 'country_daily_stats'` |
| T-3.1.009 | Insert and read-back succeeds            | Insert `('2025-06-15', 'PL', 10, ...)` and verify `SELECT moves_count = 10`                                                |
| T-3.1.010 | Duplicate PK raises error                | Insert same `(stats_date, country_code)` twice — `throws_ok`                                                               |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310300000_create_country_daily_stats.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify table exists with 18 columns and correct types
- [ ] 4. Verify composite PK on `(stats_date, country_code)`
- [ ] 5. Test insert and read-back
- [ ] 6. Run pgTAP tests T-3.1.001 through T-3.1.010

---
