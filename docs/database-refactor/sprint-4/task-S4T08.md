---
title: "Task S4T08: Waypoint Resolution & Cache Visit Trigger"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 4
tags:
  - cache-visits
  - database
  - database-refactor
  - dba
  - function
  - specification
  - sprint-4
  - sql
  - stats
  - task-index
  - task-merge
  - trigger
  - uc10
  - waypoints
depends_on:
  - S4T01
  - S4T02
  - S4T03
  - S4T04
  - S4T05
task: S4T08
step: 4.8
migration: 20260310400700_create_waypoint_cache_trigger.php
blocks:
  - S4T10
  - S4T11
changelog:
  - 2026-03-10: created by merge of task-S4T08.dba.md and task-S4T08.specification.md
  - 2026-03-10: logged the stale waypoint-column conflict in the merged trigger SQL
---

# Task S4T08: Waypoint Resolution & Cache Visit Trigger

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T08.dba.md`
- Specification source: `task-S4T08.specification.md`

## Resolved Decision

- The canonical Sprint 4 stats trigger remains `tr_gk_moves_after_waypoint_visits` calling `geokrety.fn_gk_moves_waypoint_cache()`.
- The existing website-side behavior referenced in Q-027 (`after_50_manage_waypoint_gc` calling `save_gc_waypoints()`) is treated as the upstream behavioral reference, not as the canonical stats object name.
- `stats.waypoints` uses only the canonical reduced column set: `waypoint_code`, `source`, `lat`, `lon`, `country`, `first_seen_at`.
- `INSERT`, `UPDATE`, and `DELETE` reconcile `stats.gk_cache_visits` and `stats.user_cache_visits` exactly from current `geokrety.gk_moves` state for the touched `(entity, waypoint)` keys.

## Purpose & Scope

Creates the trigger function `geokrety.fn_gk_moves_waypoint_cache()` and attaches it as trigger `tr_gk_moves_after_waypoint_visits` on `geokrety.gk_moves` (AFTER INSERT OR UPDATE OR DELETE). This trigger:

1. **Resolves** the waypoint code from `NEW.waypoint` → upserts into `stats.waypoints` (via S4T01 table)
2. **Upserts** `stats.gk_cache_visits` for the GeoKret visiting that waypoint
3. **Upserts** `stats.user_cache_visits` for the authenticated user visiting that waypoint

**Skip conditions:**

- `NEW.waypoint IS NULL` → skip waypoint resolution and cache visits
- `NEW.author IS NULL` → skip `user_cache_visits` row only (anonymous move)
- `move_type = 2` (COMMENT) → skip entirely (comments do not represent physical presence)

**On UPDATE / DELETE:** recompute the touched `(gk_id, waypoint_id)` and `(user_id, waypoint_id)` aggregates exactly from the current `geokrety.gk_moves` rows after the change is applied.

When a move is logged on `geokrety.gk_moves`, this trigger automatically:

1. Normalizes and upserts the waypoint code into `stats.waypoints`
2. Tracks per-GeoKret cache visits in `stats.gk_cache_visits`
3. Tracks per-user cache visits in `stats.user_cache_visits`

**Scope:** Trigger function + attachment. Requires S4T01 (waypoints), S4T04 (gk_cache_visits), S4T05 (user_cache_visits) to be deployed first.

---

## Requirements

| ID      | Description                                                                                         | MoSCoW |
| ------- | --------------------------------------------------------------------------------------------------- | ------ |
| REQ-470 | Function `geokrety.fn_gk_moves_waypoint_cache()` exists                                             | MUST   |
| REQ-471 | Trigger `tr_gk_moves_after_waypoint_visits` AFTER INSERT OR UPDATE OR DELETE on `geokrety.gk_moves` | MUST   |
| REQ-472 | Move type `2` (COMMENT) → skip all cache visit upserts entirely                                     | MUST   |
| REQ-473 | `NEW.waypoint IS NULL` → skip waypoint upsert and all cache visits                                  | MUST   |
| REQ-474 | Waypoint code stored as `UPPER(NEW.waypoint)` always                                                | MUST   |
| REQ-475 | `stats.waypoints` upsert uses canonical columns only: `waypoint_code`, `source`, `lat`, `lon`, `country`, `first_seen_at` | MUST   |
| REQ-476 | `stats.gk_cache_visits` is reconciled exactly from current qualifying `gk_moves` rows for the touched GK+waypoint key | MUST   |
| REQ-477 | `stats.user_cache_visits` is reconciled exactly from current qualifying `gk_moves` rows only when the move author is non-NULL | MUST   |
| REQ-478 | UPDATE and DELETE remove OLD contributions exactly; rows are deleted when no qualifying visits remain | MUST   |
| REQ-479 | `phinx rollback` drops trigger then function, no residual objects                                   | MUST   |

---

## Acceptance Criteria

| #   | Criterion                                                       | How to Verify                                          |
| --- | --------------------------------------------------------------- | ------------------------------------------------------ |
| 1   | Function created in `geokrety` schema                           | `\df geokrety.fn_gk_moves_waypoint_cache`              |
| 2   | Trigger attached to `geokrety.gk_moves`                         | `\d geokrety.gk_moves` → trigger list                  |
| 3   | COMMENT move triggers no cache row creation                     | Insert `move_type=2` with waypoint; check stats tables |
| 4   | NULL waypoint triggers no stats rows                            | Insert with `waypoint=NULL`; check stats tables        |
| 5   | Waypoint code stored uppercase                                  | Insert `'gc1ZXYA'`; SELECT from `stats.waypoints`      |
| 6   | Anonymous move (author=NULL) creates no `user_cache_visits` row | Insert with `author=NULL` + valid waypoint             |
| 7   | Two moves on same GK+waypoint → `visit_count=2`                 | Insert 2 moves; SELECT exact aggregate                 |
| 8   | UPDATE / DELETE reconciliation is exact                         | Insert, UPDATE/DELETE, then compare against source rows |
| 9   | Rollback removes trigger and function cleanly                   | `phinx rollback`; check `pg_trigger` and `pg_proc`     |

---

## Migration File

**`20260310400700_create_waypoint_cache_trigger.php`**

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateWaypointCacheTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_waypoint_cache()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
DECLARE
  v_waypoint_id  BIGINT;
  v_waypoint_code TEXT;
BEGIN
  IF TG_OP IN ('INSERT', 'UPDATE') AND NEW.move_type <> 2 AND NEW.waypoint IS NOT NULL THEN
    v_waypoint_code := UPPER(NEW.waypoint);

    INSERT INTO stats.waypoints (waypoint_code, source, lat, lon, country, first_seen_at)
    VALUES (
      v_waypoint_code,
      'UK',
      NEW.lat,
      NEW.lon,
      CASE
        WHEN NEW.country IS NULL OR BTRIM(NEW.country) = '' THEN NULL
        ELSE UPPER(NEW.country)
      END,
      NEW.moved_on_datetime
    )
    ON CONFLICT (waypoint_code) DO UPDATE SET
      lat = COALESCE(stats.waypoints.lat, EXCLUDED.lat),
      lon = COALESCE(stats.waypoints.lon, EXCLUDED.lon),
      country = COALESCE(stats.waypoints.country, EXCLUDED.country),
      first_seen_at = LEAST(stats.waypoints.first_seen_at, EXCLUDED.first_seen_at)
    RETURNING id INTO v_waypoint_id;

    IF v_waypoint_id IS NULL THEN
      SELECT id INTO v_waypoint_id
      FROM stats.waypoints
      WHERE waypoint_code = v_waypoint_code;
    END IF;

    INSERT INTO stats.gk_cache_visits (gk_id, waypoint_id, visit_count, first_visited_at, last_visited_at)
    SELECT
      NEW.geokret_id,
      v_waypoint_id,
      COUNT(*)::BIGINT,
      MIN(m.moved_on_datetime),
      MAX(m.moved_on_datetime)
    FROM geokrety.gk_moves m
    WHERE m.geokret_id = NEW.geokret_id
      AND m.waypoint IS NOT NULL
      AND UPPER(m.waypoint) = v_waypoint_code
      AND m.move_type <> 2
    GROUP BY NEW.geokret_id
    ON CONFLICT (gk_id, waypoint_id) DO UPDATE SET
      visit_count = EXCLUDED.visit_count,
      first_visited_at = EXCLUDED.first_visited_at,
      last_visited_at = EXCLUDED.last_visited_at;

    IF NEW.author IS NOT NULL THEN
      INSERT INTO stats.user_cache_visits (user_id, waypoint_id, visit_count, first_visited_at, last_visited_at)
      SELECT
        NEW.author,
        v_waypoint_id,
        COUNT(*)::BIGINT,
        MIN(m.moved_on_datetime),
        MAX(m.moved_on_datetime)
      FROM geokrety.gk_moves m
      WHERE m.author = NEW.author
        AND m.waypoint IS NOT NULL
        AND UPPER(m.waypoint) = v_waypoint_code
        AND m.move_type <> 2
      GROUP BY NEW.author
      ON CONFLICT (user_id, waypoint_id) DO UPDATE SET
        visit_count = EXCLUDED.visit_count,
        first_visited_at = EXCLUDED.first_visited_at,
        last_visited_at = EXCLUDED.last_visited_at;
    END IF;
  END IF;

  IF TG_OP IN ('DELETE', 'UPDATE') AND OLD.move_type <> 2 AND OLD.waypoint IS NOT NULL THEN
    v_waypoint_code := UPPER(OLD.waypoint);

    SELECT id INTO v_waypoint_id
    FROM stats.waypoints
    WHERE waypoint_code = v_waypoint_code;

    IF v_waypoint_id IS NOT NULL THEN
      INSERT INTO stats.gk_cache_visits (gk_id, waypoint_id, visit_count, first_visited_at, last_visited_at)
      SELECT
        OLD.geokret_id,
        v_waypoint_id,
        COUNT(*)::BIGINT,
        MIN(m.moved_on_datetime),
        MAX(m.moved_on_datetime)
      FROM geokrety.gk_moves m
      WHERE m.geokret_id = OLD.geokret_id
        AND m.waypoint IS NOT NULL
        AND UPPER(m.waypoint) = v_waypoint_code
        AND m.move_type <> 2
      GROUP BY OLD.geokret_id
      ON CONFLICT (gk_id, waypoint_id) DO UPDATE SET
        visit_count = EXCLUDED.visit_count,
        first_visited_at = EXCLUDED.first_visited_at,
        last_visited_at = EXCLUDED.last_visited_at;

      DELETE FROM stats.gk_cache_visits
      WHERE gk_id = OLD.geokret_id
        AND waypoint_id = v_waypoint_id
        AND NOT EXISTS (
          SELECT 1
          FROM geokrety.gk_moves m
          WHERE m.geokret_id = OLD.geokret_id
            AND m.waypoint IS NOT NULL
            AND UPPER(m.waypoint) = v_waypoint_code
            AND m.move_type <> 2
        );

      IF OLD.author IS NOT NULL THEN
        INSERT INTO stats.user_cache_visits (user_id, waypoint_id, visit_count, first_visited_at, last_visited_at)
        SELECT
          OLD.author,
          v_waypoint_id,
          COUNT(*)::BIGINT,
          MIN(m.moved_on_datetime),
          MAX(m.moved_on_datetime)
        FROM geokrety.gk_moves m
        WHERE m.author = OLD.author
          AND m.waypoint IS NOT NULL
          AND UPPER(m.waypoint) = v_waypoint_code
          AND m.move_type <> 2
        GROUP BY OLD.author
        ON CONFLICT (user_id, waypoint_id) DO UPDATE SET
          visit_count = EXCLUDED.visit_count,
          first_visited_at = EXCLUDED.first_visited_at,
          last_visited_at = EXCLUDED.last_visited_at;

        DELETE FROM stats.user_cache_visits
        WHERE user_id = OLD.author
          AND waypoint_id = v_waypoint_id
          AND NOT EXISTS (
            SELECT 1
            FROM geokrety.gk_moves m
            WHERE m.author = OLD.author
              AND m.waypoint IS NOT NULL
              AND UPPER(m.waypoint) = v_waypoint_code
              AND m.move_type <> 2
          );
      END IF;
    END IF;
  END IF;

  RETURN COALESCE(NEW, OLD);
END;
$$;

CREATE TRIGGER tr_gk_moves_after_waypoint_visits
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_waypoint_cache();
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP TRIGGER IF EXISTS tr_gk_moves_after_waypoint_visits ON geokrety.gk_moves;
DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_waypoint_cache();
SQL
        );
    }
}
```

