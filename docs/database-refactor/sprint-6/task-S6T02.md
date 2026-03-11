---
title: "Task S6T02: Heavy previous_move_id Orchestrator"
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
  - orchestrator
  - specification
  - sprint-6
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S6T01
task: S6T02
step: 6.2
migration: 20260310600200_create_backfill_heavy_previous_move_id.php
blocks:
  - S6T05
changelog:
  - 2026.03.10: created by merge of task-S6T02.dba.md and task-S6T02.specification.md
---

# Task S6T02: Heavy previous_move_id Orchestrator

## Sprint Context

- Sprint index: Sprint 6 Task Index
- Tags: database, database-refactor, sprint-6, task-index

## Source

- DBA source: `task-S6T02.dba.md`
- Specification source: `task-S6T02.specification.md`

## Purpose & Scope

Creates `stats.fn_backfill_heavy_previous_move_id_all()` — a convenience wrapper that calls `stats.fn_backfill_previous_move_id()` decade-by-decade to break the full history into manageable chunks, making a from-scratch production backfill safe to run without manual period calculation.

Wraps the batched S6T01 function into a full-history orchestrator that automatically divides all-time history into year-sized windows and calls `fn_backfill_previous_move_id()` for each window. Removes need for manual period calculation at deployment time.

## Requirements

| ID      | Requirement                                                                         |
| ------- | ----------------------------------------------------------------------------------- |
| REQ-710 | Function created in `stats` schema with signature `(INT)` RETURNS TEXT              |
| REQ-711 | Default `p_batch_size = 50000`                                                      |
| REQ-712 | Scans `geokrety.gk_moves` to find earliest move with `previous_move_id IS NULL`     |
| REQ-713 | Iterates in annual windows from earliest year to current year                       |
| REQ-714 | Each window delegates to `stats.fn_backfill_previous_move_id(window, p_batch_size)` |
| REQ-715 | Returns summary text including per-window results                                   |
| REQ-716 | Idempotent: re-run when table already filled exits immediately without error        |
| REQ-717 | Logs final completion to `stats.job_log`                                            |
| REQ-718 | `phinx rollback` drops function cleanly                                             |

## Migration File

**`20260310600200_create_backfill_heavy_previous_move_id.php`**

---

## pgTAP Unit Tests

| Test ID   | Assertion                                                                       | Expected |
| --------- | ------------------------------------------------------------------------------- | -------- |
| T-6.2.001 | `has_function('stats','fn_backfill_heavy_previous_move_id_all','{int}')`        | pass     |
| T-6.2.002 | Calling the orchestrator fills all NULL `previous_move_id` rows in test dataset | pass     |
| T-6.2.003 | Orchestrator is idempotent (re-run on complete data returns quickly)            | pass     |
| T-6.2.004 | Function logs to `stats.job_log` with status='ok'                               | pass     |
| T-6.2.005 | `phinx rollback` drops function cleanly                                         | pass     |

| Test ID   | Criterion | pgTAP Assertion                        | Pass Condition   |
| --------- | --------- | -------------------------------------- | ---------------- |
| T-6.2.001 | Exists    | `has_function` with `'{int}'` arg      | pass             |
| T-6.2.002 | AC-6.2.1  | NULL previous_move_id count = expected | 0 (minus firsts) |
| T-6.2.003 | AC-6.2.2  | Re-run completes without error         | no exception     |
| T-6.2.004 | REQ-717   | `job_log` row with status='ok'         | 1 row            |
| T-6.2.005 | Rollback  | Function absent after `phinx rollback` | NOT has_function |

## Implementation Checklist

