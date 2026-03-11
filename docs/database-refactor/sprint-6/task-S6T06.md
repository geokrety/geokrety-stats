---
title: "Task S6T06: Stats UC Views"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 6
tags:
  - api
  - dashboard
  - database
  - database-refactor
  - dba
  - specification
  - sprint-6
  - sql
  - stats
  - task-index
  - task-merge
  - uc
  - view
  - views
depends_on:
  - S6T05
  - S1-5
task: S6T06
step: 6.6
migration: 20260310600500_create_stats_views.php
blocks:
  - S6T07
changelog:
  - 2026.03.10: created by merge of task-S6T06.dba.md and task-S6T06.specification.md
  - 2026.03.10: resolved Q-039 by canonizing the Sprint 6 stats UC view catalog
---

# Task S6T06: Stats UC Views

## Sprint Context

- Sprint index: Sprint 6 Task Index
- Tags: database, database-refactor, sprint-6, task-index

## Source

- DBA source: `task-S6T06.dba.md`
- Specification source: `task-S6T06.specification.md`

## Resolved Decision

- S6T06 owns the canonical 12-view stats UC catalog.
- The canonical migration name is `20260310600500_create_stats_views.php`.
- The sanctioned stats-scope catalog is:
  - `stats.v_uc1_country_activity`
  - `stats.v_uc2_user_network`
  - `stats.v_uc3_gk_circulation`
  - `stats.v_uc4_user_continent_coverage`
  - `stats.v_uc6_dormancy`
  - `stats.v_uc7_country_flow`
  - `stats.v_uc8_seasonal_heatmap`
  - `stats.v_uc9_multiplier_velocity`
  - `stats.v_uc10_cache_popularity`
  - `stats.v_uc13_gk_timeline`
  - `stats.v_uc14_first_finder_hof`
  - `stats.v_uc15_distance_records`
- Achievement views UC5 and UC11 remain outside `stats` scope.
- Noncanonical merged entries such as UC12 and UC16 are deleted from this task.

## Purpose & Scope

Creates the canonical stats-facing read views used by downstream API and reporting consumers.

This task owns the stats-scope UC catalog only. It does not redefine achievement views or alternate legacy dashboards.

## Canonical UC View Catalog

| View | Canonical Focus |
| ---- | --------------- |
| `v_uc1_country_activity` | country-level movement and km activity |
| `v_uc2_user_network` | precomputed social/relation network |
| `v_uc3_gk_circulation` | GK circulation and interaction rollups |
| `v_uc4_user_continent_coverage` | user-country activity rolled to continent |
| `v_uc6_dormancy` | dormant GK state |
| `v_uc7_country_flow` | cross-country movement flows |
| `v_uc8_seasonal_heatmap` | hourly/date activity heatmap |
| `v_uc9_multiplier_velocity` | multiplier change velocity from `points.gk_multiplier_audit` |
| `v_uc10_cache_popularity` | waypoint/cache popularity |
| `v_uc13_gk_timeline` | GK event timeline |
| `v_uc14_first_finder_hof` | first-finder hall-of-fame summary |
| `v_uc15_distance_records` | distance-record views over canonical km data |

Removed as stale merge residue:

- `v_uc1_global_kpi`
- `v_uc2_recent_moves`
- `v_uc3_user_leaderboard_km`
- `v_uc4_gk_leaderboard_km`
- `v_uc6_country_stats`
- `v_uc7_gk_country_history`
- `v_uc8_user_country_stats`
- `v_uc12_gk_milestones`
- `v_uc13_first_finders`
- `v_uc14_country_pair_flows`
- `v_uc15_hourly_activity`
- `v_uc16_user_related_gks`

## Requirements

| ID      | Description                                                                 | MoSCoW |
| ------- | --------------------------------------------------------------------------- | ------ |
| REQ-760 | All 12 canonical stats UC views above are created in `stats` schema         | MUST   |
| REQ-761 | `v_uc2_user_network` reads from precomputed relationship tables, not raw `gk_moves` scans | MUST |
| REQ-762 | `v_uc9_multiplier_velocity` remains in stats scope and may read from `points.gk_multiplier_audit` | MUST |
| REQ-763 | Views use `CREATE OR REPLACE` for idempotent deployment                      | MUST   |
| REQ-764 | Rollback drops only the canonical 12-view catalog                           | MUST   |
| REQ-765 | Stale UC12/UC16 and alternate legacy names are explicitly noncanonical      | MUST   |

## Acceptance Criteria

| #   | Criterion                                  | How to Verify |
| --- | ------------------------------------------ | ------------- |
| 1   | Canonical 12-view set exists               | `\dv stats.v_uc*` |
| 2   | `v_uc2_user_network` uses relation tables  | `EXPLAIN` plan review |
| 3   | `v_uc9_multiplier_velocity` exists         | `\dv stats.v_uc9_multiplier_velocity` |
| 4   | Stale merged view names are absent         | search canonical migration body |
| 5   | Rollback drops only canonical stats views  | `phinx rollback` |

## Migration File

**`20260310600500_create_stats_views.php`**

## Full SQL DDL

```sql
CREATE OR REPLACE VIEW stats.v_uc1_country_activity AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc2_user_network AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc3_gk_circulation AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc4_user_continent_coverage AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc6_dormancy AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc7_country_flow AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc8_seasonal_heatmap AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc9_multiplier_velocity AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc10_cache_popularity AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc13_gk_timeline AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc14_first_finder_hof AS SELECT ...;
CREATE OR REPLACE VIEW stats.v_uc15_distance_records AS SELECT ...;
```

## Canonical Notes

- The 12-view catalog above is the only sanctioned stats UC catalog for Sprint 6.
- UC5 and UC11 remain out of scope because they belong to achievements.
- Legacy alternate names shown in merged drafts are deleted rather than aliased.

## Implementation Checklist

- [ ] 1. Create `20260310600500_create_stats_views.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify the canonical 12-view catalog via `\dv stats.v_uc*`
- [ ] 4. Confirm stale UC12 and UC16 entries are absent
- [ ] 5. Run representative `SELECT ... LIMIT 1` and `EXPLAIN` checks
- [ ] 6. `phinx rollback` — canonical views dropped

## Agent Loop Log

- 2026-03-10T21:05:00Z — `dba`: removed obsolete view names from the merged catalog and restored the Sprint 6 migration filename.
- 2026-03-10T21:05:00Z — `critical-thinking`: treated the Section 10-style UC numbering as authoritative and deleted drifted UC12/UC16 additions.
- 2026-03-10T21:05:00Z — `specification`: canonized the 12-view stats UC catalog and documented the removed legacy names.

## Resolution

Q-039 is resolved by canonizing the 12-view stats UC catalog in S6T06.
