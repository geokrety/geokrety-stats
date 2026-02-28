-- Migration 000002: Materialized views for leaderboard API
-- These views pre-aggregate expensive queries for fast API responses.
-- Run: SELECT refresh_leaderboard_views(); to refresh all.

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
SELECT
    u.id                                                        AS user_id,
    u.username,
    u.home_country,
    u.home_latitude,
    u.home_longitude,
    u.joined_on_datetime,
    u.last_login_datetime,
    COALESCE(t.total_points, 0)                                 AS total_points,

    -- move counts
    COUNT(DISTINCT m.id)                                        AS total_moves,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 0)        AS total_drops,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 1)        AS total_grabs,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 2)        AS total_comments,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 3)        AS total_seen,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 5)        AS total_dips,

    -- GK interactions
    COUNT(DISTINCT m.geokret)                                   AS distinct_gks_interacted,
    COUNT(DISTINCT gk.owner) FILTER (WHERE gk.owner != u.id)   AS distinct_owners,

    -- countries
    COUNT(DISTINCT m.country) FILTER (WHERE m.country IS NOT NULL) AS countries_visited_count,

    -- distance (total km of GKs moved with drops)
    COALESCE(SUM(m.distance) FILTER (WHERE m.move_type = 0), 0) AS km_contributed,

    -- activity
    MIN(m.moved_on_datetime)                                    AS first_move_at,
    MAX(m.moved_on_datetime)                                    AS last_move_at,
    COUNT(DISTINCT DATE(m.moved_on_datetime))                   AS active_days,

    -- points breakdown
    COALESCE(SUM(pl.points) FILTER (WHERE pl.label = 'base_move'), 0)       AS pts_base,
    COALESCE(SUM(pl.points) FILTER (WHERE pl.label = 'relay_bonus'), 0)     AS pts_relay,
    COALESCE(SUM(pl.points) FILTER (WHERE pl.label = 'rescuer_bonus'), 0)   AS pts_rescuer,
    COALESCE(SUM(pl.points) FILTER (WHERE pl.label = 'chain_bonus'), 0)     AS pts_chain,
    COALESCE(SUM(pl.points) FILTER (WHERE pl.label = 'country_bonus'), 0)   AS pts_country,
    COALESCE(SUM(pl.points) FILTER (WHERE pl.label = 'diversity_bonus'), 0) AS pts_diversity,
    COALESCE(SUM(pl.points) FILTER (WHERE pl.label = 'handover_bonus'), 0)  AS pts_handover,
    COALESCE(SUM(pl.points) FILTER (WHERE pl.label = 'reach_bonus'), 0)     AS pts_reach

FROM geokrety.gk_users u
LEFT JOIN geokrety_stats.user_points_totals t       ON t.user_id = u.id
LEFT JOIN geokrety.gk_moves m                       ON m.author  = u.id
LEFT JOIN geokrety.gk_geokrety gk                   ON gk.id     = m.geokret
LEFT JOIN geokrety_stats.user_points_log pl         ON pl.user_id = u.id
GROUP BY u.id, u.username, u.home_country, u.home_latitude, u.home_longitude,
         u.joined_on_datetime, u.last_login_datetime, t.total_points
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_user_stats (user_id);
CREATE INDEX ON geokrety_stats.mv_user_stats (total_points DESC);

-- ============================================================
-- GK STATS SUMMARY: per-geokret stats
-- ============================================================
CREATE MATERIALIZED VIEW geokrety_stats.mv_gk_stats AS
SELECT
    g.id                                                        AS gk_id,
    g.name,
    g.tracking_code,
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
    COUNT(DISTINCT m.id)                                        AS total_moves,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 0)        AS total_drops,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 1)        AS total_grabs,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 3)        AS total_seen,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 5)        AS total_dips,

    -- participants
    COUNT(DISTINCT m.author)   FILTER (WHERE m.author IS NOT NULL) AS distinct_users,
    COUNT(DISTINCT m.country)  FILTER (WHERE m.country IS NOT NULL) AS countries_count,
    COUNT(DISTINCT m.waypoint) FILTER (WHERE m.waypoint IS NOT NULL) AS caches_count_distinct,

    -- points generated
    COALESCE(SUM(pl.points), 0)                                 AS total_points_generated,
    COUNT(DISTINCT pl.user_id)                                  AS users_awarded,

    -- multiplier
    COALESCE(ms.current_multiplier, 1.0)                        AS current_multiplier,

    -- activity
    MIN(m.moved_on_datetime)                                    AS first_move_at,
    MAX(m.moved_on_datetime)                                    AS last_move_at

FROM geokrety.gk_geokrety g
LEFT JOIN geokrety.gk_users owner_u             ON owner_u.id  = g.owner
LEFT JOIN geokrety.gk_users holder_u            ON holder_u.id = g.holder
LEFT JOIN geokrety.gk_moves m                   ON m.geokret   = g.id
LEFT JOIN geokrety_stats.user_points_log pl     ON pl.gk_id    = g.id
LEFT JOIN geokrety_stats.gk_multiplier_state ms ON ms.gk_id    = g.id
GROUP BY g.id, g.name, g.tracking_code, g.type, g.missing, g.distance, g.caches_count,
         g.created_on_datetime, g.born_on_datetime, owner_u.username, owner_u.id,
         holder_u.username, holder_u.id, ms.current_multiplier
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

-- ============================================================
-- REFRESH FUNCTION: refresh all views atomically
-- ============================================================
CREATE OR REPLACE FUNCTION geokrety_stats.refresh_leaderboard_views()
RETURNS void LANGUAGE plpgsql AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_all_time;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_daily;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_monthly;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_yearly;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_points_daily;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_global_stats;
END;
$$;
