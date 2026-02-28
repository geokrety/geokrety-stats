-- Rollback migration 000004: Restore old (incorrect) move_type mappings
-- WARNING: Restores incorrect mappings from migration 000003
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_country_summary CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_country_stats CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_daily_activity CASCADE;
