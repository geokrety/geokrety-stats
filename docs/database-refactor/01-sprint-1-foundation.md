---
title: 'Sprint 1: Foundation & Source Table Preparation'
version: 1.0
date_created: 2026-03-08
last_updated: 2026-03-08
owner: GeoKrety Community
sprint: 1
depends_on: []
blocks: [2, 3, 4, 5, 6]
tags:
  - database
  - postgresql
  - stats
  - schema
  - migration
  - sprint-1
  - foundation
  - source-table
  - revert
---

# Sprint 1: Foundation & Source Table Preparation

This sprint establishes the foundation for the entire Stats database refactoring. It reverts five preliminary stats migrations that are being superseded by the new coherent schema design, creates the clean `stats` schema, installs operational support tables for resumable backfill jobs, seeds a continent reference table, adds two computed columns (`previous_move_id` and `km_distance`) to the source `geokrety.gk_moves` table, creates five optimized source table indexes, and enables the `btree_gist` extension required by the exclusion constraint in Sprint 3. Every subsequent sprint depends on this foundation.

## Master-Spec Alignment

The normative contract for this sprint is [00-SPEC-DRAFT-v1.obsolete.md](00-SPEC-DRAFT-v1.obsolete.md), Sections 1, 4, 5.8, 8, 9, and 12.

- Step 1.1 is a forward-only cleanup migration that drops superseded schema objects. It does not mutate Phinx bookkeeping tables.
- Any legacy `phinxlog` deletion SQL or test text later in this draft is obsolete and non-normative.
- `REQ-110` does not require Step 1.1 to recreate the reverted exploratory migrations in `down()`; rollback behavior must remain consistent with the forward-only cleanup purpose of that step.

## 1. Purpose & Scope

**Purpose:** Provide a clean slate for the stats schema by reverting superseded migrations, then lay down infrastructure (schema, operational tables, reference data, source columns, indexes, extensions) that every subsequent sprint depends on.

**Scope:**

- Revert 5 preliminary stats migrations (dropping their triggers, functions, tables, indexes, and schema)
- Create the `stats` schema
- Create `stats.backfill_progress` and `stats.job_log` operational support tables
- Create `stats.continent_reference` table and seed 249 ISO 3166-1 country-to-continent mappings
- Add `previous_move_id` (BIGINT FK) and `km_distance` (NUMERIC(8,3)) columns to `geokrety.gk_moves`
- Create 5 source table indexes on `geokrety.gk_moves` (CONCURRENTLY where possible)
- Enable `btree_gist` extension

**Intended audience:** Database engineers, backend developers, AI agents executing migration steps.

**Assumptions:**

- PostgreSQL 16.3 with PostGIS 3.4.2, pgTAP 1.3.3
- The 5 preliminary stats migrations have been previously applied and their objects exist
- The Phinx migration framework (CakePHP) is used for all migrations
- The `geokrety.gk_moves` table exists with its current schema (columns: `id`, `geokret`, `lat`, `lon`, `elevation`, `country`, `distance`, `waypoint`, `author`, `comment`, `pictures_count`, `comments_count`, `username`, `app`, `app_ver`, `created_on_datetime`, `moved_on_datetime`, `updated_on_datetime`, `move_type`, `position`)
- `geokrety.gk_waypoints_gc`, `geokrety.gk_waypoints_oc`, `geokrety.gk_geokrety`, `geokrety.gk_users`, `geokrety.gk_pictures`, `geokrety.gk_loves` tables exist

## 2. Definitions

| Term                       | Definition                                                                                                           |
| -------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| **Preliminary migrations** | Five existing stats migrations (20260228174500 through 20260307140000) being superseded by this new schema           |
| **Revert**                 | A forward-only migration that drops objects created by previous migrations, ensuring clean state                     |
| **stats schema**           | PostgreSQL schema dedicated to counters, aggregates, relationships, geography/time buckets, and helper operations    |
| **backfill_progress**      | Operational table tracking cursor position and status of resumable heavy batch jobs                                  |
| **job_log**                | Audit table recording execution metadata for all backfill/replay/snapshot operations                                 |
| **continent_reference**    | Lookup table mapping ISO 3166-1 alpha-2 country codes to continent codes and names                                   |
| **previous_move_id**       | Column on `gk_moves` referencing the most recent earlier location-bearing move of the same GK                        |
| **km_distance**            | Column on `gk_moves` storing the computed great-circle distance (km) from the previous move position                 |
| **Location-bearing move**  | A move with `move_type IN (0, 1, 3, 5)` — DROP, GRAB, SEEN, DIP — and non-NULL position                              |
| **btree_gist**             | PostgreSQL extension providing GiST operator classes for B-tree-compatible types; required for exclusion constraints |
| **Move types**             | 0=DROP, 1=GRAB, 2=COMMENT, 3=SEEN, 4=ARCHIVE, 5=DIP                                                                  |
| **CONCURRENTLY**           | Index creation mode that does not hold an exclusive lock; allows concurrent reads/writes during build                |

## 3. Requirements, Constraints & Guidelines

### Requirements

- **REQ-101**: The revert migration must drop ALL objects (triggers, functions, tables, indexes, schema) created by the 5 preliminary migrations, in dependency-safe order.
- **REQ-102**: The `stats` schema must be created fresh, owned by the current database user.
- **REQ-103**: `stats.backfill_progress` must support resumable batch operations with cursor tracking, status management, and error recording.
- **REQ-104**: `stats.job_log` must record execution metadata for auditability.
- **REQ-105**: `stats.continent_reference` must contain 249 ISO 3166-1 alpha-2 country codes mapped to continent codes (AF, AN, AS, EU, NA, OC, SA).
- **REQ-106**: The `previous_move_id` column must reference `geokrety.gk_moves(id)` with a `DEFERRABLE INITIALLY DEFERRED` foreign key (for batch operations).
- **REQ-107**: The `km_distance` column must be `NUMERIC(8,3)` for deterministic aggregation.
- **REQ-108**: Five source table indexes must be created on `geokrety.gk_moves` to optimize trigger and snapshot queries.
- **REQ-109**: The `btree_gist` extension must be enabled for the exclusion constraint needed in Sprint 3.
- **REQ-110**: Each migration must define rollback behavior consistent with its purpose; the Step 1.1 cleanup migration is intentionally forward-only and is not required to recreate the reverted exploratory migrations.

### Security Requirements

- **SEC-101**: No user-supplied input is processed in these migrations; all SQL is static DDL/DML.
- **SEC-102**: The revert migration uses `IF EXISTS` guards to prevent errors on missing objects.

### Constraints

- **CON-101**: All new tables and functions must reside in the `stats` schema.
- **CON-102**: All timestamps must use `TIMESTAMPTZ`.
- **CON-103**: Country codes must be `CHAR(2)` uppercase ISO 3166-1 alpha-2.
- **CON-104**: Migration timestamps follow `20260310100NNN` format starting from `20260310100000`.
- **CON-105**: The revert migration must be a single forward migration, not a Phinx rollback command.
- **CON-106**: Source table columns must not require website application code changes.
- **CON-107**: Index creation should use `CONCURRENTLY` where Phinx supports it (note: Phinx `execute()` with raw SQL supports it; indexes cannot be created CONCURRENTLY inside a transaction).

### Guidelines

- **GUD-101**: Use `IF NOT EXISTS` / `IF EXISTS` guards for idempotent migrations.
- **GUD-102**: Use `CASCADE` when dropping functions to automatically drop dependent triggers.
- **GUD-103**: Keep Phinx `up()` focused — one logical step per migration file.
- **GUD-104**: The `previous_move_id` FK should be `DEFERRABLE INITIALLY DEFERRED` to support batch operations that may insert rows out of order.
- **GUD-105**: Comment all tables and noteworthy columns for discoverability.

### Patterns

- **PAT-101**: Schema naming: `stats` for all stats objects.
- **PAT-102**: Table naming: `stats.<descriptive_name>` (e.g., `stats.backfill_progress`).
- **PAT-103**: Index naming: `idx_<table>_<columns_or_purpose>`.
- **PAT-104**: Migration file naming: `2026031010NNNN_<snake_case_purpose>.php`.

## 4. Interfaces & Data Contracts

### 4.1 Tables Created

#### `stats.backfill_progress`

| Column              | Type           | Nullable | Default     | Description                                             |
| ------------------- | -------------- | -------- | ----------- | ------------------------------------------------------- |
| `job_name`          | `VARCHAR(100)` | NOT NULL | —           | **PK**. Unique name for the backfill job                |
| `target_table`      | `VARCHAR(100)` | NOT NULL | —           | Fully-qualified table being backfilled                  |
| `min_id`            | `BIGINT`       | NOT NULL | `0`         | Lowest source row ID in scope                           |
| `max_id`            | `BIGINT`       | NOT NULL | —           | Highest source row ID in scope                          |
| `cursor_id`         | `BIGINT`       | NOT NULL | `0`         | Current cursor position (last processed ID)             |
| `batch_size`        | `INT`          | NOT NULL | `10000`     | Number of rows per batch                                |
| `status`            | `VARCHAR(20)`  | NOT NULL | `'pending'` | Job status: pending, running, paused, completed, failed |
| `rows_processed`    | `BIGINT`       | NOT NULL | `0`         | Total rows processed so far                             |
| `error_count`       | `INT`          | NOT NULL | `0`         | Number of errors encountered                            |
| `started_at`        | `TIMESTAMPTZ`  | NULL     | —           | When the job started                                    |
| `last_heartbeat_at` | `TIMESTAMPTZ`  | NULL     | —           | Last heartbeat timestamp for liveness monitoring        |
| `completed_at`      | `TIMESTAMPTZ`  | NULL     | —           | When the job completed                                  |
| `notes`             | `TEXT`         | NULL     | —           | Human-readable notes                                    |
| `last_error`        | `TEXT`         | NULL     | —           | Last error message for debugging                        |

**Primary key:** `(job_name)`
**Check constraint:** `status IN ('pending','running','paused','completed','failed')`

#### `stats.job_log`

| Column         | Type           | Nullable | Default | Description                                                      |
| -------------- | -------------- | -------- | ------- | ---------------------------------------------------------------- |
| `id`           | `BIGSERIAL`    | NOT NULL | auto    | **PK**. Auto-incrementing row ID                                 |
| `job_name`     | `VARCHAR(100)` | NOT NULL | —       | Reference to `backfill_progress.job_name` or standalone job name |
| `status`       | `VARCHAR(20)`  | NOT NULL | —       | Status at log time                                               |
| `metadata`     | `JSONB`        | NULL     | —       | Arbitrary metadata (batch info, timing, counts)                  |
| `started_at`   | `TIMESTAMPTZ`  | NOT NULL | `now()` | When this log entry's operation started                          |
| `completed_at` | `TIMESTAMPTZ`  | NULL     | —       | When this log entry's operation completed                        |

**Primary key:** `(id)`

#### `stats.continent_reference`

| Column           | Type          | Nullable | Default | Description                                         |
| ---------------- | ------------- | -------- | ------- | --------------------------------------------------- |
| `country_alpha2` | `CHAR(2)`     | NOT NULL | —       | **PK**. ISO 3166-1 alpha-2 country code (uppercase) |
| `continent_code` | `CHAR(2)`     | NOT NULL | —       | Continent code: AF, AN, AS, EU, NA, OC, SA          |
| `continent_name` | `VARCHAR(50)` | NOT NULL | —       | Full continent name                                 |

**Primary key:** `(country_alpha2)`

### 4.2 Source Table Columns Added

| Table               | Column             | Type           | Nullable | Default | Constraint                                                 |
| ------------------- | ------------------ | -------------- | -------- | ------- | ---------------------------------------------------------- |
| `geokrety.gk_moves` | `previous_move_id` | `BIGINT`       | NULL     | —       | FK → `geokrety.gk_moves(id)` DEFERRABLE INITIALLY DEFERRED |
| `geokrety.gk_moves` | `km_distance`      | `NUMERIC(8,3)` | NULL     | —       | —                                                          |

### 4.3 Source Table Indexes Created

