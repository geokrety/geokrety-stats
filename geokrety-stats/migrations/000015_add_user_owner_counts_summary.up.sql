CREATE TABLE IF NOT EXISTS geokrety_stats.user_owner_counts_summary (
	user_id bigint NOT NULL,
	owner_id bigint NOT NULL,
	gk_count integer NOT NULL DEFAULT 0,
	CONSTRAINT user_owner_counts_summary_pkey PRIMARY KEY (user_id, owner_id),
	CONSTRAINT user_owner_counts_summary_nonneg CHECK (gk_count >= 0)
);

INSERT INTO geokrety_stats.user_owner_counts_summary (user_id, owner_id, gk_count)
SELECT user_id, owner_id, COUNT(*)::integer
FROM geokrety_stats.user_owner_gk_counts
GROUP BY user_id, owner_id
ON CONFLICT (user_id, owner_id) DO UPDATE
SET gk_count = EXCLUDED.gk_count;

CREATE OR REPLACE FUNCTION geokrety_stats.tg_sync_user_owner_counts_summary()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
	IF TG_OP = 'INSERT' THEN
		INSERT INTO geokrety_stats.user_owner_counts_summary (user_id, owner_id, gk_count)
		VALUES (NEW.user_id, NEW.owner_id, 1)
		ON CONFLICT (user_id, owner_id) DO UPDATE
		SET gk_count = geokrety_stats.user_owner_counts_summary.gk_count + 1;
		RETURN NEW;
	ELSIF TG_OP = 'DELETE' THEN
		UPDATE geokrety_stats.user_owner_counts_summary
		SET gk_count = gk_count - 1
		WHERE user_id = OLD.user_id
		  AND owner_id = OLD.owner_id;

		DELETE FROM geokrety_stats.user_owner_counts_summary
		WHERE user_id = OLD.user_id
		  AND owner_id = OLD.owner_id
		  AND gk_count <= 0;
		RETURN OLD;
	END IF;

	RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS trg_user_owner_counts_summary_sync ON geokrety_stats.user_owner_gk_counts;

CREATE TRIGGER trg_user_owner_counts_summary_sync
AFTER INSERT OR DELETE ON geokrety_stats.user_owner_gk_counts
FOR EACH ROW
EXECUTE FUNCTION geokrety_stats.tg_sync_user_owner_counts_summary();
