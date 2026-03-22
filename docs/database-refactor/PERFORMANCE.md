# Snapshot Backfill Performance Analysis

## Goal

Reduce full 2007-10 → 2026-03 backfill time from ~5h 50m to **1–2 hours**.

---

## Baseline (Run: 2026-03-15)

Full run: `time python run_snapshot_backfill.py` → **5h 49m 38s** (20,978 seconds)

### Aggregated timing from `stats.job_log` (all 269 monthly runs)

| Function | Runs | Avg (s) | Max (s) | Total (s) | % of Total |
|---|---|---|---|---|---|
| `fn_snapshot_relationship_tables` | 269 | 57.77 | 209.77 | **15,541** | **74.1%** |
| → `fn_snapshot_cache_visits` | 269 | 30.97 | 113.41 | 8,332 | 39.7% |
| → `fn_snapshot_waypoints` | 269 | 24.64 | 91.32 | 6,627 | 31.6% |
| → `fn_snapshot_relations` | 269 | 2.16 | 12.64 | 582 | 2.8% |
| `fn_snapshot_daily_country_stats` | 268 | 7.22 | 9.86 | 1,934 | 9.2% |
| `fn_snapshot_user_country_stats` | 270 | 5.27 | 18.25 | 1,423 | 6.8% |
| `fn_snapshot_country_pair_flows` | 267 | 2.25 | 3.56 | 600 | 2.9% |
| `fn_snapshot_gk_country_stats` | 270 | 2.17 | 17.44 | 587 | 2.8% |
| `fn_snapshot_hourly_activity` | 267 | 1.72 | 2.78 | 460 | 2.2% |
| `fn_snapshot_entity_counters` | 24 | 3.07 | 3.33 | 74 | 0.4% |
| `fn_seed_daily_activity` | 268 | 0.15 | 0.82 | 40 | 0.2% |
| **Total** | | | | **~20,978** | **100%** |

### Per-month detail: May 2016 (dense month, used as benchmark)

| Function | May 2016 Timing |
|---|---|
| `fn_snapshot_waypoints` | 73.17s |
| `fn_snapshot_cache_visits` | 91.37s |
| `fn_snapshot_relations` | 3.27s |
| `fn_snapshot_relationship_tables` | 167.80s |
| `fn_snapshot_daily_country_stats` | 7.27s |
| `fn_snapshot_user_country_stats` | 9.62s |
| `fn_snapshot_gk_country_stats` | 1.55s |
| `fn_snapshot_hourly_activity` | 1.55s |
| `fn_snapshot_country_pair_flows` | 2.10s |
| `fn_seed_daily_activity` | 0.15s |
| **Total for May 2016** | **~192s (3.2 min)** |

---

## Root Cause Analysis

### #1 — RBAR (Row-By-Agonizing-Row) in `fn_snapshot_waypoints` and `fn_snapshot_cache_visits`

**71% of total runtime** comes from these two functions.

Both functions use a `FOR loop` iterating over each touched waypoint **one at a time**:

```sql
-- fn_snapshot_waypoints (simplified)
FOR v_waypoint_code IN SELECT tw.waypoint_code FROM tmp ORDER BY waypoint_code LOOP
  -- Window function query per waypoint → INSERT into stats.waypoints
  WITH ranked_waypoint AS (SELECT ... FROM gk_moves WHERE UPPER(BTRIM(waypoint)) = v_waypoint_code)
  INSERT INTO stats.waypoints ... ON CONFLICT ... DO UPDATE;
END LOOP;

-- fn_snapshot_cache_visits (simplified)
FOR v_waypoint IN SELECT tw.id, tw.waypoint_code FROM tmp ORDER BY waypoint_code LOOP
  INSERT INTO stats.gk_cache_visits ... FROM gk_moves WHERE UPPER(BTRIM(waypoint)) = v_waypoint.code;
  INSERT INTO stats.user_cache_visits ... FROM gk_moves WHERE UPPER(BTRIM(waypoint)) = v_waypoint.code;
END LOOP;
```

For May 2016: **7,325 waypoints** are touched → 7,325 iterations of each loop:

- `fn_snapshot_cache_visits`: 7,325 × 2 INSERTs = 14,650 individual queries → 91s → **6.2ms/query**

- `fn_snapshot_waypoints`: 7,325 iterations → 73s → **10ms/iteration**

**Per-iteration overhead breakdown:**

