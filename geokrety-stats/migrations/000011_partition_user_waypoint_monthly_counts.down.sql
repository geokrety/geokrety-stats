DROP FUNCTION IF EXISTS geokrety_stats.detach_or_drop_user_waypoint_month_partition(char(7), boolean);
DROP FUNCTION IF EXISTS geokrety_stats.ensure_user_waypoint_month_partitions(integer, integer);
DROP FUNCTION IF EXISTS geokrety_stats.create_user_waypoint_month_partition(char(7));

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

	IF NOT is_partitioned THEN
		RETURN;
	END IF;

	CREATE TABLE geokrety_stats.user_waypoint_monthly_counts_flat (
		user_id bigint NOT NULL,
		location_id varchar(64) NOT NULL,
		year_month char(7) NOT NULL,
		gk_id bigint NOT NULL,
		scored_at timestamptz NOT NULL,
		CONSTRAINT user_waypoint_monthly_counts_flat_pkey
			PRIMARY KEY (user_id, location_id, year_month, gk_id)
	);

	INSERT INTO geokrety_stats.user_waypoint_monthly_counts_flat (user_id, location_id, year_month, gk_id, scored_at)
	SELECT user_id, location_id, year_month, gk_id, scored_at
	FROM geokrety_stats.user_waypoint_monthly_counts;

	DROP TABLE geokrety_stats.user_waypoint_monthly_counts CASCADE;

	ALTER TABLE geokrety_stats.user_waypoint_monthly_counts_flat
		RENAME TO user_waypoint_monthly_counts;
END;
$$;

COMMENT ON TABLE geokrety_stats.user_waypoint_monthly_counts IS
	'Tracks how many distinct GKs a user has scored at each location per calendar month. Used for waypoint penalty.';
