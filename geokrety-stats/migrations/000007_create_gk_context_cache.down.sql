DROP TRIGGER IF EXISTS trg_gk_moves_refresh_gk_context_cache ON geokrety.gk_moves;
DROP FUNCTION IF EXISTS geokrety_stats.tg_refresh_gk_context_cache();
DROP FUNCTION IF EXISTS geokrety_stats.refresh_gk_context_cache(bigint);
DROP TABLE IF EXISTS geokrety_stats.gk_context_cache;
