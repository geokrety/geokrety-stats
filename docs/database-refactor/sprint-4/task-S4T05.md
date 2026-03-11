---
title: "Task S4T05: Create stats.user_cache_visits Table"
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
  - user-cache-visits
depends_on:
  - S4T01
task: S4T05
step: 4.5
migration: 20260310400400_create_user_cache_visits.php
blocks:
  - S4T08
  - S4T10
  - S4T11
changelog:
  - 2026.03.10: created by merge of task-S4T05.dba.md and task-S4T05.specification.md
---

# Task S4T05: Create stats.user_cache_visits Table

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T05.dba.md`
- Specification source: `task-S4T05.specification.md`

## Purpose & Scope

Creates `stats.user_cache_visits`, which records how many times each user's moves have been associated with each waypoint/cache. This enables per-user cache interaction analytics:

- "How many distinct caches has this user visited?"
- "What are the top caches visited by users?"
- "Has a user visited a specific cache recently?"

The structure is symmetric to `stats.gk_cache_visits` but keyed by `user_id` rather than `gk_id`.

Create `stats.user_cache_visits` to pre-aggregate per-user per-waypoint visit counts. Symmetric to `stats.gk_cache_visits` but from the user perspective.

**Scope:** One table with FK to `stats.waypoints`. No triggers (S4T08). No seed data (S4T11).

## Requirements

| ID      | Requirement                                                                         |
| ------- | ----------------------------------------------------------------------------------- |
| REQ-440 | Table must exist with 5 columns as specified                                        |
| REQ-441 | PK must be `(user_id, waypoint_id)`                                                 |
| REQ-442 | `waypoint_id` must FK to `stats.waypoints(id)` with `DEFERRABLE INITIALLY DEFERRED` |
| REQ-443 | `visit_count` default 0                                                             |
| REQ-444 | No FK from `user_id` to `geokrety.gk_users` (cross-schema FK avoided)               |
| REQ-445 | Table empty after creation                                                          |
| REQ-446 | `down()` drops with `IF EXISTS`                                                     |

## Acceptance Criteria

| Criterion        | Verification                                |
| ---------------- | ------------------------------------------- |
| Table exists     | `\dt stats.user_cache_visits`               |
| PK composite     | `(user_id, waypoint_id)` constraint         |
| FK deferred      | `is_deferrable=YES, initially_deferred=YES` |
| Empty after init | `SELECT COUNT(*) = 0`                       |

## Migration File

**`20260310400400_create_user_cache_visits.php`**

## Full SQL DDL

```sql
CREATE TABLE stats.user_cache_visits (
  user_id           INT          NOT NULL,
  waypoint_id       BIGINT       NOT NULL,
  visit_count       BIGINT       NOT NULL DEFAULT 0,
  first_visited_at  TIMESTAMPTZ  NOT NULL,
  last_visited_at   TIMESTAMPTZ  NOT NULL,
  PRIMARY KEY (user_id, waypoint_id),
  CONSTRAINT fk_user_cache_visits_waypoint
    FOREIGN KEY (waypoint_id) REFERENCES stats.waypoints(id)
    DEFERRABLE INITIALLY DEFERRED
);

COMMENT ON TABLE stats.user_cache_visits
  IS 'Per-user per-waypoint visit counter; enables user cache analytics without gk_moves scans';
COMMENT ON COLUMN stats.user_cache_visits.user_id
  IS 'User internal ID (references geokrety.gk_users.id; no cross-schema FK)';
COMMENT ON COLUMN stats.user_cache_visits.visit_count
  IS 'Number of moves by this user at this waypoint; incremented by trigger';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateUserCacheVisits extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.user_cache_visits (
  user_id           INT          NOT NULL,
  waypoint_id       BIGINT       NOT NULL,
  visit_count       BIGINT       NOT NULL DEFAULT 0,
  first_visited_at  TIMESTAMPTZ  NOT NULL,
  last_visited_at   TIMESTAMPTZ  NOT NULL,
  PRIMARY KEY (user_id, waypoint_id),
  CONSTRAINT fk_user_cache_visits_waypoint
    FOREIGN KEY (waypoint_id) REFERENCES stats.waypoints(id)
    DEFERRABLE INITIALLY DEFERRED
);

