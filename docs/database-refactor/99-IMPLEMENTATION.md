# Sprint 1 Implementation Log

Summary
- Goal: Implement Sprint 1 foundation migrations (steps 1.2–1.7). The prior exploratory migrations were already reverted per user note.

Sprint 2 Status
- Goal: Implement Sprint 2 counter foundations starting with step 2.1.

Status
- Migration files added to `geokrety-website/website/db/migrations/` for steps 1.2–1.7.
- Migration and pgTAP execution available from this workspace via the project wrappers.
- 2026-03-15 recovery step 1 restored the foundational stats reference and daily tables from the temporary cut: `20260310100300_create_continent_reference.php`, `20260310100600_enable_btree_gist.php`, and grouped migration `20260310200000_create_stats_daily_foundations.php` now cover the former `20260310200000` through `20260310200300` slice in one file.
- The `tests` bootstrap now restores `stats.continent_reference` data from the live DB and reseeds `stats.entity_counters_shard` to the canonical zeroed 25x16 catalog so foundation pgTAP runs stay deterministic after schema-only syncs.
- Recovery step 1 was applied on top of the current chain, schema-synced into `tests`, validated with the focused foundation subset, and then revalidated with the full website pgTAP suite (`86` files, `1581` tests, `PASS`).
- 2026-03-15 recovery step 2 stabilized the grouped previous-move slice under `20260310100105_add_gk_moves_source_columns.php` and `20260310100110_previous_move_backfill_chain.php`, removing the standalone `stats.fn_backfill_km_distance()` helper and deleting `test-246-backfill-km-distance.sql` because exact `km_distance` repair now stays covered on the heavy previous-move entrypoint.
- Recovery step 2 validation: the live DB no longer exposes `stats.fn_backfill_km_distance`, the focused previous-move pgTAP subset passed after schema sync, and the full website pgTAP suite passed again (`85` files, `1566` tests, `PASS`).
- 2026-03-15 recovery step 3 restored the country analytics slice as two grouped migrations: `20260310300000_create_country_stats_tables.php` now covers the former `20260310300000` through `20260310300300` tables, and `20260310300400_create_country_stats_maintenance.php` now covers the former `20260310300400` through `20260310300700` trigger, snapshot, and index maintenance slice.
- Recovery step 3 restored `test-215` through `test-222`, expanded `test-10-stats.sql` from `plan(8)` to `plan(12)` to smoke-check the four new country tables, and intentionally deferred `20260310300800_fix_previous_move_batch_update_rewire.php` because the grouped previous-move chain already passes the current full suite without the standalone compatibility migration.
- Recovery step 3 validation: the grouped country migrations were applied, the schema was synced into `tests`, the focused country pgTAP subset passed, and the full website pgTAP suite passed (`93` files, `1706` tests, `PASS`).
- 2026-03-15 recovery step 4 restored the waypoint and relationship analytics slice in four grouped migrations: `20260310400000_create_waypoint_registry.php` now covers the registry table, source view, and seed helper; `20260310400400_create_waypoint_relationship_tables.php` now covers the cache/relation tables and supporting indexes; `20260310400700_create_waypoint_relationship_triggers.php` now covers both live `gk_moves` maintenance triggers; and `20260310401000_create_waypoint_snapshot_functions.php` now covers the waypoint/cache/relation snapshot helpers and orchestration wrapper.
- Recovery step 4 restored `test-223` through `test-232` and `test-243`, rewrote `test-225-seed-waypoints.sql` to use deterministic GC/OC fixtures instead of production-scale row-count assumptions, and expanded `test-10-stats.sql` from `plan(12)` to `plan(23)` to smoke-check the new waypoint tables, view, and functions.
- Recovery step 4 validation: all four grouped waypoint migrations were applied one at a time, the schema was synced into `tests`, the focused waypoint pgTAP subset passed (`12` files, `169` tests, `PASS`), and the full website pgTAP suite passed again (`104` files, `1863` tests, `PASS`).
- 2026-03-15 recovery step 5 restored the counter-maintenance slice in two grouped migrations: `20260310200100_create_counter_live_maintenance.php` now covers the five live counter and daily-activity trigger stacks plus exact reseed SQL for the move, GeoKret, picture, and user shard catalogs, while `20260310200200_create_counter_snapshot_and_seed_functions.php` now covers `stats.fn_snapshot_entity_counters()` and `stats.fn_seed_daily_activity(tstzrange)`.
- Recovery step 5 restored `test-206`, `test-208`, `test-209`, `test-211`, `test-212`, `test-213`, and `test-214`, expanded `test-10-stats.sql` from `plan(23)` to `plan(25)` to smoke-check the counter snapshot/seed APIs, and tightened the grouped `gk_moves` daily-activity logic so `stats.daily_active_users` is reconciled exactly on delete and update instead of relying on the seed helper for cleanup.
- Recovery step 5 validation: both grouped counter migrations were applied in order, the schema was synced into `tests`, the focused counter pgTAP subset passed (`8` files, `147` tests, `PASS`), and the full website pgTAP suite passed again (`111` files, `1987` tests, `PASS`).
- 2026-03-15 recovery step 6 restored the late Sprint 5 analytics and event surface as grouped migration `20260310500000_create_analytics_event_surface.php`, covering `hourly_activity`, `country_pair_flows`, `gk_milestone_events`, `first_finder_events`, loves rollup maintenance, the points-awarder bridge trigger, milestone and first-finder triggers, batch snapshot functions, and the shared hourly/country-flow indexes.
- Recovery step 6 restored `test-233` through `test-242`, expanded `test-10-stats.sql` from `plan(25)` to `plan(40)` to smoke-check the new tables, functions, and triggers, and deliberately tightened the milestone trigger so user-threshold milestones are computed from canonical `gk_moves` history instead of depending on `stats.gk_related_users` trigger ordering.
- Recovery step 6 validation: the grouped analytics migration was applied, the schema was synced into `tests`, the focused Sprint 5 subset passed (`11` files, `171` tests, `PASS`), the migration was then rolled back to `20260310401000` and re-applied before a rollback/reapply verification subset passed (`4` files, `82` tests, `PASS`), and the full website pgTAP suite passed again (`121` files, `2133` tests, `PASS`).
- 2026-03-15 recovery step 7 restored the orchestration and read-view slice as grouped migration `20260310600400_create_snapshot_orchestration_and_views.php`, covering the scoped hourly/country-flow snapshot overloads, `stats.fn_run_snapshot_phase`, scoped and zero-arg `stats.fn_run_all_snapshots`, scoped and zero-arg `stats.fn_reconcile_stats`, and the canonical 12 `stats.v_uc*` read views.
- Recovery step 7 restored `test-248-snapshot-orchestration.sql` and `test-249-stats-views.sql`, expanded `test-10-stats.sql` from `plan(40)` to `plan(47)` to smoke-check the new orchestration wrappers and representative views, kept the removed standalone km-distance phase out of the orchestration graph, and standardized the full wrapper on the canonical 9-phase list.
- Recovery step 7 validation: the grouped orchestration/views migration was applied, the schema was synced into `tests`, the focused step-7 subset passed (`3` files, `73` tests, `PASS`), the migration was then rolled back and re-applied before the same verification subset passed again (`3` files, `73` tests, `PASS`), and the full website pgTAP suite passed again (`123` files, `2166` tests, `PASS`).
- 2026-03-15 recovery step 8 restored the optional materialized-view accelerators as migration `20260310600600_create_materialized_views.php`, adding `stats.mv_country_month_rollup`, `stats.mv_top_caches_global`, and `stats.mv_global_kpi` with their unique indexes while aligning the country-month index order with the Sprint 6 task spec.
- Recovery step 8 restored `test-250-materialized-views.sql`, expanded `test-10-stats.sql` from `plan(47)` to `plan(50)` to smoke-check the three materialized views, and kept the slice strictly additive so step 7 orchestration ownership did not move again.
- Recovery step 8 validation: the materialized-view migration was applied, the schema was synced into `tests`, the focused MV subset passed (`2` files, `57` tests, `PASS`), all three materialized views successfully completed `REFRESH MATERIALIZED VIEW CONCURRENTLY`, the migration was rolled back and re-applied before the same subset passed again, and the full website pgTAP suite passed again (`124` files, `2176` tests, `PASS`).
- 2026-03-15 recovery step 9 restored the runtime-index residue from the parked `20260310600700` slice as additive migration `20260310600700_add_snapshot_runtime_indexes.php`, keeping step 7's orchestration function ownership intact while reintroducing `idx_gk_moves_qualified_period` and `idx_gk_moves_distance_records` only.
- Recovery step 9 restored `test-251-snapshot-runtime-indexes.sql`, fixed a duplicated TAP block in the parked test file during import, and used the planner assertions to prove the restored indexes support both the replay-order scan and keyed access through `stats.v_uc15_distance_records`.
- Recovery step 9 validation: the runtime-index migration was applied, the schema was synced into `tests`, the focused planner subset passed (`2` files, `56` tests, `PASS`), the migration was rolled back and re-applied before the same subset passed again, and the full website pgTAP suite passed again (`125` files, `2182` tests, `PASS`).
- 2026-03-15 recovery step 10 restored the parked duplicate-index cleanup as corrected migration `20260314101000_remove_redundant_gk_moves_indexes.php`, retargeting the safety gate at the active-chain backfill indexes (`idx_gk_moves_prev_loc_lookup`, `idx_gk_moves_qualified_period`, `idx_gk_moves_geokret_chainlookup`) before renaming `idx_21044_primary` to `gk_moves_pkey` and dropping the legacy duplicate `gk_moves` indexes.
- Recovery step 10 added `test-252-gk-moves-index-cleanup.sql` to verify the removed duplicate indexes stay absent while the critical backfill/runtime indexes remain present, and paired it with the existing planner test so the cleanup did not quietly degrade the restored runtime-index path.
- Recovery step 10 validation: the cleanup migration was applied, the schema was synced into `tests`, the focused cleanup subset passed (`2` files, `11` tests, `PASS`), a direct live-index audit confirmed the targeted duplicates were gone, the migration was rolled back and re-applied before the same subset passed again, and the full website pgTAP suite passed again (`126` files, `2187` tests, `PASS`).
- 2026-03-15 recovery step 11 restored the final low-risk read artifact as migration `20260315101000_create_geokret_move_history_view.php`, reintroducing `geokrety.vw_geokret_move_history` as a convenience projection over `gk_moves` with `previous_position_id`, textual position rendering, and human-readable `move_type_label` values.
- Recovery step 11 added `test-253-geokret-move-history-view.sql` to cover the view surface, fixed an invalid comment-move fixture during authoring by removing the forbidden waypoint value, and left the step isolated from the stats schema so it did not perturb the earlier recovery slices.
- Recovery step 11 validation: the move-history view migration was applied, the schema was synced into `tests`, the focused view test passed (`1` file, `5` tests, `PASS`), the migration was rolled back and re-applied before the same test passed again, and the full website pgTAP suite passed again (`127` files, `2192` tests, `PASS`).
- 2026-03-15 scoped snapshot optimization follow-up rewrote the period-limited snapshot predicates to scalar lower/upper bound comparisons, aligned reconcile-country normalization with `geokrety.fn_normalize_country_code()`, made scoped hourly and country-flow runs rebuild full touched buckets, added timing/job-log coverage to the remaining backfill helpers, and introduced migration `20260315171920_add_scoped_snapshot_backfill_indexes.php` for the four scoped backfill indexes.
- 2026-03-15 scoped snapshot optimization validation: rolled back to `20260310200100` once to replay the edited grouped migrations, rollback-verified and re-applied `20260315171920`, rerolled the `20260310401000` tail while tuning the scoped waypoint/cache path, schema-synced into `tests`, passed the focused pgTAP subset (`17` files, `262` tests, `PASS`) plus a final narrow regression subset (`5` files, `80` tests, `PASS`), and bench-ran the `2007-10-01` to `2008-02-01` window with final timings of `40 ms` (`fn_seed_daily_activity`), `598 ms` (`fn_snapshot_daily_country_stats`), `194 ms` (`fn_snapshot_user_country_stats`), `23 ms` (`fn_snapshot_gk_country_stats`), `17335 ms` (`fn_snapshot_relationship_tables`), `1548 ms` (`fn_snapshot_hourly_activity`), and `2248 ms` (`fn_snapshot_country_pair_flows`).
- 2026-03-16 backfill continuation hardening updated `docs/database-refactor/run_snapshot_backfill.py` to persist exact-run phase markers in `stats.job_log` (`job_name = 'run_snapshot_backfill_step'`), filter already-completed step keys on restart, guard the resolved request/source window with a PostgreSQL advisory lock, and expose `--no-resume` plus `--clear-resume-markers` for operator control.
- 2026-03-16 backfill continuation validation: compile-checked the runner, benchmarked the default parallel one-month path (`2026-03` → `2026-04`) at `57.20s` on first run and `0.12s` on the exact rerun, benchmarked the serial two-month path with `--skip-entity-counters --no-parallel` at `01:34` on first run and `0.14s` on the exact rerun, and verified the runner-owned marker surface had `0` duplicate `step_key` values in `stats.job_log`.
- 2026-03-16 first-finder live hardening added migration `20260316133000_harden_first_finder_live_reconciliation.php`, which replaces the append-only `stats.fn_detect_first_finder(...)` behavior with per-GeoKret reconciliation ordered by `(moved_on_datetime, id)`, synchronizes `stats.gk_milestone_events(event_type = 'first_find')` in the same unit of work, and extends live maintenance to `gk_moves` insert/update/delete plus `gk_geokrety.owner` / `created_on_datetime` changes.
- 2026-03-16 first-finder live hardening validation: PHP syntax passed, `pre-commit run -a` passed, the migration was applied then rolled back and re-applied successfully, the schema was synced into `tests`, the focused first-finder subset passed (`5` files, `77` tests, `PASS`), and a live sanity audit confirmed `stats.first_finder_events` still matches the canonical source-derived count exactly (`4421` rows) while closing the previously identified out-of-order insert drift risk.
- 2026-03-16 full-suite note: the post-change full website pgTAP run failed only in unrelated country/index files already modified in the dirty worktree (`test-219-country-rollups-trigger.sql`, `test-220-country-history-trigger.sql`, `test-254-scoped-snapshot-backfill-indexes.sql`), so those regressions were left untouched while the first-finder hardening remained green on its owned validation surface.
- Remaining parked migrations have now been fully accounted for: `20260310600800_optimize_previous_move_backfill_batches.php` stays omitted because its trigger/backfill rewrites were superseded by the active grouped previous-move chain, and `20260314101100_add_geokret_covering_index.php` stays omitted because its index was already absorbed earlier by `20260310100500_create_source_table_indexes.php`.
- 2026-03-15 intermediary cleanup completed for the previous-move release cut: the active chain was reduced to `20260310100100`, `20260310100200`, `20260310100400`, `20260310100500`, and `20260315091342`, while superseded March stats migrations/tests were moved out of the active directories into `/home/kumy/GIT/geokrety-stats/tmp/20260315-intermediary-cut/`.
- The reduced chain was re-applied from the `20260220130100_add_loved_geokrety_amqp_trigger` baseline, schema-synced into `tests`, validated with the targeted split-chain pgTAP subset, and then validated again with the full website pgTAP suite (`79` files, `1474` tests, `PASS`).
- Migration `20260315091342_squash_previous_move_backfill_chain.php` now replaces the interim backfill/MV refactor chain, has been rollback-verified, schema-synced, and validated with targeted split-chain pgTAP runs.
- Sprint 2 step 2.1 has been applied, rolled back, re-applied, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.2 has been applied, rolled back, re-applied, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.3 has been applied, rolled back, re-applied, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.4 has been applied, rolled back, re-applied, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.5 has been applied, rolled back, re-applied, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.6 has been applied, rolled back, re-applied, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.7 has been applied, rolled back, re-applied, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.8 has been applied, rolled back, re-applied, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.9 has been implemented, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.10 has been implemented, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.11 has been implemented, schema-synced, and validated with targeted plus full website pgTAP runs.
- Sprint 2 step 2.12 has been implemented, schema-synced, and validated with targeted plus full website pgTAP runs.