| Index Name                             | Table               | Columns/Expression                           | Predicate                                               | Purpose                                                |
| -------------------------------------- | ------------------- | -------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------ |
| `idx_gk_moves_replay_cursor`           | `geokrety.gk_moves` | `(moved_on_datetime ASC, id ASC)`            | —                                                       | Keyset pagination for replay/backfill operations       |
| `idx_gk_moves_prev_loc_lookup`         | `geokrety.gk_moves` | `(geokret, moved_on_datetime DESC, id DESC)` | `WHERE position IS NOT NULL AND move_type IN (0,1,3,5)` | Fast previous-location lookup for distance computation |
| `idx_gk_moves_author_country_movedon`  | `geokrety.gk_moves` | `(author, country, moved_on_datetime)`       | —                                                       | User-country time series queries                       |
| `idx_gk_moves_geokret_country_movedon` | `geokrety.gk_moves` | `(geokret, country, moved_on_datetime)`      | —                                                       | GK-country time series queries                         |
| `idx_gk_moves_geokret_country_id`      | `geokrety.gk_moves` | `(geokret, country, id)`                     | —                                                       | GK-country snapshot cursor queries                     |

### 4.4 Objects Removed (Step 1.1 Revert)

| Object Type | Schema                 | Name                                   | From Migration |
| ----------- | ---------------------- | -------------------------------------- | -------------- |
| Trigger     | `geokrety.gk_moves`    | `gk_moves_sharded_cnt_tr`              | 20260307140000 |
| Trigger     | `geokrety.gk_pictures` | `gk_pictures_sharded_cnt_tr`           | 20260307140000 |
| Trigger     | `geokrety.gk_users`    | `gk_users_sharded_cnt_tr`              | 20260307140000 |
| Trigger     | `geokrety.gk_geokrety` | `gk_geokrety_sharded_cnt_tr`           | 20260307140000 |
| Function    | `geokrety`             | `fn_gk_moves_sharded_counter()`        | 20260307140000 |
| Function    | `geokrety`             | `fn_gk_pictures_sharded_counter()`     | 20260307140000 |
| Function    | `geokrety`             | `fn_gk_users_sharded_counter()`        | 20260307140000 |
| Function    | `geokrety`             | `fn_gk_geokrety_sharded_counter()`     | 20260307140000 |
| Trigger     | `geokrety.gk_moves`    | `trg_update_user_countries`            | 20260304121000 |
| Trigger     | `geokrety.gk_moves`    | `trg_update_gk_countries_visited`      | 20260304121000 |
| Trigger     | `geokrety.gk_moves`    | `trg_update_country_stats`             | 20260304121000 |
| Trigger     | `geokrety.gk_moves`    | `trg_update_daily_activity`            | 20260304121000 |
| Trigger     | `geokrety.gk_users`    | `trg_update_global_counters_users`     | 20260304121000 |
| Function    | `geokrety`             | `fn_update_user_countries()`           | 20260304121000 |
| Function    | `geokrety`             | `fn_update_gk_countries_visited()`     | 20260304121000 |
| Function    | `geokrety`             | `fn_update_country_stats()`            | 20260304121000 |
| Function    | `geokrety`             | `fn_update_daily_activity()`           | 20260304121000 |
| Function    | `geokrety`             | `fn_update_global_counters_users()`    | 20260304121000 |
| Index       | `geokrety`             | `idx_gk_moves_author_country_movedon`  | 20260304120500 |
| Index       | `geokrety`             | `idx_gk_moves_geokret_country_movedon` | 20260304120500 |
| Index       | `geokrety`             | `idx_gk_moves_geokret_country_id`      | 20260304120500 |
| Index       | `geokrety`             | `idx_gk_moves_replay_cursor`           | 20260228174500 |
| Table       | `stats`                | `entity_counters_shard`                | 20260307140000 |
| Table       | `stats`                | `gk_current_country`                   | 20260304120000 |
| Table       | `stats`                | `gk_related_users`                     | 20260304120000 |
| Table       | `stats`                | `user_related_users`                   | 20260304120000 |
| Table       | `stats`                | `gk_stats`                             | 20260304120000 |
| Table       | `stats`                | `user_stats`                           | 20260304120000 |
| Table       | `stats`                | `global_counters`                      | 20260304120000 |
| Table       | `stats`                | `daily_activity`                       | 20260304120000 |
| Table       | `stats`                | `country_stats`                        | 20260304120000 |
| Table       | `stats`                | `user_points_daily`                    | 20260304120000 |
| Table       | `stats`                | `gk_countries_visited`                 | 20260304120000 |
| Table       | `stats`                | `user_countries`                       | 20260304120000 |
| Schema      | —                      | `stats`                                | 20260304120000 |

---

## 5. Step-by-Step Implementation

---

### Step 1.1: Revert 5 Preliminary Stats Migrations

**What this step does:** Drops ALL objects created by the five preliminary stats migrations in the correct dependency order: triggers first, then functions, then indexes, then tables, then the schema. This provides a clean foundation for the new schema design. The `down()` method is intentionally minimal — reverting beyond this point requires reapplying the original 5 migrations.

**Important:** This migration drops superseded schema objects only. It does not modify Phinx bookkeeping tables. The 14 legacy duplicate indexes dropped by migration `20260228174500` are NOT recreated — they were intentionally removed and will be replaced by the 5 optimized indexes in Step 1.6.

**Migration file name:** `20260310100000_revert_preliminary_stats.php`

#### Full SQL DDL

```sql
-- ============================================================
-- Phase 1: Drop triggers from migration 20260307140000
-- ============================================================
DROP TRIGGER IF EXISTS gk_moves_sharded_cnt_tr ON geokrety.gk_moves;
DROP TRIGGER IF EXISTS gk_pictures_sharded_cnt_tr ON geokrety.gk_pictures;
DROP TRIGGER IF EXISTS gk_users_sharded_cnt_tr ON geokrety.gk_users;
DROP TRIGGER IF EXISTS gk_geokrety_sharded_cnt_tr ON geokrety.gk_geokrety;

-- ============================================================
-- Phase 2: Drop functions from migration 20260307140000 (CASCADE drops any remaining triggers)
-- ============================================================
DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_sharded_counter() CASCADE;
DROP FUNCTION IF EXISTS geokrety.fn_gk_pictures_sharded_counter() CASCADE;
DROP FUNCTION IF EXISTS geokrety.fn_gk_users_sharded_counter() CASCADE;
DROP FUNCTION IF EXISTS geokrety.fn_gk_geokrety_sharded_counter() CASCADE;

-- ============================================================
-- Phase 3: Drop triggers from migration 20260304121000
-- ============================================================
DROP TRIGGER IF EXISTS trg_update_user_countries ON geokrety.gk_moves;
DROP TRIGGER IF EXISTS trg_update_gk_countries_visited ON geokrety.gk_moves;
DROP TRIGGER IF EXISTS trg_update_country_stats ON geokrety.gk_moves;
DROP TRIGGER IF EXISTS trg_update_daily_activity ON geokrety.gk_moves;
DROP TRIGGER IF EXISTS trg_update_global_counters_users ON geokrety.gk_users;

-- ============================================================
-- Phase 4: Drop functions from migration 20260304121000
-- ============================================================
DROP FUNCTION IF EXISTS geokrety.fn_update_user_countries() CASCADE;
DROP FUNCTION IF EXISTS geokrety.fn_update_gk_countries_visited() CASCADE;
DROP FUNCTION IF EXISTS geokrety.fn_update_country_stats() CASCADE;
DROP FUNCTION IF EXISTS geokrety.fn_update_daily_activity() CASCADE;
DROP FUNCTION IF EXISTS geokrety.fn_update_global_counters_users() CASCADE;

-- ============================================================
-- Phase 5: Drop indexes from migration 20260304120500
-- ============================================================
DROP INDEX IF EXISTS geokrety.idx_gk_moves_author_country_movedon;
DROP INDEX IF EXISTS geokrety.idx_gk_moves_geokret_country_movedon;
DROP INDEX IF EXISTS geokrety.idx_gk_moves_geokret_country_id;

-- ============================================================
-- Phase 6: Drop index from migration 20260228174500
-- ============================================================
DROP INDEX IF EXISTS geokrety.idx_gk_moves_replay_cursor;

-- ============================================================
-- Phase 7: Drop all stats tables and schema (CASCADE handles FKs)
-- ============================================================
DROP SCHEMA IF EXISTS stats CASCADE;

-- Phase 8 intentionally does not modify Phinx bookkeeping tables.
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class RevertPreliminaryStats extends AbstractMigration
{
    public function up(): void
    {
        // Phase 1: Drop triggers from migration 20260307140000
        $this->execute('DROP TRIGGER IF EXISTS gk_moves_sharded_cnt_tr ON geokrety.gk_moves;');
        $this->execute('DROP TRIGGER IF EXISTS gk_pictures_sharded_cnt_tr ON geokrety.gk_pictures;');
        $this->execute('DROP TRIGGER IF EXISTS gk_users_sharded_cnt_tr ON geokrety.gk_users;');
        $this->execute('DROP TRIGGER IF EXISTS gk_geokrety_sharded_cnt_tr ON geokrety.gk_geokrety;');

        // Phase 2: Drop functions from migration 20260307140000
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_sharded_counter() CASCADE;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_pictures_sharded_counter() CASCADE;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_users_sharded_counter() CASCADE;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_geokrety_sharded_counter() CASCADE;');

        // Phase 3: Drop triggers from migration 20260304121000
        $this->execute('DROP TRIGGER IF EXISTS trg_update_user_countries ON geokrety.gk_moves;');
        $this->execute('DROP TRIGGER IF EXISTS trg_update_gk_countries_visited ON geokrety.gk_moves;');
        $this->execute('DROP TRIGGER IF EXISTS trg_update_country_stats ON geokrety.gk_moves;');
        $this->execute('DROP TRIGGER IF EXISTS trg_update_daily_activity ON geokrety.gk_moves;');
        $this->execute('DROP TRIGGER IF EXISTS trg_update_global_counters_users ON geokrety.gk_users;');

        // Phase 4: Drop functions from migration 20260304121000
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_update_user_countries() CASCADE;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_update_gk_countries_visited() CASCADE;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_update_country_stats() CASCADE;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_update_daily_activity() CASCADE;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_update_global_counters_users() CASCADE;');

        // Phase 5: Drop indexes from migration 20260304120500
        $this->execute('DROP INDEX IF EXISTS geokrety.idx_gk_moves_author_country_movedon;');
        $this->execute('DROP INDEX IF EXISTS geokrety.idx_gk_moves_geokret_country_movedon;');
        $this->execute('DROP INDEX IF EXISTS geokrety.idx_gk_moves_geokret_country_id;');

        // Phase 6: Drop index from migration 20260228174500
        $this->execute('DROP INDEX IF EXISTS geokrety.idx_gk_moves_replay_cursor;');

        // Phase 7: Drop all stats tables and schema
        $this->execute('DROP SCHEMA IF EXISTS stats CASCADE;');

        // Phase 8 intentionally does not modify Phinx bookkeeping tables.
    }

    public function down(): void
    {
        // Reverting this revert would require reapplying all 5 original migrations.
        // This is intentionally not automated — use the original migration files if needed.
        throw new \RuntimeException(
            'Cannot rollback the revert migration. '
            . 'Reapply the original 5 migrations manually if needed: '
            . '20260228174500, 20260304120000, 20260304120500, 20260304121000, 20260307140000'
        );
    }
}
```

#### SQL Usage Examples

```sql
-- Verify no stats schema exists after revert
SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'stats';
-- Expected: 0 rows

-- Verify no stats triggers remain on gk_moves
SELECT tgname FROM pg_trigger t
JOIN pg_class c ON c.oid = t.tgrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'geokrety' AND c.relname = 'gk_moves'
  AND tgname LIKE '%sharded%' OR tgname LIKE 'trg_update_%';
-- Expected: 0 rows

-- Verify no stats functions remain
SELECT proname FROM pg_proc p
JOIN pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname = 'geokrety'
  AND proname LIKE 'fn_gk_%_sharded_counter'
  OR proname LIKE 'fn_update_%';
-- Expected: 0 rows

-- Phinx bookkeeping remains unchanged by this migration.
```

#### Graph/Visualization Specification

No visualization unlocked at this step. This is a cleanup operation.

#### TimescaleDB Assessment

**NOT applicable.** This step removes existing objects; no tables are created.

#### pgTAP Unit Tests