## SQL Usage Examples

```sql
-- Verify trigger is attached
SELECT tgname, tgtype, tgevent
FROM pg_trigger
WHERE tgrelid = 'geokrety.gk_moves'::regclass
  AND tgname = 'tr_gk_moves_after_waypoint_visits';

-- Simulate: insert a DROP move with waypoint, verify stats.waypoints and gk_cache_visits
INSERT INTO geokrety.gk_moves (geokret_id, move_type, waypoint, author, lat, lon, moved_on_datetime)
VALUES (1, 0, 'GC1ZXYA', 42, 52.123, 13.456, now());

SELECT * FROM stats.waypoints WHERE waypoint_code = 'GC1ZXYA';
SELECT * FROM stats.gk_cache_visits WHERE gk_id = 1;
SELECT * FROM stats.user_cache_visits WHERE user_id = 42;

-- Verify anonymous move does NOT create user_cache_visits row
INSERT INTO geokrety.gk_moves (geokret_id, move_type, waypoint, author, lat, lon, moved_on_datetime)
VALUES (2, 0, 'OC00123', NULL, 50.0, 14.0, now());
SELECT * FROM stats.user_cache_visits WHERE user_id IS NULL; -- should be 0 rows
```

## pgTAP Unit Tests

| Test ID   | Assertion                                                                          | Expected |
| --------- | ---------------------------------------------------------------------------------- | -------- |
| T-4.8.001 | Function `geokrety.fn_gk_moves_waypoint_cache()` exists                            | pass     |
| T-4.8.002 | Trigger `tr_gk_moves_after_waypoint_visits` exists on `geokrety.gk_moves`          | pass     |
| T-4.8.003 | INSERT DROP move with waypoint → `stats.waypoints` row created (`UPPER(waypoint)`) | pass     |
| T-4.8.004 | INSERT DROP move with waypoint → `stats.gk_cache_visits` row created               | pass     |
| T-4.8.005 | INSERT DROP move with waypoint + author → `stats.user_cache_visits` row created    | pass     |
| T-4.8.006 | INSERT DROP move with waypoint, `author = NULL` → no `user_cache_visits` row       | pass     |
| T-4.8.007 | INSERT COMMENT move (type 2) with waypoint → no `gk_cache_visits` row              | pass     |
| T-4.8.008 | INSERT DROP move with `waypoint = NULL` → no `stats.waypoints` / cache visit rows  | pass     |
| T-4.8.009 | Two DROP moves same GK + cache → exact `visit_count = 2` in `gk_cache_visits`      | pass     |
| T-4.8.010 | UPDATE / DELETE move → cache visit rows are recomputed exactly or removed           | pass     |
| T-4.8.011 | Waypoint code stored as UPPER (e.g. `gc1zxya` → `GC1ZXYA`)                         | pass     |

