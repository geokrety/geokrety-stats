---
title: "Task S3T05: Create Country Rollups Trigger Function + Attach"
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
task: S3T05
step: 3.5
migration: 20260310300400_create_country_rollups_trigger.php
blocks: [5, 6]
changelog:
  - 2026-03-10: created by merge of 03-sprint-3-country-geography.md step 3.5
  - 2026-03-10: aligned test and checklist wording to exact recomputation semantics
  - 2026-03-10: marked missing canonical SQL and Phinx implementation as a blocking question
---

# Task S3T05: Create Country Rollups Trigger Function + Attach

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 3 task set in `docs/database-refactor/sprint-3/`. `../00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- `stats.country_daily_stats.unique_users` and `unique_gks` are exact online-maintained values, not approximate placeholders.
- `INSERT`, `UPDATE`, and `DELETE` handling for `stats.gk_countries_visited`, `stats.user_countries`, and `stats.gk_country_history` must maintain exact state. When earliest/latest rows are invalidated, affected rows must be recomputed from remaining qualifying moves.
- Snapshot functions seed and verify canonical state; they do not compensate for knowingly inexact live maintenance.
- Any lower text that still describes `unique_users` or `unique_gks` as approximate is obsolete and superseded by this alignment block.

## Source

- Generated from sprint document step `3.5` in `03-sprint-3-country-geography.md`.

### Step 3.5: Create Country Rollups Trigger Function + Attach

**What this step does:** Creates the trigger function `geokrety.fn_gk_moves_country_rollups()` and attaches it as `tr_gk_moves_after_country_rollups` AFTER INSERT OR UPDATE OR DELETE on `geokrety.gk_moves`. The canonical contract is exact-state maintenance for the affected country/day, GK-country, and user-country keys, including `unique_users`, `unique_gks`, `first_move_id`, `first_visited_at`, `first_visit`, and `last_visit`. NULL country values are skipped.

**Migration file name:** `20260310300400_create_country_rollups_trigger.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_country_rollups()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP IN ('UPDATE', 'DELETE') AND OLD.country IS NOT NULL THEN
    IF EXISTS (
      SELECT 1
      FROM geokrety.gk_moves m
      WHERE m.country = OLD.country
        AND m.moved_on_datetime::date = OLD.moved_on_datetime::date
    ) THEN
      INSERT INTO stats.country_daily_stats (
        stats_date,
        country_code,
        moves_count,
        drops,
        grabs,
        comments,
        sees,
        archives,
        dips,
        unique_users,
        unique_gks,
        km_contributed
      )
      SELECT
        OLD.moved_on_datetime::date,
        OLD.country,
        COUNT(*),
        COUNT(*) FILTER (WHERE move_type = 0),
        COUNT(*) FILTER (WHERE move_type = 1),
        COUNT(*) FILTER (WHERE move_type = 2),
        COUNT(*) FILTER (WHERE move_type = 3),
        COUNT(*) FILTER (WHERE move_type = 4),
        COUNT(*) FILTER (WHERE move_type = 5),
        COUNT(DISTINCT author) FILTER (WHERE author IS NOT NULL),
        COUNT(DISTINCT geokret),
        COALESCE(SUM(km_distance), 0)
      FROM geokrety.gk_moves m
      WHERE m.country = OLD.country
        AND m.moved_on_datetime::date = OLD.moved_on_datetime::date
      GROUP BY 1, 2
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
    ELSE
      DELETE FROM stats.country_daily_stats
      WHERE stats_date = OLD.moved_on_datetime::date
        AND country_code = OLD.country;
    END IF;

    IF EXISTS (
      SELECT 1
      FROM geokrety.gk_moves m
      WHERE m.geokret = OLD.geokret
        AND m.country = OLD.country
    ) THEN
      INSERT INTO stats.gk_countries_visited (
        geokrety_id,
        country_code,
        first_visited_at,
        first_move_id,
        move_count
      )
      SELECT
        OLD.geokret,
        OLD.country,
        first_row.moved_on_datetime,
        first_row.id,
        COUNT(*)
      FROM geokrety.gk_moves m
      CROSS JOIN LATERAL (
        SELECT m2.id, m2.moved_on_datetime
        FROM geokrety.gk_moves m2
        WHERE m2.geokret = OLD.geokret
          AND m2.country = OLD.country
        ORDER BY m2.moved_on_datetime ASC, m2.id ASC
        LIMIT 1
      ) AS first_row
      WHERE m.geokret = OLD.geokret
        AND m.country = OLD.country
      GROUP BY 1, 2, 3, 4
      ON CONFLICT (geokrety_id, country_code) DO UPDATE SET
        first_visited_at = EXCLUDED.first_visited_at,
        first_move_id = EXCLUDED.first_move_id,
        move_count = EXCLUDED.move_count;
    ELSE
      DELETE FROM stats.gk_countries_visited
      WHERE geokrety_id = OLD.geokret
        AND country_code = OLD.country;
    END IF;

    IF OLD.author IS NOT NULL THEN
      IF EXISTS (
        SELECT 1
        FROM geokrety.gk_moves m
        WHERE m.author = OLD.author
          AND m.country = OLD.country
      ) THEN
        INSERT INTO stats.user_countries (
          user_id,
          country_code,
          move_count,
          first_visit,
          last_visit
        )
        SELECT
          OLD.author,
          OLD.country,
          COUNT(*),
          MIN(moved_on_datetime),
          MAX(moved_on_datetime)
        FROM geokrety.gk_moves m
        WHERE m.author = OLD.author
          AND m.country = OLD.country
        GROUP BY 1, 2
        ON CONFLICT (user_id, country_code) DO UPDATE SET
          move_count = EXCLUDED.move_count,
          first_visit = EXCLUDED.first_visit,
          last_visit = EXCLUDED.last_visit;
      ELSE
        DELETE FROM stats.user_countries
        WHERE user_id = OLD.author
          AND country_code = OLD.country;
      END IF;
    END IF;
  END IF;

  IF TG_OP IN ('INSERT', 'UPDATE') AND NEW.country IS NOT NULL THEN
    INSERT INTO stats.country_daily_stats (
      stats_date,
      country_code,
      moves_count,
      drops,
      grabs,
      comments,
      sees,
      archives,
      dips,
      unique_users,
      unique_gks,
      km_contributed
    )
    SELECT
      NEW.moved_on_datetime::date,
      NEW.country,
      COUNT(*),
      COUNT(*) FILTER (WHERE move_type = 0),
      COUNT(*) FILTER (WHERE move_type = 1),
      COUNT(*) FILTER (WHERE move_type = 2),
      COUNT(*) FILTER (WHERE move_type = 3),
      COUNT(*) FILTER (WHERE move_type = 4),
      COUNT(*) FILTER (WHERE move_type = 5),
      COUNT(DISTINCT author) FILTER (WHERE author IS NOT NULL),
      COUNT(DISTINCT geokret),
      COALESCE(SUM(km_distance), 0)
    FROM geokrety.gk_moves m
    WHERE m.country = NEW.country
      AND m.moved_on_datetime::date = NEW.moved_on_datetime::date
    GROUP BY 1, 2
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

    INSERT INTO stats.gk_countries_visited (
      geokrety_id,
      country_code,
      first_visited_at,
      first_move_id,
      move_count
    )
    SELECT
      NEW.geokret,
      NEW.country,
      first_row.moved_on_datetime,
      first_row.id,
      COUNT(*)
    FROM geokrety.gk_moves m
    CROSS JOIN LATERAL (
      SELECT m2.id, m2.moved_on_datetime
      FROM geokrety.gk_moves m2
      WHERE m2.geokret = NEW.geokret
        AND m2.country = NEW.country
      ORDER BY m2.moved_on_datetime ASC, m2.id ASC
      LIMIT 1
    ) AS first_row
    WHERE m.geokret = NEW.geokret
      AND m.country = NEW.country
    GROUP BY 1, 2, 3, 4
    ON CONFLICT (geokrety_id, country_code) DO UPDATE SET
      first_visited_at = EXCLUDED.first_visited_at,
      first_move_id = EXCLUDED.first_move_id,
      move_count = EXCLUDED.move_count;

    IF NEW.author IS NOT NULL THEN
      INSERT INTO stats.user_countries (
        user_id,
        country_code,
        move_count,
        first_visit,
        last_visit
      )
      SELECT
        NEW.author,
        NEW.country,
        COUNT(*),
        MIN(moved_on_datetime),
        MAX(moved_on_datetime)
      FROM geokrety.gk_moves m
      WHERE m.author = NEW.author
        AND m.country = NEW.country
      GROUP BY 1, 2
      ON CONFLICT (user_id, country_code) DO UPDATE SET
        move_count = EXCLUDED.move_count,
        first_visit = EXCLUDED.first_visit,
        last_visit = EXCLUDED.last_visit;
    END IF;
  END IF;

  RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$;

