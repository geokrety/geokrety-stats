---
title: "Task S4T02: Create stats.v_waypoints_source_union View"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 4
tags:
  - database
  - database-refactor
  - dba
  - specification
  - sprint-4
  - sql
  - stats
  - task-index
  - task-merge
  - union
  - view
  - waypoints
depends_on:
  - S4T01
task: S4T02
step: 4.2
migration: 20260310400100_create_waypoints_source_view.php
blocks:
  - S4T03
changelog:
  - 2026.03.10: created by merge of task-S4T02.dba.md and task-S4T02.specification.md
---

# Task S4T02: Create stats.v_waypoints_source_union View

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T02.dba.md`
- Specification source: `task-S4T02.specification.md`

## Purpose & Scope

Creates a diagnostic union view over the two existing waypoint source tables (`geokrety.gk_waypoints_gc` and `geokrety.gk_waypoints_oc`). This view is used by the seeding function in S4T03 to select from both sources in a single pass. It is also useful for ad-hoc diagnostics — e.g., verifying coverage gaps, identifying duplicate codes across sources, or reconciling coordinate discrepancies.

This is a **view only** — no data is moved. It does not replace `stats.waypoints`. It is a helper layer for seeding and maintenance.

Create a diagnostic union view `stats.v_waypoints_source_union` that merges both established waypoint source tables (`geokrety.gk_waypoints_gc`, `geokrety.gk_waypoints_oc`) into a single queryable surface. The primary consumer is the seeding function in S4T03. Secondary uses include ad-hoc diagnostics and source-coverage reconciliation.

**Scope:** Create one view. No tables, no triggers, no functions, no data.

## Requirements

| ID      | Requirement                                                                                  |
| ------- | -------------------------------------------------------------------------------------------- |
| REQ-410 | View must exist in the `stats` schema                                                        |
| REQ-411 | View must include columns: `waypoint_code`, `source`, `lat`, `lon`, `country`                |
| REQ-412 | `waypoint_code` must be uppercased via `UPPER()` in the view definition                      |
| REQ-413 | `country` must be uppercased via `UPPER()` in the view definition                            |
| REQ-414 | View must use `UNION ALL` to preserve source-level duplicates for diagnostic purposes        |
| REQ-415 | Rows with `NULL` waypoint in source tables must be excluded via `WHERE waypoint IS NOT NULL` |
| REQ-416 | `down()` must drop the view with `IF EXISTS`                                                 |

## Acceptance Criteria

| Criterion                | Verification                                                                                                              |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------- |
| View exists              | `\dv stats.v_waypoints_source_union` shows the view                                                                       |
| Columns match contract   | `SELECT column_name FROM information_schema.columns WHERE table_schema='stats' AND table_name='v_waypoints_source_union'` |
| Both sources represented | `SELECT DISTINCT source FROM stats.v_waypoints_source_union` returns `GC` and `OC`                                        |
| All codes uppercase      | `SELECT COUNT(*) FROM stats.v_waypoints_source_union WHERE waypoint_code != UPPER(waypoint_code)` = 0                     |
| NULL waypoints excluded  | `SELECT COUNT(*) FROM stats.v_waypoints_source_union WHERE waypoint_code IS NULL` = 0                                     |

## Migration File

**`20260310400100_create_waypoints_source_view.php`**

## Full SQL DDL

```sql
CREATE OR REPLACE VIEW stats.v_waypoints_source_union AS
SELECT
  UPPER(waypoint)  AS waypoint_code,
  'GC'::CHAR(2)   AS source,
  lat::DOUBLE PRECISION,
  lon::DOUBLE PRECISION,
  UPPER(country)::CHAR(2) AS country
FROM geokrety.gk_waypoints_gc
WHERE waypoint IS NOT NULL

UNION ALL

SELECT
  UPPER(waypoint)  AS waypoint_code,
  'OC'::CHAR(2)   AS source,
  lat::DOUBLE PRECISION,
  lon::DOUBLE PRECISION,
  UPPER(country)::CHAR(2) AS country
FROM geokrety.gk_waypoints_oc
WHERE waypoint IS NOT NULL;

COMMENT ON VIEW stats.v_waypoints_source_union
  IS 'Union of GC and OC waypoint tables for seeding and diagnostics; does not deduplicate';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateWaypointsSourceView extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE VIEW stats.v_waypoints_source_union AS
SELECT
  UPPER(waypoint)  AS waypoint_code,
  'GC'::CHAR(2)   AS source,
  lat::DOUBLE PRECISION,
  lon::DOUBLE PRECISION,
  UPPER(country)::CHAR(2) AS country
FROM geokrety.gk_waypoints_gc
WHERE waypoint IS NOT NULL

UNION ALL

SELECT
  UPPER(waypoint)  AS waypoint_code,
  'OC'::CHAR(2)   AS source,
  lat::DOUBLE PRECISION,
  lon::DOUBLE PRECISION,
  UPPER(country)::CHAR(2) AS country
FROM geokrety.gk_waypoints_oc
WHERE waypoint IS NOT NULL;

