-- Migration 000002: Materialized views for leaderboard API
-- These views pre-aggregate expensive queries for fast API responses.
-- Run: SELECT refresh_leaderboard_views(); to refresh all.

-- ============================================================
-- DROP existing views if they exist (safe for re-runs)
-- ============================================================
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_leaderboard_all_time CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_leaderboard_daily CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_leaderboard_monthly CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_leaderboard_yearly CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_user_stats CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_gk_stats CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_user_points_daily CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_user_countries CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_gk_countries CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_global_stats CASCADE;
DROP FUNCTION IF EXISTS geokrety_stats.refresh_leaderboard_views CASCADE;

-- ============================================================
-- LEADERBOARD: all-time top users
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_leaderboard_all_time AS
SELECT
    u.id                                        AS user_id,
    u.username,
    u.home_country,
    COALESCE(t.total_points, 0)                 AS total_points,
    RANK() OVER (ORDER BY COALESCE(t.total_points, 0) DESC) AS rank,
    COUNT(DISTINCT m.geokret)                   AS distinct_gks,
    COUNT(DISTINCT m.id)                        AS total_moves,
    MAX(m.moved_on_datetime)                    AS last_active
FROM geokrety.gk_users u
LEFT JOIN geokrety_stats.user_points_totals t   ON t.user_id = u.id
LEFT JOIN geokrety.gk_moves m                   ON m.author  = u.id
WHERE COALESCE(t.total_points, 0) > 0
GROUP BY u.id, u.username, u.home_country, t.total_points
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_leaderboard_all_time (user_id);
CREATE INDEX ON geokrety_stats.mv_leaderboard_all_time (rank);
CREATE INDEX ON geokrety_stats.mv_leaderboard_all_time (total_points DESC);

-- ============================================================
-- LEADERBOARD: points earned per calendar day (rolling)
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_leaderboard_daily AS
SELECT
    u.id                        AS user_id,
    u.username,
    l.day,
    SUM(l.points)               AS points_day,
    RANK() OVER (PARTITION BY l.day ORDER BY SUM(l.points) DESC) AS rank
FROM geokrety.gk_users u
JOIN (
    SELECT user_id,
           DATE(awarded_at AT TIME ZONE 'UTC') AS day,
           SUM(points) AS points
    FROM geokrety_stats.user_points_log
    GROUP BY user_id, DATE(awarded_at AT TIME ZONE 'UTC')
) l ON l.user_id = u.id
GROUP BY u.id, u.username, l.day
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_leaderboard_daily (user_id, day);
CREATE INDEX ON geokrety_stats.mv_leaderboard_daily (day DESC);
CREATE INDEX ON geokrety_stats.mv_leaderboard_daily (day DESC, rank);

-- ============================================================
-- LEADERBOARD: points by month
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_leaderboard_monthly AS
SELECT
    u.id                        AS user_id,
    u.username,
    l.year_month,
    SUM(l.points)               AS points_month,
    RANK() OVER (PARTITION BY l.year_month ORDER BY SUM(l.points) DESC) AS rank
FROM geokrety.gk_users u
JOIN (
    SELECT user_id,
           TO_CHAR(awarded_at AT TIME ZONE 'UTC', 'YYYY-MM') AS year_month,
           SUM(points) AS points
    FROM geokrety_stats.user_points_log
    GROUP BY user_id, TO_CHAR(awarded_at AT TIME ZONE 'UTC', 'YYYY-MM')
) l ON l.user_id = u.id
GROUP BY u.id, u.username, l.year_month
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_leaderboard_monthly (user_id, year_month);
CREATE INDEX ON geokrety_stats.mv_leaderboard_monthly (year_month DESC);
CREATE INDEX ON geokrety_stats.mv_leaderboard_monthly (year_month DESC, rank);

-- ============================================================
-- LEADERBOARD: points by year
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_leaderboard_yearly AS
SELECT
    u.id                        AS user_id,
    u.username,
    l.year,
    SUM(l.points)               AS points_year,
    RANK() OVER (PARTITION BY l.year ORDER BY SUM(l.points) DESC) AS rank