Accomplished Steps
// YYYYMMDD-HHMM
- 20260313-2012 Created migration: `20260310100100_create_stats_schema.php` (creates `stats` schema).
- 20260313-2012 Created migration: `20260310100200_create_operational_support_tables.php` (creates `stats.backfill_progress`, `stats.job_log`).
- 20260313-2013 Created migration: `20260310100300_create_continent_reference.php` (creates and seeds `stats.continent_reference` with 249 entries using ON CONFLICT DO NOTHING).
- 20260313-2013 Created migration: `20260310100400_add_gk_moves_source_columns.php` (adds `previous_move_id` BIGINT and `km_distance` NUMERIC(8,3) plus deferrable FK).
- 20260313-2013 Created migration: `20260310100500_create_source_table_indexes.php` (creates five indexes on `geokrety.gk_moves` with CONCURRENTLY where specified).
- 20260313-2013 Created migration: `20260310100600_enable_btree_gist.php` (enables `btree_gist` extension).
- 20260315-1415 Recovery step 1 restored `20260310100300_create_continent_reference.php` and `20260310100600_enable_btree_gist.php`, replaced the former `20260310200000` through `20260310200300` file set with grouped migration `20260310200000_create_stats_daily_foundations.php`, restored `test-10-continent-reference.sql`, `test-10-enable-btree_gist.sql`, `test-10-stats.sql`, and `test-201` through `test-204`, then updated the test-schema copy script to load continent reference rows and reseed `stats.entity_counters_shard` to the canonical zero state in `tests`.
- 20260315-1415 Recovery step 1 validation: targeted foundation pgTAP passed after schema sync, then the full website pgTAP suite passed (`86` files, `1581` tests, `PASS`).
- 20260315-1455 Recovery step 2 kept the grouped previous-move chain under `20260310100105_add_gk_moves_source_columns.php` and `20260310100110_previous_move_backfill_chain.php`, removed the standalone `stats.fn_backfill_km_distance()` wrapper from the grouped migration, deleted `test-246-backfill-km-distance.sql`, and moved the exact heavy-path `km_distance` assertion into `test-245-backfill-previous-move-heavy.sql`.
- 20260315-1455 Recovery step 2 validation: confirmed `stats.fn_backfill_km_distance` is absent in the live DB, then passed both the focused previous-move subset and the full website pgTAP suite (`85` files, `1566` tests, `PASS`).
- 20260315-1545 Recovery step 3 restored the country analytics domain as grouped migrations `20260310300000_create_country_stats_tables.php` and `20260310300400_create_country_stats_maintenance.php`, reintroduced `test-215-country-daily-stats.sql` through `test-222-country-indexes.sql`, and expanded `test-10-stats.sql` to smoke-check the new country tables.
- 20260315-1545 Recovery step 3 validation: applied both grouped country migrations, synced the schema into `tests`, passed the focused country subset, and then passed the full website pgTAP suite (`93` files, `1706` tests, `PASS`).
- 20260315-1705 Recovery step 4 restored the waypoint analytics domain as grouped migrations `20260310400000_create_waypoint_registry.php`, `20260310400400_create_waypoint_relationship_tables.php`, `20260310400700_create_waypoint_relationship_triggers.php`, and `20260310401000_create_waypoint_snapshot_functions.php`, then reintroduced `test-223-waypoints.sql` through `test-232-cache-relation-indexes.sql` plus `test-243-waypoint-relation-snapshots.sql` and expanded `test-10-stats.sql` to smoke-check the waypoint slice.
- 20260315-1705 Recovery step 4 validation: applied all four grouped waypoint migrations one at a time, synced the schema into `tests`, passed the focused waypoint subset (`12` files, `169` tests, `PASS`), and then passed the full website pgTAP suite (`104` files, `1863` tests, `PASS`).
- 20260315-1815 Recovery step 5 restored the counter-maintenance domain as grouped migrations `20260310200100_create_counter_live_maintenance.php` and `20260310200200_create_counter_snapshot_and_seed_functions.php`, reintroduced `test-206-gk-moves-counter-trigger.sql`, `test-208-gk-moves-daily-activity-trigger.sql`, `test-209-gk-geokrety-counter-trigger.sql`, `test-211-gk-pictures-counter-trigger.sql`, `test-212-gk-users-counter-trigger.sql`, `test-213-entity-counter-snapshot.sql`, and `test-214-daily-activity-seed.sql`, and expanded `test-10-stats.sql` to smoke-check the restored counter snapshot/seed functions.
- 20260315-1815 Recovery step 5 validation: applied both grouped counter migrations, synced the schema into `tests`, passed the focused counter subset (`8` files, `147` tests, `PASS`), and then passed the full website pgTAP suite (`111` files, `1987` tests, `PASS`).
- 20260315-2005 Recovery step 6 restored the late Sprint 5 analytics and event slice as grouped migration `20260310500000_create_analytics_event_surface.php`, reintroduced `test-233-hourly-activity.sql` through `test-242-analytics-indexes.sql`, and expanded `test-10-stats.sql` to smoke-check the new analytics tables, batch functions, and trigger endpoints.
- 20260315-2005 Recovery step 6 validation: applied the grouped analytics migration, synced the schema into `tests`, passed the focused Sprint 5 subset (`11` files, `171` tests, `PASS`), rollback-verified the migration by reverting to `20260310401000` and re-applying it before a verification subset passed (`4` files, `82` tests, `PASS`), and then passed the full website pgTAP suite (`121` files, `2133` tests, `PASS`).
- 20260315-2135 Recovery step 7 restored the orchestration and read-view slice as grouped migration `20260310600400_create_snapshot_orchestration_and_views.php`, reintroduced `test-248-snapshot-orchestration.sql` and `test-249-stats-views.sql`, and expanded `test-10-stats.sql` to smoke-check the orchestration wrappers and representative `stats.v_uc*` views.
- 20260315-2135 Recovery step 7 validation: applied the grouped orchestration/views migration, synced the schema into `tests`, fixed a duplicated `test-248` TAP block uncovered during the first focused run, then passed the focused subset (`3` files, `73` tests, `PASS`), rollback-verified the migration by reverting and re-applying it before the same subset passed again, and then passed the full website pgTAP suite (`123` files, `2166` tests, `PASS`).
- 20260315-2215 Recovery step 8 restored the optional materialized-view accelerators as `20260310600600_create_materialized_views.php`, reintroduced `test-250-materialized-views.sql`, and expanded `test-10-stats.sql` to smoke-check the three `stats.mv_*` objects.
- 20260315-2215 Recovery step 8 validation: applied the MV migration, synced the schema into `tests`, passed the focused MV subset (`2` files, `57` tests, `PASS`), confirmed `REFRESH MATERIALIZED VIEW CONCURRENTLY` succeeds for all three views, rollback-verified the migration by reverting and re-applying it before the same subset passed again, and then passed the full website pgTAP suite (`124` files, `2176` tests, `PASS`).
- 20260315-2245 Recovery step 9 restored the runtime-index residue from the parked `20260310600700` slice as additive migration `20260310600700_add_snapshot_runtime_indexes.php`, reintroduced `test-251-snapshot-runtime-indexes.sql`, and corrected a duplicated TAP block that was imported from the parked test file.
- 20260315-2245 Recovery step 9 validation: applied the runtime-index migration, synced the schema into `tests`, passed the focused planner subset (`2` files, `56` tests, `PASS`), rollback-verified the migration by reverting and re-applying it before the same subset passed again, and then passed the full website pgTAP suite (`125` files, `2182` tests, `PASS`).
- 20260315-2325 Recovery step 10 restored the parked duplicate-index cleanup as corrected migration `20260314101000_remove_redundant_gk_moves_indexes.php`, adding `test-252-gk-moves-index-cleanup.sql` so the renamed primary key and removed duplicate index sets are now covered by pgTAP.
- 20260315-2325 Recovery step 10 validation: applied the cleanup migration, synced the schema into `tests`, passed the focused planner/cleanup subset (`2` files, `11` tests, `PASS`), confirmed the targeted duplicate indexes were gone in the live DB, rollback-verified the migration by rebuilding and then re-removing the duplicates, and then passed the full website pgTAP suite (`126` files, `2187` tests, `PASS`).
- 20260315-2355 Recovery step 11 restored `20260315101000_create_geokret_move_history_view.php`, added `test-253-geokret-move-history-view.sql`, and corrected the test fixture so the comment move respects the website constraint that move_type `2` rows cannot carry a waypoint.
- 20260315-2355 Recovery step 11 validation: applied the move-history view migration, synced the schema into `tests`, passed the focused view test (`1` file, `5` tests, `PASS`), rollback-verified the migration by dropping and re-creating the view, and then passed the full website pgTAP suite (`127` files, `2192` tests, `PASS`).
- 20260315-2355 Final parked-slice accounting: `20260310600800_optimize_previous_move_backfill_batches.php` remained intentionally omitted as superseded by the active grouped previous-move chain, and `20260314101100_add_geokret_covering_index.php` remained intentionally omitted because `idx_gk_moves_geokret_chainlookup` was already owned by `20260310100500_create_source_table_indexes.php`.
- 20260313-2045 Created migration: `20260310200000_create_entity_counters_shard.php` (creates and pre-seeds `stats.entity_counters_shard` with 25 entities x 16 shards).
- 20260313-2045 Added pgTAP test: `test-201-entity-counters-shard.sql` (covers schema shape, PK, types, defaults, seed cardinality, and zero-initialization).
- 20260313-2054 Applied `20260310200000_create_entity_counters_shard.php`, verified `down()` with rollback, then re-applied it successfully.
- 20260313-2054 Ran targeted pgTAP (`test-10-schema.sql`, `test-201-entity-counters-shard.sql`) and full website pgTAP (`78` files, `1390` tests, `PASS`) after syncing schema into the `tests` database.
- 20260313-2125 Created migration: `20260310200100_create_daily_activity.php` (creates `stats.daily_activity` for per-day global metrics).
- 20260313-2125 Added pgTAP test: `test-202-daily-activity.sql` (covers table shape, exact column set, comments, insert/read-back, defaults, and duplicate rejection).
- 20260313-2134 Applied `20260310200100_create_daily_activity.php`, verified `down()` with rollback, then re-applied it successfully.
- 20260313-2134 Ran targeted pgTAP (`test-10-schema.sql`, `test-10-stats.sql`, `test-202-daily-activity.sql`) and full website pgTAP (`79` files, `1422` tests, `PASS`) after syncing schema into the `tests` database.
- 20260313-2143 Created migration: `20260310200200_create_daily_active_users.php` (creates `stats.daily_active_users` for per-day user presence analytics).
- 20260313-2143 Added pgTAP test: `test-203-daily-active-users.sql` (covers PK, exact columns, comment, insert/read-back, and duplicate/null rejection).
- 20260313-2149 Applied `20260310200200_create_daily_active_users.php`, verified `down()` with rollback, then re-applied it successfully.
- 20260313-2149 Ran targeted pgTAP (`test-10-schema.sql`, `test-10-stats.sql`, `test-203-daily-active-users.sql`) and full website pgTAP (`80` files, `1438` tests, `PASS`) after syncing schema into the `tests` database.
- 20260313-2230 Created migration: `20260310200300_create_daily_entity_counts.php` (creates `stats.daily_entity_counts` for per-day entity snapshot totals).
- 20260313-2230 Added pgTAP test: `test-204-daily-entity-counts.sql` (covers schema shape, comments, defaults, insert/read-back, duplicate rejection, and upsert behavior).
- 20260313-2247 Applied `20260310200300_create_daily_entity_counts.php`, verified `down()` with rollback, then re-applied it successfully.
- 20260313-2247 Ran targeted pgTAP (`test-10-schema.sql`, `test-10-stats.sql`, `test-204-daily-entity-counts.sql`) and full website pgTAP (`81` files, `1463` tests, `PASS`) after syncing schema into the `tests` database.
- 20260313-2310 Created migration: `20260310200400_create_previous_move_trigger.php` (creates `geokrety.fn_set_previous_move_id_and_distance()` and attaches `tr_gk_moves_before_prev_move` on `geokrety.gk_moves`).
- 20260313-2310 Added pgTAP test: `test-205-previous-move-trigger.sql` (covers function/trigger existence, insert/update/delete behavior, NULL-guard branches, and km_distance rounding).
- 20260314-0008 Applied `20260310200400_create_previous_move_trigger.php`, verified `down()` with rollback, then re-applied it successfully.
- 20260314-0010 Ran targeted pgTAP (`test-205-previous-move-trigger.sql`, `PASS`) and full website pgTAP (`82` files, `1488` tests, `PASS`) after syncing schema into the `tests` database.
- 20260314-0010 Logged open question in `99-OPEN-QUESTIONS.md`: Sprint 2 treats move type `1` as qualifying for previous-move distance, while older helper/backfill contracts still only count `0,3,5`.
- 20260314-0829 Created migration: `20260310200500_create_gk_moves_counter_trigger.php` (creates `geokrety.fn_gk_moves_sharded_counter()` and attaches `tr_gk_moves_after_sharded_counters` on `geokrety.gk_moves`).
- 20260314-0829 Added pgTAP test: `test-206-gk-moves-counter-trigger.sql` (covers function/trigger existence, insert/delete deltas, shard selection, all six move types, a no-op update, and a move_type update).
- 20260314-0829 Logged open question in `99-OPEN-QUESTIONS.md`: seed `gk_moves` shard counters before historical updates or deletes touch zero-seeded rows.
- 20260314-0829 Applied `20260310200500_create_gk_moves_counter_trigger.php`, fixed decrement upserts to satisfy the nonnegative shard-counter constraint, verified `down()` with rollback twice, then re-applied it successfully.
- 20260314-0829 Ran targeted pgTAP (`test-206-gk-moves-counter-trigger.sql`, `PASS`) and full website pgTAP (`83` files, `1520` tests, `PASS`) after syncing schema into the `tests` database.
- 20260314-0830 Refactored `test-206-gk-moves-counter-trigger.sql` to keep `after_20_last_log_and_position` and `before_40_update_missing` enabled, proving coexistence with the preexisting move trigger stack while still suppressing AMQP side effects.
- 20260314-1002 Refactored `20260310200400_create_previous_move_trigger.php` so delete rewiring runs in `tr_gk_moves_after_prev_move_delete` as an `AFTER DELETE ... FOR EACH STATEMENT` trigger using `OLD TABLE`, which resolves the bulk `DELETE FROM gk_moves` conflict revealed during Sprint 2 Task 2.6 integration.
- 20260314-1002 Added pgTAP regression test: `test-207-previous-move-delete-rewire.sql` (covers bulk delete safety, multi-row rewiring from final table state, and multi-GeoKret partitioning during one delete statement).
- 20260314-1002 Re-verified `20260310200400_create_previous_move_trigger.php` and `20260310200500_create_gk_moves_counter_trigger.php` with rollback/re-apply cycles, then ran targeted pgTAP (`test-205-previous-move-trigger.sql`, `test-206-gk-moves-counter-trigger.sql`, `test-207-previous-move-delete-rewire.sql`, `test-21-stats-update.sql`, `PASS`) and full website pgTAP (`84` files, `1519` tests, `PASS`) after syncing schema into the `tests` database.
- 20260314-1150 Created migration: `20260310200600_create_gk_moves_daily_trigger.php` (creates `geokrety.fn_gk_moves_daily_activity()`, helper date refresh function, and attaches `tr_gk_moves_after_daily_activity` on `geokrety.gk_moves`).
- 20260314-1150 Added pgTAP test: `test-208-gk-moves-daily-activity-trigger.sql` (covers function/trigger existence, daily counters for inserts/updates/deletes, author presence rules, and km-contribution refreshes through the existing previous-move trigger stack).
- 20260314-1150 Applied `20260310200600_create_gk_moves_daily_trigger.php`, verified `down()` with rollback, then re-applied it successfully.
- 20260314-1150 Ran targeted pgTAP (`test-205-previous-move-trigger.sql`, `test-206-gk-moves-counter-trigger.sql`, `test-207-previous-move-delete-rewire.sql`, `test-208-gk-moves-daily-activity-trigger.sql`, `test-21-stats-update.sql`, `PASS`) and full website pgTAP (`85` files, `1538` tests, `PASS`) after syncing schema into the `tests` database.
- 20260314-1240 Created migration: `20260310200700_create_gk_geokrety_counter_trigger.php` (creates `geokrety.fn_gk_geokrety_counter()`, helper date refresh function, and attaches `tr_gk_geokrety_counters` on `geokrety.gk_geokrety`).
- 20260314-1240 Added pgTAP test: `test-209-gk-geokrety-counter-trigger.sql` (covers function/trigger existence, total/type shard counters, delete reconciliation, all 11 GK types, and `daily_activity.gk_created`).
- 20260314-1240 Applied `20260310200700_create_gk_geokrety_counter_trigger.php`, verified `down()` with rollback, then re-applied it successfully.
- 20260314-1240 Ran targeted pgTAP (`test-202-daily-activity.sql`, `test-208-gk-moves-daily-activity-trigger.sql`, `test-209-gk-geokrety-counter-trigger.sql`, `PASS`) and full website pgTAP (`86` files, `1560` tests, `PASS`) after syncing schema into the `tests` database.
- 20260314-1335 Created migration: `20260310200800_create_gk_pictures_counter_trigger.php` (creates upload-driven picture shard counters, per-type shard counters, daily-activity refresh helpers, and self-seeds historical picture shard rows before enabling the trigger).
- 20260314-1335 Added pgTAP test: `test-211-gk-pictures-counter-trigger.sql` (covers draft-vs-uploaded semantics, type transitions, date transitions, delete reconciliation, and `daily_activity` refreshes).
- 20260314-1405 Created migration: `20260310200900_create_gk_users_counter_trigger.php` (creates `geokrety.fn_gk_users_counter()`, daily registration refresh helper, and self-seeds `gk_users` shard rows before enabling the trigger).
- 20260314-1405 Added pgTAP test: `test-212-gk-users-counter-trigger.sql` (covers insert/delete deltas, shard selection by `id % 16`, and `daily_activity.users_registered`).
- 20260314-1440 Created migration: `20260310201000_create_entity_counter_snapshot.php` (creates `stats.fn_snapshot_entity_counters()` and reseeds the canonical 25 entities across all 16 shards using real `id % 16` distribution).
- 20260314-1440 Added pgTAP test: `test-213-entity-counter-snapshot.sql` (covers function existence, canonical entity catalog, 400-row snapshot shape, source-table totals, and idempotence).
- 20260314-1515 Created migration: `20260310201100_create_daily_activity_seed.php` (creates `stats.fn_seed_daily_activity(tstzrange)` with whole-day `[)` validation, exact refresh semantics, and preservation of unrelated `daily_activity` metrics during ranged re-seeds).
- 20260314-1515 Added pgTAP test: `test-214-daily-activity-seed.sql` (covers full-history seeding, range-limited seeding, idempotence, `daily_active_users`, uploaded-picture semantics, and invalid-range rejection).
- 20260314-1645 Applied Sprint 2 steps 2.9-2.12 to the development database, synced schema into `tests`, and iterated on the new migrations/tests until targeted pgTAP for `test-201`, `test-211`, `test-212`, `test-213`, and `test-214` passed.
- 20260314-1705 Created follow-up migration: `20260314103000_fix_previous_move_batch_update_rewire.php` to resolve legacy `gk_moves` bulk-update regressions uncovered by the full suite. The fix rewires both the legacy `after_30_distances` path and previous-move successor refreshes to statement-level recomputation, while preserving the old `UPDATE OF geokret, lat, lon, moved_on_datetime, move_type, position` behavior inside the trigger function so manual `distance` edits still work.
- 20260314-1715 Revalidated the legacy regression surface with targeted pgTAP (`test-43-functions-update-next-move-distance_on_gk_changed.sql`, `test-21-stats-update.sql`, `PASS`).
- 20260314-1725 Full website pgTAP passed after syncing schema into the `tests` database (`90` files, `1619` tests, `PASS`).
- 20260314-1725 Note: the shared Phinx wrapper targeted a different checkout during one rollback attempt and reverted `20260310201100` in the live dev database. The local migration SQL from `20260310201100_create_daily_activity_seed.php` and `20260314103000_fix_previous_move_batch_update_rewire.php` was then applied directly to restore the intended schema state before the final green pgTAP run.
- 20260315-0836 Updated migration: `20260314102104_split_previous_move_and_position_chain.php` (heavy previous-move backfill now refreshes the MV once per heavy run, batches by month with explicit `p_month_limit`, drops stale overloads before recreating wrappers, restores legacy state during rollback before dropping `previous_position_id`, and refreshes the MV on direct helper calls only when it is still unpopulated).
- 20260315-0836 Updated pgTAP tests: `test-245-backfill-previous-move-heavy.sql`, `test-247-backfill-km-distance-heavy.sql`, and `test-251-snapshot-runtime-indexes.sql` (adds bounded/unlimited heavy backfill coverage, one-argument compatibility checks, delegated km-distance logging coverage, fixture cleanup, and planner-stable index assertions).
- 20260315-0836 Re-applied `20260314102104_split_previous_move_and_position_chain.php`, verified `down()` with rollback twice after the final edits, synced the website schema into `tests`, and ran targeted backfill pgTAP plus the full website DB suite (`128` files, `2188` tests, `PASS`).
- 20260315-0950 Squashed the backfill/MV refactor chain into `20260315091342_squash_previous_move_backfill_chain.php`, moved superseded migrations `20260314101200` through `20260314102104` out of the active migration directory into `/home/kumy/GIT/geokrety-stats/tmp/`, and kept only the single heavy previous-move entrypoint plus split-chain online triggers.
- 20260315-0950 Finalized the squashed migration so `stats.mv_backfill_working_set` is created `WITH NO DATA`, refreshed only when `NOT ispopulated`, reused only within one heavy run, and reset back to `WITH NO DATA` afterwards to avoid stale populated-MV reuse.
- 20260315-0950 Re-applied `20260315091342_squash_previous_move_backfill_chain.php`, verified `down()` with rollback and re-apply, synced the website schema into `tests`, and ran targeted pgTAP (`test-205`, `test-240`, `test-244`, `test-245`, `test-246`, `test-247`, `test-248`, `PASS`).
- 20260316-1145 Hardened `docs/database-refactor/run_snapshot_backfill.py` for exact-run continuation: added runner-owned resume markers in `stats.job_log`, filtered completed step keys out of subsequent plans, added an advisory lock over the resolved request/source window, and exposed `--no-resume` plus `--clear-resume-markers`.
- 20260316-1145 Validated continuation semantics end to end: the default parallel `2026-03` → `2026-04` run with `--skip-entity-counters` completed in `57.20s`, the exact rerun exited in `0.12s` with `0/12` runnable step units, the serial `2026-02` → `2026-04` run with `--skip-entity-counters --no-parallel` reran in `0.14s` with `0/20` runnable step units, and a duplicate-marker audit over `stats.job_log` returned `0` duplicate `step_key` values.
- 20260316-1330 Created migration: `20260316133000_harden_first_finder_live_reconciliation.php` (adds per-GeoKret canonical first-finder reconciliation, synchronizes `first_find` milestone rows, broadens `gk_moves` trigger coverage to insert/update/delete, and adds `gk_geokrety` owner/created-at reconciliation hooks).
- 20260316-1330 Added pgTAP test: `test-258-first-finder-live-reconciliation.sql` (covers backdated inserts, loser deletes, winner disqualification fallback, owner-change invalidation, GeoKret reassignment, milestone synchronization, and the exact 168-hour eligibility boundary).
- 20260316-1330 Verified the suspicious `stats.first_finder_events` count against the canonical source query and confirmed the current live table is correct at `4421` rows; the real issue was future drift because append-only live maintenance could preserve a later-inserted but earlier-dated qualifying move incorrectly.
- 20260316-1330 Applied `20260316133000_harden_first_finder_live_reconciliation.php`, verified `down()` with rollback, re-applied it successfully, synced the schema into `tests`, and passed the focused first-finder/milestone subset (`test-236`, `test-240`, `test-256`, `test-257`, `test-258`; `77` tests, `PASS`).
- 20260316-1330 Recorded unrelated full-suite regressions after the focused pass: `test-219-country-rollups-trigger.sql`, `test-220-country-history-trigger.sql`, and `test-254-scoped-snapshot-backfill-indexes.sql` failed in the dirty worktree, so they were not modified as part of this first-finder hardening change.

