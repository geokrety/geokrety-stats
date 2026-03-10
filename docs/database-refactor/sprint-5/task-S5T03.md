---
title: "Task S5T03: stats.gk_milestone_events Table"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 5
tags:
  - analytics
  - database
  - database-refactor
  - dba
  - hall-of-fame
  - milestone-events
  - specification
  - sprint-5
  - sql
  - stats
  - table
  - task-index
  - task-merge
  - uc13
depends_on:
  - "Sprint 2"
  - "Sprint 3"
task: S5T03
step: 5.3
migration: 20260310500200_create_gk_milestone_events.php
blocks:
  - S5T07
  - S5T10
changelog:
  - 2026-03-10: created by merge of task-S5T03.dba.md and task-S5T03.specification.md
---

# Task S5T03: stats.gk_milestone_events Table

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T03.dba.md`
- Specification source: `task-S5T03.specification.md`

## Purpose & Scope

Creates `stats.gk_milestone_events`, which records one-time milestone achievements for individual GeoKrety. Each event type can only be recorded once per GeoKret. Examples: "This GK has reached 100 km", "This GK has been held by 10 different users", "This GK has crossed to a new country for the first time."

These events enable:

- Timeline views of GK history milestones (UC13)
- Milestone-based notifications (future)
- Hall-of-fame moments ("First GK to reach 10,000 km")

**Key design:** This table is an **append-only event log**. Once an event is written, it is never updated. The `UNIQUE (gk_id, event_type)` constraint ensures each milestone is recorded only once per GeoKret.

Append-only event log recording when each GeoKret crosses a significant threshold. At most 8 distinct milestone types per GeoKret (one row per `(gk_id, event_type)`). Enables milestone timeline views (UC13) and hall-of-fame queries.

**Scope:** DDL + 2 secondary indexes. Trigger population in S5T07. Backfill in Sprint 6.

---

## Requirements

| ID      | Description                                                                                                          | MoSCoW |
| ------- | -------------------------------------------------------------------------------------------------------------------- | ------ |
| REQ-540 | Table `stats.gk_milestone_events` exists                                                                             | MUST   |
| REQ-541 | Surrogate PK `id BIGSERIAL`                                                                                          | MUST   |
| REQ-542 | Milestone uniqueness is enforced by canonical trigger/batch logic; the base table contract does not require a blanket UNIQUE constraint | MUST   |
| REQ-543 | CHECK `event_type IN ('country_first','km_100','km_1000','km_10000','users_10','users_50','users_100','first_find')` | MUST   |
| REQ-544 | `metadata JSONB NULL` — extensible milestone metadata                                                                  | MUST   |
| REQ-545 | `occurred_at TIMESTAMPTZ NOT NULL` — when the milestone was actually crossed (not recorded time)                     | MUST   |
| REQ-546 | `actor_user_id INT NULL` — actor responsible for the milestone when available                                        | MUST   |
| REQ-547 | Index `idx_gk_milestone_events_gk` on `(gk_id, occurred_at DESC)`                                                    | MUST   |
| REQ-548 | Index `idx_gk_milestone_events_type` on `(event_type, occurred_at DESC)`                                             | MUST   |
| REQ-549 | Table is empty after DDL creation                                                                                    | MUST   |
| REQ-550 | `phinx rollback` drops table and both indexes                                                                        | MUST   |

---

## Acceptance Criteria

| #   | Criterion                                        | How to Verify                      |
| --- | ------------------------------------------------ | ---------------------------------- |
| 1   | Table exists in `stats`                          | `\d stats.gk_milestone_events`     |
| 2   | Canonical milestone row shape enforced           | Inspect columns and allowed event types |
| 3   | Unknown `event_type` rejected                    | Insert `'fake'` → CHECK exception  |
| 4   | `metadata` accepts valid JSONB                   | Insert with JSON object → succeeds |
| 5   | 2 secondary indexes created                      | `\di+ stats.idx_gk_milestone*`     |
| 6   | Table empty after creation                       | 0 rows                             |
| 7   | Rollback drops table and both indexes            | Both absent after rollback         |

---

## Migration File

**`20260310500200_create_gk_milestone_events.php`**

## Full SQL DDL

```sql
CREATE TABLE stats.gk_milestone_events (
  id              BIGSERIAL    NOT NULL,
  gk_id           INT          NOT NULL,
  event_type      TEXT         NOT NULL,
  event_value     NUMERIC      NULL,
  additional_data JSONB        NULL,
  occurred_at     TIMESTAMPTZ  NOT NULL,
  recorded_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id),
  UNIQUE (gk_id, event_type),
  CONSTRAINT chk_gk_milestone_event_type CHECK (
    event_type IN (
      'country_first',
      'km_100',
      'km_1000',
      'km_10000',
      'users_10',
      'users_50',
      'users_100',
      'first_find'
    )
  )
);

