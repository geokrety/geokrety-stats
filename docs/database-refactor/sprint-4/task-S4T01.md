---
title: "Task S4T01: Create stats.waypoints Table"
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
  - table
  - task-index
  - task-merge
  - waypoints
depends_on:
  - "Sprint 1"
  - "Sprint 2"
task: S4T01
step: 4.1
migration: 20260310400000_create_waypoints.php
blocks:
  - S4T02
  - S4T03
  - S4T04
  - S4T05
  - S4T08
changelog:
  - 2026-03-10: created by merge of task-S4T01.dba.md and task-S4T01.specification.md
---

# Task S4T01: Create stats.waypoints Table

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T01.dba.md`
- Specification source: `task-S4T01.specification.md`

## Purpose & Scope

Creates the `stats.waypoints` canonical lookup table. Every distinct waypoint encountered in `geokrety.gk_moves.waypoint` is normalised, deduplicated, and stored here with coordinates, country, and source provenance. This table enables O(1) cache lookups in `gk_cache_visits` and `user_cache_visits` without scanning raw move rows.

The `source` column tracks where the waypoint record originated:

- `GC` = seeded from `geokrety.gk_waypoints_gc`
- `OC` = seeded from `geokrety.gk_waypoints_oc`
- `UK` = first encountered in the `gk_moves` stream (source unknown)

Create the canonical waypoint lookup table `stats.waypoints`. This table is the foundation for cache-analytics: every move that references a waypoint string (e.g., `GC1A2B3`, `OK12345`) resolves against this table and obtains a stable surrogate integer `id`. Downstream tables `stats.gk_cache_visits` and `stats.user_cache_visits` use this FK, eliminating repeated string comparisons and enabling O(1) join lookups.

**Scope:** Create one table with two constraints. No triggers. No functions. No seed data in this step (seeding is S4T03).

## Definitions

| Term             | Definition                                                                                      |
| ---------------- | ----------------------------------------------------------------------------------------------- |
| `waypoint_code`  | Uppercase normalised waypoint identifier, e.g. `GC1A2B3`, `OKXXXX`                              |
| `source`         | Provenance of the waypoint record: `GC` (geocaching.com), `OC` (opencaching), `UK` (unknown)    |
| Surrogate key    | Auto-increment `BIGSERIAL` `id` used by FK references in `gk_cache_visits`, `user_cache_visits` |
| Canonical lookup | First occurrence wins; subsequent moves with the same code UPSERT on `waypoint_code` conflict   |

## Requirements

| ID      | Requirement                                                                                           |
| ------- | ----------------------------------------------------------------------------------------------------- |
| REQ-401 | Table `stats.waypoints` must exist in the `stats` schema with the 7 columns specified in the DBA file |
| REQ-402 | `waypoint_code` must have a UNIQUE constraint (`uq_waypoints_code`)                                   |
| REQ-403 | `source` must be constrained to `'GC'`, `'OC'`, or `'UK'` via a CHECK constraint                      |
| REQ-404 | `id` must be a `BIGSERIAL` primary key (auto-increment)                                               |
| REQ-405 | `lat`, `lon`, `country` must be nullable to support `UK`-sourced waypoints with unknown coordinates   |
| REQ-406 | Migration `down()` must drop the table with `CASCADE` to cleanly remove downstream FK dependencies    |
| REQ-407 | The table must NOT contain any rows after creation (seeding is a separate step, S4T03)                |

## Acceptance Criteria

| Criterion            | Verification                                                                                                  |
| -------------------- | ------------------------------------------------------------------------------------------------------------- |
| Table exists         | `SELECT 1 FROM information_schema.tables WHERE table_schema='stats' AND table_name='waypoints'` returns 1 row |
| Unique constraint    | Inserting duplicate `waypoint_code` raises `unique_violation` (SQLSTATE 23505)                                |
| CHECK constraint     | Inserting `source='ZZ'` raises `check_violation` (SQLSTATE 23514)                                             |
| Nullable coords      | Insert with `lat=NULL, lon=NULL` succeeds                                                                     |
| Empty after creation | `SELECT COUNT(*) FROM stats.waypoints` = 0                                                                    |
| Rollback             | `phinx rollback` removes the table and all FK constraints cascade-dropped                                     |

## Migration File

**`20260310400000_create_waypoints.php`**

## Full SQL DDL

```sql
CREATE TABLE stats.waypoints (
  id             BIGSERIAL    PRIMARY KEY,
  waypoint_code  VARCHAR(20)  NOT NULL,
  source         CHAR(2)      NOT NULL DEFAULT 'UK',
  lat            DOUBLE PRECISION,
  lon            DOUBLE PRECISION,
  country        CHAR(2),
  first_seen_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  CONSTRAINT uq_waypoints_code UNIQUE (waypoint_code),
  CONSTRAINT chk_waypoints_source CHECK (source IN ('GC','OC','UK'))
);

COMMENT ON TABLE stats.waypoints
  IS 'Canonical deduplicated waypoint registry; each distinct waypoint code appears exactly once';
