---
title: "Task S3T07: Create Country Snapshot/Seed Functions"
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
task: S3T07
step: 3.7
migration: 20260310300600_create_country_snapshot_functions.php
blocks: [5, 6]
changelog:
  - 2026-03-10: created by merge of 03-sprint-3-country-geography.md step 3.7
  - 2026-03-10: clarified the move-derived snapshot scope and logged the remaining backfill gap
  - 2026-03-10: documented the unresolved stale-row cleanup contract for partial reseeds
---

# Task S3T07: Create Country Snapshot/Seed Functions

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 3 task set in `docs/database-refactor/sprint-3/`. `../00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- `stats.country_daily_stats.unique_users` and `unique_gks` are exact online-maintained values, not approximate placeholders.
- `INSERT`, `UPDATE`, and `DELETE` handling for `stats.gk_countries_visited`, `stats.user_countries`, and `stats.gk_country_history` must maintain exact state. When earliest/latest rows are invalidated, affected rows must be recomputed from remaining qualifying moves.
- Snapshot functions seed and verify canonical state; they do not compensate for knowingly inexact live maintenance.
- Any lower text that still describes `unique_users` or `unique_gks` as approximate is obsolete and superseded by this alignment block.

## Source

- Generated from sprint document step `3.7` in `03-sprint-3-country-geography.md`.

### Step 3.7: Create Country Snapshot/Seed Functions

**What this step does:** Creates three idempotent snapshot functions that seed the move-derived country stats state from historical `gk_moves` data. These functions are used during backfill (Sprint 6) to populate the tables from the ~6.9M existing rows, and can be re-run at any time for correction or verification. All use `ON CONFLICT DO UPDATE` for idempotency and accept an optional `p_period daterange` parameter for partial reseeding of the move-derived contract.

Resolved policy for this task:

- `stats.fn_snapshot_daily_country_stats()` owns the move-derived columns in `stats.country_daily_stats`: move counts, per-type counts, `unique_users`, `unique_gks`, and `km_contributed`.
- `points_contributed` remains out of scope for Sprint 3 backfills and stays at the table default (`0`) until the points phase defines its rebuild contract.
- `loves_count` is owned by canonical source `geokrety.gk_loves`.
- `pictures_uploaded_total`, `pictures_uploaded_avatar`, `pictures_uploaded_move`, and `pictures_uploaded_user` are owned by canonical source `geokrety.gk_pictures`.
- Partial `p_period` reseeds leave rows outside the supplied date range untouched.
- Full reseed (`p_period IS NULL`) is an orchestration concern: rebuild all country stats columns from their canonical sources in sequence, rather than treating the move-derived snapshot as the sole source of truth for non-move columns.

**Migration file name:** `20260310300600_create_country_snapshot_functions.php`

#### Full SQL DDL

```sql
-- ============================================================
-- Snapshot 1: Seed country_daily_stats from gk_moves
-- ============================================================
CREATE OR REPLACE FUNCTION stats.fn_snapshot_daily_country_stats(
  p_period daterange DEFAULT NULL
)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
  v_count BIGINT := 0;
BEGIN
  INSERT INTO stats.country_daily_stats (
    stats_date, country_code,
    moves_count, drops, grabs, comments, sees, archives, dips,
    unique_users, unique_gks, km_contributed
  )
  SELECT
    m.moved_on_datetime::date AS stats_date,
    m.country AS country_code,
    COUNT(*) AS moves_count,
    COUNT(*) FILTER (WHERE m.move_type = 0) AS drops,
    COUNT(*) FILTER (WHERE m.move_type = 1) AS grabs,
    COUNT(*) FILTER (WHERE m.move_type = 2) AS comments,
    COUNT(*) FILTER (WHERE m.move_type = 3) AS sees,
    COUNT(*) FILTER (WHERE m.move_type = 4) AS archives,
    COUNT(*) FILTER (WHERE m.move_type = 5) AS dips,
    COUNT(DISTINCT m.author) FILTER (WHERE m.author IS NOT NULL) AS unique_users,
    COUNT(DISTINCT m.geokret) AS unique_gks,
    COALESCE(SUM(m.km_distance), 0) AS km_contributed
  FROM geokrety.gk_moves m
  WHERE m.country IS NOT NULL
    AND (p_period IS NULL OR m.moved_on_datetime::date <@ p_period)
  GROUP BY m.moved_on_datetime::date, m.country
  ON CONFLICT (stats_date, country_code) DO UPDATE SET
    moves_count = EXCLUDED.moves_count,
    drops = EXCLUDED.drops,
    grabs = EXCLUDED.grabs,
    comments = EXCLUDED.comments,
    sees = EXCLUDED.sees,
    archives = EXCLUDED.archives,
    dips = EXCLUDED.dips,
    unique_users = EXCLUDED.unique_users,
    unique_gks = EXCLUDED.unique_gks,
    km_contributed = EXCLUDED.km_contributed;

  GET DIAGNOSTICS v_count = ROW_COUNT;
  RETURN v_count;