Checklist & Comments
- [x] 1. Revert preliminary migrations (user indicated already reverted) — verify in DB.
- [x] 2. Migration file: `20260310100100_create_stats_schema.php` — added.
- [x] 3. Migration file: `20260310100200_create_operational_support_tables.php` — added.
- [x] 4. Migration file: `20260310100300_create_continent_reference.php` — added (includes full seed per spec).
- [x] 5. Migration file: `20260310100400_add_gk_moves_source_columns.php` — added.
- [x] 6. Migration file: `20260310100500_create_source_table_indexes.php` — added. Note: `CREATE INDEX CONCURRENTLY` cannot run inside a transaction; run-time environment must allow non-transactional execution for these statements.
- [x] 7. Migration file: `20260310100600_enable_btree_gist.php` — added.
- [x] 8. Run `phinx migrate` against the website DB (recommended: run from the website repo with DB credentials). This will actually apply the new migrations.
- [x] 9. Run pgTAP assertions T-1.1.001 through T-1.7.010 to validate schema and seeds. These tests expect a live DB and pgTAP installed.
- [x] 10. Verify `stats.continent_reference` has ~249 rows and continent counts.
- [x] 11. Verify `geokrety.gk_moves` contains the new columns and FK is `DEFERRABLE INITIALLY DEFERRED`.
- [x] 12. Verify indexes exist and, if necessary, rebuild them without CONCURRENTLY if your migration runner enforces transactions.