COMMENT ON TABLE stats.user_cache_visits
  IS 'Per-user per-waypoint visit counter; enables user cache analytics without gk_moves scans';
COMMENT ON COLUMN stats.user_cache_visits.user_id
  IS 'User internal ID (references geokrety.gk_users.id; no cross-schema FK)';
COMMENT ON COLUMN stats.user_cache_visits.visit_count
  IS 'Number of moves by this user at this waypoint; incremented by trigger';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.user_cache_visits;');
    }
}
```

## Data Contract

| Column             | Type          | Nullable | Default | Description                                     |
| ------------------ | ------------- | -------- | ------- | ----------------------------------------------- |
| `user_id`          | `INT`         | NOT NULL | —       | **PK (part 1)** — User internal integer ID      |
| `waypoint_id`      | `BIGINT`      | NOT NULL | —       | **PK (part 2)** — FK to `stats.waypoints(id)`   |
| `visit_count`      | `BIGINT`      | NOT NULL | `0`     | Number of moves by user at this waypoint        |
| `first_visited_at` | `TIMESTAMPTZ` | NOT NULL | —       | Timestamp of user's first qualifying move here  |
| `last_visited_at`  | `TIMESTAMPTZ` | NOT NULL | —       | Timestamp of user's most recent qualifying move |

**Design note:** Anonymous moves (`author IS NULL` on `gk_moves`) do NOT create rows in this table. Only moves with a non-NULL author contribute. This is enforced by the trigger (S4T08).

| Column             | Type          | Nullable | Default | Notes                              |
| ------------------ | ------------- | -------- | ------- | ---------------------------------- |
| `user_id`          | `INT`         | NOT NULL | —       | PK (1/2)                           |
| `waypoint_id`      | `BIGINT`      | NOT NULL | —       | PK (2/2) + FK to `stats.waypoints` |
| `visit_count`      | `BIGINT`      | NOT NULL | `0`     |                                    |
| `first_visited_at` | `TIMESTAMPTZ` | NOT NULL | —       |                                    |
| `last_visited_at`  | `TIMESTAMPTZ` | NOT NULL | —       |                                    |

## SQL Usage Examples

```sql
-- How many distinct caches has a user visited?
SELECT COUNT(*) AS distinct_caches
FROM stats.user_cache_visits
WHERE user_id = 9876;

-- Top caches by total user visits (globally)
SELECT w.waypoint_code, SUM(uv.visit_count) AS total_visits, COUNT(DISTINCT uv.user_id) AS distinct_users
FROM stats.user_cache_visits uv
JOIN stats.waypoints w ON w.id = uv.waypoint_id
GROUP BY w.waypoint_code
ORDER BY total_visits DESC
LIMIT 10;

-- Has a specific user visited a specific cache?
SELECT visit_count, first_visited_at, last_visited_at
FROM stats.user_cache_visits uv
JOIN stats.waypoints w ON w.id = uv.waypoint_id
WHERE uv.user_id = 9876 AND w.waypoint_code = 'GC1A2B3';

-- Most active users by number of distinct caches
SELECT user_id, COUNT(*) AS distinct_caches
FROM stats.user_cache_visits
GROUP BY user_id
ORDER BY distinct_caches DESC
LIMIT 20;