DROP TRIGGER IF EXISTS tr_gk_moves_after_country_rollups ON geokrety.gk_moves;
CREATE TRIGGER tr_gk_moves_after_country_rollups
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_country_rollups();
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateCountryRollupsTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_country_rollups()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP IN ('UPDATE', 'DELETE') AND OLD.country IS NOT NULL THEN
    IF EXISTS (
      SELECT 1
      FROM geokrety.gk_moves m
      WHERE m.country = OLD.country
        AND m.moved_on_datetime::date = OLD.moved_on_datetime::date
    ) THEN
      INSERT INTO stats.country_daily_stats (
        stats_date,
        country_code,
        moves_count,
        drops,
        grabs,
        comments,
        sees,
        archives,
        dips,
        unique_users,
        unique_gks,
        km_contributed
      )
      SELECT
        OLD.moved_on_datetime::date,
        OLD.country,
        COUNT(*),
        COUNT(*) FILTER (WHERE move_type = 0),
        COUNT(*) FILTER (WHERE move_type = 1),
        COUNT(*) FILTER (WHERE move_type = 2),
        COUNT(*) FILTER (WHERE move_type = 3),
        COUNT(*) FILTER (WHERE move_type = 4),
        COUNT(*) FILTER (WHERE move_type = 5),
        COUNT(DISTINCT author) FILTER (WHERE author IS NOT NULL),
        COUNT(DISTINCT geokret),
        COALESCE(SUM(km_distance), 0)
      FROM geokrety.gk_moves m
      WHERE m.country = OLD.country
        AND m.moved_on_datetime::date = OLD.moved_on_datetime::date
      GROUP BY 1, 2
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
    ELSE
      DELETE FROM stats.country_daily_stats
      WHERE stats_date = OLD.moved_on_datetime::date
        AND country_code = OLD.country;
    END IF;

    IF EXISTS (
      SELECT 1
      FROM geokrety.gk_moves m
      WHERE m.geokret = OLD.geokret
        AND m.country = OLD.country
    ) THEN
      INSERT INTO stats.gk_countries_visited (
        geokrety_id,
        country_code,
        first_visited_at,
        first_move_id,
        move_count
      )
      SELECT
        OLD.geokret,
        OLD.country,
        first_row.moved_on_datetime,
        first_row.id,
        COUNT(*)
      FROM geokrety.gk_moves m
      CROSS JOIN LATERAL (
        SELECT m2.id, m2.moved_on_datetime
        FROM geokrety.gk_moves m2
        WHERE m2.geokret = OLD.geokret
          AND m2.country = OLD.country
        ORDER BY m2.moved_on_datetime ASC, m2.id ASC
        LIMIT 1
      ) AS first_row
      WHERE m.geokret = OLD.geokret
        AND m.country = OLD.country
      GROUP BY 1, 2, 3, 4
      ON CONFLICT (geokrety_id, country_code) DO UPDATE SET
        first_visited_at = EXCLUDED.first_visited_at,
        first_move_id = EXCLUDED.first_move_id,
        move_count = EXCLUDED.move_count;
    ELSE
      DELETE FROM stats.gk_countries_visited
      WHERE geokrety_id = OLD.geokret
        AND country_code = OLD.country;
    END IF;

    IF OLD.author IS NOT NULL THEN
      IF EXISTS (
        SELECT 1
        FROM geokrety.gk_moves m
        WHERE m.author = OLD.author
          AND m.country = OLD.country
      ) THEN
        INSERT INTO stats.user_countries (
          user_id,
          country_code,
          move_count,
          first_visit,
          last_visit
        )
        SELECT
          OLD.author,
          OLD.country,
          COUNT(*),
          MIN(moved_on_datetime),
          MAX(moved_on_datetime)
        FROM geokrety.gk_moves m
        WHERE m.author = OLD.author
          AND m.country = OLD.country
        GROUP BY 1, 2
        ON CONFLICT (user_id, country_code) DO UPDATE SET
          move_count = EXCLUDED.move_count,
          first_visit = EXCLUDED.first_visit,
          last_visit = EXCLUDED.last_visit;
      ELSE
        DELETE FROM stats.user_countries
        WHERE user_id = OLD.author
          AND country_code = OLD.country;
      END IF;
    END IF;
  END IF;

  IF TG_OP IN ('INSERT', 'UPDATE') AND NEW.country IS NOT NULL THEN
    INSERT INTO stats.country_daily_stats (
      stats_date,
      country_code,
      moves_count,
      drops,
      grabs,
      comments,
      sees,
      archives,
      dips,
      unique_users,
      unique_gks,
      km_contributed
    )
    SELECT
      NEW.moved_on_datetime::date,
      NEW.country,
      COUNT(*),
      COUNT(*) FILTER (WHERE move_type = 0),
      COUNT(*) FILTER (WHERE move_type = 1),
      COUNT(*) FILTER (WHERE move_type = 2),
      COUNT(*) FILTER (WHERE move_type = 3),
      COUNT(*) FILTER (WHERE move_type = 4),
      COUNT(*) FILTER (WHERE move_type = 5),
      COUNT(DISTINCT author) FILTER (WHERE author IS NOT NULL),
      COUNT(DISTINCT geokret),
      COALESCE(SUM(km_distance), 0)
    FROM geokrety.gk_moves m
    WHERE m.country = NEW.country
      AND m.moved_on_datetime::date = NEW.moved_on_datetime::date
    GROUP BY 1, 2
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

    INSERT INTO stats.gk_countries_visited (
      geokrety_id,
      country_code,
      first_visited_at,
      first_move_id,
      move_count
    )
    SELECT
      NEW.geokret,
      NEW.country,
      first_row.moved_on_datetime,
      first_row.id,
      COUNT(*)
    FROM geokrety.gk_moves m
    CROSS JOIN LATERAL (
      SELECT m2.id, m2.moved_on_datetime
      FROM geokrety.gk_moves m2
      WHERE m2.geokret = NEW.geokret
        AND m2.country = NEW.country
      ORDER BY m2.moved_on_datetime ASC, m2.id ASC
      LIMIT 1
    ) AS first_row
    WHERE m.geokret = NEW.geokret
      AND m.country = NEW.country
    GROUP BY 1, 2, 3, 4
    ON CONFLICT (geokrety_id, country_code) DO UPDATE SET
      first_visited_at = EXCLUDED.first_visited_at,
      first_move_id = EXCLUDED.first_move_id,
      move_count = EXCLUDED.move_count;

    IF NEW.author IS NOT NULL THEN
      INSERT INTO stats.user_countries (
        user_id,
        country_code,
        move_count,
        first_visit,
        last_visit
      )
      SELECT
        NEW.author,
        NEW.country,
        COUNT(*),
        MIN(moved_on_datetime),
        MAX(moved_on_datetime)
      FROM geokrety.gk_moves m
      WHERE m.author = NEW.author
        AND m.country = NEW.country
      GROUP BY 1, 2
      ON CONFLICT (user_id, country_code) DO UPDATE SET
        move_count = EXCLUDED.move_count,
        first_visit = EXCLUDED.first_visit,
        last_visit = EXCLUDED.last_visit;
    END IF;
  END IF;

  RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$;

