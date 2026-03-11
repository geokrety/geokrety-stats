---
title: "Task S2T12: Create Daily Activity Seed Function"
version: 1.0
date_created: 2026-03-08
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 2
tags:
  - counters
  - daily-activity
  - database
  - migration
  - postgresql
  - schema
  - sprint-2
  - stats
  - task-merge
  - triggers
depends_on: [1]
task: S2T12
step: 2.12
migration: 20260310201100_create_daily_activity_seed.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026.03.10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.12
---

# Task S2T12: Create Daily Activity Seed Function

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `../00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.12` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.12: Create Daily Activity Seed Function

**What this step does:** Creates `stats.fn_seed_daily_activity(p_period tstzrange DEFAULT NULL)` — an idempotent function that back-fills `stats.daily_activity` and `stats.daily_active_users` from historical source table records. This function is called once during Sprint 6 historical backfill. Each call can target a specific time range or seed the entire history. It uses `ON CONFLICT DO UPDATE` throughout for idempotency.

**Migration file name:** `20260310201100_create_daily_activity_seed.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION stats.fn_seed_daily_activity(
  p_period tstzrange DEFAULT NULL
) RETURNS BIGINT LANGUAGE plpgsql AS $$
DECLARE
  v_rows BIGINT := 0;
BEGIN

  -- ============================================================
  -- Seed daily_activity from gk_moves
  -- ============================================================
  INSERT INTO stats.daily_activity (
    activity_date,
    total_moves,
    drops,
    grabs,
    comments,
    sees,
    archives,
    dips,
    km_contributed
  )
  SELECT
    moved_on_datetime::date         AS activity_date,
    COUNT(*)                        AS total_moves,
    COUNT(*) FILTER (WHERE move_type = 0) AS drops,
    COUNT(*) FILTER (WHERE move_type = 1) AS grabs,
    COUNT(*) FILTER (WHERE move_type = 2) AS comments,
    COUNT(*) FILTER (WHERE move_type = 3) AS sees,
    COUNT(*) FILTER (WHERE move_type = 4) AS archives,
    COUNT(*) FILTER (WHERE move_type = 5) AS dips,
    COALESCE(SUM(km_distance), 0)::NUMERIC(15,3) AS km_contributed
  FROM geokrety.gk_moves
  WHERE (p_period IS NULL OR moved_on_datetime <@ p_period)
  GROUP BY moved_on_datetime::date
  ON CONFLICT (activity_date) DO UPDATE SET
    total_moves    = EXCLUDED.total_moves,
    drops          = EXCLUDED.drops,
    grabs          = EXCLUDED.grabs,
    comments       = EXCLUDED.comments,
    sees           = EXCLUDED.sees,
    archives       = EXCLUDED.archives,
    dips           = EXCLUDED.dips,
    km_contributed = EXCLUDED.km_contributed;

  GET DIAGNOSTICS v_rows = ROW_COUNT;

  -- ============================================================
  -- Seed daily_activity from gk_geokrety (gk_created)
  -- ============================================================
  INSERT INTO stats.daily_activity (activity_date, gk_created)
  SELECT
    created_on_datetime::date AS activity_date,
    COUNT(*) AS gk_created
  FROM geokrety.gk_geokrety
  WHERE (p_period IS NULL OR created_on_datetime <@ p_period)
  GROUP BY created_on_datetime::date
  ON CONFLICT (activity_date) DO UPDATE SET
    gk_created = EXCLUDED.gk_created;

  -- ============================================================
  -- Seed daily_activity from gk_pictures
  -- ============================================================
  INSERT INTO stats.daily_activity (
    activity_date,
    pictures_uploaded_total,
    pictures_uploaded_avatar,
    pictures_uploaded_move,
    pictures_uploaded_user
  )
  SELECT
    created_on_datetime::date AS activity_date,
    COUNT(*) AS pictures_uploaded_total,
    COUNT(*) FILTER (WHERE type = 0) AS pictures_uploaded_avatar,
    COUNT(*) FILTER (WHERE type = 1) AS pictures_uploaded_move,
    COUNT(*) FILTER (WHERE type = 2) AS pictures_uploaded_user
  FROM geokrety.gk_pictures
  WHERE (p_period IS NULL OR created_on_datetime <@ p_period)
  GROUP BY created_on_datetime::date
  ON CONFLICT (activity_date) DO UPDATE SET
    pictures_uploaded_total  = EXCLUDED.pictures_uploaded_total,
    pictures_uploaded_avatar = EXCLUDED.pictures_uploaded_avatar,
    pictures_uploaded_move   = EXCLUDED.pictures_uploaded_move,
    pictures_uploaded_user   = EXCLUDED.pictures_uploaded_user;

  -- ============================================================
  -- Seed daily_activity from gk_users (users_registered)
  -- ============================================================
  INSERT INTO stats.daily_activity (activity_date, users_registered)
  SELECT
    joined_on_datetime::date AS activity_date,
    COUNT(*) AS users_registered
  FROM geokrety.gk_users
  WHERE (p_period IS NULL OR joined_on_datetime <@ p_period)
  GROUP BY joined_on_datetime::date
  ON CONFLICT (activity_date) DO UPDATE SET
    users_registered = EXCLUDED.users_registered;

  -- ============================================================
  -- Seed daily_active_users from gk_moves (unique authors per day)
  -- ============================================================
  INSERT INTO stats.daily_active_users (activity_date, user_id)
  SELECT DISTINCT
    moved_on_datetime::date AS activity_date,
    author
  FROM geokrety.gk_moves
  WHERE author IS NOT NULL
    AND (p_period IS NULL OR moved_on_datetime <@ p_period)
  ON CONFLICT (activity_date, user_id) DO NOTHING;

  RAISE NOTICE 'Daily activity seed completed: % rows in daily_activity affected', v_rows;
  RETURN v_rows;
END;
$$;

COMMENT ON FUNCTION stats.fn_seed_daily_activity(tstzrange) IS 'Idempotent backfill of daily_activity and daily_active_users from source tables; p_period limits to a date range; pass NULL to seed all history';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateDailyActivitySeed extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION stats.fn_seed_daily_activity(
  p_period tstzrange DEFAULT NULL
) RETURNS BIGINT LANGUAGE plpgsql AS $$
DECLARE
  v_rows BIGINT := 0;
BEGIN

  INSERT INTO stats.daily_activity (
    activity_date, total_moves, drops, grabs,
    comments, sees, archives, dips, km_contributed
  )
  SELECT
    moved_on_datetime::date, COUNT(*),
    COUNT(*) FILTER (WHERE move_type = 0), COUNT(*) FILTER (WHERE move_type = 1),
    COUNT(*) FILTER (WHERE move_type = 2), COUNT(*) FILTER (WHERE move_type = 3),
    COUNT(*) FILTER (WHERE move_type = 4), COUNT(*) FILTER (WHERE move_type = 5),
    COALESCE(SUM(km_distance), 0)::NUMERIC(15,3)
  FROM geokrety.gk_moves
  WHERE (p_period IS NULL OR moved_on_datetime <@ p_period)
  GROUP BY moved_on_datetime::date
  ON CONFLICT (activity_date) DO UPDATE SET
    total_moves    = EXCLUDED.total_moves,
    drops          = EXCLUDED.drops,
    grabs          = EXCLUDED.grabs,
    comments       = EXCLUDED.comments,
    sees           = EXCLUDED.sees,
    archives       = EXCLUDED.archives,
    dips           = EXCLUDED.dips,
    km_contributed = EXCLUDED.km_contributed;
  GET DIAGNOSTICS v_rows = ROW_COUNT;

  INSERT INTO stats.daily_activity (activity_date, gk_created)
  SELECT created_on_datetime::date, COUNT(*)
  FROM geokrety.gk_geokrety
  WHERE (p_period IS NULL OR created_on_datetime <@ p_period)
  GROUP BY created_on_datetime::date
  ON CONFLICT (activity_date) DO UPDATE SET
    gk_created = EXCLUDED.gk_created;

  INSERT INTO stats.daily_activity (
    activity_date, pictures_uploaded_total, pictures_uploaded_avatar,
    pictures_uploaded_move, pictures_uploaded_user
  )
  SELECT
    created_on_datetime::date, COUNT(*),
    COUNT(*) FILTER (WHERE type = 0), COUNT(*) FILTER (WHERE type = 1),
    COUNT(*) FILTER (WHERE type = 2)
  FROM geokrety.gk_pictures
  WHERE (p_period IS NULL OR created_on_datetime <@ p_period)
  GROUP BY created_on_datetime::date
  ON CONFLICT (activity_date) DO UPDATE SET
    pictures_uploaded_total  = EXCLUDED.pictures_uploaded_total,
    pictures_uploaded_avatar = EXCLUDED.pictures_uploaded_avatar,
    pictures_uploaded_move   = EXCLUDED.pictures_uploaded_move,
    pictures_uploaded_user   = EXCLUDED.pictures_uploaded_user;

  INSERT INTO stats.daily_activity (activity_date, users_registered)
  SELECT joined_on_datetime::date, COUNT(*)
  FROM geokrety.gk_users
  WHERE (p_period IS NULL OR joined_on_datetime <@ p_period)
  GROUP BY joined_on_datetime::date
  ON CONFLICT (activity_date) DO UPDATE SET
    users_registered = EXCLUDED.users_registered;

  INSERT INTO stats.daily_active_users (activity_date, user_id)
  SELECT DISTINCT moved_on_datetime::date, author
  FROM geokrety.gk_moves
  WHERE author IS NOT NULL
    AND (p_period IS NULL OR moved_on_datetime <@ p_period)
  ON CONFLICT (activity_date, user_id) DO NOTHING;

  RAISE NOTICE 'Daily activity seed completed: % rows in daily_activity affected', v_rows;
  RETURN v_rows;
END;
$$;

COMMENT ON FUNCTION stats.fn_seed_daily_activity(tstzrange) IS 'Idempotent backfill of daily_activity and daily_active_users from source tables';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP FUNCTION IF EXISTS stats.fn_seed_daily_activity(tstzrange);');
    }
}
```