END;
$$;

COMMENT ON FUNCTION stats.fn_snapshot_daily_country_stats IS 'Seeds country_daily_stats from gk_moves. Idempotent via ON CONFLICT DO UPDATE. Optional p_period limits date range.';

-- ============================================================
-- Snapshot 2: Seed user_countries from gk_moves
-- ============================================================
CREATE OR REPLACE FUNCTION stats.fn_snapshot_user_country_stats(
  p_period daterange DEFAULT NULL
)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
  v_count BIGINT := 0;
BEGIN
  INSERT INTO stats.user_countries (
    user_id, country_code, move_count, first_visit, last_visit
  )
  SELECT
    m.author AS user_id,
    m.country AS country_code,
    COUNT(*) AS move_count,
    MIN(m.moved_on_datetime) AS first_visit,
    MAX(m.moved_on_datetime) AS last_visit
  FROM geokrety.gk_moves m
  WHERE m.country IS NOT NULL
    AND m.author IS NOT NULL
    AND (p_period IS NULL OR m.moved_on_datetime::date <@ p_period)
  GROUP BY m.author, m.country
  ON CONFLICT (user_id, country_code) DO UPDATE SET
    move_count = EXCLUDED.move_count,
    first_visit = LEAST(stats.user_countries.first_visit, EXCLUDED.first_visit),
    last_visit = GREATEST(stats.user_countries.last_visit, EXCLUDED.last_visit);

  GET DIAGNOSTICS v_count = ROW_COUNT;
  RETURN v_count;
END;
$$;

COMMENT ON FUNCTION stats.fn_snapshot_user_country_stats IS 'Seeds user_countries from gk_moves. Idempotent via ON CONFLICT DO UPDATE. Optional p_period limits date range.';

-- ============================================================
-- Snapshot 3: Seed gk_countries_visited from gk_moves
-- ============================================================
CREATE OR REPLACE FUNCTION stats.fn_snapshot_gk_country_stats(
  p_period daterange DEFAULT NULL
)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
  v_count BIGINT := 0;
BEGIN
  INSERT INTO stats.gk_countries_visited (
    geokrety_id, country_code, first_visited_at, first_move_id, move_count
  )
  SELECT
    sub.geokret AS geokrety_id,
    sub.country AS country_code,
    sub.first_visited_at,
    sub.first_move_id,
    sub.move_count
  FROM (
    SELECT
      m.geokret,
      m.country,
      MIN(m.moved_on_datetime) AS first_visited_at,
      (array_agg(m.id ORDER BY m.moved_on_datetime ASC, m.id ASC))[1] AS first_move_id,
      COUNT(*) AS move_count
    FROM geokrety.gk_moves m
    WHERE m.country IS NOT NULL
      AND (p_period IS NULL OR m.moved_on_datetime::date <@ p_period)
    GROUP BY m.geokret, m.country
  ) sub
  ON CONFLICT (geokrety_id, country_code) DO UPDATE SET
    move_count = EXCLUDED.move_count,
    first_visited_at = LEAST(stats.gk_countries_visited.first_visited_at, EXCLUDED.first_visited_at),
    first_move_id = CASE
      WHEN EXCLUDED.first_visited_at < stats.gk_countries_visited.first_visited_at THEN EXCLUDED.first_move_id
      ELSE stats.gk_countries_visited.first_move_id
    END;

  GET DIAGNOSTICS v_count = ROW_COUNT;
  RETURN v_count;
END;
$$;

COMMENT ON FUNCTION stats.fn_snapshot_gk_country_stats IS 'Seeds gk_countries_visited from gk_moves. Idempotent via ON CONFLICT DO UPDATE. Optional p_period limits date range.';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateCountrySnapshotFunctions extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION stats.fn_snapshot_daily_country_stats(
  p_period daterange DEFAULT NULL
)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
  v_count BIGINT := 0;
