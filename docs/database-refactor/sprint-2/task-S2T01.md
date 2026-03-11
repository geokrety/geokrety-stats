---
title: "Task S2T01: Create stats.entity_counters_shard Table"
version: 1.0
date_created: 2026-03-08
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 2
tags:
  - counters
  - daily-activity
  - database
  - migration
  - postgresql
  - schema
  - sprint-2
  - stats
  - task-merge
  - triggers
depends_on: [1]
task: S2T01
step: 2.1
migration: 20260310200000_create_entity_counters_shard.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026-03-10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.1
---

# Task S2T01: Create stats.entity_counters_shard Table

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `../00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.1` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.1: Create `stats.entity_counters_shard` Table

**What this step does:** Creates the `stats.entity_counters_shard` table that provides low-contention exact counts for all major entity categories. Instead of a single counter row (which causes hot-row lock contention under high write load), each entity is split across 16 shard rows (shard 0–15). The shard for any given row is computed as `id % 16`. To read the total for an entity, sum across all shards: `SELECT SUM(cnt) FROM stats.entity_counters_shard WHERE entity = 'gk_moves'`. After table creation, 400 shard rows (25 entities × 16 shards) are pre-initialized to zero.

**Migration file name:** `20260310200000_create_entity_counters_shard.php`

#### Full SQL DDL

```sql
CREATE TABLE stats.entity_counters_shard (
  entity VARCHAR(32) NOT NULL,
  shard INT NOT NULL,
  cnt BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (entity, shard)
);

COMMENT ON TABLE stats.entity_counters_shard IS 'Sharded counter table for exact entity counts; use SUM(cnt) per entity for total';
COMMENT ON COLUMN stats.entity_counters_shard.entity IS 'Counter entity name, e.g.: gk_moves, gk_moves_type_0, gk_geokrety_type_3';
COMMENT ON COLUMN stats.entity_counters_shard.shard IS 'Shard index (0-15) for low-contention concurrent increments';

-- Pre-initialize 16 shard rows for each of the 25 tracked entities
-- Entities: gk_moves total + 6 types, gk_geokrety total + 11 types,
--           gk_pictures total + 3 types, gk_users, gk_loves
INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
SELECT e.entity, s.shard, 0
FROM (
  VALUES
    ('gk_moves'), ('gk_moves_type_0'), ('gk_moves_type_1'), ('gk_moves_type_2'),
    ('gk_moves_type_3'), ('gk_moves_type_4'), ('gk_moves_type_5'),
    ('gk_geokrety'), ('gk_geokrety_type_0'), ('gk_geokrety_type_1'),
    ('gk_geokrety_type_2'), ('gk_geokrety_type_3'), ('gk_geokrety_type_4'),
    ('gk_geokrety_type_5'), ('gk_geokrety_type_6'), ('gk_geokrety_type_7'),
    ('gk_geokrety_type_8'), ('gk_geokrety_type_9'), ('gk_geokrety_type_10'),
    ('gk_pictures'), ('gk_pictures_type_0'), ('gk_pictures_type_1'),
    ('gk_pictures_type_2'), ('gk_users'), ('gk_loves')
) AS e(entity)
CROSS JOIN generate_series(0, 15) AS s(shard);
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateEntityCountersShard extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.entity_counters_shard (
  entity VARCHAR(32) NOT NULL,
  shard INT NOT NULL,
  cnt BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (entity, shard)
);

COMMENT ON TABLE stats.entity_counters_shard IS 'Sharded counter table for exact entity counts; use SUM(cnt) per entity for total';
COMMENT ON COLUMN stats.entity_counters_shard.entity IS 'Counter entity name, e.g.: gk_moves, gk_moves_type_0, gk_geokrety_type_3';
COMMENT ON COLUMN stats.entity_counters_shard.shard IS 'Shard index (0-15) for low-contention concurrent increments';

INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
SELECT e.entity, s.shard, 0
FROM (
  VALUES
    ('gk_moves'), ('gk_moves_type_0'), ('gk_moves_type_1'), ('gk_moves_type_2'),
    ('gk_moves_type_3'), ('gk_moves_type_4'), ('gk_moves_type_5'),
    ('gk_geokrety'), ('gk_geokrety_type_0'), ('gk_geokrety_type_1'),
    ('gk_geokrety_type_2'), ('gk_geokrety_type_3'), ('gk_geokrety_type_4'),
    ('gk_geokrety_type_5'), ('gk_geokrety_type_6'), ('gk_geokrety_type_7'),
    ('gk_geokrety_type_8'), ('gk_geokrety_type_9'), ('gk_geokrety_type_10'),
    ('gk_pictures'), ('gk_pictures_type_0'), ('gk_pictures_type_1'),
    ('gk_pictures_type_2'), ('gk_users'), ('gk_loves')
) AS e(entity)
CROSS JOIN generate_series(0, 15) AS s(shard);
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.entity_counters_shard;');
    }
}
```