Sprint 2 Checklist & Comments
- [x] 1. Create migration file: `20260310200000_create_entity_counters_shard.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify table exists with 3 columns and correct types.
- [x] 4. Verify composite PK on `(entity, shard)`.
- [x] 5. Verify 400 rows were inserted (25 entities × 16 shards).
- [x] 6. Verify all `cnt` values are `0`.
- [x] 7. Ensure script /home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh copy the data if necessary.
- [x] 8. Run pgTAP tests T-2.1.001 through T-2.1.010.

Sprint 2 Task 2.2 Checklist & Comments
- [x] 1. Create migration file `20260310200100_create_daily_activity.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify table exists with 17 columns and correct types.
- [x] 4. Verify PK on `activity_date`.
- [x] 5. Test insert and read-back.
- [x] 6. Ensure script /home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh copy the data if necessary.
- [x] 7. Run pgTAP tests T-2.2.001 through T-2.2.010.

Sprint 2 Task 2.3 Checklist & Comments
- [x] 1. Create migration file `20260310200200_create_daily_active_users.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify table exists with 2 columns and correct types.
- [x] 4. Verify composite PK on `(activity_date, user_id)`.
- [x] 5. Test insert and duplicate rejection.
- [x] 6. Ensure script /home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh copy the data if necessary.
- [x] 7. Run pgTAP tests T-2.3.001 through T-2.3.008.

Sprint 2 Task 2.4 Checklist & Comments
- [x] 1. Create migration file `20260310200300_create_daily_entity_counts.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify table exists with 3 columns and correct types.
- [x] 4. Verify composite PK on `(count_date, entity)`.
- [x] 5. Test insert, read-back, and upsert behavior.
- [x] 6. Ensure script /home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh copy the data if necessary.
- [x] 7. Run pgTAP tests T-2.4.001 through T-2.4.009.

