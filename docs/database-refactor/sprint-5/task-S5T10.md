---
title: "Task S5T10: Sprint 5 Analytics Indexes"
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
  - index
  - performance
  - specification
  - sprint-5
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S5T01
  - S5T02
  - S5T05
  - S5T07
  - S5T09
task: S5T10
step: 5.10
migration: 20260310500900_create_analytics_indexes.php
blocks:
  - S6T05
changelog:
  - 2026-03-10: created by merge of task-S5T10.dba.md and task-S5T10.specification.md
---

# Task S5T10: Sprint 5 Analytics Indexes

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T10.dba.md`
- Specification source: `task-S5T10.specification.md`

## Purpose & Scope

Creates performance indexes on all Sprint 5 analytics tables. Sprint 5 tables that survive live-data writes (hourly_activity, country_pair_flows, outbox_events) are the primary targets. The milestone and first_finder tables have low cardinality PK-based access, so additional indexes are minimal.

Adds the incremental B-tree indexes needed by canonical Sprint 5 analytics tables. No new tables are created.

## Requirements

| ID      | Requirement                                                                                          |
| ------- | ---------------------------------------------------------------------------------------------------- |
| REQ-670 | Index `idx_hourly_activity_date_desc` on `stats.hourly_activity(activity_date DESC)` created         |
| REQ-671 | Index `idx_country_pair_flows_month_desc` on `stats.country_pair_flows(year_month DESC)` created     |
| REQ-672 | Index `idx_country_pair_flows_from` on `stats.country_pair_flows(from_country, year_month DESC)`     |
| REQ-673 | Index `idx_country_pair_flows_to` on `stats.country_pair_flows(to_country, year_month DESC)` created |
| REQ-674 | Event-bridge read index is defined only if a concrete bridge table is approved                        |
| REQ-675 | No requirement may depend on a non-canonical `processed` column                                       |
| REQ-676 | Indexes already owned by Sprint 5 table-creation tasks are not redefined here                         |
| REQ-677 | This task covers only incremental indexes not already required by the canonical table contracts       |
| REQ-678 | All indexes use `IF NOT EXISTS` to allow idempotent re-run                                           |
| REQ-679 | `phinx rollback` drops only the indexes actually created by this task                                |

## Migration File

**`20260310500900_create_analytics_indexes.php`**

---

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateAnalyticsIndexes extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
-- hourly_activity
CREATE INDEX IF NOT EXISTS idx_hourly_activity_date_desc
  ON stats.hourly_activity (activity_date DESC);

-- country_pair_flows
CREATE INDEX IF NOT EXISTS idx_country_pair_flows_month_desc
  ON stats.country_pair_flows (year_month DESC);

CREATE INDEX IF NOT EXISTS idx_country_pair_flows_from
  ON stats.country_pair_flows (from_country, year_month DESC);

CREATE INDEX IF NOT EXISTS idx_country_pair_flows_to
  ON stats.country_pair_flows (to_country, year_month DESC);

-- outbox_events
CREATE INDEX IF NOT EXISTS idx_outbox_events_created_at
  ON stats.outbox_events (created_at ASC);

CREATE INDEX IF NOT EXISTS idx_outbox_events_unprocessed
  ON stats.outbox_events (created_at ASC)
  WHERE processed = FALSE;

-- gk_milestone_events
CREATE INDEX IF NOT EXISTS idx_gk_milestone_events_type
  ON stats.gk_milestone_events (event_type);

-- first_finder_events
CREATE INDEX IF NOT EXISTS idx_first_finder_events_user
  ON stats.first_finder_events (finder_user_id);
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP INDEX IF EXISTS stats.idx_hourly_activity_date_desc;
DROP INDEX IF EXISTS stats.idx_country_pair_flows_month_desc;
DROP INDEX IF EXISTS stats.idx_country_pair_flows_from;
DROP INDEX IF EXISTS stats.idx_country_pair_flows_to;
DROP INDEX IF EXISTS stats.idx_outbox_events_created_at;
DROP INDEX IF EXISTS stats.idx_outbox_events_unprocessed;
DROP INDEX IF EXISTS stats.idx_gk_milestone_events_type;
DROP INDEX IF EXISTS stats.idx_first_finder_events_user;
SQL
        );
    }
}
```

## Data Contract

| Artifact                            | Table               | Columns                           |
| ----------------------------------- | ------------------- | --------------------------------- |
| `idx_hourly_activity_date_desc`     | hourly_activity     | `(activity_date DESC)`            |
| `idx_country_pair_flows_month_desc` | country_pair_flows  | `(year_month DESC)`               |
| `idx_country_pair_flows_from`       | country_pair_flows  | `(from_country, year_month DESC)` |
| `idx_country_pair_flows_to`         | country_pair_flows  | `(to_country, year_month DESC)`   |
| event-bridge read index             | approved bridge table only | implementation-specific        |
| optional bridge read index          | approved bridge table only | implementation-specific     |

