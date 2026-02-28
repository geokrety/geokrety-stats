-- Migration 000004: Fix move_type column mappings in country and daily activity views
-- Correct mapping: 0=drop, 1=grab, 2=comment, 3=seen, 4=archived, 5=dip
-- Previous migration 000003 had 0=grab, 1=drop, 3=dip, 4=seen which was incorrect.

-- ── Rebuild mv_country_stats with correct move_type mappings ──────────────
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_country_summary CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_country_stats CASCADE;

CREATE MATERIALIZED VIEW geokrety_stats.mv_country_stats AS
SELECT
    m.country,
    COUNT(*)                                                                    AS total_moves,
    COUNT(DISTINCT m.geokret)                                                   AS unique_gks,
    COUNT(DISTINCT m.author)                                                    AS unique_users,
    COUNT(*) FILTER (WHERE m.move_type = 0)                                    AS drops,
    COUNT(*) FILTER (WHERE m.move_type = 1)                                    AS grabs,
    COUNT(*) FILTER (WHERE m.move_type = 5)                                    AS dips,
    COUNT(*) FILTER (WHERE m.move_type = 2)                                    AS comments,
    COUNT(*) FILTER (WHERE m.move_type = 3)                                    AS seen,
    COALESCE(SUM(CASE WHEN p.gk_id IS NOT NULL THEN p.points ELSE 0 END), 0)  AS total_points_awarded
FROM geokrety.gk_moves m
LEFT JOIN geokrety_stats.user_points_log p
      ON  m.geokret = p.gk_id
      AND m.author  = p.user_id
WHERE m.country IS NOT NULL
GROUP BY m.country;

CREATE INDEX idx_mv_country_stats_country ON geokrety_stats.mv_country_stats(country);

-- ── Rebuild mv_country_summary ────────────────────────────────────────────
CREATE MATERIALIZED VIEW geokrety_stats.mv_country_summary AS
SELECT
    country,
    total_moves,
    unique_gks,
    unique_users,
    drops,
    grabs,
    dips,
    comments,
    seen,
    total_points_awarded
FROM geokrety_stats.mv_country_stats
ORDER BY total_points_awarded DESC;

-- ── Rebuild mv_daily_activity with correct move_type mappings ─────────────
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_daily_activity CASCADE;

CREATE MATERIALIZED VIEW geokrety_stats.mv_daily_activity AS
SELECT
    DATE(moved_on_datetime AT TIME ZONE 'UTC')  AS activity_date,
    COUNT(*)                                    AS total_moves,
    COUNT(DISTINCT author)                      AS active_users,
    COUNT(DISTINCT geokret)                     AS active_gks,
    COUNT(*) FILTER (WHERE move_type = 0)       AS drops,
    COUNT(*) FILTER (WHERE move_type = 1)       AS grabs,
    COUNT(*) FILTER (WHERE move_type = 5)       AS dips,
    COUNT(*) FILTER (WHERE move_type = 2)       AS comments,
    COUNT(*) FILTER (WHERE move_type = 3)       AS seen
FROM geokrety.gk_moves
WHERE moved_on_datetime >= NOW() - INTERVAL '90 days'
GROUP BY DATE(moved_on_datetime AT TIME ZONE 'UTC');

CREATE INDEX idx_mv_daily_activity_date ON geokrety_stats.mv_daily_activity(activity_date);

-- ── Update the refresh function to include these views ────────────────────
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

    -- Country and daily activity stats (non-concurrent, no unique index)
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_country_stats;
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_country_summary;
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_daily_activity;

    -- Single-row global stats (non-concurrent, single row so no contention)
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_global_stats;
END;
$$;
