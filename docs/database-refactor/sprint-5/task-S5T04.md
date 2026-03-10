---
title: "Task S5T04: stats.first_finder_events Table"
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
  - first-finder
  - hall-of-fame
  - specification
  - sprint-5
  - sql
  - stats
  - table
  - task-index
  - task-merge
  - uc14
depends_on:
  - "Sprint 2"
task: S5T04
step: 5.4
migration: 20260310500300_create_first_finder_events.php
blocks:
  - S5T07
  - S5T10
changelog:
  - 2026-03-10: created by merge of task-S5T04.dba.md and task-S5T04.specification.md
---

# Task S5T04: stats.first_finder_events Table

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T04.dba.md`
- Specification source: `task-S5T04.specification.md`

## Purpose & Scope

Creates `stats.first_finder_events`, which records who was the **first non-owner** to interact with each GeoKret, and how quickly that happened after the GK was created. One row per GeoKret maximum.

This table powers:

- UC14: First Finder Hall of Fame ("fastest finders", "most first finds per user")
- "First find within 7 days" statistics
- User achievement: first-finder badge

**Key design:** PK is `gk_id` (one first-finder per GeoKret). The `finder_user_id` is the first non-owner to log a qualifying move. `hours_since_creation` is the time delta between GK creation and the first-finder event (as SMALLINT to cap at 65535 hours ≈ 7.5 years).

Records one row per GeoKret for the first non-owner interaction, capturing who found it first and how quickly. PK is `gk_id`. Powers UC14 First Finder Hall of Fame and "fastest finders" analytics.

**Scope:** DDL + 2 secondary indexes. Trigger population in S5T07 (handles first-find detection on INSERT). Backfill in Sprint 6.

---

## Requirements

| ID      | Description                                                                                         | MoSCoW |
| ------- | --------------------------------------------------------------------------------------------------- | ------ |
| REQ-560 | Table `stats.first_finder_events` exists                                                            | MUST   |
| REQ-561 | PK `gk_id INT` — one row per GeoKret maximum                                                        | MUST   |
| REQ-562 | `finder_user_id INT NOT NULL` — first non-owner user                                                | MUST   |
| REQ-563 | `move_id BIGINT NOT NULL` — source move that established first-finder status                        | MUST   |
| REQ-564 | `CHECK (move_type IN (0,1,3,5))` — only qualifying move types                                       | MUST   |
| REQ-565 | `hours_since_creation SMALLINT NOT NULL CHECK (>= 0)`                                               | MUST   |
| REQ-566 | `found_at TIMESTAMPTZ NOT NULL` — move timestamp                                                    | MUST   |
| REQ-567 | `gk_created_at TIMESTAMPTZ NOT NULL` — GK creation timestamp used for the 7-day cutoff             | MUST   |
| REQ-568 | Index `idx_first_finder_events_user` on `(finder_user_id, found_at DESC)` for leaderboard           | MUST   |
| REQ-569 | Partial index `idx_first_finder_events_hours` on `hours_since_creation WHERE <= 168` (7-day window) | MUST   |
| REQ-570 | Table is empty after DDL creation                                                                   | MUST   |
| REQ-571 | `phinx rollback` drops table and both indexes                                                       | MUST   |

---

## Acceptance Criteria

| #   | Criterion                                          | How to Verify                               |
| --- | -------------------------------------------------- | ------------------------------------------- |
| 1   | Table exists in `stats`                            | `\d stats.first_finder_events`              |
| 2   | PK is `gk_id`                                      | Description shows PK                        |
| 3   | `move_type = 2` (COMMENT) rejected by CHECK        | Manual INSERT → exception                   |
| 4   | `hours_since_creation = -1` rejected               | Manual INSERT → exception                   |
| 5   | First-finder row requires `move_id` and `gk_created_at` | Manual INSERT missing either field → exception |
| 6   | Second INSERT for same `gk_id` raises PK violation | Manual INSERT → exception                   |
| 6   | Partial index on `hours_since_creation <= 168`     | `\d+ stats.first_finder_events` shows WHERE |
| 7   | Table empty after creation                         | 0 rows                                      |
| 8   | Rollback drops table and both indexes              | Both absent after rollback                  |

---

## Migration File

**`20260310500300_create_first_finder_events.php`**

## Full SQL DDL

```sql
CREATE TABLE stats.first_finder_events (
  gk_id                INT          NOT NULL,
  finder_user_id       INT          NOT NULL,
  move_type            SMALLINT     NOT NULL,
  hours_since_creation SMALLINT     NOT NULL,
  found_at             TIMESTAMPTZ  NOT NULL,
  recorded_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  PRIMARY KEY (gk_id),
  CONSTRAINT chk_ffe_move_type CHECK (move_type IN (0, 1, 3, 5)),
  CONSTRAINT chk_ffe_hours_non_negative CHECK (hours_since_creation >= 0)
);

