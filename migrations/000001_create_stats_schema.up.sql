-- Migration 000001: Create geokrety_stats schema and all tables

-- Create the schema
CREATE SCHEMA IF NOT EXISTS geokrety_stats;

-- ============================================================
-- processed_events: idempotency record for move events
-- ============================================================
CREATE TABLE geokrety_stats.processed_events (
    move_id         BIGINT      NOT NULL PRIMARY KEY,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    pipeline_result TEXT        -- 'scored', 'rejected', 'error'
);

COMMENT ON TABLE geokrety_stats.processed_events IS
    'Tracks which gk_moves entries have been processed to ensure exactly-once execution.';

-- ============================================================
-- gk_multiplier_state: per-GK multiplier tracking
-- ============================================================
CREATE TABLE geokrety_stats.gk_multiplier_state (
    gk_id               BIGINT          NOT NULL PRIMARY KEY,
    current_multiplier  DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    last_updated_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    -- who last held the GK and when (for decay calculation)
    current_holder_id   BIGINT,
    holder_acquired_at  TIMESTAMPTZ,
    CONSTRAINT multiplier_min CHECK (current_multiplier >= 1.0),
    CONSTRAINT multiplier_max CHECK (current_multiplier <= 2.0)
);

COMMENT ON TABLE geokrety_stats.gk_multiplier_state IS
    'Tracks the current multiplier for each GeoKret. Used by the points pipeline.';

-- ============================================================
-- gk_countries_visited: per-GK countries visited set
-- ============================================================
CREATE TABLE geokrety_stats.gk_countries_visited (
    gk_id           BIGINT          NOT NULL,
    country_code    CHAR(2)         NOT NULL,  -- ISO 3166-1 alpha-2
    first_visited_at TIMESTAMPTZ   NOT NULL,
    first_move_id   BIGINT,
    PRIMARY KEY (gk_id, country_code)
);

CREATE INDEX idx_gk_countries_gk_id ON geokrety_stats.gk_countries_visited (gk_id);

COMMENT ON TABLE geokrety_stats.gk_countries_visited IS
    'Records all countries a GeoKret has visited. Used for country crossing detection.';

-- ============================================================
-- user_move_history: tracks (user, gk, log_type) tuples
-- Used for first-move detection and multiplier tracking.
-- ============================================================
CREATE TABLE geokrety_stats.user_move_history (
    user_id     BIGINT      NOT NULL,
    gk_id       BIGINT      NOT NULL,
    log_type    SMALLINT    NOT NULL,  -- 0=drop,1=grab,3=seen,5=dip
    first_at    TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, gk_id, log_type)
);

CREATE INDEX idx_user_move_history_user_gk ON geokrety_stats.user_move_history (user_id, gk_id);

COMMENT ON TABLE geokrety_stats.user_move_history IS
    'Records which (user, gk, log_type) combinations have been seen. Used to enforce first-move-only base points.';

-- ============================================================
-- user_owner_gk_counts: tracks per-user, per-owner GK interaction counts
-- Used for the anti-farming 10-GK-per-owner limit.
-- ============================================================
CREATE TABLE geokrety_stats.user_owner_gk_counts (
    user_id     BIGINT  NOT NULL,
    owner_id    BIGINT  NOT NULL,
    gk_id       BIGINT  NOT NULL,
    first_earned_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, owner_id, gk_id)
);

CREATE INDEX idx_user_owner_gk_user_owner ON geokrety_stats.user_owner_gk_counts (user_id, owner_id);

COMMENT ON TABLE geokrety_stats.user_owner_gk_counts IS
    'Tracks distinct GKs per (user, owner) pair where base points were earned. Enforces max-10-per-owner limit.';

