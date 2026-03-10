---
title: "Task S5T09: Batch Aggregation Functions"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 5
tags:
  - analytics
  - backfill
  - batch
  - database
  - database-refactor
  - dba
  - function
  - snapshot
  - specification
  - sprint-5
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S5T01
  - S5T02
task: S5T09
step: 5.9
migration: 20260310500800_create_batch_aggregation_functions.php
blocks:
  - S5T10
  - S6T05
changelog:
  - 2026-03-10: created by merge of task-S5T09.dba.md and task-S5T09.specification.md
  - 2026-03-10: resolved Q-035 by narrowing S5T09 to hourly and country-pair batch functions
---

# Task S5T09: Batch Aggregation Functions

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T09.dba.md`
- Specification source: `task-S5T09.specification.md`

## Resolved Decision

- S5T09 canonically owns the batch aggregation helpers for `stats.hourly_activity` and `stats.country_pair_flows` only.
- The canonical migration name is `20260310500800_create_batch_aggregation_functions.php`.
- Milestone and first-finder backfill functions are stale merge residue and are not part of this task.
- Batch helpers must use the canonical `stats.job_log` contract with `job_name`, `status`, `metadata`, `started_at`, and `completed_at`.

## Purpose & Scope

Creates idempotent batch functions that seed Sprint 5 aggregate tables from existing move history and can be safely re-run during deployment or repair.

| Function                                 | Seeds into                 |
| ---------------------------------------- | -------------------------- |
| `stats.fn_snapshot_hourly_activity()`    | `stats.hourly_activity`    |
| `stats.fn_snapshot_country_pair_flows()` | `stats.country_pair_flows` |

**Out of scope:**

- milestone backfills
- first-finder backfills
- live trigger logic

## Requirements

| ID      | Description                                                                                                      | MoSCoW |
| ------- | ---------------------------------------------------------------------------------------------------------------- | ------ |
| REQ-650 | `stats.fn_snapshot_hourly_activity()` aggregates `gk_moves` into `(date, hour, move_type)` buckets              | MUST   |
| REQ-651 | `stats.fn_snapshot_hourly_activity()` is idempotent via `ON CONFLICT DO UPDATE`                                 | MUST   |
| REQ-652 | `stats.fn_snapshot_country_pair_flows()` detects cross-country pairs from canonical qualifying move history      | MUST   |
| REQ-653 | `stats.fn_snapshot_country_pair_flows()` counts distinct GeoKrety per month/from/to tuple                       | MUST   |
| REQ-654 | Both functions are defined in `stats` schema with `SECURITY DEFINER`                                             | MUST   |
| REQ-655 | Both functions write canonical `stats.job_log` rows using `job_name`, `status`, `metadata`, `started_at`, `completed_at` | MUST |
| REQ-656 | Both functions return `BIGINT` affected-row counts, not status text                                              | MUST   |
| REQ-657 | `phinx rollback` drops both functions cleanly                                                                    | MUST   |

## Acceptance Criteria

| #   | Criterion                                               | How to Verify                                        |
| --- | ------------------------------------------------------- | ---------------------------------------------------- |
| 1   | Both snapshot functions exist in `stats` schema         | `\df stats.fn_snapshot_*`                            |
| 2   | Hourly snapshot produces expected bucket counts         | Seed known sample and compare exact rows             |
| 3   | Country-pair snapshot produces only cross-country pairs | Verify same-country pairs are absent                 |
| 4   | Re-running both functions is idempotent                 | Row counts and values unchanged on second run        |
| 5   | Both functions write canonical `stats.job_log` rows     | Inspect `job_name`, `status`, `metadata`, timestamps |
| 6   | Rollback removes both functions                         | `phinx rollback`                                     |

## Migration File

**`20260310500800_create_batch_aggregation_functions.php`**

## Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION stats.fn_snapshot_hourly_activity()
  RETURNS BIGINT
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
BEGIN
  -- Recompute hourly buckets from canonical move history, upsert into
  -- `stats.hourly_activity`, log execution in `stats.job_log`, and
  -- return the affected-row count.
  RETURN 0;
END;
$$;

CREATE OR REPLACE FUNCTION stats.fn_snapshot_country_pair_flows()
  RETURNS BIGINT
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
BEGIN
  -- Recompute cross-country month buckets from canonical qualifying move
  -- history, upsert into `stats.country_pair_flows`, log execution in
  -- `stats.job_log`, and return the affected-row count.
  RETURN 0;
END;
$$;
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateBatchAggregationFunctions extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION stats.fn_snapshot_hourly_activity()
  RETURNS BIGINT LANGUAGE plpgsql SECURITY DEFINER
AS $$
BEGIN
  RETURN 0;
END;
$$;

CREATE OR REPLACE FUNCTION stats.fn_snapshot_country_pair_flows()
  RETURNS BIGINT LANGUAGE plpgsql SECURITY DEFINER
AS $$
BEGIN
  RETURN 0;
END;
$$;
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP FUNCTION IF EXISTS stats.fn_snapshot_hourly_activity();
DROP FUNCTION IF EXISTS stats.fn_snapshot_country_pair_flows();
SQL
        );
    }
}
```

