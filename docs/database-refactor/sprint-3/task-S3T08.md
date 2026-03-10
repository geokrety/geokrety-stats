---
title: "Task S3T08: Create Country Indexes"
version: 1.0
date_created: 2026-03-08
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 3
tags:
  - country
  - database
  - geography
  - migration
  - postgresql
  - schema
  - sprint-3
  - stats
  - task-merge
  - traversal
depends_on: [1, 2]
task: S3T08
step: 3.8
migration: 20260310300700_create_country_indexes.php
blocks: [5, 6]
changelog:
  - 2026-03-10: created by merge of 03-sprint-3-country-geography.md step 3.8
  - 2026-03-10: removed sprint-wide appendix content from this step-scoped task file
---

# Task S3T08: Create Country Indexes

## Master-Spec Alignment

The normative contract for this sprint is [00-SPEC-DRAFT-v1.md](00-SPEC-DRAFT-v1.md), Sections 5.3, 5.4, 8.4, 9.2, and 11.

- `stats.country_daily_stats.unique_users` and `unique_gks` are exact online-maintained values, not approximate placeholders.
- `INSERT`, `UPDATE`, and `DELETE` handling for `stats.gk_countries_visited`, `stats.user_countries`, and `stats.gk_country_history` must maintain exact state. When earliest/latest rows are invalidated, affected rows must be recomputed from remaining qualifying moves.
- Snapshot functions seed and verify canonical state; they do not compensate for knowingly inexact live maintenance.
- Any lower text that still describes `unique_users` or `unique_gks` as approximate is obsolete and superseded by this alignment block.

## Source

- Generated from sprint document step `3.8` in `03-sprint-3-country-geography.md`.

### Step 3.8: Create Country Indexes

**What this step does:** Creates three supporting indexes on the country stats tables to optimize common query patterns: country-first time series, active GKs by country, and GK travel timelines. These indexes support the frontend API queries for choropleth maps, country detail pages, and GK timeline views.

**Migration file name:** `20260310300700_create_country_indexes.php`

#### Full SQL DDL

```sql
-- Index 1: Country-first lookups for single-country time series.
-- Used by: Country detail page — daily trend for a specific country.
-- Query pattern: WHERE country_code = $cc ORDER BY stats_date
-- The PK is (stats_date, country_code) which is date-first;
-- this index reverses the order for country-first access.
CREATE INDEX idx_country_daily_stats_country_date
  ON stats.country_daily_stats (country_code, stats_date);

-- Index 2: Fast lookup of GKs currently in a specific country.
-- Used by: "GKs in [country]" map overlay, country population count.
-- Query pattern: WHERE country_code = $cc AND departed_at IS NULL
-- Partial index excludes closed intervals, keeping the index small.
CREATE INDEX idx_gk_country_history_active_by_country
  ON stats.gk_country_history (country_code)
  WHERE departed_at IS NULL;

-- Index 3: GK travel timeline queries.
-- Used by: GK story page — chronological list of countries visited.
-- Query pattern: WHERE geokrety_id = $gk ORDER BY arrived_at DESC
CREATE INDEX idx_gk_country_history_gk_arrived
  ON stats.gk_country_history (geokrety_id, arrived_at DESC);
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateCountryIndexes extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE INDEX IF NOT EXISTS idx_country_daily_stats_country_date
  ON stats.country_daily_stats (country_code, stats_date);
SQL
        );

        $this->execute(<<<'SQL'
CREATE INDEX IF NOT EXISTS idx_gk_country_history_active_by_country
  ON stats.gk_country_history (country_code)
  WHERE departed_at IS NULL;
SQL
        );

        $this->execute(<<<'SQL'
CREATE INDEX IF NOT EXISTS idx_gk_country_history_gk_arrived
  ON stats.gk_country_history (geokrety_id, arrived_at DESC);
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP INDEX IF EXISTS stats.idx_country_daily_stats_country_date;');
        $this->execute('DROP INDEX IF EXISTS stats.idx_gk_country_history_active_by_country;');
        $this->execute('DROP INDEX IF EXISTS stats.idx_gk_country_history_gk_arrived;');
    }
}
```

#### SQL Usage Examples