Sprint 2 Task 2.5 Checklist & Comments
- [x] 1. Verify `geokrety.gk_moves.previous_move_id` and `km_distance` columns exist (Sprint 1, Step 1.5).
- [x] 2. Create migration file `20260310200400_create_previous_move_trigger.php`.
- [x] 3. Run `phinx migrate`.
- [x] 4. Verify function `geokrety.fn_set_previous_move_id_and_distance()` exists.
- [x] 5. Verify trigger `tr_gk_moves_before_prev_move` exists on `gk_moves`.
- [x] 6. Test: INSERT drop after prior drop -> `previous_move_id` set, `km_distance > 0`.
- [x] 7. Test: COMMENT/ARCHIVE moves -> `previous_move_id` and `km_distance` remain NULL.
- [x] 8. Test: Move with NULL position -> `km_distance` remains NULL.
- [x] 9. Test: First move for a GK -> `previous_move_id` remains NULL.
- [x] 10. Test: UPDATE recalculates `previous_move_id` and `km_distance`.
- [x] 11. Test: DELETE path executes safely for both single-row and bulk delete statements.
- [x] 12. Run pgTAP tests T-2.5.001 through T-2.5.012.

Sprint 2 Task 2.6 Checklist & Comments
- [x] 1. Verify `stats.entity_counters_shard` table exists (Step 2.1).
- [x] 2. Create migration file `20260310200500_create_gk_moves_counter_trigger.php`.
- [x] 3. Run `phinx migrate`.
- [x] 4. Verify function `geokrety.fn_gk_moves_sharded_counter` exists.
- [x] 5. Verify trigger `tr_gk_moves_after_sharded_counters` exists on `gk_moves`.
- [x] 6. Test: INSERT increments counters.
- [x] 7. Test: DELETE decrements counters.
- [x] 8. Test: Correct shard row updated based on `id % 16`.
- [x] 9. Test: UPDATE reverses old typed contribution and applies the new typed contribution while legacy move triggers remain enabled.
- [x] 10. Run pgTAP tests T-2.6.001 through T-2.6.010.

Sprint 2 Task 2.7 Checklist & Comments
- [x] 1. Verify `stats.daily_activity` table exists (Step 2.2).
- [x] 2. Verify `stats.daily_active_users` table exists (Step 2.3).
- [x] 3. Create migration file `20260310200600_create_gk_moves_daily_trigger.php`.
- [x] 4. Run `phinx migrate`.
- [x] 5. Verify function `geokrety.fn_gk_moves_daily_activity` exists.
- [x] 6. Verify trigger `tr_gk_moves_after_daily_activity` exists on `gk_moves`.
- [x] 7. Test INSERT DROP increments `daily_activity.drops`.
- [x] 8. Test INSERT with author creates `daily_active_users` row.
- [x] 9. Test INSERT with NULL author skips `daily_active_users`.
- [x] 10. Test UPDATE reconciles old/new daily buckets exactly.
- [x] 11. Test DELETE decrements `daily_activity` counters.
- [x] 12. Run pgTAP tests T-2.7.001 through T-2.7.014.

Sprint 2 Task 2.8 Checklist & Comments
- [x] 1. Verify `stats.entity_counters_shard` and `stats.daily_activity` tables exist (Steps 2.1, 2.2).
- [x] 2. Create migration file `20260310200700_create_gk_geokrety_counter_trigger.php`.
- [x] 3. Run `phinx migrate`.
- [x] 4. Verify function `geokrety.fn_gk_geokrety_counter` exists.
- [x] 5. Verify trigger `tr_gk_geokrety_counters` exists on `gk_geokrety`.
- [x] 6. Test INSERT increments shard counters and `daily_activity.gk_created`.
- [x] 7. Test DELETE decrements counters and refreshes `daily_activity.gk_created`.
- [x] 8. Run pgTAP tests T-2.8.001 through T-2.8.008.

Sprint 2 Task 2.9 Checklist & Comments
- [x] 1. Verify `stats.entity_counters_shard` and `stats.daily_activity` tables exist (Steps 2.1, 2.2).
- [x] 2. Create migration file `20260310200800_create_gk_pictures_counter_trigger.php`.
- [x] 3. Verify function `geokrety.fn_gk_pictures_counter` exists.
- [x] 4. Verify trigger `tr_gk_pictures_after_counter` exists on `gk_pictures`.
- [x] 5. Test uploaded pictures increment total/type shard counters while draft rows stay out of the counters.
- [x] 6. Test UPDATE reclassifies picture type and upload day correctly in both shard counters and `daily_activity`.
- [x] 7. Test DELETE decrements counters and refreshes `daily_activity` exactly.
- [x] 8. Run pgTAP tests T-2.9.001 through T-2.9.016.

Sprint 2 Task 2.10 Checklist & Comments
- [x] 1. Verify `stats.entity_counters_shard` and `stats.daily_activity` tables exist (Steps 2.1, 2.2).
- [x] 2. Create migration file `20260310200900_create_gk_users_counter_trigger.php`.
- [x] 3. Verify function `geokrety.fn_gk_users_counter` exists.
- [x] 4. Verify trigger `tr_gk_users_activity` exists on `gk_users`.
- [x] 5. Test INSERT increments the `gk_users` shard chosen by `id % 16` and refreshes `daily_activity.users_registered`.
- [x] 6. Test DELETE decrements the total/shard counters and refreshes `daily_activity.users_registered` back down.
- [x] 7. Run pgTAP tests T-2.10.001 through T-2.10.008.

Sprint 2 Task 2.11 Checklist & Comments
- [x] 1. Verify `stats.entity_counters_shard` exists and is pre-seeded with the canonical entity catalog (Step 2.1).
- [x] 2. Create migration file `20260310201000_create_entity_counter_snapshot.php`.
- [x] 3. Verify function `stats.fn_snapshot_entity_counters()` exists.
- [x] 4. Verify the snapshot recreates exactly 25 entities × 16 shards (`400` rows).
- [x] 5. Test totals for moves, GeoKrety, pictures, users, and loves against the source tables.
- [x] 6. Test idempotence when the snapshot function is re-run.
- [x] 7. Run pgTAP tests T-2.11.001 through T-2.11.015.

Sprint 2 Task 2.12 Checklist & Comments
- [x] 1. Verify `stats.daily_activity` and `stats.daily_active_users` exist (Steps 2.2, 2.3).
- [x] 2. Create migration file `20260310201100_create_daily_activity_seed.php`.
- [x] 3. Verify function `stats.fn_seed_daily_activity(tstzrange)` exists.
- [x] 4. Test full-history seeding of `daily_activity` and `daily_active_users`.
- [x] 5. Test re-running the full seed is idempotent.
- [x] 6. Test ranged seeding refreshes only the requested days and preserves unrelated `daily_activity` metrics.
- [x] 7. Test non-day-aligned ranges are rejected with SQLSTATE `22023`.
- [x] 8. Run pgTAP tests T-2.12.001 through T-2.12.020.

Sprint 3 Status
- Goal: Implement Sprint 3 country geography foundations and exact online maintenance for country rollups and country history.
- Sprint 3 steps 3.1 through 3.8 have been implemented, applied in the development database, schema-synced into `tests`, and validated with targeted plus full website pgTAP runs.

