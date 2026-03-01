DROP MATERIALIZED VIEW IF EXISTS geokrety_stats.mv_user_stats;

CREATE MATERIALIZED VIEW geokrety_stats.mv_user_stats AS
WITH user_moves AS (
    SELECT
        m.author AS user_id,
        COUNT(*)                                              AS total_moves,
        COUNT(*) FILTER (WHERE m.move_type = 0)               AS total_drops,
        COUNT(*) FILTER (WHERE m.move_type = 1)               AS total_grabs,
        COUNT(*) FILTER (WHERE m.move_type = 2)               AS total_comments,
        COUNT(*) FILTER (WHERE m.move_type = 3)               AS total_seen,
        COUNT(*) FILTER (WHERE m.move_type = 4)               AS total_archived,
        COUNT(*) FILTER (WHERE m.move_type = 5)               AS total_dips,
        COUNT(DISTINCT m.geokret)                             AS distinct_gks_interacted,
        COUNT(DISTINCT gk.owner) FILTER (WHERE gk.owner IS NOT NULL AND gk.owner != m.author) AS distinct_owners,
        COUNT(DISTINCT m.country) FILTER (WHERE m.country IS NOT NULL) AS countries_visited_count,
        COALESCE(SUM(m.distance) FILTER (WHERE m.move_type = 0), 0) AS km_contributed,
        MIN(m.moved_on_datetime)                              AS first_move_at,
        MAX(m.moved_on_datetime)                              AS last_move_at,
        COUNT(DISTINCT DATE(m.moved_on_datetime))             AS active_days
    FROM geokrety.gk_moves m
    LEFT JOIN geokrety.gk_geokrety gk ON gk.id = m.geokret
    WHERE m.author IS NOT NULL
    GROUP BY m.author
),
points_by_label AS (
    SELECT
        user_id,
        COALESCE(SUM(points) FILTER (WHERE label = 'base_move'), 0)       AS pts_base,
        COALESCE(SUM(points) FILTER (WHERE label = 'relay_bonus'), 0)     AS pts_relay,
        COALESCE(SUM(points) FILTER (WHERE label = 'rescuer_bonus'), 0)   AS pts_rescuer,
        COALESCE(SUM(points) FILTER (WHERE label = 'chain_bonus'), 0)     AS pts_chain,
        COALESCE(SUM(points) FILTER (WHERE label = 'country_bonus'), 0)   AS pts_country,
        COALESCE(SUM(points) FILTER (WHERE label = 'diversity_bonus'), 0) AS pts_diversity,
        COALESCE(SUM(points) FILTER (WHERE label = 'handover_bonus'), 0)  AS pts_handover,
        COALESCE(SUM(points) FILTER (WHERE label = 'reach_bonus'), 0)     AS pts_reach
    FROM geokrety_stats.user_points_log
    GROUP BY user_id
)
SELECT
    u.id                                                         AS user_id,
    u.username,
    u.home_country,
    u.home_latitude,
    u.home_longitude,
    u.joined_on_datetime,
    COALESCE(t.total_points, 0)                                  AS total_points,
    COALESCE(um.total_moves, 0)                                  AS total_moves,
    COALESCE(um.total_drops, 0)                                  AS total_drops,
    COALESCE(um.total_grabs, 0)                                  AS total_grabs,
    COALESCE(um.total_comments, 0)                               AS total_comments,
    COALESCE(um.total_seen, 0)                                   AS total_seen,
    COALESCE(um.total_dips, 0)                                   AS total_dips,
    COALESCE(um.total_archived, 0)                               AS total_archived,
    COALESCE(um.distinct_gks_interacted, 0)                      AS distinct_gks_interacted,
    COALESCE(um.distinct_owners, 0)                              AS distinct_owners,
    COALESCE(um.countries_visited_count, 0)                      AS countries_visited_count,
    COALESCE(um.km_contributed, 0)                               AS km_contributed,
    um.first_move_at,
    um.last_move_at,
    COALESCE(um.active_days, 0)                                  AS active_days,
    COALESCE(pl.pts_base, 0)                                     AS pts_base,
    COALESCE(pl.pts_relay, 0)                                    AS pts_relay,
    COALESCE(pl.pts_rescuer, 0)                                  AS pts_rescuer,
    COALESCE(pl.pts_chain, 0)                                    AS pts_chain,
    COALESCE(pl.pts_country, 0)                                  AS pts_country,
    COALESCE(pl.pts_diversity, 0)                                AS pts_diversity,
    COALESCE(pl.pts_handover, 0)                                 AS pts_handover,
    COALESCE(pl.pts_reach, 0)                                    AS pts_reach
FROM geokrety.gk_users u
LEFT JOIN geokrety_stats.user_points_totals t ON t.user_id = u.id
LEFT JOIN user_moves um ON um.user_id = u.id
LEFT JOIN points_by_label pl ON pl.user_id = u.id
WITH DATA;

CREATE UNIQUE INDEX ON geokrety_stats.mv_user_stats (user_id);
CREATE INDEX ON geokrety_stats.mv_user_stats (total_points DESC);
