# Database migration runbook

This runbook describes the manual steps and commands used to apply and validate the snapshot/materialized-view/index migrations and to run scoped snapshot runs (1-month and 3-month). Use a safe test/staging copy first and always take a backup.

**Prerequisites**
- PostgreSQL client tools: `psql`, `pg_dump`, `pg_restore` (or equivalent)
- `pg_prove` (optional) for running pgTAP tests
- Project's phinx helper script at `.github/skills/phinx/scripts/phinx.sh`
- Database credentials with a user `geokrety` (adjust as necessary)

**Backup (required)**
- Full custom-format dump (fast restore):

```bash
mkdir -p ~/db-backups
BACKUP=~/db-backups/geokrety-$(date +%Y%m%d%H%M).dump
pg_dump -U geokrety -h localhost -Fc geokrety > "$BACKUP"
ls -lh "$BACKUP"
```

**Apply migrations**
- From repository root apply migrations in `website/db` using the project script:

```bash
cd website/db
/home/kumy/GIT/geokrety-stats/.github/skills/phinx/scripts/phinx.sh migrate
```

If you need to rollback the last applied migration:

```bash
/home/kumy/GIT/geokrety-stats/.github/skills/phinx/scripts/phinx.sh rollback --count=1
```

**Materialized views refresh (manual verification)**
After the migrations that create MVs, you can refresh them concurrently:

```sql
-- run in psql or here-doc
REFRESH MATERIALIZED VIEW CONCURRENTLY stats.mv_country_month_rollup;
REFRESH MATERIALIZED VIEW CONCURRENTLY stats.mv_top_caches_global;
REFRESH MATERIALIZED VIEW CONCURRENTLY stats.mv_global_kpi;
```

Example as a shell heredoc:

```bash
psql -U geokrety -d geokrety <<'EOF'
REFRESH MATERIALIZED VIEW CONCURRENTLY stats.mv_country_month_rollup;
REFRESH MATERIALIZED VIEW CONCURRENTLY stats.mv_top_caches_global;
REFRESH MATERIALIZED VIEW CONCURRENTLY stats.mv_global_kpi;
EOF
```

**Run snapshots (scoped by period)**
The main runner is `stats.fn_run_all_snapshots(p_phases, p_period, p_batch_size)`.
Pass `NULL::text[]` to run the default phase list. Use a tstzrange for `p_period` with the form `tstzrange('YYYY-MM-DD HH24:MI:SS+TZ','YYYY-MM-DD HH24:MI:SS+TZ','[)')` (left-inclusive, right-exclusive).

For full historical backfills, do not call the full runner with `NULL` period. Run the full pre-phases once, replay the sliceable phases month by month from `2007-10-01` up to tomorrow `00:00 UTC`, then run the full post-phases once.

The four new full rebuild phases for the backfilled tables are:
- `fn_snapshot_daily_entity_counts`
- `fn_snapshot_gk_country_history`
- `fn_snapshot_first_finder_events`
- `fn_snapshot_gk_milestone_events`

For these four phases, the automation script uses an explicit transaction with `SET LOCAL session_replication_role = replica` to bypass target-table constraint checks during the bulk rebuild. `SET LOCAL` only works inside a transaction, so the manual SQL recipe must do the same.

**Copy-paste monthly backfill block**

