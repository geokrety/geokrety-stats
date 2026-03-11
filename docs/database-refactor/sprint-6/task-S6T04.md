---
title: "Task S6T04: Heavy km Distance Orchestrator"
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
  - km
  - orchestrator
  - specification
  - sprint-6
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S6T03
task: S6T04
step: 6.4
migration: 20260310600300_create_backfill_km_distance_heavy.php
blocks:
  - S6T05
changelog:
  - 2026.03.10: created by merge of task-S6T04.dba.md and task-S6T04.specification.md
---

# Task S6T04: Heavy km Distance Orchestrator

## Sprint Context

- Sprint index: Sprint 6 Task Index
- Tags: database, database-refactor, sprint-6, task-index

## Source

- DBA source: `task-S6T04.dba.md`
- Specification source: `task-S6T04.specification.md`

## Purpose & Scope

Creates `stats.fn_backfill_heavy_km_distance_all(p_batch_size INT DEFAULT 50000)` — year-windowed wrapper around `fn_backfill_km_distance()` for production full-history deployment.

Convenience wrapper for full-history km backfill that splits history into annual windows and calls `fn_backfill_km_distance()` per window. Removes manual period planning at deployment.

## Requirements

| ID      | Requirement                                                                     |
| ------- | ------------------------------------------------------------------------------- |
| REQ-730 | Function in `stats` schema, signature `(INT)` RETURNS TEXT                      |
| REQ-731 | Default batch_size = 50000                                                      |
| REQ-732 | Finds the earliest qualifying move in scope and iterates through full-history windows to now |
| REQ-733 | Each iteration calls `stats.fn_backfill_km_distance(window, batch_size)`        |
| REQ-734 | Returns combined summary output                                                 |
| REQ-735 | Logs final entry to `stats.job_log`                                             |
| REQ-736 | `phinx rollback` drops function cleanly                                         |

## Migration File

**`20260310600400_create_backfill_heavy_km_distance.php`**

---

## pgTAP Unit Tests

| Test ID   | Assertion                                                           | Expected |
| --------- | ------------------------------------------------------------------- | -------- |
| T-6.4.001 | `has_function('stats','fn_backfill_heavy_km_distance_all','{int}')` | pass     |
| T-6.4.002 | Calling orchestrator populates `gk_km_counter` from test dataset    | pass     |
| T-6.4.003 | `job_log` row exists with status='ok'                               | pass     |
| T-6.4.004 | `phinx rollback` drops function cleanly                             | pass     |

| Test ID   | Criterion | pgTAP Assertion                       | Pass Condition   |
| --------- | --------- | ------------------------------------- | ---------------- |
| T-6.4.001 | Exists    | `has_function` with `'{int}'`         | pass             |
| T-6.4.002 | AC-6.4.1  | km_distance populated on test dataset | not empty        |
| T-6.4.003 | REQ-735   | job_log row status='ok'               | 1 row            |
| T-6.4.004 | Rollback  | Function absent after rollback        | NOT has_function |

## Implementation Checklist

- [ ] 1. Create `20260310600400_create_backfill_heavy_km_distance.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Run on test dataset, verify gk_km_counter populated
- [ ] 4. Run pgTAP T-6.4.001 through T-6.4.004
- [ ] 5. `phinx rollback` — function dropped

- [ ] Create `20260310600300_create_backfill_km_distance_heavy.php`
- [ ] Run `phinx migrate` — no errors
- [ ] Test on multi-year test dataset
- [ ] Run pgTAP T-6.4.001 through T-6.4.004
- [ ] `phinx rollback` — function dropped

## SQL DDL

```sql
CREATE OR REPLACE FUNCTION stats.fn_backfill_heavy_km_distance_all(
  p_batch_size INT DEFAULT 50000
)
  RETURNS TEXT LANGUAGE plpgsql SECURITY DEFINER
AS $$
DECLARE
  v_start       TIMESTAMPTZ := clock_timestamp();
  v_earliest    TIMESTAMPTZ;
  v_slice_start TIMESTAMPTZ;
  v_slice_end   TIMESTAMPTZ;
  v_window      INTERVAL := INTERVAL '1 year';
  v_result      TEXT;
  v_total       TEXT := '';
BEGIN
  SELECT MIN(moved_on_datetime) INTO v_earliest
  FROM geokrety.gk_moves
  WHERE move_type IN (0, 3, 5);

  IF v_earliest IS NULL THEN
    RETURN 'fn_backfill_heavy_km_distance_all: no qualifying moves found';
  END IF;

  v_slice_start := DATE_TRUNC('year', v_earliest);

  LOOP
    EXIT WHEN v_slice_start > NOW();

    v_slice_end := v_slice_start + v_window;

    RAISE NOTICE 'km distance backfill: processing %', tstzrange(v_slice_start, v_slice_end);

    v_result     := stats.fn_backfill_km_distance(tstzrange(v_slice_start, v_slice_end), p_batch_size);
    v_total      := v_total || E'\n' || v_result;

    v_slice_start := v_slice_end;
  END LOOP;

  INSERT INTO stats.job_log (job_name, rows_affected, duration, status)
  VALUES ('fn_backfill_heavy_km_distance_all', NULL, clock_timestamp() - v_start, 'ok');

  RETURN 'fn_backfill_heavy_km_distance_all done in ' ||
         (clock_timestamp() - v_start)::TEXT || ':' || v_total;
END;
$$;
```

## Phinx PHP Migration

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateBackfillHeavyKmDistance extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION stats.fn_backfill_heavy_km_distance_all(
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
DROP FUNCTION IF EXISTS stats.fn_backfill_heavy_km_distance_all(INT);
SQL
        );
    }
}
```

## Usage Examples

```sql
-- Full-history km backfill
SELECT stats.fn_backfill_heavy_km_distance_all();

-- Check top GKs by km
SELECT geokret_id, total_km FROM stats.gk_km_counter ORDER BY total_km DESC LIMIT 10;

-- Check job log
SELECT * FROM stats.job_log WHERE job_name LIKE '%km%' ORDER BY started_at DESC LIMIT 5;
```

## Master-Spec Alignment

This task is governed by [../00-SPEC-DRAFT-v1.obsolete.md](../00-SPEC-DRAFT-v1.obsolete.md), Section 8.2.

- The heavy wrapper orchestrates full-history backfill of `geokrety.gk_moves.km_distance` via the canonical batched function.
- Acceptance criteria must validate coverage of `km_distance` on `geokrety.gk_moves`, not population of off-spec counter tables.
- `stats.job_log` references in this task must use the canonical column set only.

## AC-6.4.1 — Full Coverage

**Given** km-contributing moves across multiple years
**When** `fn_backfill_heavy_km_distance_all()` called
**Then** qualifying `geokrety.gk_moves` rows have `km_distance` populated for the full historical range

## AC-6.4.2 — No-Data Early Exit

**Given** no ko_moves with `move_type IN (0,3,5)` exist
**When** function called
**Then** returns informational string, no error
