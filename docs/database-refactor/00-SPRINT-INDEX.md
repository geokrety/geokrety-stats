---
title: 'Stats Schema Implementation Plan — Sprint Index'
version: 1.0
date_created: 2026-03-08
last_updated: 2026-03-08
owner: GeoKrety Community
scope: database-refactor
tags:
  - database
  - postgresql
  - stats
  - implementation-plan
  - sprint-index
---

# Database Refactor Implementation Plan — Sprint Index

> **Scope:** This index tracks the database-refactor sprint documents in this folder and must stay aligned with the full canonical scope in [00-SPEC-DRAFT-v1.obsolete.md](00-SPEC-DRAFT-v1.obsolete.md), including cross-schema dependencies.

## Overview

This document is the master index for the GeoKrety database refactoring sprint plan maintained in this folder. Each sprint is a self-contained specification file with numbered steps, full SQL DDL, Phinx migration code, trigger/function implementations, TimescaleDB assessments, pgTAP test tables, graph specifications, and checklists.

## Reference Documents

| Document             | Path                                                                                                                           | Purpose                                                     |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------- |
| Schema Specification | [00-SPEC-DRAFT-v1.obsolete.md](00-SPEC-DRAFT-v1.obsolete.md)                                                                                     | Canonical stats/points/achievements draft specification |
| Gamification Rules   | [../../.github/instructions/gamification-rules.instructions.md](../../.github/instructions/gamification-rules.instructions.md) | Authoritative points/multiplier/chain rules                 |
| Open Questions       | [99-OPEN-QUESTIONS.md](99-OPEN-QUESTIONS.md)                                                                                   | Deferred decisions and questions for future phases          |

## Sprint Dependency Graph

```
Sprint 1 ─── Foundation & Source Table Preparation
   │
   ├──► Sprint 2 ─── Sharded Counters, Daily Activity & Previous-Move Trigger
   │       │
   │       ├──► Sprint 3 ─── Country, Geography & Traversal
   │       │       │
   │       │       └──► Sprint 5 ─── Advanced Analytics, Events & AMQP
   │       │               │
   │       │               └──► Sprint 6 ─── Backfill, Views & Data Migration
   │       │
   │       └──► Sprint 4 ─── Waypoints, Cache Analytics & Social Relations
   │               │
   │               └──► Sprint 6 (also depends on Sprint 4)
   │
   └──► Sprint 6 (depends on all previous sprints)
```

## Sprint Summary

### Sprint 1: Foundation & Source Table Preparation ✅

**File:** [01-sprint-1-foundation.md](01-sprint-1-foundation.md)
**Depends on:** nothing
**Blocks:** Sprints 2, 3, 4, 5, 6

| Step | Description                                                        | Migration File                                         |
| ---- | ------------------------------------------------------------------ | ------------------------------------------------------ |
| 1.1  | Revert 5 preliminary stats migrations                              | `20260310100000_revert_preliminary_stats.php`          |
| 1.2  | Create `stats` schema                                              | `20260310100100_create_stats_schema.php`               |
| 1.3  | Create operational support tables (`backfill_progress`, `job_log`) | `20260310100200_create_operational_support_tables.php` |
| 1.4  | Create continent reference table + seed 249 countries              | `20260310100300_create_continent_reference.php`        |
| 1.5  | Add `previous_move_id` + `km_distance` columns to `gk_moves`       | `20260310100400_add_gk_moves_source_columns.php`       |
| 1.6  | Create 5 source table indexes (CONCURRENTLY)                       | `20260310100500_create_source_table_indexes.php`       |
| 1.7  | Enable `btree_gist` extension                                      | `20260310100600_enable_btree_gist.php`                 |

**Tests:** 62 pgTAP assertions (T-1.1.001 — T-1.7.002)
**Migrations:** 7 files

---

### Sprint 2: Sharded Counters, Daily Activity & Previous-Move Trigger

**File:** [02-sprint-2-counters-daily-activity.md](02-sprint-2-counters-daily-activity.md)
**Depends on:** Sprint 1
**Blocks:** Sprints 3, 4, 5, 6

