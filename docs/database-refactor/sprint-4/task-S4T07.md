---
title: "Task S4T07: stats.user_related_users Table"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 4
tags:
  - database
  - database-refactor
  - dba
  - social
  - specification
  - sprint-4
  - sql
  - stats
  - table
  - task-index
  - task-merge
  - uc2
  - user-related-users
depends_on:
  - "Sprint 2"
task: S4T07
step: 4.7
migration: 20260310400600_create_user_related_users.php
blocks:
  - S4T09
  - S4T10
  - S4T11
changelog:
  - 2026.03.10: created by merge of task-S4T07.dba.md and task-S4T07.specification.md
---

# Task S4T07: stats.user_related_users Table

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T07.dba.md`
- Specification source: `task-S4T07.specification.md`

## Purpose & Scope

Creates `stats.user_related_users`, which records social connections between pairs of users as inferred from shared GeoKrety interactions. A row `(user_id, related_user_id)` means these two users have both interacted with the same GeoKret. `shared_geokrety_count` counts how many distinct GeoKrety they share. This directly powers:

- UC2: User social-network graph (most-shared-GK pairs)
- "Most connected user" leaderboard
- Social proximity analytics

The relationship is **directional** in storage but semantically symmetric: if `(A, B)` exists, `(B, A)` also exists. The trigger maintains both directions.

Provides a persistent social-graph adjacency table between authenticated users. Pair `(user_id, related_user_id)` exists when both users have interacted with the same GeoKret. Symmetric: both `(A,B)` and `(B,A)` are stored. This table directly serves UC2 (User Network Graph) and any social proximity queries.

**Scope:** DDL only. Trigger population is in S4T09.

---

## Requirements

| ID      | Description                                                                                                | MoSCoW |
| ------- | ---------------------------------------------------------------------------------------------------------- | ------ |
| REQ-460 | Table `stats.user_related_users` exists in the `stats` schema                                              | MUST   |
| REQ-461 | `PRIMARY KEY (user_id, related_user_id)` — composite, no duplicate pairs                                   | MUST   |
| REQ-462 | `CHECK (user_id <> related_user_id)` — self-links forbidden                                                | MUST   |
| REQ-463 | `shared_geokrety_count BIGINT DEFAULT 0` — incremented by trigger S4T09                                    | MUST   |
| REQ-464 | `first_seen_at / last_seen_at TIMESTAMPTZ NOT NULL` — temporal metadata required                           | MUST   |
| REQ-465 | Table is empty after DDL creation (trigger in S4T09 populates it from existing moves in Sprint 6 backfill) | MUST   |
| REQ-466 | `phinx rollback` drops the table cleanly                                                                   | MUST   |

---

## Acceptance Criteria

| #   | Criterion                                                         | How to Verify                                                                 |
| --- | ----------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| 1   | Table created under `stats` schema                                | `\d stats.user_related_users`                                                 |
| 2   | PK is composite `(user_id, related_user_id)`                      | `\d+ stats.user_related_users` → "PRIMARY KEY"                                |
| 3   | Self-link INSERT raises constraint violation                      | `INSERT INTO stats.user_related_users VALUES (1,1,0,now(),now())` → exception |
| 4   | Both `first_seen_at` and `last_seen_at` are NOT NULL (no default) | Attempt insert without them → exception                                       |
| 5   | `shared_geokrety_count` defaults to 0                             | INSERT without column → `SELECT shared_geokrety_count` = 0                    |
| 6   | `phinx rollback` drops table                                      | Table absent after rollback                                                   |

---

## Migration File

**`20260310400600_create_user_related_users.php`**

## Full SQL DDL

```sql
CREATE TABLE stats.user_related_users (
  user_id                INT          NOT NULL,
  related_user_id        INT          NOT NULL,
  shared_geokrety_count  BIGINT       NOT NULL DEFAULT 0,
  first_seen_at          TIMESTAMPTZ  NOT NULL,
  last_seen_at           TIMESTAMPTZ  NOT NULL,
  PRIMARY KEY (user_id, related_user_id),
  CONSTRAINT chk_user_related_users_no_self
    CHECK (user_id <> related_user_id)
);

COMMENT ON TABLE stats.user_related_users
  IS 'Directional user-user relation via shared GeoKrety; both directions stored; powers UC2 social graph';
COMMENT ON COLUMN stats.user_related_users.shared_geokrety_count
  IS 'Number of distinct GeoKrety that both users have interacted with';
COMMENT ON COLUMN stats.user_related_users.user_id
  IS 'Source user (authenticated only)';
