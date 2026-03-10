---
title: "Task S3T06: Create Country History Trigger Function + Attach"
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
task: S3T06
step: 3.6
migration: 20260310300500_create_country_history_trigger.php
blocks: [5, 6]
changelog:
  - 2026-03-10: created by merge of 03-sprint-3-country-geography.md step 3.6
  - 2026-03-10: marked missing canonical SQL and Phinx implementation as a blocking question
---

# Task S3T06: Create Country History Trigger Function + Attach

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 3 task set in `docs/database-refactor/sprint-3/`. `00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- `stats.country_daily_stats.unique_users` and `unique_gks` are exact online-maintained values, not approximate placeholders.
- `INSERT`, `UPDATE`, and `DELETE` handling for `stats.gk_countries_visited`, `stats.user_countries`, and `stats.gk_country_history` must maintain exact state. When earliest/latest rows are invalidated, affected rows must be recomputed from remaining qualifying moves.
- Snapshot functions seed and verify canonical state; they do not compensate for knowingly inexact live maintenance.
- Any lower text that still describes `unique_users` or `unique_gks` as approximate is obsolete and superseded by this alignment block.

## Source

- Generated from sprint document step `3.6` in `03-sprint-3-country-geography.md`.

### Step 3.6: Create Country History Trigger Function + Attach

**What this step does:** Creates the trigger function `geokrety.fn_gk_moves_country_history()` and attaches it as `tr_gk_moves_after_country_history` AFTER INSERT OR UPDATE OR DELETE on `geokrety.gk_moves`. The canonical contract is exact interval maintenance for `stats.gk_country_history`, including repair of neighboring intervals when a transition move is updated or deleted. COMMENT and ARCHIVE moves (types 2, 4) do not affect country history.

**Migration file name:** `20260310300500_create_country_history_trigger.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_country_history()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP IN ('UPDATE', 'DELETE') THEN
    DELETE FROM stats.gk_country_history
    WHERE geokrety_id = OLD.geokret;

    INSERT INTO stats.gk_country_history (
      geokrety_id,
      country_code,
      arrived_at,
      departed_at,
      move_id
    )
    WITH ordered_moves AS (
      SELECT
        m.id,
        m.geokret,
        m.country,
        m.moved_on_datetime,
        LAG(m.country) OVER (
          PARTITION BY m.geokret
          ORDER BY m.moved_on_datetime, m.id
        ) AS prev_country
      FROM geokrety.gk_moves m
      WHERE m.geokret = OLD.geokret
        AND m.country IS NOT NULL
        AND m.move_type IN (0, 1, 3, 5)
    ),
    transitions AS (
      SELECT
        id,
        geokret,
        country,
        moved_on_datetime,
        LEAD(moved_on_datetime) OVER (ORDER BY moved_on_datetime, id) AS next_arrived_at
      FROM ordered_moves
      WHERE prev_country IS DISTINCT FROM country
    )
    SELECT
      geokret,
      country,
      moved_on_datetime,
      next_arrived_at,
      id
    FROM transitions;
  END IF;

  IF TG_OP = 'INSERT' OR (TG_OP = 'UPDATE' AND NEW.geokret <> OLD.geokret) THEN
    DELETE FROM stats.gk_country_history
    WHERE geokrety_id = NEW.geokret;

    INSERT INTO stats.gk_country_history (
      geokrety_id,
      country_code,
      arrived_at,
      departed_at,
      move_id
    )
    WITH ordered_moves AS (
      SELECT
        m.id,
        m.geokret,
        m.country,
        m.moved_on_datetime,
        LAG(m.country) OVER (
          PARTITION BY m.geokret
          ORDER BY m.moved_on_datetime, m.id
        ) AS prev_country
      FROM geokrety.gk_moves m
      WHERE m.geokret = NEW.geokret
        AND m.country IS NOT NULL
        AND m.move_type IN (0, 1, 3, 5)
    ),
    transitions AS (
      SELECT
        id,
        geokret,
        country,
        moved_on_datetime,
        LEAD(moved_on_datetime) OVER (ORDER BY moved_on_datetime, id) AS next_arrived_at
      FROM ordered_moves
      WHERE prev_country IS DISTINCT FROM country
    )
    SELECT
      geokret,
      country,
      moved_on_datetime,
      next_arrived_at,
      id
    FROM transitions;
  END IF;

  RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$;

