---
title: "Task S6T03: Batched km Distance Backfill Function"
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
  - distance
  - function
  - km
  - specification
  - sprint-6
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S6T01
  - S2T05
task: S6T03
step: 6.3
migration: 20260310600200_create_backfill_km_distance.php
blocks:
  - S6T04
  - S6T05
changelog:
  - 2026-03-10: created by merge of task-S6T03.dba.md and task-S6T03.specification.md
  - 2026-03-10: resolved Q-037 by narrowing S6T03 to km_distance backfill only
---

# Task S6T03: Batched km Distance Backfill Function

## Sprint Context

- Sprint index: Sprint 6 Task Index
- Tags: database, database-refactor, sprint-6, task-index

## Source

- DBA source: `task-S6T03.dba.md`
- Specification source: `task-S6T03.specification.md`

## Resolved Decision

- S6T03 backfills only `geokrety.gk_moves.km_distance`.
- The canonical migration name is `20260310600200_create_backfill_km_distance.php`.
- Any lower text referencing `stats.gk_km_counter` or `stats.user_km_counter` as S6T03 targets is obsolete.
- Historical aggregate tables are rebuilt elsewhere from canonical move history; they are not direct outputs of this task.

## Purpose & Scope

Creates `stats.fn_backfill_km_distance(p_period TSTZRANGE DEFAULT tstzrange('-infinity', 'infinity'), p_batch_size INT DEFAULT 50000)` to populate historical `geokrety.gk_moves.km_distance` values from canonical predecessor links.

This task owns only row-level distance backfill on `geokrety.gk_moves`.

## Requirements

| ID      | Description                                                                                  | MoSCoW |
| ------- | -------------------------------------------------------------------------------------------- | ------ |
| REQ-720 | Function `stats.fn_backfill_km_distance(TSTZRANGE, INT)` exists                              | MUST   |
| REQ-721 | Only qualifying kilometer-counting moves with a valid predecessor can receive `km_distance`  | MUST   |
| REQ-722 | Rows without a valid predecessor keep `km_distance` as `NULL`                                | MUST   |
| REQ-723 | Re-running the helper safely recomputes `km_distance` for the selected period                | MUST   |
| REQ-724 | Completion writes a canonical `stats.job_log` row using only `job_name`, `status`, `metadata`, `started_at`, `completed_at` | MUST   |
| REQ-725 | `metadata` includes at least `period`, `batch_size`, and `rows_updated`                      | MUST   |
| REQ-726 | No direct writes to legacy km counter tables occur in this task                              | MUST   |
| REQ-727 | `phinx rollback` drops the function cleanly                                                  | MUST   |

## Acceptance Criteria

| #   | Criterion                                           | How to Verify |
| --- | --------------------------------------------------- | ------------- |
| 1   | Function exists in `stats` schema                   | `\df stats.fn_backfill_km_distance` |
| 2   | Qualifying move with predecessor gets `km_distance` | Seed known coordinates and compare result |
| 3   | First locatable move keeps `km_distance` as `NULL`  | Verify sample row |
| 4   | Re-run is idempotent for selected period            | Recompute and confirm stable values |
| 5   | `stats.job_log` row uses canonical columns only     | Inspect log row |
| 6   | Rollback removes function                           | `phinx rollback` |

## Migration File

**`20260310600200_create_backfill_km_distance.php`**

## Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION stats.fn_backfill_km_distance(
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
  -- 1. Batch qualifying moves inside the requested period.
  -- 2. Use previous_move_id-derived source and canonical move coordinates.
  -- 3. Set km_distance only for qualifying rows with a valid predecessor.

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES (
    'fn_backfill_km_distance',
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

final class CreateBackfillKmDistance extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION stats.fn_backfill_km_distance(
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
DROP FUNCTION IF EXISTS stats.fn_backfill_km_distance(TSTZRANGE, INT);
SQL
        );
    }
}
```

The placeholder body above must be replaced by the canonical `km_distance` recomputation logic from REQ-721 through REQ-726.

## Canonical Notes

- S6T03 does not seed `stats.gk_km_counter` or `stats.user_km_counter`.
- Any validation here must target `geokrety.gk_moves.km_distance` directly.
- Full-history orchestration belongs to S6T04.

## pgTAP Unit Tests

| Test ID   | Assertion                                                 | Pass Condition |
| --------- | --------------------------------------------------------- | -------------- |
| T-6.3.001 | `stats.fn_backfill_km_distance()` exists                  | `has_function()` |
| T-6.3.002 | Qualifying move receives expected `km_distance`           | exact match |
| T-6.3.003 | First locatable move keeps `km_distance IS NULL`          | exact match |
| T-6.3.004 | Non-qualifying move does not receive `km_distance`        | exact match |
| T-6.3.005 | Re-run preserves correct values                           | exact match |
| T-6.3.006 | `stats.job_log` row uses canonical fields only            | pass |
| T-6.3.007 | Rollback removes function                                 | pass |

## Implementation Checklist

- [ ] 1. Create `20260310600200_create_backfill_km_distance.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify `\df stats.fn_backfill_km_distance`
- [ ] 4. Validate sample `km_distance` recomputation on known moves
- [ ] 5. Verify no off-spec counter-table dependency remains
- [ ] 6. Run pgTAP T-6.3.001 through T-6.3.007
- [ ] 7. `phinx rollback` — function dropped

## Agent Loop Log

- 2026-03-10T21:05:00Z — `dba`: removed stale km-counter-table ownership from S6T03 and restored the correct migration file.
- 2026-03-10T21:05:00Z — `critical-thinking`: separated row-level distance recomputation from aggregate-table rebuild responsibilities.
- 2026-03-10T21:05:00Z — `specification`: canonized S6T03 as a `geokrety.gk_moves.km_distance` backfill task only.

## Resolution

Q-037 is resolved by canonizing S6T03 as the `km_distance` backfill task only.