COMMENT ON COLUMN stats.user_related_users.related_user_id
  IS 'Target user (authenticated only); never equals user_id (self-relation prevented by CHECK)';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateUserRelatedUsers extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.user_related_users (
  user_id                INT          NOT NULL,
  related_user_id        INT          NOT NULL,
  shared_geokrety_count  BIGINT       NOT NULL DEFAULT 0,
  first_seen_at          TIMESTAMPTZ  NOT NULL,
  last_seen_at           TIMESTAMPTZ  NOT NULL,
  PRIMARY KEY (user_id, related_user_id),
  CONSTRAINT chk_user_related_users_no_self
    CHECK (user_id <> related_user_id)
);

COMMENT ON TABLE stats.user_related_users
  IS 'Directional user-user relation via shared GeoKrety; both directions stored; powers UC2 social graph';
COMMENT ON COLUMN stats.user_related_users.shared_geokrety_count
  IS 'Number of distinct GeoKrety that both users have interacted with';
COMMENT ON COLUMN stats.user_related_users.user_id
  IS 'Source user (authenticated only)';
COMMENT ON COLUMN stats.user_related_users.related_user_id
  IS 'Target user (authenticated only); never equals user_id (self-relation prevented by CHECK)';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.user_related_users;');
    }
}
```

## Data Contract

| Column                  | Type          | Nullable | Default | Description                                           |
| ----------------------- | ------------- | -------- | ------- | ----------------------------------------------------- |
| `user_id`               | `INT`         | NOT NULL | —       | **PK (part 1)** — Source user ID                      |
| `related_user_id`       | `INT`         | NOT NULL | —       | **PK (part 2)** — Target user ID (≠ `user_id`)        |
| `shared_geokrety_count` | `BIGINT`      | NOT NULL | `0`     | Distinct GeoKrety both users have interacted with     |
| `first_seen_at`         | `TIMESTAMPTZ` | NOT NULL | —       | When this user pair first "met" via a GeoKret         |
| `last_seen_at`          | `TIMESTAMPTZ` | NOT NULL | —       | When this user pair most recently "met" via a GeoKret |

**Constraint:** `chk_user_related_users_no_self` — CHECK `user_id <> related_user_id` prevents self-links.

**Trigger behaviour (S4T09):** When user A interacts with a GeoKret that users B and C have also touched:

1. Upsert `(A, B)` — increment `shared_geokrety_count`
2. Upsert `(B, A)` — same increment for reverse direction
3. Do the same for `(A, C)` and `(C, A)`

## SQL Usage Examples

```sql
-- UC2: Who has user 9876 connected with, ordered by strength?
SELECT related_user_id, shared_geokrety_count, last_seen_at
FROM stats.user_related_users
WHERE user_id = 9876
ORDER BY shared_geokrety_count DESC
LIMIT 20;

-- Symmetric check: if (A,B) exists, (B,A) must also exist
SELECT a.user_id, a.related_user_id, a.shared_geokrety_count,
       b.shared_geokrety_count AS reverse_count
FROM stats.user_related_users a
JOIN stats.user_related_users b ON b.user_id = a.related_user_id AND b.related_user_id = a.user_id
LIMIT 10;

-- Most socially connected users (hub detection)
SELECT user_id, COUNT(*) AS distinct_connections, SUM(shared_geokrety_count) AS total_shared
FROM stats.user_related_users
GROUP BY user_id
ORDER BY distinct_connections DESC
LIMIT 20;

-- Common connections between two users (via SQL join on gk_related_users)
-- (used in advanced graph queries, not direct from this table)
SELECT DISTINCT r1.geokrety_id
FROM stats.gk_related_users r1
JOIN stats.gk_related_users r2 ON r1.geokrety_id = r2.geokrety_id
WHERE r1.user_id = 100 AND r2.user_id = 200;

-- Self-relation guard check: returns 0 if integrity holds
SELECT COUNT(*) AS self_relations
FROM stats.user_related_users
WHERE user_id = related_user_id;
```

## Graph / Visualization Specification

1. **UC2: User social-network graph** — Force-directed graph (D3.js)
   - Nodes: users; edges: `(user_id, related_user_id)` where `shared_geokrety_count >= 2`
   - Edge weight: `shared_geokrety_count`
   - Only one direction needed for undirected graph visualisation: `WHERE user_id < related_user_id`
   - Data: `SELECT user_id, related_user_id, shared_geokrety_count FROM stats.user_related_users WHERE user_id < related_user_id AND shared_geokrety_count >= 2 LIMIT 500`

2. **Most connected users** — Table/leaderboard
   - Data: `SELECT user_id, COUNT(*) AS connections FROM stats.user_related_users GROUP BY user_id ORDER BY connections DESC LIMIT 50`

```
ASCII Sample — Social graph fragment:
User A ──── (23 GKs) ──── User B
     |                       |
  (8 GKs)                 (5 GKs)
     |                       |
