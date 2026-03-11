---
title: "Task S5T01: stats.hourly_activity Table"
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
  - hourly-activity
  - specification
  - sprint-5
  - sql
  - stats
  - table
  - task-index
  - task-merge
  - uc8
depends_on:
  - "Sprint 2 daily_activity"
task: S5T01
step: 5.1
migration: 20260310500000_create_hourly_activity.php
blocks:
  - S5T07
  - S5T08
  - S5T10
changelog:
  - 2026.03.10: created by merge of task-S5T01.dba.md and task-S5T01.specification.md
---

# Task S5T01: stats.hourly_activity Table

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T01.dba.md`
- Specification source: `task-S5T01.specification.md`

## Purpose & Scope

Creates `stats.hourly_activity`, which aggregates move counts by **calendar date**, **UTC hour** (0–23), and **move type** (0–5). This is the sub-daily granularity complement to Sprint 2's `daily_move_counts`. It enables:

- UC8: Seasonal/time-of-day heatmap ("when do people log moves?")
- Peak usage hour analytics
- Move type distribution per hour

The table is **append-optimized**: each unique `(activity_date, hour_utc, move_type)` triplet has exactly one row, incremented by trigger.

Stores sub-daily, per-move-type counters indexed by `(date, hour_utc, move_type)`. Each cell holds the total count of moves of a given type logged during a specific hour on a specific date. The table powers the UC8 time-of-day heatmap and peak usage analytics.

**Scope:** DDL only. Trigger population is in S5T07 (covers Sprint 5 analytics triggers). Backfill in Sprint 6.

---

## Requirements

| ID      | Description                                                                    | MoSCoW |
| ------- | ------------------------------------------------------------------------------ | ------ |
| REQ-520 | Table `stats.hourly_activity` exists in the `stats` schema                     | MUST   |
| REQ-521 | 3-part composite PK: `(activity_date, hour_utc, move_type)`                    | MUST   |
| REQ-522 | `hour_utc SMALLINT CHECK(0..23)`                                               | MUST   |
| REQ-523 | `move_type SMALLINT CHECK(0..5)`                                               | MUST   |
| REQ-524 | `move_count BIGINT DEFAULT 0 NOT NULL`                                         | MUST   |
| REQ-525 | Table is empty after DDL creation (trigger populates it; backfill in Sprint 6) | MUST   |
| REQ-526 | `phinx rollback` drops table cleanly                                           | MUST   |

---

## Acceptance Criteria

| #   | Criterion                              | How to Verify                 |
| --- | -------------------------------------- | ----------------------------- |
| 1   | Table `stats.hourly_activity` created  | `\d stats.hourly_activity`    |
| 2   | 3-part composite PK                    | Description shows PRIMARY KEY |
| 3   | `hour_utc = 24` raises CHECK violation | Manual INSERT test            |
| 4   | `move_type = 6` raises CHECK violation | Manual INSERT test            |
| 5   | Table is empty after DDL               | `SELECT COUNT(*) = 0`         |
| 6   | `phinx rollback` drops table           | Table absent                  |

---

## Migration File

**`20260310500000_create_hourly_activity.php`**

## Full SQL DDL

```sql
CREATE TABLE stats.hourly_activity (
  activity_date  DATE       NOT NULL,
  hour_utc       SMALLINT   NOT NULL,
  move_type      SMALLINT   NOT NULL,
  move_count     BIGINT     NOT NULL DEFAULT 0,
  PRIMARY KEY (activity_date, hour_utc, move_type),
  CONSTRAINT chk_hourly_activity_hour   CHECK (hour_utc   BETWEEN 0 AND 23),
  CONSTRAINT chk_hourly_activity_mtype  CHECK (move_type  BETWEEN 0 AND 5)
);

COMMENT ON TABLE stats.hourly_activity
  IS 'Aggregate move count by date, UTC hour (0-23), and move type; powers UC8 heatmap';
COMMENT ON COLUMN stats.hourly_activity.activity_date
  IS 'Calendar date (UTC) of the moves';
COMMENT ON COLUMN stats.hourly_activity.hour_utc
  IS 'Hour of day in UTC (0=midnight, 23=11pm)';
COMMENT ON COLUMN stats.hourly_activity.move_type
  IS 'Move type: 0=DROP, 1=GRAB, 2=COMMENT, 3=SEEN, 4=ARCHIVE, 5=DIP';
COMMENT ON COLUMN stats.hourly_activity.move_count
  IS 'Number of moves of this type logged at this date/hour';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateHourlyActivity extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.hourly_activity (
  activity_date  DATE       NOT NULL,
  hour_utc       SMALLINT   NOT NULL,
  move_type      SMALLINT   NOT NULL,
  move_count     BIGINT     NOT NULL DEFAULT 0,
  PRIMARY KEY (activity_date, hour_utc, move_type),
  CONSTRAINT chk_hourly_activity_hour   CHECK (hour_utc   BETWEEN 0 AND 23),
  CONSTRAINT chk_hourly_activity_mtype  CHECK (move_type  BETWEEN 0 AND 5)
);

