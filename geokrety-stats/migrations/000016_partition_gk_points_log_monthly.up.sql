DO $$
DECLARE
	is_partitioned BOOLEAN;
BEGIN
	SELECT c.relkind = 'p'
	INTO is_partitioned
	FROM pg_class c
	JOIN pg_namespace n ON n.oid = c.relnamespace
	WHERE n.nspname = 'geokrety_stats' AND c.relname = 'gk_points_log';

	IF COALESCE(is_partitioned, FALSE) THEN
		RETURN;
	END IF;

	IF to_regclass('geokrety_stats.gk_points_log') IS NOT NULL
	   AND to_regclass('geokrety_stats.gk_points_log_legacy') IS NULL THEN
		ALTER TABLE geokrety_stats.gk_points_log RENAME TO gk_points_log_legacy;
	END IF;

	IF EXISTS (
		SELECT 1
		FROM pg_constraint con
		JOIN pg_class cls ON cls.oid = con.conrelid
		JOIN pg_namespace nsp ON nsp.oid = cls.relnamespace
		WHERE nsp.nspname = 'geokrety_stats'
		  AND cls.relname = 'gk_points_log_legacy'
		  AND con.conname = 'gk_points_log_pkey'
	) AND NOT EXISTS (
		SELECT 1
		FROM pg_constraint con
		JOIN pg_class cls ON cls.oid = con.conrelid
		JOIN pg_namespace nsp ON nsp.oid = cls.relnamespace
		WHERE nsp.nspname = 'geokrety_stats'
		  AND cls.relname = 'gk_points_log_legacy'
		  AND con.conname = 'gk_points_log_legacy_pkey'
	) THEN
		ALTER TABLE geokrety_stats.gk_points_log_legacy
			RENAME CONSTRAINT gk_points_log_pkey TO gk_points_log_legacy_pkey;
	END IF;

	IF to_regclass('geokrety_stats.gk_points_log_id_seq') IS NULL THEN
		CREATE SEQUENCE geokrety_stats.gk_points_log_id_seq;
	END IF;
END;
$$;

CREATE TABLE IF NOT EXISTS geokrety_stats.gk_points_log (
	id               BIGINT NOT NULL DEFAULT nextval('geokrety_stats.gk_points_log_id_seq'::regclass),
	gk_id            BIGINT NOT NULL,
	move_id          BIGINT,
	old_multiplier   DOUBLE PRECISION,
	new_multiplier   DOUBLE PRECISION,
	multiplier_delta DOUBLE PRECISION,
	reason           TEXT,
	module_source    VARCHAR(64),
	updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT gk_points_log_pkey PRIMARY KEY (id, updated_at)
) PARTITION BY RANGE (updated_at);

CREATE TABLE IF NOT EXISTS geokrety_stats.gk_points_log_default
	PARTITION OF geokrety_stats.gk_points_log DEFAULT;

CREATE INDEX IF NOT EXISTS idx_gk_points_log_gk_id ON geokrety_stats.gk_points_log (gk_id);
CREATE INDEX IF NOT EXISTS idx_gk_points_log_move_id ON geokrety_stats.gk_points_log (move_id);
CREATE INDEX IF NOT EXISTS idx_gk_points_log_updated_at ON geokrety_stats.gk_points_log (updated_at);

DO $$
BEGIN
	IF to_regclass('geokrety_stats.gk_points_log_legacy') IS NOT NULL THEN
		IF to_regclass('geokrety_stats.gk_points_log_id_seq') IS NOT NULL THEN
			ALTER SEQUENCE geokrety_stats.gk_points_log_id_seq OWNED BY NONE;
		END IF;

		INSERT INTO geokrety_stats.gk_points_log (
			id, gk_id, move_id, old_multiplier, new_multiplier, multiplier_delta, reason, module_source, updated_at
		)
		SELECT
			id, gk_id, move_id, old_multiplier, new_multiplier, multiplier_delta, reason, module_source, updated_at
		FROM geokrety_stats.gk_points_log_legacy
		ON CONFLICT DO NOTHING;

		DROP TABLE geokrety_stats.gk_points_log_legacy;

		IF to_regclass('geokrety_stats.gk_points_log_id_seq') IS NOT NULL THEN
			ALTER SEQUENCE geokrety_stats.gk_points_log_id_seq OWNED BY geokrety_stats.gk_points_log.id;
		END IF;
	END IF;
END;
$$;

COMMENT ON TABLE geokrety_stats.gk_points_log IS
	'Partitioned monthly log of GeoKret multiplier changes with full audit trail.';

CREATE OR REPLACE FUNCTION geokrety_stats.ensure_gk_points_log_month_partition(p_month DATE)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
	month_start DATE := date_trunc('month', p_month)::date;
	next_month DATE := (date_trunc('month', p_month) + interval '1 month')::date;
	partition_name TEXT := format('gk_points_log_%s', to_char(month_start, 'YYYYMM'));
BEGIN
	EXECUTE format(
		'CREATE TABLE IF NOT EXISTS geokrety_stats.%I PARTITION OF geokrety_stats.gk_points_log FOR VALUES FROM (%L) TO (%L)',
		partition_name,
		month_start,
		next_month
	);
END;
$$;

CREATE OR REPLACE FUNCTION geokrety_stats.rotate_gk_points_log_partitions(
	p_as_of TIMESTAMPTZ,
	p_retain_months INTEGER DEFAULT 6,
	p_future_months INTEGER DEFAULT 2
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
	base_month DATE := date_trunc('month', p_as_of)::date;
	month_cursor DATE;
	cutoff DATE := (base_month - make_interval(months => GREATEST(p_retain_months, 1)))::date;
	partition_name TEXT;
	record_row RECORD;
BEGIN
	FOR i IN -GREATEST(p_retain_months, 1)..GREATEST(p_future_months, 0) LOOP
		month_cursor := (base_month + make_interval(months => i))::date;
		PERFORM geokrety_stats.ensure_gk_points_log_month_partition(month_cursor);
	END LOOP;

	FOR record_row IN
		SELECT
			c.relname AS partition_name,
			to_date(substring(c.relname from 'gk_points_log_(\\d{6})$'), 'YYYYMM') AS part_month
		FROM pg_inherits i
		JOIN pg_class p ON p.oid = i.inhparent
		JOIN pg_class c ON c.oid = i.inhrelid
		JOIN pg_namespace n ON n.oid = p.relnamespace
		WHERE n.nspname = 'geokrety_stats'
		  AND p.relname = 'gk_points_log'
		  AND c.relname ~ '^gk_points_log_[0-9]{6}$'
	LOOP
		IF record_row.part_month < cutoff THEN
			partition_name := record_row.partition_name;
			EXECUTE format('DROP TABLE IF EXISTS geokrety_stats.%I', partition_name);
		END IF;
	END LOOP;
END;
$$;
