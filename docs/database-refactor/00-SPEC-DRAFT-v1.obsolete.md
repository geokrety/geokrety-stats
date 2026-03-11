
# GeoKrety Stats Schema Comprehensive Specification

## Table of Contents

- [1. Scope and Constraints](#1-scope-and-constraints)
- [2. Canonical Domain Rules](#2-canonical-domain-rules)
- [3. Fresh-Start Architecture](#3-fresh-start-architecture)
- [4. Source Table Enhancements (No Website Code Change)](#4-source-table-enhancements-no-website-code-change)
- [5. Stats Schema (`stats`)](#5-stats-schema-stats)
- [6. Points Schema (`points`)](#6-points-schema-points)
- [7. Achievements Schema (`achievements`)](#7-achievements-schema-achievements)
- [8. Trigger and Function Design](#8-trigger-and-function-design)
- [8.5 Trigger and Function Summary](#85-trigger-and-function-summary)
- [9. Index Strategy (Required)](#9-index-strategy-required)
- [10. Use Cases and Views (UC1-UC15)](#10-use-cases-and-views-uc1-uc15)
- [11. Snapshot Ingestion and Backfill Helpers](#11-snapshot-ingestion-and-backfill-helpers)
- [12. Migration and Rollback Runbook](#12-migration-and-rollback-runbook)
- [13. pgtap Test Matrix (Large)](#13-pgtap-test-matrix-large)
- [14. Implementation Checklist](#14-implementation-checklist)

## 1. Scope and Constraints

This document is a fresh schema plan. Assume the following migrations are reverted before implementation and therefore no tables/functions/triggers from them are kept:

- `20260307140000_create_entity_sharded_counters.php`
- `20260304121000_create_stats_triggers.php`
- `20260304120500_add_gk_moves_trigger_indexes.php`
- `20260304120000_create_stats_tables.php`
- `20260228174500_optimize_gk_moves_indexes.php`

Design goals:

- Support fast analytics and scoring on replay of about 10,000,000 `geokrety.gk_moves` rows.
- Avoid hot rows and full table scans in runtime API queries.
- Keep scoring (`points`) and badge logic (`achievements`) in separate schemas.
- Do not require website application code changes.
- Prefer append-friendly structures, idempotent replay, and resumable heavy jobs.

## 2. Canonical Domain Rules

Move type mapping:

- `0 DROP`
- `1 GRAB`
- `2 COMMENT`
- `3 SEEN`
- `4 ARCHIVE`
- `5 DIP`

Geokret type mapping:

- Standard transferable: `0,1,2,3,4,5,7,9`
- Non-transferable: `6,8,10`

Resolved design decisions (integrated from OQ items):

- Cache questions must be answered by dedicated stats tables (`OQ1`).
- `previous_move_id` is stored on `gk_moves` and used to compute live distance (`OQ2`).
- UC2 must not query `gk_moves` directly in frontend-facing views (`OQ3/UC2`).
- Previous-location source move types are `DROP, GRAB(with location), SEEN(with location), DIP` (`OQ4`).
- PK/index ordering follows query shape (user-first where user-first lookup dominates) (`OQ5`).
- Existing anti-farming/owner-limit granularity is sufficient and explicitly persisted (`OQ6`).
- No backward compatibility layer is required (`OQ7`).

## 3. Fresh-Start Architecture

Schemas:

- `stats`: counters, aggregates, relationships, geography/time buckets, helper operations.
- `points`: scoring state, ledger, multiplier state, chain state, anti-farming state.
- `achievements`: definitions, criteria, progress, awards.

Write path:

1. Insert/update/delete in source tables (`geokrety.gk_moves`, `gk_geokrety`, `gk_users`, `gk_pictures`, `gk_loves`).
2. Triggers update small state tables and append event records.
3. Heavy recomputation is done by manual helper functions and replay jobs.
4. API/graphs query views/materialized views over pre-aggregated tables.

Hybrid execution model:

- PostgreSQL remains the source of truth for deterministic state updates (counters, relationships, minimal chain state).
- `points-awarder` (Go + RabbitMQ consumer) is recommended for heavy/rule-rich awarding logic and replay orchestration.
- DB triggers should emit compact events and maintain only low-cost synchronous facts.
- `points-awarder` should calculate awards, update `points.*`, and publish audit/result events.

Read path:

- Runtime dashboards and leaderboards should read from `stats.v_*`, `points.v_*`, and materialized views.
- Avoid scanning `geokrety.gk_moves` in frontend-facing queries.

## 4. Source Table Enhancements (No Website Code Change)

### 4.1 `geokrety.gk_moves` additional columns

```sql
ALTER TABLE geokrety.gk_moves
  ADD COLUMN previous_move_id BIGINT,
  ADD COLUMN km_distance NUMERIC(8,3);

ALTER TABLE geokrety.gk_moves
  ADD CONSTRAINT fk_gk_moves_previous_move
  FOREIGN KEY (previous_move_id) REFERENCES geokrety.gk_moves(id)
  DEFERRABLE INITIALLY DEFERRED;
```

Notes:

- `NUMERIC(8,3)` is used for deterministic aggregates and precise sums.
- `previous_move_id` and `km_distance` are trigger-managed for new rows.
- `previous_move_id` is preferred over storing a duplicate geography payload.
- `geokrety.gk_geokrety.last_position` is used as the primary fast-path candidate.
- Historical rows are backfilled by explicit heavy/manual helper functions.

### 4.2 Previous-location semantics

For a move `m`, `previous_move_id` references the most recent earlier move of the same GK where:

- move type in `(0, 1, 3, 5)`
- effective location exists
- sort key: `(moved_on_datetime DESC, id DESC)`

Effective location:

- `DROP` and `DIP`: location expected.
- `SEEN`: only if location exists.
- `GRAB`: only if location exists.

Distance computation:

- `km_distance` on current move is computed from geometry of `(previous_move_id -> current_move)`.
- If no valid previous move exists, `previous_move_id` and `km_distance` remain NULL.

## 5. Stats Schema (`stats`)

### 5.1 Exact sharded counters

What this unlocks:

- Instant counters for hero KPIs (total moves, users, GKs, pictures, loves).
- Fast move-type and GK-type pies for dashboard cards.
- Low-contention increments during heavy ingest/replay.

Stats/graph samples:

- Total moves over time baseline (joined with `stats.daily_entity_counts`).
- Distribution chart: `gk_moves_type_0..5` and `gk_geokrety_type_0..10`.

Trigger responsibility:

- `AFTER INSERT/DELETE/UPDATE` triggers on `geokrety.gk_moves`, `gk_geokrety`, `gk_pictures`, `gk_users`, `gk_loves` update shard counts.

```sql
CREATE TABLE stats.entity_counters_shard (
  entity VARCHAR(32) NOT NULL,
  shard INT NOT NULL,
  cnt BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (entity, shard)
);
```

Important: use exact column names from migration analysis: `entity`, `shard`, `cnt`.

Counter entities:

- `gk_moves`, `gk_moves_type_0..5`
- `gk_geokrety`, `gk_geokrety_type_0..10`
- `gk_pictures`, `gk_pictures_type_0..2`
- `gk_users`
- `gk_loves`

### 5.2 Daily/global activity tables

What this unlocks:

- Global daily activity chart (moves/km/points/loves).
- Daily content production chart (GK created, pictures uploaded by type, user signups).
- Daily platform contribution leaderboard (top point-contributing days/weeks).

Stats/graph samples:

- Stacked area by `drops/grabs/sees/dips/comments/archives`.
- Daily points contribution trend and 7-day moving average.
- Picture upload mix chart by picture type.

```sql
CREATE TABLE stats.daily_activity (
  activity_date DATE PRIMARY KEY,
  total_moves BIGINT NOT NULL DEFAULT 0,
  drops BIGINT NOT NULL DEFAULT 0,
  grabs BIGINT NOT NULL DEFAULT 0,
  comments BIGINT NOT NULL DEFAULT 0,
  sees BIGINT NOT NULL DEFAULT 0,
  archives BIGINT NOT NULL DEFAULT 0,
  dips BIGINT NOT NULL DEFAULT 0,
  km_contributed NUMERIC(14,3) NOT NULL DEFAULT 0,
  points_contributed NUMERIC(16,4) NOT NULL DEFAULT 0,
  loves_count BIGINT NOT NULL DEFAULT 0,
  gk_created BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_total BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_avatar BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_move BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_user BIGINT NOT NULL DEFAULT 0,
  users_registered BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE stats.daily_active_users (
  activity_date DATE NOT NULL,
  user_id INT NOT NULL,
  PRIMARY KEY (activity_date, user_id)
);

CREATE TABLE stats.daily_entity_counts (
  count_date DATE NOT NULL,
  entity VARCHAR(32) NOT NULL,
  cnt BIGINT NOT NULL,
  PRIMARY KEY (count_date, entity)
);
```

Trigger responsibility:

- `gk_moves` daily trigger updates move-type counters and `km_contributed`.
- `points.gk_moves_points` award trigger (or points-awarder upsert) updates `points_contributed`.
- `gk_pictures` trigger updates total + per-type picture counters.
- `gk_users`/`gk_geokrety`/`gk_loves` triggers update corresponding daily columns.

### 5.3 Country and GK traversal tables

What this unlocks:

- Country choropleth and country-versus-country competition panels.
- Country points contribution leaderboard.
- Country media engagement chart (pictures/loves) by day or month.

Stats/graph samples:

- Monthly choropleth by `moves_count`, `km_contributed`, `points_contributed`.
- Top countries by `unique_gks` and `unique_users`.
- Country gallery activity chart (`pictures_uploaded_*`).

```sql
CREATE TABLE stats.country_daily_stats (
  stats_date DATE NOT NULL,
  country_code CHAR(2) NOT NULL,
  moves_count BIGINT NOT NULL DEFAULT 0,
  drops BIGINT NOT NULL DEFAULT 0,
  grabs BIGINT NOT NULL DEFAULT 0,
  comments BIGINT NOT NULL DEFAULT 0,
  sees BIGINT NOT NULL DEFAULT 0,
  archives BIGINT NOT NULL DEFAULT 0,
  dips BIGINT NOT NULL DEFAULT 0,
  unique_users BIGINT NOT NULL DEFAULT 0,
  unique_gks BIGINT NOT NULL DEFAULT 0,
  km_contributed NUMERIC(14,3) NOT NULL DEFAULT 0,
  points_contributed NUMERIC(16,4) NOT NULL DEFAULT 0,
  loves_count BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_total BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_avatar BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_move BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_user BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (stats_date, country_code)
);

CREATE TABLE stats.gk_countries_visited (
  geokrety_id INT NOT NULL,
  country_code CHAR(2) NOT NULL,
  first_visited_at TIMESTAMPTZ NOT NULL,
  first_move_id BIGINT NOT NULL,
  move_count INT NOT NULL DEFAULT 1,
  PRIMARY KEY (geokrety_id, country_code)
);

CREATE TABLE stats.user_countries (
  user_id INT NOT NULL,
  country_code CHAR(2) NOT NULL,
  move_count BIGINT NOT NULL DEFAULT 0,
  first_visit TIMESTAMPTZ NOT NULL,
  last_visit TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (user_id, country_code)
);
```

Trigger responsibility:

- `gk_moves` trigger maintains `country_daily_stats`, `gk_countries_visited`, `user_countries`.
- `points` upsert path updates `country_daily_stats.points_contributed` by actor country.
- `gk_pictures` and `gk_loves` triggers update country picture/love counters when location context exists.

### 5.4 Country history timeline

What this unlocks:

- Current country map of all active GKs without scanning move history.
- Country dwell-time and transition analytics.
- Country arrival/departure timelines for GK story graphs.

Trigger responsibility:

- `gk_moves` trigger opens/closes intervals in `stats.gk_country_history` on location-bearing moves.

```sql
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE stats.gk_country_history (
  id BIGSERIAL PRIMARY KEY,
  geokrety_id INT NOT NULL,
  country_code CHAR(2) NOT NULL,
  arrived_at TIMESTAMPTZ NOT NULL,
  departed_at TIMESTAMPTZ,
  move_id BIGINT NOT NULL,
  CONSTRAINT gk_country_history_excl
    EXCLUDE USING gist (
      geokrety_id WITH =,
      tstzrange(arrived_at, COALESCE(departed_at, 'infinity')) WITH &&
    )
);
```

### 5.5 Waypoint/cache model for optimized cache queries

What this unlocks:

- Most-visited cache leaderboards (global, per-country, per-user).
- Cache discovery heatmaps and retention analytics.
- Direct answers to cache-based OQ queries without replaying raw moves.

`source` column purpose:

- `GC`: waypoint originated from `geokrety.gk_waypoints_gc`.
- `OC`: waypoint originated from `geokrety.gk_waypoints_oc`.
- `UK`: unknown/unmapped source when seen first in move stream.
- This supports provenance, deduping quality checks, and source-specific refresh jobs.

Reuse existing waypoint tables:

- Yes, reuse both waypoint tables and seed `stats.waypoints` from them.
- Also create an optional union view for diagnostics:

```sql
CREATE VIEW stats.v_waypoints_source_union AS
SELECT UPPER(waypoint) AS waypoint_code, 'GC'::char(2) AS source, lat, lon, UPPER(country) AS country
FROM geokrety.gk_waypoints_gc
UNION ALL
SELECT UPPER(waypoint) AS waypoint_code, 'OC'::char(2) AS source, lat, lon, UPPER(country) AS country
FROM geokrety.gk_waypoints_oc;
```

```sql
CREATE TABLE stats.waypoints (
  id BIGSERIAL PRIMARY KEY,
  waypoint_code VARCHAR(20) NOT NULL UNIQUE,
  source CHAR(2) NOT NULL DEFAULT 'UK',
  lat DOUBLE PRECISION,
  lon DOUBLE PRECISION,
  country CHAR(2),
  first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE stats.gk_cache_visits (
  gk_id INT NOT NULL,
  waypoint_id BIGINT NOT NULL REFERENCES stats.waypoints(id),
  visit_count BIGINT NOT NULL DEFAULT 0,
  first_visited_at TIMESTAMPTZ NOT NULL,
  last_visited_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (gk_id, waypoint_id)
);

CREATE TABLE stats.user_cache_visits (
  user_id INT NOT NULL,
  waypoint_id BIGINT NOT NULL REFERENCES stats.waypoints(id),
  visit_count BIGINT NOT NULL DEFAULT 0,
  first_visited_at TIMESTAMPTZ NOT NULL,
  last_visited_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (user_id, waypoint_id)
);
```

This directly answers:

- which cache a geokrety/user visited
- how many caches a geokrety visited

without full `gk_moves` scans.

Trigger responsibility:

- `gk_moves` trigger resolves/creates waypoint rows and upserts cache visit counters.

### 5.6 User/GK relation tables for UC2 and social graph

What this unlocks:

- UC2 social-network graph without direct `gk_moves` reads.
- Related-user leaderboard (most shared GKs).
- GK reach/interaction graph and “most social GK” ranking.

Trigger responsibility:

- `gk_moves` trigger upserts `gk_related_users` and directional/normalized `user_related_users`.

```sql
CREATE TABLE stats.gk_related_users (
  geokrety_id INT NOT NULL,
  user_id INT NOT NULL,
  interaction_count BIGINT NOT NULL DEFAULT 0,
  first_interaction TIMESTAMPTZ NOT NULL,
  last_interaction TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (geokrety_id, user_id)
);

CREATE TABLE stats.user_related_users (
  user_id INT NOT NULL,
  related_user_id INT NOT NULL,
  shared_geokrety_count BIGINT NOT NULL DEFAULT 0,
  first_seen_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (user_id, related_user_id)
);
```

UC2 views must use these tables only.

### 5.7 Additional analytics tables (new UC support)

What this unlocks:

- Hour/day heatmaps, country-to-country flow maps, milestone timelines, first-finder hall of fame.
- Rich graph panels for UC8-UC15.

Trigger/job responsibility:

- Lightweight triggers append milestone/first-finder facts.
- Batch jobs compute `country_pair_flows` and `hourly_activity` buckets.

```sql
CREATE TABLE stats.continent_reference (
  country_alpha2 CHAR(2) PRIMARY KEY,
  continent_code CHAR(2) NOT NULL,
  continent_name VARCHAR(50) NOT NULL
);

CREATE TABLE stats.hourly_activity (
  activity_date DATE NOT NULL,
  hour_utc SMALLINT NOT NULL CHECK (hour_utc BETWEEN 0 AND 23),
  move_type SMALLINT NOT NULL CHECK (move_type BETWEEN 0 AND 5),
  move_count BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (activity_date, hour_utc, move_type)
);

CREATE TABLE stats.country_pair_flows (
  year_month DATE NOT NULL,
  from_country CHAR(2) NOT NULL,
  to_country CHAR(2) NOT NULL,
  move_count BIGINT NOT NULL DEFAULT 0,
  unique_gk_count BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (year_month, from_country, to_country)
);

CREATE TABLE stats.gk_milestone_events (
  id BIGSERIAL PRIMARY KEY,
  gk_id INT NOT NULL,
  event_type VARCHAR(50) NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL,
  actor_user_id INT,
  metadata JSONB,
  CHECK (event_type IN (
    'country_first', 'km_100', 'km_1000', 'km_10000',
    'users_10', 'users_50', 'users_100', 'first_find'
  ))
);

CREATE TABLE stats.first_finder_events (
  gk_id INT PRIMARY KEY,
  finder_user_id INT NOT NULL,
  move_id BIGINT NOT NULL,
  move_type SMALLINT NOT NULL,
  found_at TIMESTAMPTZ NOT NULL,
  gk_created_at TIMESTAMPTZ NOT NULL,
  hours_since_creation SMALLINT NOT NULL
);
```

### 5.8 Operational support tables

What this unlocks:

- Safe resumable heavy operations with cursor checkpoints.
- Auditability and operations dashboards for long-running replay/backfill jobs.

```sql
CREATE TABLE stats.backfill_progress (
  job_name VARCHAR(100) PRIMARY KEY,
  target_table VARCHAR(100) NOT NULL,
  min_id BIGINT NOT NULL DEFAULT 0,
  max_id BIGINT NOT NULL,
  cursor_id BIGINT NOT NULL DEFAULT 0,
  batch_size INT NOT NULL DEFAULT 10000,
  status VARCHAR(20) NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending','running','paused','completed','failed')),
  rows_processed BIGINT NOT NULL DEFAULT 0,
  error_count INT NOT NULL DEFAULT 0,
  started_at TIMESTAMPTZ,
  last_heartbeat_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  notes TEXT,
  last_error TEXT
);

CREATE TABLE stats.job_log (
  id BIGSERIAL PRIMARY KEY,
  job_name VARCHAR(100) NOT NULL,
  status VARCHAR(20) NOT NULL,
  metadata JSONB,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);
```

## 6. Points Schema (`points`)

### 6.1 Core scoring ledger and daily totals

What this unlocks:

- Transparent score audit per move (for disputes and debugging).
- Daily/weekly/monthly user leaderboards from pre-aggregated totals.
- Bonus-component charts (relay/rescuer/chain/country/diversity contribution mix).

Design choice: `points.gk_moves_points` vs JSON on `gk_moves`

- Keep `points.gk_moves_points` as authoritative normalized ledger.
- Do not store/maintain a `points_breakdown JSONB` on `gk_moves` to avoid duplication and drift.
- Optional denormalized JSON can be exposed in API layer/view if needed, generated from ledger columns.

```sql
CREATE TABLE points.gk_moves_points (
  move_id BIGINT PRIMARY KEY,
  gk_id INT NOT NULL,
  actor_user_id INT,
  owner_user_id INT,
  move_type SMALLINT NOT NULL,
  multiplier_used NUMERIC(6,4) NOT NULL,
  base_points NUMERIC(10,4) NOT NULL DEFAULT 0,
  relay_bonus NUMERIC(10,4) NOT NULL DEFAULT 0,
  rescuer_bonus NUMERIC(10,4) NOT NULL DEFAULT 0,
  chain_bonus NUMERIC(10,4) NOT NULL DEFAULT 0,
  country_bonus NUMERIC(10,4) NOT NULL DEFAULT 0,
  reach_bonus NUMERIC(10,4) NOT NULL DEFAULT 0,
  diversity_bonus NUMERIC(10,4) NOT NULL DEFAULT 0,
  location_penalty_factor NUMERIC(5,2) NOT NULL DEFAULT 1.00,
  total_points NUMERIC(12,4) NOT NULL,
  formula_version SMALLINT NOT NULL DEFAULT 1,
  calculated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE points.user_daily_points (
  user_id INT NOT NULL,
  points_date DATE NOT NULL,
  base_points_totals NUMERIC(12,4) NOT NULL DEFAULT 0,
  relay_bonus_totals NUMERIC(12,4) NOT NULL DEFAULT 0,
  rescuer_bonus_totals NUMERIC(12,4) NOT NULL DEFAULT 0,
  chain_bonus_totals NUMERIC(12,4) NOT NULL DEFAULT 0,
  country_bonus_totals NUMERIC(12,4) NOT NULL DEFAULT 0,
  reach_bonus_totals NUMERIC(12,4) NOT NULL DEFAULT 0,
  diversity_bonus_totals NUMERIC(12,4) NOT NULL DEFAULT 0,
  total_points_totals NUMERIC(14,4) NOT NULL DEFAULT 0,
  awards_count_totals BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, points_date)
);
```

Location penalty only affects `base_points`, not other bonuses.

Trigger/job responsibility:

- If pure SQL mode: trigger on `gk_moves` computes and upserts ledger + daily totals.
- If `points-awarder` mode: DB trigger emits event, awarder computes and writes both tables.

### 6.2 Multiplier state and audit

What this unlocks:

- Multiplier trend charts and volatility dashboards.
- High-value GK leaderboard by current multiplier and multiplier velocity.

Trigger/job responsibility:

- Insert-time move handler applies +0.01/+0.05 increments (rule-based).
- Scheduled decay job applies time-based decrements and writes audit rows.

```sql
CREATE TABLE points.gk_multiplier (
  gk_id INT PRIMARY KEY,
  current_multiplier NUMERIC(6,4) NOT NULL DEFAULT 1.0000,
  holder_since TIMESTAMPTZ,
  cache_since TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (current_multiplier >= 1.0000 AND current_multiplier <= 2.0000)
);

CREATE TABLE points.gk_multiplier_audit (
  id BIGSERIAL PRIMARY KEY,
  gk_id INT NOT NULL,
  multiplier_before NUMERIC(6,4) NOT NULL,
  multiplier_after NUMERIC(6,4) NOT NULL,
  reason VARCHAR(64) NOT NULL,
  source_move_id BIGINT,
  calculated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 6.3 Owner history and anti-farming

What this unlocks:

- Correct anti-farming enforcement across ownership changes.
- Owner circulation contribution leaderboard without false positives.

Trigger responsibility:

- `gk_geokrety.owner` change trigger maintains ownership intervals.
- Awarding logic checks previous-owner and max-10-per-owner scope tables.

```sql
CREATE TABLE points.gk_owner_history (
  geokrety_id INT NOT NULL,
  owner_user_id INT NOT NULL,
  owned_from TIMESTAMPTZ NOT NULL,
  owned_to TIMESTAMPTZ,
  PRIMARY KEY (geokrety_id, owner_user_id, owned_from)
);

CREATE TABLE points.owner_gk_earning_scope (
  user_id INT NOT NULL,
  owner_user_id INT NOT NULL,
  geokrety_id INT NOT NULL,
  first_earned_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (user_id, owner_user_id, geokrety_id)
);
```

`owner_gk_earning_scope` is used to enforce max 10 GKs per owner per user.

### 6.4 Chain model (normalized, no array)

What this unlocks:

- Chain length leaderboard, chain closure stats, and chain quality dashboards.
- Deterministic membership tracking with simple SQL semantics.

`points-awarder` integration:

- On chain-close detection, publish RabbitMQ message with `chain_id`, `gk_id`, `chain_length`, participants.
- `points-awarder` computes payouts and writes `points.gk_moves_points` + `points.user_daily_points`.
- Database still enforces idempotency via `user_gk_chain_bonuses` and unique keys.

```sql
CREATE TABLE points.movement_chains (
  chain_id BIGSERIAL PRIMARY KEY,
  gk_id INT NOT NULL,
  chain_status VARCHAR(16) NOT NULL CHECK (chain_status IN ('active','closed')),
  started_at TIMESTAMPTZ NOT NULL,
  latest_move_at TIMESTAMPTZ NOT NULL,
  closed_at TIMESTAMPTZ,
  chain_length INT NOT NULL DEFAULT 0,
  closure_reason VARCHAR(64),
  bonus_awarded BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX idx_movement_chains_one_active_per_gk
  ON points.movement_chains (gk_id)
  WHERE chain_status = 'active';

CREATE TABLE points.movement_chain_members (
  chain_id BIGINT NOT NULL REFERENCES points.movement_chains(chain_id),
  user_id INT NOT NULL,
  join_order SMALLINT NOT NULL,
  joined_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (chain_id, user_id)
);

CREATE TABLE points.user_gk_chain_bonuses (
  user_id INT NOT NULL,
  gk_id INT NOT NULL,
  last_chain_bonus_at TIMESTAMPTZ NOT NULL,
  bonus_count INT NOT NULL DEFAULT 1,
  PRIMARY KEY (user_id, gk_id)
);
```

### 6.5 Location anti-farming tables (corrected split)

What this unlocks:

- Anti-farming compliance dashboards.
- Monthly location pressure charts and suspicious-pattern detection.

Trigger responsibility:

- On eligible move events, update distinct GK set and monthly count.
- Penalty factor derived as 1.0 / 0.5 / 0.25 / 0.0 from `gk_count`.

```sql
CREATE TABLE points.location_monthly_gk_count (
  user_id INT NOT NULL,
  location_key TEXT NOT NULL,
  penalty_year_month DATE NOT NULL,
  gk_count INT NOT NULL DEFAULT 0,
  last_updated TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, location_key, penalty_year_month)
);

CREATE TABLE points.location_monthly_gk_set (
  user_id INT NOT NULL,
  location_key TEXT NOT NULL,
  penalty_year_month DATE NOT NULL,
  gk_id INT NOT NULL,
  PRIMARY KEY (user_id, location_key, penalty_year_month, gk_id)
);
```

`location_key` identity rule:

- `WP:<UPPER(waypoint)>` if waypoint exists.
- otherwise `LL:<lat_rounded_4>,<lon_rounded_4>` if coordinates exist.
- otherwise NULL and no location penalty update.

### 6.6 Diversity and monthly trackers

What this unlocks:

- Monthly diversity progress cards and streak-like charts.
- Diversity leaderboard (unique GKs, owner interactions, country diversity).

Trigger/job responsibility:

- Awarding pipeline updates monthly counters and flips awarded flags once.

```sql
CREATE TABLE points.diversity_bonus_tracking (
  user_id INT NOT NULL,
  bonus_year_month DATE NOT NULL,
  distinct_gks_dropped INT NOT NULL DEFAULT 0,
  distinct_owners_interacted INT NOT NULL DEFAULT 0,
  distinct_countries_visited INT NOT NULL DEFAULT 0,
  gks_bonus_awarded BOOLEAN NOT NULL DEFAULT FALSE,
  owners_bonus_awarded BOOLEAN NOT NULL DEFAULT FALSE,
  countries_bonus_awarded BOOLEAN NOT NULL DEFAULT FALSE,
  PRIMARY KEY (user_id, bonus_year_month)
);
```

## 7. Achievements Schema (`achievements`)

What this unlocks:

- Evolving multi-level achievements with XP progression.
- Progress bars, level-up feeds, and achievement leaderboards.
- Progressive badge families (logger, mover, kilometer, unique GKs, countries, rescuer, chain closer).

```sql
CREATE TABLE achievements.achievement_definitions (
  id SERIAL PRIMARY KEY,
  code VARCHAR(64) UNIQUE NOT NULL,
  title VARCHAR(128) NOT NULL,
  description TEXT NOT NULL,
  metric_key VARCHAR(64) NOT NULL,
  threshold NUMERIC(12,4) NOT NULL,
  scope VARCHAR(16) NOT NULL CHECK (scope IN ('lifetime','monthly','rolling_6m')),
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE achievements.achievement_levels (
  achievement_id INT NOT NULL REFERENCES achievements.achievement_definitions(id),
  level_no INT NOT NULL,
  threshold_value NUMERIC(14,4) NOT NULL,
  xp_reward INT NOT NULL DEFAULT 0,
  title_override VARCHAR(128),
  PRIMARY KEY (achievement_id, level_no)
);

CREATE TABLE achievements.user_xp (
  user_id INT PRIMARY KEY,
  total_xp BIGINT NOT NULL DEFAULT 0,
  current_level INT NOT NULL DEFAULT 1,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE achievements.user_progress (
  user_id INT NOT NULL,
  achievement_id INT NOT NULL REFERENCES achievements.achievement_definitions(id),
  period_start DATE,
  progress_value NUMERIC(14,4) NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, achievement_id, period_start)
);

CREATE TABLE achievements.user_awards (
  user_id INT NOT NULL,
  achievement_id INT NOT NULL REFERENCES achievements.achievement_definitions(id),
  awarded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  source_move_id BIGINT,
  details JSONB,
  PRIMARY KEY (user_id, achievement_id, awarded_at)
);
```

Progressive achievement examples:

- `progressive_logger`: 1, 5, 100, 1000 logs.
- `progressive_mover`: 1, 5, 100, 1000 non-comment moves.
- `progressive_kilometer`: 10, 100, 1000, 10000 km.
- `progressive_unique_gk`: 1, 10, 100, 500 unique moved GKs.
- `progressive_countries`: 1, 5, 25, 100 countries.
- `progressive_chain_closer`: 1, 5, 25 closed chains.
- `progressive_rescuer`: 1, 5, 25 rescue events.

Examples that map to prior UC ideas:

- UC5 achievement candidate: `first_finder`.
- UC11 achievement candidate: `social_hub` (high shared-geokrety graph centrality bucket).

Trigger/job responsibility:

- Lightweight trigger/event path increments raw counters.
- `achievements.fn_rebuild_progress_for_period` and incremental worker evaluate thresholds and insert awards.
- XP updates are atomic with award insertion.

## 8. Trigger and Function Design

### 8.1 Previous-location trigger

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_set_previous_move_id_and_distance()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_prev_move_id BIGINT;
BEGIN
  IF NEW.move_type NOT IN (0,1,3,5) OR NEW.position IS NULL THEN
    NEW.previous_move_id := NULL;
    NEW.km_distance := NULL;
    RETURN NEW;
  END IF;

  -- Fast path: previous move often matches gk_geokrety.last_position.
  SELECT g.last_position INTO v_prev_move_id
  FROM geokrety.gk_geokrety g
  WHERE g.id = NEW.geokret;

  IF v_prev_move_id IS NULL THEN
    SELECT m.id
      INTO v_prev_move_id
    FROM geokrety.gk_moves m
    WHERE m.geokret = NEW.geokret
      AND m.position IS NOT NULL
      AND m.move_type IN (0,1,3,5)
      AND (m.moved_on_datetime < NEW.moved_on_datetime
        OR (m.moved_on_datetime = NEW.moved_on_datetime AND m.id < NEW.id))
    ORDER BY m.moved_on_datetime DESC, m.id DESC
    LIMIT 1;
  END IF;

  NEW.previous_move_id := v_prev_move_id;

  IF NEW.previous_move_id IS NOT NULL THEN
    NEW.km_distance := (
      SELECT (public.ST_Distance(pm.position, NEW.position) / 1000.0)::NUMERIC(8,3)
      FROM geokrety.gk_moves pm
      WHERE pm.id = NEW.previous_move_id
        AND pm.position IS NOT NULL
    );
  ELSE
    NEW.km_distance := NULL;
  END IF;

  RETURN NEW;
END;
$$;
```

Attach as BEFORE trigger after GIS/location-normalization trigger order.

### 8.2 Helper functions for heavy backfill (manual execution)

```sql
CREATE OR REPLACE FUNCTION stats.fn_backfill_previous_move_id(
  p_period tstzrange DEFAULT NULL,
  p_batch_size INT DEFAULT 50000
) RETURNS BIGINT;

CREATE OR REPLACE FUNCTION stats.fn_backfill_heavy_previous_move_id_all()
RETURNS BIGINT;

CREATE OR REPLACE FUNCTION stats.fn_backfill_km_distance(
  p_period tstzrange DEFAULT NULL,
  p_batch_size INT DEFAULT 50000
) RETURNS BIGINT;

CREATE OR REPLACE FUNCTION stats.fn_backfill_heavy_km_distance_all()
RETURNS BIGINT;
```

Naming intentionally marks heavy variants.

Implementation note:

- Function names are kept for continuity, but they backfill `previous_move_id` and `km_distance`.

### 8.3 Snapshot/replay helpers

```sql
CREATE OR REPLACE FUNCTION stats.fn_snapshot_entity_counters() RETURNS VOID;
CREATE OR REPLACE FUNCTION stats.fn_snapshot_daily_country_stats(p_period daterange DEFAULT NULL) RETURNS BIGINT;
CREATE OR REPLACE FUNCTION stats.fn_snapshot_user_country_stats(p_period daterange DEFAULT NULL) RETURNS BIGINT;
CREATE OR REPLACE FUNCTION stats.fn_snapshot_gk_country_stats(p_period daterange DEFAULT NULL) RETURNS BIGINT;
CREATE OR REPLACE FUNCTION stats.fn_snapshot_relationship_tables(p_period daterange DEFAULT NULL) RETURNS BIGINT;
CREATE OR REPLACE FUNCTION points.fn_replay_points_for_period(p_period tstzrange DEFAULT NULL, p_batch_size INT DEFAULT 50000) RETURNS BIGINT;
CREATE OR REPLACE FUNCTION points.fn_replay_points_heavy_all() RETURNS BIGINT;
CREATE OR REPLACE FUNCTION achievements.fn_rebuild_progress_for_period(p_period daterange DEFAULT NULL) RETURNS BIGINT;
```

### 8.4 Trigger families

- `gk_moves` trigger family:
  - `tr_gk_moves_before_prev_move`: set `previous_move_id` + `km_distance`
  - `tr_gk_moves_after_sharded_counters`: update `stats.entity_counters_shard`
  - `tr_gk_moves_after_daily_activity`: update `stats.daily_activity`, `stats.daily_active_users`
  - `tr_gk_moves_after_country_rollups`: update `stats.country_daily_stats`, `stats.gk_countries_visited`, `stats.user_countries`
  - `tr_gk_moves_after_country_history`: update `stats.gk_country_history`
  - `tr_gk_moves_after_waypoint_visits`: update `stats.waypoints`, `stats.gk_cache_visits`, `stats.user_cache_visits`
  - `tr_gk_moves_after_relations`: update `stats.gk_related_users`, `stats.user_related_users`
  - `tr_gk_moves_emit_points_event`: enqueue RabbitMQ event payload row or `NOTIFY` payload for `points-awarder`
- `gk_geokrety` trigger family:
  - `tr_gk_geokrety_owner_history`: maintain `points.gk_owner_history`
  - `tr_gk_geokrety_counters`: maintain entity and daily GK creation counters
- `gk_loves`, `gk_users`, `gk_pictures` trigger families:
  - update sharded counters
  - update daily totals and country breakdowns where context is available

### 8.5 Trigger and Function Summary

| Name                                            | Type     | Event                              | Writes To                                                                          | Purpose                                     |
| ----------------------------------------------- | -------- | ---------------------------------- | ---------------------------------------------------------------------------------- | ------------------------------------------- |
| `geokrety.fn_set_previous_move_id_and_distance` | function | BEFORE INSERT/UPDATE on `gk_moves` | `NEW.previous_move_id`, `NEW.km_distance`                                          | link previous move and compute leg distance |
| `tr_gk_moves_after_sharded_counters`            | trigger  | AFTER I/U/D `gk_moves`             | `stats.entity_counters_shard`                                                      | exact fast counters                         |
| `tr_gk_moves_after_daily_activity`              | trigger  | AFTER I/U/D `gk_moves`             | `stats.daily_activity`, `stats.daily_active_users`                                 | daily macro KPIs                            |
| `tr_gk_moves_after_country_rollups`             | trigger  | AFTER I/U/D `gk_moves`             | `stats.country_daily_stats`, `stats.gk_countries_visited`, `stats.user_countries`  | country + traversal rollups                 |
| `tr_gk_moves_after_country_history`             | trigger  | AFTER I/U/D `gk_moves`             | `stats.gk_country_history`                                                         | temporal country intervals                  |
| `tr_gk_moves_after_waypoint_visits`             | trigger  | AFTER I/U/D `gk_moves`             | `stats.waypoints`, `stats.gk_cache_visits`, `stats.user_cache_visits`              | optimized cache analytics                   |
| `tr_gk_moves_after_relations`                   | trigger  | AFTER I/U/D `gk_moves`             | `stats.gk_related_users`, `stats.user_related_users`                               | UC2/social graph data                       |
| `tr_gk_moves_emit_points_event`                 | trigger  | AFTER INSERT `gk_moves`            | event table/queue bridge                                                           | asynchronous scoring for `points-awarder`   |
| `tr_gk_geokrety_owner_history`                  | trigger  | AFTER UPDATE OF owner              | `points.gk_owner_history`                                                          | anti-farming ownership timeline             |
| `tr_gk_geokrety_counters`                       | trigger  | AFTER INSERT/DELETE                | `stats.entity_counters_shard`, `stats.daily_activity`                              | GK volume KPIs                              |
| `tr_gk_pictures_activity`                       | trigger  | AFTER I/U/D `gk_pictures`          | `stats.entity_counters_shard`, `stats.daily_activity`, `stats.country_daily_stats` | picture totals + type breakdown             |
| `tr_gk_loves_activity`                          | trigger  | AFTER I/U/D `gk_loves`             | `stats.entity_counters_shard`, `stats.daily_activity`, `stats.country_daily_stats` | loves contribution                          |
| `tr_gk_users_activity`                          | trigger  | AFTER INSERT/DELETE `gk_users`     | `stats.entity_counters_shard`, `stats.daily_activity`                              | user growth KPI                             |
| `stats.fn_backfill_previous_move_id`            | function | manual                             | `geokrety.gk_moves`                                                                | batch backfill of previous-move links       |
| `stats.fn_backfill_km_distance`                 | function | manual                             | `geokrety.gk_moves`                                                                | batch distance backfill                     |
| `points.fn_replay_points_for_period`            | function | manual/job                         | `points.gk_moves_points`, `points.user_daily_points`                               | deterministic scoring replay                |
| `achievements.fn_rebuild_progress_for_period`   | function | manual/job                         | `achievements.user_progress`, `achievements.user_awards`, `achievements.user_xp`   | rebuild progress and awards                 |

## 9. Index Strategy (Required)

### 9.1 Source table indexes

```sql
CREATE INDEX idx_gk_moves_replay_cursor
  ON geokrety.gk_moves (moved_on_datetime ASC, id ASC);

CREATE INDEX idx_gk_moves_prev_loc_lookup
  ON geokrety.gk_moves (geokret, moved_on_datetime DESC, id DESC)
  WHERE position IS NOT NULL AND move_type IN (0,1,3,5);

CREATE INDEX idx_gk_moves_author_country_movedon
  ON geokrety.gk_moves (author, country, moved_on_datetime);

CREATE INDEX idx_gk_moves_geokret_country_movedon
  ON geokrety.gk_moves (geokret, country, moved_on_datetime);

CREATE INDEX idx_gk_moves_geokret_country_id
  ON geokrety.gk_moves (geokret, country, id);
```

### 9.2 Stats indexes

```sql
CREATE INDEX idx_country_daily_stats_country_date
  ON stats.country_daily_stats (country_code, stats_date);

CREATE INDEX idx_gk_country_history_active_by_country
  ON stats.gk_country_history (country_code)
  WHERE departed_at IS NULL;

CREATE INDEX idx_gk_country_history_gk_arrived
  ON stats.gk_country_history (geokrety_id, arrived_at DESC);

CREATE INDEX idx_gk_related_users_user
  ON stats.gk_related_users (user_id);

CREATE INDEX idx_user_cache_visits_waypoint
  ON stats.user_cache_visits (waypoint_id, user_id);

CREATE INDEX idx_gk_cache_visits_waypoint
  ON stats.gk_cache_visits (waypoint_id, gk_id);

CREATE INDEX idx_waypoints_country
  ON stats.waypoints (country);
```

### 9.3 Points indexes

```sql
CREATE INDEX idx_gk_moves_points_actor_date
  ON points.gk_moves_points (actor_user_id, calculated_at DESC);

CREATE INDEX idx_user_daily_points_date_total
  ON points.user_daily_points (points_date DESC, total_points_totals DESC);

CREATE INDEX idx_movement_chains_closure_check
  ON points.movement_chains (latest_move_at)
  WHERE chain_status = 'active';

CREATE INDEX idx_chain_members_user
  ON points.movement_chain_members (user_id, chain_id);

CREATE INDEX idx_location_monthly_count_month
  ON points.location_monthly_gk_count (penalty_year_month, user_id);

CREATE INDEX idx_owner_scope_owner_user
  ON points.owner_gk_earning_scope (owner_user_id, user_id);

CREATE INDEX idx_gk_owner_history_gk_time
  ON points.gk_owner_history (geokrety_id, owned_from DESC);
```

## 10. Use Cases and Views (UC1-UC15)

### 10.1 View policy

For UC1, UC3, UC4, UC6, UC7, UC8, UC9, UC10, UC13, UC14, UC15 create explicit views.
UC2 must not rely on direct `gk_moves` scans.

### 10.2 View definitions (logical)

```sql
CREATE VIEW stats.v_uc1_country_activity AS
SELECT country_code,
       SUM(moves_count) AS moves,
       SUM(km_contributed) AS km
FROM stats.country_daily_stats
GROUP BY country_code;

CREATE VIEW stats.v_uc2_user_network AS
SELECT u.user_id,
       u.related_user_id,
       u.shared_geokrety_count,
       u.last_seen_at
FROM stats.user_related_users u;

CREATE VIEW stats.v_uc3_gk_circulation AS
SELECT geokrety_id,
       COUNT(*) AS users,
       SUM(interaction_count) AS interactions
FROM stats.gk_related_users
GROUP BY geokrety_id;

CREATE VIEW stats.v_uc4_user_continent_coverage AS
SELECT uc.user_id,
       cr.continent_code,
       SUM(uc.move_count) AS moves
FROM stats.user_countries uc
JOIN stats.continent_reference cr ON cr.country_alpha2 = uc.country_code
GROUP BY uc.user_id, cr.continent_code;

CREATE VIEW stats.v_uc6_dormancy AS
SELECT geokrety_id,
       MAX(last_interaction) AS last_touch,
       now() - MAX(last_interaction) AS dormancy_interval
FROM stats.gk_related_users
GROUP BY geokrety_id;

CREATE VIEW stats.v_uc7_country_flow AS
SELECT year_month, from_country, to_country, move_count, unique_gk_count
FROM stats.country_pair_flows;

CREATE VIEW stats.v_uc8_seasonal_heatmap AS
SELECT activity_date, hour_utc, move_type, move_count
FROM stats.hourly_activity;

CREATE VIEW stats.v_uc9_multiplier_velocity AS
SELECT gk_id,
       MAX(calculated_at) AS last_change,
       AVG(multiplier_after - multiplier_before) AS avg_delta
FROM points.gk_multiplier_audit
GROUP BY gk_id;

CREATE VIEW stats.v_uc10_cache_popularity AS
SELECT w.waypoint_code,
       SUM(g.visit_count) AS total_gk_visits,
       COUNT(DISTINCT g.gk_id) AS distinct_gks
FROM stats.gk_cache_visits g
JOIN stats.waypoints w ON w.id = g.waypoint_id
GROUP BY w.waypoint_code;

CREATE VIEW stats.v_uc13_gk_timeline AS
SELECT gk_id, event_type, occurred_at, actor_user_id
FROM stats.gk_milestone_events;

CREATE VIEW stats.v_uc14_first_finder_hof AS
SELECT finder_user_id,
       COUNT(*) AS first_finds
FROM stats.first_finder_events
GROUP BY finder_user_id;

CREATE VIEW stats.v_uc15_distance_records AS
SELECT geokret AS gk_id,
       SUM(km_distance) AS km_total
FROM geokrety.gk_moves
WHERE km_distance IS NOT NULL
GROUP BY geokret;
```

UC5 and UC11 are modeled as achievements with optional summary views:

```sql
CREATE VIEW achievements.v_uc5_first_finder_progress AS
SELECT p.user_id, p.progress_value
FROM achievements.user_progress p
JOIN achievements.achievement_definitions d ON d.id = p.achievement_id
WHERE d.code = 'first_finder';

CREATE VIEW achievements.v_uc11_social_hub_progress AS
SELECT p.user_id, p.progress_value
FROM achievements.user_progress p
JOIN achievements.achievement_definitions d ON d.id = p.achievement_id
WHERE d.code = 'social_hub';
```

UC12 correction applied: love is GK-level (`gk_loves`/GK entities), not move-level.

## 11. Snapshot Ingestion and Backfill Helpers

### 11.1 Snapshot order

1. Seed `stats.waypoints` from known waypoint sources.
2. Seed `stats.entity_counters_shard` from source tables.
3. Seed `stats.daily_entity_counts`, `stats.daily_activity`.
4. Seed `stats.user_countries`, `stats.gk_countries_visited`, `stats.country_daily_stats`.
5. Seed `stats.gk_related_users`, `stats.user_related_users`.
6. Backfill `previous_move_id` and `km_distance`.
7. Seed `stats.gk_country_history`.
8. Seed `points` tables via replay.
9. Seed `achievements` progress and awards.

### 11.2 Snapshot ingestion SQL patterns

```sql
-- keyset replay pattern
SELECT id
FROM geokrety.gk_moves
WHERE (moved_on_datetime, id) > (:last_ts, :last_id)
ORDER BY moved_on_datetime, id
LIMIT :batch_size;

-- daily activity seed
INSERT INTO stats.daily_activity (...)
SELECT date_trunc('day', moved_on_datetime)::date AS activity_date,
       COUNT(*) AS total_moves,
       SUM((move_type=0)::int) AS drops,
       SUM((move_type=1)::int) AS grabs,
       SUM((move_type=2)::int) AS comments,
       SUM((move_type=3)::int) AS sees,
       SUM((move_type=4)::int) AS archives,
       SUM((move_type=5)::int) AS dips,
      COALESCE(SUM(km_distance),0) AS km_contributed,
      0 AS points_contributed,
      0 AS loves_count,
      0 AS gk_created,
      0 AS pictures_uploaded_total,
      0 AS pictures_uploaded_avatar,
      0 AS pictures_uploaded_move,
      0 AS pictures_uploaded_user,
      0 AS users_registered
FROM geokrety.gk_moves
GROUP BY 1
ON CONFLICT (activity_date) DO UPDATE SET
  total_moves = EXCLUDED.total_moves;
```

### 11.3 Manual heavy functions that must be explicit

- `stats.fn_backfill_heavy_previous_move_id_all()`
- `stats.fn_backfill_heavy_km_distance_all()`
- `points.fn_replay_points_heavy_all()`

These are intentionally manual and not auto-run in migration.

## 12. Migration and Rollback Runbook

### 12.1 Deploy flow

1. Roll back listed temporary migrations.
2. Apply fresh schema migrations (`stats`, `points`, `achievements`, source column additions, indexes).
3. Create triggers/functions with safe order.
4. Run manual heavy backfills in batches.
5. Run replay for points and achievements.
6. Validate with pgtap and reconciliation queries.
7. Enable scheduled jobs.

### 12.2 Rollback flow

- If migration DDL fails: full transaction rollback.
- If heavy backfill fails: resume using `stats.backfill_progress` cursor.
- If replay output invalid: truncate target replay tables (`points.gk_moves_points`, derived daily tables), fix function, replay again.

### 12.3 Manual function run examples

```sql
SELECT stats.fn_backfill_previous_move_id(NULL, 50000);
SELECT stats.fn_backfill_previous_move_id('[2025-01-01,2025-12-31)'::tstzrange, 50000);
SELECT stats.fn_backfill_heavy_previous_move_id_all();
SELECT stats.fn_backfill_km_distance('[2024-01-01,2024-12-31)'::tstzrange, 50000);
SELECT points.fn_replay_points_for_period('[2020-01-01,2026-01-01)'::tstzrange, 25000);
SELECT points.fn_replay_points_heavy_all();
SELECT achievements.fn_rebuild_progress_for_period(NULL);
```

## 13. pgtap Test Matrix (Large)

Target file family:

- `tests/test-203-schema-core.sql`
- `tests/test-204-previous-location.sql`
- `tests/test-205-points-replay.sql`
- `tests/test-206-achievements.sql`
- `tests/test-207-performance-guards.sql`

| Test ID | Area         | Assertion                                                          |
| ------- | ------------ | ------------------------------------------------------------------ |
| T001    | schema       | `stats`, `points`, `achievements` schemas exist                    |
| T002    | schema       | `entity_counters_shard` columns are exactly `entity, shard, cnt`   |
| T003    | schema       | `entity_counters_shard` PK is `(entity, shard)`                    |
| T004    | schema       | `daily_activity` has `archives` column                             |
| T005    | schema       | `daily_activity` has `km_contributed` numeric                      |
| T006    | schema       | `daily_active_users` PK is `(activity_date, user_id)`              |
| T007    | schema       | `gk_countries_visited.move_count` exists                           |
| T008    | schema       | `gk_country_history` exclusion constraint exists                   |
| T009    | schema       | `waypoints.waypoint_code` unique exists                            |
| T010    | schema       | `gk_cache_visits` PK `(gk_id, waypoint_id)`                        |
| T011    | schema       | `user_cache_visits` PK `(user_id, waypoint_id)`                    |
| T012    | schema       | `gk_related_users` PK `(geokrety_id, user_id)`                     |
| T013    | schema       | `user_related_users` PK `(user_id, related_user_id)`               |
| T014    | schema       | `points.gk_moves_points` PK `(move_id)`                            |
| T015    | schema       | `points.gk_multiplier` check 1.0..2.0 exists                       |
| T016    | schema       | `points.location_monthly_gk_count` PK order is user/location/month |
| T017    | schema       | `points.location_monthly_gk_set` PK includes `gk_id`               |
| T018    | schema       | `movement_chains` partial unique active-per-gk exists              |
| T019    | schema       | `movement_chain_members` PK `(chain_id, user_id)`                  |
| T020    | schema       | `user_gk_chain_bonuses` PK `(user_id, gk_id)`                      |
| T021    | schema       | `achievements.achievement_definitions.code` unique                 |
| T022    | schema       | `achievements.user_progress` composite PK exists                   |
| T023    | schema       | `achievements.user_awards` PK exists                               |
| T024    | trigger      | `fn_set_previous_move_id_and_distance` exists                      |
| T025    | trigger      | previous-move trigger attached to `gk_moves`                       |
| T026    | trigger      | `DROP` with position sets `previous_move_id` correctly             |
| T027    | trigger      | first locatable move has null `previous_move_id`                   |
| T028    | trigger      | `SEEN` without location does not set previous                      |
| T029    | trigger      | `GRAB` with location participates in chain                         |
| T030    | trigger      | `GRAB` without location does not set distance                      |
| T031    | trigger      | `DIP` computes distance from previous locatable move               |
| T032    | trigger      | tie-break by `id` for equal `moved_on_datetime`                    |
| T033    | trigger      | non-locatable move leaves `km_distance` null                       |
| T034    | trigger      | country history insert opens interval                              |
| T035    | trigger      | country change closes old interval                                 |
| T036    | trigger      | same-country move does not create overlap                          |
| T037    | trigger      | user-country upsert increments move_count                          |
| T038    | trigger      | gk-country upsert increments move_count                            |
| T039    | trigger      | waypoint upsert fills unknown lat/lon                              |
| T040    | trigger      | cache visit upsert increments visit_count                          |
| T041    | trigger      | user-cache visit upsert increments visit_count                     |
| T042    | points       | anonymous move awards zero                                         |
| T043    | points       | owner standard move awards zero                                    |
| T044    | points       | owner non-transferable move can award base                         |
| T045    | points       | first user move on GK awards base +3                               |
| T046    | points       | repeated same-user move on GK awards 0 base                        |
| T047    | points       | waypoint-required move without waypoint gives 0 base               |
| T048    | points       | GRAB without waypoint may still award base                         |
| T049    | points       | multiplier used is pre-update value                                |
| T050    | points       | first drop by user adds +0.01 multiplier                           |
| T051    | points       | first grab by user adds +0.01 multiplier                           |
| T052    | points       | first seen by user adds +0.01 multiplier                           |
| T053    | points       | first dip by user adds +0.01 multiplier                            |
| T054    | points       | comment does not increase multiplier                               |
| T055    | points       | country first-visit adds +0.05 multiplier                          |
| T056    | points       | home country does not add +0.05                                    |
| T057    | points       | multiplier ceiling 2.0 enforced                                    |
| T058    | points       | multiplier floor 1.0 enforced                                      |
| T059    | points       | daily hands decay -0.008/day applied                               |
| T060    | points       | weekly cache decay -0.02/week applied                              |
| T061    | points       | multiplier audit row created per change                            |
| T062    | points       | location first distinct GK factor 1.0                              |
| T063    | points       | location second distinct GK factor 0.5                             |
| T064    | points       | location third distinct GK factor 0.25                             |
| T065    | points       | location fourth distinct GK factor 0.0                             |
| T066    | points       | penalty applies only to base component                             |
| T067    | points       | month boundary resets location penalty                             |
| T068    | points       | location key waypoint precedence over coords                       |
| T069    | points       | fallback location key rounds to 4 decimals                         |
| T070    | points       | owner limit allows first 10 owner-GKs                              |
| T071    | points       | owner limit blocks 11th owner-GK                                   |
| T072    | points       | owner limit is per owner, not global                               |
| T073    | points       | previous owner cannot farm base points                             |
| T074    | points       | relay bonus mover +2 rule                                          |
| T075    | points       | relay bonus previous dropper +1 rule                               |
| T076    | points       | rescuer bonus +2 grabber, +1 owner                                 |
| T077    | points       | rescuer not triggered for same holder                              |
| T078    | points       | reach bonus on 10 distinct users                                   |
| T079    | points       | diversity bonus +3 at 5 distinct dropped GKs                       |
| T080    | points       | diversity bonus +7 at 10 owner interactions                        |
| T081    | points       | diversity bonus +5 country diversity monthly                       |
| T082    | chain        | active chain created on first eligible move                        |
| T083    | chain        | member added once per user per chain                               |
| T084    | chain        | self-grab treated as DIP behavior                                  |
| T085    | chain        | comment does not affect timer                                      |
| T086    | chain        | archive closes chain immediately                                   |
| T087    | chain        | inactivity closes chain after 14 days                              |
| T088    | chain        | chain length <3 no bonus                                           |
| T089    | chain        | chain bonus formula `min(n^2,8n)` for n=3                          |
| T090    | chain        | chain bonus formula cap for n=10                                   |
| T091    | chain        | owner gets 25% of distributed chain points                         |
| T092    | chain        | 6-month anti-farming chain bonus lockout                           |
| T093    | achievements | first_finder definition exists                                     |
| T094    | achievements | social_hub definition exists                                       |
| T095    | achievements | progress recompute function updates rows                           |
| T096    | achievements | awards inserted only when threshold crossed                        |
| T097    | achievements | duplicate award prevented for same timestamp tuple                 |
| T098    | uc           | `v_uc1_country_activity` exists                                    |
| T099    | uc           | `v_uc2_user_network` reads no `gk_moves`                           |
| T100    | uc           | `v_uc3_gk_circulation` exists                                      |
| T101    | uc           | `v_uc4_user_continent_coverage` exists                             |
| T102    | uc           | `v_uc6_dormancy` exists                                            |
| T103    | uc           | `v_uc7_country_flow` exists                                        |
| T104    | uc           | `v_uc8_seasonal_heatmap` exists                                    |
| T105    | uc           | `v_uc9_multiplier_velocity` exists                                 |
| T106    | uc           | `v_uc10_cache_popularity` exists                                   |
| T107    | uc           | `v_uc13_gk_timeline` exists                                        |
| T108    | uc           | `v_uc14_first_finder_hof` exists                                   |
| T109    | uc           | `v_uc15_distance_records` exists                                   |
| T110    | replay       | `fn_backfill_previous_move_id` accepts period arg                  |
| T111    | replay       | heavy previous-location function exists                            |
| T112    | replay       | `fn_backfill_km_distance` accepts period arg                       |
| T113    | replay       | heavy km function exists                                           |
| T114    | replay       | points replay period function exists                               |
| T115    | replay       | heavy points replay function exists                                |
| T116    | replay       | replay idempotency on second run                                   |
| T117    | replay       | backfill progress cursor advances                                  |
| T118    | replay       | resume works after forced failure                                  |
| T119    | replay       | job log row inserted per run                                       |
| T120    | replay       | checksum query matches expected sample                             |
| T121    | index        | replay cursor index exists                                         |
| T122    | index        | previous-location partial index exists                             |
| T123    | index        | author-country-time index exists                                   |
| T124    | index        | geokret-country-time index exists                                  |
| T125    | index        | geokret-country-id index exists                                    |
| T126    | index        | country_daily_stats(country,date) exists                           |
| T127    | index        | country history active partial index exists                        |
| T128    | index        | gk_related_users(user_id) index exists                             |
| T129    | index        | user_daily_points(date,total_points_totals) index exists           |
| T130    | index        | movement chain closure partial index exists                        |
| T131    | perf         | explain for v_uc2 uses relation tables                             |
| T132    | perf         | explain for cache popularity avoids full gk_moves                  |
| T133    | perf         | explain for distance records uses km index/filter                  |
| T134    | perf         | explain for leaderboard uses daily points index                    |
| T135    | perf         | explain for country map uses country_daily_stats                   |
| T136    | perf         | explain for active chains uses partial active index                |
| T137    | perf         | explain for previous-location lookup uses partial index            |
| T138    | perf         | explain for owner scope check uses owner index                     |
| T139    | data         | country code normalization to uppercase in stats outputs           |
| T140    | data         | archives are counted in daily and country stats                    |
| T141    | data         | loves tied to GK not moves                                         |
| T142    | data         | first finder requires 7-day cutoff                                 |
| T143    | data         | first finder only one row per GK                                   |
| T144    | data         | non-transferable country bonus owner path works                    |
| T145    | data         | standard owner country bonus owner path works                      |
| T146    | data         | diversity+country bonus stacking works                             |
| T147    | data         | chain closure payout updates user_daily_points                     |
| T148    | data         | chain payout writes move-independent event                         |
| T149    | data         | rescuer event requires cache dormancy threshold                    |
| T150    | data         | seen without location does not join chain                          |
| T151    | data         | seen with location joins chain                                     |
| T152    | data         | dip extends timer limited amount                                   |
| T153    | data         | archive stops chain timer                                          |
| T154    | data         | deleted user logs become anonymous with 0 points                   |
| T155    | data         | location penalty skipped when no waypoint and no coords            |
| T156    | data         | monthly tracker rows are UTC-month anchored                        |
| T157    | data         | owner history open interval uniqueness holds                       |
| T158    | data         | owner history update on owner change trigger                       |
| T159    | data         | UC12 love dashboard uses GK-love tables                            |
| T160    | data         | all helper functions executable by migration role                  |

## 14. Implementation Checklist

- [ ] Create schemas and base tables.
- [ ] Add source-table columns (`previous_move_id`, `km_distance`).
- [ ] Add all required indexes.
- [ ] Create trigger functions and attach in proper order.
- [ ] Seed `stats.waypoints` and counters.
- [ ] Run period backfills then heavy all backfills.
- [ ] Run points replay and achievements rebuild.
- [ ] Run pgtap suite T001-T160.
- [ ] Validate explain plans for key UC views.
- [ ] Pass quality gates: pgTap green, no missing required indexes, replay idempotency verified, key queries under target latency.
- [ ] Enable daily/weekly scheduled jobs.
