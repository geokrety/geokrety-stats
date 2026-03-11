---
title: Geokrety Schema
description: Transactional source schema and the trigger layer that feeds the canonical stats schema.
icon: material/database-arrow-right
---

# `geokrety` schema

`geokrety` is the operational source of truth. The current stats branch did not move write ownership into `stats`; instead, it added source-side columns, functions, and triggers so analytics state is updated where the original business event occurs.

## Table of contents

- [Role](#role)
- [Stats-critical objects](#stats-critical-objects)
- [Core tables](#core-tables)
- [Trigger ownership](#trigger-ownership)
- [Read surfaces and helper objects](#read-surfaces-and-helper-objects)
- [Operational cautions](#operational-cautions)

## Role

Three source tables matter most to the current branch:

- `gk_moves`: the event stream for movement, comments, seen logs, archives, and dips
- `gk_geokrety`: GeoKret identity and owner state
- `gk_users`, `gk_pictures`, `gk_loves`: supporting event sources for user, content, and engagement counters

The live database shows `gk_moves` at roughly `6.9M` estimated rows and `17 GB`, making it the dominant source relation for analytics cost.

## Stats-critical objects

### `gk_moves`

The March 2026 chain added and now depends on:

- `previous_move_id`
- `previous_position_id`
- `km_distance`

These columns support chain traversal, deterministic distance calculation, and snapshot backfills. They are maintained by:

- `fn_set_previous_move_id_and_distance()`
- `fn_refresh_previous_move_ids_after_insert()`
- `fn_refresh_previous_move_ids_after_update()`
- `fn_rewire_previous_move_ids_after_delete()`

### `gk_geokrety`

The key branch addition here is live first-finder reconciliation. The trigger:

- `tr_gk_geokrety_after_first_finder`

fires `fn_gk_geokrety_first_finder()` on owner or creation-time changes and on delete.

## Core tables

### Movement and GeoKret domain

- `gk_geokrety`: main GK identity, ownership, holder, counters, and lifecycle fields
- `gk_moves`: movement log and analytics event source
- `gk_moves_comments`: comment and missing-report domain attached to moves
- `gk_pictures`: picture uploads linked to GK, move, or user
- `gk_loves`: lightweight engagement table
- `gk_watched`: watchlist subscriptions
- `vw_geokret_move_history`: human-readable move-chain diagnostic view added on this branch

### User and authentication domain

- `gk_users`
- `gk_account_activation`
- `gk_email_activation`
- `gk_email_revalidate`
- `gk_password_tokens`
- `gk_users_authentication_history`
- `gk_users_settings` and `gk_users_settings_parameters`
- `gk_users_social_auth`
- `gk_users_username_history`

### Content, messaging, and community domain

- `gk_news`, `gk_news_comments`, `gk_news_comments_access`
- `gk_mails`
- `gk_labels`
- `gk_awards`, `gk_awards_group`, `gk_awards_won`, `gk_yearly_ranking`

### Waypoints and geospatial source domain

- `gk_waypoints_gc`
- `gk_waypoints_oc`
- `gk_waypoints_country`
- `gk_waypoints_sync`
- `gk_waypoints_types`

### Operational and metadata tables

- `phinxlog`
- `schema_migrations`
- `scripts`
- `sessions`
- `gk_site_settings` and `gk_site_settings_parameters`
- `gk_rate_limit_overrides`

## Trigger ownership

The current branch relies on these `gk_moves` triggers in particular:

- `tr_gk_moves_before_prev_move`
- `tr_gk_moves_after_prev_move_insert`
- `tr_gk_moves_after_prev_move_update`
- `tr_gk_moves_after_prev_move_delete`
- `tr_gk_moves_after_daily_activity`
- `tr_gk_moves_after_sharded_counters`
- `tr_gk_moves_after_country_rollups`
- `tr_gk_moves_after_country_history`
- `tr_gk_moves_after_waypoint_visits`
- `tr_gk_moves_after_relations`
- `tr_gk_moves_after_milestones`
- `tr_gk_moves_after_first_finder`

This matters operationally: if a bulk load disables trigger semantics or bypasses application writes, the analytics surfaces in `stats` will drift unless a targeted rebuild or snapshot is run afterward.

## Read surfaces and helper objects

Useful live read helpers include:

- `vw_geokret_move_history`: move chain inspection with human labels
- `gk_geokrety_with_details`: denormalized current GK view
- `gk_geokrety_in_caches`: materialized subset for cache-located GKs
- `gk_statistics_country_trends` and `gk_statistics_daily_counters`: older statistics still present outside the canonical `stats` schema

## Operational cautions

- Do not treat `distance` on `gk_moves` as a replacement for `km_distance`; the branch standardized deterministic numeric distance through the new lineage columns.
- Trigger ordering matters. The source schema now owns a substantial part of analytics correctness.
- For bulk repair work, prefer session flags and dedicated repair functions over disabling triggers globally. If trigger semantics were bypassed and `stats` drifted, recover through [the stats snapshot and backfill workflow](specs.stats.md#snapshot-and-backfill-commands).
- When changing source function signatures, explicitly drop stale overloads. PostgreSQL will otherwise keep older wrappers alive.
