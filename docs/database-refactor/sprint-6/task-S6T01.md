---
title: "Task S6T01: Batched previous_move_id Backfill Function"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 6
tags:
  - backfill
  - database
  - database-refactor
  - dba
  - function
  - previous-move-id
  - specification
  - sprint-6
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S1
  - S2T05
task: S6T01
step: 6.1
migration: 20260310600000_create_backfill_previous_move.php
blocks:
  - S6T02
  - S6T03
changelog:
  - 2026.03.10: created by merge of task-S6T01.dba.md and task-S6T01.specification.md
  - 2026.03.10: resolved Q-036 by canonizing the backfill job_log contract
---

# Task S6T01: Batched previous_move_id Backfill Function

## Sprint Context

- Sprint index: Sprint 6 Task Index
- Tags: database, database-refactor, sprint-6, task-index

## Source

- DBA source: `task-S6T01.dba.md`
- Specification source: `task-S6T01.specification.md`

## Resolved Decision

- S6T01 keeps the batched `previous_move_id` backfill scope only.
- The canonical migration name is `20260310600000_create_backfill_previous_move.php`.
- Backfill helpers write to `stats.job_log` using only `job_name`, `status`, `metadata`, `started_at`, and `completed_at`.
- Legacy `rows_affected` and `duration` columns are obsolete and must not appear in this task.

## Purpose & Scope

Creates `stats.fn_backfill_previous_move_id(p_period TSTZRANGE DEFAULT tstzrange('-infinity', 'infinity'), p_batch_size INT DEFAULT 50000)` to populate `geokrety.gk_moves.previous_move_id` for historical rows that predate the live Sprint 2 trigger.

The helper is batch-oriented, idempotent, and operationally safe to re-run. It owns only historical `previous_move_id` reconstruction on `geokrety.gk_moves`.

## Canonical job_log Contract

Every S6 backfill helper that writes a completion row into `stats.job_log` must follow this contract:

- `job_name`: stable helper identifier such as `fn_backfill_previous_move_id`
- `status`: terminal state such as `ok` or `error`
- `metadata`: JSON payload carrying run-specific details such as `period`, `batch_size`, `rows_updated`, and any notes
- `started_at`: timestamp captured before work begins
- `completed_at`: timestamp captured when the helper finishes or fails

Canonical constraints for this task:

- `started_at <= completed_at`
- `metadata` is the only place for per-run counters or duration-like detail
- legacy columns `rows_affected` and `duration` are not part of the contract and must not be referenced

## Requirements

| ID      | Description                                                                                              | MoSCoW |
| ------- | -------------------------------------------------------------------------------------------------------- | ------ |
| REQ-700 | Function `stats.fn_backfill_previous_move_id(TSTZRANGE, INT)` exists                                     | MUST   |
| REQ-701 | Default period is all-time and default batch size is `50000`                                             | MUST   |
| REQ-702 | Each GK row receives the most recent earlier qualifying move from the same GK only                       | MUST   |
| REQ-703 | First qualifying move per GK keeps `previous_move_id` as `NULL`                                          | MUST   |
| REQ-704 | Re-running the helper does not corrupt already-correct links                                              | MUST   |
| REQ-705 | Completion writes a canonical `stats.job_log` row using only `job_name`, `status`, `metadata`, `started_at`, `completed_at` | MUST   |
| REQ-706 | `metadata` includes at least `period`, `batch_size`, and `rows_updated`                                  | MUST   |
| REQ-707 | `phinx rollback` drops the function cleanly                                                               | MUST   |

## Acceptance Criteria

| #   | Criterion                                                | How to Verify |
| --- | -------------------------------------------------------- | ------------- |
| 1   | Function exists in `stats` schema                        | `\df stats.fn_backfill_previous_move_id` |
| 2   | Historical rows get correct predecessor IDs             | Seed ordered GK history and compare IDs |
| 3   | First move per GK stays `NULL`                           | Verify seeded first row |
| 4   | Re-run is idempotent                                     | Second run does not change correct rows |
| 5   | `stats.job_log` row uses canonical columns only          | Inspect `job_name`, `status`, `metadata`, timestamps |
| 6   | Rollback removes the helper                              | `phinx rollback` |

## Migration File

**`20260310600000_create_backfill_previous_move.php`**

## Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION stats.fn_backfill_previous_move_id(
  p_period     TSTZRANGE DEFAULT tstzrange('-infinity', 'infinity'),
  p_batch_size INT       DEFAULT 50000
)
RETURNS BIGINT
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
  v_started_at TIMESTAMPTZ := clock_timestamp();
  v_rows_updated BIGINT := 0;
BEGIN
  -- Canonical implementation:
  -- 1. Process qualifying rows inside the requested period in batches.
  -- 2. Reconstruct predecessor links per GK only.
  -- 3. Leave first qualifying move per GK with NULL previous_move_id.
  -- 4. Write a canonical stats.job_log row with JSON metadata.

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES (
    'fn_backfill_previous_move_id',
    'ok',
    jsonb_build_object(
      'period', p_period,
      'batch_size', p_batch_size,
      'rows_updated', v_rows_updated
    ),
    v_started_at,
    clock_timestamp()
  );

  RETURN v_rows_updated;
END;
$$;
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateBackfillPreviousMove extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION stats.fn_backfill_previous_move_id(
  p_period     TSTZRANGE DEFAULT tstzrange('-infinity', 'infinity'),
  p_batch_size INT       DEFAULT 50000
)
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
DROP FUNCTION IF EXISTS stats.fn_backfill_previous_move_id(TSTZRANGE, INT);
SQL
        );
    }
}
```

The placeholder body above must be replaced by the canonical batch logic from REQ-702 through REQ-706.

## Canonical Notes

- This task owns only `geokrety.gk_moves.previous_move_id` backfill.
- The helper may expose row counts to callers via its `BIGINT` return value and via `metadata`; it must not depend on removed `stats.job_log` columns.
- Full-history orchestration belongs to S6T02.

## pgTAP Unit Tests

| Test ID   | Assertion                                                        | Pass Condition |
| --------- | ---------------------------------------------------------------- | -------------- |
| T-6.1.001 | `stats.fn_backfill_previous_move_id()` exists                    | `has_function()` |
| T-6.1.002 | Middle move points to correct predecessor                        | exact match |
| T-6.1.003 | First move per GK remains `NULL`                                 | exact match |
| T-6.1.004 | Cross-GK contamination does not occur                            | exact match |
| T-6.1.005 | Re-run is idempotent                                             | exact match |
| T-6.1.006 | `stats.job_log` row uses canonical fields only                   | pass |
| T-6.1.007 | Rollback removes function                                        | pass |

## Implementation Checklist

- [ ] 1. Create `20260310600000_create_backfill_previous_move.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify `\df stats.fn_backfill_previous_move_id`
- [ ] 4. Test ordered predecessor reconstruction on sample history
- [ ] 5. Verify canonical `stats.job_log` metadata payload
- [ ] 6. Run pgTAP T-6.1.001 through T-6.1.007
- [ ] 7. `phinx rollback` — function dropped

## Agent Loop Log

- 2026-03-10T21:05:00Z — `dba`: removed obsolete `rows_affected` and `duration` references and restored the Sprint 6 migration name.
- 2026-03-10T21:05:00Z — `critical-thinking`: treated the job-log payload as metadata, not schema drift, to preserve the canonical support-table contract.
- 2026-03-10T21:05:00Z — `specification`: narrowed S6T01 to predecessor backfill plus explicit job-log write requirements.

## Resolution

Q-036 is resolved by canonizing the S6 backfill `stats.job_log` write contract in S6T01.