FROM geokrety.gk_users u
JOIN (
    SELECT user_id,
           EXTRACT(YEAR FROM awarded_at AT TIME ZONE 'UTC')::int AS year,
           SUM(points) AS points
    FROM geokrety_stats.user_points_log
    GROUP BY user_id, EXTRACT(YEAR FROM awarded_at AT TIME ZONE 'UTC')::int
) l ON l.user_id = u.id
GROUP BY u.id, u.username, l.year
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_leaderboard_yearly (user_id, year);
CREATE INDEX ON geokrety_stats.mv_leaderboard_yearly (year DESC);
CREATE INDEX ON geokrety_stats.mv_leaderboard_yearly (year DESC, rank);

-- ============================================================
-- USER STATS SUMMARY: comprehensive per-user stats
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_user_stats AS
WITH user_moves AS (
    SELECT
        m.author AS user_id,
        COUNT(*)                                              AS total_moves,
        COUNT(*) FILTER (WHERE m.move_type = 0)               AS total_drops,
        COUNT(*) FILTER (WHERE m.move_type = 1)               AS total_grabs,
        COUNT(*) FILTER (WHERE m.move_type = 2)               AS total_comments,
        COUNT(*) FILTER (WHERE m.move_type = 3)               AS total_seen,
        COUNT(*) FILTER (WHERE m.move_type = 4)               AS total_archived,
        COUNT(*) FILTER (WHERE m.move_type = 5)               AS total_dips,
        COUNT(DISTINCT m.geokret)                             AS distinct_gks_interacted,
        COUNT(DISTINCT gk.owner) FILTER (WHERE gk.owner IS NOT NULL AND gk.owner != m.author) AS distinct_owners,
        COUNT(DISTINCT m.country) FILTER (WHERE m.country IS NOT NULL) AS countries_visited_count,
        COALESCE(SUM(m.distance) FILTER (WHERE m.move_type = 0), 0) AS km_contributed,
        MIN(m.moved_on_datetime)                             AS first_move_at,
        MAX(m.moved_on_datetime)                             AS last_move_at,
        COUNT(DISTINCT DATE(m.moved_on_datetime))             AS active_days
    FROM geokrety.gk_moves m
    LEFT JOIN geokrety.gk_geokrety gk ON gk.id = m.geokret
    WHERE m.author IS NOT NULL
    GROUP BY m.author
),
points_by_label AS (
    SELECT
        user_id,
        COALESCE(SUM(points) FILTER (WHERE label = 'base_move'), 0)       AS pts_base,
        COALESCE(SUM(points) FILTER (WHERE label = 'relay_bonus'), 0)     AS pts_relay,
        COALESCE(SUM(points) FILTER (WHERE label = 'rescuer_bonus'), 0)   AS pts_rescuer,
        COALESCE(SUM(points) FILTER (WHERE label = 'chain_bonus'), 0)     AS pts_chain,
        COALESCE(SUM(points) FILTER (WHERE label = 'country_bonus'), 0)   AS pts_country,
        COALESCE(SUM(points) FILTER (WHERE label = 'diversity_bonus'), 0) AS pts_diversity,
        COALESCE(SUM(points) FILTER (WHERE label = 'handover_bonus'), 0)  AS pts_handover,
        COALESCE(SUM(points) FILTER (WHERE label = 'reach_bonus'), 0)     AS pts_reach
    FROM geokrety_stats.user_points_log
    GROUP BY user_id
)
SELECT
    u.id                                                        AS user_id,
    u.username,
    u.home_country,
    u.home_latitude,
    u.home_longitude,
    u.joined_on_datetime,
    COALESCE(t.total_points, 0)                                 AS total_points,
    COALESCE(um.total_moves, 0)                                 AS total_moves,
    COALESCE(um.total_drops, 0)                                 AS total_drops,
    COALESCE(um.total_grabs, 0)                                 AS total_grabs,
    COALESCE(um.total_comments, 0)                              AS total_comments,
    COALESCE(um.total_seen, 0)                                  AS total_seen,
    COALESCE(um.total_dips, 0)                                  AS total_dips,
    COALESCE(um.total_archived, 0)                              AS total_archived,
    COALESCE(um.distinct_gks_interacted, 0)                      AS distinct_gks_interacted,
    COALESCE(um.distinct_owners, 0)                              AS distinct_owners,
    COALESCE(um.countries_visited_count, 0)                      AS countries_visited_count,
    COALESCE(um.km_contributed, 0)                               AS km_contributed,
    um.first_move_at,
    um.last_move_at,
    COALESCE(um.active_days, 0)                                  AS active_days,
    COALESCE(pl.pts_base, 0)                                     AS pts_base,
    COALESCE(pl.pts_relay, 0)                                    AS pts_relay,
    COALESCE(pl.pts_rescuer, 0)                                  AS pts_rescuer,
    COALESCE(pl.pts_chain, 0)                                    AS pts_chain,
    COALESCE(pl.pts_country, 0)                                  AS pts_country,
    COALESCE(pl.pts_diversity, 0)                                AS pts_diversity,
    COALESCE(pl.pts_handover, 0)                                 AS pts_handover,
    COALESCE(pl.pts_reach, 0)                                    AS pts_reach
