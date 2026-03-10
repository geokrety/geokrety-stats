---
title: "Task S6T07: Materialized Views"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 6
tags:
  - cache
  - dashboard
  - database
  - database-refactor
  - dba
  - materialized-view
  - performance
  - specification
  - sprint-6
  - sql
  - stats
  - task-index
  - task-merge
depends_on:
  - S6T06
task: S6T07
step: 6.7
migration: 20260310600600_create_materialized_views.php
blocks:
  - S6T09
changelog:
  - 2026-03-10: created by merge of task-S6T07.dba.md and task-S6T07.specification.md
  - 2026-03-10: resolved Q-017 by standardizing the materialized-view catalog
  - 2026-03-10: resolved Q-006 by canonizing the refresh schedule owner and cadence
---

# Task S6T07: Materialized Views

## Sprint Context

- Sprint index: Sprint 6 Task Index
- Tags: database, database-refactor, sprint-6, task-index

## Source

- DBA source: `task-S6T07.dba.md`
- Specification source: `task-S6T07.specification.md`

## Resolved Decision

- Materialized views are standardized in the master spec; they are not left open as an undefined optional catalog.
- The canonical Sprint 6 MV catalog contains exactly three objects:
  - `stats.mv_country_month_rollup`
  - `stats.mv_top_caches_global`
  - `stats.mv_global_kpi`
- Refresh is executed by an external scheduler, not by a canonical `pg_cron` dependency.
- `REFRESH MATERIALIZED VIEW CONCURRENTLY` is required for the standardized MV set.

## Purpose & Scope

Creates the canonical materialized-view accelerators for the heaviest dashboard and reporting read paths.

These objects are part of the accepted Sprint 6 spec and must be defined consistently across environments.

## Canonical MV Catalog

1. `stats.mv_country_month_rollup`
2. `stats.mv_top_caches_global`
3. `stats.mv_global_kpi`

## Refresh Schedule

Canonical target cadence:

- `stats.mv_global_kpi`: every 5 minutes
- `stats.mv_top_caches_global`: every 30 minutes
- `stats.mv_country_month_rollup`: every 60 minutes

Canonical refresh owner:

- external scheduler invoking `REFRESH MATERIALIZED VIEW CONCURRENTLY`
- no hard dependency on `pg_cron`

## Requirements

| ID      | Description                                                                 | MoSCoW |
| ------- | --------------------------------------------------------------------------- | ------ |
| REQ-770 | The three-view MV catalog above is standardized and created by this task    | MUST   |
| REQ-771 | Each MV is derived from canonical stats tables or views only                | MUST   |
| REQ-772 | Each MV has the unique index required for concurrent refresh                | MUST   |
| REQ-773 | Refresh cadence and ownership are documented as part of the task            | MUST   |
| REQ-774 | `REFRESH MATERIALIZED VIEW CONCURRENTLY` succeeds for the standardized set  | MUST   |
| REQ-775 | No off-spec km counter tables are required by the standardized MV catalog   | MUST   |
| REQ-776 | `phinx rollback` drops the standardized MV set cleanly                      | MUST   |

## Acceptance Criteria

| #   | Criterion                                         | How to Verify |
| --- | ------------------------------------------------- | ------------- |
| 1   | All 3 standardized MVs exist                      | `\dm stats.*` |
| 2   | Each MV has its required unique index             | inspect `pg_indexes` |
| 3   | Concurrent refresh succeeds for each MV           | run refresh commands |
| 4   | MV definitions use canonical sources only         | inspect DDL |
| 5   | Refresh cadence is documented                     | inspect refresh section |
| 6   | Rollback removes all 3 MVs                        | `phinx rollback` |

## Migration File

**`20260310600600_create_materialized_views.php`**

## Full SQL DDL

```sql
CREATE MATERIALIZED VIEW stats.mv_country_month_rollup AS
SELECT
  from_country,
  to_country,
  year_month,
  move_count,
  unique_gk_count
FROM stats.country_pair_flows
WITH DATA;

CREATE UNIQUE INDEX idx_mv_country_month_rollup_pk
  ON stats.mv_country_month_rollup (from_country, to_country, year_month);

CREATE MATERIALIZED VIEW stats.mv_top_caches_global AS
SELECT
  waypoint_code,
  total_gk_visits,
  distinct_gks
FROM stats.v_uc10_cache_popularity
WITH DATA;

CREATE UNIQUE INDEX idx_mv_top_caches_global_pk
  ON stats.mv_top_caches_global (waypoint_code);

CREATE MATERIALIZED VIEW stats.mv_global_kpi AS
SELECT
  COUNT(*) FILTER (WHERE TRUE) OVER () AS total_geokrety,
  (SELECT COUNT(*) FROM geokrety.gk_moves) AS total_moves,
  (SELECT COALESCE(SUM(km_distance), 0) FROM geokrety.gk_moves WHERE km_distance IS NOT NULL) AS total_km,
  (SELECT COUNT(*) FROM geokrety.gk_users) AS total_users,
  clock_timestamp() AS computed_at
FROM geokrety.gk_geokrety
LIMIT 1
WITH DATA;

CREATE UNIQUE INDEX idx_mv_global_kpi_pk
  ON stats.mv_global_kpi ((1));
```

## Canonical Notes

- The MV catalog is standardized but still derived from canonical stats objects.
- Refresh execution is an Ops concern owned by an external scheduler, using the cadence above.
- `pg_cron` may be used by deployment choice, but it is not the canonical infrastructure dependency.

## Implementation Checklist

- [ ] 1. Create `20260310600600_create_materialized_views.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify all 3 MVs via `\dm stats.*`
- [ ] 4. Verify the three required unique indexes
- [ ] 5. Test `REFRESH MATERIALIZED VIEW CONCURRENTLY` on all 3 MVs
- [ ] 6. Configure the external refresh scheduler with the canonical cadence
- [ ] 7. `phinx rollback` — all 3 MVs dropped

## Agent Loop Log

- 2026-03-10T21:35:00Z — `dba`: replaced the old optional-MV framing with a standardized three-view catalog and canonical refresh schedule.
- 2026-03-10T21:35:00Z — `critical-thinking`: kept refresh ownership outside PostgreSQL extension requirements while still standardizing the objects and cadence.
- 2026-03-10T21:35:00Z — `specification`: aligned S6T07 with the user decision that the MV catalog belongs in the master spec.

## Resolution

Q-017 is resolved by standardizing the Sprint 6 three-view materialized-view catalog.
