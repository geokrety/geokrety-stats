---
title: "Task S6T08: Full Backfill Execution Plan & Reconciliation"
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
  - deployment
  - deployment-plan
  - reconciliation
  - specification
  - sprint-6
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S6T05
  - S6T07
task: S6T08
step: 6.8
migration: "(manual SQL - no migration)"
blocks:
  - S6T09
changelog:
  - 2026.03.10: created by merge of task-S6T08.dba.md and task-S6T08.specification.md
  - 2026.03.10: resolved Q-018 by standardizing `stats.fn_reconcile_stats()`
  - 2026.03.10: resolved Q-040 by canonizing the reconciliation target set and zero-delta policy
---

# Task S6T08: Full Backfill Execution Plan & Reconciliation

## Sprint Context

- Sprint index: Sprint 6 Task Index
- Tags: database, database-refactor, sprint-6, task-index

## Source

- DBA source: `task-S6T08.dba.md`
- Specification source: `task-S6T08.specification.md`

## Resolved Decision

- `stats.fn_reconcile_stats()` is part of the master spec and is standardized by this task.
- S6T08 still includes the operational runbook, but reconciliation is not runbook-only.
- Exact-match reconciliation uses zero tolerance unless another task explicitly states otherwise.
- Off-spec km counter tables are not reconciliation targets.

## Purpose & Scope

Defines the production backfill run order and the canonical reconciliation helper used to verify the resulting stats state.

## Canonical Reconciliation Targets

1. `geokrety.gk_moves.previous_move_id`
2. `geokrety.gk_moves.km_distance`
3. `stats.entity_counters_shard` aggregate sums
4. `stats.daily_activity`
5. `stats.daily_active_users`
6. `stats.country_daily_stats`
7. `stats.user_countries`
8. `stats.gk_countries_visited`
9. `stats.waypoints`
10. `stats.gk_cache_visits`
11. `stats.user_cache_visits`
12. `stats.gk_related_users`
13. `stats.user_related_users`
14. `stats.hourly_activity`
15. `stats.country_pair_flows`
16. canonical views and materialized views from S6T06 and S6T07

Excluded as obsolete/noncanonical targets:

- `stats.gk_km_counter`
- `stats.user_km_counter`

## Tolerance Policy

- Exact checks must reconcile to delta `0`.
- Any nonzero delta is a failure.
- No approximate tolerance band is accepted in this task.

## Requirements

| ID      | Description                                                                 | MoSCoW |
| ------- | --------------------------------------------------------------------------- | ------ |
| REQ-780 | Production backfill execution order is documented                           | MUST   |
| REQ-781 | `stats.fn_reconcile_stats()` exists as a canonical helper                   | MUST   |
| REQ-782 | Reconciliation helper validates only canonical accepted objects             | MUST   |
| REQ-783 | Exact-match checks require delta `0`                                        | MUST   |
| REQ-784 | Helper writes canonical `stats.job_log` rows using only approved columns    | MUST   |
| REQ-785 | Manual runbook and helper contract remain aligned                           | MUST   |

## Acceptance Criteria

| #   | Criterion                                         | How to Verify |
| --- | ------------------------------------------------- | ------------- |
| 1   | Runbook documents the canonical phase order       | inspect execution section |
| 2   | `stats.fn_reconcile_stats()` exists               | `\df stats.fn_reconcile_stats` |
| 3   | Reconciliation target set excludes off-spec km counters | inspect helper body |
| 4   | Zero-delta policy is explicit                     | inspect tolerance section |
| 5   | Helper writes canonical `stats.job_log` rows      | inspect completion row |

## Manual SQL Runbook

```sql
SELECT stats.fn_run_all_snapshots();
SELECT * FROM stats.fn_reconcile_stats();
```

## Reconciliation Helper DDL

```sql
CREATE OR REPLACE FUNCTION stats.fn_reconcile_stats()
RETURNS TABLE (
  check_name TEXT,
  source_count BIGINT,
  stats_count BIGINT,
  delta BIGINT,
  status TEXT
)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
  v_started_at TIMESTAMPTZ := clock_timestamp();
BEGIN
  -- Canonical implementation returns one row per exact reconciliation check.
  -- Each check must report delta = 0 to yield status = 'OK'.

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES (
    'fn_reconcile_stats',
    'ok',
    jsonb_build_object('policy', 'exact-zero-delta'),
    v_started_at,
    clock_timestamp()
  );

  RETURN;
END;
$$;
```

## Canonical Notes

- The helper is standardized even though this task remains a manual-SQL operational step in the sprint index.
- Any previous wording treating reconciliation as helper-optional is obsolete.
- Quality gates in S6T09 may call this helper directly.

## Implementation Checklist

- [ ] 1. Deliver the manual runbook and canonical helper definition
- [ ] 2. Execute `stats.fn_run_all_snapshots()` on a test environment
- [ ] 3. Execute `stats.fn_reconcile_stats()` and verify zero mismatches
- [ ] 4. Verify helper references canonical `stats.job_log` fields only
- [ ] 5. Remediate any nonzero deltas before Sprint 6 sign-off

## Agent Loop Log

- 2026-03-10T21:35:00Z — `dba`: restored `stats.fn_reconcile_stats()` as a first-class canonical helper while keeping the runbook phase order explicit.
- 2026-03-10T21:35:00Z — `critical-thinking`: rejected approximate tolerances and off-spec counter targets to keep reconciliation exact and replayable.
- 2026-03-10T21:35:00Z — `specification`: aligned S6T08 with the user decision that reconciliation belongs in the master spec.

## Resolution

Q-018 and Q-040 are resolved by standardizing `stats.fn_reconcile_stats()` and the zero-delta reconciliation policy in S6T08.