DROP TRIGGER IF EXISTS tr_gk_moves_after_country_history ON geokrety.gk_moves;
CREATE TRIGGER tr_gk_moves_after_country_history
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_country_history();
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateCountryHistoryTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_country_history()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP IN ('UPDATE', 'DELETE') THEN
    DELETE FROM stats.gk_country_history
    WHERE geokrety_id = OLD.geokret;

    INSERT INTO stats.gk_country_history (
      geokrety_id,
      country_code,
      arrived_at,
      departed_at,
      move_id
    )
    WITH ordered_moves AS (
      SELECT
        m.id,
        m.geokret,
        m.country,
        m.moved_on_datetime,
        LAG(m.country) OVER (
          PARTITION BY m.geokret
          ORDER BY m.moved_on_datetime, m.id
        ) AS prev_country
      FROM geokrety.gk_moves m
      WHERE m.geokret = OLD.geokret
        AND m.country IS NOT NULL
        AND m.move_type IN (0, 1, 3, 5)
    ),
    transitions AS (
      SELECT
        id,
        geokret,
        country,
        moved_on_datetime,
        LEAD(moved_on_datetime) OVER (ORDER BY moved_on_datetime, id) AS next_arrived_at
      FROM ordered_moves
      WHERE prev_country IS DISTINCT FROM country
    )
    SELECT
      geokret,
      country,
      moved_on_datetime,
      next_arrived_at,
      id
    FROM transitions;
  END IF;

  IF TG_OP = 'INSERT' OR (TG_OP = 'UPDATE' AND NEW.geokret <> OLD.geokret) THEN
    DELETE FROM stats.gk_country_history
    WHERE geokrety_id = NEW.geokret;

    INSERT INTO stats.gk_country_history (
      geokrety_id,
      country_code,
      arrived_at,
      departed_at,
      move_id
    )
    WITH ordered_moves AS (
      SELECT
        m.id,
        m.geokret,
        m.country,
        m.moved_on_datetime,
        LAG(m.country) OVER (
          PARTITION BY m.geokret
          ORDER BY m.moved_on_datetime, m.id
        ) AS prev_country
      FROM geokrety.gk_moves m
      WHERE m.geokret = NEW.geokret
        AND m.country IS NOT NULL
        AND m.move_type IN (0, 1, 3, 5)
    ),
    transitions AS (
      SELECT
        id,
        geokret,
        country,
        moved_on_datetime,
        LEAD(moved_on_datetime) OVER (ORDER BY moved_on_datetime, id) AS next_arrived_at
      FROM ordered_moves
      WHERE prev_country IS DISTINCT FROM country
    )
    SELECT
      geokret,
      country,
      moved_on_datetime,
      next_arrived_at,
      id
    FROM transitions;
  END IF;

  RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$;

DROP TRIGGER IF EXISTS tr_gk_moves_after_country_history ON geokrety.gk_moves;
CREATE TRIGGER tr_gk_moves_after_country_history
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_country_history();
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TRIGGER IF EXISTS tr_gk_moves_after_country_history ON geokrety.gk_moves;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_country_history() CASCADE;');
    }
}
```

#### SQL Usage Examples

```sql
-- Verify trigger exists
SELECT tgname FROM pg_trigger t
JOIN pg_class c ON c.oid = t.tgrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'geokrety' AND c.relname = 'gk_moves'
  AND t.tgname = 'tr_gk_moves_after_country_history';

-- Verify function exists
SELECT proname FROM pg_proc
WHERE proname = 'fn_gk_moves_country_history'
  AND pronamespace = 'geokrety'::regnamespace;

