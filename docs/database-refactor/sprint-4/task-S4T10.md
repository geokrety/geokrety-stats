---
title: "Task S4T10: Indexes for Sprint 4 Tables"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 4
tags:
  - database
  - database-refactor
  - dba
  - index
  - performance
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
task: S4T10
step: 4.10
migration: 20260310400900_create_cache_relation_indexes.php
blocks:
  - S4T11
changelog:
  - 2026-03-10: created by merge of task-S4T10.dba.md and task-S4T10.specification.md
  - 2026-03-10: logged the canonical-four-index versus legacy-ten-index conflict
---

# Task S4T10: Indexes for Sprint 4 Tables

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T10.dba.md`
- Specification source: `task-S4T10.specification.md`

## Resolved Decision

- Sprint 4 creates exactly four canonical secondary indexes in this task.
- The earlier ten-index draft is obsolete because it depended on non-canonical columns and unapproved performance assumptions.
- Additional indexes remain out of scope until approved as a master-spec amendment.

## Purpose & Scope

Creates secondary indexes on the tables introduced in S4T01–S4T07 to support efficient:

- Waypoint lookup by country (UC1, UC10)
- Cache visit queries by waypoint (UC10: most-visited caches)
- Social relation queries by user (UC2, UC3)
- Social relation queries by GeoKret (UC3)

Tables already have appropriate PK indexes. Only non-PK access patterns requiring additional indexes are created here.

Adds the canonical four secondary indexes to the Sprint 4 tables. These indexes are performance optimizations, not correctness requirements. They support the approved UC2, UC3, and UC10 query shapes without reintroducing the deprecated ten-index draft.

---

## Requirements

| ID      | Description                                                                                          | MoSCoW |
| ------- | ---------------------------------------------------------------------------------------------------- | ------ |
| REQ-490 | 4 canonical indexes created across Sprint 4 `stats.*` tables                                          | MUST   |
| REQ-491 | `idx_waypoints_country` — on `stats.waypoints(country)`                                                | MUST   |
| REQ-492 | `idx_gk_cache_visits_waypoint` — on `(waypoint_id, gk_id)` for cache-centric queries                  | MUST   |
| REQ-493 | `idx_user_cache_visits_waypoint` — on `(waypoint_id, user_id)` for cache-centric user lookup          | MUST   |
| REQ-494 | `idx_gk_related_users_user` — on `(user_id)` for user-first relation lookup                            | MUST   |
| REQ-495 | Any additional non-canonical performance indexes are out of scope until the master spec is amended     | MUST   |
| REQ-496 | `phinx rollback` drops all canonical Sprint 4 indexes cleanly                                          | MUST   |

---

## Acceptance Criteria

| #   | Criterion                                                          | How to Verify                                     |
| --- | ------------------------------------------------------------------ | ------------------------------------------------- |
| 1   | All 4 canonical indexes exist in `stats` schema                  | `\di+ stats.*` shows the canonical set           |
| 2   | No duplicate index is created for `waypoint_code` unique lookup  | Confirm the unique constraint remains the only key |
| 3   | `phinx rollback` drops all canonical indexes cleanly             | 0 of the canonical indexes remain after rollback  |

---

## Migration File

**`20260310400900_create_cache_relation_indexes.php`**

## Full SQL DDL

```sql
CREATE INDEX idx_waypoints_country
  ON stats.waypoints (country)
  WHERE country IS NOT NULL;

CREATE INDEX idx_gk_cache_visits_waypoint
  ON stats.gk_cache_visits (waypoint_id, gk_id);

CREATE INDEX idx_user_cache_visits_waypoint
  ON stats.user_cache_visits (waypoint_id, user_id);
CREATE INDEX idx_gk_related_users_user
  ON stats.gk_related_users (user_id);
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateCacheRelationIndexes extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE INDEX idx_waypoints_country
  ON stats.waypoints (country) WHERE country IS NOT NULL;

CREATE INDEX idx_gk_cache_visits_waypoint
  ON stats.gk_cache_visits (waypoint_id, gk_id);

CREATE INDEX idx_user_cache_visits_waypoint
  ON stats.user_cache_visits (waypoint_id, user_id);