CREATE INDEX idx_gk_milestone_events_gk
  ON stats.gk_milestone_events (gk_id, occurred_at DESC);

CREATE INDEX idx_gk_milestone_events_type
  ON stats.gk_milestone_events (event_type, occurred_at DESC);

COMMENT ON TABLE stats.gk_milestone_events
  IS 'Append-only log of per-GK milestones; each event_type recorded at most once per GK';
COMMENT ON COLUMN stats.gk_milestone_events.event_type
  IS 'Milestone type: country_first, km_100, km_1000, km_10000, users_10, users_50, users_100, first_find';
COMMENT ON COLUMN stats.gk_milestone_events.event_value
  IS 'Numeric value at threshold (e.g. total km, user count)';
COMMENT ON COLUMN stats.gk_milestone_events.additional_data
  IS 'JSONB metadata (e.g. country code for country_first, actor user_id)';
COMMENT ON COLUMN stats.gk_milestone_events.occurred_at
  IS 'When the GK actually crossed the milestone (from move timestamp)';
COMMENT ON COLUMN stats.gk_milestone_events.recorded_at
  IS 'When this row was inserted into the stats DB';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkMilestoneEvents extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.gk_milestone_events (
  id              BIGSERIAL    NOT NULL,
  gk_id           INT          NOT NULL,
  event_type      TEXT         NOT NULL,
  event_value     NUMERIC      NULL,
  additional_data JSONB        NULL,
  occurred_at     TIMESTAMPTZ  NOT NULL,
  recorded_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id),
  UNIQUE (gk_id, event_type),
  CONSTRAINT chk_gk_milestone_event_type CHECK (
    event_type IN (
      'country_first','km_100','km_1000','km_10000',
      'users_10','users_50','users_100','first_find'
    )
  )
);

CREATE INDEX idx_gk_milestone_events_gk
  ON stats.gk_milestone_events (gk_id, occurred_at DESC);

CREATE INDEX idx_gk_milestone_events_type
  ON stats.gk_milestone_events (event_type, occurred_at DESC);
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP TABLE IF EXISTS stats.gk_milestone_events;
SQL
        );
    }
}
```

## Data Contract

| Column            | Type          | Nullable | Default | Description                              |
| ----------------- | ------------- | -------- | ------- | ---------------------------------------- |
| `id`              | `BIGSERIAL`   | NOT NULL | auto    | **PK** — Surrogate key                   |
| `gk_id`           | `INT`         | NOT NULL | —       | GeoKret ID (no cross-schema FK)          |
| `event_type`      | `TEXT`        | NOT NULL | —       | Milestone type (constrained to 8 values) |
| `event_value`     | `NUMERIC`     | NULL     | —       | E.g. km count or user count at threshold |
| `additional_data` | `JSONB`       | NULL     | —       | Extra context (country, actor, etc.)     |
| `occurred_at`     | `TIMESTAMPTZ` | NOT NULL | —       | When milestone was actually crossed      |
| `recorded_at`     | `TIMESTAMPTZ` | NOT NULL | `NOW()` | When row was inserted                    |

**Unique constraint:** `(gk_id, event_type)` — each milestone type fires at most once per GeoKret.

## SQL Usage Examples

```sql
-- UC13: Milestone timeline for GK 1234
SELECT event_type, event_value, occurred_at, additional_data
FROM stats.gk_milestone_events
WHERE gk_id = 1234
ORDER BY occurred_at;

-- Hall of fame: first GKs to reach 10,000 km
SELECT gk_id, event_value, occurred_at
FROM stats.gk_milestone_events
WHERE event_type = 'km_10000'
ORDER BY occurred_at
LIMIT 10;

