-- Replace waypoint partition helper with temp staging table approach
-- and remove old persistent uwmc_stage_* helper tables.

CREATE OR REPLACE FUNCTION geokrety_stats.create_user_waypoint_month_partition(_year_month char(7))
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
	part_name text;
	part_reg regclass;
	has_default_rows boolean := false;
BEGIN
	IF _year_month IS NULL OR _year_month::text !~ '^[0-9]{4}-[0-9]{2}$' THEN
		RAISE EXCEPTION 'Invalid year_month: % (expected YYYY-MM)', _year_month;
	END IF;

	part_name := format('user_waypoint_monthly_counts_p_%s', replace(_year_month::text, '-', '_'));
	part_reg := to_regclass(format('geokrety_stats.%I', part_name));
	IF part_reg IS NOT NULL THEN
		RETURN part_name;
	END IF;

	SELECT EXISTS (
		SELECT 1
		FROM geokrety_stats.user_waypoint_monthly_counts_p_default
		WHERE year_month = _year_month
	)
	INTO has_default_rows;

	IF has_default_rows THEN
		CREATE TEMP TABLE IF NOT EXISTS uwmc_temp_stage (
			user_id bigint NOT NULL,
			location_id varchar(64) NOT NULL,
			year_month char(7) NOT NULL,
			gk_id bigint NOT NULL,
			scored_at timestamptz NOT NULL,
			PRIMARY KEY (user_id, location_id, year_month, gk_id)
		) ON COMMIT DROP;

		TRUNCATE uwmc_temp_stage;

		INSERT INTO uwmc_temp_stage (user_id, location_id, year_month, gk_id, scored_at)
		SELECT user_id, location_id, year_month, gk_id, scored_at
		FROM geokrety_stats.user_waypoint_monthly_counts_p_default
		WHERE year_month = _year_month;

		DELETE FROM geokrety_stats.user_waypoint_monthly_counts_p_default
		WHERE year_month = _year_month;
	END IF;

	EXECUTE format(
		'CREATE TABLE IF NOT EXISTS geokrety_stats.%I PARTITION OF geokrety_stats.user_waypoint_monthly_counts FOR VALUES IN (%L)',
		part_name,
		_year_month::text
	);

	IF has_default_rows THEN
		EXECUTE format(
			'INSERT INTO geokrety_stats.%I (user_id, location_id, year_month, gk_id, scored_at)
			 SELECT user_id, location_id, year_month, gk_id, scored_at
			 FROM uwmc_temp_stage
			 ON CONFLICT DO NOTHING',
			part_name
		);
	END IF;

	RETURN part_name;
END;
$$;

DO $$
DECLARE
	r record;
BEGIN
	FOR r IN
		SELECT relname
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'geokrety_stats'
		  AND c.relkind = 'r'
		  AND c.relname LIKE 'uwmc_stage_%'
	LOOP
		EXECUTE format('DROP TABLE IF EXISTS geokrety_stats.%I', r.relname);
	END LOOP;
END;
$$;
