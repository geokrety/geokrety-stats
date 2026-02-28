-- Drop related users materialized views
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_geokrety_related_users;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_user_related_users;

-- Restore original refresh function (without the new views)
DROP FUNCTION geokrety_stats.refresh_materialized_views();

CREATE FUNCTION geokrety_stats.refresh_materialized_views() RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_global_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_geokrety_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_daily_points;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_country_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_bonus_breakdown;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_daily_points;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_geokrety_daily_points;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_year_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_geokrety_year_stats;
    RAISE NOTICE 'All materialized views refreshed successfully at %', NOW();
END;
$$ LANGUAGE plpgsql;