CREATE INDEX idx_gk_related_users_user
  ON stats.gk_related_users (user_id);
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP INDEX IF EXISTS stats.idx_waypoints_country;
DROP INDEX IF EXISTS stats.idx_gk_cache_visits_waypoint;
DROP INDEX IF EXISTS stats.idx_user_cache_visits_waypoint;
DROP INDEX IF EXISTS stats.idx_gk_related_users_user;
SQL
        );
    }
}
```

## pgTAP Unit Tests

| Test ID    | Assertion                                                                     | Expected |
| ---------- | ----------------------------------------------------------------------------- | -------- |
| T-4.10.001 | `has_index('stats', 'waypoints', 'idx_waypoints_country')`                    | pass     |
| T-4.10.002 | `has_index('stats', 'gk_cache_visits', 'idx_gk_cache_visits_waypoint')`       | pass     |
| T-4.10.003 | `has_index('stats', 'user_cache_visits', 'idx_user_cache_visits_waypoint')`   | pass     |
| T-4.10.004 | `has_index('stats', 'gk_related_users', 'idx_gk_related_users_user')`         | pass     |
| T-4.10.005 | `phinx rollback` drops the canonical four indexes                             | pass     |

| Test ID    | Assertion                                 | Pass Condition            |
| ---------- | ----------------------------------------- | ------------------------- |
| T-4.10.001 | `idx_waypoints_country` exists            | `has_index()` pass        |
| T-4.10.002 | `idx_gk_cache_visits_waypoint` exists     | `has_index()` pass        |
| T-4.10.003 | `idx_gk_related_users_user` exists        | `has_index()` pass        |
| T-4.10.004 | Rollback drops the canonical indexes      | All absent after rollback |

---

## Implementation Checklist

- [ ] 1. Create migration `20260310400900_create_cache_relation_indexes.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. `\di+ stats.*` — confirm exactly the canonical 4 indexes listed
- [ ] 4. Verify partial index on `idx_waypoints_country` (only non-NULL countries)
- [ ] 5. Run pgTAP T-4.10.001 through T-4.10.005 — all pass
- [ ] 6. `phinx rollback` — canonical 4 indexes dropped

- [ ] 1. Write `20260310400900_create_cache_relation_indexes.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. `\di+ stats.*` — exactly the canonical 4 indexes listed
- [ ] 4. Verify no duplicate index is created for `waypoint_code` unique lookup
- [ ] 5. Run pgTAP T-4.10.001 through T-4.10.005 — all pass
- [ ] 6. `phinx rollback` — canonical 4 indexes removed cleanly

## Index Reference Table

| Index Name                         | Table                      | Columns                                 | Type  | Partial?      | Use Case                      |
| ---------------------------------- | -------------------------- | --------------------------------------- | ----- | ------------- | ----------------------------- |
| `idx_waypoints_country`            | `stats.waypoints`          | `(country)`                             | BTREE | `IS NOT NULL` | UC1 heatmap, geo filter       |
| `idx_gk_cache_visits_waypoint`     | `stats.gk_cache_visits`    | `(waypoint_id, gk_id)`                  | BTREE | No            | UC10: GKs at cache            |
| `idx_user_cache_visits_waypoint`   | `stats.user_cache_visits`  | `(waypoint_id, user_id)`                | BTREE | No            | Users at cache                |
| `idx_gk_related_users_user`        | `stats.gk_related_users`   | `(user_id)`                             | BTREE | No            | UC2: GKs touched by user Y    |

## Notes on Missing Indexes

- `stats.waypoints (waypoint_code)` — already covered by the UNIQUE constraint (implicit index)
- `stats.gk_cache_visits (gk_id, waypoint_id)` — PK covers this already
- `stats.user_cache_visits (user_id, waypoint_id)` — PK covers this already
- `stats.gk_related_users (geokrety_id, user_id)` — PK covers this already
- `stats.user_related_users (user_id, related_user_id)` — PK covers this already

## Master-Spec Alignment

This task is governed by [../00-SPEC-DRAFT-v1.md](../00-SPEC-DRAFT-v1.md), Section 9.2.

- The canonical Sprint 4 secondary-index set is exactly: `idx_gk_related_users_user`, `idx_user_cache_visits_waypoint`, `idx_gk_cache_visits_waypoint`, and `idx_waypoints_country`.
- References to non-canonical or nonexistent columns such as `source_type`, `touch_count`, or alternate relation-key names are obsolete draft text.
- If additional performance indexes are desired beyond the canonical four, they must first be approved as a master-spec amendment rather than asserted here as settled requirements.

## Agent Loop Log

- 2026-03-10T18:40:00Z — `dba`: removed the unapproved ten-index draft and kept only the four indexes referenced by the Sprint 4 contract.
- 2026-03-10T18:40:00Z — `critical-thinking`: retained only indexes tied to explicit UC2, UC3, and UC10 access paths; broader performance tuning stays out of scope until approved.
- 2026-03-10T18:40:00Z — `specification`: aligned SQL, rollback, tests, checklist, and reference tables with Q-029 and the canonical Sprint 4 column names.

## Resolution

Q-029 is resolved by reducing this task to the canonical four-index set only.

## Objects Created (Secondary Indexes)

| Index Name                         | Table                      | Notes                                |
| ---------------------------------- | -------------------------- | ------------------------------------ |
| `idx_waypoints_country`          | `stats.waypoints`         | Country filter support            |
| `idx_gk_cache_visits_waypoint`   | `stats.gk_cache_visits`   | Cache-centric visit lookup        |
| `idx_user_cache_visits_waypoint` | `stats.user_cache_visits` | Cache-centric user lookup         |
| `idx_gk_related_users_user`      | `stats.gk_related_users`  | User-first relation lookup        |

---
