-- Ensure heavy refresh can run CONCURRENTLY for related-users materialized views.
-- PostgreSQL requires a unique index without WHERE clause on the materialized view.

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_user_related_users_unique_pair
	ON geokrety_stats.mv_user_related_users (user_id, related_user_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_geokrety_related_users_unique_pair
	ON geokrety_stats.mv_geokrety_related_users (geokret, user_id);