#### SQL Usage Examples

```sql
-- Full historical backfill (run once during Sprint 6)
SELECT stats.fn_seed_daily_activity(NULL);

-- Seed only January 2026
SELECT stats.fn_seed_daily_activity('[2026-01-01, 2026-02-01)'::tstzrange);

-- Refresh last 7 days (incremental refresh pattern)
SELECT stats.fn_seed_daily_activity(
  tstzrange(
    (CURRENT_DATE - INTERVAL '7 days')::timestamptz,
    CURRENT_TIMESTAMP
  )
);

-- Verify seed outputs match source table
SELECT
  da.activity_date,
  da.total_moves AS seeded_total,
  src.move_count AS source_total
FROM stats.daily_activity da
JOIN (
  SELECT moved_on_datetime::date AS day, COUNT(*) AS move_count
  FROM geokrety.gk_moves GROUP BY 1
) src ON src.day = da.activity_date
WHERE da.total_moves != src.move_count
LIMIT 10;
-- Zero rows means seed is accurate
```

#### Graph/Visualization Specification

No new chart types unlocked at this step. This function populates all `stats.daily_activity` columns which feed the activity timeline charts defined in Steps 2.2 through 2.4.

#### TimescaleDB Assessment

**NOT applicable.** This step creates a utility function, not a table.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.12.001 | Function fn_seed_daily_activity exists | `has_function('stats', 'fn_seed_daily_activity', ARRAY['tstzrange'])` |
| T-2.12.002 | Function returns bigint | `function_returns('stats', 'fn_seed_daily_activity', ARRAY['tstzrange'], 'bigint')` |
| T-2.12.003 | Function with NULL executes without error | `lives_ok($$ SELECT stats.fn_seed_daily_activity(NULL); $$)` |
| T-2.12.004 | Function with date range executes without error | `lives_ok($$ SELECT stats.fn_seed_daily_activity('[2020-01-01,2020-01-31)'::tstzrange); $$)` |
| T-2.12.005 | total_moves matches gk_moves count per day | After seed, per-day counts in daily_activity match direct COUNT from gk_moves |
| T-2.12.006 | km_contributed matches sum of km_distance per day | Verify km_contributed matches aggregated source values |
| T-2.12.007 | daily_active_users populated | After seed, at least one row in daily_active_users per active day |
| T-2.12.008 | Function is idempotent | Call twice, verify counts identical after second call |
| T-2.12.009 | NULL author_id rows excluded from daily_active_users | Moves with NULL author never appear in daily_active_users |
| T-2.12.010 | Returns count of affected daily_activity rows | Verify return value > 0 when source tables are non-empty |