- [ ] 1. Create `20260310600200_create_backfill_heavy_previous_move_id.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify `\df stats.fn_backfill_heavy_previous_move_id_all`
- [ ] 4. Run on test dataset — verify all NULL previous_move_id rows filled
- [ ] 5. Run pgTAP T-6.2.001 through T-6.2.005
- [ ] 6. `phinx rollback` — function dropped

- [ ] Create `20260310600200_create_backfill_heavy_previous_move_id.php`
- [ ] Run `phinx migrate` — no errors
- [ ] Test on sample dataset — verify NULL count drops to expected minimum
- [ ] Run pgTAP T-6.2.001 through T-6.2.005
- [ ] `phinx rollback` — function dropped

## SQL DDL

```sql
CREATE OR REPLACE FUNCTION stats.fn_backfill_heavy_previous_move_id_all(
  p_batch_size INT DEFAULT 50000
)
  RETURNS TEXT LANGUAGE plpgsql SECURITY DEFINER
AS $$
DECLARE
  v_start    TIMESTAMPTZ := clock_timestamp();
  v_earliest TIMESTAMPTZ;
  v_period   TSTZRANGE;
  v_window   INTERVAL := INTERVAL '1 year';
  v_slice_start TIMESTAMPTZ;
  v_slice_end   TIMESTAMPTZ;
  v_result   TEXT;
  v_total    TEXT := '';
BEGIN
  -- Find earliest move date in the database
  SELECT MIN(moved_on_datetime) INTO v_earliest
  FROM geokrety.gk_moves
  WHERE previous_move_id IS NULL;

  IF v_earliest IS NULL THEN
    RETURN 'fn_backfill_heavy_previous_move_id_all: nothing to backfill';
  END IF;

  v_slice_start := DATE_TRUNC('year', v_earliest);

  LOOP
    v_slice_end := v_slice_start + v_window;

    EXIT WHEN v_slice_start > NOW();

    v_period := tstzrange(v_slice_start, v_slice_end);

    RAISE NOTICE 'Backfilling period: %', v_period;

    v_result := stats.fn_backfill_previous_move_id(v_period, p_batch_size);
    v_total  := v_total || E'\n' || v_result;

    v_slice_start := v_slice_end;
  END LOOP;

  INSERT INTO stats.job_log (job_name, rows_affected, duration, status)
  VALUES ('fn_backfill_heavy_previous_move_id_all', NULL, clock_timestamp() - v_start, 'ok');

  RETURN 'fn_backfill_heavy_previous_move_id_all completed in ' ||
         (clock_timestamp() - v_start)::TEXT || ':' || v_total;
END;
$$;
```

## Phinx PHP Migration

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateBackfillHeavyPreviousMoveId extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION stats.fn_backfill_heavy_previous_move_id_all(
  p_batch_size INT DEFAULT 50000
)
  RETURNS TEXT LANGUAGE plpgsql SECURITY DEFINER
AS $func$
/* ... full function body ... */
$func$;
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP FUNCTION IF EXISTS stats.fn_backfill_heavy_previous_move_id_all(INT);
SQL
        );
    }
}
```

## Usage Examples

```sql
-- Full history backfill (default 50k batch)
SELECT stats.fn_backfill_heavy_previous_move_id_all();

-- Smaller batch for production under load
SELECT stats.fn_backfill_heavy_previous_move_id_all(20000);

-- Check progress
SELECT COUNT(*) FROM geokrety.gk_moves WHERE previous_move_id IS NULL;

-- View job log
SELECT * FROM stats.job_log WHERE job_name LIKE '%previous_move_id%' ORDER BY started_at DESC LIMIT 20;
```

## AC-6.2.1 — Full-History Completion

**Given** 5000 historical moves with NULL `previous_move_id`
**When** `fn_backfill_heavy_previous_move_id_all()` runs
**Then** `SELECT COUNT(*) FROM geokrety.gk_moves WHERE previous_move_id IS NULL` = 0 (only first-per-GK remain)

## AC-6.2.2 — Idempotency

**Given** all `previous_move_id` already filled
**When** function runs
**Then** immediate return without error; no corruption

## AC-6.2.3 — Progress Reporting

**When** function runs
**Then** NOTICE messages emitted per annual window processed
