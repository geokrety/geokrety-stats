---
title: "Task S4T06: Create stats.gk_related_users Table"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 4
tags:
  - database
  - database-refactor
  - dba
  - gk-related-users
  - social
  - specification
  - sprint-4
  - sql
  - stats
  - table
  - task-index
  - task-merge
depends_on:
  - "Sprint 2"
task: S4T06
step: 4.6
migration: 20260310400500_create_gk_related_users.php
blocks:
  - S4T09
  - S4T10
  - S4T11
changelog:
  - 2026-03-10: created by merge of task-S4T06.dba.md and task-S4T06.specification.md
---

# Task S4T06: Create stats.gk_related_users Table

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T06.dba.md`
- Specification source: `task-S4T06.specification.md`

## Purpose & Scope

Creates `stats.gk_related_users`, which tracks which users have interacted with each GeoKret and how many times. A row `(geokrety_id, user_id)` means that user has logged at least one non-comment move on that GK. This directly powers:

- "How many distinct users has this GeoKret touched?" (reach metric)
- UC3: GK circulation graph (number of interactions per GK)
- Owner `+1 handover` bonus detection (via reach of 10 distinct users in 6 months)

Create `stats.gk_related_users` to track user-GeoKret interactions. Powers UC3 (GK circulation graph) and the 10-distinct-users reach-bonus detection logic.

**Scope:** One table. No FK to `geokrety.gk_geokrety` (cross-schema avoided). No triggers (S4T09). No seed data (S4T11).

## Requirements

| ID      | Requirement                                                                          |
| ------- | ------------------------------------------------------------------------------------ |
| REQ-450 | Table must exist with 5 columns as specified                                         |
| REQ-451 | PK must be `(geokrety_id, user_id)`                                                  |
| REQ-452 | `interaction_count` default 0                                                        |
| REQ-453 | `first_interaction` and `last_interaction` must be `TIMESTAMPTZ`                     |
| REQ-454 | Trigger (S4T09) will only insert rows for authenticated users (`author IS NOT NULL`) |
| REQ-455 | Trigger (S4T09) will only count qualifying moves: `move_type IN (0,1,3,5)`           |
| REQ-456 | Table empty after creation                                                           |

## Acceptance Criteria

| Criterion              | Verification                                   |
| ---------------------- | ---------------------------------------------- |
| Table exists           | `\dt stats.gk_related_users`                   |
| PK composite           | `(geokrety_id, user_id)` confirmed             |
| Timestamps TIMESTAMPTZ | `pg_typeof(first_interaction)` = `timestamptz` |
| Empty after init       | `SELECT COUNT(*) = 0`                          |

## Migration File

**`20260310400500_create_gk_related_users.php`**

## Full SQL DDL

```sql
CREATE TABLE stats.gk_related_users (
  geokrety_id         INT          NOT NULL,
  user_id             INT          NOT NULL,
  interaction_count   BIGINT       NOT NULL DEFAULT 0,
  first_interaction   TIMESTAMPTZ  NOT NULL,
  last_interaction    TIMESTAMPTZ  NOT NULL,
  PRIMARY KEY (geokrety_id, user_id)
);

COMMENT ON TABLE stats.gk_related_users
  IS 'Per-GeoKret per-user interaction counter; powers UC3, reach bonus, and social graph';
COMMENT ON COLUMN stats.gk_related_users.geokrety_id
  IS 'GeoKret internal ID (no cross-schema FK)';
COMMENT ON COLUMN stats.gk_related_users.user_id
  IS 'User internal ID — only authenticated users (author IS NOT NULL)';
COMMENT ON COLUMN stats.gk_related_users.interaction_count
  IS 'Count of qualifying moves (DROP/GRAB/SEEN/DIP, not COMMENT) by this user on this GK';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkRelatedUsers extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.gk_related_users (
  geokrety_id         INT          NOT NULL,
  user_id             INT          NOT NULL,
  interaction_count   BIGINT       NOT NULL DEFAULT 0,
  first_interaction   TIMESTAMPTZ  NOT NULL,
  last_interaction    TIMESTAMPTZ  NOT NULL,
  PRIMARY KEY (geokrety_id, user_id)
);