FROM geokrety.gk_users u
LEFT JOIN geokrety_stats.user_points_totals t       ON t.user_id = u.id
LEFT JOIN user_moves um                            ON um.user_id = u.id
LEFT JOIN points_by_label pl                       ON pl.user_id = u.id
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_user_stats (user_id);
CREATE INDEX ON geokrety_stats.mv_user_stats (total_points DESC);

-- ============================================================
-- GK STATS SUMMARY: per-geokret stats
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_gk_stats AS
WITH move_stats AS (
    SELECT
        m.geokret                                           AS gk_id,
        COUNT(*)                                             AS total_moves,
        COUNT(*) FILTER (WHERE m.move_type = 0)             AS total_drops,
        COUNT(*) FILTER (WHERE m.move_type = 1)             AS total_grabs,
        COUNT(*) FILTER (WHERE m.move_type = 3)             AS total_seen,
        COUNT(*) FILTER (WHERE m.move_type = 5)             AS total_dips,
        COUNT(DISTINCT m.author) FILTER (WHERE m.author IS NOT NULL) AS distinct_users,
        COUNT(DISTINCT m.country) FILTER (WHERE m.country IS NOT NULL) AS countries_count,
        COUNT(DISTINCT m.waypoint) FILTER (WHERE m.waypoint IS NOT NULL) AS caches_count_distinct,
        MIN(m.moved_on_datetime)                            AS first_move_at,
        MAX(m.moved_on_datetime)                            AS last_move_at
    FROM geokrety.gk_moves m
    GROUP BY m.geokret
),
points_summary AS (
    SELECT
        gk_id,
        COALESCE(SUM(points), 0)                              AS total_points_generated,
        COUNT(DISTINCT user_id)                               AS users_awarded
    FROM geokrety_stats.user_points_log
    GROUP BY gk_id
)
SELECT
    g.id                                                        AS gk_id,
    g.name,
    g.type                                                      AS gk_type,
    g.missing,
    g.distance,
    g.caches_count,
    g.created_on_datetime,
    g.born_on_datetime,
    owner_u.username                                            AS owner_username,
    owner_u.id                                                  AS owner_id,
    holder_u.username                                           AS holder_username,
    holder_u.id                                                 AS holder_id,

    -- move counts
    COALESCE(move_stats.total_moves, 0)                         AS total_moves,
    COALESCE(move_stats.total_drops, 0)                         AS total_drops,
    COALESCE(move_stats.total_grabs, 0)                         AS total_grabs,
    COALESCE(move_stats.total_seen, 0)                          AS total_seen,
    COALESCE(move_stats.total_dips, 0)                          AS total_dips,

    -- participants
    COALESCE(move_stats.distinct_users, 0)                      AS distinct_users,
    COALESCE(move_stats.countries_count, 0)                     AS countries_count,
    COALESCE(move_stats.caches_count_distinct, 0)               AS caches_count_distinct,

    -- points generated
    COALESCE(points_summary.total_points_generated, 0)          AS total_points_generated,
    COALESCE(points_summary.users_awarded, 0)                   AS users_awarded,

    -- multiplier
    COALESCE(ms.current_multiplier, 1.0)                        AS current_multiplier,

    -- activity
    move_stats.first_move_at,
    move_stats.last_move_at

