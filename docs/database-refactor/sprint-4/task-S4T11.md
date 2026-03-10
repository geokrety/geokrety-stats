---
title: "Task S4T11: Snapshot Functions for Waypoints & Relations"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 4
tags:
  - backfill
  - database
  - database-refactor
  - dba
  - function
  - snapshot
  - specification
  - sprint-4
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S4T01
  - S4T04
  - S4T05
  - S4T06
  - S4T07
  - S4T10
task: S4T11
step: 4.11
migration: 20260310401000_create_waypoint_relation_snapshots.php
changelog:
  - 2026-03-10: created by merge of task-S4T11.dba.md and task-S4T11.specification.md
  - 2026-03-10: logged the canonical snapshot-helper scope conflict
---

# Task S4T11: Snapshot Functions for Waypoints & Relations

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T11.dba.md`
- Specification source: `task-S4T11.specification.md`

## Resolved Decision

- Q-013 promotes the focused Sprint 4 snapshot helpers into the canonical spec.
- Sprint 6 still expects the master wrapper name `stats.fn_snapshot_relationship_tables(p_period daterange DEFAULT NULL)`.
- Sprint 4 therefore canonizes both the three concrete helpers and the wrapper facade.
- The three helpers are the authoritative implementation surface; the wrapper is a thin orchestration facade kept for master-spec continuity.

## Purpose & Scope

Creates **idempotent snapshot / backfill functions** that can be safely called multiple times to populate `stats.waypoints`, `stats.gk_cache_visits`, `stats.user_cache_visits`, `stats.gk_related_users`, and `stats.user_related_users` from the existing `geokrety.gk_moves` history.

These functions are called during the Sprint 6 full backfill phase. They are also safe to run in production as delta-repair functions (e.g. after a trigger bug fix or partial import).

Each function:

- Uses `ON CONFLICT DO UPDATE` / `DO NOTHING` so re-running is safe
- Logs progress into `stats.job_log`
- Returns the number of rows affected

Four functions are created:

1. `stats.fn_snapshot_waypoints()` — seeds `stats.waypoints` from `gk_moves`
2. `stats.fn_snapshot_cache_visits()` — populates `stats.gk_cache_visits` + `stats.user_cache_visits`
3. `stats.fn_snapshot_relations()` — populates `stats.gk_related_users` + `stats.user_related_users`
4. `stats.fn_snapshot_relationship_tables(p_period daterange DEFAULT NULL)` — orchestration wrapper over the three helpers

The helper-scope question is now resolved: waypoint, cache-visit, and relation snapshot helpers are promoted into the canonical Sprint 4 contract, while `stats.fn_snapshot_relationship_tables(...)` remains the stable master-spec entry point.

**Execution order must be respected:** follow the master snapshot order from [00-SPEC-DRAFT-v1.md](../00-SPEC-DRAFT-v1.md), with relationship-table seeding occurring after waypoint/cache prerequisites are available.

---

## Requirements

| ID      | Description                                                                                       | MoSCoW |
| ------- | ------------------------------------------------------------------------------------------------- | ------ |
| REQ-510 | `stats.fn_snapshot_waypoints()`, `stats.fn_snapshot_cache_visits()`, `stats.fn_snapshot_relations()`, and `stats.fn_snapshot_relationship_tables(p_period daterange DEFAULT NULL)` all exist | MUST   |
| REQ-511 | The promoted helper set is idempotent: re-running produces the same final rows without duplicates    | MUST   |
| REQ-512 | Snapshot helpers only process qualifying rows (`waypoint IS NOT NULL`, `move_type` filters, `author IS NOT NULL` where required) | MUST   |
| REQ-513 | `stats.fn_snapshot_relations()` preserves canonical `shared_geokrety_count` distinct-shared-GK semantics | MUST   |
| REQ-514 | All snapshot functions write canonical `stats.job_log` metadata using `job_name`, `status`, `metadata`, `started_at`, `completed_at` | MUST   |
| REQ-515 | `stats.fn_snapshot_relationship_tables(...)` is a thin facade over the promoted helper set; `phinx rollback` drops wrapper and helpers cleanly | MUST   |

---

## Acceptance Criteria

| #   | Criterion                                                                   | How to Verify                                                            |
| --- | --------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| 1   | Promoted helper set and wrapper all exist in `stats` schema                 | `\df stats.fn_snapshot_*`                                               |
| 2   | Running the helper set twice produces the same final row counts             | Run, record counts, run again; counts equal                              |
| 3   | Anonymous moves are not represented in user relationships                   | `SELECT * FROM stats.user_related_users WHERE user_id IS NULL` → 0        |
| 4   | `user_related_users` symmetry check returns 0 asymmetric pairs              | See symmetry query in dba.md                                             |
| 5   | Wrapper delegates to the promoted helper set                                | Execute wrapper; verify rows and matching `job_log` entries               |
| 6   | `stats.job_log` contains canonical execution rows for helper calls          | `SELECT COUNT(*) FROM stats.job_log WHERE job_name LIKE 'fn_snapshot_%'` |
| 7   | Rollback drops wrapper and helpers                                          | None present in `\df stats.fn_snapshot_*` after rollback                 |

---

## Migration File

**`20260310401000_create_waypoint_relation_snapshots.php`**

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateWaypointRelationSnapshots extends AbstractMigration
{
    public function up(): void
    {
  // NOTE: the three promoted snapshot helpers and the wrapper facade
  // are created inline here. See the DDL sections below for canonical bodies.
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION stats.fn_snapshot_waypoints() RETURNS BIGINT ...;
CREATE OR REPLACE FUNCTION stats.fn_snapshot_cache_visits() RETURNS BIGINT ...;
CREATE OR REPLACE FUNCTION stats.fn_snapshot_relations() RETURNS BIGINT ...;
CREATE OR REPLACE FUNCTION stats.fn_snapshot_relationship_tables(p_period daterange DEFAULT NULL) RETURNS BIGINT ...;
SQL
        );
  // In practice: paste full function bodies from the DDL sections below
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP FUNCTION IF EXISTS stats.fn_snapshot_waypoints();
DROP FUNCTION IF EXISTS stats.fn_snapshot_cache_visits();
DROP FUNCTION IF EXISTS stats.fn_snapshot_relations();
DROP FUNCTION IF EXISTS stats.fn_snapshot_relationship_tables(daterange);
SQL
        );
    }
}
```

