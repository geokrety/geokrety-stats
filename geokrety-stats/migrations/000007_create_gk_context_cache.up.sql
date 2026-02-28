-- Trigger-maintained GK context cache to avoid repeated historical lookups
-- Useful for modules needing previous holder, home country, last drop/cache event.

CREATE TABLE IF NOT EXISTS geokrety_stats.gk_context_cache (
    gk_id                bigint PRIMARY KEY,
    home_country         char(2),
    previous_holder_id   bigint,
    last_drop_at         timestamptz,
    last_drop_user_id    bigint,
    last_cache_entry_at  timestamptz,
    refreshed_at         timestamptz NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION geokrety_stats.refresh_gk_context_cache(_gk_id bigint)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    _home_country       char(2);
    _previous_holder_id bigint;
    _last_drop_at       timestamptz;
    _last_drop_user_id  bigint;
    _last_cache_entry_at timestamptz;
BEGIN
    IF _gk_id IS NULL THEN
        RETURN;
    END IF;

    SELECT country::char(2)
    INTO _home_country
    FROM geokrety.gk_moves
    WHERE geokret = _gk_id
      AND country IS NOT NULL
      AND move_type IN (0, 3, 5)
    ORDER BY moved_on_datetime ASC
    LIMIT 1;

    SELECT author
    INTO _previous_holder_id
    FROM geokrety.gk_moves
    WHERE geokret = _gk_id
      AND move_type = 1
      AND author IS NOT NULL
    ORDER BY moved_on_datetime DESC
    LIMIT 1 OFFSET 1;

    SELECT moved_on_datetime, author
    INTO _last_drop_at, _last_drop_user_id
    FROM geokrety.gk_moves
    WHERE geokret = _gk_id
      AND move_type = 0
      AND author IS NOT NULL
    ORDER BY moved_on_datetime DESC
    LIMIT 1;

    SELECT moved_on_datetime
    INTO _last_cache_entry_at
    FROM geokrety.gk_moves
    WHERE geokret = _gk_id
      AND move_type IN (0, 3)
    ORDER BY moved_on_datetime DESC
    LIMIT 1;

    INSERT INTO geokrety_stats.gk_context_cache (
        gk_id,
        home_country,
        previous_holder_id,
        last_drop_at,
        last_drop_user_id,
        last_cache_entry_at,
        refreshed_at
    ) VALUES (
        _gk_id,
        _home_country,
        _previous_holder_id,
        _last_drop_at,
        _last_drop_user_id,
        _last_cache_entry_at,
        now()
    )
    ON CONFLICT (gk_id) DO UPDATE SET
        home_country = EXCLUDED.home_country,
        previous_holder_id = EXCLUDED.previous_holder_id,
        last_drop_at = EXCLUDED.last_drop_at,
        last_drop_user_id = EXCLUDED.last_drop_user_id,
        last_cache_entry_at = EXCLUDED.last_cache_entry_at,
        refreshed_at = now();
END;
$$;

CREATE OR REPLACE FUNCTION geokrety_stats.tg_refresh_gk_context_cache()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        PERFORM geokrety_stats.refresh_gk_context_cache(NEW.geokret);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        PERFORM geokrety_stats.refresh_gk_context_cache(NEW.geokret);
        IF OLD.geokret IS DISTINCT FROM NEW.geokret THEN
            PERFORM geokrety_stats.refresh_gk_context_cache(OLD.geokret);
        END IF;
        RETURN NEW;
    ELSE
        PERFORM geokrety_stats.refresh_gk_context_cache(OLD.geokret);
        RETURN OLD;
    END IF;
END;
$$;

DROP TRIGGER IF EXISTS trg_gk_moves_refresh_gk_context_cache ON geokrety.gk_moves;

CREATE TRIGGER trg_gk_moves_refresh_gk_context_cache
AFTER INSERT OR UPDATE OF geokret, move_type, author, country, moved_on_datetime OR DELETE
ON geokrety.gk_moves
FOR EACH ROW
EXECUTE FUNCTION geokrety_stats.tg_refresh_gk_context_cache();

CREATE INDEX IF NOT EXISTS idx_gk_context_cache_home_country
    ON geokrety_stats.gk_context_cache (home_country);

CREATE INDEX IF NOT EXISTS idx_gk_context_cache_previous_holder
    ON geokrety_stats.gk_context_cache (previous_holder_id);
