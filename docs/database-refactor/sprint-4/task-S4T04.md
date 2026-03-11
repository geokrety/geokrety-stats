---
title: "Task S4T04: Create stats.gk_cache_visits Table"
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
  - specification
  - sprint-4
  - sql
  - stats
  - table
  - task-index
  - task-merge
depends_on:
  - S4T01
task: S4T04
step: 4.4
migration: 20260310400300_create_gk_cache_visits.php
blocks:
  - S4T08
  - S4T10
  - S4T11
changelog:
  - 2026.03.10: created by merge of task-S4T04.dba.md and task-S4T04.specification.md
---

# Task S4T04: Create stats.gk_cache_visits Table

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T04.dba.md`
- Specification source: `task-S4T04.specification.md`

## Purpose & Scope

Creates `stats.gk_cache_visits`, which records how many times each GeoKret has visited each waypoint/cache, along with first and last visit timestamps. This enables instant answers to:

- "How many caches has this GeoKret visited?"
- "Which caches did this GeoKret visit?"
- "What is the most visited cache by a specific GeoKret?"

Without this table these answers require full scans of `gk_moves`. The FK to `stats.waypoints(id)` ensures only registered waypoints appear here.

Create `stats.gk_cache_visits` to pre-aggregate per-GeoKret per-waypoint visit counts. This eliminates `gk_moves` scans for cache popularity and GK travel analytics.

**Scope:** One table with FK to `stats.waypoints`. No triggers (added in S4T08). No seed data.

## Requirements

| ID      | Requirement                                                                         |
| ------- | ----------------------------------------------------------------------------------- |
| REQ-430 | Table `stats.gk_cache_visits` must exist with 5 columns as specified                |
| REQ-431 | PK must be `(gk_id, waypoint_id)`                                                   |
| REQ-432 | `waypoint_id` must FK to `stats.waypoints(id)` with `DEFERRABLE INITIALLY DEFERRED` |
| REQ-433 | `visit_count` must default to 0                                                     |
| REQ-434 | No FK from `gk_id` to `geokrety.gk_geokrety` (cross-schema FK avoided)              |
| REQ-435 | Table must be empty after creation (seeding in S4T11)                               |
| REQ-436 | `down()` drops the table with `IF EXISTS`                                           |

## Acceptance Criteria

| Criterion            | Verification                                                                                                                                                        |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Table exists         | `\dt stats.gk_cache_visits` returns result                                                                                                                          |
| PK is composite      | `SELECT constraint_name FROM information_schema.table_constraints WHERE table_name='gk_cache_visits' AND constraint_type='PRIMARY KEY'`                             |
| FK deferred          | `SELECT is_deferrable, initially_deferred FROM information_schema.referential_constraints WHERE constraint_name='fk_gk_cache_visits_waypoint'` returns `YES`, `YES` |
| Empty after creation | `SELECT COUNT(*) FROM stats.gk_cache_visits` = 0                                                                                                                    |

## Migration File

**`20260310400300_create_gk_cache_visits.php`**

## Full SQL DDL

```sql
CREATE TABLE stats.gk_cache_visits (
  gk_id             INT          NOT NULL,
  waypoint_id       BIGINT       NOT NULL,
  visit_count       BIGINT       NOT NULL DEFAULT 0,
  first_visited_at  TIMESTAMPTZ  NOT NULL,
  last_visited_at   TIMESTAMPTZ  NOT NULL,
  PRIMARY KEY (gk_id, waypoint_id),
  CONSTRAINT fk_gk_cache_visits_waypoint
    FOREIGN KEY (waypoint_id) REFERENCES stats.waypoints(id)
    DEFERRABLE INITIALLY DEFERRED
);

COMMENT ON TABLE stats.gk_cache_visits
  IS 'Per-GeoKret per-waypoint visit counter; enables cache analytics without gk_moves scans';
COMMENT ON COLUMN stats.gk_cache_visits.gk_id
  IS 'GeoKret internal ID (references geokrety.gk_geokrety.id, not FK to avoid cross-schema dep)';
COMMENT ON COLUMN stats.gk_cache_visits.waypoint_id
  IS 'FK to stats.waypoints(id); surrogated to allow rename/merge without cascading issues';
COMMENT ON COLUMN stats.gk_cache_visits.visit_count
  IS 'Count of moves referencing this waypoint for this GeoKret; incremented by trigger';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkCacheVisits extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.gk_cache_visits (
  gk_id             INT          NOT NULL,
  waypoint_id       BIGINT       NOT NULL,
  visit_count       BIGINT       NOT NULL DEFAULT 0,
  first_visited_at  TIMESTAMPTZ  NOT NULL,
  last_visited_at   TIMESTAMPTZ  NOT NULL,
  PRIMARY KEY (gk_id, waypoint_id),
  CONSTRAINT fk_gk_cache_visits_waypoint
    FOREIGN KEY (waypoint_id) REFERENCES stats.waypoints(id)
    DEFERRABLE INITIALLY DEFERRED
);

