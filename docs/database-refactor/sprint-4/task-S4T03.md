---
title: "Task S4T03: Seed stats.waypoints from GC/OC Sources"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 4
tags:
  - database
  - database-refactor
  - dba
  - seed
  - specification
  - sprint-4
  - sql
  - stats
  - task-index
  - task-merge
  - waypoints
depends_on:
  - S4T01
  - S4T02
task: S4T03
step: 4.3
migration: 20260310400200_seed_waypoints.php
blocks:
  - S4T04
  - S4T05
  - S4T08
changelog:
  - 2026.03.10: created by merge of task-S4T03.dba.md and task-S4T03.specification.md
  - 2026.03.10: logged the seed-source conflict between direct-table and canonical-view implementations
---

# Task S4T03: Seed stats.waypoints from GC/OC Sources

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T03.dba.md`
- Specification source: `task-S4T03.specification.md`

## Resolved Decision

- Waypoint seeding uses `stats.v_waypoints_source_union` exclusively.
- Direct reads from `geokrety.gk_waypoints_gc` and `geokrety.gk_waypoints_oc` are no longer canonical in this task.
- `GC` precedence is preserved by ordering seed rows before `OC` rows and relying on `ON CONFLICT DO NOTHING`.

## Purpose & Scope

Populates `stats.waypoints` from the canonical union view `stats.v_waypoints_source_union` using an idempotent `INSERT ... ON CONFLICT DO NOTHING` pattern. `GC` rows take precedence when the same code appears in both sources by being ordered ahead of `OC` rows in the seed query. After seeding, the migration also creates a helper function `stats.fn_seed_waypoints()` that can be re-run for incremental refreshes without a full migration rollback.

**Scope:** Seed INSERT statements + one helper function. No new tables or views.

## Requirements

| ID      | Requirement                                                                                               |
| ------- | --------------------------------------------------------------------------------------------------------- |
| REQ-420 | Seed input must come from `stats.v_waypoints_source_union` only                                             |
| REQ-421 | `GC` rows from the union view must win over `OC` duplicates via seed ordering and `ON CONFLICT DO NOTHING` |
| REQ-422 | `waypoint_code` must be uppercased before insertion                                                       |
| REQ-423 | `country` must be uppercased before insertion                                                             |
| REQ-424 | Seed is idempotent: re-running inserts 0 rows if all codes already present                                |
| REQ-425 | `stats.fn_seed_waypoints()` must log a row in `stats.job_log` per run                                     |
| REQ-426 | `down()` must truncate `stats.waypoints` (not drop) and drop the function                                 |

## Acceptance Criteria

| Criterion           | Verification                                                                           |
| ------------------- | -------------------------------------------------------------------------------------- |
| Table populated     | `SELECT COUNT(*) FROM stats.waypoints` > 100,000                                       |
| GC source present   | `SELECT COUNT(*) FROM stats.waypoints WHERE source='GC'` > 0                           |
| OC source present   | `SELECT COUNT(*) FROM stats.waypoints WHERE source='OC'` > 0                           |
| All codes uppercase | `SELECT COUNT(*) FROM stats.waypoints WHERE waypoint_code != UPPER(waypoint_code)` = 0 |
| Idempotency         | Second `SELECT stats.fn_seed_waypoints()` returns `0`                                  |
| Job log entry       | `SELECT COUNT(*) FROM stats.job_log WHERE job_name='fn_seed_waypoints'` > 0            |

## Migration File

**`20260310400200_seed_waypoints.php`**

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class SeedWaypoints extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
INSERT INTO stats.waypoints (waypoint_code, source, lat, lon, country, first_seen_at)
SELECT
  waypoint_code,
  source,
  lat,
  lon,
  country,
  now()
FROM stats.v_waypoints_source_union
ORDER BY CASE source WHEN 'GC' THEN 0 ELSE 1 END, waypoint_code
ON CONFLICT (waypoint_code) DO NOTHING;

CREATE OR REPLACE FUNCTION stats.fn_seed_waypoints()
RETURNS BIGINT
LANGUAGE plpgsql AS $$
DECLARE
  v_inserted BIGINT := 0;
BEGIN
  INSERT INTO stats.waypoints (waypoint_code, source, lat, lon, country, first_seen_at)
  SELECT waypoint_code, source, lat, lon, country, now()
  FROM stats.v_waypoints_source_union
  ORDER BY CASE source WHEN 'GC' THEN 0 ELSE 1 END, waypoint_code
  ON CONFLICT (waypoint_code) DO NOTHING;
  GET DIAGNOSTICS v_inserted = ROW_COUNT;

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES ('fn_seed_waypoints', 'completed',
    jsonb_build_object('inserted', v_inserted), now(), now());

  RAISE NOTICE 'fn_seed_waypoints: inserted % new waypoints', v_inserted;
  RETURN v_inserted;
END;
$$;
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP FUNCTION IF EXISTS stats.fn_seed_waypoints();
TRUNCATE TABLE stats.waypoints;
SQL
        );
    }
}
```