-- Users who visited the most caches in a given country
SELECT uv.user_id, COUNT(DISTINCT uv.waypoint_id) AS caches_in_country
FROM stats.user_cache_visits uv
JOIN stats.waypoints w ON w.id = uv.waypoint_id
WHERE w.country = 'PL'
GROUP BY uv.user_id
ORDER BY caches_in_country DESC
LIMIT 10;
```

## Graph / Visualization Specification

1. **Cache visit distribution by user** — Histogram
   - Data: `SELECT COUNT(*) AS distinct_caches, COUNT(user_id) AS users FROM stats.user_cache_visits GROUP BY user_id ORDER BY distinct_caches`
   - Shows how many users visited 1, 2-5, 5-10, 10-50, 50+ caches

2. **Top caches by unique users** — Bar chart
   - Data: `SELECT w.waypoint_code, COUNT(DISTINCT uv.user_id) FROM stats.user_cache_visits uv JOIN stats.waypoints w ... GROUP BY ... ORDER BY ... LIMIT 20`
   - x-axis: cache code, y-axis: unique user count

```
ASCII Sample — Cache visit distribution:
Visited 1 cache:    45,231 users (30%)
Visited 2-5:        38,192 users (26%)
Visited 6-20:       31,440 users (21%)
Visited 21-100:     25,000 users (17%)
Visited 100+:        9,000 users  (6%)
```

## TimescaleDB Assessment

**NOT recommended.** Same reasoning as `stats.gk_cache_visits` — this is a compressed rollup keyed by `(user_id, waypoint_id)`. No time-series dimension on the PK. Standard PostgreSQL is optimal.

## pgTAP Unit Tests

| Test ID   | Assertion                                                                  | Expected |
| --------- | -------------------------------------------------------------------------- | -------- |
| T-4.5.001 | `has_table('stats', 'user_cache_visits')`                                  | pass     |
| T-4.5.002 | `col_is_pk('stats', 'user_cache_visits', ARRAY['user_id', 'waypoint_id'])` | pass     |
| T-4.5.003 | `col_type_is('stats', 'user_cache_visits', 'user_id', 'integer')`          | pass     |
| T-4.5.004 | `col_type_is('stats', 'user_cache_visits', 'waypoint_id', 'bigint')`       | pass     |
| T-4.5.005 | `col_type_is('stats', 'user_cache_visits', 'visit_count', 'bigint')`       | pass     |
| T-4.5.006 | `col_default_is('stats', 'user_cache_visits', 'visit_count', '0')`         | pass     |
| T-4.5.007 | FK `fk_user_cache_visits_waypoint` exists and is deferrable                | pass     |
| T-4.5.008 | Table is empty after creation                                              | pass     |
| T-4.5.009 | FK violation on invalid `waypoint_id`                                      | pass     |
| T-4.5.010 | `phinx rollback` drops the table                                           | pass     |

| Test ID   | Area     | Description                           | Method    |
| --------- | -------- | ------------------------------------- | --------- |
| T-4.5.001 | schema   | Table exists                          | pgTAP     |
| T-4.5.002 | schema   | PK `(user_id, waypoint_id)`           | pgTAP     |
| T-4.5.003 | schema   | `user_id` is `integer`                | pgTAP     |
| T-4.5.004 | schema   | `waypoint_id` is `bigint`             | pgTAP     |
| T-4.5.005 | schema   | `visit_count` is `bigint`             | pgTAP     |
| T-4.5.006 | schema   | `visit_count` default 0               | pgTAP     |
| T-4.5.007 | schema   | FK to `stats.waypoints(id)` deferred  | pgTAP/SQL |
| T-4.5.008 | data     | Empty after migration                 | SQL       |
| T-4.5.009 | data     | FK violation on invalid `waypoint_id` | SQL       |
| T-4.5.010 | rollback | `down()` drops table                  | phinx     |

## Implementation Checklist

- [ ] 1. Verify S4T01 (`stats.waypoints`) migration applied
- [ ] 2. Create migration `20260310400400_create_user_cache_visits.php`
- [ ] 3. Run `phinx migrate` — confirm no errors
- [ ] 4. `\d stats.user_cache_visits` — 5 columns with correct types
- [ ] 5. Verify PK on `(user_id, waypoint_id)`
- [ ] 6. Verify FK is `DEFERRABLE INITIALLY DEFERRED`
- [ ] 7. Table should be empty
- [ ] 8. Run pgTAP tests T-4.5.001 through T-4.5.010
- [ ] 9. Verify `phinx rollback` drops table cleanly

- - [DBA  ] DBA file reviewed: [task-S4T05.dba.md](task-S4T05.md#dba)
- [ ] S4T01 applied
- [ ] Migration `20260310400400_create_user_cache_visits.php` created
- [ ] Migration runs successfully
- [ ] All T-4.5.xxx tests pass
- [ ] Rollback verified