| Step | Description                                                 | Migration File                                          |
| ---- | ----------------------------------------------------------- | ------------------------------------------------------- |
| 2.1  | Create `stats.entity_counters_shard` table                  | `20260310200000_create_entity_counters_shard.php`       |
| 2.2  | Create `stats.daily_activity` table                         | `20260310200100_create_daily_activity.php`              |
| 2.3  | Create `stats.daily_active_users` table                     | `20260310200200_create_daily_active_users.php`          |
| 2.4  | Create `stats.daily_entity_counts` table                    | `20260310200300_create_daily_entity_counts.php`         |
| 2.5  | Create previous-move trigger function + attach              | `20260310200400_create_previous_move_trigger.php`       |
| 2.6  | Create `gk_moves` sharded counter trigger function + attach | `20260310200500_create_gk_moves_counter_trigger.php`    |
| 2.7  | Create `gk_moves` daily activity trigger function + attach  | `20260310200600_create_gk_moves_daily_trigger.php`      |
| 2.8  | Create `gk_geokrety` counter trigger function + attach      | `20260310200700_create_gk_geokrety_counter_trigger.php` |
| 2.9  | Create `gk_pictures` counter trigger function + attach      | `20260310200800_create_gk_pictures_counter_trigger.php` |
| 2.10 | Create `gk_users` counter trigger function + attach         | `20260310200900_create_gk_users_counter_trigger.php`    |
| 2.11 | Create entity counter snapshot function                     | `20260310201000_create_entity_counter_snapshot.php`     |
| 2.12 | Create daily activity seed function                         | `20260310201100_create_daily_activity_seed.php`         |

**Tests:** ~80 pgTAP assertions
**Migrations:** 12 files

---

### Sprint 3: Country, Geography & Traversal

**File:** [03-sprint-3-country-geography.md](03-sprint-3-country-geography.md)
**Depends on:** Sprint 1, Sprint 2
**Blocks:** Sprint 5, Sprint 6

| Step | Description                                                    | Migration File                                         |
| ---- | -------------------------------------------------------------- | ------------------------------------------------------ |
| 3.1  | Create `stats.country_daily_stats` table                       | `20260310300000_create_country_daily_stats.php`        |
| 3.2  | Create `stats.gk_countries_visited` table                      | `20260310300100_create_gk_countries_visited.php`       |
| 3.3  | Create `stats.user_countries` table                            | `20260310300200_create_user_countries.php`             |
| 3.4  | Create `stats.gk_country_history` table (exclusion constraint) | `20260310300300_create_gk_country_history.php`         |
| 3.5  | Create country rollups trigger function + attach               | `20260310300400_create_country_rollups_trigger.php`    |
| 3.6  | Create country history trigger function + attach               | `20260310300500_create_country_history_trigger.php`    |
| 3.7  | Create country snapshot/seed functions                         | `20260310300600_create_country_snapshot_functions.php` |
| 3.8  | Create country indexes                                         | `20260310300700_create_country_indexes.php`            |

**Tests:** ~65 pgTAP assertions
**Migrations:** 8 files

---

### Sprint 4: Waypoints, Cache Analytics & Social Relations

**File:** [sprint-4/00-index.md](sprint-4/00-index.md)
**Depends on:** Sprint 1, Sprint 2
**Blocks:** Sprint 6

| Step | Description                                                        | Migration File                                          |
| ---- | ------------------------------------------------------------------ | ------------------------------------------------------- |
| 4.1  | Create `stats.waypoints` table                                     | `20260310400000_create_waypoints.php`                   |
| 4.2  | Create `stats.v_waypoints_source_union` view                       | `20260310400100_create_waypoints_source_view.php`       |
| 4.3  | Seed waypoints from GC/OC sources                                  | `20260310400200_seed_waypoints.php`                     |
| 4.4  | Create `stats.gk_cache_visits` table                               | `20260310400300_create_gk_cache_visits.php`             |
| 4.5  | Create `stats.user_cache_visits` table                             | `20260310400400_create_user_cache_visits.php`           |
| 4.6  | Create `stats.gk_related_users` table                              | `20260310400500_create_gk_related_users.php`            |
| 4.7  | Create `stats.user_related_users` table                            | `20260310400600_create_user_related_users.php`          |
| 4.8  | Create waypoint resolution + cache visit trigger function + attach | `20260310400700_create_waypoint_cache_trigger.php`      |
| 4.9  | Create social relation trigger function + attach                   | `20260310400800_create_relation_trigger.php`            |
| 4.10 | Create waypoint/cache/relation indexes                             | `20260310400900_create_cache_relation_indexes.php`      |
| 4.11 | Create snapshot functions for waypoints/relations                  | `20260310401000_create_waypoint_relation_snapshots.php` |