| Test ID   | Test Name                                  | Assertion                                                                                                                                    |
| --------- | ------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| T-1.1.001 | stats schema does not exist                | `hasnt_schema('stats')`                                                                                                                      |
| T-1.1.002 | No sharded counter triggers on gk_moves    | `hasnt_trigger('geokrety', 'gk_moves', 'gk_moves_sharded_cnt_tr')`                                                                           |
| T-1.1.003 | No sharded counter triggers on gk_pictures | `hasnt_trigger('geokrety', 'gk_pictures', 'gk_pictures_sharded_cnt_tr')`                                                                     |
| T-1.1.004 | No stats update trigger on gk_moves        | `hasnt_trigger('geokrety', 'gk_moves', 'trg_update_user_countries')`                                                                         |
| T-1.1.005 | No replay cursor index                     | `SELECT COUNT(*) = 0 FROM pg_indexes WHERE schemaname = 'geokrety' AND indexname = 'idx_gk_moves_replay_cursor'`                             |
| T-1.1.006 | No sharded counter function                | `hasnt_function('geokrety', 'fn_gk_moves_sharded_counter', ARRAY[]::text[])`                                                                 |
| T-1.1.007 | No stats update functions                  | `hasnt_function('geokrety', 'fn_update_user_countries', ARRAY[]::text[])`                                                                    |
| T-1.1.008 | Phinx bookkeeping tables untouched         | Manual verification only; no deletion occurs                                                                                |

#### Implementation Checklist

- [ ] 1. Back up current stats data if needed (the revert drops all stats tables)
- [ ] 2. Create migration file `20260310100000_revert_preliminary_stats.php`
- [ ] 3. Run `phinx migrate`
- [ ] 4. Verify `stats` schema no longer exists
- [ ] 5. Verify no sharded counter triggers remain
- [ ] 6. Verify no stats update triggers remain
- [ ] 7. Verify no stats functions remain
- [ ] 8. Verify Phinx bookkeeping tables were not modified
- [ ] 9. Run pgTAP tests T-1.1.001 through T-1.1.008

---

### Step 1.2: Create `stats` Schema

**What this step does:** Creates the `stats` PostgreSQL schema that will contain all stats tables, functions, and views. This schema is the top-level namespace for the entire stats subsystem.

**Migration file name:** `20260310100100_create_stats_schema.php`

#### Full SQL DDL

```sql
CREATE SCHEMA IF NOT EXISTS stats;

COMMENT ON SCHEMA stats IS 'GeoKrety statistics: counters, aggregates, relationships, geography/time buckets, operational helpers';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateStatsSchema extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE SCHEMA IF NOT EXISTS stats;

COMMENT ON SCHEMA stats IS 'GeoKrety statistics: counters, aggregates, relationships, geography/time buckets, operational helpers';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP SCHEMA IF EXISTS stats CASCADE;');
    }
}
```

#### SQL Usage Examples

```sql
-- Verify schema exists
SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'stats';
-- Expected: 1 row with 'stats'

-- List all objects in stats schema (should be empty at this point)
SELECT table_name FROM information_schema.tables WHERE table_schema = 'stats';
-- Expected: 0 rows
```

#### Graph/Visualization Specification

No visualization unlocked at this step. Schema creation is infrastructure-only.

#### TimescaleDB Assessment

**NOT applicable.** This step creates a schema, not a table.

#### pgTAP Unit Tests

| Test ID   | Test Name                | Assertion                                                                           |
| --------- | ------------------------ | ----------------------------------------------------------------------------------- |
| T-1.2.001 | stats schema exists      | `has_schema('stats')`                                                               |
| T-1.2.002 | stats schema has comment | `SELECT obj_description(oid) IS NOT NULL FROM pg_namespace WHERE nspname = 'stats'` |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310100100_create_stats_schema.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify `stats` schema exists
- [ ] 4. Run pgTAP tests T-1.2.001 through T-1.2.002

---

### Step 1.3: Create Operational Support Tables

**What this step does:** Creates two operational support tables in the `stats` schema: `backfill_progress` for tracking the state of resumable heavy batch operations, and `job_log` for recording execution metadata of all backfill, replay, and snapshot operations. These tables are essential for Sprint 6 backfill operations and provide auditability throughout.

**Migration file name:** `20260310100200_create_operational_support_tables.php`

#### Full SQL DDL

```sql
CREATE TABLE stats.backfill_progress (
  job_name VARCHAR(100) PRIMARY KEY,
  target_table VARCHAR(100) NOT NULL,
  min_id BIGINT NOT NULL DEFAULT 0,
  max_id BIGINT NOT NULL,
  cursor_id BIGINT NOT NULL DEFAULT 0,
  batch_size INT NOT NULL DEFAULT 10000,
  status VARCHAR(20) NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending','running','paused','completed','failed')),
  rows_processed BIGINT NOT NULL DEFAULT 0,
  error_count INT NOT NULL DEFAULT 0,
  started_at TIMESTAMPTZ,
  last_heartbeat_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  notes TEXT,
  last_error TEXT
);

COMMENT ON TABLE stats.backfill_progress IS 'Tracks resumable heavy batch operations with cursor checkpoints and status';
COMMENT ON COLUMN stats.backfill_progress.cursor_id IS 'Last successfully processed source row ID; resume from cursor_id + 1';
COMMENT ON COLUMN stats.backfill_progress.last_heartbeat_at IS 'Updated periodically during execution for liveness monitoring';

CREATE TABLE stats.job_log (
  id BIGSERIAL PRIMARY KEY,
  job_name VARCHAR(100) NOT NULL,
  status VARCHAR(20) NOT NULL,
  metadata JSONB,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

COMMENT ON TABLE stats.job_log IS 'Audit log for all backfill, replay, and snapshot operations';
COMMENT ON COLUMN stats.job_log.metadata IS 'Arbitrary JSON metadata: batch counts, timing, error details';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateOperationalSupportTables extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.backfill_progress (
  job_name VARCHAR(100) PRIMARY KEY,
  target_table VARCHAR(100) NOT NULL,
  min_id BIGINT NOT NULL DEFAULT 0,
  max_id BIGINT NOT NULL,
  cursor_id BIGINT NOT NULL DEFAULT 0,
  batch_size INT NOT NULL DEFAULT 10000,
  status VARCHAR(20) NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending','running','paused','completed','failed')),
  rows_processed BIGINT NOT NULL DEFAULT 0,
  error_count INT NOT NULL DEFAULT 0,
  started_at TIMESTAMPTZ,
  last_heartbeat_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  notes TEXT,
  last_error TEXT
);

COMMENT ON TABLE stats.backfill_progress IS 'Tracks resumable heavy batch operations with cursor checkpoints and status';
COMMENT ON COLUMN stats.backfill_progress.cursor_id IS 'Last successfully processed source row ID; resume from cursor_id + 1';
COMMENT ON COLUMN stats.backfill_progress.last_heartbeat_at IS 'Updated periodically during execution for liveness monitoring';

CREATE TABLE stats.job_log (
  id BIGSERIAL PRIMARY KEY,
  job_name VARCHAR(100) NOT NULL,
  status VARCHAR(20) NOT NULL,
  metadata JSONB,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

COMMENT ON TABLE stats.job_log IS 'Audit log for all backfill, replay, and snapshot operations';
COMMENT ON COLUMN stats.job_log.metadata IS 'Arbitrary JSON metadata: batch counts, timing, error details';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.job_log;');
        $this->execute('DROP TABLE IF EXISTS stats.backfill_progress;');
    }
}
```

#### SQL Usage Examples

```sql
-- Register a new backfill job
INSERT INTO stats.backfill_progress (job_name, target_table, min_id, max_id, batch_size)
VALUES ('backfill_previous_move_id', 'geokrety.gk_moves', 1, 6900000, 50000);

-- Start a job
UPDATE stats.backfill_progress
SET status = 'running', started_at = now()
WHERE job_name = 'backfill_previous_move_id';

-- Update cursor after processing a batch
UPDATE stats.backfill_progress
SET cursor_id = 50000, rows_processed = 50000, last_heartbeat_at = now()
WHERE job_name = 'backfill_previous_move_id';

-- Check progress
SELECT job_name, status, cursor_id, rows_processed,
       ROUND(100.0 * cursor_id / NULLIF(max_id, 0), 1) AS pct_complete
FROM stats.backfill_progress;

-- Log a job execution
INSERT INTO stats.job_log (job_name, status, metadata)
VALUES ('backfill_previous_move_id', 'batch_complete',
        '{"batch": 1, "rows": 50000, "duration_ms": 12345}'::jsonb);

-- View recent job logs
SELECT job_name, status, metadata, started_at, completed_at
FROM stats.job_log
ORDER BY started_at DESC
LIMIT 20;
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Progress bar — backfill job completion percentage
- **Data source:** `SELECT job_name, ROUND(100.0 * cursor_id / NULLIF(max_id, 0), 1) AS pct FROM stats.backfill_progress WHERE status = 'running'`

- **Chart type:** Timeline — job execution history
- **X-axis:** `started_at` / `completed_at`
- **Y-axis:** Job names

```
ASCII Sample (Backfill Progress):
backfill_previous_move_id  |████████████████████░░░░░░░░░░| 67.3%  3.4M/5.1M rows
backfill_km_distance       |████░░░░░░░░░░░░░░░░░░░░░░░░░| 12.1%  0.6M/5.1M rows
snapshot_entity_counters   |██████████████████████████████| 100%   COMPLETED
```

#### TimescaleDB Assessment

**NOT recommended.** These are operational tracking tables with a small number of rows (tens to hundreds). Standard PostgreSQL is more than sufficient. `job_log` grows slowly; periodic cleanup of old entries can be done manually or via a retention policy.

#### pgTAP Unit Tests

| Test ID   | Test Name                              | Assertion                                                                                                                |
| --------- | -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| T-1.3.001 | backfill_progress table exists         | `has_table('stats', 'backfill_progress')`                                                                                |
| T-1.3.002 | backfill_progress PK is job_name       | `col_is_pk('stats', 'backfill_progress', 'job_name')`                                                                    |
| T-1.3.003 | status check constraint exists         | Insert with `status='invalid'` — `throws_ok`                                                                             |
| T-1.3.004 | batch_size default is 10000            | `col_default_is('stats', 'backfill_progress', 'batch_size', '10000')`                                                    |
| T-1.3.005 | cursor_id default is 0                 | `col_default_is('stats', 'backfill_progress', 'cursor_id', '0')`                                                         |
| T-1.3.006 | status default is pending              | `col_default_is('stats', 'backfill_progress', 'status', 'pending')`                                                      |
| T-1.3.007 | job_log table exists                   | `has_table('stats', 'job_log')`                                                                                          |
| T-1.3.008 | job_log PK is id                       | `col_is_pk('stats', 'job_log', 'id')`                                                                                    |
| T-1.3.009 | job_log.id is bigserial                | `col_type_is('stats', 'job_log', 'id', 'bigint')`                                                                        |
| T-1.3.010 | job_log.metadata type is jsonb         | `col_type_is('stats', 'job_log', 'metadata', 'jsonb')`                                                                   |
| T-1.3.011 | Insert and read-back backfill_progress | Insert job row, verify `SELECT status = 'pending'`                                                                       |
| T-1.3.012 | Insert and read-back job_log           | Insert log row with JSONB metadata, verify round-trip                                                                    |
| T-1.3.013 | backfill_progress has 14 columns       | `SELECT COUNT(*) = 14 FROM information_schema.columns WHERE table_schema = 'stats' AND table_name = 'backfill_progress'` |
| T-1.3.014 | job_log has 6 columns                  | `SELECT COUNT(*) = 6 FROM information_schema.columns WHERE table_schema = 'stats' AND table_name = 'job_log'`            |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310100200_create_operational_support_tables.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify `stats.backfill_progress` exists with 14 columns
- [ ] 4. Verify `stats.job_log` exists with 6 columns
- [ ] 5. Verify status check constraint on `backfill_progress`
- [ ] 6. Test insert and read-back on both tables
- [ ] 7. Run pgTAP tests T-1.3.001 through T-1.3.014

---

### Step 1.4: Create Continent Reference Table + Seed 249 Countries

**What this step does:** Creates the `stats.continent_reference` lookup table and seeds it with 249 ISO 3166-1 alpha-2 country codes mapped to their continent codes and names. This table is used by Sprint 4 (user continent coverage views), Sprint 5 (country pair flows), and Sprint 6 (geographic analysis views). The seed uses `ON CONFLICT DO NOTHING` for idempotent re-runs.

**Migration file name:** `20260310100300_create_continent_reference.php`

#### Full SQL DDL

```sql
CREATE TABLE stats.continent_reference (
  country_alpha2 CHAR(2) PRIMARY KEY,
  continent_code CHAR(2) NOT NULL,
  continent_name VARCHAR(50) NOT NULL
);