| Test ID   | Scenario                                        | Pass Condition                                                         |
| --------- | ----------------------------------------------- | ---------------------------------------------------------------------- |
| T-4.8.001 | Function `fn_gk_moves_waypoint_cache` exists    | `has_function('geokrety','fn_gk_moves_waypoint_cache')`                |
| T-4.8.002 | Trigger on `geokrety.gk_moves`                  | `has_trigger('geokrety.gk_moves','tr_gk_moves_after_waypoint_visits')` |
| T-4.8.003 | DROP move + waypoint → `stats.waypoints` row    | 1 row with UPPER code                                                  |
| T-4.8.004 | DROP move + waypoint → `gk_cache_visits` row    | 1 row with count=1                                                     |
| T-4.8.005 | DROP move + author → `user_cache_visits` row    | 1 row with count=1                                                     |
| T-4.8.006 | DROP move, author=NULL → no `user_cache_visits` | 0 rows                                                                 |
| T-4.8.007 | COMMENT move (type 2) → no cache rows           | 0 rows in both cache tables                                            |
| T-4.8.008 | NULL waypoint move → no rows in any stats table | 0 rows in all three stats tables                                       |
| T-4.8.009 | 2 DROP moves same GK+cache → `visit_count = 2`  | Single row, count=2                                                    |
| T-4.8.010 | UPDATE / DELETE reconciliation exact            | Aggregate matches qualifying `gk_moves` source rows                    |
| T-4.8.011 | Uppercase normalisation: `gc1zxya` → `GC1ZXYA`  | `stats.waypoints.waypoint_code = 'GC1ZXYA'`                            |