## SQL Usage Examples

```sql
-- Run snapshots in order
SELECT stats.fn_snapshot_waypoints();
SELECT stats.fn_snapshot_cache_visits();
SELECT stats.fn_snapshot_relations();
SELECT stats.fn_snapshot_relationship_tables();

-- Check how many rows were written
SELECT * FROM stats.job_log
WHERE job_name LIKE 'fn_snapshot_%'
ORDER BY completed_at DESC
LIMIT 10;

-- Verify symmetry after relations snapshot
SELECT COUNT(*) AS asymmetric_pairs
FROM stats.user_related_users a
WHERE NOT EXISTS (
  SELECT 1 FROM stats.user_related_users b
  WHERE b.user_id = a.related_user_id
    AND b.related_user_id = a.user_id
);
-- Must return 0
```

## pgTAP Unit Tests

| Test ID    | Assertion                                                                        | Expected |
| ---------- | -------------------------------------------------------------------------------- | -------- |
| T-4.11.001 | Function `stats.fn_snapshot_waypoints()` exists                                  | pass     |
| T-4.11.002 | Function `stats.fn_snapshot_cache_visits()` exists                               | pass     |
| T-4.11.003 | Function `stats.fn_snapshot_relations()` exists                                  | pass     |
| T-4.11.004 | Function `stats.fn_snapshot_relationship_tables()` exists                        | pass     |
| T-4.11.005 | All promoted helpers and the wrapper return `BIGINT`                             | pass     |
| T-4.11.006 | Helper set is idempotent                                                         | pass     |
| T-4.11.007 | `fn_snapshot_waypoints()` stores `waypoint_code` as UPPER()                      | pass     |
| T-4.11.008 | `fn_snapshot_relations()` skips `move_type IN (2,4)` and null authors            | pass     |
| T-4.11.009 | `fn_snapshot_relations()` produces symmetric `user_related_users` pairs          | pass     |
| T-4.11.010 | Wrapper execution writes canonical `stats.job_log` rows                          | pass     |
| T-4.11.011 | `phinx rollback` drops wrapper and helpers                                       | pass     |