**Tests:** ~70 pgTAP assertions
**Migrations:** 11 files

---

### Sprint 5: Advanced Analytics, Events & AMQP

**File:** [sprint-5/00-index.md](sprint-5/00-index.md)
**Depends on:** Sprint 1, Sprint 2, Sprint 3, Sprint 4
**Blocks:** Sprint 6

| Step | Description                                               | Migration File                                          |
| ---- | --------------------------------------------------------- | ------------------------------------------------------- |
| 5.1  | Create `stats.hourly_activity` table                      | `20260310500000_create_hourly_activity.php`             |
| 5.2  | Create `stats.country_pair_flows` table                   | `20260310500100_create_country_pair_flows.php`          |
| 5.3  | Create `stats.gk_milestone_events` table                  | `20260310500200_create_gk_milestone_events.php`         |
| 5.4  | Create `stats.first_finder_events` table                  | `20260310500300_create_first_finder_events.php`         |
| 5.5  | Create `gk_loves` activity trigger function + attach        | `20260310500400_create_gk_loves_counter_trigger.php`   |
| 5.6  | Create AMQP event emission trigger function + attach      | `20260310500500_create_amqp_event_trigger.php`          |
| 5.7  | Create milestone detection trigger function               | `20260310500600_create_milestone_trigger.php`           |
| 5.8  | Create first finder detection trigger function            | `20260310500700_create_first_finder_trigger.php`        |
| 5.9  | Create batch aggregation functions (hourly, country-pair) | `20260310500800_create_batch_aggregation_functions.php` |
| 5.10 | Create analytics indexes                                  | `20260310500900_create_analytics_indexes.php`           |

**Tests:** ~60 pgTAP assertions
**Migrations:** 10 files

---

### Sprint 6: Backfill Helpers, Views & Data Migration

**File:** [sprint-6/00-index.md](sprint-6/00-index.md)
**Depends on:** All previous sprints
**Blocks:** nothing (final sprint for stats)

| Step | Description                                       | Migration File                                           |
| ---- | ------------------------------------------------- | -------------------------------------------------------- |
| 6.1  | Create `fn_backfill_previous_move_id` (batched)   | `20260310600000_create_backfill_previous_move.php`       |
| 6.2  | Create `fn_backfill_heavy_previous_move_id_all`   | `20260310600100_create_backfill_previous_move_heavy.php` |
| 6.3  | Create `fn_backfill_km_distance` (batched)        | `20260310600200_create_backfill_km_distance.php`         |
| 6.4  | Create `fn_backfill_heavy_km_distance_all`        | `20260310600300_create_backfill_km_distance_heavy.php`   |
| 6.5  | Create snapshot ingestion orchestration functions | `20260310600400_create_snapshot_orchestration.php`       |
| 6.6  | Create the canonical 12 stats UC views            | `20260310600500_create_stats_views.php`                  |
| 6.7  | Define optional materialized-view accelerators    | `20260310600600_create_materialized_views.php`           |
| 6.8  | Full backfill execution plan + reconciliation     | (manual SQL — no migration)                              |
| 6.9  | Performance validation + quality gates            | (manual SQL — no migration)                              |

**Tests:** ~55 pgTAP assertions
**Migrations:** 7 files + manual SQL

---

## Totals

| Metric                          | Count |
| ------------------------------- | ----- |
| **Sprints**                     | 6     |
| **Steps**                       | 55    |
| **Migration files**             | 55    |
| **pgTAP assertions**            | ~392  |
| **Stats tables created**        | 18    |
| **Trigger functions**           | 13    |
| **Snapshot/backfill functions** | 12    |
| **Views**                       | 13    |
| **Indexes (stats schema)**      | 12    |
| **Indexes (source table)**      | 5     |