---

## Implementation Checklist

- [ ] 1. Create migration file `20260310400700_create_waypoint_cache_trigger.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. `\df geokrety.fn_gk_moves_waypoint_cache` — function exists
- [ ] 4. Trigger `tr_gk_moves_after_waypoint_visits` listed in `\d geokrety.gk_moves`
- [ ] 5. Test COMMENT move (type 2) → no cache rows
- [ ] 6. Test NULL waypoint → no cache rows
- [ ] 7. Test anonymous move → no `user_cache_visits` row
- [ ] 8. Test uppercase normalisation
- [ ] 9. Test UPDATE / DELETE reconciliation: row removed when count reaches 0 and timestamps match source rows
- [ ] 10. Run pgTAP T-4.8.001 through T-4.8.011
- [ ] 11. `phinx rollback` — trigger and function dropped

- [ ] 1. Write `20260310400700_create_waypoint_cache_trigger.php` with `up()` + `down()`
- [ ] 2. `phinx migrate` — confirm no errors
- [ ] 3. Trigger exists in `\d geokrety.gk_moves`
- [ ] 4. Function listed in `\df geokrety`
- [ ] 5. Run acceptance criteria 1–9 manually with test inserts
- [ ] 6. Run pgTAP T-4.8.001 through T-4.8.011 — all pass
- [ ] 7. `phinx rollback` — trigger + function absent

## Full SQL DDL — Trigger Function

The canonical function body is the same one embedded in the migration block above under `20260310400700_create_waypoint_cache_trigger.php`. This duplicate lower DDL section was intentionally collapsed to avoid preserving stale merged SQL that referenced obsolete waypoint columns.

## Full SQL DDL — Trigger Attachment

```sql
CREATE TRIGGER tr_gk_moves_after_waypoint_visits
  AFTER INSERT OR UPDATE OR DELETE
  ON geokrety.gk_moves
  FOR EACH ROW
  EXECUTE FUNCTION geokrety.fn_gk_moves_waypoint_cache();

