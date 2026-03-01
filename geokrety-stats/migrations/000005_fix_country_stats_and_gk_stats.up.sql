-- Migration 000004: Fix move_type mapping in country stats views
-- and add loves_count, status fields to mv_gk_stats
--
-- FIXES:
-- 1. mv_country_stats had drops/grabs swapped (move_type 0=drop, 1=grab)
--    and dips/seen were wrong (5=dip, 3=seen, NOT 3=dip, 4=seen)
-- 2. mv_gk_stats missing loves_count, in_cache, parked, non_collectible
--
-- Canonical move_type mapping (from geokrety.gk_moves):
--   0 = drop (placed in cache)
--   1 = grab (picked up from cache / person)
--   2 = comment
--   3 = seen  (seen it somewhere)
--   4 = archived / dipped  (legacy)
--   5 = dip   (virtual / carrying)

-- ============================================================
-- Fix mv_country_stats: rebuild with correct move_type filters
-- ============================================================
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_country_summary CASCADE;
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_country_stats CASCADE;

CREATE MATERIALIZED VIEW geokrety_stats.mv_country_stats AS
WITH move_countries AS (
    SELECT
        id,
        author,
        geokret,
        move_type,
        COALESCE(country, LAG(country) OVER (PARTITION BY geokret ORDER BY id)) as effective_country
    FROM geokrety.gk_moves
)
SELECT
    m.effective_country as country,
    COUNT(*)                                              AS total_moves,
    COUNT(DISTINCT m.geokret)                             AS unique_gks,
    COUNT(DISTINCT m.author)                              AS unique_users,
    COUNT(*) FILTER (WHERE m.move_type = 0)               AS drops,
    COUNT(*) FILTER (WHERE m.move_type = 1)               AS grabs,
    COUNT(*) FILTER (WHERE m.move_type = 5)               AS dips,
    COUNT(*) FILTER (WHERE m.move_type = 2)               AS comments,
    COUNT(*) FILTER (WHERE m.move_type = 3)               AS seen,
    COALESCE(SUM(CASE WHEN p.gk_id IS NOT NULL THEN p.points ELSE 0 END), 0) AS total_points_awarded
FROM move_countries m
LEFT JOIN geokrety_stats.user_points_log p ON m.id = p.move_id
WHERE m.effective_country IS NOT NULL
GROUP BY m.effective_country;

CREATE INDEX idx_mv_country_stats_country ON geokrety_stats.mv_country_stats(country);
CREATE INDEX idx_mv_country_stats_points  ON geokrety_stats.mv_country_stats(total_points_awarded DESC);

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

CREATE INDEX idx_mv_country_summary_country ON geokrety_stats.mv_country_summary(country);

-- ============================================================
-- Rebuild mv_gk_stats with loves_count, in_cache, status flags
-- ============================================================
DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_gk_stats CASCADE;

CREATE MATERIALIZED VIEW geokrety_stats.mv_gk_stats AS
SELECT
    g.id                                                        AS gk_id,
    g.name,
    g.type                                                      AS gk_type,
    g.missing,
    g.distance,
    g.caches_count,
    g.created_on_datetime,
    g.born_on_datetime,
    owner_u.username                                            AS owner_username,
    owner_u.id                                                  AS owner_id,
    holder_u.username                                           AS holder_username,
    holder_u.id                                                 AS holder_id,

    -- status flags
    (g.holder IS NULL)                                          AS in_cache,
    (g.non_collectible IS NOT NULL)                             AS is_non_collectible,
    (g.parked IS NOT NULL)                                      AS is_parked,
    g.loves_count,

    -- move counts
    COUNT(DISTINCT m.id)                                        AS total_moves,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 0)        AS total_drops,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 1)        AS total_grabs,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 3)        AS total_seen,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 5)        AS total_dips,

    -- participants
    COUNT(DISTINCT m.author)   FILTER (WHERE m.author IS NOT NULL)   AS distinct_users,
    COUNT(DISTINCT m.country)  FILTER (WHERE m.country IS NOT NULL)  AS countries_count,
    COUNT(DISTINCT m.waypoint) FILTER (WHERE m.waypoint IS NOT NULL) AS caches_count_distinct,

    -- points generated
    COALESCE(SUM(pl.points), 0)                                 AS total_points_generated,
    COUNT(DISTINCT pl.user_id)                                  AS users_awarded,

    -- multiplier
    COALESCE(ms.current_multiplier, 1.0)                        AS current_multiplier,

    -- activity
    MIN(m.moved_on_datetime)                                    AS first_move_at,
    MAX(m.moved_on_datetime)                                    AS last_move_at

FROM geokrety.gk_geokrety g
LEFT JOIN geokrety.gk_users owner_u             ON owner_u.id  = g.owner
LEFT JOIN geokrety.gk_users holder_u            ON holder_u.id = g.holder
LEFT JOIN geokrety.gk_moves m                   ON m.geokret   = g.id
LEFT JOIN geokrety_stats.user_points_log pl     ON pl.gk_id    = g.id
LEFT JOIN geokrety_stats.gk_multiplier_state ms ON ms.gk_id    = g.id
GROUP BY g.id, g.name, g.type, g.missing, g.distance, g.caches_count,
         g.created_on_datetime, g.born_on_datetime, owner_u.username, owner_u.id,
         holder_u.username, holder_u.id, ms.current_multiplier,
         g.holder, g.non_collectible, g.parked, g.loves_count
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_gk_stats (gk_id);
CREATE INDEX ON geokrety_stats.mv_gk_stats (owner_id);
CREATE INDEX ON geokrety_stats.mv_gk_stats (total_points_generated DESC);
CREATE INDEX ON geokrety_stats.mv_gk_stats (loves_count DESC);

-- ============================================================
-- Update refresh function to include new views
-- ============================================================
CREATE OR REPLACE FUNCTION geokrety_stats.refresh_leaderboard_views()
RETURNS void LANGUAGE plpgsql AS $$
BEGIN
    -- Core leaderboards (most queried, refresh concurrently)
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_all_time;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_daily;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_monthly;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_yearly;

    -- User and GK stats (concurrent)
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_stats;

    -- User/GK country breakdowns (concurrent)
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_points_daily;

    -- Country and daily activity stats (non-concurrent, no unique index)
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_country_stats;
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_country_summary;
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_daily_activity;

    -- Single-row global stats (non-concurrent, single row so no contention)
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_global_stats;
END;
$$;
