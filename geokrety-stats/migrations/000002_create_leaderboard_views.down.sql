-- Migration 000002: Drop leaderboard materialized views

DROP FUNCTION IF EXISTS geokrety_stats.refresh_leaderboard_views();
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_global_stats;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_gk_countries;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_user_countries;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_user_points_daily;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_gk_stats;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_user_stats;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_leaderboard_yearly;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_leaderboard_monthly;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_leaderboard_daily;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_leaderboard_all_time;
