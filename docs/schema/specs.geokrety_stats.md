---
title: Geokrety Stats Schema
description: Legacy and gamification analytics schema still present in the live database.
icon: material/trophy-outline
---

# `geokrety_stats` schema

The live database still contains a large `geokrety_stats` schema. It predates the canonical `stats` schema and also carries the gamification and points pipeline. It is real production surface area, not a placeholder.

## Table of contents

- [Role](#role)
- [Key tables](#key-tables)
- [Materialized views](#materialized-views)
- [Partitioned tables](#partitioned-tables)
- [Relation to canonical stats](#relation-to-canonical-stats)

## Role

This schema owns the points and gamification model, including:

- user points ledger and totals
- multiplier state per GK
- chains and chain-completion logic
- monthly diversity tracking
- location-based monthly scoring caps
- legacy leaderboard and summary materialized views

## Key tables

- `user_points_log`: append-only user points ledger
- `user_points_totals`: current totals per user
- `gk_multiplier_state`: current multiplier per GK
- `gk_points_log`: multiplier-change audit log
- `processed_events`: exactly-once processing registry for source moves
- `gk_chains`, `gk_chain_members`, `gk_chain_completions`: chain gameplay surfaces
- `gk_countries_visited`: legacy country-visit registry per GK
- `user_move_history`: first-move-only enforcement surface
- `user_owner_gk_counts` and `user_owner_counts_summary`: anti-farming and owner-limit enforcement
- `user_monthly_diversity`, `user_monthly_diversity_countries`, `user_monthly_diversity_drops`, `user_monthly_diversity_owners`: monthly diversity logic
- `user_waypoint_monthly_counts`: partitioned monthly location cap tracking

## Materialized views

Representative read models include:

- `mv_global_stats`
- `mv_gk_stats`
- `mv_user_stats`
- `mv_country_stats` and `mv_country_summary`
- `mv_daily_activity`
- `mv_gk_countries`
- `mv_user_countries`
- `mv_user_points_daily`
- `mv_user_related_users`
- `mv_geokrety_related_users`
- `mv_leaderboard_all_time`, `mv_leaderboard_daily`, `mv_leaderboard_monthly`, `mv_leaderboard_yearly`

## Partitioned tables

Two large operational patterns stand out:

- `gk_points_log` is range partitioned by `updated_at`, with monthly partitions spanning historical periods through `2026-02` plus a default partition
- `user_waypoint_monthly_counts` is list partitioned by `year_month`, pre-created far into the future for retention and maintenance simplicity

Live maintenance functions also matter here, especially `refresh_leaderboard_views`, `refresh_leaderboard_views_light`, `refresh_leaderboard_views_heavy`, `ensure_gk_points_log_month_partition`, `rotate_gk_points_log_partitions`, and `ensure_user_waypoint_month_partitions`.

Partition rotation is operationally significant. If current partitions are not created on time, writes fall into default partitions and complicate retention and performance work.

This is important because the schema already demonstrates partition-heavy operational design, whereas canonical `stats` currently stays on plain PostgreSQL tables and materialized views.

## Relation to canonical stats

The current branch does not replace `geokrety_stats`. Instead:

- `stats` becomes the canonical analytics schema for branch-owned operational and reporting surfaces
- `geokrety_stats` remains the gamification and legacy analytics domain
- future APIs should make a deliberate distinction between canonical operational analytics and points-specific or legacy summaries

For product design, that usually means:

- `stats` backs dashboards, trend charts, timelines, and country exploration
- `geokrety_stats` backs scoring, rankings, streaks, and rewards

### Public versus internal boundary

Reasonable public read contracts live around leaderboard and summary views. Internal processing state includes at least:

- `processed_events`
- `user_points_log`
- `gk_points_log`
- `user_owner_gk_counts`
- `user_waypoint_monthly_counts`

Those tables are appropriate for operator tools and repair workflows, not as public API contracts.
