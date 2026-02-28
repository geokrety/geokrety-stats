-- Create simpler base views first
-- mv_user_related_users: Users who moved same geokrety
CREATE MATERIALIZED VIEW geokrety_stats.mv_user_related_users AS
WITH user_geokrety AS (
    SELECT DISTINCT m1.author as user_id, m1.geokret
    FROM geokrety.gk_moves m1
)
SELECT
    u1.user_id,
    u2.user_id as related_user_id,
    COUNT(DISTINCT u1.geokret) as shared_geokrety_count
FROM user_geokrety u1
INNER JOIN user_geokrety u2 ON u1.geokret = u2.geokret
WHERE u1.user_id != u2.user_id
GROUP BY u1.user_id, u2.user_id;

CREATE INDEX ON geokrety_stats.mv_user_related_users (user_id);
CREATE INDEX ON geokrety_stats.mv_user_related_users (related_user_id);

-- mv_geokrety_related_users: Users who moved specific geokrety
CREATE MATERIALIZED VIEW geokrety_stats.mv_geokrety_related_users AS
SELECT
    m.geokret,
    m.author as user_id,
    COUNT(*) as interaction_count,
    MAX(m.moved_on_datetime) as last_interaction
FROM geokrety.gk_moves m
GROUP BY m.geokret, m.author;

CREATE INDEX ON geokrety_stats.mv_geokrety_related_users (geokret);
CREATE INDEX ON geokrety_stats.mv_geokrety_related_users (user_id);

CREATE INDEX ON geokrety_stats.mv_geokrety_related_users (geokret, interaction_count DESC);
CREATE INDEX ON geokrety_stats.mv_geokrety_related_users (user_id);

-- Update refresh function to include new materialized views
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
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_related_users;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_geokrety_related_users;
    RAISE NOTICE 'All materialized views refreshed successfully at %', NOW();
END;
$$ LANGUAGE plpgsql;