COMMENT ON TABLE stats.gk_cache_visits
  IS 'Per-GeoKret per-waypoint visit counter; enables cache analytics without gk_moves scans';
COMMENT ON COLUMN stats.gk_cache_visits.gk_id
  IS 'GeoKret internal ID (references geokrety.gk_geokrety.id, not FK to avoid cross-schema dep)';
COMMENT ON COLUMN stats.gk_cache_visits.waypoint_id
  IS 'FK to stats.waypoints(id); surrogated to allow rename/merge without cascading issues';
COMMENT ON COLUMN stats.gk_cache_visits.visit_count
  IS 'Count of moves referencing this waypoint for this GeoKret; incremented by trigger';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.gk_cache_visits;');
    }
}
```

## Data Contract

| Column             | Type          | Nullable | Default | Description                                   |
| ------------------ | ------------- | -------- | ------- | --------------------------------------------- |
| `gk_id`            | `INT`         | NOT NULL | —       | **PK (part 1)** — GeoKret internal integer ID |
| `waypoint_id`      | `BIGINT`      | NOT NULL | —       | **PK (part 2)** — FK to `stats.waypoints(id)` |
| `visit_count`      | `BIGINT`      | NOT NULL | `0`     | Number of moves at this waypoint for this GK  |
| `first_visited_at` | `TIMESTAMPTZ` | NOT NULL | —       | Timestamp of earliest qualifying move         |
| `last_visited_at`  | `TIMESTAMPTZ` | NOT NULL | —       | Timestamp of most recent qualifying move      |

**Constraints:**

- PK: `(gk_id, waypoint_id)`
- FK: `waypoint_id → stats.waypoints(id)` — `DEFERRABLE INITIALLY DEFERRED`

**Design note:** `gk_id` has no FK to `geokrety.gk_geokrety` intentionally — cross-schema FKs have been avoided throughout to allow independent schema migrations. Referential integrity is maintained by the trigger.

| Column             | Type          | Nullable | Default | Notes                                  |
| ------------------ | ------------- | -------- | ------- | -------------------------------------- |
| `gk_id`            | `INT`         | NOT NULL | —       | PK (1/2)                               |
| `waypoint_id`      | `BIGINT`      | NOT NULL | —       | PK (2/2) + FK to `stats.waypoints(id)` |
| `visit_count`      | `BIGINT`      | NOT NULL | `0`     |                                        |
| `first_visited_at` | `TIMESTAMPTZ` | NOT NULL | —       |                                        |
| `last_visited_at`  | `TIMESTAMPTZ` | NOT NULL | —       |                                        |

## SQL Usage Examples

```sql
-- How many distinct caches has a GeoKret visited?
SELECT COUNT(*) AS distinct_caches
FROM stats.gk_cache_visits
WHERE gk_id = 12345;

-- Which caches did a GeoKret visit, most recent first?
SELECT w.waypoint_code, cv.visit_count, cv.last_visited_at
FROM stats.gk_cache_visits cv
JOIN stats.waypoints w ON w.id = cv.waypoint_id
WHERE cv.gk_id = 12345
ORDER BY cv.last_visited_at DESC;

-- Most visited cache globally (used in v_uc10_cache_popularity)
SELECT w.waypoint_code, SUM(cv.visit_count) AS total_visits, COUNT(DISTINCT cv.gk_id) AS distinct_gks
FROM stats.gk_cache_visits cv
JOIN stats.waypoints w ON w.id = cv.waypoint_id
GROUP BY w.waypoint_code
ORDER BY total_visits DESC
LIMIT 10;

-- Caches visited multiple times by the same GK (repeat visitors)
SELECT w.waypoint_code, cv.visit_count
FROM stats.gk_cache_visits cv
JOIN stats.waypoints w ON w.id = cv.waypoint_id
WHERE cv.gk_id = 12345 AND cv.visit_count > 1
ORDER BY cv.visit_count DESC;