COMMENT ON TRIGGER tr_gk_moves_after_waypoint_visits ON geokrety.gk_moves
  IS 'Fires on each row insert/update/delete to maintain stats.waypoints, gk_cache_visits, user_cache_visits';
```

## Trigger Logic Decision Table

| Condition                   | waypoints upsert | gk_cache_visits | user_cache_visits |
| --------------------------- | ---------------- | --------------- | ----------------- |
| `move_type = 2` (COMMENT)   | SKIP             | SKIP            | SKIP              |
| `NEW.waypoint IS NULL`      | SKIP             | SKIP            | SKIP              |
| `NEW.author IS NULL` (anon) | Upsert           | Upsert          | **SKIP**          |
| Normal authenticated move   | Upsert           | Upsert          | Upsert            |

## Master-Spec Alignment

This task is governed by [../00-SPEC-DRAFT-v1.obsolete.md](../00-SPEC-DRAFT-v1.obsolete.md), Sections 5.5 and 8.4.

- `stats.waypoints` uses `waypoint_code`, not `code`, and has no `last_seen_at` column in the canonical contract.
- `tr_gk_moves_after_waypoint_visits` is an `AFTER INSERT OR UPDATE OR DELETE` trigger, and `UPDATE` handling must reconcile `OLD` then `NEW` exactly.
- `visit_count` in `stats.gk_cache_visits` and `stats.user_cache_visits` counts qualifying move rows for the `(entity, waypoint)` pair.
- The website trigger `after_50_manage_waypoint_gc` / helper `save_gc_waypoints()` is the behavioral reference captured in Q-027; this task standardizes the stats-owned trigger/function names and reduced column contract.
- Any lower examples that use non-canonical waypoint column names are obsolete and superseded by this alignment block.

## Objects Created

| Object Type | Name                                    | Owning Schema          |
| ----------- | --------------------------------------- | ---------------------- |
| Function    | `geokrety.fn_gk_moves_waypoint_cache()` | `geokrety`             |
| Trigger     | `tr_gk_moves_after_waypoint_visits`     | on `geokrety.gk_moves` |

## Side-effects (rows modified in `stats.*`)

| Table                     | Operation                 | Condition                                 |
| ------------------------- | ------------------------- | ----------------------------------------- |
| `stats.waypoints`         | INSERT ON CONFLICT UPDATE | `waypoint IS NOT NULL AND move_type != 2` |
| `stats.gk_cache_visits`   | INSERT ON CONFLICT UPDATE | same as above                             |
| `stats.user_cache_visits` | INSERT ON CONFLICT UPDATE | same + `author IS NOT NULL`               |

## Agent Loop Log

- 2026-03-10T18:40:00Z — `dba`: replaced obsolete `code` / `name` / `last_seen_at` usage with canonical `stats.waypoints` columns and exact source-row reconciliation.
- 2026-03-10T18:40:00Z — `critical-thinking`: kept the Sprint 4 stats trigger names canonical while documenting the existing website trigger naming as upstream behavioral context.
- 2026-03-10T18:40:00Z — `specification`: updated requirements, SQL, tests, and checklist so UPDATE / DELETE semantics are exact instead of incremental draft behavior.

## Resolution

Q-027 is resolved by canonizing the reduced-column waypoint/cache trigger contract in this task.

---