COMMENT ON COLUMN stats.waypoints.waypoint_code
  IS 'Uppercase normalised waypoint identifier, e.g. GC1A2B3, OKXXXX';
COMMENT ON COLUMN stats.waypoints.source
  IS 'Provenance: GC=geocaching.com seed, OC=opencaching seed, UK=first seen in move stream';
COMMENT ON COLUMN stats.waypoints.country
  IS 'ISO 3166-1 alpha-2 country code derived from waypoint seed tables; may be NULL for UK-sourced';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateWaypoints extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.waypoints (
  id             BIGSERIAL    PRIMARY KEY,
  waypoint_code  VARCHAR(20)  NOT NULL,
  source         CHAR(2)      NOT NULL DEFAULT 'UK',
  lat            DOUBLE PRECISION,
  lon            DOUBLE PRECISION,
  country        CHAR(2),
  first_seen_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  CONSTRAINT uq_waypoints_code UNIQUE (waypoint_code),
  CONSTRAINT chk_waypoints_source CHECK (source IN ('GC','OC','UK'))
);

COMMENT ON TABLE stats.waypoints
  IS 'Canonical deduplicated waypoint registry; each distinct waypoint code appears exactly once';
COMMENT ON COLUMN stats.waypoints.waypoint_code
  IS 'Uppercase normalised waypoint identifier, e.g. GC1A2B3, OKXXXX';
COMMENT ON COLUMN stats.waypoints.source
  IS 'Provenance: GC=geocaching.com seed, OC=opencaching seed, UK=first seen in move stream';
COMMENT ON COLUMN stats.waypoints.country
  IS 'ISO 3166-1 alpha-2 country code derived from waypoint seed tables; may be NULL for UK-sourced';
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP TABLE IF EXISTS stats.waypoints CASCADE;
SQL
        );
    }
}
```

## Data Contract

| Column          | Type               | Nullable | Default | Description                                             |
| --------------- | ------------------ | -------- | ------- | ------------------------------------------------------- |
| `id`            | `BIGSERIAL`        | NOT NULL | auto    | **PK** — surrogate key used by FK in cache-visit tables |
| `waypoint_code` | `VARCHAR(20)`      | NOT NULL | —       | Uppercase waypoint identifier; unique                   |
| `source`        | `CHAR(2)`          | NOT NULL | `'UK'`  | Provenance: `GC`, `OC`, or `UK`                         |
| `lat`           | `DOUBLE PRECISION` | YES      | NULL    | Latitude from seed table or move coordinates            |
| `lon`           | `DOUBLE PRECISION` | YES      | NULL    | Longitude from seed table or move coordinates           |
| `country`       | `CHAR(2)`          | YES      | NULL    | Country code from seed table; NULL for UK-sourced       |
| `first_seen_at` | `TIMESTAMPTZ`      | NOT NULL | `now()` | When this waypoint was first registered                 |

**Constraints:**

- `uq_waypoints_code` — UNIQUE on `waypoint_code`
- `chk_waypoints_source` — CHECK `source IN ('GC','OC','UK')`

**Table: `stats.waypoints`**

| Column          | Type               | Nullable | Default | Constraint                   |
| --------------- | ------------------ | -------- | ------- | ---------------------------- |
| `id`            | `BIGSERIAL`        | NOT NULL | auto    | PRIMARY KEY                  |
| `waypoint_code` | `VARCHAR(20)`      | NOT NULL | —       | UNIQUE (`uq_waypoints_code`) |
| `source`        | `CHAR(2)`          | NOT NULL | `'UK'`  | CHECK `IN ('GC','OC','UK')`  |
| `lat`           | `DOUBLE PRECISION` | YES      | NULL    |                              |
| `lon`           | `DOUBLE PRECISION` | YES      | NULL    |                              |
| `country`       | `CHAR(2)`          | YES      | NULL    |                              |
| `first_seen_at` | `TIMESTAMPTZ`      | NOT NULL | `now()` |                              |

## SQL Usage Examples

```sql
-- Look up a waypoint by code (case-insensitive lookup helper)
SELECT id, waypoint_code, source, lat, lon, country
FROM stats.waypoints
WHERE waypoint_code = UPPER('gc1a2b3');

-- Count waypoints by source
SELECT source, COUNT(*) AS total
FROM stats.waypoints
GROUP BY source
ORDER BY source;

-- Find waypoints with no coordinates (UK-sourced, need enrichment)
SELECT waypoint_code, first_seen_at
FROM stats.waypoints
WHERE lat IS NULL AND source = 'UK'
ORDER BY first_seen_at DESC
LIMIT 50;

-- Count total registered waypoints
SELECT COUNT(*) AS total_waypoints FROM stats.waypoints;