Sprint 3 Accomplished Steps
- 20260314-1135 Created migration: `20260310300000_create_country_daily_stats.php` (creates `stats.country_daily_stats`).
- 20260314-1135 Created migration: `20260310300100_create_gk_countries_visited.php` (creates `stats.gk_countries_visited`).
- 20260314-1135 Created migration: `20260310300200_create_user_countries.php` (creates `stats.user_countries`).
- 20260314-1135 Created migration: `20260310300300_create_gk_country_history.php` (creates `stats.gk_country_history` with the `gk_country_history_excl` exclusion constraint).
- 20260314-1135 Created migration: `20260310300400_create_country_rollups_trigger.php` (creates exact recompute helpers for `country_daily_stats`, `gk_countries_visited`, and `user_countries`, then attaches `tr_gk_moves_after_country_rollups`).
- 20260314-1135 Created migration: `20260310300500_create_country_history_trigger.php` (creates `geokrety.fn_refresh_gk_country_history()` plus `tr_gk_moves_after_country_history`).
- 20260314-1135 Created migration: `20260310300600_create_country_snapshot_functions.php` (creates exact country snapshot functions and normalizes country codes to lowercase in snapshot output).
- 20260314-1135 Created migration: `20260310300700_create_country_indexes.php` (creates the three Sprint 3 country indexes).
- 20260314-1135 Added pgTAP tests: `test-215-country-daily-stats.sql` through `test-222-country-indexes.sql`.
- 20260314-1135 Updated shared schema tests: `test-10-schema.sql`, `test-10-stats.sql`.
- 20260314-1135 Updated `website/db/tests-copy-schema-geokrety-to-tests.sh` to recreate `btree_gist` in the `tests` database before importing the Sprint 3 exclusion-constraint table.
- 20260314-1135 Applied Sprint 3 migrations to the development database via the Phinx wrapper, then revalidated the corrected trigger segment by rolling back to `20260310300300` and reapplying forward after fixing runtime bigint helper signatures.
- 20260314-1135 Note: the current Phinx wrapper stack ignored `migrate --count=1` during the initial Sprint 3 apply and migrated the full pending Sprint 3 stack at once; subsequent surgical validation used `rollback -t 20260310300300` plus forward re-apply to keep the live database aligned with the edited migration files.
- 20260314-1135 Ran targeted pgTAP (`test-10-enable-btree_gist.sql`, `test-10-schema.sql`, `test-10-stats.sql`, `test-215` through `test-222`, `PASS`) after syncing schema into the `tests` database.
- 20260314-1135 Full website pgTAP passed after syncing schema into the `tests` database (`98` files, `1763` tests, `PASS`).

Sprint 3 Task 3.1 Checklist & Comments
- [x] 1. Create migration file `20260310300000_create_country_daily_stats.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify table exists with 18 columns and correct types.
- [x] 4. Verify composite PK on `(stats_date, country_code)`.
- [x] 5. Test insert and read-back.
- [x] 6. Run pgTAP tests T-3.1.001 through T-3.1.010.

Sprint 3 Task 3.2 Checklist & Comments
- [x] 1. Create migration file `20260310300100_create_gk_countries_visited.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify table exists with 5 columns and correct types.
- [x] 4. Verify composite PK on `(geokrety_id, country_code)`.
- [x] 5. Test insert and duplicate rejection.
- [x] 6. Run pgTAP tests T-3.2.001 through T-3.2.009.

Sprint 3 Task 3.3 Checklist & Comments
- [x] 1. Create migration file `20260310300200_create_user_countries.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify table exists with 5 columns and correct types.
- [x] 4. Verify composite PK on `(user_id, country_code)`.
- [x] 5. Test insert and duplicate rejection.
- [x] 6. Run pgTAP tests T-3.3.001 through T-3.3.009.

Sprint 3 Task 3.4 Checklist & Comments
- [x] 1. Verify `btree_gist` extension is enabled (Sprint 1, Step 1.7).
- [x] 2. Create migration file `20260310300300_create_gk_country_history.php`.
- [x] 3. Run `phinx migrate`.
- [x] 4. Verify table exists with 6 columns and correct types.
- [x] 5. Verify exclusion constraint `gk_country_history_excl` exists.
- [x] 6. Test non-overlapping intervals succeed.
- [x] 7. Test overlapping intervals are rejected.
- [x] 8. Test different GKs can have overlapping intervals.
- [x] 9. Run pgTAP tests T-3.4.001 through T-3.4.013.

Sprint 3 Task 3.5 Checklist & Comments
- [x] 1. Create migration file `20260310300400_create_country_rollups_trigger.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify function `geokrety.fn_gk_moves_country_rollups` exists.
- [x] 4. Verify trigger `tr_gk_moves_after_country_rollups` exists on `gk_moves`.
- [x] 5. Test INSERT with country updates all three exact-state tables.
- [x] 6. Test INSERT with NULL country creates no stats rows.
- [x] 7. Test DELETE recomputes all touched exact-state keys.
- [x] 8. Test UPDATE of country and same-country updates repair aggregates and first/last metadata.
- [x] 9. Test anonymous move skips `user_countries` while still updating the other tables.
- [x] 10. Run pgTAP tests T-3.5.001 through T-3.5.013.

Sprint 3 Task 3.6 Checklist & Comments
- [x] 1. Create migration file `20260310300500_create_country_history_trigger.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify function `geokrety.fn_gk_moves_country_history` exists.
- [x] 4. Verify trigger `tr_gk_moves_after_country_history` exists on `gk_moves`.
- [x] 5. Test first move opens an interval.
- [x] 6. Test same-country move is a no-op.
- [x] 7. Test different-country move closes the old interval and opens the new one.
- [x] 8. Test COMMENT and ARCHIVE moves do not affect intervals.
- [x] 9. Test NULL-country moves do not affect intervals.
- [x] 10. Test DELETE and UPDATE repairs rebuild exact neighboring intervals.
- [x] 11. Test exclusion constraint still rejects overlaps.
- [x] 12. Run pgTAP tests T-3.6.001 through T-3.6.015.

Sprint 3 Task 3.7 Checklist & Comments
- [x] 1. Create migration file `20260310300600_create_country_snapshot_functions.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify all three functions exist with correct signatures.
- [x] 4. Test full snapshot behavior and stale-row repair.
- [x] 5. Test idempotent reruns.
- [x] 6. Test `p_period` filtering for `country_daily_stats` and exact full-source rebuild behavior for the all-time user/GK snapshot tables.
- [x] 7. Run pgTAP tests T-3.7.001 through T-3.7.011.

Sprint 3 Task 3.8 Checklist & Comments
- [x] 1. Create migration file `20260310300700_create_country_indexes.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify all 3 indexes exist and are valid.
- [x] 4. Verify the partial index predicate on `idx_gk_country_history_active_by_country`.
- [x] 5. Run pgTAP tests T-3.8.001 through T-3.8.006.

Sprint 4 Status
- Goal: Implement Sprint 4 waypoint, cache-visit, and social-relation foundations through step 4.10.
- Sprint 4 steps 4.1 through 4.10 have been implemented, applied in the development database, rollback-verified as a stack, schema-synced into `tests`, and validated with targeted plus full website pgTAP runs.

Sprint 4 Accomplished Steps
- 20260314-1330 Created migrations: `20260310400000_create_waypoints.php` through `20260310400900_create_cache_relation_indexes.php` for Sprint 4 tasks 4.1-4.10.
- 20260314-1330 Added pgTAP tests: `test-223-waypoints.sql` through `test-232-cache-relation-indexes.sql`.
- 20260314-1330 Updated shared schema smoke test: `test-10-stats.sql`.
- 20260314-1330 Updated `website/db/tests-copy-schema-geokrety-to-tests.sh` to copy seeded `stats.waypoints` data into the `tests` database.
- 20260314-1330 Ran the review loop (`dba`, `critical-thinking`, `quality-engineer`) before applying; logged the remaining S4T03 rollback-contract tradeoff in `99-OPEN-QUESTIONS.md` as Q-030.
- 20260314-1330 Applied Sprint 4 migrations to the development database. Note: the current Phinx wrapper again ignored `migrate --count=1` and applied the full pending Sprint 4 stack in one run.
- 20260314-1330 Verified rollback by reverting from `20260314103000` back to `20260310300700`, which exercised Sprint 4 `down()` methods in reverse order. Then re-applied forward to `20260314103000` to restore both Sprint 4 and the preexisting `20260314103000_fix_previous_move_batch_update_rewire.php` migration.
- 20260314-1330 Synced the live schema into `tests` with the updated copy script and confirmed seeded `stats.waypoints` data was available in the test database.
- 20260314-1330 Ran targeted pgTAP (`test-10-stats.sql`, `test-223` through `test-232`, `PASS`) after two test-fixture fixes for seeded waypoint IDs and deferred-FK metadata checks.
- 20260314-1330 Full website pgTAP passed after syncing schema into the `tests` database (`108` files, `1896` tests, `PASS`).

Sprint 4 Task 4.1 Checklist & Comments
- [x] 1. Create migration file `20260310400000_create_waypoints.php`.
- [x] 2. Verify PHP class name is `CreateWaypoints`.
- [x] 3. Run `phinx migrate`.
- [x] 4. Verify `stats.waypoints` exists with the expected 7 columns and constraints.
- [x] 5. Verify `uq_waypoints_code` and `chk_waypoints_source` exist.
- [x] 6. Verify nullable coordinate behavior and duplicate/source constraint rejection.
- [x] 7. Run pgTAP tests T-4.1.001 through T-4.1.013.
- [x] 8. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 4 Task 4.2 Checklist & Comments
- [x] 1. Create migration file `20260310400100_create_waypoints_source_view.php`.
- [x] 2. Verify source waypoint tables exist and expose the expected columns.
- [x] 3. Run `phinx migrate`.
- [x] 4. Verify the view exposes both GC and OC sources.
- [x] 5. Verify uppercase normalization and blank/NULL waypoint filtering.
- [x] 6. Run pgTAP tests T-4.2.001 through T-4.2.012.
- [x] 7. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 4 Task 4.3 Checklist & Comments
- [x] 1. Verify S4T01 and S4T02 dependencies are applied before seeding.
- [x] 2. Create migration file `20260310400200_seed_waypoints.php`.
- [x] 3. Run `phinx migrate`.
- [x] 4. Verify `stats.waypoints` is populated from the union view and includes both GC and OC rows.
- [x] 5. Verify `stats.fn_seed_waypoints()` is idempotent and writes to `stats.job_log`.
- [x] 6. Update the tests DB copy workflow to include seeded `stats.waypoints` data.
- [x] 7. Run pgTAP tests T-4.3.001 through T-4.3.010.
- [x] 8. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 4 Task 4.4 Checklist & Comments
- [x] 1. Create migration file `20260310400300_create_gk_cache_visits.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify composite PK `(gk_id, waypoint_id)` and deferred FK `fk_gk_cache_visits_waypoint`.
- [x] 4. Verify the table remains empty after creation and rejects invalid waypoint references when deferred constraints are checked.
- [x] 5. Run pgTAP tests T-4.4.001 through T-4.4.014.
- [x] 6. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 4 Task 4.5 Checklist & Comments
- [x] 1. Create migration file `20260310400400_create_user_cache_visits.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify composite PK `(user_id, waypoint_id)` and deferred FK `fk_user_cache_visits_waypoint`.
- [x] 4. Verify the table remains empty after creation and rejects invalid waypoint references when deferred constraints are checked.
- [x] 5. Run pgTAP tests T-4.5.001 through T-4.5.014.
- [x] 6. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 4 Task 4.6 Checklist & Comments
- [x] 1. Create migration file `20260310400500_create_gk_related_users.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify composite PK `(geokrety_id, user_id)` and the expected timestamptz columns.
- [x] 4. Verify valid inserts and required-timestamp enforcement.
- [x] 5. Run pgTAP tests T-4.6.001 through T-4.6.012.
- [x] 6. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 4 Task 4.7 Checklist & Comments
- [x] 1. Create migration file `20260310400600_create_user_related_users.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify composite PK `(user_id, related_user_id)` and check constraint `chk_user_related_users_no_self`.
- [x] 4. Verify required-timestamp enforcement and self-link rejection.
- [x] 5. Run pgTAP tests T-4.7.001 through T-4.7.013.
- [x] 6. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 4 Task 4.8 Checklist & Comments
- [x] 1. Create migration file `20260310400700_create_waypoint_cache_trigger.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify function `geokrety.fn_gk_moves_waypoint_cache()` and trigger `tr_gk_moves_after_waypoint_visits` exist.
- [x] 4. Verify uppercase waypoint registration, anonymous-user skipping, comment skipping, NULL-waypoint skipping, and exact UPDATE/DELETE reconciliation.
- [x] 5. Run pgTAP tests T-4.8.001 through T-4.8.015.
- [x] 6. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 4 Task 4.9 Checklist & Comments
- [x] 1. Create migration file `20260310400800_create_relation_trigger.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify function `geokrety.fn_gk_moves_relations()` and trigger `tr_gk_moves_after_relations` exist.
- [x] 4. Verify qualifying move filtering, anonymous/comment skipping, symmetric user-pair maintenance, and exact UPDATE/DELETE reconciliation.
- [x] 5. Run pgTAP tests T-4.9.001 through T-4.9.016.
- [x] 6. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 4 Task 4.10 Checklist & Comments
- [x] 1. Create migration file `20260310400900_create_cache_relation_indexes.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify the canonical four Sprint 4 secondary indexes exist and are valid.
- [x] 4. Verify the `idx_waypoints_country` partial-index predicate and the absence of a duplicate waypoint-code lookup index.
- [x] 5. Run pgTAP tests T-4.10.001 through T-4.10.008.
- [x] 6. Verify rollback as part of the Sprint 4 stack rollback cycle.