```bash
set -euo pipefail

export PGDATABASE="${PGDATABASE:-geokrety}"
export BATCH_SIZE="${BATCH_SIZE:-50000}"
export FULL_START_DATE="${FULL_START_DATE:-2007-10-01}"
export FULL_END_DATE="${FULL_END_DATE:-$(date -u -d tomorrow +%F)}"

run_full_phase_replica() {
	local phase="$1"
	psql -v ON_ERROR_STOP=1 -P pager=off <<EOF
BEGIN;
SET LOCAL session_replication_role = replica;
SELECT stats.fn_run_snapshot_phase('${phase}', NULL, ${BATCH_SIZE});
COMMIT;
EOF
}

psql -v ON_ERROR_STOP=1 -P pager=off -c \
	"SELECT stats.fn_run_snapshot_phase('fn_snapshot_entity_counters', NULL, ${BATCH_SIZE});"
run_full_phase_replica "fn_snapshot_daily_entity_counts"

cursor_date="$FULL_START_DATE"
while [[ "$cursor_date" < "$FULL_END_DATE" ]]; do
	slice_start="$(date -u -d "$cursor_date" '+%Y-%m-%d 00:00:00+00')"
	next_month_date="$(date -u -d "$cursor_date +1 month" +%F)"
	slice_end="$(date -u -d "$next_month_date" '+%Y-%m-%d 00:00:00+00')"

	if [[ "$next_month_date" > "$FULL_END_DATE" ]]; then
		slice_end="$(date -u -d "$FULL_END_DATE" '+%Y-%m-%d 00:00:00+00')"
	fi

	echo "== Slice ${slice_start} -> ${slice_end} =="

	psql -v ON_ERROR_STOP=1 -P pager=off -c \
		"SELECT stats.fn_run_snapshot_phase('fn_backfill_heavy_previous_move_id_all', tstzrange('${slice_start}','${slice_end}','[)'), ${BATCH_SIZE});"
	psql -v ON_ERROR_STOP=1 -P pager=off -c \
		"SELECT stats.fn_run_snapshot_phase('fn_seed_daily_activity', tstzrange('${slice_start}','${slice_end}','[)'), ${BATCH_SIZE});"
	psql -v ON_ERROR_STOP=1 -P pager=off -c \
		"SELECT stats.fn_run_snapshot_phase('fn_snapshot_daily_country_stats', tstzrange('${slice_start}','${slice_end}','[)'), ${BATCH_SIZE});"
	psql -v ON_ERROR_STOP=1 -P pager=off -c \
		"SELECT stats.fn_run_snapshot_phase('fn_snapshot_user_country_stats', tstzrange('${slice_start}','${slice_end}','[)'), ${BATCH_SIZE});"
	psql -v ON_ERROR_STOP=1 -P pager=off -c \
		"SELECT stats.fn_run_snapshot_phase('fn_snapshot_gk_country_stats', tstzrange('${slice_start}','${slice_end}','[)'), ${BATCH_SIZE});"
	psql -v ON_ERROR_STOP=1 -P pager=off -c \
		"SELECT stats.fn_run_snapshot_phase('fn_snapshot_relationship_tables', tstzrange('${slice_start}','${slice_end}','[)'), ${BATCH_SIZE});"
	psql -v ON_ERROR_STOP=1 -P pager=off -c \
		"SELECT stats.fn_run_snapshot_phase('fn_snapshot_hourly_activity', tstzrange('${slice_start}','${slice_end}','[)'), ${BATCH_SIZE});"
	psql -v ON_ERROR_STOP=1 -P pager=off -c \
		"SELECT stats.fn_run_snapshot_phase('fn_snapshot_country_pair_flows', tstzrange('${slice_start}','${slice_end}','[)'), ${BATCH_SIZE});"

	cursor_date="$next_month_date"
done

run_full_phase_replica "fn_snapshot_gk_country_history"
run_full_phase_replica "fn_snapshot_first_finder_events"
run_full_phase_replica "fn_snapshot_gk_milestone_events"
```

Notes:
- `fn_snapshot_entity_counters` is fast enough to run once without slicing and does not use replica mode.
- `fn_snapshot_daily_entity_counts` runs once before the monthly loop because it rebuilds the entire trend table from source history.
- `fn_backfill_heavy_previous_move_id_all` and other sliced phases run once per month (monthly basis) for efficient backfill.
- `fn_snapshot_gk_country_history`, `fn_snapshot_first_finder_events`, and `fn_snapshot_gk_milestone_events` run once after the monthly loop because they rebuild whole-history tables.
- The last slice is partial for the current month and stops at tomorrow `00:00 UTC`.
- Override `FULL_START_DATE`, `FULL_END_DATE`, or `BATCH_SIZE` in the shell if you need a narrower replay window.

**Automation script**

Use the helper script next to this runbook for the same phase order with per-phase timings, progress tracking, ETA with date/time, colors, and emoji:

```bash
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py --start 2007-10 --end 2008-01
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py --start 2026-01 --end 2026-02 --batch-size 50000
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py --start 2007-10 --end 2008-01 --dry-run
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py --no-replica-role
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py --clear-resume-markers --dry-run
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py --no-resume
```