| Test ID    | Scenario                                     | Pass Condition                |
| ---------- | -------------------------------------------- | ----------------------------- |
| T-4.11.001 | `fn_snapshot_relationship_tables` exists     | `has_function()` pgTAP call   |
| T-4.11.002 | Wrapper delegates to promoted helpers        | Matching helper and wrapper outputs |
| T-4.11.003 | Helper set is idempotent                     | Row count stable on 2nd run   |
| T-4.11.004 | COMMENT + anon moves excluded from relations | Row absence check             |
| T-4.11.005 | `user_related_users` symmetry after snapshot | 0 asymmetric pairs            |
| T-4.11.006 | `stats.job_log` written after helper call    | 1 new log row                 |
| T-4.11.007 | Rollback removes wrapper and helpers         | `hasnt_function()` pgTAP pass |

---

## Implementation Checklist

- [ ] 1. Create `20260310401000_create_waypoint_relation_snapshots.php` with full function bodies
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. `\df stats.fn_snapshot_*` — helper set and wrapper present
- [ ] 4. Run `SELECT stats.fn_snapshot_waypoints()` on a dev DB with sample moves — verify rows
- [ ] 5. Run `SELECT stats.fn_snapshot_cache_visits()` — verify cache visit rows
- [ ] 6. Run `SELECT stats.fn_snapshot_relations()` — verify relation rows + symmetry
- [ ] 7. Re-run each function (idempotency test) — counts stable, no duplicates
- [ ] 8. Check `stats.job_log` entries after each run
- [ ] 9. Run pgTAP T-4.11.001 through T-4.11.011
- [ ] 10. `phinx rollback` — wrapper and helpers dropped

- [ ] 1. Write `20260310401000_create_waypoint_relation_snapshots.php` with the canonical helper body
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. `\df stats.fn_snapshot_relationship_tables` — wrapper listed
- [ ] 4. Run functions in order on dev DB; check row counts
- [ ] 5. Re-run for idempotency; counts match
- [ ] 6. Verify COMMENT and anonymous excludes
- [ ] 7. Verify `user_related_users` symmetry
- [ ] 8. Check `stats.job_log` entries
- [ ] 9. Run pgTAP T-4.11.001 through T-4.11.007 — all pass
- [ ] 10. `phinx rollback` — canonical helper gone
- [ ] 11. **Sprint 4 complete** — ready to proceed to Sprint 5

## `stats.fn_snapshot_waypoints()`

```sql
CREATE OR REPLACE FUNCTION stats.fn_snapshot_waypoints()
  RETURNS BIGINT
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
DECLARE
  v_rows BIGINT;
  v_started_at TIMESTAMPTZ := clock_timestamp();
BEGIN
  INSERT INTO stats.waypoints (waypoint_code, source, lat, lon, country, first_seen_at)
  SELECT
    UPPER(m.waypoint) AS waypoint_code,
    'UK' AS source,
    MAX(m.lat) FILTER (WHERE m.lat IS NOT NULL) AS lat,
    MAX(m.lon) FILTER (WHERE m.lon IS NOT NULL) AS lon,
    MAX(UPPER(m.country)) FILTER (WHERE m.country IS NOT NULL AND BTRIM(m.country) <> '') AS country,
    MIN(m.moved_on_datetime) AS first_seen_at
  FROM geokrety.gk_moves m
  WHERE m.waypoint IS NOT NULL
    AND m.move_type <> 2
  GROUP BY UPPER(m.waypoint)
  ON CONFLICT (waypoint_code) DO UPDATE SET
    lat = COALESCE(stats.waypoints.lat, EXCLUDED.lat),
    lon = COALESCE(stats.waypoints.lon, EXCLUDED.lon),
    country = COALESCE(stats.waypoints.country, EXCLUDED.country),
    first_seen_at = LEAST(stats.waypoints.first_seen_at, EXCLUDED.first_seen_at);

  GET DIAGNOSTICS v_rows = ROW_COUNT;

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES (
    'fn_snapshot_waypoints',
    'completed',
    jsonb_build_object('rows_affected', v_rows),
    v_started_at,
    clock_timestamp()
  );

  RETURN v_rows;
END;
$$;
```