-- Waypoints in a specific country
SELECT waypoint_code, source
FROM stats.waypoints
WHERE country = 'PL'
ORDER BY waypoint_code;
```

## Graph / Visualization Specification

**Unlocked visualisations:**

1. **Waypoints by source** — Donut chart
   - Data: `SELECT source, COUNT(*) FROM stats.waypoints GROUP BY source`
   - Segments: GC (green), OC (blue), UK (orange)

2. **Cache discovery over time** — Line chart
   - Data: `SELECT DATE_TRUNC('month', first_seen_at) AS month, COUNT(*) FROM stats.waypoints GROUP BY 1 ORDER BY 1`
   - x-axis: months, y-axis: new waypoints registered

```
ASCII Sample — Waypoints by Source:
GC  ████████████████████ 64%   (241,000)
OC  ████████             26%   (97,000)
UK  ███                  10%   (37,000)
Total: 375,000 waypoints
```

## TimescaleDB Assessment

**NOT recommended.** `stats.waypoints` is a dimension table with a surrogate `BIGSERIAL` PK, not a time-series. `first_seen_at` is informational, not a partitioning key. The table grows by insertion of new unique codes (rare after initial seeding). Standard PostgreSQL with a B-tree unique index on `waypoint_code` is sufficient. TimescaleDB would add overhead with no benefit.

## pgTAP Unit Tests

| Test ID   | Assertion                                                                     | Expected         |
| --------- | ----------------------------------------------------------------------------- | ---------------- |
| T-4.1.001 | `has_table('stats', 'waypoints')`                                             | pass             |
| T-4.1.002 | `col_is_pk('stats', 'waypoints', ARRAY['id'])`                                | pass             |
| T-4.1.003 | `col_type_is('stats', 'waypoints', 'waypoint_code', 'character varying(20)')` | pass             |
| T-4.1.004 | `col_type_is('stats', 'waypoints', 'source', 'character(2)')`                 | pass             |
| T-4.1.005 | `col_default_is('stats', 'waypoints', 'source', 'UK')`                        | pass             |
| T-4.1.006 | `col_type_is('stats', 'waypoints', 'country', 'character(2)')`                | pass             |
| T-4.1.007 | `col_is_null('stats', 'waypoints', 'lat')`                                    | pass (nullable)  |
| T-4.1.008 | Unique constraint `uq_waypoints_code` exists                                  | pass             |
| T-4.1.009 | CHECK constraint `chk_waypoints_source` rejects invalid value (e.g., `'ZZ'`)  | raises exception |
| T-4.1.010 | Duplicate `waypoint_code` insert raises unique violation                      | raises exception |
| T-4.1.011 | `DOWN` migration drops table and cascade removes dependent FKs                | pass             |

| Test ID   | Area     | Description                                       | Method         |
| --------- | -------- | ------------------------------------------------- | -------------- |
| T-4.1.001 | schema   | Table `stats.waypoints` exists                    | pgTAP          |
| T-4.1.002 | schema   | PK is `(id)`                                      | pgTAP          |
| T-4.1.003 | schema   | `waypoint_code` is `VARCHAR(20)`                  | pgTAP          |
| T-4.1.004 | schema   | `source` is `CHAR(2)`                             | pgTAP          |
| T-4.1.005 | schema   | `source` default is `'UK'`                        | pgTAP          |
| T-4.1.006 | schema   | `country` is `CHAR(2)` nullable                   | pgTAP          |
| T-4.1.007 | schema   | `lat` is nullable                                 | pgTAP          |
| T-4.1.008 | schema   | Unique constraint `uq_waypoints_code` exists      | pgTAP          |
| T-4.1.009 | data     | Invalid source raises CHECK violation             | SQL execute    |
| T-4.1.010 | data     | Duplicate `waypoint_code` raises UNIQUE violation | SQL execute    |
| T-4.1.011 | rollback | Migration `down()` removes table cleanly          | phinx rollback |

## Implementation Checklist

- [ ] 1. Create migration file `20260310400000_create_waypoints.php`
- [ ] 2. Verify PHP class name is `CreateWaypoints` (matches Phinx convention)
- [ ] 3. Run `phinx migrate` — confirm no errors
- [ ] 4. `\d stats.waypoints` — verify all 7 columns present with correct types
- [ ] 5. Verify `uq_waypoints_code` UNIQUE constraint exists
- [ ] 6. Verify `chk_waypoints_source` CHECK constraint exists
- [ ] 7. Test CHECK: `INSERT INTO stats.waypoints (waypoint_code, source) VALUES ('TEST', 'ZZ')` → expect error
- [ ] 8. Test UNIQUE: insert same `waypoint_code` twice → expect error
- [ ] 9. Run pgTAP tests T-4.1.001 through T-4.1.011
- [ ] 10. Verify `phinx rollback` drops the table cleanly (no orphaned objects)

- - [DBA  ] DBA file reviewed and approved: [task-S4T01.dba.md](task-S4T01.md#dba)
- [ ] Migration file `20260310400000_create_waypoints.php` created
- [ ] Migration runs successfully (`phinx migrate`)
- [ ] All T-4.1.xxx pgTAP tests pass
- [ ] Migration rolls back successfully (`phinx rollback`)
- [ ] Spec sign-off: table matches data contract above