- PL/pgSQL executor overhead: ~0.5–1ms per iteration
- Index scan (using `idx_gk_moves_waypoint_code_hist`): ~0.3ms per iteration
- Individual INSERT into indexed table: ~5ms per iteration (buffer flush, index maintenance)

**Note**: The functional index `idx_gk_moves_waypoint_code_hist` on `UPPER(BTRIM(waypoint))` already exists, so index scans are already efficient. The bottleneck is **per-iteration overhead** + **individual INSERT overhead**.

### #2 — Serial execution of independent monthly phases

The 8 monthly phases run **sequentially** but only `fn_backfill_heavy_previous_move_id_all` is truly dependent (writes `previous_move_id`). All other 7 phases write to distinct tables and can run **in parallel**.

For May 2016 with estimated optimized times:

- Serial: ~30s
- Parallel (critical path = fn_snapshot_relationship_tables): ~11.5s

### #3 — `fn_normalize_country_code` is VOLATILE (prevents parallelism)

`geokrety.fn_normalize_country_code` is declared VOLATILE despite being a pure computation. This prevents:

- PostgreSQL from executing aggregates in parallel
- Creation of functional indexes on the country code

---

## Optimization Plan

### OPT-1: Eliminate RBAR in `fn_snapshot_waypoints`

**Strategy**: Replace the `FOR loop` with a single set-based SQL using **period-only moves**.

**Key insight**: The original function queries ALL history for each touched waypoint to find the "best" representative (lat/lon/country). But because:
1. The `ON CONFLICT` clause already uses `COALESCE` to preserve existing good data
2. The backfill processes months in chronological order
3. Earlier periods have already updated `stats.waypoints` with historical data

→ **We only need to process the current period's moves**. The COALESCE in `ON CONFLICT` will preserve any better data from previous runs.

**New approach**:
```sql
WITH ranked AS (
  SELECT tw.waypoint_code, lat, lon, country, moved_on_datetime AS first_seen_at,
    ROW_NUMBER() OVER (PARTITION BY tw.waypoint_code ORDER BY ...) AS rn
  FROM geokrety.gk_moves m
  JOIN tmp_snapshot_touched_waypoints tw ON tw.waypoint_code = UPPER(BTRIM(m.waypoint))
  WHERE m.moved_on_datetime >= v_period_start AND m.moved_on_datetime < v_period_end
    AND m.waypoint IS NOT NULL AND BTRIM(m.waypoint) <> '' AND m.move_type <> 2
)
INSERT INTO stats.waypoints (...) SELECT ... FROM ranked WHERE rn = 1
ON CONFLICT (waypoint_code) DO UPDATE SET ...;
```

**Expected speedup**: From 73s to **< 2s** (period has ~10K–30K rows vs full 6.9M; hash joins trivial)

### OPT-2: Eliminate RBAR in `fn_snapshot_cache_visits`

**Strategy**: Replace the `FOR loop` with two `CROSS JOIN LATERAL` queries.

`LATERAL` forces a nested-loop join which reuses the existing functional index per waypoint, but eliminates:
1. PL/pgSQL per-iteration overhead
2. Individual INSERT overhead (single bulk INSERT instead)

```sql
-- Replace loop with LATERAL-based bulk INSERT:
INSERT INTO stats.gk_cache_visits (gk_id, waypoint_id, visit_count, first_visited_at, last_visited_at)
SELECT ld.geokret, tw.id, ld.visit_count, ld.first_visited_at, ld.last_visited_at
FROM tmp_snapshot_touched_waypoints tw
CROSS JOIN LATERAL (
  SELECT m.geokret, COUNT(*)::BIGINT, MIN(m.moved_on_datetime), MAX(m.moved_on_datetime)
  FROM geokrety.gk_moves m
  WHERE m.waypoint IS NOT NULL AND BTRIM(m.waypoint) <> '' AND m.move_type <> 2
    AND UPPER(BTRIM(m.waypoint)) = tw.waypoint_code
  GROUP BY m.geokret
) ld;
```

**Expected speedup**: From 91s to **~10s** (LATERAL uses same functional index, eliminates 14,650 individual inserts)

### OPT-3: Mark `fn_normalize_country_code` as IMMUTABLE

Change from `VOLATILE` to `IMMUTABLE STRICT PARALLEL SAFE`. This function is a pure computation (BTRIM + UPPER + length check) with no side effects.

