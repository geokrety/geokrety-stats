-- Pre-create historical and future month partitions for replay efficiency.
-- Range: 2007-01 through (current month + 10 years).

-- Harden helper function so creating a new month partition also works when
-- matching rows are still present in the default partition.
CREATE OR REPLACE FUNCTION geokrety_stats.create_user_waypoint_month_partition(_year_month char(7))
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
	part_name text;
	stage_name text;
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
		stage_name := format('uwmc_stage_%s', replace(_year_month::text, '-', '_'));

		EXECUTE format(
			'CREATE TABLE IF NOT EXISTS geokrety_stats.%I (
				user_id bigint NOT NULL,
				location_id varchar(64) NOT NULL,
				year_month char(7) NOT NULL,
				gk_id bigint NOT NULL,
				scored_at timestamptz NOT NULL,
				CONSTRAINT %I PRIMARY KEY (user_id, location_id, year_month, gk_id)
			)',
			stage_name,
			stage_name || '_pkey'
		);

		EXECUTE format('TRUNCATE geokrety_stats.%I', stage_name);

		EXECUTE format(
			'INSERT INTO geokrety_stats.%I (user_id, location_id, year_month, gk_id, scored_at)
			 SELECT user_id, location_id, year_month, gk_id, scored_at
			 FROM geokrety_stats.user_waypoint_monthly_counts_p_default
			 WHERE year_month = %L',
			stage_name,
			_year_month::text
		);

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
			 FROM geokrety_stats.%I
			 ON CONFLICT DO NOTHING',
			part_name,
			stage_name
		);
	END IF;

	RETURN part_name;
END;
$$;

DO $$
DECLARE
	is_partitioned boolean := false;
	month_cursor date := DATE '2007-01-01';
	month_end date := date_trunc('month', now() + INTERVAL '10 years')::date;
BEGIN
	SELECT (c.relkind = 'p')
	INTO is_partitioned
	FROM pg_class c
	JOIN pg_namespace n ON n.oid = c.relnamespace
	WHERE n.nspname = 'geokrety_stats'
	  AND c.relname = 'user_waypoint_monthly_counts';

	IF NOT is_partitioned THEN
		RAISE NOTICE 'user_waypoint_monthly_counts is not partitioned; skipping precreate migration';
		RETURN;
	END IF;

	WHILE month_cursor <= month_end LOOP
		PERFORM geokrety_stats.create_user_waypoint_month_partition(
			to_char(month_cursor, 'YYYY-MM')::char(7)
		);
		month_cursor := (month_cursor + INTERVAL '1 month')::date;
	END LOOP;
END;
$$;
