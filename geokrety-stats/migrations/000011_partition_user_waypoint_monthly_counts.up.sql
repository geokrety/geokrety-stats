-- Partition geokrety_stats.user_waypoint_monthly_counts by month key (year_month)
-- to enable cheap retention operations: detach/drop old months.

DO $$
DECLARE
	is_partitioned boolean := false;
BEGIN
	SELECT (c.relkind = 'p')
	INTO is_partitioned
	FROM pg_class c
	JOIN pg_namespace n ON n.oid = c.relnamespace
	WHERE n.nspname = 'geokrety_stats'
	  AND c.relname = 'user_waypoint_monthly_counts';

	IF is_partitioned THEN
		RETURN;
	END IF;

	IF to_regclass('geokrety_stats.user_waypoint_monthly_counts_legacy') IS NULL THEN
		ALTER TABLE geokrety_stats.user_waypoint_monthly_counts
			RENAME TO user_waypoint_monthly_counts_legacy;
	END IF;

	IF EXISTS (
		SELECT 1
		FROM pg_constraint con
		JOIN pg_class cls ON cls.oid = con.conrelid
		JOIN pg_namespace nsp ON nsp.oid = cls.relnamespace
		WHERE nsp.nspname = 'geokrety_stats'
		  AND cls.relname = 'user_waypoint_monthly_counts_legacy'
		  AND con.conname = 'user_waypoint_monthly_counts_pkey'
	) THEN
		ALTER TABLE geokrety_stats.user_waypoint_monthly_counts_legacy
			RENAME CONSTRAINT user_waypoint_monthly_counts_pkey TO user_waypoint_monthly_counts_legacy_pkey;
	END IF;
END;
$$;

CREATE TABLE IF NOT EXISTS geokrety_stats.user_waypoint_monthly_counts (
	user_id bigint NOT NULL,
	location_id varchar(64) NOT NULL,
	year_month char(7) NOT NULL,
	gk_id bigint NOT NULL,
	scored_at timestamptz NOT NULL,
	CONSTRAINT user_waypoint_monthly_counts_pkey
		PRIMARY KEY (user_id, location_id, year_month, gk_id)
) PARTITION BY LIST (year_month);

CREATE TABLE IF NOT EXISTS geokrety_stats.user_waypoint_monthly_counts_p_default
	PARTITION OF geokrety_stats.user_waypoint_monthly_counts
	DEFAULT;

INSERT INTO geokrety_stats.user_waypoint_monthly_counts (user_id, location_id, year_month, gk_id, scored_at)
SELECT user_id, location_id, year_month, gk_id, scored_at
FROM geokrety_stats.user_waypoint_monthly_counts_legacy
ON CONFLICT DO NOTHING;

DROP TABLE IF EXISTS geokrety_stats.user_waypoint_monthly_counts_legacy;

COMMENT ON TABLE geokrety_stats.user_waypoint_monthly_counts IS
	'Tracks how many distinct GKs a user has scored at each location per calendar month. Partitioned by year_month for retention operations.';

CREATE OR REPLACE FUNCTION geokrety_stats.create_user_waypoint_month_partition(_year_month char(7))
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
	part_name text;
BEGIN
	IF _year_month IS NULL OR _year_month::text !~ '^[0-9]{4}-[0-9]{2}$' THEN
		RAISE EXCEPTION 'Invalid year_month: % (expected YYYY-MM)', _year_month;
	END IF;

	part_name := format('user_waypoint_monthly_counts_p_%s', replace(_year_month::text, '-', '_'));

	EXECUTE format(
		'CREATE TABLE IF NOT EXISTS geokrety_stats.%I PARTITION OF geokrety_stats.user_waypoint_monthly_counts FOR VALUES IN (%L)',
		part_name,
		_year_month::text
	);

	EXECUTE format(
		'INSERT INTO geokrety_stats.%I (user_id, location_id, year_month, gk_id, scored_at)
		 SELECT user_id, location_id, year_month, gk_id, scored_at
		 FROM geokrety_stats.user_waypoint_monthly_counts_p_default
		 WHERE year_month = %L
		 ON CONFLICT DO NOTHING',
		part_name,
		_year_month::text
	);

	DELETE FROM geokrety_stats.user_waypoint_monthly_counts_p_default
	WHERE year_month = _year_month;

	RETURN part_name;
END;
$$;

CREATE OR REPLACE FUNCTION geokrety_stats.ensure_user_waypoint_month_partitions(
	_months_back integer DEFAULT 1,
	_months_ahead integer DEFAULT 2
)
RETURNS integer
LANGUAGE plpgsql
AS $$
DECLARE
	offset_month integer;
	month_key char(7);
	created_count integer := 0;
BEGIN
	FOR offset_month IN -_months_back.._months_ahead LOOP
		month_key := to_char(date_trunc('month', now()) + make_interval(months => offset_month), 'YYYY-MM')::char(7);
		PERFORM geokrety_stats.create_user_waypoint_month_partition(month_key);
		created_count := created_count + 1;
	END LOOP;

	RETURN created_count;
END;
$$;

CREATE OR REPLACE FUNCTION geokrety_stats.detach_or_drop_user_waypoint_month_partition(
	_year_month char(7),
	_drop boolean DEFAULT false
)
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
	part_name text;
	part_reg regclass;
BEGIN
	IF _year_month IS NULL OR _year_month::text !~ '^[0-9]{4}-[0-9]{2}$' THEN
		RAISE EXCEPTION 'Invalid year_month: % (expected YYYY-MM)', _year_month;
	END IF;

	part_name := format('user_waypoint_monthly_counts_p_%s', replace(_year_month::text, '-', '_'));
	part_reg := to_regclass(format('geokrety_stats.%I', part_name));

	IF part_reg IS NULL THEN
		RETURN format('partition %s does not exist', part_name);
	END IF;

	IF _drop THEN
		EXECUTE format('DROP TABLE geokrety_stats.%I', part_name);
		RETURN format('dropped partition %s', part_name);
	END IF;

	EXECUTE format('ALTER TABLE geokrety_stats.user_waypoint_monthly_counts DETACH PARTITION geokrety_stats.%I', part_name);
	RETURN format('detached partition %s', part_name);
END;
$$;

-- Prepare near-term partitions so current traffic avoids default partition.
SELECT geokrety_stats.ensure_user_waypoint_month_partitions(1, 2);