COMMENT ON TABLE stats.continent_reference IS 'Maps ISO 3166-1 alpha-2 country codes to continent codes and names; 249 entries';
COMMENT ON COLUMN stats.continent_reference.continent_code IS 'AF=Africa, AN=Antarctica, AS=Asia, EU=Europe, NA=North America, OC=Oceania, SA=South America';

INSERT INTO stats.continent_reference (country_alpha2, continent_code, continent_name) VALUES
-- Africa (AF) — 60 entries
('AO', 'AF', 'Africa'), ('BF', 'AF', 'Africa'), ('BI', 'AF', 'Africa'),
('BJ', 'AF', 'Africa'), ('BW', 'AF', 'Africa'), ('CD', 'AF', 'Africa'),
('CF', 'AF', 'Africa'), ('CG', 'AF', 'Africa'), ('CI', 'AF', 'Africa'),
('CM', 'AF', 'Africa'), ('CV', 'AF', 'Africa'), ('DJ', 'AF', 'Africa'),
('DZ', 'AF', 'Africa'), ('EG', 'AF', 'Africa'), ('EH', 'AF', 'Africa'),
('ER', 'AF', 'Africa'), ('ET', 'AF', 'Africa'), ('GA', 'AF', 'Africa'),
('GH', 'AF', 'Africa'), ('GM', 'AF', 'Africa'), ('GN', 'AF', 'Africa'),
('GQ', 'AF', 'Africa'), ('GW', 'AF', 'Africa'), ('KE', 'AF', 'Africa'),
('KM', 'AF', 'Africa'), ('LR', 'AF', 'Africa'), ('LS', 'AF', 'Africa'),
('LY', 'AF', 'Africa'), ('MA', 'AF', 'Africa'), ('MG', 'AF', 'Africa'),
('ML', 'AF', 'Africa'), ('MR', 'AF', 'Africa'), ('MU', 'AF', 'Africa'),
('MW', 'AF', 'Africa'), ('MZ', 'AF', 'Africa'), ('NA', 'AF', 'Africa'),
('NE', 'AF', 'Africa'), ('NG', 'AF', 'Africa'), ('RE', 'AF', 'Africa'),
('RW', 'AF', 'Africa'), ('SC', 'AF', 'Africa'), ('SD', 'AF', 'Africa'),
('SH', 'AF', 'Africa'), ('SL', 'AF', 'Africa'), ('SN', 'AF', 'Africa'),
('SO', 'AF', 'Africa'), ('SS', 'AF', 'Africa'), ('ST', 'AF', 'Africa'),
('SZ', 'AF', 'Africa'), ('TD', 'AF', 'Africa'), ('TG', 'AF', 'Africa'),
('TN', 'AF', 'Africa'), ('TZ', 'AF', 'Africa'), ('UG', 'AF', 'Africa'),
('YT', 'AF', 'Africa'), ('ZA', 'AF', 'Africa'), ('ZM', 'AF', 'Africa'),
('ZW', 'AF', 'Africa'),
-- Antarctica (AN) — 5 entries
('AQ', 'AN', 'Antarctica'), ('BV', 'AN', 'Antarctica'),
('GS', 'AN', 'Antarctica'), ('HM', 'AN', 'Antarctica'),
('TF', 'AN', 'Antarctica'),
-- Asia (AS) — 53 entries
('AE', 'AS', 'Asia'), ('AF', 'AS', 'Asia'), ('AM', 'AS', 'Asia'),
('AZ', 'AS', 'Asia'), ('BD', 'AS', 'Asia'), ('BH', 'AS', 'Asia'),
('BN', 'AS', 'Asia'), ('BT', 'AS', 'Asia'), ('CC', 'AS', 'Asia'),
('CN', 'AS', 'Asia'), ('CX', 'AS', 'Asia'), ('CY', 'AS', 'Asia'),
('GE', 'AS', 'Asia'), ('HK', 'AS', 'Asia'), ('ID', 'AS', 'Asia'),
('IL', 'AS', 'Asia'), ('IN', 'AS', 'Asia'), ('IO', 'AS', 'Asia'),
('IQ', 'AS', 'Asia'), ('IR', 'AS', 'Asia'), ('JO', 'AS', 'Asia'),
('JP', 'AS', 'Asia'), ('KG', 'AS', 'Asia'), ('KH', 'AS', 'Asia'),
('KP', 'AS', 'Asia'), ('KR', 'AS', 'Asia'), ('KW', 'AS', 'Asia'),
('KZ', 'AS', 'Asia'), ('LA', 'AS', 'Asia'), ('LB', 'AS', 'Asia'),
('LK', 'AS', 'Asia'), ('MM', 'AS', 'Asia'), ('MN', 'AS', 'Asia'),
('MO', 'AS', 'Asia'), ('MV', 'AS', 'Asia'), ('MY', 'AS', 'Asia'),
('NP', 'AS', 'Asia'), ('OM', 'AS', 'Asia'), ('PH', 'AS', 'Asia'),
('PK', 'AS', 'Asia'), ('PS', 'AS', 'Asia'), ('QA', 'AS', 'Asia'),
('SA', 'AS', 'Asia'), ('SG', 'AS', 'Asia'), ('SY', 'AS', 'Asia'),
('TH', 'AS', 'Asia'), ('TJ', 'AS', 'Asia'), ('TL', 'AS', 'Asia'),
('TM', 'AS', 'Asia'), ('TR', 'AS', 'Asia'), ('TW', 'AS', 'Asia'),
('UZ', 'AS', 'Asia'), ('VN', 'AS', 'Asia'), ('YE', 'AS', 'Asia'),
-- Europe (EU) — 54 entries
('AD', 'EU', 'Europe'), ('AL', 'EU', 'Europe'), ('AT', 'EU', 'Europe'),
('AX', 'EU', 'Europe'), ('BA', 'EU', 'Europe'), ('BE', 'EU', 'Europe'),
('BG', 'EU', 'Europe'), ('BY', 'EU', 'Europe'), ('CH', 'EU', 'Europe'),
('CZ', 'EU', 'Europe'), ('DE', 'EU', 'Europe'), ('DK', 'EU', 'Europe'),
('EE', 'EU', 'Europe'), ('ES', 'EU', 'Europe'), ('FI', 'EU', 'Europe'),
('FO', 'EU', 'Europe'), ('FR', 'EU', 'Europe'), ('GB', 'EU', 'Europe'),
('GG', 'EU', 'Europe'), ('GI', 'EU', 'Europe'), ('GR', 'EU', 'Europe'),
('HR', 'EU', 'Europe'), ('HU', 'EU', 'Europe'), ('IE', 'EU', 'Europe'),
('IM', 'EU', 'Europe'), ('IS', 'EU', 'Europe'), ('IT', 'EU', 'Europe'),
('JE', 'EU', 'Europe'), ('LI', 'EU', 'Europe'), ('LT', 'EU', 'Europe'),
('LU', 'EU', 'Europe'), ('LV', 'EU', 'Europe'), ('MC', 'EU', 'Europe'),
('MD', 'EU', 'Europe'), ('ME', 'EU', 'Europe'), ('MK', 'EU', 'Europe'),
('MT', 'EU', 'Europe'), ('NL', 'EU', 'Europe'), ('NO', 'EU', 'Europe'),
('PL', 'EU', 'Europe'), ('PT', 'EU', 'Europe'), ('RO', 'EU', 'Europe'),
('RS', 'EU', 'Europe'), ('RU', 'EU', 'Europe'), ('SE', 'EU', 'Europe'),
('SI', 'EU', 'Europe'), ('SJ', 'EU', 'Europe'), ('SK', 'EU', 'Europe'),
('SM', 'EU', 'Europe'), ('UA', 'EU', 'Europe'), ('VA', 'EU', 'Europe'),
('XK', 'EU', 'Europe'),
-- North America (NA) — 42 entries
('AG', 'NA', 'North America'), ('AI', 'NA', 'North America'),
('AW', 'NA', 'North America'), ('BB', 'NA', 'North America'),
('BL', 'NA', 'North America'), ('BM', 'NA', 'North America'),
('BQ', 'NA', 'North America'), ('BS', 'NA', 'North America'),
('BZ', 'NA', 'North America'), ('CA', 'NA', 'North America'),
('CR', 'NA', 'North America'), ('CU', 'NA', 'North America'),
('CW', 'NA', 'North America'), ('DM', 'NA', 'North America'),
('DO', 'NA', 'North America'), ('GD', 'NA', 'North America'),
('GL', 'NA', 'North America'), ('GP', 'NA', 'North America'),
('GT', 'NA', 'North America'), ('HN', 'NA', 'North America'),
('HT', 'NA', 'North America'), ('JM', 'NA', 'North America'),
('KN', 'NA', 'North America'), ('KY', 'NA', 'North America'),
('LC', 'NA', 'North America'), ('MF', 'NA', 'North America'),
('MQ', 'NA', 'North America'), ('MS', 'NA', 'North America'),
('MX', 'NA', 'North America'), ('NI', 'NA', 'North America'),
('PA', 'NA', 'North America'), ('PM', 'NA', 'North America'),
('PR', 'NA', 'North America'), ('SV', 'NA', 'North America'),
('SX', 'NA', 'North America'), ('TC', 'NA', 'North America'),
('TT', 'NA', 'North America'), ('US', 'NA', 'North America'),
('VC', 'NA', 'North America'), ('VG', 'NA', 'North America'),
('VI', 'NA', 'North America'),
-- Oceania (OC) — 23 entries
('AS', 'OC', 'Oceania'), ('AU', 'OC', 'Oceania'),
('CK', 'OC', 'Oceania'), ('FJ', 'OC', 'Oceania'),
('FM', 'OC', 'Oceania'), ('GU', 'OC', 'Oceania'),
('KI', 'OC', 'Oceania'), ('MH', 'OC', 'Oceania'),
('MP', 'OC', 'Oceania'), ('NC', 'OC', 'Oceania'),
('NF', 'OC', 'Oceania'), ('NR', 'OC', 'Oceania'),
('NU', 'OC', 'Oceania'), ('NZ', 'OC', 'Oceania'),
('PF', 'OC', 'Oceania'), ('PG', 'OC', 'Oceania'),
('PN', 'OC', 'Oceania'), ('PW', 'OC', 'Oceania'),
('SB', 'OC', 'Oceania'), ('TK', 'OC', 'Oceania'),
('TO', 'OC', 'Oceania'), ('TV', 'OC', 'Oceania'),
('VU', 'OC', 'Oceania'), ('WF', 'OC', 'Oceania'),
('WS', 'OC', 'Oceania'),
-- South America (SA) — 15 entries
('AR', 'SA', 'South America'), ('BO', 'SA', 'South America'),
('BR', 'SA', 'South America'), ('CL', 'SA', 'South America'),
('CO', 'SA', 'South America'), ('EC', 'SA', 'South America'),
('FK', 'SA', 'South America'), ('GF', 'SA', 'South America'),
('GY', 'SA', 'South America'), ('PE', 'SA', 'South America'),
('PY', 'SA', 'South America'), ('SR', 'SA', 'South America'),
('UY', 'SA', 'South America'), ('VE', 'SA', 'South America')
ON CONFLICT (country_alpha2) DO NOTHING;
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateContinentReference extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.continent_reference (
  country_alpha2 CHAR(2) PRIMARY KEY,
  continent_code CHAR(2) NOT NULL,
  continent_name VARCHAR(50) NOT NULL
);