```sql
-- Verify all 3 indexes exist
SELECT indexname, indexdef
FROM pg_indexes
WHERE schemaname = 'stats'
  AND indexname IN (
    'idx_country_daily_stats_country_date',
    'idx_gk_country_history_active_by_country',
    'idx_gk_country_history_gk_arrived'
  )
ORDER BY indexname;

-- Verify partial index predicate on active_by_country
SELECT indexdef FROM pg_indexes
WHERE indexname = 'idx_gk_country_history_active_by_country';
-- Expected: ... WHERE (departed_at IS NULL)

-- Test country-first index is used for time series
EXPLAIN (COSTS OFF)
SELECT stats_date, moves_count
FROM stats.country_daily_stats
WHERE country_code = 'PL'
ORDER BY stats_date;

-- Test active-by-country partial index
EXPLAIN (COSTS OFF)
SELECT geokrety_id, arrived_at
FROM stats.gk_country_history
WHERE country_code = 'PL' AND departed_at IS NULL;

-- Test GK timeline index
EXPLAIN (COSTS OFF)
SELECT country_code, arrived_at, departed_at
FROM stats.gk_country_history
WHERE geokrety_id = 1
ORDER BY arrived_at DESC;
```

#### Graph/Visualization Specification

No new visualization unlocked at this step. These indexes optimize query performance for the visualizations defined in Steps 3.1–3.4.

#### TimescaleDB Assessment

**NOT recommended.** These are standard B-tree indexes on stats tables. Standard PostgreSQL indexes are appropriate for the expected data volumes (~1.3M rows for `country_daily_stats`, ~2M rows for `gk_country_history`).

#### pgTAP Unit Tests

| Test ID   | Test Name                                        | Assertion                                                                                                             |
| --------- | ------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------- |
| T-3.8.001 | idx_country_daily_stats_country_date exists      | `has_index('stats', 'country_daily_stats', 'idx_country_daily_stats_country_date')`                                   |
| T-3.8.002 | idx_gk_country_history_active_by_country exists  | `has_index('stats', 'gk_country_history', 'idx_gk_country_history_active_by_country')`                                |
| T-3.8.003 | idx_gk_country_history_gk_arrived exists         | `has_index('stats', 'gk_country_history', 'idx_gk_country_history_gk_arrived')`                                       |
| T-3.8.004 | active_by_country is a partial index             | `SELECT indexdef LIKE '%WHERE%' FROM pg_indexes WHERE indexname = 'idx_gk_country_history_active_by_country'` is TRUE |
| T-3.8.005 | active_by_country filters by departed_at IS NULL | Parse indexdef and verify predicate                                                                                   |
| T-3.8.006 | All 3 indexes are valid                          | `SELECT bool_and(indisvalid) FROM pg_index WHERE indexrelid IN (SELECT oid FROM pg_class WHERE relname IN (...))`     |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310300700_create_country_indexes.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify all 3 indexes exist and are valid
- [ ] 4. Verify partial index predicate on `idx_gk_country_history_active_by_country`
- [ ] 5. Test EXPLAIN on sample queries to confirm index usage
- [ ] 6. Run pgTAP tests T-3.8.001 through T-3.8.006

### Edge Case 7: Concurrent moves for same GK

If two moves for the same GK arrive concurrently, the exclusion constraint acts as a serialization point — the second INSERT will block until the first transaction commits or rolls back. This prevents race conditions in interval management.

```sql
-- Edge Case 3: GK revisits same country
-- After GK #1 moves: PL → DE → PL
SELECT country_code, arrived_at, departed_at
FROM stats.gk_country_history
WHERE geokrety_id = 1
ORDER BY arrived_at;
-- Result:
-- PL | 2025-01-01 | 2025-03-01  (first PL stay)
-- DE | 2025-03-01 | 2025-06-01  (DE visit)
-- PL | 2025-06-01 | NULL        (returned to PL, still there)
```

## 11. Validation Criteria

