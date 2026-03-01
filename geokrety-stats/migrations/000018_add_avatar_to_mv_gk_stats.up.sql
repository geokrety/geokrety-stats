DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_gk_stats;

CREATE MATERIALIZED VIEW geokrety_stats.mv_gk_stats AS
SELECT
    g.id                                                        AS gk_id,
    g.name,
    g.avatar,
    pic.bucket                                                  AS avatar_bucket,
    pic.key                                                     AS avatar_key,
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

    (g.holder IS NULL)                                          AS in_cache,
    (g.non_collectible IS NOT NULL)                             AS is_non_collectible,
    (g.parked IS NOT NULL)                                      AS is_parked,
    g.loves_count,

    COUNT(DISTINCT m.id)                                        AS total_moves,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 0)         AS total_drops,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 1)         AS total_grabs,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 2)         AS total_comments,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 3)         AS total_seen,
    COUNT(DISTINCT m.id) FILTER (WHERE m.move_type = 5)         AS total_dips,

    COUNT(DISTINCT m.author)   FILTER (WHERE m.author IS NOT NULL)   AS distinct_users,
    COUNT(DISTINCT m.country)  FILTER (WHERE m.country IS NOT NULL)  AS countries_count,
    COUNT(DISTINCT m.waypoint) FILTER (WHERE m.waypoint IS NOT NULL) AS caches_count_distinct,

    COALESCE(SUM(pl.points), 0)                                 AS total_points_generated,
    COUNT(DISTINCT pl.user_id)                                  AS users_awarded,

    COALESCE(ms.current_multiplier, 1.0)                        AS current_multiplier,

    MIN(m.moved_on_datetime)                                    AS first_move_at,
    MAX(m.moved_on_datetime)                                    AS last_move_at

FROM geokrety.gk_geokrety g
LEFT JOIN geokrety.gk_pictures pic           ON pic.id      = g.avatar
LEFT JOIN geokrety.gk_users owner_u          ON owner_u.id  = g.owner
LEFT JOIN geokrety.gk_users holder_u         ON holder_u.id = g.holder
LEFT JOIN geokrety.gk_moves m                ON m.geokret   = g.id
LEFT JOIN geokrety_stats.user_points_log pl  ON pl.gk_id    = g.id
LEFT JOIN geokrety_stats.gk_multiplier_state ms ON ms.gk_id = g.id
GROUP BY g.id, g.name, g.avatar, pic.bucket, pic.key, g.type, g.missing, g.distance, g.caches_count,
         g.created_on_datetime, g.born_on_datetime, owner_u.username, owner_u.id,
         holder_u.username, holder_u.id, ms.current_multiplier,
         g.holder, g.non_collectible, g.parked, g.loves_count
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_gk_stats (gk_id);
CREATE INDEX ON geokrety_stats.mv_gk_stats (owner_id);
CREATE INDEX ON geokrety_stats.mv_gk_stats (total_points_generated DESC);
CREATE INDEX ON geokrety_stats.mv_gk_stats (loves_count DESC);