-- ============================================================
-- user_waypoint_monthly_counts: tracks per-user, per-waypoint, per-month GK counts
-- Used for waypoint penalty calculation.
-- ============================================================
CREATE TABLE geokrety_stats.user_waypoint_monthly_counts (
    user_id         BIGINT          NOT NULL,
    -- waypoint or NULL (use coordinates as text when no waypoint)
    location_id     VARCHAR(64)     NOT NULL,
    year_month      CHAR(7)         NOT NULL,  -- YYYY-MM
    gk_id           BIGINT          NOT NULL,
    scored_at       TIMESTAMPTZ     NOT NULL,
    PRIMARY KEY (user_id, location_id, year_month, gk_id)
);

CREATE INDEX idx_user_waypoint_user_loc_month
    ON geokrety_stats.user_waypoint_monthly_counts (user_id, location_id, year_month);

COMMENT ON TABLE geokrety_stats.user_waypoint_monthly_counts IS
    'Tracks how many distinct GKs a user has scored at each location per calendar month. Used for waypoint penalty.';

-- ============================================================
-- user_monthly_diversity: tracks monthly diversity bonuses per user
-- ============================================================
CREATE TABLE geokrety_stats.user_monthly_diversity (
    user_id                     BIGINT      NOT NULL,
    year_month                  CHAR(7)     NOT NULL,  -- YYYY-MM pattern
    gks_dropped_count           INTEGER     NOT NULL DEFAULT 0,
    gks_dropped_bonus_awarded   BOOLEAN     NOT NULL DEFAULT FALSE,
    distinct_owners_count       INTEGER     NOT NULL DEFAULT 0,
    distinct_owners_bonus_awarded BOOLEAN   NOT NULL DEFAULT FALSE,
    PRIMARY KEY (user_id, year_month)
);

COMMENT ON TABLE geokrety_stats.user_monthly_diversity IS
    'Monthly diversity tracking per user: drops count and distinct owner count (with bonus flags).';

-- Per (user, month, country) diversity country bonus tracking
CREATE TABLE geokrety_stats.user_monthly_diversity_countries (
    user_id     BIGINT      NOT NULL,
    year_month  CHAR(7)     NOT NULL,
    country     CHAR(2)     NOT NULL,
    awarded_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, year_month, country)
);

-- Per (user, month) distinct owners tracking (which owners counted)
CREATE TABLE geokrety_stats.user_monthly_diversity_owners (
    user_id     BIGINT      NOT NULL,
    year_month  CHAR(7)     NOT NULL,
    owner_id    BIGINT      NOT NULL,
    counted_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, year_month, owner_id)
);

-- Per (user, month) distinct GKs dropped tracking
CREATE TABLE geokrety_stats.user_monthly_diversity_drops (
    user_id     BIGINT      NOT NULL,
    year_month  CHAR(7)     NOT NULL,
    gk_id       BIGINT      NOT NULL,
    dropped_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, year_month, gk_id)
);

-- ============================================================
-- gk_chains: active and completed chains
-- ============================================================
CREATE TABLE geokrety_stats.gk_chains (
    id                  BIGSERIAL   PRIMARY KEY,
    gk_id               BIGINT      NOT NULL,
    status              VARCHAR(16) NOT NULL DEFAULT 'active',  -- 'active', 'ended'
    started_at          TIMESTAMPTZ NOT NULL,
    ended_at            TIMESTAMPTZ,
    chain_last_active   TIMESTAMPTZ NOT NULL,
    holder_acquired_at  TIMESTAMPTZ,
    end_reason          VARCHAR(64), -- 'archived', 'timeout', 'superseded'
    CONSTRAINT chain_status CHECK (status IN ('active', 'ended'))
);

CREATE INDEX idx_gk_chains_gk_active ON geokrety_stats.gk_chains (gk_id, status);

COMMENT ON TABLE geokrety_stats.gk_chains IS
    'Tracks movement chains per GeoKret. One active chain per GK at most.';