-- Top GeoKrety by cache diversity
SELECT gk_id, COUNT(*) AS distinct_caches_visited
FROM stats.gk_cache_visits
GROUP BY gk_id
ORDER BY distinct_caches_visited DESC
LIMIT 20;
```

## Graph / Visualization Specification

1. **Most visited caches** — Horizontal bar chart
   - Data: `SELECT w.waypoint_code, SUM(cv.visit_count) AS total FROM stats.gk_cache_visits cv JOIN stats.waypoints w ON w.id = cv.waypoint_id GROUP BY w.waypoint_code ORDER BY total DESC LIMIT 20`
   - x-axis: total visit count, y-axis: waypoint code labels

2. **Cache visit distribution** — Histogram
   - Data: `SELECT visit_count, COUNT(*) AS gk_count FROM stats.gk_cache_visits GROUP BY visit_count ORDER BY visit_count`
   - Shows how many GKs visited each cache 1 vs 2 vs 3+ times (most: 1)

```
ASCII Sample — Top Caches:
GC1A2B3  ████████████████████  2,341 GK visits
GC00001  ██████████████████    1,992 GK visits
GC5ZZZZ  ████████████          1,234 GK visits
...
```

## TimescaleDB Assessment

**NOT recommended.** `stats.gk_cache_visits` is a compressed rollup table with PK `(gk_id, waypoint_id)`. There is no time-series dimension on the PK. `first_visited_at` and `last_visited_at` are metadata columns, not partitioning keys. All writes are `INSERT ... ON CONFLICT DO UPDATE` upserts on fixed PK rows. TimescaleDB hypertable partitioning is inappropriate.

## pgTAP Unit Tests

| Test ID   | Assertion                                                                    | Expected |
| --------- | ---------------------------------------------------------------------------- | -------- |
| T-4.4.001 | `has_table('stats', 'gk_cache_visits')`                                      | pass     |
| T-4.4.002 | `col_is_pk('stats', 'gk_cache_visits', ARRAY['gk_id', 'waypoint_id'])`       | pass     |
| T-4.4.003 | `col_type_is('stats', 'gk_cache_visits', 'gk_id', 'integer')`                | pass     |
| T-4.4.004 | `col_type_is('stats', 'gk_cache_visits', 'waypoint_id', 'bigint')`           | pass     |
| T-4.4.005 | `col_type_is('stats', 'gk_cache_visits', 'visit_count', 'bigint')`           | pass     |
| T-4.4.006 | `col_default_is('stats', 'gk_cache_visits', 'visit_count', '0')`             | pass     |
| T-4.4.007 | FK `fk_gk_cache_visits_waypoint` exists and references `stats.waypoints(id)` | pass     |
| T-4.4.008 | Table is empty after creation                                                | pass     |
| T-4.4.009 | Insert with non-existent `waypoint_id` raises FK violation (after commit)    | pass     |
| T-4.4.010 | `phinx rollback` drops the table cleanly                                     | pass     |

| Test ID   | Area     | Description                           | Method         |
| --------- | -------- | ------------------------------------- | -------------- |
| T-4.4.001 | schema   | Table exists                          | pgTAP          |
| T-4.4.002 | schema   | PK `(gk_id, waypoint_id)`             | pgTAP          |
| T-4.4.003 | schema   | `gk_id` is `integer`                  | pgTAP          |
| T-4.4.004 | schema   | `waypoint_id` is `bigint`             | pgTAP          |
| T-4.4.005 | schema   | `visit_count` is `bigint`             | pgTAP          |
| T-4.4.006 | schema   | `visit_count` default 0               | pgTAP          |
| T-4.4.007 | schema   | FK to `stats.waypoints(id)` deferred  | pgTAP / SQL    |
| T-4.4.008 | data     | Empty after migration                 | SQL            |
| T-4.4.009 | data     | FK violation on invalid `waypoint_id` | SQL            |
| T-4.4.010 | rollback | `down()` drops table                  | phinx rollback |

## Implementation Checklist

- [ ] 1. Verify S4T01 (`stats.waypoints`) migration applied
- [ ] 2. Create migration `20260310400300_create_gk_cache_visits.php`
- [ ] 3. Run `phinx migrate` — confirm no errors
- [ ] 4. `\d stats.gk_cache_visits` — verify 5 columns with correct types
- [ ] 5. Verify PK on `(gk_id, waypoint_id)`
- [ ] 6. Verify FK `fk_gk_cache_visits_waypoint` is `DEFERRABLE INITIALLY DEFERRED`
- [ ] 7. Table should be empty (no seed in this step)
- [ ] 8. Run pgTAP tests T-4.4.001 through T-4.4.010
- [ ] 9. Verify `phinx rollback` drops the table cleanly

- - [DBA  ] DBA file reviewed: [task-S4T04.dba.md](task-S4T04.md#dba)
- [ ] S4T01 applied (waypoints table exists)
- [ ] Migration `20260310400300_create_gk_cache_visits.php` created
- [ ] Migration runs successfully
- [ ] PK and FK verified
- [ ] All T-4.4.xxx tests pass
- [ ] Rollback verified