#### SQL Usage Examples

```sql
-- Read total move count (sum across 16 shards)
SELECT SUM(cnt) AS total_moves
FROM stats.entity_counters_shard
WHERE entity = 'gk_moves';

-- Read per-type breakdown for all move types
SELECT entity, SUM(cnt) AS total
FROM stats.entity_counters_shard
WHERE entity LIKE 'gk_moves_type_%'
GROUP BY entity
ORDER BY entity;

-- Dashboard KPI: all entity totals in one query
SELECT entity, SUM(cnt) AS total
FROM stats.entity_counters_shard
GROUP BY entity
ORDER BY entity;

-- Read total GeoKrety by type
SELECT entity, SUM(cnt) AS count
FROM stats.entity_counters_shard
WHERE entity LIKE 'gk_geokrety%'
GROUP BY entity
ORDER BY entity;

-- Verify 25 entities × 16 shards = 400 rows pre-initialized
SELECT COUNT(*) AS total_shard_rows FROM stats.entity_counters_shard;
-- Expected: 400
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Donut chart — Move type distribution
  - **Data source:** `SELECT entity, SUM(cnt) FROM stats.entity_counters_shard WHERE entity LIKE 'gk_moves_type_%' GROUP BY entity`
  - **Series:** 6 segments (drop, grab, comment, seen, archive, dip)

- **Chart type:** Donut chart — GeoKrety type distribution
  - **Data source:** `SELECT entity, SUM(cnt) FROM stats.entity_counters_shard WHERE entity LIKE 'gk_geokrety_type_%' GROUP BY entity`
  - **Series:** 11 segments (one per GK type)

- **Chart type:** Horizontal bar — Platform entity totals
  - **Data source:** `SELECT entity, SUM(cnt) FROM stats.entity_counters_shard WHERE entity IN ('gk_moves','gk_geokrety','gk_users','gk_pictures','gk_loves') GROUP BY entity`

```
ASCII Sample (KPI Counter Card Sources):
Total Moves     |  6,931,442  (SUM over gk_moves)
Total GeoKrety  |    412,881  (SUM over gk_geokrety)
Total Users     |    148,332  (SUM over gk_users)
Total Pictures  |    234,541  (SUM over gk_pictures)
Total Loves     |     87,209  (SUM over gk_loves)

Move Type Distribution:
drop ██████████████████ 38%
grab ████████████████   34%
dip  ████               8%
seen ████               8%
com  ████               8%
arch █                  4%
```

#### TimescaleDB Assessment

**NOT recommended.** `entity_counters_shard` is a 400-row dimension table with PK `(entity, shard)`. It is not a time-series table and has no time column. All writes are idempotent `INSERT ... ON CONFLICT DO UPDATE` upserts on fixed rows. TimescaleDB hypertable partitioning requires a time column and is entirely inappropriate here. Standard PostgreSQL B-tree index on the PK handles all reads and writes optimally.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.1.001 | entity_counters_shard table exists | `has_table('stats', 'entity_counters_shard')` |
| T-2.1.002 | PK is (entity, shard) | `col_is_pk('stats', 'entity_counters_shard', ARRAY['entity', 'shard'])` |
| T-2.1.003 | entity column type is varchar(32) | `col_type_is('stats', 'entity_counters_shard', 'entity', 'character varying(32)')` |
| T-2.1.004 | shard column type is integer | `col_type_is('stats', 'entity_counters_shard', 'shard', 'integer')` |
| T-2.1.005 | cnt default is 0 | `col_default_is('stats', 'entity_counters_shard', 'cnt', '0')` |
| T-2.1.006 | 400 rows pre-initialized | `SELECT is(COUNT(*)::INT, 400) FROM stats.entity_counters_shard` |
| T-2.1.007 | 25 distinct entities pre-initialized | `SELECT is(COUNT(DISTINCT entity)::INT, 25) FROM stats.entity_counters_shard` |
| T-2.1.008 | Each entity has 16 shards | `SELECT is(MIN(cnt_shards)::INT, 16) FROM (SELECT entity, COUNT(*) AS cnt_shards FROM stats.entity_counters_shard GROUP BY entity) sub` |
| T-2.1.009 | gk_moves entity present | `SELECT is(COUNT(*)::INT, 16) FROM stats.entity_counters_shard WHERE entity = 'gk_moves'` |
| T-2.1.010 | gk_loves entity present | `SELECT is(COUNT(*)::INT, 16) FROM stats.entity_counters_shard WHERE entity = 'gk_loves'` |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310200000_create_entity_counters_shard.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify table exists with 3 columns and correct types
- [ ] 4. Verify composite PK on `(entity, shard)`
- [ ] 5. Verify 400 rows were inserted (25 entities × 16 shards)
- [ ] 6. Verify all cnt values are 0
- [ ] 7. Run pgTAP tests T-2.1.001 through T-2.1.010

---
