DROP TRIGGER IF EXISTS trg_user_owner_counts_summary_sync ON geokrety_stats.user_owner_gk_counts;
DROP FUNCTION IF EXISTS geokrety_stats.tg_sync_user_owner_counts_summary();
DROP TABLE IF EXISTS geokrety_stats.user_owner_counts_summary;