Sprint 5 Status
- Goal: Implement Sprint 5 analytical aggregates, live milestone/first-finder automation, and the points-awarder event bridge through step 5.10.
- Sprint 5 steps 5.1 through 5.10 have been implemented, applied in the development database, rollback-verified across the edited trigger/function stack, schema-synced into `tests`, and validated with targeted plus full website pgTAP runs.

Sprint 5 Accomplished Steps
- 20260314-1400 Created migrations: `20260310500000_create_hourly_activity.php` through `20260310500900_create_analytics_indexes.php` for Sprint 5 tasks 5.1-5.10.
- 20260314-1400 Added pgTAP tests: `test-233-hourly-activity.sql` through `test-242-analytics-indexes.sql`.
- 20260314-1400 Updated shared schema smoke tests: `test-10-schema.sql` and `test-10-stats.sql`.
- 20260314-1400 Ran the review loop (`dba`, `critical-thinking`, `quality-engineer`) before applying, then iterated on the implementation to address four issues the reviews exposed: historical loves-country attribution, missing seed behavior for preexisting loves, points-bridge coexistence with the legacy AMQP trigger, and invalid/idempotence gaps in the snapshot-function tests.
- 20260314-1400 Adjusted `20260310500400_create_gk_loves_counter_trigger.php` so loves rollups resolve the GeoKret country at love-time via `stats.gk_country_history`, with fallback to the current country, and so migration apply seeds `stats.entity_counters_shard`, `stats.daily_activity.loves_count`, and `stats.country_daily_stats.loves_count` from existing `geokrety.gk_loves` data.
- 20260314-1400 Adjusted `20260310500600_create_milestone_trigger.php` after first execution to cast `NEW.geokret` to `INT` in calls to `geokrety.fn_record_gk_milestone_event(...)`, matching the helper signature installed by the migration.
- 20260314-1400 Adjusted `20260310500800_create_batch_aggregation_functions.php` after targeted pgTAP to defensively drop its temp snapshot tables before recreating them, making repeated execution in a single session idempotent.
- 20260314-1400 Applied Sprint 5 migrations to the development database, then rollback-verified and re-applied the edited subset by rewinding to `20260310500500`, replaying `20260310500600`, and later rewinding to `20260310500700` to reinstall the corrected `20260310500800` function definitions before restoring `20260310500900`.
- 20260314-1400 Synced the live schema into `tests` after each replay that changed installed database functions or triggers.
- 20260314-1400 Ran targeted pgTAP (`test-10-schema.sql`, `test-10-stats.sql`, `test-233` through `test-242`, `PASS`) after fixing the final assertion and same-session snapshot idempotence issues.
- 20260314-1400 Full website pgTAP passed after syncing schema into the `tests` database (`118` files, `2035` tests, `PASS`).

Sprint 5 Task 5.1 Checklist & Comments
- [x] 1. Create migration file `20260310500000_create_hourly_activity.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify `stats.hourly_activity` exists with the expected composite PK and UTC hour / move type constraints.
- [x] 4. Verify valid inserts succeed and invalid hour / move-type values are rejected.
- [x] 5. Run pgTAP tests T-5.1.001 through T-5.1.012.

Sprint 5 Task 5.2 Checklist & Comments
- [x] 1. Create migration file `20260310500100_create_country_pair_flows.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify `stats.country_pair_flows` exists with the expected composite PK and uppercase country constraints.
- [x] 4. Verify the month bucket is constrained to the first day of the month and self-loops are rejected.
- [x] 5. Run pgTAP tests T-5.2.001 through T-5.2.016.

Sprint 5 Task 5.3 Checklist & Comments
- [x] 1. Create migration file `20260310500200_create_gk_milestone_events.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify `stats.gk_milestone_events` exists with the expected unique `(gk_id, event_type)` contract and supporting indexes.
- [x] 4. Verify allowed `event_type` enforcement and JSONB payload persistence.
- [x] 5. Run pgTAP tests T-5.3.001 through T-5.3.014.

Sprint 5 Task 5.4 Checklist & Comments
- [x] 1. Create migration file `20260310500300_create_first_finder_events.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify `stats.first_finder_events` exists with the expected one-row-per-GK PK, move metadata columns, and partial 168-hour index.
- [x] 4. Verify invalid move types, negative hours, and duplicate GK rows are rejected.
- [x] 5. Run pgTAP tests T-5.4.001 through T-5.4.015.

Sprint 5 Task 5.5 Checklist & Comments
- [x] 1. Create migration file `20260310500400_create_gk_loves_counter_trigger.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify `geokrety.fn_gk_loves_activity` and `tr_gk_loves_activity` exist on `geokrety.gk_loves`.
- [x] 4. Verify insert/update/delete paths refresh shard counters, `stats.daily_activity.loves_count`, and `stats.country_daily_stats.loves_count` exactly.
- [x] 5. Verify historical updates/deletes repair the original love-time country bucket even when the GeoKret later moves.
- [x] 6. Verify existing loves are seeded into the shard and daily rollup tables during migration apply.
- [x] 7. Run pgTAP tests T-5.5.001 through T-5.5.016.

Sprint 5 Task 5.6 Checklist & Comments
- [x] 1. Create migration file `20260310500500_create_amqp_event_trigger.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify `geokrety.fn_gk_moves_emit_points_event` and `tr_gk_moves_emit_points_event` exist on `geokrety.gk_moves`.
- [x] 4. Verify qualifying move inserts append one `notify_queues.geokrety_changes` row with `channel = 'points-awarder'`, `action = 'gk_move_created'`, and payload = move id.
- [x] 5. Verify the legacy `after_99_notify_amqp_moves` trigger still coexists with the new bridge path.
- [x] 6. Run pgTAP tests T-5.6.001 through T-5.6.011.

Sprint 5 Task 5.7 Checklist & Comments
- [x] 1. Create migration file `20260310500600_create_milestone_trigger.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify `geokrety.fn_gk_moves_milestones` and `tr_gk_moves_after_milestones` exist on `geokrety.gk_moves`.
- [x] 4. Verify milestone insertion for `km_100`, `km_1000`, and user-count thresholds, including `additional_data` payloads.
- [x] 5. Verify non-qualifying moves do not create milestone rows.
- [x] 6. Run pgTAP tests T-5.7.001 through T-5.7.014.

Sprint 5 Task 5.8 Checklist & Comments
- [x] 1. Create migration file `20260310500700_create_first_finder_trigger.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify `stats.fn_detect_first_finder`, `geokrety.fn_gk_moves_first_finder`, and `tr_gk_moves_after_first_finder` exist.
- [x] 4. Verify the first qualifying non-owner authenticated move records `stats.first_finder_events` and appends the `first_find` milestone.
- [x] 5. Verify owner, anonymous, late, and repeated qualifying moves do not create duplicate first-finder rows.
- [x] 6. Run pgTAP tests T-5.8.001 through T-5.8.011.