The placeholder bodies above must be replaced by the canonical batch aggregation logic required by REQ-650 through REQ-656.

## Canonical Notes

- `stats.job_log` writes must use the canonical column set only; `rows_affected` and `duration` are obsolete.
- Batch helpers are idempotent repair/backfill tools, not live triggers.
- Any merged text assigning milestone or first-finder batch ownership to S5T09 is obsolete.

## SQL Usage Examples

```sql
SELECT stats.fn_snapshot_hourly_activity();
SELECT stats.fn_snapshot_country_pair_flows();

SELECT *
FROM stats.job_log
WHERE job_name IN ('fn_snapshot_hourly_activity', 'fn_snapshot_country_pair_flows')
ORDER BY completed_at DESC;
```

## pgTAP Unit Tests

| Test ID   | Assertion                                                   | Pass Condition |
| --------- | ----------------------------------------------------------- | -------------- |
| T-5.9.001 | `stats.fn_snapshot_hourly_activity()` exists                | `has_function()` |
| T-5.9.002 | `stats.fn_snapshot_country_pair_flows()` exists             | `has_function()` |
| T-5.9.003 | Hourly snapshot produces expected aggregate rows            | exact match     |
| T-5.9.004 | Country-pair snapshot produces expected cross-country rows  | exact match     |
| T-5.9.005 | Re-running functions is idempotent                          | exact match     |
| T-5.9.006 | Canonical `stats.job_log` rows are written                  | pass            |
| T-5.9.007 | Rollback removes both functions                             | pass            |

## Implementation Checklist

- [ ] 1. Create `20260310500800_create_batch_aggregation_functions.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify `\df stats.fn_snapshot_hourly_activity`
- [ ] 4. Verify `\df stats.fn_snapshot_country_pair_flows`
- [ ] 5. Validate hourly aggregation against known sample data
- [ ] 6. Validate country-pair aggregation against sequence test data
- [ ] 7. Verify canonical `stats.job_log` writes
- [ ] 8. Run pgTAP T-5.9.001 through T-5.9.007
- [ ] 9. `phinx rollback` — both functions dropped

## Canonical Alignment

- S5T09 is the batch/manual aggregation task for hourly activity and country-pair flows only.
- Live milestone and first-finder detection belong to S5T07 and S5T08 respectively.
- Any stale return type or obsolete `stats.job_log` column usage in merged drafts is superseded by this task definition.

## Agent Loop Log

- 2026-03-10T19:55:00Z — `dba`: removed milestone and first-finder batch bodies from S5T09 and restored the canonical two-function batch scope.
- 2026-03-10T19:55:00Z — `critical-thinking`: reconciled the abbreviated Sprint 5 index description with the merged draft by treating extra batch helpers as stale residue.
- 2026-03-10T19:55:00Z — `specification`: canonized the migration name, `BIGINT` return contract, and canonical `stats.job_log` usage.

## Resolution

Q-035 is resolved by canonizing S5T09 as the hourly and country-pair batch aggregation task only.