## `stats.fn_snapshot_cache_visits()`

```sql
CREATE OR REPLACE FUNCTION stats.fn_snapshot_cache_visits()
  RETURNS BIGINT
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
DECLARE
  v_rows BIGINT := 0;
  v_partial BIGINT := 0;
  v_started_at TIMESTAMPTZ := clock_timestamp();
BEGIN
  TRUNCATE TABLE stats.gk_cache_visits, stats.user_cache_visits;

  INSERT INTO stats.gk_cache_visits (
    gk_id, waypoint_id, visit_count, first_visited_at, last_visited_at
  )
  SELECT
    m.geokret_id,
    w.id,
    COUNT(*)::BIGINT,
    MIN(m.moved_on_datetime),
    MAX(m.moved_on_datetime)
  FROM geokrety.gk_moves m
  JOIN stats.waypoints w ON w.waypoint_code = UPPER(m.waypoint)
  WHERE m.waypoint IS NOT NULL
    AND m.move_type <> 2
  GROUP BY m.geokret_id, w.id;

  GET DIAGNOSTICS v_partial = ROW_COUNT;
  v_rows := v_rows + v_partial;

  INSERT INTO stats.user_cache_visits (
    user_id, waypoint_id, visit_count, first_visited_at, last_visited_at
  )
  SELECT
    m.author,
    w.id,
    COUNT(*)::BIGINT,
    MIN(m.moved_on_datetime),
    MAX(m.moved_on_datetime)
  FROM geokrety.gk_moves m
  JOIN stats.waypoints w ON w.waypoint_code = UPPER(m.waypoint)
  WHERE m.waypoint IS NOT NULL
    AND m.author IS NOT NULL
    AND m.move_type <> 2
  GROUP BY m.author, w.id;

  GET DIAGNOSTICS v_partial = ROW_COUNT;
  v_rows := v_rows + v_partial;

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES (
    'fn_snapshot_cache_visits',
    'completed',
    jsonb_build_object('rows_affected', v_rows),
    v_started_at,
    clock_timestamp()
  );

  RETURN v_rows;
END;
$$;
```

## `stats.fn_snapshot_relations()`

```sql
CREATE OR REPLACE FUNCTION stats.fn_snapshot_relations()
  RETURNS BIGINT
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
DECLARE
  v_rows BIGINT := 0;
  v_partial BIGINT := 0;
  v_started_at TIMESTAMPTZ := clock_timestamp();
BEGIN
  TRUNCATE TABLE stats.gk_related_users, stats.user_related_users;

  INSERT INTO stats.gk_related_users (
    geokrety_id, user_id, interaction_count, first_interaction, last_interaction
  )
  SELECT
    m.geokret_id,
    m.author,
    COUNT(*)::BIGINT,
    MIN(m.moved_on_datetime),
    MAX(m.moved_on_datetime)
  FROM geokrety.gk_moves m
  WHERE m.author IS NOT NULL
    AND m.move_type IN (0, 1, 3, 5)
  GROUP BY m.geokret_id, m.author;

  GET DIAGNOSTICS v_partial = ROW_COUNT;
  v_rows := v_rows + v_partial;

  INSERT INTO stats.user_related_users (
    user_id, related_user_id, shared_geokrety_count, first_seen_at, last_seen_at
  )
  SELECT
    a.user_id,
    b.user_id AS related_user_id,
    COUNT(DISTINCT a.geokrety_id)::BIGINT,
    MIN(LEAST(a.first_interaction, b.first_interaction)),
    MAX(GREATEST(a.last_interaction, b.last_interaction))
  FROM stats.gk_related_users a
  JOIN stats.gk_related_users b
    ON a.geokrety_id = b.geokrety_id
   AND a.user_id <> b.user_id
  GROUP BY a.user_id, b.user_id;

  GET DIAGNOSTICS v_partial = ROW_COUNT;
  v_rows := v_rows + v_partial;

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES (
    'fn_snapshot_relations',
    'completed',
    jsonb_build_object('rows_affected', v_rows),
    v_started_at,
    clock_timestamp()
  );

  RETURN v_rows;
END;
$$;
```

## `stats.fn_snapshot_relationship_tables(p_period daterange DEFAULT NULL)`