## TimescaleDB Assessment

`stats.hourly_activity` and `stats.country_pair_flows` are natural candidates for TimescaleDB hypertables if the dataset grows beyond 50M rows. However, PostgreSQL B-tree indexes on `(activity_date DESC)` and `(year_month DESC)` provide equivalent performance for dashboard queries in the near term. TimescaleDB conversion is a Sprint 7+ operation if warranted.

## pgTAP Unit Tests

| Test ID    | Assertion                                                                                                 | Expected |
| ---------- | --------------------------------------------------------------------------------------------------------- | -------- |
| T-5.10.001 | `has_index('stats','hourly_activity','idx_hourly_activity_date_desc')`                                    | pass     |
| T-5.10.002 | `has_index('stats','country_pair_flows','idx_country_pair_flows_month_desc')`                             | pass     |
| T-5.10.003 | `has_index('stats','country_pair_flows','idx_country_pair_flows_from')`                                   | pass     |
| T-5.10.004 | `has_index('stats','country_pair_flows','idx_country_pair_flows_to')`                                     | pass     |
| T-5.10.005 | `has_index('stats','outbox_events','idx_outbox_events_created_at')`                                       | pass     |
| T-5.10.006 | `has_index('stats','outbox_events','idx_outbox_events_unprocessed')`                                      | pass     |
| T-5.10.007 | `has_index('stats','gk_milestone_events','idx_gk_milestone_events_type')`                                 | pass     |
| T-5.10.008 | `has_index('stats','first_finder_events','idx_first_finder_events_user')`                                 | pass     |
| T-5.10.009 | EXPLAIN plan for `SELECT … FROM hourly_activity WHERE activity_date > now()-interval '7 days'` uses index | pass     |
| T-5.10.010 | `phinx rollback` — all 8 indexes dropped                                                                  | pass     |

| Test ID    | Criterion | pgTAP Assertion                                                | Pass Condition  |
| ---------- | --------- | -------------------------------------------------------------- | --------------- |
| T-5.10.001 | REQ-670   | `has_index('stats','hourly_activity','idx_...')`               | pass            |
| T-5.10.002 | REQ-671   | `has_index('stats','country_pair_flows','idx_..._month_desc')` | pass            |
| T-5.10.003 | REQ-672   | `has_index('stats','country_pair_flows','idx_..._from')`       | pass            |
| T-5.10.004 | REQ-673   | `has_index('stats','country_pair_flows','idx_..._to')`         | pass            |
| T-5.10.005 | REQ-674   | optional bridge read index is checked only if a bridge table exists | pass       |
| T-5.10.009 | AC-5.10.1 | EXPLAIN shows index scan for recent `activity_date` query      | index_scan      |
| T-5.10.010 | AC-5.10.5 | Task-owned indexes absent after rollback                       | `NOT has_index` |

## Implementation Checklist

- [ ] 1. Create `20260310500900_create_analytics_indexes.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify all 8 index names via `\di stats.*`
- [ ] 4. EXPLAIN on recent-activity query — confirm index scan
- [ ] 5. EXPLAIN on country-flows from-country query — confirm index scan
- [ ] 6. Verify partial index on outbox via `pg_indexes` system view
- [ ] 7. Run pgTAP T-5.10.001 through T-5.10.010
- [ ] 8. `phinx rollback` — all 8 indexes dropped cleanly

- [ ] Create `20260310500900_create_analytics_indexes.php`
- [ ] Run `phinx migrate` — no errors
- [ ] Verify the task-owned canonical indexes via `\di stats.*`
- [ ] Run EXPLAIN on key dashboard queries to confirm index scans
- [ ] Confirm any implementation-specific bridge index only if a bridge table exists
- [ ] Run pgTAP T-5.10.001 through T-5.10.010
- [ ] `phinx rollback` — all 8 indexes dropped

## hourly_activity — Descending Date

```sql
CREATE INDEX IF NOT EXISTS idx_hourly_activity_date_desc
  ON stats.hourly_activity (activity_date DESC);
```

_Used by: dashboard queries reading the last N days of activity. The primary key already covers (activity_date, hour_utc, move_type) for exact lookups but a single-column DESC index speeds up range scans from the API._

---

## country_pair_flows — year_month Descending

```sql
CREATE INDEX IF NOT EXISTS idx_country_pair_flows_month_desc
  ON stats.country_pair_flows (year_month DESC);