Sprint 5 Task 5.9 Checklist & Comments
- [x] 1. Create migration file `20260310500800_create_batch_aggregation_functions.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify `stats.fn_snapshot_hourly_activity()` and `stats.fn_snapshot_country_pair_flows()` exist.
- [x] 4. Verify both snapshot functions upsert exact aggregates, remove stale rows, and append canonical `stats.job_log` entries.
- [x] 5. Verify both functions are idempotent across repeated executions in the same SQL session.
- [x] 6. Run pgTAP tests T-5.9.001 through T-5.9.012.

Sprint 5 Task 5.10 Checklist & Comments
- [x] 1. Create migration file `20260310500900_create_analytics_indexes.php`.
- [x] 2. Run `phinx migrate`.
- [x] 3. Verify the canonical four Sprint 5 analytics indexes exist and are valid.
- [x] 4. Verify representative ordered/filter queries use the intended indexes.
- [x] 5. Run pgTAP tests T-5.10.001 through T-5.10.010.

Sprint 4 Addendum: Task 4.11
- Goal: Add the canonical waypoint/cache/relation snapshot helpers required by Sprint 6 orchestration.
- Sprint 4 task 4.11 has been implemented, applied in the development database, rollback-verified, schema-synced into `tests`, and validated with targeted plus full website pgTAP runs.

Sprint 4 Task 4.11 Checklist & Comments
- [x] 1. Create migration file `20260310401000_create_waypoint_relation_snapshots.php`.
- [x] 2. Verify `stats.fn_snapshot_waypoints()`, `stats.fn_snapshot_cache_visits()`, `stats.fn_snapshot_relations()`, and `stats.fn_snapshot_relationship_tables(daterange)` exist.
- [x] 3. Verify helper reruns are idempotent and refresh earliest-waypoint facts correctly for UK-sourced waypoints.
- [x] 4. Verify helper execution writes canonical `stats.job_log` rows.
- [x] 5. Verify rollback removes the helper set cleanly.
- [x] 6. Run pgTAP tests T-4.11.001 through T-4.11.011.

Sprint 6 Status
- Goal: Implement Sprint 6 backfills, orchestration, canonical stats views/materialized views, reconciliation, and the database-side quality-gate follow-up.
- Sprint 6 tasks 6.1 through 6.9 have been implemented in the development database, rollback-verified across the edited migration tail, schema-synced into `tests`, and validated with targeted plus full website pgTAP runs.

Sprint 6 Accomplished Steps
- 20260314-0000 Created migrations: `20260310600000_create_backfill_previous_move.php` through `20260310600600_create_materialized_views.php` for Sprint 6 tasks 6.1-6.7, plus the standardized reconciliation helper in task 6.8.
- 20260314-0000 Added pgTAP tests: `test-244-backfill-previous-move.sql` through `test-250-materialized-views.sql`.
- 20260314-0000 Ran the review loop before applying, then iterated on the implementation to fix rerun semantics in waypoint snapshots, heavy-wrapper bounds, reconciliation catalog handling, materialized-view refreshability, and pgTAP fixture accuracy.
- 20260314-0000 Verified rollback/re-apply behavior for every Sprint 6 migration and reinstalled edited migrations after each post-test correction so the live development database always matched the checked-in migration sources.
- 20260314-0000 Verified `REFRESH MATERIALIZED VIEW CONCURRENTLY` for `stats.mv_country_month_rollup`, `stats.mv_top_caches_global`, and `stats.mv_global_kpi` after adding a real singleton key to `mv_global_kpi`.
- 20260314-0000 Created follow-up migration `20260310600700_split_snapshot_runs_and_add_runtime_indexes.php` to add phase-split/scoped orchestration overloads, selective reconciliation overloads, and runtime indexes `idx_gk_moves_qualified_period` plus `idx_gk_moves_distance_records` after the monolithic quality-gate run proved impractical on the development dataset.
- 20260314-0000 Added pgTAP test `test-251-snapshot-runtime-indexes.sql` and expanded `test-248-snapshot-orchestration.sql` to validate the scoped execution path.
- 20260314-0000 Targeted pgTAP passed for the affected Sprint 5/6 files after the follow-up migration, and the full website pgTAP suite passed after syncing schema into the `tests` database (`127` files, `2120` tests, `PASS`).
- 20260315-0000 Updated `20260314102104_split_previous_move_and_position_chain.php` so `stats.mv_backfill_working_set` precomputes `expected_km_distance`, `stats.fn_backfill_km_distance()` consumes that cached value instead of recomputing it per batch row, the stale duplicate heavy-wrapper override was removed, and both heavy backfill entry points now include a temporary 3-month processing cap. Re-applied the migration, re-synced the `tests` database, and re-ran targeted pgTAP for `test-30`, `test-205`, `test-244`, and `test-246` (`85` tests, `PASS`).

Sprint 6 Quality-Gate Evidence
- `stats.fn_reconcile_stats(ARRAY['stats.daily_activity','stats.country_daily_stats','stats.hourly_activity','stats.country_pair_flows'], tstzrange('2026-03-01', '2026-04-01', '[)'))` completed in about `6.2s` on the development dataset. The earlier non-period-scoped selective run over the same logical check set took about `19.3s` before the overload was added.
- `stats.fn_run_all_snapshots(ARRAY['fn_seed_daily_activity','fn_snapshot_daily_country_stats','fn_snapshot_hourly_activity','fn_snapshot_country_pair_flows'], tstzrange('2026-03-01', '2026-04-01', '[)'), 5000)` completed in about `11.9s` for the March 2026 subset. Phase-level profiling showed approximately `3.4s` for `fn_seed_daily_activity`, `7.4s` for `fn_snapshot_daily_country_stats`, `0.45s` for `fn_snapshot_hourly_activity`, and `0.45s` for `fn_snapshot_country_pair_flows`.
- `EXPLAIN (ANALYZE, BUFFERS)` for `stats.v_uc15_distance_records` keyed lookup now uses `idx_gk_moves_distance_records` and completed in about `0.955 ms` on the development dataset.
- Practical rollout note: the legacy monolithic no-arg `stats.fn_run_all_snapshots()` / `stats.fn_reconcile_stats()` path remains too heavy for comfortable interactive use on the current development dataset. The new phase-split/scoped overloads are the operational path for Sprint 6 validation and repair runs.

Sprint 6 Task 6.1 Checklist & Comments
- [x] 1. Create migration file `20260310600000_create_backfill_previous_move.php`.
- [x] 2. Verify `stats.fn_backfill_previous_move_id(tstzrange, int)` exists and updates only mismatched predecessor links.
- [x] 3. Verify idempotence and canonical `stats.job_log` metadata.
- [x] 4. Run pgTAP tests T-6.1.001 through T-6.1.010.

Sprint 6 Task 6.2 Checklist & Comments
- [x] 1. Create migration file `20260310600100_create_backfill_previous_move_heavy.php`.
- [x] 2. Verify the heavy wrapper iterates year slices through the latest mismatched year and logs a canonical summary even on early exit.
- [x] 3. Run pgTAP tests T-6.2.001 through T-6.2.008.

Sprint 6 Task 6.3 Checklist & Comments
- [x] 1. Create migration file `20260310600200_create_backfill_km_distance.php`.
- [x] 2. Verify `stats.fn_backfill_km_distance(tstzrange, int)` recomputes canonical row-level km distance exactly.
- [x] 3. Verify idempotence and canonical `stats.job_log` metadata.
- [x] 4. Run pgTAP tests T-6.3.001 through T-6.3.009.

Sprint 6 Task 6.4 Checklist & Comments
- [x] 1. Create migration file `20260310600300_create_backfill_km_distance_heavy.php`.
- [x] 2. Verify the heavy wrapper iterates year slices through the latest mismatched year and exits cleanly when no work remains.
- [x] 3. Run pgTAP tests T-6.4.001 through T-6.4.008.

Sprint 6 Task 6.5 Checklist & Comments
- [x] 1. Create migration file `20260310600400_create_snapshot_orchestration.php`.
- [x] 2. Verify `stats.fn_run_all_snapshots()` and `stats.fn_reconcile_stats()` exist and execute successfully on the canonical fixture state.
- [x] 3. Verify the orchestration metadata records the canonical phase order and the reconciliation helper enforces the exact-zero-delta policy.
- [x] 4. Run pgTAP tests T-6.5.001 through T-6.5.013.

Sprint 6 Task 6.6 Checklist & Comments
- [x] 1. Create migration file `20260310600500_create_stats_views.php`.
- [x] 2. Verify the canonical twelve UC views exist and read from the intended precomputed stats sources.
- [x] 3. Run pgTAP tests T-6.6.001 through T-6.6.010.

Sprint 6 Task 6.7 Checklist & Comments
- [x] 1. Create migration file `20260310600600_create_materialized_views.php`.
- [x] 2. Verify the canonical three materialized views exist with the required unique indexes.
- [x] 3. Verify all three materialized views support `REFRESH MATERIALIZED VIEW CONCURRENTLY`.
- [x] 4. Run pgTAP tests T-6.7.001 through T-6.7.007.

Sprint 6 Task 6.8 Checklist & Comments
- [x] 1. Standardize `stats.fn_reconcile_stats()` in the migration set.
- [x] 2. Verify canonical reconciliation returns zero mismatches on the task fixture state.
- [x] 3. Verify reconciliation writes canonical `stats.job_log` rows using the approved columns only.
- [x] 4. Run pgTAP tests T-6.8.001 through T-6.8.006.

Sprint 6 Task 6.9 Checklist & Comments
- [x] 1. Run `EXPLAIN` / latency checks on canonical Sprint 6 read paths.
- [x] 2. Validate concurrent refresh and keyed lookup performance for the new materialized-view / UC-view hot paths.
- [x] 3. Add a split/scoped execution path for orchestration and reconciliation after the monolithic path proved too slow for a single interactive run.
- [x] 4. Add runtime indexes targeted at the measured slow paths.
- [x] 5. Record the remaining operational constraint that phase-split/scoped calls are the recommended path on the current development dataset.

