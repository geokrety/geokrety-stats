---
title: "Task S6T05: Snapshot Orchestration Function"
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
  - orchestration
  - snapshot
  - specification
  - sprint-6
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S2T11
  - S2T12
  - S3T07
  - S4T11
  - S5T09
  - S6T02
  - S6T04
task: S6T05
step: 6.5
migration: 20260310600400_create_snapshot_orchestration.php
blocks:
  - S6T08
changelog:
  - 2026.03.10: created by merge of task-S6T05.dba.md and task-S6T05.specification.md
  - 2026.03.10: resolved Q-038 by canonizing the orchestration helper list and order
---

# Task S6T05: Snapshot Orchestration Function

## Sprint Context

- Sprint index: Sprint 6 Task Index
- Tags: database, database-refactor, sprint-6, task-index

## Source

- DBA source: `task-S6T05.dba.md`
- Specification source: `task-S6T05.specification.md`

## Resolved Decision

- S6T05 owns the single orchestration wrapper for the canonical helper chain.
- The canonical migration name is `20260310600400_create_snapshot_orchestration.php`.
- Stale helper names such as `fn_snapshot_km_counters()`, `fn_snapshot_country_stats()`, `fn_snapshot_milestones()`, and `fn_snapshot_first_finders()` are removed from the orchestration contract.
- Sprint 4 orchestration uses `stats.fn_snapshot_relationship_tables(...)` as the stable wrapper, not bespoke per-helper sequencing inside S6T05.

## Purpose & Scope

Creates `stats.fn_run_all_snapshots()` as the deployment and repair entry point that runs the canonical backfill and snapshot helpers in dependency order.

This wrapper owns orchestration only. It does not redefine the underlying helper contracts.

## Canonical Helper List and Order

The canonical orchestration order is:

1. `stats.fn_backfill_heavy_previous_move_id_all()`
2. `stats.fn_backfill_heavy_km_distance_all()`
3. `stats.fn_snapshot_entity_counters()`
4. `stats.fn_seed_daily_activity()`
5. `stats.fn_snapshot_daily_country_stats()`
6. `stats.fn_snapshot_user_country_stats()`
7. `stats.fn_snapshot_gk_country_stats()`
8. `stats.fn_snapshot_relationship_tables(p_period daterange DEFAULT NULL)`
9. `stats.fn_snapshot_hourly_activity()`
10. `stats.fn_snapshot_country_pair_flows()`

Removed from the canonical order:

- `stats.fn_snapshot_km_counters()`
- `stats.fn_snapshot_country_stats()`
- `stats.fn_snapshot_milestones()`
- `stats.fn_snapshot_first_finders()`
- bespoke Sprint 4 direct wrapper logic that bypasses `stats.fn_snapshot_relationship_tables(...)`

## Requirements

| ID      | Description                                                                                                      | MoSCoW |
| ------- | ---------------------------------------------------------------------------------------------------------------- | ------ |
| REQ-740 | Function `stats.fn_run_all_snapshots()` exists in `stats` schema                                                 | MUST   |
| REQ-741 | Execution order follows the canonical helper list above                                                          | MUST   |
| REQ-742 | Waypoint/cache/relation rebuild is invoked via `stats.fn_snapshot_relationship_tables(...)`                     | MUST   |
| REQ-743 | Sprint 5 batch helpers run only after prerequisite backfills and snapshots complete                              | MUST   |
| REQ-744 | Completion writes a canonical `stats.job_log` row using only `job_name`, `status`, `metadata`, `started_at`, `completed_at` | MUST   |
| REQ-745 | Wrapper returns summary text or structured phase summary without depending on removed job_log columns            | MUST   |
| REQ-746 | `phinx rollback` drops the wrapper cleanly                                                                       | MUST   |

## Acceptance Criteria

| #   | Criterion                                           | How to Verify |
| --- | --------------------------------------------------- | ------------- |
| 1   | Wrapper exists                                      | `\df stats.fn_run_all_snapshots` |
| 2   | Phase order matches canonical list                  | Inspect wrapper body |
| 3   | Sprint 4 orchestration uses wrapper helper          | Search for `fn_snapshot_relationship_tables` |
| 4   | No stale helper names remain in wrapper             | Search for removed names |
| 5   | `stats.job_log` row uses canonical fields only      | Inspect completion row |
| 6   | Rollback removes wrapper                            | `phinx rollback` |

## Migration File

**`20260310600400_create_snapshot_orchestration.php`**

## Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION stats.fn_run_all_snapshots()
RETURNS TEXT
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
  v_started_at TIMESTAMPTZ := clock_timestamp();
BEGIN
  PERFORM stats.fn_backfill_heavy_previous_move_id_all();
  PERFORM stats.fn_backfill_heavy_km_distance_all();
  PERFORM stats.fn_snapshot_entity_counters();
  PERFORM stats.fn_seed_daily_activity();
  PERFORM stats.fn_snapshot_daily_country_stats();
  PERFORM stats.fn_snapshot_user_country_stats();
  PERFORM stats.fn_snapshot_gk_country_stats();
  PERFORM stats.fn_snapshot_relationship_tables();
  PERFORM stats.fn_snapshot_hourly_activity();
  PERFORM stats.fn_snapshot_country_pair_flows();

  INSERT INTO stats.job_log (job_name, status, metadata, started_at, completed_at)
  VALUES (
    'fn_run_all_snapshots',
    'ok',
    jsonb_build_object(
      'phases', jsonb_build_array(
        'fn_backfill_heavy_previous_move_id_all',
        'fn_backfill_heavy_km_distance_all',
        'fn_snapshot_entity_counters',
        'fn_seed_daily_activity',
        'fn_snapshot_daily_country_stats',
        'fn_snapshot_user_country_stats',
        'fn_snapshot_gk_country_stats',
        'fn_snapshot_relationship_tables',
        'fn_snapshot_hourly_activity',
        'fn_snapshot_country_pair_flows'
      )
    ),
    v_started_at,
    clock_timestamp()
  );

  RETURN 'fn_run_all_snapshots completed';
END;
$$;
```

## Canonical Notes

- This task is about orchestration, not alternative helper definitions.
- Milestone and first-finder logic are live-trigger concerns from Sprint 5 and do not belong in the batch orchestration chain.
- Country snapshot seeding uses the three canonical Sprint 3 helpers, not a collapsed stale alias.

## Implementation Checklist

- [ ] 1. Create `20260310600400_create_snapshot_orchestration.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify `\df stats.fn_run_all_snapshots`
- [ ] 4. Confirm canonical helper order in the wrapper body
- [ ] 5. Verify canonical `stats.job_log` metadata payload
- [ ] 6. Run orchestration on sample dataset
- [ ] 7. `phinx rollback` — wrapper dropped

## Agent Loop Log

- 2026-03-10T21:05:00Z — `dba`: removed stale helper names and restored the correct Sprint 6 migration number.
- 2026-03-10T21:05:00Z — `critical-thinking`: collapsed Sprint 4 orchestration to the stable wrapper to avoid dual ownership of helper ordering.
- 2026-03-10T21:05:00Z — `specification`: canonized the final 10-step orchestration chain and documented the deleted helper names.

## Resolution

Q-038 is resolved by canonizing the final orchestration helper list and order in S6T05.