```

_Used by: leaderboard / map queries that ORDER BY year_month DESC LIMIT N._

---

## country_pair_flows — from_country, year_month Descending

```sql
CREATE INDEX IF NOT EXISTS idx_country_pair_flows_from
  ON stats.country_pair_flows (from_country, year_month DESC);
```

_Used by: "show recent outbound flows from PL" queries. Covers WHERE from_country='PL' ORDER BY year_month DESC._

---

## country_pair_flows — to_country, year_month Descending

```sql
CREATE INDEX IF NOT EXISTS idx_country_pair_flows_to
  ON stats.country_pair_flows (to_country, year_month DESC);
```

_Used by: "show recent inbound flows to DE" queries._

---

## outbox_events — created_at Ascending (Relay Consumption)

```sql
CREATE INDEX IF NOT EXISTS idx_outbox_events_created_at
  ON stats.outbox_events (created_at ASC);
```

_Used by: AMQP relay worker that reads unprocessed events ORDER BY created_at ASC LIMIT N. Replaces the partial index design (which required `WHERE processed = FALSE`) to keep it future-proof as other states could be added._

---

## outbox_events — Partial Index on Unprocessed Events (Optional Fast-Path)

```sql
CREATE INDEX IF NOT EXISTS idx_outbox_events_unprocessed
  ON stats.outbox_events (created_at ASC)
  WHERE processed = FALSE;
```

_Small partial index for relay worker hot-path. Invisible to index scanner once all rows are processed._

---

## gk_milestone_events — event_type

```sql
CREATE INDEX IF NOT EXISTS idx_gk_milestone_events_type
  ON stats.gk_milestone_events (event_type);
```

_Used by: "how many GKs have hit km_100?" aggregate queries._

---

## first_finder_events — finder_user_id

```sql
CREATE INDEX IF NOT EXISTS idx_first_finder_events_user
  ON stats.first_finder_events (finder_user_id);
```

_Used by: "list all GKs first found by user X" profile queries._

---

## Index Summary Table

| Index Name                        | Table               | Columns                       | Type   | Rationale                       |
| --------------------------------- | ------------------- | ----------------------------- | ------ | ------------------------------- |
| idx_hourly_activity_date_desc     | hourly_activity     | activity_date DESC            | B-tree | Date-range scans for dashboards |
| idx_country_pair_flows_month_desc | country_pair_flows  | year_month DESC               | B-tree | Latest-month ordering           |
| idx_country_pair_flows_from       | country_pair_flows  | from_country, year_month DESC | B-tree | Outbound flows per country      |
| idx_country_pair_flows_to         | country_pair_flows  | to_country, year_month DESC   | B-tree | Inbound flows per country       |
| idx_outbox_events_created_at      | outbox_events       | created_at ASC                | B-tree | Relay worker full scan          |
| idx_outbox_events_unprocessed     | outbox_events       | created_at ASC, partial       | B-tree | Relay worker hot-path           |
| idx_gk_milestone_events_type      | gk_milestone_events | event_type                    | B-tree | Aggregate by milestone type     |
| idx_first_finder_events_user      | first_finder_events | finder_user_id                | B-tree | User profile "firsts" queries   |

## Master-Spec Alignment

This task is governed by [../00-SPEC-DRAFT-v1.md](../00-SPEC-DRAFT-v1.md), Section 9 and the canonical table contracts for Sprint 5.

- Do not assume a `processed` column on any event bridge table unless the master spec is amended to define one.
- Indexes already owned by table-creation tasks must not be re-specified here as new mandatory Sprint 5 indexes.
- Any lower text that depends on non-canonical event-bridge columns is obsolete and superseded by this alignment block.

## AC-5.10.1 — Hourly Activity Index Used

**Given** `hourly_activity` table populated, query filters `activity_date > NOW() - INTERVAL '7 days'`
**When** EXPLAIN ANALYZE run
**Then** query plan shows Index Scan on `idx_hourly_activity_date_desc`

## AC-5.10.2 — Country Pair Flows From-Country Filter

**Given** `country_pair_flows` rows exist, query does `WHERE from_country = 'PL' ORDER BY year_month DESC`
**When** EXPLAIN ANALYZE run
**Then** plan shows Index Scan on `idx_country_pair_flows_from`

## AC-5.10.3 — Event Bridge Read Index

**Given** a concrete bridge-table implementation is approved
**When** the bridge reader query is benchmarked
**Then** the approved implementation-specific read index is used

## AC-5.10.4 — All Indexes Present

**Given** migration ran
**When** `SELECT indexname FROM pg_indexes WHERE schemaname='stats'` queried
**Then** the canonical indexes created by this task are listed

## AC-5.10.5 — Rollback Removes All Indexes

**Given** migration has been applied
**When** `phinx rollback` run
**Then** none of the indexes created by this task appear in `pg_indexes WHERE schemaname='stats'`