**Note:** `down()` truncates rather than drops the table (the table is dropped by S4T01's rollback). This keeps the rollback idempotent.

## SQL Usage Examples

```sql
-- Run initial seed
SELECT stats.fn_seed_waypoints();

-- Check how many waypoints were loaded by source
SELECT source, COUNT(*) AS total
FROM stats.waypoints
GROUP BY source
ORDER BY source;

-- Check total after seed
SELECT COUNT(*) AS total FROM stats.waypoints;

-- Verify idempotency: second run inserts 0 rows
SELECT stats.fn_seed_waypoints();  -- returns 0

-- Show last job log entry
SELECT * FROM stats.job_log
WHERE job_name = 'fn_seed_waypoints'
ORDER BY started_at DESC LIMIT 1;

-- Count waypoints with full coordinate coverage
SELECT
  COUNT(*)                                           AS total,
  COUNT(*) FILTER (WHERE lat IS NOT NULL)            AS with_coords,
  ROUND(100.0 * COUNT(*) FILTER (WHERE lat IS NOT NULL) / COUNT(*), 2) AS pct_coords
FROM stats.waypoints;
```

## Graph / Visualization Specification

**Unlocked (via stats.waypoints after seeding):**

1. **Waypoint coverage map** — dot-density world map
   - Data: `SELECT lat, lon, source FROM stats.waypoints WHERE lat IS NOT NULL`
   - Layer: one dot per waypoint, coloured by source

2. **Waypoints by country** — Choropleth
   - Data: `SELECT country, COUNT(*) FROM stats.waypoints GROUP BY country`
   - Legend: green intensity = concentration of registered caches

```
ASCII Sample — After seed:
GC waypoints:  241,320
OC waypoints:  104,071 (97,450 new + 6,621 already seeded as GC)
UK waypoints:        0 (none from migration; added live by trigger)
Total:         338,770
```

## TimescaleDB Assessment

**NOT applicable.** Seed operation — not time-series.

## pgTAP Unit Tests

| Test ID   | Assertion                                                        | Expected             |
| --------- | ---------------------------------------------------------------- | -------------------- |
| T-4.3.001 | `stats.waypoints` row count > 100,000 after seed                 | pass                 |
| T-4.3.002 | `GC` source rows > 0                                             | pass                 |
| T-4.3.003 | `OC` source rows > 0                                             | pass                 |
| T-4.3.004 | No row has `waypoint_code != UPPER(waypoint_code)`               | pass                 |
| T-4.3.005 | No row has `country != UPPER(country)` where country is not NULL | pass                 |
| T-4.3.006 | `stats.fn_seed_waypoints` function exists                        | pgTAP `has_function` |
| T-4.3.007 | Second run of `fn_seed_waypoints()` returns 0 (idempotent)       | pass                 |
| T-4.3.008 | `stats.job_log` has at least one row for `fn_seed_waypoints`     | pass                 |
| T-4.3.009 | No `NULL` waypoint_code in `stats.waypoints`                     | pass                 |
| T-4.3.010 | Rollback truncates waypoints and drops function                  | phinx rollback       |

| Test ID   | Area     | Description                                        | Method         |
| --------- | -------- | -------------------------------------------------- | -------------- |
| T-4.3.001 | data     | `stats.waypoints` count > 100,000 after seed       | SQL            |
| T-4.3.002 | data     | GC rows > 0                                        | SQL            |
| T-4.3.003 | data     | OC rows > 0                                        | SQL            |
| T-4.3.004 | data     | All `waypoint_code` uppercase                      | SQL            |
| T-4.3.005 | data     | All `country` (non-null) uppercase                 | SQL            |
| T-4.3.006 | schema   | `stats.fn_seed_waypoints` function exists          | pgTAP          |
| T-4.3.007 | behavior | Second `fn_seed_waypoints()` call returns 0        | SQL            |
| T-4.3.008 | audit    | `stats.job_log` row exists for `fn_seed_waypoints` | SQL            |
| T-4.3.009 | data     | No NULL `waypoint_code`                            | SQL            |
| T-4.3.010 | rollback | `down()` truncates table and drops function        | phinx rollback |

## Implementation Checklist

- [ ] 1. Verify S4T01 (table) and S4T02 (view) migrations applied
- [ ] 2. Create migration `20260310400200_seed_waypoints.php`
- [ ] 3. Run `phinx migrate` — confirm no errors
- [ ] 4. Check row count: `SELECT COUNT(*) FROM stats.waypoints` — should be >100,000
- [ ] 5. Check `GC` and `OC` source counts are both >0
- [ ] 6. Verify all `waypoint_code` are uppercase
- [ ] 7. Call `SELECT stats.fn_seed_waypoints()` a second time — verify returns 0 (idempotent)
- [ ] 8. Verify `stats.job_log` has a row for `fn_seed_waypoints`
- [ ] 9. Run pgTAP tests T-4.3.001 through T-4.3.010
- [ ] 10. Verify `phinx rollback` truncates waypoints and drops function

- - [DBA  ] DBA file reviewed: [task-S4T03.dba.md](task-S4T03.md#dba)
- [ ] Dependencies S4T01 and S4T02 applied
- [ ] Migration `20260310400200_seed_waypoints.php` created
- [ ] Migration runs successfully; row count > 100,000
- [ ] Idempotency verified
- [ ] Job log entry verified
- [ ] All T-4.3.xxx tests pass
- [ ] Rollback verified

## Seed INSERT

```sql
INSERT INTO stats.waypoints (waypoint_code, source, lat, lon, country, first_seen_at)
SELECT
  waypoint_code,
  source,
  lat,
  lon,
  country,
  now()
FROM stats.v_waypoints_source_union
ORDER BY CASE source WHEN 'GC' THEN 0 ELSE 1 END, waypoint_code
ON CONFLICT (waypoint_code) DO NOTHING;
```

## Idempotent Seed Function

```sql
CREATE OR REPLACE FUNCTION stats.fn_seed_waypoints()
RETURNS BIGINT
LANGUAGE plpgsql AS $$
DECLARE
  v_inserted BIGINT := 0;
BEGIN
  INSERT INTO stats.waypoints (waypoint_code, source, lat, lon, country, first_seen_at)
  SELECT waypoint_code, source, lat, lon, country, now()
  FROM stats.v_waypoints_source_union
  ORDER BY CASE source WHEN 'GC' THEN 0 ELSE 1 END, waypoint_code
  ON CONFLICT (waypoint_code) DO NOTHING;

  GET DIAGNOSTICS v_inserted = ROW_COUNT;

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES ('fn_seed_waypoints', 'completed',
    jsonb_build_object('inserted', v_inserted), now(), now());

  RAISE NOTICE 'fn_seed_waypoints: inserted % new waypoints', v_inserted;
  RETURN v_inserted;
END;
$$;
```

## Function Contract

**`stats.fn_seed_waypoints() RETURNS BIGINT`**

- Seeds from `stats.v_waypoints_source_union` only, with `GC` rows ordered ahead of `OC`
- Returns total rows inserted (0 on re-run)
- Logs result to `stats.job_log`
- Emits `RAISE NOTICE` with count

## Agent Loop Log

- 2026-03-10T18:16:57Z - Loop 1 - `dba`: Accepted union-view-only seed path with explicit source ordering so `GC` precedence remains deterministic.
- 2026-03-10T18:16:57Z - Loop 1 - `critical-thinking`: No blocking contradiction; removing direct-table SQL eliminates duplicated seed logic and keeps S4T03 aligned with S4T02.
- 2026-03-10T18:16:57Z - Loop 1 - `specification`: Rewrote migration/helper SQL to consume `stats.v_waypoints_source_union`, removed the draft conflict, and preserved acceptance criteria semantics.

## Resolution

- Waypoint seeding is canonicalized to the union view only; see Q-026 reference update in `99-OPEN-QUESTIONS.md`.