**Enables**:

- Parallel aggregate queries in `fn_snapshot_daily_country_stats`, `fn_snapshot_user_country_stats`, `fn_snapshot_gk_country_stats`
- Functional indexes on `fn_normalize_country_code(country)`

**Expected speedup**: ~30–40% reduction for the three country stats functions (parallelism enabled)

### OPT-4: Parallel execution of monthly phases in Python script

Run all 7 independent monthly phases concurrently (they write to disjoint tables).

**Dependency constraint**: `fn_backfill_heavy_previous_move_id_all` must complete first (writes `previous_move_id`), then all others can run in parallel.

**Expected speedup**: Per-month time drops from ~30s (optimized-serial) to ~11.5s (parallel, critical path = fn_snapshot_relationship_tables).

---

## Benchmark Results

### Baseline (2026-03-15, before any optimization)

| Period | Total Time | May 2016 Month |
|---|---|---|
| 2007-10 → 2026-03 (full run) | **5h 49m 38s** | 192s |

---

### After OPT-1 + OPT-2 — Migration `20260316110000_optimize_snapshot_rbar_functions.php`

#### Per-function comparison: May 2016

| Function | Before | After | Speedup |
|---|---|---|---|
| `fn_snapshot_waypoints` | 73.17s | **0.18s** | **406×** |
| `fn_snapshot_cache_visits` | 91.37s | **3.00s** | **30×** |
| `fn_snapshot_relationship_tables` | 167.80s | **6.73s** | **25×** |
| `fn_snapshot_daily_country_stats` | 7.27s | 7.36s | ~1× |
| `fn_snapshot_user_country_stats` | 9.62s | 10.17s | ~1× |
| `fn_snapshot_gk_country_stats` | 1.55s | 1.69s | ~1× |
| `fn_snapshot_hourly_activity` | 1.55s | 1.67s | ~1× |
| `fn_snapshot_country_pair_flows` | 2.10s | 3.94s | ~1× |
| `fn_seed_daily_activity` | 0.15s | 0.27s | ~1× |
| **Total May 2016** | **192s** | **37s** | **5.2×** |

#### Average per-function times (59 month runs across multiple years)

| Function | Avg (s) | Min (s) | Max (s) |
|---|---|---|---|
| `fn_snapshot_daily_country_stats` | 7.68 | 7.36 | 8.51 |
| `fn_snapshot_user_country_stats` | 7.07 | 0.21 | 14.86 |
| `fn_snapshot_relationship_tables` | 5.67 | 1.90 | 9.71 |
| `fn_snapshot_country_pair_flows` | 3.91 | 3.71 | 4.26 |
| `fn_snapshot_entity_counters` | 3.24 | 3.14 | 3.47 |
| `fn_snapshot_relations` | 3.07 | 1.05 | 4.68 |
| `fn_snapshot_gk_country_stats` | 3.01 | 0.07 | 11.37 |
| `fn_snapshot_cache_visits` | 2.39 | 0.76 | 5.02 |
| `fn_snapshot_hourly_activity` | 1.93 | 1.49 | 2.73 |
| `fn_seed_daily_activity` | 0.27 | 0.10 | 0.77 |
| `fn_snapshot_waypoints` | 0.21 | 0.01 | 0.75 |
| **Total serial per-month** | **~38.5s** | — | — |

#### Full run timing by year (serial, after OPT-1+2)

| Period | Total | Per-month avg |
|---|---|---|
| 2009 (sparse) | 220s | 18.3s/month |
| 2015 (medium) | 347s | 28.9s/month |
| 2016 Q2 (3 months, dense) | 101s | 33.7s/month |
| 2022 (dense recent) | 545s | 45.4s/month |
| 2025 (most recent) | 490s | 40.8s/month |

Projected full run serial (2007–2026): **~7,000–8,500s ≈ 2–2.5h**

---

### After OPT-1 + OPT-2 + OPT-4 — Parallel Python phases

#### Full run timing by year (parallel)

| Period | Serial (after OPT-1+2) | Parallel | Per-month parallel | Speedup |
|---|---|---|---|---|
| May 2016 (1 month) | 37s | **17s** | 17s | **2.2×** |
| 2016 Q2 (3 months) | 101s | **44s** | 14.7s | **2.3×** |
| 2022 (12 months) | 545s | **205s** | 17.1s | **2.7×** |
| 2025 (12 months) | 490s | **179s** | 14.9s | **2.7×** |