The script provides:
- **Live progress tracking** with table format showing phase, slice period, elapsed time, and ETA datetime
- **Color-coded output** for easy visual scanning
- **Emoji indicators** for status (⏳ pending, ⚙️ running, ✅ done, 📊 stats)
- **ETA with datetime** (not just duration), so you know when the run will complete
- **Per-phase timing** from both wall-clock and server-side measurements
- **Throughput statistics** (steps per hour) at the end
- **--dry-run** mode to preview all planned steps before execution
- **Replica-role fast path by default** for `daily_entity_counts`, `gk_country_history`, `first_finder_events`, and `gk_milestone_events`; pass `--no-replica-role` to disable it
- **Exact-run resume by default** using runner-owned `stats.job_log` markers (`job_name = 'run_snapshot_backfill_step'`)
- **--clear-resume-markers** to delete only the markers for the exact resolved run key before planning
- **--no-resume** to force a full replay even when exact-run markers already exist

**Resume semantics**

- The runner computes a `run_key` from the resolved request window, current source bounds, batch size, parallel mode, replica-role mode, and the `--skip-entity-counters` flag.
- Every completed phase/slice step writes one canonical marker row into `stats.job_log` with `job_name = 'run_snapshot_backfill_step'`.
- Re-running the script with the exact same parameters and unchanged source bounds filters those completed steps out of the plan, so the process continues instead of replaying already-finished work.
- If the source bounds change, the `run_key` changes too, so stale markers are ignored automatically.
- The runner takes a PostgreSQL advisory lock for the resolved request/source window so two identical runs cannot overlap.
- For an operator reset, use `--clear-resume-markers`; for a deliberate full replay, use `--no-resume`.

Example continuation flow:

```bash
# First run for a bounded window.
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py \
	--start 2026-03 --end 2026-04

# Same command again: completed steps are skipped and the runner exits quickly.
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py \
	--start 2026-03 --end 2026-04

# Force a fresh replay for the same window.
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py \
	--start 2026-03 --end 2026-04 --no-resume

# Clear only the exact-run markers, then preview the rebuilt plan.
python /home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py \
	--start 2026-03 --end 2026-04 --clear-resume-markers --dry-run
```

Example output:

```
ℹ️ Source bounds: 2007-10-01 00:00:00+00 → 2026-03-15 00:00:00+00
ℹ️ Requested run: 2007-10-01 00:00:00+00 → 2007-11-01 00:00:00+00
📊 Slices: 1 | Steps: 7 | Batch size: 50000

  #  │ Phase (50 chars)                          │ Slice (25 chars)         │ Elapsed  │ ETA (UTC)
  ─────────────────────────────────────────────────────────────────────────────────────────────────────

  ⚙️ 1 │ fn_snapshot_entity_counters                 │ full                     │ 00:02:15 │ 2026-03-15 12:45:30
	✅ 1 │ fn_snapshot_entity_counters                 │ full                     │ 00:02:15 │ 2026-03-15 12:30:00
	⚙️ 2 │ fn_snapshot_daily_entity_counts             │ full                     │ 00:00:04 │ 2026-03-15 12:30:04
	✅ 2 │ fn_snapshot_daily_entity_counts             │ full                     │ 00:00:04 │ 2026-03-15 12:30:04
	⚙️ 3 │ fn_backfill_heavy_previous_move_id_all      │ 2007-10-01..2007-11-01   │ 00:00:01 │ 2026-03-15 12:30:05
	⚙️ 4 │ PARALLEL (7 phases)                         │ 2007-10-01..2007-11-01   │ 00:00:08 │ 2026-03-15 12:30:13
	⚙️ 5 │ fn_snapshot_gk_country_history              │ full                     │ 00:00:24 │ 2026-03-15 12:30:37
	⚙️ 6 │ fn_snapshot_first_finder_events             │ full                     │ 00:00:01 │ 2026-03-15 12:30:38
	⚙️ 7 │ fn_snapshot_gk_milestone_events             │ full                     │ 00:00:17 │ 2026-03-15 12:30:55
  ...

  ═====════════════════════════════════════════════════════════════════════════════════════════════════
	📊 Completed 13 phases in 00:00:58 (799.1 steps/hour)
```