COMMENT ON TABLE stats.hourly_activity
  IS 'Aggregate move count by date, UTC hour (0-23), and move type; powers UC8 heatmap';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.hourly_activity;');
    }
}
```

## Data Contract

| Column          | Type       | Nullable | Default | Description                         |
| --------------- | ---------- | -------- | ------- | ----------------------------------- |
| `activity_date` | `DATE`     | NOT NULL | —       | **PK (part 1)** — Calendar date UTC |
| `hour_utc`      | `SMALLINT` | NOT NULL | —       | **PK (part 2)** — UTC hour 0–23     |
| `move_type`     | `SMALLINT` | NOT NULL | —       | **PK (part 3)** — Move type 0–5     |
| `move_count`    | `BIGINT`   | NOT NULL | `0`     | Aggregate count for this cell       |

**Constraints:**

- `chk_hourly_activity_hour`: `hour_utc BETWEEN 0 AND 23`
- `chk_hourly_activity_mtype`: `move_type BETWEEN 0 AND 5`

## SQL Usage Examples

```sql
-- UC8: Heatmap — total moves per hour across all dates (last 12 months)
SELECT hour_utc, SUM(move_count) AS total
FROM stats.hourly_activity
WHERE activity_date >= NOW() - INTERVAL '12 months'
GROUP BY hour_utc
ORDER BY hour_utc;

-- UC8: Move type distribution by hour (last 30 days)
SELECT hour_utc, move_type, SUM(move_count) AS total
FROM stats.hourly_activity
WHERE activity_date >= NOW() - INTERVAL '30 days'
GROUP BY hour_utc, move_type
ORDER BY hour_utc, move_type;

-- Peak days
SELECT activity_date, SUM(move_count) AS daily_total
FROM stats.hourly_activity
GROUP BY activity_date
ORDER BY daily_total DESC
LIMIT 10;
```

## TimescaleDB Assessment

**CANDIDATE for hypertable** (time dimension: `activity_date`). With years of data and potentially 24×6 = 144 rows per day, volume grows manageable but chunk-based queries would benefit. Chunk by `activity_date` 1-month intervals.

```sql
-- Optional hypertable conversion (run after table creation and before loading data):
SELECT create_hypertable('stats.hourly_activity', 'activity_date',
  chunk_time_interval => INTERVAL '1 month',
  if_not_exists => TRUE
);
```

**Trade-off:** TimescaleDB hypertables break standard `ON CONFLICT DO UPDATE` for some PostgreSQL versions. If using vanilla PostgreSQL, standard table is sufficient given small row volume (~144 rows/day max).

## pgTAP Unit Tests

| Test ID   | Assertion                                                                              | Expected  |
| --------- | -------------------------------------------------------------------------------------- | --------- |
| T-5.1.001 | `has_table('stats', 'hourly_activity')`                                                | pass      |
| T-5.1.002 | `col_is_pk('stats', 'hourly_activity', ARRAY['activity_date','hour_utc','move_type'])` | pass      |
| T-5.1.003 | `col_type_is('stats', 'hourly_activity', 'hour_utc', 'smallint')`                      | pass      |
| T-5.1.004 | `col_type_is('stats', 'hourly_activity', 'move_type', 'smallint')`                     | pass      |
| T-5.1.005 | `col_type_is('stats', 'hourly_activity', 'move_count', 'bigint')`                      | pass      |
| T-5.1.006 | `col_default_is('stats', 'hourly_activity', 'move_count', '0')`                        | pass      |
| T-5.1.007 | INSERT `hour_utc = 24` → CHECK violation                                               | exception |
| T-5.1.008 | INSERT `move_type = 6` → CHECK violation                                               | exception |
| T-5.1.009 | Table is empty after creation                                                          | pass      |
| T-5.1.010 | `phinx rollback` drops table                                                           | pass      |

| Test ID   | Assertion                         | Pass Condition     |
| --------- | --------------------------------- | ------------------ |
| T-5.1.001 | Table exists                      | `has_table()`      |
| T-5.1.002 | 3-part PK                         | `col_is_pk()`      |
| T-5.1.003 | `hour_utc` is SMALLINT            | `col_type_is()`    |
| T-5.1.004 | `move_type` is SMALLINT           | `col_type_is()`    |
| T-5.1.005 | `move_count` is BIGINT            | `col_type_is()`    |
| T-5.1.006 | `move_count` defaults to 0        | `col_default_is()` |
| T-5.1.007 | `hour_utc = 24` → CHECK violation | Exception raised   |
| T-5.1.008 | `move_type = 6` → CHECK violation | Exception raised   |
| T-5.1.009 | Table empty after creation        | `is_empty()`       |
| T-5.1.010 | Rollback drops table              | `hasnt_table()`    |

---

## Implementation Checklist

- [ ] 1. Create migration `20260310500000_create_hourly_activity.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. `\d stats.hourly_activity` — 4 columns, 3-part PK, 2 CHECKs
- [ ] 4. Test `hour_utc = 24` violation
- [ ] 5. Test `move_type = 6` violation
- [ ] 6. Run pgTAP T-5.1.001 through T-5.1.010 — all pass
- [ ] 7. `phinx rollback` — table gone

- [ ] 1. Write `20260310500000_create_hourly_activity.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. 4 columns with correct types and 3-part PK
- [ ] 4. Both CHECK constraints verified
- [ ] 5. Run pgTAP T-5.1.001 through T-5.1.010 — all pass
- [ ] 6. `phinx rollback` — table gone

## Table Created

```
stats.hourly_activity (activity_date, hour_utc, move_type, move_count)
```

| Column          | Type     | Constraints                        |
| --------------- | -------- | ---------------------------------- |
| `activity_date` | DATE     | PK (part 1), NOT NULL              |
| `hour_utc`      | SMALLINT | PK (part 2), NOT NULL, CHECK 0..23 |
| `move_type`     | SMALLINT | PK (part 3), NOT NULL, CHECK 0..5  |
| `move_count`    | BIGINT   | NOT NULL, DEFAULT 0                |

---
