-- Add operational refresh profiles and context-cache validation helper.

CREATE OR REPLACE FUNCTION geokrety_stats.refresh_leaderboard_views_light()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_all_time;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_daily;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_monthly;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_leaderboard_yearly;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_points_daily;
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_global_stats;
END;
$$;

CREATE OR REPLACE FUNCTION geokrety_stats.refresh_leaderboard_views_heavy()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_gk_countries;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_user_related_users;
    REFRESH MATERIALIZED VIEW CONCURRENTLY geokrety_stats.mv_geokrety_related_users;
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_country_stats;
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_country_summary;
    REFRESH MATERIALIZED VIEW geokrety_stats.mv_daily_activity;
END;
$$;

CREATE OR REPLACE FUNCTION geokrety_stats.validate_gk_context_cache()
RETURNS TABLE (
    checked_rows bigint,
    bad_home_country_null bigint,
    bad_prev_holder_null bigint,
    bad_last_drop_null bigint,
    bad_last_drop_user_null bigint,
    bad_last_cache_entry_null bigint,
    home_country_mismatch bigint,
    previous_holder_mismatch bigint,
    last_drop_at_mismatch bigint,
    last_drop_user_mismatch bigint,
    last_cache_entry_mismatch bigint,
    missing_cache_rows bigint
)
LANGUAGE sql
AS $$
WITH src AS (
    SELECT
        m.geokret AS gk_id,
        COUNT(*) FILTER (WHERE m.country IS NOT NULL AND m.move_type IN (0,3,5)) AS country_events,
        COUNT(*) FILTER (WHERE m.move_type = 1 AND m.author IS NOT NULL) AS grabs_with_author,
        COUNT(*) FILTER (WHERE m.move_type = 0 AND m.author IS NOT NULL) AS drops_with_author,
        COUNT(*) FILTER (WHERE m.move_type IN (0,3)) AS cache_entries
    FROM geokrety.gk_moves m
    GROUP BY m.geokret
),
home_country AS (
    SELECT DISTINCT ON (m.geokret)
        m.geokret AS gk_id,
        m.country::char(2) AS expected_home_country
    FROM geokrety.gk_moves m
    WHERE m.country IS NOT NULL
      AND m.move_type IN (0,3,5)
    ORDER BY m.geokret, m.moved_on_datetime ASC, m.id ASC
),
prev_holder AS (
    SELECT y.geokret AS gk_id, y.author AS expected_previous_holder_id
    FROM (
        SELECT
            m.geokret,
            m.author,
            ROW_NUMBER() OVER (PARTITION BY m.geokret ORDER BY m.moved_on_datetime DESC, m.id DESC) AS rn
        FROM geokrety.gk_moves m
        WHERE m.move_type = 1
          AND m.author IS NOT NULL
    ) y
    WHERE y.rn = 2
),
last_drop AS (
    SELECT DISTINCT ON (m.geokret)
        m.geokret AS gk_id,
        m.moved_on_datetime AS expected_last_drop_at,
        m.author AS expected_last_drop_user_id
    FROM geokrety.gk_moves m
    WHERE m.move_type = 0
      AND m.author IS NOT NULL
    ORDER BY m.geokret, m.moved_on_datetime DESC, m.id DESC
),
last_cache AS (
    SELECT DISTINCT ON (m.geokret)
        m.geokret AS gk_id,
        m.moved_on_datetime AS expected_last_cache_entry_at
    FROM geokrety.gk_moves m
    WHERE m.move_type IN (0,3)
    ORDER BY m.geokret, m.moved_on_datetime DESC, m.id DESC
),
joined AS (
    SELECT
        c.*,
        s.country_events,
        s.grabs_with_author,
        s.drops_with_author,
        s.cache_entries,
        hc.expected_home_country,
        ph.expected_previous_holder_id,
        ld.expected_last_drop_at,
        ld.expected_last_drop_user_id,
        lc.expected_last_cache_entry_at
    FROM geokrety_stats.gk_context_cache c
    JOIN src s ON s.gk_id = c.gk_id
    LEFT JOIN home_country hc ON hc.gk_id = c.gk_id
    LEFT JOIN prev_holder ph ON ph.gk_id = c.gk_id
    LEFT JOIN last_drop ld ON ld.gk_id = c.gk_id
    LEFT JOIN last_cache lc ON lc.gk_id = c.gk_id
)
SELECT
    (SELECT COUNT(*) FROM geokrety_stats.gk_context_cache) AS checked_rows,
    COUNT(*) FILTER (WHERE home_country IS NULL AND country_events > 0) AS bad_home_country_null,
    COUNT(*) FILTER (WHERE previous_holder_id IS NULL AND grabs_with_author >= 2) AS bad_prev_holder_null,
    COUNT(*) FILTER (WHERE last_drop_at IS NULL AND drops_with_author > 0) AS bad_last_drop_null,
    COUNT(*) FILTER (WHERE last_drop_user_id IS NULL AND drops_with_author > 0) AS bad_last_drop_user_null,
    COUNT(*) FILTER (WHERE last_cache_entry_at IS NULL AND cache_entries > 0) AS bad_last_cache_entry_null,
    COUNT(*) FILTER (WHERE home_country IS DISTINCT FROM expected_home_country) AS home_country_mismatch,
    COUNT(*) FILTER (WHERE previous_holder_id IS DISTINCT FROM expected_previous_holder_id) AS previous_holder_mismatch,
    COUNT(*) FILTER (WHERE last_drop_at IS DISTINCT FROM expected_last_drop_at) AS last_drop_at_mismatch,
    COUNT(*) FILTER (WHERE last_drop_user_id IS DISTINCT FROM expected_last_drop_user_id) AS last_drop_user_mismatch,
    COUNT(*) FILTER (WHERE last_cache_entry_at IS DISTINCT FROM expected_last_cache_entry_at) AS last_cache_entry_mismatch,
    (
        SELECT COUNT(*)
        FROM (
            SELECT DISTINCT geokret AS gk_id
            FROM geokrety.gk_moves
            WHERE geokret IS NOT NULL
        ) x
        LEFT JOIN geokrety_stats.gk_context_cache c ON c.gk_id = x.gk_id
        WHERE c.gk_id IS NULL
    ) AS missing_cache_rows
FROM joined;
$$;