-- Test scenario: GK travels PL → DE → CZ
-- After first move in PL: one open interval (PL, arrived, NULL)
-- After move in DE: PL closed, DE opens
-- After move in CZ: DE closed, CZ opens
SELECT country_code, arrived_at, departed_at
FROM stats.gk_country_history
WHERE geokrety_id = 1
ORDER BY arrived_at;
```

#### Graph/Visualization Specification

No new visualization unlocked at this step. The trigger maintains data for the country timeline visualization defined in Step 3.4.

#### TimescaleDB Assessment

**NOT applicable.** This step creates a trigger function, not a table.

#### pgTAP Unit Tests

| Test ID   | Test Name                                                | Assertion                                                                                 |
| --------- | -------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| T-3.6.001 | Function fn_gk_moves_country_history exists              | `has_function('geokrety', 'fn_gk_moves_country_history', ARRAY[]::text[])`                |
| T-3.6.002 | Function returns trigger                                 | `function_returns('geokrety', 'fn_gk_moves_country_history', ARRAY[]::text[], 'trigger')` |
| T-3.6.003 | Trigger tr_gk_moves_after_country_history exists         | `has_trigger('geokrety', 'gk_moves', 'tr_gk_moves_after_country_history')`                |
| T-3.6.004 | First move in PL opens interval                          | Insert DROP in PL, verify `gk_country_history` has 1 row with `departed_at IS NULL`       |
| T-3.6.005 | Second move in same country (PL) is no-op                | Insert another DROP in PL, verify still 1 open interval                                   |
| T-3.6.006 | Move in DE closes PL and opens DE                        | Insert DROP in DE, verify PL has `departed_at` set and DE has open interval               |
| T-3.6.007 | COMMENT move does not affect history                     | Insert COMMENT (type 2) with country, verify no new interval                              |
| T-3.6.008 | ARCHIVE move does not affect history                     | Insert ARCHIVE (type 4) with country, verify no new interval                              |
| T-3.6.009 | Move with NULL country does not affect history           | Insert DROP with `country=NULL`, verify no new interval                                   |
| T-3.6.010 | DELETE repairs neighboring intervals exactly             | Delete a transition move and verify predecessor/successor intervals remain exact          |
| T-3.6.011 | UPDATE country from PL to DE repairs interval boundaries | Update a transition move and verify exact predecessor/successor boundaries                |
| T-3.6.012 | SEEN (type 3) with country opens interval                | Insert SEEN in FR, verify open interval for FR                                            |
| T-3.6.013 | DIP (type 5) with country opens interval                 | Insert DIP in AT, verify open interval for AT                                             |
| T-3.6.014 | GRAB (type 1) with country opens interval if new country | Insert GRAB in SK when open interval is PL, verify PL closed and SK opened                |
| T-3.6.015 | Multiple GKs have independent intervals                  | Two different GKs both in PL, verify two separate open intervals                          |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310300500_create_country_history_trigger.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify function `geokrety.fn_gk_moves_country_history` exists
- [ ] 4. Verify trigger `tr_gk_moves_after_country_history` exists on `gk_moves`
- [ ] 5. Test: first move in PL → open interval
- [ ] 6. Test: same-country move → no new interval
- [ ] 7. Test: different-country move → old closed, new opened
- [ ] 8. Test: COMMENT/ARCHIVE → no effect on intervals
- [ ] 9. Test: NULL country → no effect
- [ ] 10. Test: DELETE move → exact interval repair
- [ ] 11. Test: UPDATE country → exact interval repair
- [ ] 12. Test: exclusion constraint rejects overlapping manual inserts
- [ ] 13. Run pgTAP tests T-3.6.001 through T-3.6.015

## Agent Loop Log

- 2026-03-10T18:14:47Z - Loop 1 - `dba`: Preferred full per-GK history rebuild from qualifying transition moves because it guarantees exact interval repair after updates and deletes.
- 2026-03-10T18:14:47Z - Loop 1 - `critical-thinking`: No blocking contradiction; accepted a full-GK recompute path as canonical because correctness is more important than micro-optimization in rare edit/delete cases.
- 2026-03-10T18:14:47Z - Loop 1 - `specification`: Inserted full SQL and Phinx bodies, replaced the simplified-interval placeholder, and preserved the exact-history test contract.

## Resolution

- Blocking SQL/Phinx gap for this task has been resolved in this file; see Q-024 reference update in `99-OPEN-QUESTIONS.md`.

---
