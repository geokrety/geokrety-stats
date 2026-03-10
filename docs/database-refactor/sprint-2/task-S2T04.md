---
title: "Task S2T04: Create stats.daily_entity_counts Table"
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
task: S2T04
step: 2.4
migration: 20260310200300_create_daily_entity_counts.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026-03-10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.4
---

# Task S2T04: Create stats.daily_entity_counts Table

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.4` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.4: Create `stats.daily_entity_counts` Table

**What this step does:** Creates the `stats.daily_entity_counts` table that stores daily entity snapshot totals by calendar day. This is the time-series companion to `entity_counters_shard`: while the shard table gives the current live total, `daily_entity_counts` records the daily seeded value used by charts and reconciliations. A nightly/manual job (Sprint 6) inserts one row per `(count_date, entity)` by reading the shard totals.

**Migration file name:** `20260310200300_create_daily_entity_counts.php`

#### Full SQL DDL

```sql
CREATE TABLE stats.daily_entity_counts (
  count_date DATE NOT NULL,
  entity VARCHAR(32) NOT NULL,
  cnt BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (count_date, entity)
);

COMMENT ON TABLE stats.daily_entity_counts IS 'Daily cumulative entity counts for trend charts; populated by nightly snapshot job';
COMMENT ON COLUMN stats.daily_entity_counts.entity IS 'Entity name matching entity_counters_shard.entity';
COMMENT ON COLUMN stats.daily_entity_counts.cnt IS 'Entity count snapshot value for count_date';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateDailyEntityCounts extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.daily_entity_counts (
  count_date DATE NOT NULL,
  entity VARCHAR(32) NOT NULL,
  cnt BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (count_date, entity)
);

COMMENT ON TABLE stats.daily_entity_counts IS 'Daily cumulative entity counts for trend charts; populated by nightly snapshot job';
COMMENT ON COLUMN stats.daily_entity_counts.entity IS 'Entity name matching entity_counters_shard.entity';
COMMENT ON COLUMN stats.daily_entity_counts.cnt IS 'Entity count snapshot value for count_date';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.daily_entity_counts;');
    }
}
```

#### SQL Usage Examples

```sql
-- GeoKrety daily snapshot counts for the last year
SELECT count_date, cnt
FROM stats.daily_entity_counts
WHERE entity = 'gk_geokrety'
  AND count_date >= CURRENT_DATE - INTERVAL '1 year'
ORDER BY count_date;

-- Platform growth chart: all major entities over time
SELECT count_date,
  MAX(CASE WHEN entity = 'gk_moves' THEN cnt END) AS total_moves,
  MAX(CASE WHEN entity = 'gk_geokrety' THEN cnt END) AS total_geokrety,
  MAX(CASE WHEN entity = 'gk_users' THEN cnt END) AS total_users
FROM stats.daily_entity_counts
WHERE entity IN ('gk_moves', 'gk_geokrety', 'gk_users')
GROUP BY count_date
ORDER BY count_date;

-- Growth rate: day-over-day new moves (delta from snapshot counts)
SELECT count_date,
  cnt,
  cnt - LAG(cnt, 1, 0) OVER (ORDER BY count_date) AS daily_new
FROM stats.daily_entity_counts
WHERE entity = 'gk_moves'
ORDER BY count_date;

-- Upsert today's snapshot (called by nightly job)
INSERT INTO stats.daily_entity_counts (count_date, entity, cnt)
SELECT CURRENT_DATE, entity, SUM(cnt)
FROM stats.entity_counters_shard
GROUP BY entity
ON CONFLICT (count_date, entity) DO UPDATE SET
  cnt = EXCLUDED.cnt;
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Multi-line area chart — Platform Growth over time
  - **X-axis:** `count_date`
  - **Y-axis (left):** `cnt` for `gk_moves`
  - **Y-axis (right):** `cnt` for `gk_geokrety` and `gk_users`
  - **Data sources:** Three series filtered by entity name

- **Chart type:** Stacked area chart — GK type composition evolution over time
  - **X-axis:** `count_date`
  - **Series:** `gk_geokrety_type_0` through `gk_geokrety_type_10`

```
ASCII Sample (Platform Growth):
Moves (M)  Users (K)  GeoKrety (K)
7.0  |     ●                      |
6.0  |   ●   ●                   |  150 ·   ·
5.0  |  ●        ●               |  120       ·   ·
4.0  | ●            ●            |   90
3.0  |●                ●  ●      |   60
2.0  |                    ●  ●  ●|   30
     +---------------------------+
     2015 2017 2019 2021 2023 2025
```

#### TimescaleDB Assessment

**RECOMMENDED.** Rationale:

- `daily_entity_counts` is a canonical time-series table: rows are inserted once per day per entity, never updated, and accessed primarily via time-range queries (e.g., "last year of data").
- With 25 entities × 365 days/year × 15 years = ~136,875 rows today, growing at ~9,125 rows/year.
- While small now, TimescaleDB hypertable on `count_date` would enable:
  - Automatic chunk-based time pruning for date-range scans
  - Column-based compression (up to 90% storage savings for such repetitive data)
  - Continuous aggregates if sub-daily granularity is added later
- **Recommendation:** Convert to TimescaleDB hypertable if TimescaleDB is installed. Otherwise, standard PostgreSQL PK index is sufficient at current scale.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.4.001 | daily_entity_counts table exists | `has_table('stats', 'daily_entity_counts')` |
| T-2.4.002 | PK is (count_date, entity) | `col_is_pk('stats', 'daily_entity_counts', ARRAY['count_date', 'entity'])` |
| T-2.4.003 | count_date type is date | `col_type_is('stats', 'daily_entity_counts', 'count_date', 'date')` |
| T-2.4.004 | entity type is varchar(32) | `col_type_is('stats', 'daily_entity_counts', 'entity', 'character varying(32)')` |
| T-2.4.005 | cnt default is 0 | `col_default_is('stats', 'daily_entity_counts', 'cnt', '0')` |
| T-2.4.006 | cnt type is bigint | `col_type_is('stats', 'daily_entity_counts', 'cnt', 'bigint')` |
| T-2.4.007 | Insert succeeds | Insert `('2025-06-15', 'gk_moves', 6000000)` and verify |
| T-2.4.008 | Duplicate PK raises error | Insert same `(count_date, entity)` twice — `throws_ok` |
| T-2.4.009 | ON CONFLICT upsert works | Insert then re-insert with different value using ON CONFLICT, verify updated |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310200300_create_daily_entity_counts.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify table exists with 3 columns and correct types
- [ ] 4. Verify composite PK on `(count_date, entity)`
- [ ] 5. Test insert, read-back, and upsert behavior
- [ ] 6. Run pgTAP tests T-2.4.001 through T-2.4.009

---