-- How many GKs have reached each milestone?
SELECT event_type, COUNT(*) AS gk_count
FROM stats.gk_milestone_events
GROUP BY event_type
ORDER BY event_type;
```

## TimescaleDB Assessment

**NOT recommended.** Low-cardinality table: at most 8 events per GeoKret. Total rows: O(N_geokrety × 8). Hypertable overhead not justified.

## pgTAP Unit Tests

| Test ID   | Assertion                                                                 | Expected  |
| --------- | ------------------------------------------------------------------------- | --------- |
| T-5.3.001 | `has_table('stats', 'gk_milestone_events')`                               | pass      |
| T-5.3.002 | `col_is_pk('stats', 'gk_milestone_events', ARRAY['id'])`                  | pass      |
| T-5.3.003 | UNIQUE constraint `(gk_id, event_type)` exists                            | pass      |
| T-5.3.004 | `col_type_is('stats', 'gk_milestone_events', 'event_type', 'text')`       | pass      |
| T-5.3.005 | `col_type_is('stats', 'gk_milestone_events', 'additional_data', 'jsonb')` | pass      |
| T-5.3.006 | INSERT invalid `event_type = 'fake_type'` → CHECK violation               | exception |
| T-5.3.007 | Insert same `(gk_id, event_type)` twice → UNIQUE violation                | exception |
| T-5.3.008 | `recorded_at` defaults to `NOW()`                                         | pass      |
| T-5.3.009 | Table empty after creation                                                | pass      |
| T-5.3.010 | `phinx rollback` drops table and its indexes                              | pass      |

| Test ID   | Assertion                                          | Pass Condition     |
| --------- | -------------------------------------------------- | ------------------ |
| T-5.3.001 | Table exists                                       | `has_table()`      |
| T-5.3.002 | PK is `id`                                         | `col_is_pk()`      |
| T-5.3.003 | Canonical milestone columns present               | Column check       |
| T-5.3.004 | `event_type` type is text                          | `col_type_is()`    |
| T-5.3.005 | `metadata` is JSONB                                | `col_type_is()`    |
| T-5.3.006 | Invalid event_type → CHECK violation               | Exception          |
| T-5.3.007 | Milestone deduplication is handled by trigger/batch logic | behavior verified |
| T-5.3.009 | Table empty after creation                         | `is_empty()`       |
| T-5.3.010 | Rollback drops table + indexes                     | `hasnt_table()`    |

---

## Implementation Checklist

- [ ] 1. Create `20260310500200_create_gk_milestone_events.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. `\d stats.gk_milestone_events` — 7 columns, BIGSERIAL PK, UNIQUE + CHECK constraints
- [ ] 4. Test invalid event type → exception
- [ ] 5. Test duplicate `(gk_id, event_type)` → UNIQUE exception
- [ ] 6. Verify 2 secondary indexes created
- [ ] 7. Run pgTAP T-5.3.001 through T-5.3.010 — all pass
- [ ] 8. `phinx rollback` — table and indexes dropped

- [ ] 1. Write `20260310500200_create_gk_milestone_events.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify canonical milestone columns, PK, event-type CHECK, and supporting indexes
- [ ] 4. Test invalid event type and canonical deduplication behavior
- [ ] 5. Run pgTAP T-5.3.001 through the canonical milestone table checks — all pass
- [ ] 6. `phinx rollback` — table and indexes gone

## Milestone Event Types Reference

| event_type      | Trigger condition                     | event_value | additional_data example          |
| --------------- | ------------------------------------- | ----------- | -------------------------------- |
| `country_first` | GK reaches its first non-home country | —           | `{"country": "DE", "actor": 42}` |
| `km_100`        | Cumulative distance reaches 100 km    | 100         | `{"actor": 42}`                  |
| `km_1000`       | Distance reaches 1000 km              | 1000        | `{"actor": 42}`                  |
| `km_10000`      | Distance reaches 10000 km             | 10000       | `{"actor": 42}`                  |
| `users_10`      | GK touched by 10 distinct users       | 10          | `{"actor": 42}`                  |
| `users_50`      | GK touched by 50 distinct users       | 50          | `{"actor": 42}`                  |
| `users_100`     | GK touched by 100 distinct users      | 100         | `{"actor": 42}`                  |
| `first_find`    | First non-owner logs on this GK       | —           | `{"actor": 42}`                  |

## Master-Spec Alignment

This task is governed by [../00-SPEC-DRAFT-v1.obsolete.md](../00-SPEC-DRAFT-v1.obsolete.md), Section 5.7.

- Canonical table contract: `stats.gk_milestone_events(id, gk_id, event_type, occurred_at, actor_user_id, metadata)`.
- `event_type` remains `VARCHAR(50)` constrained to the master milestone set.
- Fields such as `event_value`, `additional_data`, and `recorded_at`, and a blanket uniqueness rule on `(gk_id, event_type)`, are not part of the canonical contract unless the master spec is amended.

## Table Created

```
stats.gk_milestone_events (id, gk_id, event_type, occurred_at, actor_user_id, metadata)
```

| Column            | Type        | Constraints                               |
| ----------------- | ----------- | ----------------------------------------- |
| `id`              | BIGSERIAL   | PK NOT NULL                               |
| `gk_id`           | INT         | NOT NULL                                  |
| `event_type`      | VARCHAR(50) | NOT NULL, CHECK 8 values                  |
| `actor_user_id`   | INT         | NULL                                      |
| `metadata`        | JSONB       | NULL                                      |
| `occurred_at`     | TIMESTAMPTZ | NOT NULL                                  |

---