User C ──── (12 GKs) ─── User D

Top connected users:
User_912  ████████████████████ 430 connections
User_003  ███████████████████  407 connections
```

## TimescaleDB Assessment

**NOT recommended.** Social-relation rollup table with PK `(user_id, related_user_id)`. Writes are UPSERT patterns. No time-series partitioning appropriate.

## pgTAP Unit Tests

| Test ID   | Assertion                                                                                     | Expected         |
| --------- | --------------------------------------------------------------------------------------------- | ---------------- |
| T-4.7.001 | `has_table('stats', 'user_related_users')`                                                    | pass             |
| T-4.7.002 | `col_is_pk('stats', 'user_related_users', ARRAY['user_id', 'related_user_id'])`               | pass             |
| T-4.7.003 | `col_type_is('stats', 'user_related_users', 'user_id', 'integer')`                            | pass             |
| T-4.7.004 | `col_type_is('stats', 'user_related_users', 'related_user_id', 'integer')`                    | pass             |
| T-4.7.005 | `col_type_is('stats', 'user_related_users', 'shared_geokrety_count', 'bigint')`               | pass             |
| T-4.7.006 | `col_default_is('stats', 'user_related_users', 'shared_geokrety_count', '0')`                 | pass             |
| T-4.7.007 | CHECK constraint `chk_user_related_users_no_self` prevents `user_id = related_user_id` insert | raises exception |
| T-4.7.008 | Table is empty after creation                                                                 | pass             |
| T-4.7.009 | `phinx rollback` drops table                                                                  | pass             |

| Test ID   | pgTAP Call                                                                      | Pass Condition                              |
| --------- | ------------------------------------------------------------------------------- | ------------------------------------------- |
| T-4.7.001 | `has_table('stats', 'user_related_users')`                                      | Table exists                                |
| T-4.7.002 | `col_is_pk('stats', 'user_related_users', ARRAY['user_id','related_user_id'])`  | Composite PK                                |
| T-4.7.003 | `col_type_is('stats', 'user_related_users', 'user_id', 'integer')`              | Type integer                                |
| T-4.7.004 | `col_type_is('stats', 'user_related_users', 'related_user_id', 'integer')`      | Type integer                                |
| T-4.7.005 | `col_type_is('stats', 'user_related_users', 'shared_geokrety_count', 'bigint')` | Type bigint                                 |
| T-4.7.006 | `col_default_is('stats', 'user_related_users', 'shared_geokrety_count', '0')`   | Default 0                                   |
| T-4.7.007 | Self-link INSERT on `user_id = related_user_id`                                 | Raises CHECK exception                      |
| T-4.7.008 | `is_empty('SELECT * FROM stats.user_related_users')`                            | 0 rows after DDL                            |
| T-4.7.009 | Rollback removes table                                                          | `hasnt_table('stats','user_related_users')` |

---

## Implementation Checklist

- [ ] 1. Create migration `20260310400600_create_user_related_users.php`
- [ ] 2. Run `phinx migrate` — confirm no errors
- [ ] 3. `\d stats.user_related_users` — 5 columns with correct types
- [ ] 4. Verify PK on `(user_id, related_user_id)`
- [ ] 5. Test self-link: `INSERT INTO stats.user_related_users VALUES (1,1,0,now(),now())` → expect CHECK violation
- [ ] 6. Table empty after creation
- [ ] 7. Run pgTAP tests T-4.7.001 through T-4.7.009
- [ ] 8. Verify `phinx rollback` drops table cleanly

- [ ] 1. Write `20260310400600_create_user_related_users.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Confirm 5 columns in `\d stats.user_related_users`
- [ ] 4. Verify composite PK and CHECK constraint present
- [ ] 5. Test self-link violation manually
- [ ] 6. Table is empty
- [ ] 7. Run pgTAP T-4.7.001 through T-4.7.009 — all pass
- [ ] 8. `phinx rollback` — table gone

## Table Created

```
stats.user_related_users (user_id, related_user_id, shared_geokrety_count, first_seen_at, last_seen_at)
```

| Column                  | Type        | Constraints                      |
| ----------------------- | ----------- | -------------------------------- |
| `user_id`               | INT         | PK (part 1), NOT NULL            |
| `related_user_id`       | INT         | PK (part 2), NOT NULL, ≠ user_id |
| `shared_geokrety_count` | BIGINT      | NOT NULL DEFAULT 0               |
| `first_seen_at`         | TIMESTAMPTZ | NOT NULL                         |
| `last_seen_at`          | TIMESTAMPTZ | NOT NULL                         |

## No FK to geokrety.\* schema

`user_id` and `related_user_id` are bare `INT`s. No cross-schema foreign keys. Referential integrity enforced by trigger logic (S4T09).

---