FROM geokrety.gk_geokrety g
LEFT JOIN geokrety.gk_users owner_u             ON owner_u.id  = g.owner
LEFT JOIN geokrety.gk_users holder_u            ON holder_u.id = g.holder
LEFT JOIN move_stats                          ON move_stats.gk_id = g.id
LEFT JOIN points_summary                      ON points_summary.gk_id = g.id
LEFT JOIN geokrety_stats.gk_multiplier_state ms ON ms.gk_id    = g.id
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_gk_stats (gk_id);
CREATE INDEX ON geokrety_stats.mv_gk_stats (owner_id);
CREATE INDEX ON geokrety_stats.mv_gk_stats (total_points_generated DESC);

-- ============================================================
-- USER POINTS TIMELINE: per-user points per day (for charts)
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_user_points_daily AS
SELECT
    user_id,
    DATE(awarded_at AT TIME ZONE 'UTC')         AS day,
    SUM(points)                                 AS points,
    COUNT(*)                                    AS awards_count,
    COUNT(DISTINCT move_id)                     AS moves_count
FROM geokrety_stats.user_points_log
GROUP BY user_id, DATE(awarded_at AT TIME ZONE 'UTC')
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_user_points_daily (user_id, day);
CREATE INDEX ON geokrety_stats.mv_user_points_daily (user_id, day DESC);

-- ============================================================
-- USER COUNTRIES: per-user countries visited
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_user_countries AS
SELECT
    m.author                    AS user_id,
    m.country,
    COUNT(*)                    AS move_count,
    MIN(m.moved_on_datetime)    AS first_visit,
    MAX(m.moved_on_datetime)    AS last_visit
FROM geokrety.gk_moves m
WHERE m.country IS NOT NULL
  AND m.author  IS NOT NULL
GROUP BY m.author, m.country
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_user_countries (user_id, country);
CREATE INDEX ON geokrety_stats.mv_user_countries (user_id);

-- ============================================================
-- GK COUNTRIES: per-gk countries visited
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_gk_countries AS
SELECT
    cv.gk_id,
    cv.country_code             AS country,
    cv.first_visited_at,
    cv.first_move_id
FROM geokrety_stats.gk_countries_visited cv
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_gk_countries (gk_id, country);
CREATE INDEX ON geokrety_stats.mv_gk_countries (gk_id);

-- ============================================================
-- GLOBAL STATS: site-wide counters
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_global_stats AS
SELECT
    (SELECT COUNT(*) FROM geokrety.gk_users)                        AS total_users,
    (SELECT COUNT(*) FROM geokrety.gk_geokrety)                     AS total_gks,
    (SELECT COUNT(*) FROM geokrety.gk_moves)                        AS total_moves,
    (SELECT COUNT(*) FROM geokrety_stats.user_points_totals WHERE total_points > 0) AS scored_users,
    (SELECT COALESCE(SUM(total_points), 0) FROM geokrety_stats.user_points_totals)  AS total_points_awarded,
    (SELECT COUNT(DISTINCT country) FROM geokrety.gk_moves WHERE country IS NOT NULL) AS countries_reached,
    (SELECT COALESCE(SUM(distance), 0) FROM geokrety.gk_geokrety)   AS total_km,
    NOW()                                                            AS computed_at
WITH DATA;

-- Unique index for mv_global_stats (required for concurrent refresh)
CREATE UNIQUE INDEX mv_global_stats_uniq_idx ON geokrety_stats.mv_global_stats ((1));

-- ============================================================
-- REFRESH FUNCTION: refresh all views with intelligent timing
-- ============================================================
CREATE OR REPLACE FUNCTION geokrety_stats.refresh_leaderboard_views()
RETURNS void LANGUAGE plpgsql AS $$
BEGIN
    -- Core leaderboards (most queried, refresh concurrently)
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_all_time;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_daily;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_monthly;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_yearly;

    -- User and GK stats (concurrent)
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_stats;

    -- User/GK country breakdowns (concurrent)
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_points_daily;

    -- Single-row global stats (non-concurrent, single row so no contention)
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_global_stats;
END;
$$;