CREATE INDEX idx_first_finder_events_user
  ON stats.first_finder_events (finder_user_id, found_at DESC);

CREATE INDEX idx_first_finder_events_hours
  ON stats.first_finder_events (hours_since_creation)
  WHERE hours_since_creation <= 168;  -- within 7 days (168 hours)

COMMENT ON TABLE stats.first_finder_events
  IS 'First non-owner interaction per GeoKret; one row per GK; powers UC14 first-finder leaderboard';
COMMENT ON COLUMN stats.first_finder_events.gk_id
  IS 'GeoKret ID — PK; one first-find per GeoKret';
COMMENT ON COLUMN stats.first_finder_events.finder_user_id
  IS 'User ID of the first non-owner to interact with this GK';
COMMENT ON COLUMN stats.first_finder_events.move_type
  IS 'Move type of the first-find interaction: 0=DROP, 1=GRAB, 3=SEEN, 5=DIP';
COMMENT ON COLUMN stats.first_finder_events.hours_since_creation
  IS 'Hours between GK creation and this first-find event (capped at SMALLINT max ~65535h)';
COMMENT ON COLUMN stats.first_finder_events.found_at
  IS 'Timestamp of the first-find move';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateFirstFinderEvents extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.first_finder_events (
  gk_id                INT          NOT NULL,
  finder_user_id       INT          NOT NULL,
  move_type            SMALLINT     NOT NULL,
  hours_since_creation SMALLINT     NOT NULL,
  found_at             TIMESTAMPTZ  NOT NULL,
  recorded_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  PRIMARY KEY (gk_id),
  CONSTRAINT chk_ffe_move_type CHECK (move_type IN (0, 1, 3, 5)),
  CONSTRAINT chk_ffe_hours_non_negative CHECK (hours_since_creation >= 0)
);

CREATE INDEX idx_first_finder_events_user
  ON stats.first_finder_events (finder_user_id, found_at DESC);

CREATE INDEX idx_first_finder_events_hours
  ON stats.first_finder_events (hours_since_creation)
  WHERE hours_since_creation <= 168;
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.first_finder_events;');
    }
}
```

## Data Contract

| Column                 | Type          | Nullable | Default | Description                            |
| ---------------------- | ------------- | -------- | ------- | -------------------------------------- |
| `gk_id`                | `INT`         | NOT NULL | —       | **PK** — GeoKret ID                    |
| `finder_user_id`       | `INT`         | NOT NULL | —       | First non-owner user ID                |
| `move_type`            | `SMALLINT`    | NOT NULL | —       | Move type of first-find (0,1,3,5 only) |
| `hours_since_creation` | `SMALLINT`    | NOT NULL | —       | Hours from GK creation to first find   |
| `found_at`             | `TIMESTAMPTZ` | NOT NULL | —       | When the first-find occurred           |
| `recorded_at`          | `TIMESTAMPTZ` | NOT NULL | `NOW()` | When this row was inserted             |

## Graph / Visualization Specification

**UC14: First Finder Hall of Fame**

```sql
-- Top first finders (most GKs found first)
SELECT finder_user_id, COUNT(*) AS gks_found_first
FROM stats.first_finder_events
GROUP BY finder_user_id
ORDER BY gks_found_first DESC
LIMIT 20;

-- Fastest finders (only within 7-day window = 168 hours)
SELECT gk_id, finder_user_id, hours_since_creation, found_at
FROM stats.first_finder_events
WHERE hours_since_creation <= 168
ORDER BY hours_since_creation
LIMIT 20;

-- Distribution: how quickly are GKs first found?
SELECT
  CASE
    WHEN hours_since_creation <= 24   THEN '< 1 day'
    WHEN hours_since_creation <= 72   THEN '1-3 days'
    WHEN hours_since_creation <= 168  THEN '3-7 days'
    ELSE '> 7 days'
  END AS bucket,
  COUNT(*) AS gk_count
