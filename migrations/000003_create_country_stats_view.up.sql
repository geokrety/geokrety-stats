-- Create materialized view for country statistics
-- This replaces direct aggregation on gk_moves table (which is huge)
-- move_type: 0=drop, 1=grab, 2=comment, 3=seen, 4=archived, 5=dip
CREATE MATERIALIZED VIEW geokrety_stats.mv_country_stats AS
SELECT
    m.country,
    COUNT(*) AS total_moves,
    COUNT(DISTINCT m.geokret) AS unique_gks,
    COUNT(DISTINCT m.author) AS unique_users,
    COUNT(*) FILTER (WHERE m.move_type = 0) AS drops,
    COUNT(*) FILTER (WHERE m.move_type = 1) AS grabs,
    COUNT(*) FILTER (WHERE m.move_type = 5) AS dips,
    COUNT(*) FILTER (WHERE m.move_type = 2) AS comments,
    COUNT(*) FILTER (WHERE m.move_type = 3) AS seen,
    COALESCE(SUM(CASE WHEN p.gk_id IS NOT NULL THEN p.points ELSE 0 END), 0) AS total_points_awarded
FROM geokrety.gk_moves m
LEFT JOIN geokrety_stats.user_points_log p ON m.geokret = p.gk_id AND m.author = p.user_id
WHERE m.country IS NOT NULL
GROUP BY m.country;

CREATE INDEX idx_mv_country_stats_country ON geokrety_stats.mv_country_stats(country);

-- Aggregate by country only (without move_type breakdown)
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

-- Daily activity aggregation (more efficient than direct query)
CREATE MATERIALIZED VIEW geokrety_stats.mv_daily_activity AS
SELECT
    DATE(moved_on_datetime AT TIME ZONE 'UTC') AS activity_date,
    COUNT(*) AS total_moves,
    COUNT(DISTINCT author) AS active_users,
    COUNT(DISTINCT geokret) AS active_gks,
    COUNT(*) FILTER (WHERE move_type = 0) AS drops,
    COUNT(*) FILTER (WHERE move_type = 1) AS grabs,
    COUNT(*) FILTER (WHERE move_type = 5) AS dips,
    COUNT(*) FILTER (WHERE move_type = 2) AS comments,
    COUNT(*) FILTER (WHERE move_type = 3) AS seen
FROM geokrety.gk_moves
GROUP BY DATE(moved_on_datetime AT TIME ZONE 'UTC');

CREATE INDEX idx_mv_daily_activity_date ON geokrety_stats.mv_daily_activity(activity_date);