DROP TRIGGER IF EXISTS tr_gk_moves_after_country_rollups ON geokrety.gk_moves;
CREATE TRIGGER tr_gk_moves_after_country_rollups
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_country_rollups();
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TRIGGER IF EXISTS tr_gk_moves_after_country_rollups ON geokrety.gk_moves;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_country_rollups() CASCADE;');
    }
}
```

#### SQL Usage Examples

```sql
-- Verify trigger exists on gk_moves
SELECT tgname, tgtype, tgenabled
FROM pg_trigger t
JOIN pg_class c ON c.oid = t.tgrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'geokrety' AND c.relname = 'gk_moves'
  AND t.tgname = 'tr_gk_moves_after_country_rollups';

-- Verify function exists
SELECT proname, pronamespace::regnamespace
FROM pg_proc
WHERE proname = 'fn_gk_moves_country_rollups';

-- Test: insert a move and verify country_daily_stats updated
-- (In test harness with test data)
-- INSERT INTO geokrety.gk_moves (..., country, move_type, ...) VALUES (..., 'PL', 0, ...);
-- SELECT moves_count, drops FROM stats.country_daily_stats WHERE country_code = 'PL' AND stats_date = CURRENT_DATE;
```

#### Graph/Visualization Specification

No new visualization unlocked at this step. The trigger maintains data for the visualizations defined in Steps 3.1–3.3.

#### TimescaleDB Assessment

**NOT applicable.** This step creates a trigger function, not a table.

#### pgTAP Unit Tests

| Test ID   | Test Name                                                | Assertion                                                                                                   |
| --------- | -------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| T-3.5.001 | Function fn_gk_moves_country_rollups exists              | `has_function('geokrety', 'fn_gk_moves_country_rollups', ARRAY[]::text[])`                                  |
| T-3.5.002 | Function returns trigger                                 | `function_returns('geokrety', 'fn_gk_moves_country_rollups', ARRAY[]::text[], 'trigger')`                   |
| T-3.5.003 | Trigger tr_gk_moves_after_country_rollups exists         | `has_trigger('geokrety', 'gk_moves', 'tr_gk_moves_after_country_rollups')`                                  |
| T-3.5.004 | INSERT with country updates country_daily_stats          | Insert move with `country='PL', move_type=0`, verify `country_daily_stats` row has `drops=1, moves_count=1` |
| T-3.5.005 | INSERT with NULL country creates no stats row            | Insert move with `country=NULL`, verify no new row in `country_daily_stats`                                 |
| T-3.5.006 | INSERT updates gk_countries_visited                      | Insert move with `country='PL'`, verify `gk_countries_visited` row exists with `move_count=1`               |
| T-3.5.007 | Second INSERT same GK same country increments move_count | Insert two moves for same GK in 'PL', verify `gk_countries_visited.move_count=2`                            |
| T-3.5.008 | INSERT with author updates user_countries                | Insert move with `author=42, country='PL'`, verify `user_countries` row exists                              |
| T-3.5.009 | INSERT with NULL author skips user_countries             | Insert anonymous move with `country='PL'`, verify no `user_countries` row for NULL user                     |
| T-3.5.010 | DELETE recomputes affected country/day aggregates        | Insert then delete move, verify `country_daily_stats`, `gk_countries_visited`, and `user_countries` are recomputed exactly |
| T-3.5.011 | UPDATE country from PL to DE recomputes both key sets    | Insert with `country='PL'`, update to `country='DE'`, verify both the old and new country keys are recomputed exactly |
| T-3.5.012 | UPDATE affecting earliest/latest visit timestamps repairs exact state | Update a move that was the first or latest visit, then verify `first_move_id`, `first_visited_at`, `first_visit`, and `last_visit` are repaired from remaining moves |
| T-3.5.013 | km_contributed recomputes correctly for touched country/day | Insert or update move with `km_distance=150.500`, verify the touched `country_daily_stats.km_contributed` value matches the exact recomputed sum |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310300400_create_country_rollups_trigger.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify function `geokrety.fn_gk_moves_country_rollups` exists
- [ ] 4. Verify trigger `tr_gk_moves_after_country_rollups` exists on `gk_moves`
- [ ] 5. Test INSERT with country → stats tables updated
- [ ] 6. Test INSERT with NULL country → no stats rows created
- [ ] 7. Test DELETE → exact recomputation for all touched country/day, GK-country, and user-country keys
- [ ] 8. Test UPDATE of country → exact recomputation for both old and new key sets
- [ ] 9. Test earliest/latest invalidation → first/last metadata repaired from remaining qualifying moves
- [ ] 10. Test anonymous move → no user_countries row
- [ ] 11. Run pgTAP tests T-3.5.001 through T-3.5.013

## Agent Loop Log

- 2026-03-10T18:12:26Z - Loop 1 - `dba`: Preferred exact recomputation of touched old/new keys over incremental math so earliest/latest metadata repairs remain correct after deletes and country edits.
- 2026-03-10T18:12:26Z - Loop 1 - `critical-thinking`: No blocking contradiction; accepted a less-optimized but exact recompute path as the canonical replacement for the removed incremental draft.
- 2026-03-10T18:12:26Z - Loop 1 - `specification`: Inserted full SQL and Phinx bodies, preserved test matrix semantics, and removed the blocking open-question placeholder.

## Resolution

- Blocking SQL/Phinx gap for this task has been resolved in this file; see Q-023 reference update in `99-OPEN-QUESTIONS.md`.

---