1. `phinx status` shows all 8 Sprint 3 migrations as `up`
2. `stats.country_daily_stats`, `stats.gk_countries_visited`, `stats.user_countries`, `stats.gk_country_history` all exist with correct columns and types
3. Exclusion constraint `gk_country_history_excl` exists and rejects overlapping intervals
4. Triggers `tr_gk_moves_after_country_rollups` and `tr_gk_moves_after_country_history` exist on `gk_moves`
5. Functions `fn_gk_moves_country_rollups()` and `fn_gk_moves_country_history()` exist in `geokrety` schema
6. Functions `fn_snapshot_daily_country_stats`, `fn_snapshot_user_country_stats`, `fn_snapshot_gk_country_stats` exist in `stats` schema
7. All 3 indexes exist and are valid
8. INSERT/UPDATE/DELETE on `gk_moves` correctly maintains all 4 stats tables
9. Anonymous moves do not create `user_countries` rows
10. NULL-country moves do not affect any stats tables
11. All 86 pgTAP tests pass

## 12. Related Specifications / Further Reading

- [00-SCHEMA.md](00-SCHEMA.md) — Comprehensive schema design document (sections 5.3, 5.4, 8.4, 9.2)
- [01-SPRINT-INDEX.md](01-SPRINT-INDEX.md) — Sprint dependency graph and summary
- [spec-schema-sprint-1-foundation.md](../../spec/spec-schema-sprint-1-foundation.md) — Sprint 1: Foundation
- [gamification-rules.instructions.md](../../.github/instructions/gamification-rules.instructions.md) — Country crossing rules (+0.05 multiplier, +3 points)
- [PostgreSQL EXCLUDE constraint](https://www.postgresql.org/docs/16/sql-createtable.html#SQL-CREATETABLE-EXCLUDE)
- [btree_gist Extension](https://www.postgresql.org/docs/16/btree-gist.html)
- [pgTAP Documentation](https://pgtap.org/documentation.html)
- [Phinx Migration Documentation](https://book.cakephp.org/phinx/0/en/migrations.html)

---

## Appendix A: Migration Execution Order

| Order | Migration ID     | File Name                                              | Step |
| ----- | ---------------- | ------------------------------------------------------ | ---- |
| 1     | `20260310300000` | `20260310300000_create_country_daily_stats.php`        | 3.1  |
| 2     | `20260310300100` | `20260310300100_create_gk_countries_visited.php`       | 3.2  |
| 3     | `20260310300200` | `20260310300200_create_user_countries.php`             | 3.3  |
| 4     | `20260310300300` | `20260310300300_create_gk_country_history.php`         | 3.4  |
| 5     | `20260310300400` | `20260310300400_create_country_rollups_trigger.php`    | 3.5  |
| 6     | `20260310300500` | `20260310300500_create_country_history_trigger.php`    | 3.6  |
| 7     | `20260310300600` | `20260310300600_create_country_snapshot_functions.php` | 3.7  |
| 8     | `20260310300700` | `20260310300700_create_country_indexes.php`            | 3.8  |

## Appendix B: Objects Created Summary

| Object Type | Schema              | Name                                         | Step |
| ----------- | ------------------- | -------------------------------------------- | ---- |
| Table       | `stats`             | `country_daily_stats`                        | 3.1  |
| Table       | `stats`             | `gk_countries_visited`                       | 3.2  |
| Table       | `stats`             | `user_countries`                             | 3.3  |
| Table       | `stats`             | `gk_country_history`                         | 3.4  |
| Constraint  | `stats`             | `gk_country_history_excl` (exclusion)        | 3.4  |
| Function    | `geokrety`          | `fn_gk_moves_country_rollups()`              | 3.5  |
| Trigger     | `geokrety.gk_moves` | `tr_gk_moves_after_country_rollups`          | 3.5  |
| Function    | `geokrety`          | `fn_gk_moves_country_history()`              | 3.6  |
| Trigger     | `geokrety.gk_moves` | `tr_gk_moves_after_country_history`          | 3.6  |
| Function    | `stats`             | `fn_snapshot_daily_country_stats(daterange)` | 3.7  |
| Function    | `stats`             | `fn_snapshot_user_country_stats(daterange)`  | 3.7  |
| Function    | `stats`             | `fn_snapshot_gk_country_stats(daterange)`    | 3.7  |
| Index       | `stats`             | `idx_country_daily_stats_country_date`       | 3.8  |
| Index       | `stats`             | `idx_gk_country_history_active_by_country`   | 3.8  |
| Index       | `stats`             | `idx_gk_country_history_gk_arrived`          | 3.8  |

## Appendix C: Objects Removed Summary

No objects are removed in Sprint 3. All objects are new additions.