Projected full run parallel (2007–2026): **~3,000–3,500s ≈ 50–58 min** ✅

#### Overall speedup summary (vs original baseline)

| Optimization | May 2016 | Full run estimate | vs Baseline |
|---|---|---|---|
| Baseline (original) | 192s | 5h 49m 38s (20,978s) | 1× |
| OPT-1+2: SQL RBAR fix | 37s | ~2–2.5h | ~3× |
| OPT-1+2+4: SQL + parallel | **17s** | **~50–58 min** | **~6.5×** |

**Goal achieved**: Full run projected at under 1 hour (target was 1–2 hours).

---

## Optimizations Applied

| Migration | Status | Description |
|---|---|---|
| `20260316110000_optimize_snapshot_rbar_functions.php` | ✅ Applied | OPT-1: period-only set-based CTE for `fn_snapshot_waypoints`; OPT-2: CROSS JOIN LATERAL for `fn_snapshot_cache_visits` |
| `run_snapshot_backfill.py --parallel` (default) | ✅ Implemented | OPT-4: all 7 snapshot phases run concurrently per month via `ThreadPoolExecutor`; use `--no-parallel` to revert |

OPT-3 (`fn_normalize_country_code` IMMUTABLE): **Not needed** — DBA review confirmed the function is already `IMMUTABLE STRICT`. The remaining 3 country-stats functions are already set-based and PostgreSQL-level parallel-safe.

---

## Final full backfill — Verified run (2026-03-16)

The final full backfill was executed and completed successfully. Key facts from the run and verification:

- Completed steps: **1777**

- Wall time (real): **00:46:42.538** (46m42s)
- Throughput: **2282.8 steps/hour** (measured)

Recent `stats.job_log` entries (last 2 hours) grouped by `job_name` and `status`:

| job_name | status | count |
|---|---:|---:|
| `fn_backfill_previous_move_id` | ok | 250 |
| `fn_seed_daily_activity` | ok | 249 |
| `fn_snapshot_cache_visits` | ok | 249 |
| `fn_snapshot_country_pair_flows` | completed | 249 |
| `fn_snapshot_daily_country_stats` | ok | 249 |
| `fn_snapshot_entity_counters` | ok | 7 |
| `fn_snapshot_gk_country_stats` | ok | 249 |
| `fn_snapshot_hourly_activity` | completed | 249 |
| `fn_snapshot_relations` | ok | 250 |
| `fn_snapshot_relationship_tables` | ok | 249 |
| `fn_snapshot_user_country_stats` | ok | 249 |
| `fn_snapshot_waypoints` | ok | 249 |

No `error` status entries were observed in the recent `stats.job_log` window — the run completed without logged failures.

Log file (runtime events): `docs/database-refactor/run_snapshot_backfill.log` — contains per-phase start/done lines with wall time and job_log timing for each phase.

### Resume and continuation timing (2026-03-16)

The runner now persists one completion marker per exact phase/slice step in `stats.job_log`
under `job_name = 'run_snapshot_backfill_step'`. The `run_key` includes the resolved
request window, current source bounds, batch size, parallel mode, replica-role mode,
and `--skip-entity-counters`, so the same command with unchanged source bounds can
continue or short-circuit safely.

| Window | Mode | First run | Exact rerun | Result |
|---|---|---:|---:|---|
| `2026-02` → `2026-04`, `--skip-entity-counters --no-parallel` | serial | `01:34` | `0.14s` | 20/20 step units skipped on rerun |
| `2026-03` → `2026-04`, `--skip-entity-counters` | parallel | `57.20s` | `0.12s` | 12/12 step units skipped on rerun |

Validation notes:

- The exact serial rerun printed `Plan items: 0` and exited with `Nothing to execute.` in `0.14s`.
- The exact parallel rerun printed `Plan items: 0` and exited with `Nothing to execute.` in `0.12s`.
- The parallel one-month run wrote `12` runner-owned markers and `12` distinct `step_key` values for the resolved `run_key`, proving one marker per completed step with no duplicates.
- A global duplicate audit returned `0` duplicate `step_key` rows across all `run_snapshot_backfill_step` markers.

Operational consequence: the historical backfill can now be stopped and restarted with the same parameters without replaying already-completed steps, which removes the previous restart penalty for long recovery runs.

---