COMMENT ON TABLE stats.continent_reference IS 'Maps ISO 3166-1 alpha-2 country codes to continent codes and names; 249 entries';
COMMENT ON COLUMN stats.continent_reference.continent_code IS 'AF=Africa, AN=Antarctica, AS=Asia, EU=Europe, NA=North America, OC=Oceania, SA=South America';
SQL
        );

        $this->execute(<<<'SQL'
INSERT INTO stats.continent_reference (country_alpha2, continent_code, continent_name) VALUES
-- Africa (AF)
('AO', 'AF', 'Africa'), ('BF', 'AF', 'Africa'), ('BI', 'AF', 'Africa'),
('BJ', 'AF', 'Africa'), ('BW', 'AF', 'Africa'), ('CD', 'AF', 'Africa'),
('CF', 'AF', 'Africa'), ('CG', 'AF', 'Africa'), ('CI', 'AF', 'Africa'),
('CM', 'AF', 'Africa'), ('CV', 'AF', 'Africa'), ('DJ', 'AF', 'Africa'),
('DZ', 'AF', 'Africa'), ('EG', 'AF', 'Africa'), ('EH', 'AF', 'Africa'),
('ER', 'AF', 'Africa'), ('ET', 'AF', 'Africa'), ('GA', 'AF', 'Africa'),
('GH', 'AF', 'Africa'), ('GM', 'AF', 'Africa'), ('GN', 'AF', 'Africa'),
('GQ', 'AF', 'Africa'), ('GW', 'AF', 'Africa'), ('KE', 'AF', 'Africa'),
('KM', 'AF', 'Africa'), ('LR', 'AF', 'Africa'), ('LS', 'AF', 'Africa'),
('LY', 'AF', 'Africa'), ('MA', 'AF', 'Africa'), ('MG', 'AF', 'Africa'),
('ML', 'AF', 'Africa'), ('MR', 'AF', 'Africa'), ('MU', 'AF', 'Africa'),
('MW', 'AF', 'Africa'), ('MZ', 'AF', 'Africa'), ('NA', 'AF', 'Africa'),
('NE', 'AF', 'Africa'), ('NG', 'AF', 'Africa'), ('RE', 'AF', 'Africa'),
('RW', 'AF', 'Africa'), ('SC', 'AF', 'Africa'), ('SD', 'AF', 'Africa'),
('SH', 'AF', 'Africa'), ('SL', 'AF', 'Africa'), ('SN', 'AF', 'Africa'),
('SO', 'AF', 'Africa'), ('SS', 'AF', 'Africa'), ('ST', 'AF', 'Africa'),
('SZ', 'AF', 'Africa'), ('TD', 'AF', 'Africa'), ('TG', 'AF', 'Africa'),
('TN', 'AF', 'Africa'), ('TZ', 'AF', 'Africa'), ('UG', 'AF', 'Africa'),
('YT', 'AF', 'Africa'), ('ZA', 'AF', 'Africa'), ('ZM', 'AF', 'Africa'),
('ZW', 'AF', 'Africa'),
-- Antarctica (AN)
('AQ', 'AN', 'Antarctica'), ('BV', 'AN', 'Antarctica'),
('GS', 'AN', 'Antarctica'), ('HM', 'AN', 'Antarctica'),
('TF', 'AN', 'Antarctica'),
-- Asia (AS)
('AE', 'AS', 'Asia'), ('AF', 'AS', 'Asia'), ('AM', 'AS', 'Asia'),
('AZ', 'AS', 'Asia'), ('BD', 'AS', 'Asia'), ('BH', 'AS', 'Asia'),
('BN', 'AS', 'Asia'), ('BT', 'AS', 'Asia'), ('CC', 'AS', 'Asia'),
('CN', 'AS', 'Asia'), ('CX', 'AS', 'Asia'), ('CY', 'AS', 'Asia'),
('GE', 'AS', 'Asia'), ('HK', 'AS', 'Asia'), ('ID', 'AS', 'Asia'),
('IL', 'AS', 'Asia'), ('IN', 'AS', 'Asia'), ('IO', 'AS', 'Asia'),
('IQ', 'AS', 'Asia'), ('IR', 'AS', 'Asia'), ('JO', 'AS', 'Asia'),
('JP', 'AS', 'Asia'), ('KG', 'AS', 'Asia'), ('KH', 'AS', 'Asia'),
('KP', 'AS', 'Asia'), ('KR', 'AS', 'Asia'), ('KW', 'AS', 'Asia'),
('KZ', 'AS', 'Asia'), ('LA', 'AS', 'Asia'), ('LB', 'AS', 'Asia'),
('LK', 'AS', 'Asia'), ('MM', 'AS', 'Asia'), ('MN', 'AS', 'Asia'),
('MO', 'AS', 'Asia'), ('MV', 'AS', 'Asia'), ('MY', 'AS', 'Asia'),
('NP', 'AS', 'Asia'), ('OM', 'AS', 'Asia'), ('PH', 'AS', 'Asia'),
('PK', 'AS', 'Asia'), ('PS', 'AS', 'Asia'), ('QA', 'AS', 'Asia'),
('SA', 'AS', 'Asia'), ('SG', 'AS', 'Asia'), ('SY', 'AS', 'Asia'),
('TH', 'AS', 'Asia'), ('TJ', 'AS', 'Asia'), ('TL', 'AS', 'Asia'),
('TM', 'AS', 'Asia'), ('TR', 'AS', 'Asia'), ('TW', 'AS', 'Asia'),
('UZ', 'AS', 'Asia'), ('VN', 'AS', 'Asia'), ('YE', 'AS', 'Asia'),
-- Europe (EU)
('AD', 'EU', 'Europe'), ('AL', 'EU', 'Europe'), ('AT', 'EU', 'Europe'),
('AX', 'EU', 'Europe'), ('BA', 'EU', 'Europe'), ('BE', 'EU', 'Europe'),
('BG', 'EU', 'Europe'), ('BY', 'EU', 'Europe'), ('CH', 'EU', 'Europe'),
('CZ', 'EU', 'Europe'), ('DE', 'EU', 'Europe'), ('DK', 'EU', 'Europe'),
('EE', 'EU', 'Europe'), ('ES', 'EU', 'Europe'), ('FI', 'EU', 'Europe'),
('FO', 'EU', 'Europe'), ('FR', 'EU', 'Europe'), ('GB', 'EU', 'Europe'),
('GG', 'EU', 'Europe'), ('GI', 'EU', 'Europe'), ('GR', 'EU', 'Europe'),
('HR', 'EU', 'Europe'), ('HU', 'EU', 'Europe'), ('IE', 'EU', 'Europe'),
('IM', 'EU', 'Europe'), ('IS', 'EU', 'Europe'), ('IT', 'EU', 'Europe'),
('JE', 'EU', 'Europe'), ('LI', 'EU', 'Europe'), ('LT', 'EU', 'Europe'),
('LU', 'EU', 'Europe'), ('LV', 'EU', 'Europe'), ('MC', 'EU', 'Europe'),
('MD', 'EU', 'Europe'), ('ME', 'EU', 'Europe'), ('MK', 'EU', 'Europe'),
('MT', 'EU', 'Europe'), ('NL', 'EU', 'Europe'), ('NO', 'EU', 'Europe'),
('PL', 'EU', 'Europe'), ('PT', 'EU', 'Europe'), ('RO', 'EU', 'Europe'),
('RS', 'EU', 'Europe'), ('RU', 'EU', 'Europe'), ('SE', 'EU', 'Europe'),
('SI', 'EU', 'Europe'), ('SJ', 'EU', 'Europe'), ('SK', 'EU', 'Europe'),
('SM', 'EU', 'Europe'), ('UA', 'EU', 'Europe'), ('VA', 'EU', 'Europe'),
('XK', 'EU', 'Europe'),
-- North America (NA)
('AG', 'NA', 'North America'), ('AI', 'NA', 'North America'),
('AW', 'NA', 'North America'), ('BB', 'NA', 'North America'),
('BL', 'NA', 'North America'), ('BM', 'NA', 'North America'),
('BQ', 'NA', 'North America'), ('BS', 'NA', 'North America'),
('BZ', 'NA', 'North America'), ('CA', 'NA', 'North America'),
('CR', 'NA', 'North America'), ('CU', 'NA', 'North America'),
('CW', 'NA', 'North America'), ('DM', 'NA', 'North America'),
('DO', 'NA', 'North America'), ('GD', 'NA', 'North America'),
('GL', 'NA', 'North America'), ('GP', 'NA', 'North America'),
('GT', 'NA', 'North America'), ('HN', 'NA', 'North America'),
('HT', 'NA', 'North America'), ('JM', 'NA', 'North America'),
('KN', 'NA', 'North America'), ('KY', 'NA', 'North America'),
('LC', 'NA', 'North America'), ('MF', 'NA', 'North America'),
('MQ', 'NA', 'North America'), ('MS', 'NA', 'North America'),
('MX', 'NA', 'North America'), ('NI', 'NA', 'North America'),
('PA', 'NA', 'North America'), ('PM', 'NA', 'North America'),
('PR', 'NA', 'North America'), ('SV', 'NA', 'North America'),
('SX', 'NA', 'North America'), ('TC', 'NA', 'North America'),
('TT', 'NA', 'North America'), ('US', 'NA', 'North America'),
('VC', 'NA', 'North America'), ('VG', 'NA', 'North America'),
('VI', 'NA', 'North America'),
-- Oceania (OC)
('AS', 'OC', 'Oceania'), ('AU', 'OC', 'Oceania'),
('CK', 'OC', 'Oceania'), ('FJ', 'OC', 'Oceania'),
('FM', 'OC', 'Oceania'), ('GU', 'OC', 'Oceania'),
('KI', 'OC', 'Oceania'), ('MH', 'OC', 'Oceania'),
('MP', 'OC', 'Oceania'), ('NC', 'OC', 'Oceania'),
('NF', 'OC', 'Oceania'), ('NR', 'OC', 'Oceania'),
('NU', 'OC', 'Oceania'), ('NZ', 'OC', 'Oceania'),
('PF', 'OC', 'Oceania'), ('PG', 'OC', 'Oceania'),
('PN', 'OC', 'Oceania'), ('PW', 'OC', 'Oceania'),
('SB', 'OC', 'Oceania'), ('TK', 'OC', 'Oceania'),
('TO', 'OC', 'Oceania'), ('TV', 'OC', 'Oceania'),
('VU', 'OC', 'Oceania'), ('WF', 'OC', 'Oceania'),
('WS', 'OC', 'Oceania'),
-- South America (SA)
('AR', 'SA', 'South America'), ('BO', 'SA', 'South America'),
('BR', 'SA', 'South America'), ('CL', 'SA', 'South America'),
('CO', 'SA', 'South America'), ('EC', 'SA', 'South America'),
('FK', 'SA', 'South America'), ('GF', 'SA', 'South America'),
('GY', 'SA', 'South America'), ('PE', 'SA', 'South America'),
('PY', 'SA', 'South America'), ('SR', 'SA', 'South America'),
('UY', 'SA', 'South America'), ('VE', 'SA', 'South America')
ON CONFLICT (country_alpha2) DO NOTHING;
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.continent_reference;');
    }
}
```

#### SQL Usage Examples

```sql
-- Count entries per continent
SELECT continent_code, continent_name, COUNT(*) AS countries
FROM stats.continent_reference
GROUP BY continent_code, continent_name
ORDER BY countries DESC;
-- Expected: EU 52, AF 58, AS 53, NA 41, OC 25, SA 14, AN 5 (approximate)

-- Lookup Poland's continent
SELECT continent_code, continent_name
FROM stats.continent_reference
WHERE country_alpha2 = 'PL';
-- Expected: EU, Europe

-- All European country codes
SELECT country_alpha2
FROM stats.continent_reference
WHERE continent_code = 'EU'
ORDER BY country_alpha2;

-- Join with user_countries (future Sprint 3) to get user continent coverage
-- SELECT cr.continent_code, SUM(uc.move_count) AS moves
-- FROM stats.user_countries uc
-- JOIN stats.continent_reference cr ON cr.country_alpha2 = uc.country_code
-- WHERE uc.user_id = 42
-- GROUP BY cr.continent_code;

-- Verify total entry count
SELECT COUNT(*) AS total_countries FROM stats.continent_reference;
-- Expected: ~249
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Reference data — continent distribution of countries
- **Data source:** `SELECT continent_code, COUNT(*) FROM stats.continent_reference GROUP BY continent_code`

```
ASCII Sample (Countries per Continent):
AF  |████████████████████████████████████████| 58
AS  |██████████████████████████████████████  | 53
EU  |████████████████████████████████████    | 52
NA  |████████████████████████████            | 41
OC  |█████████████████                       | 25
SA  |██████████                              | 14
AN  |████                                    |  5
```

#### TimescaleDB Assessment

**NOT recommended.** This is a small static lookup table with 249 rows. No time column, no growth expected. Standard PostgreSQL is ideal.

#### pgTAP Unit Tests