COMMENT ON TABLE stats.gk_related_users
  IS 'Per-GeoKret per-user interaction counter; powers UC3, reach bonus, and social graph';
COMMENT ON COLUMN stats.gk_related_users.geokrety_id
  IS 'GeoKret internal ID (no cross-schema FK)';
COMMENT ON COLUMN stats.gk_related_users.user_id
  IS 'User internal ID — only authenticated users (author IS NOT NULL)';
COMMENT ON COLUMN stats.gk_related_users.interaction_count
  IS 'Count of qualifying moves (DROP/GRAB/SEEN/DIP, not COMMENT) by this user on this GK';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.gk_related_users;');
    }
}
```

## Data Contract

| Column              | Type          | Nullable | Default | Description                                          |
| ------------------- | ------------- | -------- | ------- | ---------------------------------------------------- |
| `geokrety_id`       | `INT`         | NOT NULL | —       | **PK (part 1)** — GeoKret internal ID                |
| `user_id`           | `INT`         | NOT NULL | —       | **PK (part 2)** — User internal ID (authenticated)   |
| `interaction_count` | `BIGINT`      | NOT NULL | `0`     | Non-comment moves by user on this GK                 |
| `first_interaction` | `TIMESTAMPTZ` | NOT NULL | —       | Timestamp of user's first qualifying move on this GK |
| `last_interaction`  | `TIMESTAMPTZ` | NOT NULL | —       | Timestamp of user's most recent qualifying move      |

**Qualifying moves:** `move_type IN (0,1,3,5)` — DROP, GRAB, SEEN, DIP. Comments (type 2) and Archives (type 4) do NOT qualify.

| Column              | Type          | Nullable | Default | Notes                              |
| ------------------- | ------------- | -------- | ------- | ---------------------------------- |
| `geokrety_id`       | `INT`         | NOT NULL | —       | PK (1/2)                           |
| `user_id`           | `INT`         | NOT NULL | —       | PK (2/2) — authenticated user only |
| `interaction_count` | `BIGINT`      | NOT NULL | `0`     | Qualifying moves (type 0,1,3,5)    |
| `first_interaction` | `TIMESTAMPTZ` | NOT NULL | —       |                                    |
| `last_interaction`  | `TIMESTAMPTZ` | NOT NULL | —       |                                    |

## SQL Usage Examples

```sql
-- How many distinct users has a GeoKret touched?
SELECT COUNT(*) AS distinct_users
FROM stats.gk_related_users
WHERE geokrety_id = 12345;

-- Total interactions on a GeoKret
SELECT SUM(interaction_count) AS total_interactions
FROM stats.gk_related_users
WHERE geokrety_id = 12345;

-- Has a specific user interacted with a specific GK?
SELECT interaction_count, first_interaction, last_interaction
FROM stats.gk_related_users
WHERE geokrety_id = 12345 AND user_id = 9876;

-- GeoKrety that have reached 10+ distinct users (for reach bonus check)
SELECT geokrety_id, COUNT(*) AS distinct_users
FROM stats.gk_related_users
GROUP BY geokrety_id
HAVING COUNT(*) >= 10
ORDER BY distinct_users DESC;

-- Most circulated GeoKrety (UC3 base query)
SELECT geokrety_id,
       COUNT(*) AS distinct_users,
       SUM(interaction_count) AS total_interactions
FROM stats.gk_related_users
GROUP BY geokrety_id
ORDER BY distinct_users DESC
LIMIT 20;