**Sliced heavy phases**

The `fn_backfill_heavy_previous_move_id_all` function runs on a per-month basis (sliced). This allows:
- Progressive backfill of historical data
- Better resource usage with bounded batch sizes
- Resumable runs if one month fails (re-run just that month)

Examples below use UTC timestamps — adjust timezone offsets if needed.

- Run for a 1-month period (example: 2026-02-01 => 2026-03-01):

```bash
START='2026-02-01 00:00:00+00'
END='2026-03-01 00:00:00+00'
psql -U geokrety -d geokrety -c "SELECT stats.fn_run_all_snapshots(NULL::text[], tstzrange('$START','$END','[)'), 50000);"
```

- Run for a 3-month period (example: 2026-01-01 => 2026-04-01):

```bash
START='2026-01-01 00:00:00+00'
END='2026-04-01 00:00:00+00'
psql -U geokrety -d geokrety -c "SELECT stats.fn_run_all_snapshots(NULL::text[], tstzrange('$START','$END','[)'), 50000);"
```

Notes:
- If you want to run a single phase only, pass the phase names as the first argument, e.g. `ARRAY['fn_snapshot_hourly_activity']::text[]`.
- If the snapshot run is long or times out, lower `p_batch_size` (third argument) and retry.

**Run focused pgTAP tests**
- Preferred: `pg_prove` (if available):

```bash
pg_prove -d geokrety website/db/tests/test-250-materialized-views.sql
```

- Fallback: run SQL test file with `psql` (some tests expect pgTAP harness):

```bash
psql -U geokrety -d geokrety -f website/db/tests/test-250-materialized-views.sql
```

**Full test suite**
- Use your project test runner or `pg_prove` over the `website/db/tests` directory if set up. Example:

```bash
pg_prove -d geokrety website/db/tests/*.sql
```

or use the project's test script if present.

**Verification queries**
- Check recent snapshot job results:

```sql
SELECT * FROM stats.job_log ORDER BY started_at DESC LIMIT 50;
```

- Inspect runner-owned resume markers for the latest exact run:

```sql
SELECT
	metadata->>'run_key' AS run_key,
	metadata->>'phase' AS phase,
	COALESCE(metadata->>'slice_start', 'full') AS slice_start,
	metadata->>'parallel_mode' AS parallel_mode,
	completed_at
FROM stats.job_log
WHERE job_name = 'run_snapshot_backfill_step'
ORDER BY completed_at DESC, id DESC
LIMIT 50;
```

- Confirm materialized views exist:

```sql
SELECT matviewname FROM pg_matviews WHERE schemaname = 'stats';
```

- Inspect gk_moves indexes (verify the expected indexes remain):

```sql
SELECT schemaname, tablename, indexname
FROM pg_indexes
WHERE schemaname = 'geokrety' AND tablename = 'gk_moves'
ORDER BY indexname;
```

**Rollback plan (if a run fails or you need to undo a migration)**
- Rollback last migration:

```bash
/home/kumy/GIT/geokrety-stats/.github/skills/phinx/scripts/phinx.sh rollback --count=1
```

- If you need to restore DB from backup (fast):

```bash
pg_restore -U geokrety -d geokrety --clean --no-owner /path/to/geokrety-YYYYMMDDHHMM.dump
```

**Troubleshooting & tips**
- If pgTAP reports "duplicate plan()" errors, inspect the SQL test file for multiple `plan()` declarations and remove duplicates.
- Pre-commit hooks may auto-fix EOF/trailing whitespace; if a commit fails, `git add` the modified files again and re-commit.
- Long-running `REFRESH MATERIALIZED VIEW` should be run with `CONCURRENTLY` when possible to avoid exclusive locks.
- Use `screen`/`tmux` or background jobs for long snapshot runs and capture output to a log file:

```bash
psql -U geokrety -d geokrety -c "SELECT stats.fn_run_all_snapshots(NULL::text[], tstzrange('$START','$END','[)'), 50000);" | tee snapshot-run.log
```

**Change log & notes**
- This runbook covers the manual steps used during the March 2026 recovery: restoring orchestration functions, materialized views, runtime snapshot indexes, index cleanup, and move-history view. Adjust dates and phase lists to your environment.