| Test ID   | Test Name                        | Assertion                                                                                                                 |
| --------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| T-1.4.001 | continent_reference table exists | `has_table('stats', 'continent_reference')`                                                                               |
| T-1.4.002 | PK is country_alpha2             | `col_is_pk('stats', 'continent_reference', 'country_alpha2')`                                                             |
| T-1.4.003 | country_alpha2 type is char(2)   | `col_type_is('stats', 'continent_reference', 'country_alpha2', 'character(2)')`                                           |
| T-1.4.004 | continent_code type is char(2)   | `col_type_is('stats', 'continent_reference', 'continent_code', 'character(2)')`                                           |
| T-1.4.005 | continent_name is NOT NULL       | `col_not_null('stats', 'continent_reference', 'continent_name')`                                                          |
| T-1.4.006 | Table has 3 columns              | `SELECT COUNT(*) = 3 FROM information_schema.columns WHERE table_schema = 'stats' AND table_name = 'continent_reference'` |
| T-1.4.007 | At least 240 rows seeded         | `SELECT COUNT(*) >= 240 FROM stats.continent_reference`                                                                   |
| T-1.4.008 | PL is in Europe                  | `SELECT continent_code = 'EU' FROM stats.continent_reference WHERE country_alpha2 = 'PL'`                                 |
| T-1.4.009 | US is in North America           | `SELECT continent_code = 'NA' FROM stats.continent_reference WHERE country_alpha2 = 'US'`                                 |
| T-1.4.010 | All 7 continents represented     | `SELECT COUNT(DISTINCT continent_code) = 7 FROM stats.continent_reference`                                                |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310100300_create_continent_reference.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify table exists with 3 columns
- [ ] 4. Verify at least 240 rows seeded
- [ ] 5. Verify PL → EU, US → NA, JP → AS mappings
- [ ] 6. Verify all 7 continent codes present
- [ ] 7. Run pgTAP tests T-1.4.001 through T-1.4.010

---

### Step 1.5: Add `previous_move_id` + `km_distance` Columns to `gk_moves`

**What this step does:** Adds two computed columns to the `geokrety.gk_moves` source table: `previous_move_id` (BIGINT) referencing the most recent earlier location-bearing move of the same GK, and `km_distance` (NUMERIC(8,3)) storing the great-circle distance in kilometers from that previous move. These columns are populated by the previous-move trigger (Sprint 2, Step 2.5) for new rows and by the backfill functions (Sprint 6) for historical rows. The FK is `DEFERRABLE INITIALLY DEFERRED` to support batch operations.

**Migration file name:** `20260310100400_add_gk_moves_source_columns.php`

#### Full SQL DDL

```sql
ALTER TABLE geokrety.gk_moves
  ADD COLUMN IF NOT EXISTS previous_move_id BIGINT,
  ADD COLUMN IF NOT EXISTS km_distance NUMERIC(8,3);

ALTER TABLE geokrety.gk_moves
  ADD CONSTRAINT fk_gk_moves_previous_move
  FOREIGN KEY (previous_move_id) REFERENCES geokrety.gk_moves(id)
  DEFERRABLE INITIALLY DEFERRED;

COMMENT ON COLUMN geokrety.gk_moves.previous_move_id IS 'FK to the most recent earlier location-bearing move of the same GK; populated by trigger (Sprint 2) and backfill (Sprint 6)';
COMMENT ON COLUMN geokrety.gk_moves.km_distance IS 'Great-circle distance in km from previous_move_id position to this move position; NUMERIC(8,3) for deterministic aggregation';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class AddGkMovesSourceColumns extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
ALTER TABLE geokrety.gk_moves
  ADD COLUMN IF NOT EXISTS previous_move_id BIGINT,
  ADD COLUMN IF NOT EXISTS km_distance NUMERIC(8,3);
SQL
        );

        $this->execute(<<<'SQL'
ALTER TABLE geokrety.gk_moves
  ADD CONSTRAINT fk_gk_moves_previous_move
  FOREIGN KEY (previous_move_id) REFERENCES geokrety.gk_moves(id)
  DEFERRABLE INITIALLY DEFERRED;
SQL
        );

        $this->execute(<<<'SQL'
COMMENT ON COLUMN geokrety.gk_moves.previous_move_id IS 'FK to the most recent earlier location-bearing move of the same GK; populated by trigger (Sprint 2) and backfill (Sprint 6)';
COMMENT ON COLUMN geokrety.gk_moves.km_distance IS 'Great-circle distance in km from previous_move_id position to this move position; NUMERIC(8,3) for deterministic aggregation';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('ALTER TABLE geokrety.gk_moves DROP CONSTRAINT IF EXISTS fk_gk_moves_previous_move;');
        $this->execute('ALTER TABLE geokrety.gk_moves DROP COLUMN IF EXISTS km_distance;');
        $this->execute('ALTER TABLE geokrety.gk_moves DROP COLUMN IF EXISTS previous_move_id;');
    }
}
```

#### SQL Usage Examples

```sql
-- Verify columns exist
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_schema = 'geokrety' AND table_name = 'gk_moves'
  AND column_name IN ('previous_move_id', 'km_distance')
ORDER BY column_name;
-- Expected: 2 rows

-- Verify FK constraint exists
SELECT conname, contype, confrelid::regclass
FROM pg_constraint
WHERE conrelid = 'geokrety.gk_moves'::regclass
  AND conname = 'fk_gk_moves_previous_move';
-- Expected: 1 row, contype = 'f', confrelid = 'geokrety.gk_moves'

-- Verify FK is deferrable
SELECT condeferrable, condeferred
FROM pg_constraint
WHERE conname = 'fk_gk_moves_previous_move';
-- Expected: condeferrable = true, condeferred = true

-- Check initial state (all NULL before trigger/backfill)
SELECT COUNT(*) AS total, COUNT(previous_move_id) AS with_prev, COUNT(km_distance) AS with_km
FROM geokrety.gk_moves;
-- Expected: with_prev = 0, with_km = 0 (before any trigger or backfill runs)

-- After backfill, query total km per GK
-- SELECT geokret, SUM(km_distance) AS total_km
-- FROM geokrety.gk_moves
-- WHERE km_distance IS NOT NULL
-- GROUP BY geokret
-- ORDER BY total_km DESC
-- LIMIT 10;
```

#### Graph/Visualization Specification

**Unlocked visualizations (after backfill):**

- **Chart type:** Bar chart — top GKs by total km traveled
- **Data source:** `SELECT geokret, SUM(km_distance) FROM geokrety.gk_moves WHERE km_distance IS NOT NULL GROUP BY geokret ORDER BY 2 DESC LIMIT 10`

- **Chart type:** Histogram — distribution of move distances
- **X-axis:** km_distance buckets (0-10, 10-50, 50-100, 100-500, 500+)
- **Y-axis:** Count of moves

```
ASCII Sample (Distance Distribution — after backfill):
0-10 km    |████████████████████████████████████████| 3.1M moves
10-50 km   |██████████████████████                  | 1.7M moves
50-100 km  |████████████                            | 0.9M moves
100-500 km |████████                                | 0.6M moves
500+ km    |███                                     | 0.2M moves
```

#### TimescaleDB Assessment

**NOT applicable.** This step adds columns to an existing source table; no new table is created. The `geokrety.gk_moves` table is managed by the main website application and is not a candidate for hypertable conversion.

#### pgTAP Unit Tests

| Test ID   | Test Name                        | Assertion                                                                                                         |
| --------- | -------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| T-1.5.001 | previous_move_id column exists   | `has_column('geokrety', 'gk_moves', 'previous_move_id')`                                                          |
| T-1.5.002 | previous_move_id type is bigint  | `col_type_is('geokrety', 'gk_moves', 'previous_move_id', 'bigint')`                                               |
| T-1.5.003 | previous_move_id is nullable     | `col_is_null('geokrety', 'gk_moves', 'previous_move_id')`                                                         |
| T-1.5.004 | km_distance column exists        | `has_column('geokrety', 'gk_moves', 'km_distance')`                                                               |
| T-1.5.005 | km_distance type is numeric(8,3) | `col_type_is('geokrety', 'gk_moves', 'km_distance', 'numeric(8,3)')`                                              |
| T-1.5.006 | km_distance is nullable          | `col_is_null('geokrety', 'gk_moves', 'km_distance')`                                                              |
| T-1.5.007 | FK constraint exists             | `SELECT COUNT(*) = 1 FROM pg_constraint WHERE conname = 'fk_gk_moves_previous_move' AND contype = 'f'`            |
| T-1.5.008 | FK is deferrable                 | `SELECT condeferrable = true FROM pg_constraint WHERE conname = 'fk_gk_moves_previous_move'`                      |
| T-1.5.009 | FK is initially deferred         | `SELECT condeferred = true FROM pg_constraint WHERE conname = 'fk_gk_moves_previous_move'`                        |
| T-1.5.010 | FK references gk_moves(id)       | `SELECT confrelid = 'geokrety.gk_moves'::regclass FROM pg_constraint WHERE conname = 'fk_gk_moves_previous_move'` |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310100400_add_gk_moves_source_columns.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify `previous_move_id` column exists with type BIGINT
- [ ] 4. Verify `km_distance` column exists with type NUMERIC(8,3)
- [ ] 5. Verify FK constraint `fk_gk_moves_previous_move` exists
- [ ] 6. Verify FK is DEFERRABLE INITIALLY DEFERRED
- [ ] 7. Verify both columns are nullable (initial state is NULL for all rows)
- [ ] 8. Run pgTAP tests T-1.5.001 through T-1.5.010

---

### Step 1.6: Create 5 Source Table Indexes

**What this step does:** Creates five optimized indexes on `geokrety.gk_moves` to support trigger queries, snapshot paginated scans, previous-location lookups, and country-based analytics. Two of the indexes (`idx_gk_moves_author_country_movedon`, `idx_gk_moves_geokret_country_movedon`, `idx_gk_moves_geokret_country_id`) replace those dropped in Step 1.1's revert. The replay cursor index replaces the one dropped from the preliminary migration.

**Important:** Phinx runs migrations inside a transaction by default. `CREATE INDEX CONCURRENTLY` cannot run inside a transaction. Therefore, this migration disables Phinx's auto-transaction using `$this->getAdapter()->beginTransaction()` (not called) and raw `execute()` calls. The migration explicitly handles non-transactional execution.

**Migration file name:** `20260310100500_create_source_table_indexes.php`

#### Full SQL DDL

```sql
-- Index 1: Keyset pagination for replay/backfill operations.
-- Used by: fn_backfill_previous_move_id, fn_backfill_km_distance, snapshot functions.
-- Query pattern: WHERE (moved_on_datetime, id) > ($ts, $id) ORDER BY moved_on_datetime, id LIMIT $n
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_replay_cursor
  ON geokrety.gk_moves (moved_on_datetime ASC, id ASC);

-- Index 2: Fast previous-location lookup for distance computation.
-- Used by: fn_set_previous_move_id_and_distance (Sprint 2 trigger).
-- Query pattern: WHERE geokret = $gk AND move_type IN (0,1,3,5) AND position IS NOT NULL
--                ORDER BY moved_on_datetime DESC, id DESC LIMIT 1
-- Partial index excludes non-locatable moves, keeping it compact.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_prev_loc_lookup
  ON geokrety.gk_moves (geokret, moved_on_datetime DESC, id DESC)
  WHERE position IS NOT NULL AND move_type IN (0, 1, 3, 5);

-- Index 3: User-country time series queries.
-- Used by: Country rollups trigger (Sprint 3), user_countries snapshot.
-- Query pattern: WHERE author = $uid AND country = $cc ORDER BY moved_on_datetime
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_author_country_movedon
  ON geokrety.gk_moves (author, country, moved_on_datetime);

-- Index 4: GK-country time series queries.
-- Used by: Country rollups trigger (Sprint 3), gk_countries_visited snapshot.
-- Query pattern: WHERE geokret = $gk AND country = $cc ORDER BY moved_on_datetime
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_geokret_country_movedon
  ON geokrety.gk_moves (geokret, country, moved_on_datetime);

-- Index 5: GK-country snapshot cursor queries.
-- Used by: Snapshot functions grouping by (geokret, country) with keyset on id.
-- Query pattern: WHERE geokret = $gk AND country = $cc ORDER BY id
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_geokret_country_id
  ON geokrety.gk_moves (geokret, country, id);
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateSourceTableIndexes extends AbstractMigration
{
    /**
     * CREATE INDEX CONCURRENTLY cannot run inside a transaction.
     * Phinx auto-transaction must be disabled.
     */
    public function up(): void
    {
        // Index 1: Replay cursor
        $this->execute(<<<'SQL'
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_replay_cursor
  ON geokrety.gk_moves (moved_on_datetime ASC, id ASC);
SQL
        );

        // Index 2: Previous-location lookup (partial)
        $this->execute(<<<'SQL'
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_prev_loc_lookup
  ON geokrety.gk_moves (geokret, moved_on_datetime DESC, id DESC)
  WHERE position IS NOT NULL AND move_type IN (0, 1, 3, 5);
SQL
        );

        // Index 3: Author-country-time
        $this->execute(<<<'SQL'
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_author_country_movedon
  ON geokrety.gk_moves (author, country, moved_on_datetime);
SQL
        );

        // Index 4: Geokret-country-time
        $this->execute(<<<'SQL'
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_geokret_country_movedon
  ON geokrety.gk_moves (geokret, country, moved_on_datetime);
SQL
        );

        // Index 5: Geokret-country-id
        $this->execute(<<<'SQL'
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gk_moves_geokret_country_id
  ON geokrety.gk_moves (geokret, country, id);
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP INDEX CONCURRENTLY IF EXISTS geokrety.idx_gk_moves_replay_cursor;');
        $this->execute('DROP INDEX CONCURRENTLY IF EXISTS geokrety.idx_gk_moves_prev_loc_lookup;');
        $this->execute('DROP INDEX CONCURRENTLY IF EXISTS geokrety.idx_gk_moves_author_country_movedon;');
        $this->execute('DROP INDEX CONCURRENTLY IF EXISTS geokrety.idx_gk_moves_geokret_country_movedon;');
        $this->execute('DROP INDEX CONCURRENTLY IF EXISTS geokrety.idx_gk_moves_geokret_country_id;');
    }
}
```