FROM stats.first_finder_events
GROUP BY 1
ORDER BY MIN(hours_since_creation);
```

## TimescaleDB Assessment

**NOT recommended.** One row per GeoKret (natural upper bound). TimescaleDB hypertable overhead not justified.

## pgTAP Unit Tests

| Test ID   | Assertion                                                                         | Expected  |
| --------- | --------------------------------------------------------------------------------- | --------- |
| T-5.4.001 | `has_table('stats', 'first_finder_events')`                                       | pass      |
| T-5.4.002 | `col_is_pk('stats', 'first_finder_events', ARRAY['gk_id'])`                       | pass      |
| T-5.4.003 | `col_type_is('stats', 'first_finder_events', 'hours_since_creation', 'smallint')` | pass      |
| T-5.4.004 | `col_type_is('stats', 'first_finder_events', 'move_type', 'smallint')`            | pass      |
| T-5.4.005 | CHECK: invalid `move_type = 2` → violation                                        | exception |
| T-5.4.006 | CHECK: `hours_since_creation = -1` → violation                                    | exception |
| T-5.4.007 | Second INSERT for same `gk_id` → PK violation                                     | exception |
| T-5.4.008 | Index `idx_first_finder_events_user` exists                                       | pass      |
| T-5.4.009 | Index `idx_first_finder_events_hours` exists (partial `<= 168`)                   | pass      |
| T-5.4.010 | Table empty after creation                                                        | pass      |
| T-5.4.011 | `phinx rollback` drops table and both indexes                                     | pass      |

| Test ID   | Assertion                                     | Pass Condition  |
| --------- | --------------------------------------------- | --------------- |
| T-5.4.001 | Table exists                                  | `has_table()`   |
| T-5.4.002 | PK is `gk_id`                                 | `col_is_pk()`   |
| T-5.4.003 | `hours_since_creation` is SMALLINT            | `col_type_is()` |
| T-5.4.004 | `move_type` is SMALLINT                       | `col_type_is()` |
| T-5.4.005 | `move_type = 2` → CHECK violation             | Exception       |
| T-5.4.006 | `hours_since_creation = -1` → CHECK violation | Exception       |
| T-5.4.007 | Duplicate `gk_id` → PK violation              | Exception       |
| T-5.4.008 | User index exists                             | `has_index()`   |
| T-5.4.009 | Partial hours index exists (`<= 168`)         | `has_index()`   |
| T-5.4.010 | Table empty after creation                    | `is_empty()`    |
| T-5.4.011 | Rollback removes table + indexes              | `hasnt_table()` |

---

## Implementation Checklist

- [ ] 1. Create `20260310500300_create_first_finder_events.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. `\d stats.first_finder_events` — 6 columns, `gk_id` PK, 2 CHECKs
- [ ] 4. Test `move_type = 2` (COMMENT) → CHECK violation
- [ ] 5. Test `hours_since_creation = -1` → CHECK violation
- [ ] 6. Test duplicate `gk_id` → PK violation
- [ ] 7. Verify partial index on `hours_since_creation <= 168`
- [ ] 8. Run pgTAP T-5.4.001 through T-5.4.011 — all pass
- [ ] 9. `phinx rollback` — table and indexes gone

- [ ] 1. Write `20260310500300_create_first_finder_events.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify canonical first-finder columns, PK, CHECK constraints, and indexes
- [ ] 4. Test CHECK and PK violations
- [ ] 5. Run pgTAP T-5.4.001 through T-5.4.011 — all pass
- [ ] 6. `phinx rollback` — clean removal

## Master-Spec Alignment

This task is governed by [../00-SPEC-DRAFT-v1.md](../00-SPEC-DRAFT-v1.md), Section 5.7 and Test Matrix items T142-T143.

- Canonical table contract: `stats.first_finder_events(gk_id, finder_user_id, move_id, move_type, found_at, gk_created_at, hours_since_creation)`.
- First-finder eligibility is limited to qualifying non-owner moves within 168 hours of GK creation.
- Any lower text that omits `move_id` or `gk_created_at`, or that treats the first-finder window as open-ended, is obsolete and superseded by this alignment block.

## Table Created

```
stats.first_finder_events (gk_id, finder_user_id, move_id, move_type, found_at, gk_created_at, hours_since_creation)
```

| Column                 | Type        | Constraints                  |
| ---------------------- | ----------- | ---------------------------- |
| `gk_id`                | INT         | PK, NOT NULL                 |
| `finder_user_id`       | INT         | NOT NULL                     |
| `move_id`              | BIGINT      | NOT NULL                     |
| `move_type`            | SMALLINT    | NOT NULL, CHECK IN (0,1,3,5) |
| `found_at`             | TIMESTAMPTZ | NOT NULL                     |
| `gk_created_at`        | TIMESTAMPTZ | NOT NULL                     |
| `hours_since_creation` | SMALLINT    | NOT NULL, CHECK >= 0         |

---