COMMENT ON VIEW stats.v_waypoints_source_union
  IS 'Union of GC and OC waypoint tables for seeding and diagnostics; does not deduplicate';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP VIEW IF EXISTS stats.v_waypoints_source_union;');
    }
}
```

## Data Contract

**View: `stats.v_waypoints_source_union`**

| Column          | Type               | Description                                  |
| --------------- | ------------------ | -------------------------------------------- |
| `waypoint_code` | `TEXT`             | Uppercased waypoint identifier               |
| `source`        | `CHAR(2)`          | `'GC'` or `'OC'`                             |
| `lat`           | `DOUBLE PRECISION` | Latitude (may be NULL if missing in source)  |
| `lon`           | `DOUBLE PRECISION` | Longitude (may be NULL if missing in source) |
| `country`       | `CHAR(2)`          | Uppercased ISO country code (may be NULL)    |

## SQL Usage Examples

```sql
-- Total waypoints available from all sources
SELECT source, COUNT(*) AS total
FROM stats.v_waypoints_source_union
GROUP BY source
ORDER BY source;

-- Find waypoints present in both GC and OC (potential duplicates)
SELECT waypoint_code, COUNT(DISTINCT source) AS sources
FROM stats.v_waypoints_source_union
GROUP BY waypoint_code
HAVING COUNT(DISTINCT source) > 1
LIMIT 20;

-- Preview first 10 GC waypoints
SELECT * FROM stats.v_waypoints_source_union
WHERE source = 'GC'
LIMIT 10;

-- Waypoints without coordinates (will result in UK enrichment needed)
SELECT waypoint_code, source
FROM stats.v_waypoints_source_union
WHERE lat IS NULL OR lon IS NULL
LIMIT 20;
```

## TimescaleDB Assessment

**NOT applicable.** This is a view over source tables, not a time-series. No TimescaleDB involvement.

## pgTAP Unit Tests

| Test ID   | Assertion                                                             | Expected |
| --------- | --------------------------------------------------------------------- | -------- |
| T-4.2.001 | `has_view('stats', 'v_waypoints_source_union')`                       | pass     |
| T-4.2.002 | View has column `waypoint_code`                                       | pass     |
| T-4.2.003 | View has column `source`                                              | pass     |
| T-4.2.004 | View has column `lat`                                                 | pass     |
| T-4.2.005 | View has column `lon`                                                 | pass     |
| T-4.2.006 | View has column `country`                                             | pass     |
| T-4.2.007 | View returns at least 2 distinct `source` values (`GC` and `OC`)      | pass     |
| T-4.2.008 | All `waypoint_code` values are uppercase (spot-check 100 rows)        | pass     |
| T-4.2.009 | `DROP VIEW IF EXISTS stats.v_waypoints_source_union` removes the view | pass     |

| Test ID   | Area     | Description                                  | Method         |
| --------- | -------- | -------------------------------------------- | -------------- |
| T-4.2.001 | schema   | View `stats.v_waypoints_source_union` exists | pgTAP          |
| T-4.2.002 | schema   | Column `waypoint_code` present               | pgTAP          |
| T-4.2.003 | schema   | Column `source` present                      | pgTAP          |
| T-4.2.004 | schema   | Column `lat` present                         | pgTAP          |
| T-4.2.005 | schema   | Column `lon` present                         | pgTAP          |
| T-4.2.006 | schema   | Column `country` present                     | pgTAP          |
| T-4.2.007 | data     | Both `GC` and `OC` sources present           | SQL spot-check |
| T-4.2.008 | data     | All `waypoint_code` values uppercase         | SQL assertion  |
| T-4.2.009 | rollback | View dropped by `down()`                     | phinx rollback |

## Implementation Checklist

- [ ] 1. Create migration file `20260310400100_create_waypoints_source_view.php`
- [ ] 2. Verify `geokrety.gk_waypoints_gc` and `geokrety.gk_waypoints_oc` exist with `waypoint`, `lat`, `lon`, `country` columns
- [ ] 3. Run `phinx migrate` — confirm no errors
- [ ] 4. `SELECT source, COUNT(*) FROM stats.v_waypoints_source_union GROUP BY source` — verify both GC and OC appear
- [ ] 5. Verify all `waypoint_code` values are uppercase
- [ ] 6. Run pgTAP tests T-4.2.001 through T-4.2.009
- [ ] 7. Verify `phinx rollback` drops the view cleanly

- - [DBA  ] DBA file reviewed: [task-S4T02.dba.md](task-S4T02.md#dba)
- [ ] Migration `20260310400100_create_waypoints_source_view.php` created
- [ ] Migration runs successfully
- [ ] Both sources visible in view
- [ ] All T-4.2.xxx pgTAP tests pass
- [ ] Migration rolled back successfully

## View Column Reference

| Column          | Type               | Description                                        |
| --------------- | ------------------ | -------------------------------------------------- |
| `waypoint_code` | `TEXT`             | Uppercase waypoint identifier from GC or OC source |
| `source`        | `CHAR(2)`          | `'GC'` or `'OC'`                                   |
| `lat`           | `DOUBLE PRECISION` | Latitude from source table                         |
| `lon`           | `DOUBLE PRECISION` | Longitude from source table                        |
| `country`       | `CHAR(2)`          | Uppercase ISO country code from source             |

**Note:** This view uses `UNION ALL` (not `UNION`) to avoid the cost of deduplication at query time. The seeding function (S4T03) handles deduplication via `ON CONFLICT DO NOTHING`.