BEGIN
  INSERT INTO stats.country_daily_stats (
    stats_date, country_code,
    moves_count, drops, grabs, comments, sees, archives, dips,
    unique_users, unique_gks, km_contributed
  )
  SELECT
    m.moved_on_datetime::date AS stats_date,
    m.country AS country_code,
    COUNT(*) AS moves_count,
    COUNT(*) FILTER (WHERE m.move_type = 0) AS drops,
    COUNT(*) FILTER (WHERE m.move_type = 1) AS grabs,
    COUNT(*) FILTER (WHERE m.move_type = 2) AS comments,
    COUNT(*) FILTER (WHERE m.move_type = 3) AS sees,
    COUNT(*) FILTER (WHERE m.move_type = 4) AS archives,
    COUNT(*) FILTER (WHERE m.move_type = 5) AS dips,
    COUNT(DISTINCT m.author) FILTER (WHERE m.author IS NOT NULL) AS unique_users,
    COUNT(DISTINCT m.geokret) AS unique_gks,
    COALESCE(SUM(m.km_distance), 0) AS km_contributed
  FROM geokrety.gk_moves m
  WHERE m.country IS NOT NULL
    AND (p_period IS NULL OR m.moved_on_datetime::date <@ p_period)
  GROUP BY m.moved_on_datetime::date, m.country
  ON CONFLICT (stats_date, country_code) DO UPDATE SET
    moves_count = EXCLUDED.moves_count,
    drops = EXCLUDED.drops,
    grabs = EXCLUDED.grabs,
    comments = EXCLUDED.comments,
    sees = EXCLUDED.sees,
    archives = EXCLUDED.archives,
    dips = EXCLUDED.dips,
    unique_users = EXCLUDED.unique_users,
    unique_gks = EXCLUDED.unique_gks,
    km_contributed = EXCLUDED.km_contributed;

  GET DIAGNOSTICS v_count = ROW_COUNT;
  RETURN v_count;
END;
$$;

COMMENT ON FUNCTION stats.fn_snapshot_daily_country_stats IS 'Seeds country_daily_stats from gk_moves. Idempotent via ON CONFLICT DO UPDATE. Optional p_period limits date range.';

CREATE OR REPLACE FUNCTION stats.fn_snapshot_user_country_stats(
  p_period daterange DEFAULT NULL
)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
  v_count BIGINT := 0;
BEGIN
  INSERT INTO stats.user_countries (
    user_id, country_code, move_count, first_visit, last_visit
  )
  SELECT
    m.author AS user_id,
    m.country AS country_code,
    COUNT(*) AS move_count,
    MIN(m.moved_on_datetime) AS first_visit,
    MAX(m.moved_on_datetime) AS last_visit
  FROM geokrety.gk_moves m
  WHERE m.country IS NOT NULL
    AND m.author IS NOT NULL
    AND (p_period IS NULL OR m.moved_on_datetime::date <@ p_period)
  GROUP BY m.author, m.country
  ON CONFLICT (user_id, country_code) DO UPDATE SET
    move_count = EXCLUDED.move_count,
    first_visit = LEAST(stats.user_countries.first_visit, EXCLUDED.first_visit),
    last_visit = GREATEST(stats.user_countries.last_visit, EXCLUDED.last_visit);

  GET DIAGNOSTICS v_count = ROW_COUNT;
  RETURN v_count;
END;
$$;

COMMENT ON FUNCTION stats.fn_snapshot_user_country_stats IS 'Seeds user_countries from gk_moves. Idempotent via ON CONFLICT DO UPDATE. Optional p_period limits date range.';

CREATE OR REPLACE FUNCTION stats.fn_snapshot_gk_country_stats(
  p_period daterange DEFAULT NULL
)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
  v_count BIGINT := 0;
BEGIN
  INSERT INTO stats.gk_countries_visited (
    geokrety_id, country_code, first_visited_at, first_move_id, move_count
  )
  SELECT
    sub.geokret AS geokrety_id,
    sub.country AS country_code,
    sub.first_visited_at,
    sub.first_move_id,
    sub.move_count
  FROM (
    SELECT
      m.geokret,
      m.country,
      MIN(m.moved_on_datetime) AS first_visited_at,
      (array_agg(m.id ORDER BY m.moved_on_datetime ASC, m.id ASC))[1] AS first_move_id,
      COUNT(*) AS move_count
    FROM geokrety.gk_moves m
    WHERE m.country IS NOT NULL
      AND (p_period IS NULL OR m.moved_on_datetime::date <@ p_period)
    GROUP BY m.geokret, m.country
  ) sub
  ON CONFLICT (geokrety_id, country_code) DO UPDATE SET
    move_count = EXCLUDED.move_count,
    first_visited_at = LEAST(stats.gk_countries_visited.first_visited_at, EXCLUDED.first_visited_at),
    first_move_id = CASE
      WHEN EXCLUDED.first_visited_at < stats.gk_countries_visited.first_visited_at THEN EXCLUDED.first_move_id
      ELSE stats.gk_countries_visited.first_move_id
    END;

  GET DIAGNOSTICS v_count = ROW_COUNT;
  RETURN v_count;
END;
$$;