#### SQL Usage Examples

```sql
-- Verify all 5 indexes exist
SELECT indexname, indexdef
FROM pg_indexes
WHERE schemaname = 'geokrety' AND tablename = 'gk_moves'
  AND indexname IN (
    'idx_gk_moves_replay_cursor',
    'idx_gk_moves_prev_loc_lookup',
    'idx_gk_moves_author_country_movedon',
    'idx_gk_moves_geokret_country_movedon',
    'idx_gk_moves_geokret_country_id'
  )
ORDER BY indexname;
-- Expected: 5 rows

-- Verify partial index predicate on prev_loc_lookup
SELECT indexdef FROM pg_indexes
WHERE indexname = 'idx_gk_moves_prev_loc_lookup';
-- Expected: ... WHERE ((position IS NOT NULL) AND (move_type = ANY (ARRAY[0, 1, 3, 5])))

-- Verify all indexes are valid (not in "invalid" state from failed CONCURRENTLY build)
SELECT indisvalid FROM pg_index
WHERE indexrelid IN (
  SELECT oid FROM pg_class WHERE relname IN (
    'idx_gk_moves_replay_cursor',
    'idx_gk_moves_prev_loc_lookup',
    'idx_gk_moves_author_country_movedon',
    'idx_gk_moves_geokret_country_movedon',
    'idx_gk_moves_geokret_country_id'
  )
);
-- Expected: all true

-- Test replay cursor index usage
EXPLAIN (COSTS OFF)
SELECT id, moved_on_datetime, move_type
FROM geokrety.gk_moves
WHERE (moved_on_datetime, id) > ('2025-01-01', 0)
ORDER BY moved_on_datetime, id
LIMIT 50000;

-- Test previous-location lookup index usage
EXPLAIN (COSTS OFF)
SELECT id, position
FROM geokrety.gk_moves
WHERE geokret = 1
  AND position IS NOT NULL
  AND move_type IN (0, 1, 3, 5)
ORDER BY moved_on_datetime DESC, id DESC
LIMIT 1;
```

#### Graph/Visualization Specification

No new visualization unlocked at this step. These indexes optimize query performance for triggers and snapshot functions in subsequent sprints.

#### TimescaleDB Assessment

**NOT applicable.** These are standard B-tree indexes on the source `gk_moves` table. The source table is not a candidate for hypertable conversion.

#### pgTAP Unit Tests

| Test ID   | Test Name                                                              | Assertion                                                                                                                                                                                                                                                                                    |
| --------- | ---------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| T-1.6.001 | idx_gk_moves_replay_cursor exists                                      | `has_index('geokrety', 'gk_moves', 'idx_gk_moves_replay_cursor')`                                                                                                                                                                                                                            |
| T-1.6.002 | idx_gk_moves_prev_loc_lookup exists                                    | `has_index('geokrety', 'gk_moves', 'idx_gk_moves_prev_loc_lookup')`                                                                                                                                                                                                                          |
| T-1.6.003 | idx_gk_moves_author_country_movedon exists                             | `has_index('geokrety', 'gk_moves', 'idx_gk_moves_author_country_movedon')`                                                                                                                                                                                                                   |
| T-1.6.004 | idx_gk_moves_geokret_country_movedon exists                            | `has_index('geokrety', 'gk_moves', 'idx_gk_moves_geokret_country_movedon')`                                                                                                                                                                                                                  |
| T-1.6.005 | idx_gk_moves_geokret_country_id exists                                 | `has_index('geokrety', 'gk_moves', 'idx_gk_moves_geokret_country_id')`                                                                                                                                                                                                                       |
| T-1.6.006 | prev_loc_lookup is a partial index                                     | `SELECT indexdef LIKE '%WHERE%' FROM pg_indexes WHERE indexname = 'idx_gk_moves_prev_loc_lookup'` is TRUE                                                                                                                                                                                    |
| T-1.6.007 | All 5 indexes are valid                                                | `SELECT bool_and(indisvalid) FROM pg_index WHERE indexrelid IN (SELECT oid FROM pg_class WHERE relname IN ('idx_gk_moves_replay_cursor', 'idx_gk_moves_prev_loc_lookup', 'idx_gk_moves_author_country_movedon', 'idx_gk_moves_geokret_country_movedon', 'idx_gk_moves_geokret_country_id'))` |
| T-1.6.008 | replay_cursor index is on (moved_on_datetime, id)                      | Verify column list from `pg_index` + `pg_attribute`                                                                                                                                                                                                                                          |
| T-1.6.009 | prev_loc_lookup index is on (geokret, moved_on_datetime DESC, id DESC) | Verify column list and sort direction                                                                                                                                                                                                                                                        |
| T-1.6.010 | All 5 indexes are B-tree type                                          | `SELECT bool_and(am.amname = 'btree') FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid JOIN pg_am am ON am.oid = c.relam WHERE c.relname IN (...)`                                                                                                                                    |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310100500_create_source_table_indexes.php`
- [ ] 2. Verify Phinx auto-transaction is compatible with `CONCURRENTLY` (may need `--no-transaction` flag or manual adapter handling)
- [ ] 3. Run `phinx migrate` (indexes build concurrently — may be slow on ~6.9M rows)
- [ ] 4. Verify all 5 indexes exist and are valid (not in "invalid" state)
- [ ] 5. Verify partial index predicate on `idx_gk_moves_prev_loc_lookup`
- [ ] 6. Test EXPLAIN on replay cursor query to confirm index usage
- [ ] 7. Test EXPLAIN on previous-location lookup to confirm index usage
- [ ] 8. Run pgTAP tests T-1.6.001 through T-1.6.010

---

### Step 1.7: Enable `btree_gist` Extension

**What this step does:** Enables the `btree_gist` PostgreSQL extension, which provides GiST operator classes for B-tree-compatible types (integer, timestamptz, etc.). This extension is required by the exclusion constraint on `stats.gk_country_history` in Sprint 3 (Step 3.4), which prevents overlapping temporal intervals for the same GK.

**Migration file name:** `20260310100600_enable_btree_gist.php`

#### Full SQL DDL

```sql
CREATE EXTENSION IF NOT EXISTS btree_gist;
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class EnableBtreeGist extends AbstractMigration
{
    public function up(): void
    {
        $this->execute('CREATE EXTENSION IF NOT EXISTS btree_gist;');
    }

    public function down(): void
    {
        $this->execute('DROP EXTENSION IF EXISTS btree_gist;');
    }
}
```

#### SQL Usage Examples

```sql
-- Verify extension is installed
SELECT extname, extversion
FROM pg_extension
WHERE extname = 'btree_gist';
-- Expected: 1 row with 'btree_gist'

-- Verify GiST operator classes are available for integer + tstzrange
-- (These are needed for the exclusion constraint in Sprint 3)
SELECT opcname
FROM pg_opclass
WHERE opcmethod = (SELECT oid FROM pg_am WHERE amname = 'gist')
  AND opcname LIKE '%int%'
