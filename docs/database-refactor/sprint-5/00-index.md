---
title: 'Sprint 5: Advanced Analytics and Event Bridge — Index'
version: 1.0
date_created: 2026-03-08
last_updated: 2026-03-08
owner: GeoKrety Community
sprint: 5
depends_on: [1, 2, 3, 4]
blocks: [6]
tags:
  - database
  - postgresql
  - stats
  - sprint-5
  - analytics
  - events
  - event-bridge
  - milestones
  - hourly
---

# Sprint 5: Advanced Analytics, Events & AMQP

## Overview

Sprint 5 adds the advanced analytics and event-emission layer. It creates four tables that track hourly activity buckets, country-pair flow matrices, GK milestone events, and first-finder records. Live trigger work includes `gk_loves` activity rollups, the `gk_moves` points-event bridge, and lightweight milestone and first-finder fact capture. Batch aggregation functions fill `hourly_activity` and `country_pair_flows` for historical data. Supporting indexes complete the sprint.

The bridge integration remains abstract at the master-spec level: the trigger writes to an approved event table or queue bridge so that `points-awarder` can consume scoring events asynchronously.

## Master-Spec Alignment

Sprint 5 must remain aligned to [../00-SPEC-DRAFT-v1.md](../00-SPEC-DRAFT-v1.md), Sections 5.7, 8.4, 8.5, 10, and 11.

- Canonical Sprint 5 tables from the master spec are `stats.hourly_activity`, `stats.country_pair_flows`, `stats.gk_milestone_events`, and `stats.first_finder_events`.
- Live trigger work in the master spec includes `gk_loves` activity updates, lightweight milestone and first-finder fact capture, and the `tr_gk_moves_emit_points_event` bridge. `stats.hourly_activity` and `stats.country_pair_flows` are batch/manual computed.
- The master spec allows an event table or queue bridge but does not canonically standardize a concrete `stats.outbox_events` schema.
- Any lower text that removes the Sprint 5 `gk_loves` trigger task, introduces non-canonical bridge names, or assigns live trigger ownership to hourly/country-pair tables is obsolete and superseded by this alignment block.

## Prerequisites

- Sprint 1 completed: `stats` schema, `btree_gist`, source table indexes, operational tables.
- Sprint 2 completed: entity counters, daily activity tables, previous-move trigger.
- Sprint 3 completed: country daily stats, gk_countries_visited, user_countries, gk_country_history.
- Sprint 4 completed: waypoints, cache visits, relation tables.

## Time Estimate

| Phase                        | Effort   |
| ---------------------------- | -------- |
| Schema + DDL authoring       | 2 h      |
| Trigger function authoring   | 4 h      |
| AMQP/outbox design + trigger | 2 h      |
| Phinx migration authoring    | 2 h      |
| pgTAP test authoring         | 2 h      |
| Batch aggregation functions  | 1.5 h    |
| Code review + QA             | 1.5 h    |
| **Total**                    | **15 h** |

## Task Inventory

| Task  | File (DBA)                             | File (Specification)                                       | Step | Description                                                | Migration File                                          |
| ----- | -------------------------------------- | ---------------------------------------------------------- | ---- | ---------------------------------------------------------- | ------------------------------------------------------- |
| S5T01 | [task-S5T01.dba.md](task-S5T01.dba.md) | [task-S5T01.specification.md](task-S5T01.specification.md) | 5.1  | Create `stats.hourly_activity` table                       | `20260310500000_create_hourly_activity.php`             |
| S5T02 | [task-S5T02.dba.md](task-S5T02.dba.md) | [task-S5T02.specification.md](task-S5T02.specification.md) | 5.2  | Create `stats.country_pair_flows` table                    | `20260310500100_create_country_pair_flows.php`          |
| S5T03 | [task-S5T03.dba.md](task-S5T03.dba.md) | [task-S5T03.specification.md](task-S5T03.specification.md) | 5.3  | Create `stats.gk_milestone_events` table                   | `20260310500200_create_gk_milestone_events.php`         |
| S5T04 | [task-S5T04.dba.md](task-S5T04.dba.md) | [task-S5T04.specification.md](task-S5T04.specification.md) | 5.4  | Create `stats.first_finder_events` table                   | `20260310500300_create_first_finder_events.php`         |
| S5T05 | [task-S5T05.dba.md](task-S5T05.dba.md) | [task-S5T05.specification.md](task-S5T05.specification.md) | 5.5  | Create `gk_loves` activity trigger function + attach        | `20260310500400_create_gk_loves_counter_trigger.php`   |
| S5T06 | [task-S5T06.dba.md](task-S5T06.dba.md) | [task-S5T06.specification.md](task-S5T06.specification.md) | 5.6  | Create canonical points-event emission trigger              | `20260310500500_create_amqp_event_trigger.php`          |
| S5T07 | [task-S5T07.dba.md](task-S5T07.dba.md) | [task-S5T07.specification.md](task-S5T07.specification.md) | 5.7  | Create milestone detection trigger function + attach       | `20260310500600_create_milestone_trigger.php`           |
| S5T08 | [task-S5T08.dba.md](task-S5T08.dba.md) | [task-S5T08.specification.md](task-S5T08.specification.md) | 5.8  | Create first finder detection trigger function + attach    | `20260310500700_create_first_finder_trigger.php`        |
| S5T09 | [task-S5T09.dba.md](task-S5T09.dba.md) | [task-S5T09.specification.md](task-S5T09.specification.md) | 5.9  | Create batch aggregation functions (hourly, country-pair)  | `20260310500800_create_batch_aggregation_functions.php` |
| S5T10 | [task-S5T10.dba.md](task-S5T10.dba.md) | [task-S5T10.specification.md](task-S5T10.specification.md) | 5.10 | Create analytics indexes                                   | `20260310500900_create_analytics_indexes.php`           |

