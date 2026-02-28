-- Extra indexes for high-volume replay joins and leaderboard aggregations.

-- Reach bonus query joins user_points_log on move_id and filters points > 0.
CREATE INDEX IF NOT EXISTS idx_user_points_log_move_id_points_pos
    ON geokrety_stats.user_points_log (move_id)
    WHERE points > 0;

-- Frequent per-user time-window aggregations.
CREATE INDEX IF NOT EXISTS idx_user_points_log_user_awarded_desc
    ON geokrety_stats.user_points_log (user_id, awarded_at DESC);

-- GK-centric aggregations for mv_gk_stats and related reports.
CREATE INDEX IF NOT EXISTS idx_user_points_log_gk_awarded_desc
    ON geokrety_stats.user_points_log (gk_id, awarded_at DESC)
    WHERE gk_id IS NOT NULL;