## Migration Timestamp Allocation

| Sprint | Range Start      | Range End        |
| ------ | ---------------- | ---------------- |
| 1      | `20260310100000` | `20260310100600` |
| 2      | `20260310200000` | `20260310201100` |
| 3      | `20260310300000` | `20260310300700` |
| 4      | `20260310400000` | `20260310401000` |
| 5      | `20260310500000` | `20260310500900` |
| 6      | `20260310600000` | `20260310600600` |

## Naming Conventions

| Object Type                        | Pattern                                   | Example                                           |
| ---------------------------------- | ----------------------------------------- | ------------------------------------------------- |
| Schema                             | `stats`                                   | `stats`                                           |
| Table                              | `stats.<descriptive_name>`                | `stats.entity_counters_shard`                     |
| Function                           | `stats.fn_<purpose>`                      | `stats.fn_backfill_previous_move_id`              |
| Trigger function (geokrety schema) | `geokrety.fn_<purpose>`                   | `geokrety.fn_set_previous_move_id_and_distance`   |
| Trigger                            | `tr_<table>_<purpose>`                    | `tr_gk_moves_after_sharded_counters`              |
| Index                              | `idx_<table>_<columns_or_purpose>`        | `idx_country_daily_stats_country_date`            |
| View                               | `stats.v_<uc_number>_<name>`              | `stats.v_uc1_country_activity`                    |
| Migration file                     | `2026031X0Y0Z00_<snake_case_purpose>.php` | `20260310200000_create_entity_counters_shard.php` |
| Test file                          | `test-2XX-<description>.sql`              | `test-210-sprint2-counters-daily.sql`             |

## Test File Number Allocation

| Sprint | Test File Numbers            | Description                                  |
| ------ | ---------------------------- | -------------------------------------------- |
| 1      | `test-200-*` to `test-203-*` | (already allocated: test-200, 201, 202, 203) |
| 2      | `test-210-*` to `test-212-*` | Counters, daily, previous-move               |
| 3      | `test-220-*` to `test-222-*` | Country, geography, traversal                |
| 4      | `test-230-*` to `test-232-*` | Waypoints, cache, relations                  |
| 5      | `test-240-*` to `test-242-*` | Analytics, events, AMQP                      |
| 6      | `test-250-*` to `test-252-*` | Backfill, views, validation                  |

## Quality Gates (per sprint)

Each sprint must pass before the next begins:

- [ ] All migration files created and syntactically valid
- [ ] `phinx migrate` succeeds with no errors
- [ ] All pgTAP assertions pass (`pg_prove`)
- [ ] EXPLAIN plans for key queries confirm index usage
- [ ] No orphaned triggers or functions
- [ ] All tables have correct PKs, FKs, and CHECK constraints
- [ ] Trigger execution order is correct (BEFORE precedes AFTER)
- [ ] `phinx rollback` for each migration succeeds cleanly

## Cross-Sprint Coherence Rules

1. **Column naming:** All columns use snake_case. Counter columns are `cnt` (in sharded counters) or descriptive (`move_count`, `visit_count`).
2. **Timestamp columns:** Always `TIMESTAMPTZ`, never bare `TIMESTAMP`.
3. **Country codes:** Always `CHAR(2)` uppercase ISO 3166-1 alpha-2.
4. **Numeric distances:** Always `NUMERIC(8,3)` for km.
5. **Numeric points:** Always `NUMERIC(X,4)` for money-like values.
6. **Move types:** Always referenced as integers 0-5, never strings.
7. **GK types:** Standard transferable = `{0,1,2,3,4,5,7,9}`, Non-transferable = `{6,8,10}`.
8. **Trigger order on gk_moves:** BEFORE triggers run first (previous-move), then AFTER triggers in alphabetical trigger name order.
9. **Idempotency:** All seed/snapshot functions use `ON CONFLICT ... DO UPDATE` or `DO NOTHING`.
10. **FK constraints to source tables:** Use `DEFERRABLE INITIALLY DEFERRED` where batch operations may violate ordering.