#### Implementation Checklist

- [ ] 1. Verify `stats.daily_activity` and `stats.daily_active_users` tables exist (Steps 2.2, 2.3)
- [ ] 2. Create migration file `20260310201100_create_daily_activity_seed.php`
- [ ] 3. Run `phinx migrate`
- [ ] 4. Verify function `stats.fn_seed_daily_activity` exists
- [ ] 5. Test function with NULL parameter — full seed
- [ ] 6. Test function with a date range parameter
- [ ] 7. Verify total_moves, km_contributed, users_registered match source aggregates
- [ ] 8. Verify daily_active_users populated; NULL authors excluded
- [ ] 9. Test idempotency: run twice, verify identical results
- [ ] 10. Run pgTAP tests T-2.12.001 through T-2.12.010

---

## 6. Consolidated Tail Note

The detailed acceptance, testing, rationale, dependency, and appendix material that followed in earlier drafts is now superseded by the canonical requirements and data contracts defined in this document's aligned Sprint 2 task sections and by the master spec.

Use the canonical identifiers and plain-table contracts only:

- `geokrety.fn_set_previous_move_id_and_distance()`
- `geokrety.gk_moves.km_distance`
- `stats.daily_activity.total_moves`, `drops`, `km_contributed`, `gk_created`
- `stats.daily_entity_counts.cnt`

Deprecated legacy identifiers and outdated examples are obsolete and non-normative.
