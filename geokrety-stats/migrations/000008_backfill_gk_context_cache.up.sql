-- Manual backfill function for geokrety_stats.gk_context_cache
-- Usage (manual, one-shot or periodic):
--   SELECT geokrety_stats.backfill_gk_context_cache();
--
-- This is intentionally manual to avoid long migration runtime during deploy.

CREATE OR REPLACE FUNCTION geokrety_stats.backfill_gk_context_cache()
RETURNS bigint
LANGUAGE plpgsql
AS $$
DECLARE
    _affected bigint := 0;
BEGIN
    WITH gk_ids AS (
        SELECT DISTINCT geokret AS gk_id
        FROM geokrety.gk_moves
        WHERE geokret IS NOT NULL
    ),
    home_country AS (
        SELECT DISTINCT ON (m.geokret)
            m.geokret AS gk_id,
            m.country::char(2) AS home_country
        FROM geokrety.gk_moves m
        WHERE m.country IS NOT NULL
          AND m.move_type IN (0, 3, 5)
        ORDER BY m.geokret, m.moved_on_datetime ASC
    ),
    prev_holder AS (
        SELECT x.geokret AS gk_id, x.author AS previous_holder_id
        FROM (
            SELECT
                m.geokret,
                m.author,
                row_number() OVER (PARTITION BY m.geokret ORDER BY m.moved_on_datetime DESC) AS rn
            FROM geokrety.gk_moves m
            WHERE m.move_type = 1
              AND m.author IS NOT NULL
        ) x
        WHERE x.rn = 2
    ),
    last_drop AS (
        SELECT DISTINCT ON (m.geokret)
            m.geokret AS gk_id,
            m.moved_on_datetime AS last_drop_at,
            m.author AS last_drop_user_id
        FROM geokrety.gk_moves m
        WHERE m.move_type = 0
          AND m.author IS NOT NULL
        ORDER BY m.geokret, m.moved_on_datetime DESC
    ),
    last_cache_entry AS (
        SELECT DISTINCT ON (m.geokret)
            m.geokret AS gk_id,
            m.moved_on_datetime AS last_cache_entry_at
        FROM geokrety.gk_moves m
        WHERE m.move_type IN (0, 3)
        ORDER BY m.geokret, m.moved_on_datetime DESC
    )
    INSERT INTO geokrety_stats.gk_context_cache (
        gk_id,
        home_country,
        previous_holder_id,
        last_drop_at,
        last_drop_user_id,
        last_cache_entry_at,
        refreshed_at
    )
    SELECT
        gk.gk_id,
        hc.home_country,
        ph.previous_holder_id,
        ld.last_drop_at,
        ld.last_drop_user_id,
        lce.last_cache_entry_at,
        now()
    FROM gk_ids gk
    LEFT JOIN home_country hc ON hc.gk_id = gk.gk_id
    LEFT JOIN prev_holder ph ON ph.gk_id = gk.gk_id
    LEFT JOIN last_drop ld ON ld.gk_id = gk.gk_id
    LEFT JOIN last_cache_entry lce ON lce.gk_id = gk.gk_id
    ON CONFLICT (gk_id) DO UPDATE SET
        home_country = EXCLUDED.home_country,
        previous_holder_id = EXCLUDED.previous_holder_id,
        last_drop_at = EXCLUDED.last_drop_at,
        last_drop_user_id = EXCLUDED.last_drop_user_id,
        last_cache_entry_at = EXCLUDED.last_cache_entry_at,
        refreshed_at = now();

    GET DIAGNOSTICS _affected = ROW_COUNT;
    RETURN _affected;
END;
$$;