COMMENT ON FUNCTION stats.fn_snapshot_gk_country_stats IS 'Seeds gk_countries_visited from gk_moves. Idempotent via ON CONFLICT DO UPDATE. Optional p_period limits date range.';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP FUNCTION IF EXISTS stats.fn_snapshot_daily_country_stats(daterange);');
        $this->execute('DROP FUNCTION IF EXISTS stats.fn_snapshot_user_country_stats(daterange);');
        $this->execute('DROP FUNCTION IF EXISTS stats.fn_snapshot_gk_country_stats(daterange);');
    }
}
```

#### SQL Usage Examples

```sql
-- Full snapshot of all country daily stats (all time)
SELECT stats.fn_snapshot_daily_country_stats(NULL);

-- Snapshot only for January 2025
SELECT stats.fn_snapshot_daily_country_stats('[2025-01-01, 2025-02-01)'::daterange);

-- Snapshot user country stats for all time
SELECT stats.fn_snapshot_user_country_stats(NULL);

-- Snapshot GK country stats for a specific quarter
SELECT stats.fn_snapshot_gk_country_stats('[2025-01-01, 2025-04-01)'::daterange);

-- Verify idempotency: run twice, check row counts match
SELECT stats.fn_snapshot_daily_country_stats('[2025-06-01, 2025-07-01)'::daterange) AS first_run;
SELECT stats.fn_snapshot_daily_country_stats('[2025-06-01, 2025-07-01)'::daterange) AS second_run;
-- Both return same count; ON CONFLICT DO UPDATE overwrites identical values
```

#### Graph/Visualization Specification

No new visualization unlocked at this step. These functions populate data for the visualizations defined in Steps 3.1–3.4.

#### TimescaleDB Assessment

**NOT applicable.** This step creates functions, not tables.

#### pgTAP Unit Tests

| Test ID   | Test Name                                      | Assertion                                                                                    |
| --------- | ---------------------------------------------- | -------------------------------------------------------------------------------------------- |
| T-3.7.001 | fn_snapshot_daily_country_stats exists         | `has_function('stats', 'fn_snapshot_daily_country_stats', ARRAY['daterange'])`               |
| T-3.7.002 | fn_snapshot_user_country_stats exists          | `has_function('stats', 'fn_snapshot_user_country_stats', ARRAY['daterange'])`                |
| T-3.7.003 | fn_snapshot_gk_country_stats exists            | `has_function('stats', 'fn_snapshot_gk_country_stats', ARRAY['daterange'])`                  |
| T-3.7.004 | fn_snapshot_daily_country_stats returns bigint | `function_returns('stats', 'fn_snapshot_daily_country_stats', ARRAY['daterange'], 'bigint')` |
| T-3.7.005 | fn_snapshot_user_country_stats returns bigint  | `function_returns('stats', 'fn_snapshot_user_country_stats', ARRAY['daterange'], 'bigint')`  |
| T-3.7.006 | fn_snapshot_gk_country_stats returns bigint    | `function_returns('stats', 'fn_snapshot_gk_country_stats', ARRAY['daterange'], 'bigint')`    |
| T-3.7.007 | Snapshot daily produces correct row count      | Insert 3 test moves (2 in PL, 1 in DE), run snapshot, verify 2 rows in `country_daily_stats` |
| T-3.7.008 | Snapshot is idempotent                         | Run snapshot twice on same data, verify same row count both times                            |
| T-3.7.009 | Snapshot with p_period filters correctly       | Insert moves in Jan and Feb, snapshot for Jan only, verify only Jan rows populated           |
| T-3.7.010 | User snapshot skips anonymous moves            | Insert move with `author=NULL`, run user snapshot, verify no row for NULL user               |
| T-3.7.011 | GK snapshot captures first_move_id correctly   | Insert 3 moves for GK #1 in PL with different IDs, verify `first_move_id` is the earliest    |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310300600_create_country_snapshot_functions.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify all three functions exist with correct signatures
- [ ] 4. Test each function with NULL period (full snapshot)
- [ ] 5. Test each function with a specific date range
- [ ] 6. Test idempotency: run twice with same data
- [ ] 7. Test that anonymous moves are excluded from user snapshot
- [ ] 8. Run pgTAP tests T-3.7.001 through T-3.7.011

## Agent Loop Log

- 2026-03-10T18:16:06Z - Loop 1 - `dba`: Confirmed the existing SQL should remain move-derived only; non-move columns need separate canonical-source rebuild ownership instead of being silently zeroed by this function.
- 2026-03-10T18:16:06Z - Loop 1 - `critical-thinking`: No blocking contradiction; clarified that partial reseeds protect out-of-range rows while full reseeds are orchestration-level rebuilds across all owning sources.
- 2026-03-10T18:16:06Z - Loop 1 - `specification`: Folded the ownership and cleanup policy into the task text, removed the open-question placeholders, and preserved the existing snapshot SQL scope.

## Resolution

- Backfill ownership and cleanup policy are canonicalized in this file; see Q-025 reference update in `99-OPEN-QUESTIONS.md`.

---