-- ============================================================
-- gk_chain_members: ordered members of each chain
-- ============================================================
CREATE TABLE geokrety_stats.gk_chain_members (
    chain_id    BIGINT      NOT NULL REFERENCES geokrety_stats.gk_chains(id) ON DELETE CASCADE,
    user_id     BIGINT      NOT NULL,
    position    INTEGER     NOT NULL,  -- order joined
    joined_at   TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (chain_id, user_id)
);

CREATE INDEX idx_gk_chain_members_chain ON geokrety_stats.gk_chain_members (chain_id);

COMMENT ON TABLE geokrety_stats.gk_chain_members IS
    'Members of each chain, in order of joining.';

-- ============================================================
-- gk_chain_completions: anti-farming record per user per chain
-- ============================================================
CREATE TABLE geokrety_stats.gk_chain_completions (
    user_id         BIGINT      NOT NULL,
    gk_id           BIGINT      NOT NULL,
    chain_id        BIGINT      NOT NULL,
    completed_at    TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, chain_id)
);

CREATE INDEX idx_gk_chain_completions_user_gk ON geokrety_stats.gk_chain_completions (user_id, gk_id, completed_at);

COMMENT ON TABLE geokrety_stats.gk_chain_completions IS
    'Records chain bonus awards per user per chain. Used for anti-farming (6-month cooldown per user per GK).';

-- ============================================================
-- user_points_log: detailed points award ledger
-- ============================================================
CREATE TABLE geokrety_stats.user_points_log (
    id              BIGSERIAL   PRIMARY KEY,
    user_id         BIGINT      NOT NULL,
    points          DOUBLE PRECISION NOT NULL,
    reason          TEXT        NOT NULL,
    label           VARCHAR(64) NOT NULL,   -- machine-readable category
    module_source   VARCHAR(64) NOT NULL,   -- which computer awarded this
    is_owner_reward BOOLEAN     NOT NULL DEFAULT FALSE,
    move_id         BIGINT,                 -- source event (nullable for maintenance jobs)
    gk_id           BIGINT,
    chain_id        BIGINT,
    awarded_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    superseded_by   BIGINT,                 -- for recalculation audit trail
    CONSTRAINT points_non_negative CHECK (points >= 0)
);

CREATE INDEX idx_user_points_log_user_id ON geokrety_stats.user_points_log (user_id);
CREATE INDEX idx_user_points_log_move_id ON geokrety_stats.user_points_log (move_id);
CREATE INDEX idx_user_points_log_awarded_at ON geokrety_stats.user_points_log (awarded_at);

COMMENT ON TABLE geokrety_stats.user_points_log IS
    'Append-only ledger of all points awarded to users. One row per award entry.';

-- ============================================================
-- gk_points_log: detailed GK-level points/multiplier changes log
-- ============================================================
CREATE TABLE geokrety_stats.gk_points_log (
    id                  BIGSERIAL   PRIMARY KEY,
    gk_id               BIGINT      NOT NULL,
    move_id             BIGINT,
    old_multiplier      DOUBLE PRECISION,
    new_multiplier      DOUBLE PRECISION,
    multiplier_delta    DOUBLE PRECISION,
    reason              TEXT,
    module_source       VARCHAR(64),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gk_points_log_gk_id ON geokrety_stats.gk_points_log (gk_id);
CREATE INDEX idx_gk_points_log_move_id ON geokrety_stats.gk_points_log (move_id);

COMMENT ON TABLE geokrety_stats.gk_points_log IS
    'Log of GeoKret multiplier changes with full audit trail.';

-- ============================================================
-- user_points_totals: aggregated points per user
-- ============================================================
CREATE TABLE geokrety_stats.user_points_totals (
    user_id         BIGINT          NOT NULL PRIMARY KEY,
    total_points    DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_updated_at TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    CONSTRAINT total_points_non_negative CHECK (total_points >= 0)
);

COMMENT ON TABLE geokrety_stats.user_points_totals IS
    'Running total of points per user. Updated after each scoring event.';
