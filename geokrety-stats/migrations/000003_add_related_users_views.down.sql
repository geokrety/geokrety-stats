-- Drop related users materialized views
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_geokrety_related_users;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_user_related_users;

-- Restore original refresh function (without the new views)
DROP FUNCTION IF EXISTS geokrety_stats.refresh_leaderboard_views();

CREATE FUNCTION geokrety_stats.refresh_leaderboard_views() RETURNS void AS $$
BEGIN
    -- Core leaderboards
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_all_time;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_daily;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_monthly;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_yearly;

    -- User and GK stats
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_stats;

    -- User/GK country breakdowns
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_points_daily;

    -- Global stats
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_global_stats;
END;
$$ LANGUAGE plpgsql;