```sql
CREATE OR REPLACE FUNCTION stats.fn_snapshot_relationship_tables(p_period daterange DEFAULT NULL)
  RETURNS BIGINT
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
DECLARE
  v_rows BIGINT := 0;
  v_started_at TIMESTAMPTZ := clock_timestamp();
BEGIN
  -- Sprint 4 canonizes a thin facade: the wrapper keeps the stable master-spec
  -- entry point while delegating to the promoted concrete helpers.
  v_rows := v_rows + stats.fn_snapshot_waypoints();
  v_rows := v_rows + stats.fn_snapshot_cache_visits();
  v_rows := v_rows + stats.fn_snapshot_relations();

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES (
    'fn_snapshot_relationship_tables',
    'completed',
    jsonb_build_object(
      'rows_affected', v_rows,
      'requested_period', p_period,
      'delegated_helpers', jsonb_build_array(
        'fn_snapshot_waypoints',
        'fn_snapshot_cache_visits',
        'fn_snapshot_relations'
      )
    ),
    v_started_at,
    clock_timestamp()
  );

  RETURN v_rows;
END;
$$;
```

## Execution Order (Sprint 6 Backfill)

```
1. stats.fn_snapshot_waypoints()        -- fast, seeds waypoints first
2. stats.fn_snapshot_cache_visits()     -- depends on waypoints being seeded
3. stats.fn_snapshot_relations()        -- can run in parallel with cache visits if needed
4. stats.fn_snapshot_relationship_tables() -- orchestration facade over the promoted helper set
```

## Master-Spec Alignment

This task is governed by [../00-SPEC-DRAFT-v1.md](../00-SPEC-DRAFT-v1.md), Sections 8.3 and 11.

- The canonical Sprint 4 contract now includes both the promoted helper set (`fn_snapshot_waypoints`, `fn_snapshot_cache_visits`, `fn_snapshot_relations`) and the stable wrapper `stats.fn_snapshot_relationship_tables(p_period daterange DEFAULT NULL)`.
- `stats.job_log` columns are `id`, `job_name`, `status`, `metadata`, `started_at`, and `completed_at`.
- `stats.fn_snapshot_relationship_tables(...)` is a thin facade kept for master-spec continuity and Sprint 6 orchestration compatibility; it must not diverge behaviorally from the promoted helpers.
- Any lower text that asserts alternate `stats.job_log` columns or wrapper-only helper scope is obsolete and superseded by this alignment block.

## Functions Created

| Function                           | Returns | Side-effects                                                 |
| ---------------------------------- | ------- | ------------------------------------------------------------ |
| `stats.fn_snapshot_waypoints()` | BIGINT  | Upserts newly discovered move-derived waypoints into `stats.waypoints` |
| `stats.fn_snapshot_cache_visits()` | BIGINT  | Rebuilds `stats.gk_cache_visits` and `stats.user_cache_visits` |
| `stats.fn_snapshot_relations()` | BIGINT  | Rebuilds `stats.gk_related_users` and `stats.user_related_users` |
| `stats.fn_snapshot_relationship_tables()` | BIGINT  | Delegates to the promoted helper set and logs orchestration metadata |

## Agent Loop Log

- 2026-03-10T19:05:00Z — `dba`: reconciled the stale helper-only and wrapper-only drafts by keeping the three concrete helpers plus a thin wrapper facade for continuity.
- 2026-03-10T19:05:00Z — `critical-thinking`: used Q-013 and Sprint 6 call sites as the deciding evidence, avoiding another naming break while removing the wrapper-only contradiction.
- 2026-03-10T19:05:00Z — `specification`: updated requirements, tests, migration notes, DDL, and alignment text so S4T11 matches the promoted helper set and canonical `stats.job_log` contract.

## Resolution

Q-013 and Q-030 are resolved here by canonizing both the promoted helper set and the wrapper facade.

---

## Sprint 4 Completion Checkpoint

After S4T11 is deployed and validated:

| Table/Object               | Populated by          | Trigger maintaining live? |
| -------------------------- | --------------------- | ------------------------- |
| `stats.gk_related_users`   | S4T11 + S4T09         | ✅ S4T09                  |
| `stats.user_related_users` | S4T11 + S4T09         | ✅ S4T09                  |

---