## pgTAP Test File Allocation

| Sprint | Test File                             | Description                           |
| ------ | ------------------------------------- | ------------------------------------- |
| 5      | `test-240-hourly-country-flows.sql`   | Hourly activity + country pair flows  |
| 5      | `test-241-milestones-firstfinder.sql` | Milestone events + first finder       |
| 5      | `test-242-event-bridge.sql`           | `gk_loves` + event-bridge trigger tests |

Expected assertions: ~60 pgTAP assertions.

## Migration Timestamp Allocation

| Task  | Timestamp        |
| ----- | ---------------- |
| S5T01 | `20260310500000` |
| S5T02 | `20260310500100` |
| S5T03 | `20260310500200` |
| S5T04 | `20260310500300` |
| S5T05 | `20260310500400` |
| S5T06 | `20260310500500` |
| S5T07 | `20260310500600` |
| S5T08 | `20260310500700` |
| S5T09 | `20260310500800` |
| S5T10 | `20260310500900` |

## Objects Created

| Object                                | Type             | Schema                      |
| ------------------------------------- | ---------------- | --------------------------- |
| `stats.hourly_activity`               | Table            | `stats`                     |
| `stats.country_pair_flows`            | Table            | `stats`                     |
| `stats.gk_milestone_events`           | Table            | `stats`                     |
| `stats.first_finder_events`           | Table            | `stats`                     |
| `geokrety.fn_gk_loves_activity()`     | Trigger function | `geokrety`                  |
| canonical points-event bridge function | Trigger function | `geokrety`                  |
| `geokrety.fn_gk_moves_milestones()`   | Trigger function | `geokrety`                  |
| `geokrety.fn_gk_moves_first_finder()` | Trigger function | `geokrety`                  |
| `tr_gk_loves_activity`                | Trigger          | `geokrety.gk_loves`         |
| `tr_gk_moves_emit_points_event`       | Trigger          | `geokrety.gk_moves`         |
| `tr_gk_moves_after_milestones`        | Trigger          | `geokrety.gk_moves`         |
| `tr_gk_moves_after_first_finder`      | Trigger          | `geokrety.gk_moves`         |
| `stats.fn_snapshot_hourly_activity()`    | Batch function   | `stats`                     |
| `stats.fn_snapshot_country_pair_flows()` | Batch function   | `stats`                     |
| `idx_hourly_activity_date_desc`          | Index            | `stats.hourly_activity`  |
| `idx_country_pair_flows_month_desc`      | Index            | `stats.country_pair_flows` |
| `idx_country_pair_flows_from`            | Index            | `stats.country_pair_flows` |
| `idx_country_pair_flows_to`              | Index            | `stats.country_pair_flows` |
| `idx_gk_milestone_events_gk`          | Index            | `stats.gk_milestone_events` |
| `idx_first_finder_events_user`        | Index            | `stats.first_finder_events` |

## Event Bridge Architecture

The database uses the canonical loves and points trigger bridge layer:

1. `tr_gk_loves_activity` updates love-related rollups within the same transaction as insert, update, and delete operations on `gk_loves`.
2. `tr_gk_moves_emit_points_event` emits a scoring event within the same transaction as inserts on `gk_moves`.
3. The bridge transport may use an event table, `NOTIFY`, or another approved queue bridge.
4. `points-awarder` consumes scoring events from the approved bridge implementation.
5. Delivery lifecycle semantics are implementation-specific and are not part of the canonical stats contract.

## Dependency Graph

```
Sprint 1 → Sprint 2 → Sprint 3 → Sprint 4
                                    └──► Sprint 5 (this sprint)
                                              └──► Sprint 6
```

## Quality Gates

Before marking sprint complete:

- [ ] All 10 migration files created and syntactically valid
- [ ] `phinx migrate` succeeds with no errors
- [ ] All ~60 pgTAP assertions pass
- [ ] `EXPLAIN` for hourly heatmap uses `stats.hourly_activity`
- [ ] `EXPLAIN` for country flow uses `stats.country_pair_flows`
- [ ] Milestone trigger inserts row on km_100, km_1000 threshold crossing
- [ ] First finder trigger inserts row only once per GK and only within 7-day window
- [ ] `gk_loves` trigger handles insert, update, and delete changes for shard, daily, and country love rollups
- [ ] Event-bridge trigger inserts one row per qualifying move
- [ ] `phinx rollback` for each migration succeeds cleanly