-- Users that have interacted with the most GeoKrety
SELECT user_id, COUNT(*) AS distinct_gks_touched
FROM stats.gk_related_users
GROUP BY user_id
ORDER BY distinct_gks_touched DESC
LIMIT 20;
```

## Graph / Visualization Specification

1. **UC3: GK circulation chart** — Horizontal bar
   - Data: `SELECT geokrety_id, COUNT(*) AS users FROM stats.gk_related_users GROUP BY geokrety_id ORDER BY users DESC LIMIT 20`
   - x-axis: distinct users, y-axis: GK identifier

2. **Reach milestone distribution** — Histogram
   - Data: `SELECT reached_users, COUNT(*) AS gk_count FROM (SELECT geokrety_id, COUNT(*) AS reached_users FROM stats.gk_related_users GROUP BY geokrety_id) sub GROUP BY reached_users ORDER BY reached_users`

```
ASCII Sample — GK reach distribution:
1 user   ██████████████████████████████ 60%   (very new GKs)
2-5      ████████████████               30%
6-9      ████                            7%
10+      █                               3%

Top 3 most circulated GKs:
GK-0001  ██████████████████████  412 distinct users
GK-ABCD  █████████████████████   387 distinct users
GK-F001  ████████████████████    359 distinct users
```

## TimescaleDB Assessment

**NOT recommended.** `stats.gk_related_users` is a compressed rollup with PK `(geokrety_id, user_id)`. No time dimension on the PK. Standard PostgreSQL with B-tree index handles all workloads.

## pgTAP Unit Tests

| Test ID   | Assertion                                                                                   | Expected |
| --------- | ------------------------------------------------------------------------------------------- | -------- |
| T-4.6.001 | `has_table('stats', 'gk_related_users')`                                                    | pass     |
| T-4.6.002 | `col_is_pk('stats', 'gk_related_users', ARRAY['geokrety_id', 'user_id'])`                   | pass     |
| T-4.6.003 | `col_type_is('stats', 'gk_related_users', 'geokrety_id', 'integer')`                        | pass     |
| T-4.6.004 | `col_type_is('stats', 'gk_related_users', 'user_id', 'integer')`                            | pass     |
| T-4.6.005 | `col_type_is('stats', 'gk_related_users', 'interaction_count', 'bigint')`                   | pass     |
| T-4.6.006 | `col_default_is('stats', 'gk_related_users', 'interaction_count', '0')`                     | pass     |
| T-4.6.007 | `col_type_is('stats', 'gk_related_users', 'first_interaction', 'timestamp with time zone')` | pass     |
| T-4.6.008 | Table is empty after creation                                                               | pass     |
| T-4.6.009 | `phinx rollback` drops table cleanly                                                        | pass     |

| Test ID   | Area     | Description                          | Method |
| --------- | -------- | ------------------------------------ | ------ |
| T-4.6.001 | schema   | Table exists                         | pgTAP  |
| T-4.6.002 | schema   | PK `(geokrety_id, user_id)`          | pgTAP  |
| T-4.6.003 | schema   | `geokrety_id` is `integer`           | pgTAP  |
| T-4.6.004 | schema   | `user_id` is `integer`               | pgTAP  |
| T-4.6.005 | schema   | `interaction_count` is `bigint`      | pgTAP  |
| T-4.6.006 | schema   | `interaction_count` default 0        | pgTAP  |
| T-4.6.007 | schema   | `first_interaction` is `timestamptz` | pgTAP  |
| T-4.6.008 | data     | Empty after migration                | SQL    |
| T-4.6.009 | rollback | `down()` drops table                 | phinx  |

## Implementation Checklist

- [ ] 1. Create migration `20260310400500_create_gk_related_users.php`
- [ ] 2. Run `phinx migrate` — confirm no errors
- [ ] 3. `\d stats.gk_related_users` — 5 columns with correct types
- [ ] 4. Verify PK on `(geokrety_id, user_id)`
- [ ] 5. Table empty after creation
- [ ] 6. Run pgTAP tests T-4.6.001 through T-4.6.009
- [ ] 7. Verify `phinx rollback` drops table cleanly

- [ ] DBA file reviewed: [task-S4T06.dba.md](task-S4T06.dba.md)
- [ ] Migration `20260310400500_create_gk_related_users.php` created
- [ ] Migration runs successfully
- [ ] All T-4.6.xxx tests pass
- [ ] Rollback verified