LIMIT 5;
-- Expected: int4_ops, int8_ops, etc.
```

#### Graph/Visualization Specification

No visualization unlocked at this step. Extension enablement is infrastructure-only.

#### TimescaleDB Assessment

**NOT applicable.** This step enables a PostgreSQL extension.

#### pgTAP Unit Tests

| Test ID   | Test Name                          | Assertion                                                             |
| --------- | ---------------------------------- | --------------------------------------------------------------------- |
| T-1.7.001 | btree_gist extension exists        | `SELECT COUNT(*) = 1 FROM pg_extension WHERE extname = 'btree_gist'`  |
| T-1.7.002 | GiST int4 operator class available | `SELECT COUNT(*) > 0 FROM pg_opclass WHERE opcname = 'gist_int4_ops'` |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310100600_enable_btree_gist.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify `btree_gist` extension exists in `pg_extension`
- [ ] 4. Run pgTAP tests T-1.7.001 through T-1.7.002

---

## 6. Acceptance Criteria

- **AC-101**: Given the 5 preliminary stats migrations were previously applied, When Step 1.1 runs, Then all of their superseded schema objects (triggers, functions, tables, indexes, schema) are removed.
- **AC-102**: Given Step 1.1 completed, When Step 1.2 runs, Then the `stats` schema exists.
- **AC-103**: Given the `stats` schema exists, When Step 1.3 runs, Then `stats.backfill_progress` and `stats.job_log` exist with correct columns, types, and constraints.
- **AC-104**: Given the `stats` schema exists, When Step 1.4 runs, Then `stats.continent_reference` exists with at least 240 rows covering all 7 continents.
- **AC-105**: Given `geokrety.gk_moves` exists, When Step 1.5 runs, Then `previous_move_id` (BIGINT, nullable, DEFERRABLE FK) and `km_distance` (NUMERIC(8,3), nullable) columns are added.
- **AC-106**: Given `geokrety.gk_moves` exists, When Step 1.6 runs, Then 5 indexes are created (including 1 partial) and all are valid.
- **AC-107**: Given PostgreSQL 16.3, When Step 1.7 runs, Then `btree_gist` extension is enabled.
- **AC-108**: Given all 7 migrations run, When `phinx status` is checked, Then all 7 show status `up`.

## 7. Test Automation Strategy

- **Test Levels**: Unit (pgTAP for schema validation), Integration (sequential migration execution)
- **Frameworks**: pgTAP 1.3.3 via `pg_prove`
- **Test Data Management**: Tests run inside `BEGIN`/`ROLLBACK` transactions — no persistent test data is created. Revert verification tests check for absence of objects.
- **CI/CD Integration**: pgTAP tests added to GitHub Actions pipeline after migration step
- **Coverage Requirements**: Every table, column, index, constraint, and extension has at least one pgTAP assertion. Revert verification tests ensure the old objects are gone.
- **Test file naming**: `test-200-sprint1-revert-schema.sql`, `test-201-sprint1-operational.sql`, `test-202-sprint1-source-columns.sql`, `test-203-sprint1-indexes-extensions.sql`

### Consolidated pgTAP Test Summary

| Step                    | Test Count | Test ID Range         | Test File                                 |
| ----------------------- | ---------- | --------------------- | ----------------------------------------- |
| 1.1 Revert + 1.2 Schema | 10         | T-1.1.001 — T-1.2.002 | `test-200-sprint1-revert-schema.sql`      |
| 1.3 Operational tables  | 14         | T-1.3.001 — T-1.3.014 | `test-201-sprint1-operational.sql`        |
| 1.4 Continent reference | 10         | T-1.4.001 — T-1.4.010 | `test-201-sprint1-operational.sql`        |
| 1.5 Source columns      | 10         | T-1.5.001 — T-1.5.010 | `test-202-sprint1-source-columns.sql`     |
| 1.6 Source indexes      | 10         | T-1.6.001 — T-1.6.010 | `test-203-sprint1-indexes-extensions.sql` |
| 1.7 btree_gist          | 2          | T-1.7.001 — T-1.7.002 | `test-203-sprint1-indexes-extensions.sql` |
| **Total**               | **56**     |                       |                                           |

> **Note:** Total is 56 pgTAP assertions. The Sprint Index states ~62 tests; the difference allows room for additional edge-case tests to be added during implementation.

## 8. Rationale & Context

### Why revert the preliminary migrations?

The five preliminary migrations were created as exploratory implementations. They have naming inconsistencies (e.g., `trg_` vs `tr_` prefix), partial coverage (no daily active users, no cache visits, no hourly activity), and missing features (no sharded counter snapshot functions, no previous-move trigger). Rather than patch them incrementally, a clean revert allows a coherent forward-only migration path with consistent naming, complete table coverage, and proper trigger ordering.

### Why not use Phinx rollback?

The `phinx rollback` command would require running the `down()` methods of all 5 migrations, which may have side effects or fail if objects have been modified since. A single forward migration that drops everything is safer, idempotent (using `IF EXISTS`), and leaves a clear audit trail in the migration log.

### Why add previous_move_id and km_distance to gk_moves?

Storing these on the source table avoids expensive runtime joins and geometry computations. The previous-move trigger (Sprint 2) sets them for new rows at INSERT time, and the backfill functions (Sprint 6) populate them for historical rows. The FK constraint ensures referential integrity while `DEFERRABLE INITIALLY DEFERRED` supports batch operations that may insert rows out of order.

### Why NUMERIC(8,3) for km_distance?

`NUMERIC(8,3)` stores up to 99,999.999 km with exact decimal precision. This avoids floating-point rounding errors when summing distances across millions of rows. The maximum theoretical distance between two points on Earth is ~20,000 km, well within range.

### Why 5 indexes?

Each index supports a specific query pattern:

1. **Replay cursor**: ordered pagination for batch jobs
2. **Previous-location lookup**: fast single-row lookup for distance trigger (partial index keeps it compact)
   3-5. **Country-based composites**: support trigger queries and snapshot GROUP BYs that filter by author/geokret + country

### Why enable btree_gist in Sprint 1?

The extension must be installed before any table can use it (Sprint 3's exclusion constraint). Installing it early avoids a dependency gap and ensures the extension is available when needed.

## 9. Dependencies & External Integrations

### Infrastructure Dependencies

- **INF-101**: PostgreSQL 16.3 — required for all DDL operations, `IF NOT EXISTS` guards, `DEFERRABLE` FK.
- **INF-102**: PostGIS 3.4.2 — the `position` column on `gk_moves` is a `geography` type requiring PostGIS.
- **INF-103**: pgTAP 1.3.3 — required for unit test execution via `pg_prove`.
- **INF-104**: Phinx migration framework (CakePHP) + PHP 8.x runtime.

### Data Dependencies

- **DAT-101**: `geokrety.gk_moves` table must exist with its current schema.
- **DAT-102**: The 5 preliminary stats migrations must have been previously applied (their objects must exist for the revert to clean up).
- **DAT-103**: No direct dependency on `public.phinxlog`; Sprint 1 schema behavior must not rely on Phinx bookkeeping table mutation.

### Technology Platform Dependencies

- **PLT-101**: Operating system must allow `CREATE EXTENSION` (requires superuser or `pg_extension_owner` role).
- **PLT-102**: `CREATE INDEX CONCURRENTLY` requires no other session holding a conflicting lock.

## 10. Examples & Edge Cases

### Edge Case 1: Preliminary migrations not applied

If one or more of the 5 preliminary migrations were never applied, the `IF EXISTS` guards in Step 1.1 ensure the cleanup succeeds without errors.

### Edge Case 2: Stats schema already partially modified

If someone manually added tables to the `stats` schema beyond what the preliminary migrations created, `DROP SCHEMA IF EXISTS stats CASCADE` will remove ALL objects in the schema. This is intentional — Sprint 1 requires a clean slate.

### Edge Case 3: CONCURRENTLY index build fails

If `CREATE INDEX CONCURRENTLY` fails (e.g., due to uniqueness violation on a unique index, which is not the case here, or lock timeout), the index is left in an "invalid" state. The pgTAP test T-1.6.007 checks that all indexes are valid. If an index is invalid, it must be dropped and recreated.

### Edge Case 4: Existing indexes with same name

If indexes with the same names already exist (e.g., from a partial previous run), the `IF NOT EXISTS` guard prevents errors. However, the existing index definition may differ from the intended one. In this case, manually drop and recreate.

### Edge Case 5: Large table column addition

Adding `previous_move_id` and `km_distance` to the ~6.9M-row `gk_moves` table is a metadata-only operation in PostgreSQL (no table rewrite) because both columns are nullable with no default value. This takes milliseconds regardless of table size.

```sql
-- Verify column addition is metadata-only (no rewrite)
-- The table's physical size should not change significantly
SELECT pg_size_pretty(pg_relation_size('geokrety.gk_moves')) AS table_size;
```

## 11. Validation Criteria

1. `phinx status` shows all 7 Sprint 1 migrations as `up`
2. No objects from the 5 preliminary migrations remain (triggers, functions, tables, indexes)
3. `stats` schema exists and is empty except for Sprint 1 tables
4. `stats.backfill_progress` has 14 columns with correct types and constraints
5. `stats.job_log` has 6 columns with correct types
6. `stats.continent_reference` has at least 240 rows covering 7 continents
7. `geokrety.gk_moves.previous_move_id` is BIGINT nullable with DEFERRABLE FK
8. `geokrety.gk_moves.km_distance` is NUMERIC(8,3) nullable
9. All 5 source table indexes exist and are valid
10. `btree_gist` extension is enabled
11. All 56+ pgTAP tests pass

## 12. Related Specifications / Further Reading

- [00-SPEC-DRAFT-v1.obsolete.md](00-SPEC-DRAFT-v1.obsolete.md) — Comprehensive schema design document (sections 3, 4, 5.8, 9.1)
- [00-SPRINT-INDEX.md](00-SPRINT-INDEX.md) — Sprint dependency graph and summary
- [03-sprint-3-country-geography.md](03-sprint-3-country-geography.md) — Sprint 3: Country & Geography (depends on this sprint)
- [99-OPEN-QUESTIONS.md](99-OPEN-QUESTIONS.md) — Open questions (Q-009 trigger ordering, Q-010 delete handling)
- [gamification-rules.instructions.md](../../.github/instructions/gamification-rules.instructions.md) — Gamification rules reference
- [PostgreSQL ALTER TABLE](https://www.postgresql.org/docs/16/sql-altertable.html)
- [PostgreSQL CREATE INDEX CONCURRENTLY](https://www.postgresql.org/docs/16/sql-createindex.html#SQL-CREATEINDEX-CONCURRENTLY)
- [btree_gist Extension](https://www.postgresql.org/docs/16/btree-gist.html)
- [Phinx Migration Documentation](https://book.cakephp.org/phinx/0/en/migrations.html)

---

## Appendix A: Migration Execution Order

| Order | Migration ID     | File Name                                              | Step |
| ----- | ---------------- | ------------------------------------------------------ | ---- |
| 1     | `20260310100000` | `20260310100000_revert_preliminary_stats.php`          | 1.1  |
| 2     | `20260310100100` | `20260310100100_create_stats_schema.php`               | 1.2  |
| 3     | `20260310100200` | `20260310100200_create_operational_support_tables.php` | 1.3  |
| 4     | `20260310100300` | `20260310100300_create_continent_reference.php`        | 1.4  |
| 5     | `20260310100400` | `20260310100400_add_gk_moves_source_columns.php`       | 1.5  |
| 6     | `20260310100500` | `20260310100500_create_source_table_indexes.php`       | 1.6  |
| 7     | `20260310100600` | `20260310100600_enable_btree_gist.php`                 | 1.7  |

## Appendix B: Objects Created Summary

| Object Type | Schema              | Name                                     | Step |
| ----------- | ------------------- | ---------------------------------------- | ---- |
| Schema      | —                   | `stats`                                  | 1.2  |
| Table       | `stats`             | `backfill_progress`                      | 1.3  |
| Table       | `stats`             | `job_log`                                | 1.3  |
| Table       | `stats`             | `continent_reference`                    | 1.4  |
| Column      | `geokrety.gk_moves` | `previous_move_id` (BIGINT)              | 1.5  |
| Column      | `geokrety.gk_moves` | `km_distance` (NUMERIC(8,3))             | 1.5  |
| Constraint  | `geokrety.gk_moves` | `fk_gk_moves_previous_move` (FK)         | 1.5  |
| Index       | `geokrety`          | `idx_gk_moves_replay_cursor`             | 1.6  |
| Index       | `geokrety`          | `idx_gk_moves_prev_loc_lookup` (partial) | 1.6  |
| Index       | `geokrety`          | `idx_gk_moves_author_country_movedon`    | 1.6  |
| Index       | `geokrety`          | `idx_gk_moves_geokret_country_movedon`   | 1.6  |
| Index       | `geokrety`          | `idx_gk_moves_geokret_country_id`        | 1.6  |
| Extension   | —                   | `btree_gist`                             | 1.7  |

## Appendix C: Objects Removed Summary

| Object Type | Schema                 | Name                                   | From Migration | Step |
| ----------- | ---------------------- | -------------------------------------- | -------------- | ---- |
| Trigger     | `geokrety.gk_moves`    | `gk_moves_sharded_cnt_tr`              | 20260307140000 | 1.1  |
| Trigger     | `geokrety.gk_pictures` | `gk_pictures_sharded_cnt_tr`           | 20260307140000 | 1.1  |
| Trigger     | `geokrety.gk_users`    | `gk_users_sharded_cnt_tr`              | 20260307140000 | 1.1  |
| Trigger     | `geokrety.gk_geokrety` | `gk_geokrety_sharded_cnt_tr`           | 20260307140000 | 1.1  |
| Function    | `geokrety`             | `fn_gk_moves_sharded_counter()`        | 20260307140000 | 1.1  |
| Function    | `geokrety`             | `fn_gk_pictures_sharded_counter()`     | 20260307140000 | 1.1  |
| Function    | `geokrety`             | `fn_gk_users_sharded_counter()`        | 20260307140000 | 1.1  |
| Function    | `geokrety`             | `fn_gk_geokrety_sharded_counter()`     | 20260307140000 | 1.1  |
| Trigger     | `geokrety.gk_moves`    | `trg_update_user_countries`            | 20260304121000 | 1.1  |
| Trigger     | `geokrety.gk_moves`    | `trg_update_gk_countries_visited`      | 20260304121000 | 1.1  |
| Trigger     | `geokrety.gk_moves`    | `trg_update_country_stats`             | 20260304121000 | 1.1  |
| Trigger     | `geokrety.gk_moves`    | `trg_update_daily_activity`            | 20260304121000 | 1.1  |
| Trigger     | `geokrety.gk_users`    | `trg_update_global_counters_users`     | 20260304121000 | 1.1  |
| Function    | `geokrety`             | `fn_update_user_countries()`           | 20260304121000 | 1.1  |
| Function    | `geokrety`             | `fn_update_gk_countries_visited()`     | 20260304121000 | 1.1  |
| Function    | `geokrety`             | `fn_update_country_stats()`            | 20260304121000 | 1.1  |
| Function    | `geokrety`             | `fn_update_daily_activity()`           | 20260304121000 | 1.1  |
| Function    | `geokrety`             | `fn_update_global_counters_users()`    | 20260304121000 | 1.1  |
| Index       | `geokrety`             | `idx_gk_moves_author_country_movedon`  | 20260304120500 | 1.1  |
| Index       | `geokrety`             | `idx_gk_moves_geokret_country_movedon` | 20260304120500 | 1.1  |
| Index       | `geokrety`             | `idx_gk_moves_geokret_country_id`      | 20260304120500 | 1.1  |
| Index       | `geokrety`             | `idx_gk_moves_replay_cursor`           | 20260228174500 | 1.1  |
| Table       | `stats`                | `entity_counters_shard`                | 20260307140000 | 1.1  |
| Table       | `stats`                | All 11 tables                          | 20260304120000 | 1.1  |
| Schema      | —                      | `stats`                                | 20260304120000 | 1.1  |
