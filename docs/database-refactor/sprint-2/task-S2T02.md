---
title: "Task S2T02: Create stats.daily_activity Table"
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
task: S2T02
step: 2.2
migration: 20260310200100_create_daily_activity.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026.03.10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.2
---

# Task S2T02: Create stats.daily_activity Table

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `../00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.2` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.2: Create `stats.daily_activity` Table

**What this step does:** Creates the `stats.daily_activity` table that stores per-calendar-day aggregate activity metrics. This is the primary source for the global activity chart, the KM contributed timeline, and dashboard KPI per-day breakdowns. The `activity_date` primary key means at most one row per day. Columns `points_contributed`, `pictures_*`, `loves_count`, and `users_registered` are NOT populated by the gk_moves trigger in this step — they are updated by their respective triggers in Steps 2.8 (gk_geokrety), 2.9 (gk_pictures), and 2.10 (gk_users), and by the points-awarder service in Sprint 4.

**Migration file name:** `20260310200100_create_daily_activity.php`

#### Full SQL DDL

```sql
CREATE TABLE stats.daily_activity (
  activity_date DATE PRIMARY KEY,
  total_moves BIGINT NOT NULL DEFAULT 0,
  drops BIGINT NOT NULL DEFAULT 0,
  grabs BIGINT NOT NULL DEFAULT 0,
  comments BIGINT NOT NULL DEFAULT 0,
  sees BIGINT NOT NULL DEFAULT 0,
  archives BIGINT NOT NULL DEFAULT 0,
  dips BIGINT NOT NULL DEFAULT 0,
  km_contributed NUMERIC(14,3) NOT NULL DEFAULT 0,
  points_contributed NUMERIC(16,4) NOT NULL DEFAULT 0,
  gk_created BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_total BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_avatar BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_move BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_user BIGINT NOT NULL DEFAULT 0,
  loves_count BIGINT NOT NULL DEFAULT 0,
  users_registered BIGINT NOT NULL DEFAULT 0
);

COMMENT ON TABLE stats.daily_activity IS 'Per-calendar-day aggregate activity metrics; one row per day';
COMMENT ON COLUMN stats.daily_activity.points_contributed IS 'Total gamification points awarded on this date; updated by points-awarder service (Sprint 4)';
COMMENT ON COLUMN stats.daily_activity.gk_created IS 'New GeoKrety created on this date; updated by gk_geokrety trigger (Step 2.8)';
COMMENT ON COLUMN stats.daily_activity.pictures_uploaded_total IS 'Total pictures uploaded; updated by gk_pictures trigger (Step 2.9)';
COMMENT ON COLUMN stats.daily_activity.loves_count IS 'Loves given on this date; updated by loves trigger (Sprint 5)';
COMMENT ON COLUMN stats.daily_activity.users_registered IS 'New user registrations; updated by gk_users trigger (Step 2.10)';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateDailyActivity extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.daily_activity (
  activity_date DATE PRIMARY KEY,
  total_moves BIGINT NOT NULL DEFAULT 0,
  drops BIGINT NOT NULL DEFAULT 0,
  grabs BIGINT NOT NULL DEFAULT 0,
  comments BIGINT NOT NULL DEFAULT 0,
  sees BIGINT NOT NULL DEFAULT 0,
  archives BIGINT NOT NULL DEFAULT 0,
  dips BIGINT NOT NULL DEFAULT 0,
  km_contributed NUMERIC(14,3) NOT NULL DEFAULT 0,
  points_contributed NUMERIC(16,4) NOT NULL DEFAULT 0,
  gk_created BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_total BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_avatar BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_move BIGINT NOT NULL DEFAULT 0,
  pictures_uploaded_user BIGINT NOT NULL DEFAULT 0,
  loves_count BIGINT NOT NULL DEFAULT 0,
  users_registered BIGINT NOT NULL DEFAULT 0
);

COMMENT ON TABLE stats.daily_activity IS 'Per-calendar-day aggregate activity metrics; one row per day';
COMMENT ON COLUMN stats.daily_activity.points_contributed IS 'Total gamification points awarded on this date; updated by points-awarder service (Sprint 4)';
COMMENT ON COLUMN stats.daily_activity.gk_created IS 'New GeoKrety created on this date; updated by gk_geokrety trigger (Step 2.8)';
COMMENT ON COLUMN stats.daily_activity.pictures_uploaded_total IS 'Total pictures uploaded; updated by gk_pictures trigger (Step 2.9)';
COMMENT ON COLUMN stats.daily_activity.loves_count IS 'Loves given on this date; updated by loves trigger (Sprint 5)';
COMMENT ON COLUMN stats.daily_activity.users_registered IS 'New user registrations; updated by gk_users trigger (Step 2.10)';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.daily_activity;');
    }
}
```

#### SQL Usage Examples

```sql
-- Last 30 days of global activity
SELECT activity_date, total_moves, drops, grabs, sees, dips, km_contributed
FROM stats.daily_activity
WHERE activity_date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY activity_date DESC;

-- Weekly totals for the current year
SELECT
  date_trunc('week', activity_date)::date AS week_start,
  SUM(total_moves) AS moves,
  SUM(km_contributed) AS km
FROM stats.daily_activity
WHERE activity_date >= date_trunc('year', CURRENT_DATE)
GROUP BY week_start
ORDER BY week_start;

-- Move type breakdown for a given month
SELECT
  SUM(drops) AS drops,
  SUM(grabs) AS grabs,
  SUM(sees) AS sees,
  SUM(dips) AS dips,
  SUM(comments) AS comments,
  SUM(archives) AS archives,
  SUM(km_contributed) AS total_km
FROM stats.daily_activity
WHERE activity_date BETWEEN '2025-06-01' AND '2025-06-30';

-- Most active day ever
SELECT activity_date, total_moves
FROM stats.daily_activity
ORDER BY total_moves DESC
LIMIT 1;
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Stacked area chart — Global daily activity by move type
  - **X-axis:** `activity_date`
  - **Y-axis:** `drops`, `grabs`, `sees`, `dips`, `comments`, `archives` (stacked)
  - **Colors:** Distinct per move type (drop=green, grab=blue, seen=teal, dip=yellow, comment=gray, archive=red)

- **Chart type:** Line chart — Daily km contributed with 7-day moving average overlay
  - **X-axis:** `activity_date`
  - **Primary line:** `km_contributed` per day
  - **Secondary line:** `AVG(km_contributed) OVER (ORDER BY activity_date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW)`

```
ASCII Sample (Stacked Daily Moves, last 7 days):
2025-06-15 |drop████grab████dip██see█| 1245 moves, 8,432 km
2025-06-16 |drop████████grab███dip█  | 987 moves,  6,210 km
2025-06-17 |drop███grab████dip████   | 1102 moves, 9,140 km
2025-06-18 |drop█████grab████        | 832 moves,  5,432 km
2025-06-19 |drop████████grab█████dip | 1389 moves, 11,230 km
2025-06-20 |drop████grab███dip██see█ | 943 moves,  7,891 km
2025-06-21 |drop███grab██            | 611 moves,  3,210 km
```

#### TimescaleDB Assessment

**RECOMMENDED (conditional).** Rationale:

- `daily_activity` is an append-heavy time-series table with one row per calendar day. With 15+ years of historical data, it holds ~5,500 rows and grows at ~365 rows/year.
- At this tiny scale, standard PostgreSQL with PK-based access is entirely sufficient and TimescaleDB adds no benefit.
- **However**, if the table is extended with sub-daily granularity in the future (hourly buckets), TimescaleDB hypertable conversion would become highly valuable for compression and time-range pruning.
- **Recommendation:** Use standard PostgreSQL table now. The PK `activity_date` already provides excellent access patterns.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.2.001 | daily_activity table exists | `has_table('stats', 'daily_activity')` |
| T-2.2.002 | PK is activity_date | `col_is_pk('stats', 'daily_activity', 'activity_date')` |
| T-2.2.003 | activity_date type is date | `col_type_is('stats', 'daily_activity', 'activity_date', 'date')` |
| T-2.2.004 | total_moves default is 0 | `col_default_is('stats', 'daily_activity', 'total_moves', '0')` |
| T-2.2.005 | km_contributed type is numeric(14,3) | `col_type_is('stats', 'daily_activity', 'km_contributed', 'numeric(14,3)')` |
| T-2.2.006 | points_contributed type is numeric(16,4) | `col_type_is('stats', 'daily_activity', 'points_contributed', 'numeric(16,4)')` |
| T-2.2.007 | 17 columns exist | `SELECT is(COUNT(*)::INT, 17) FROM information_schema.columns WHERE table_schema = 'stats' AND table_name = 'daily_activity'` |
| T-2.2.008 | Insert succeeds | Insert `('2025-06-15', 100, ...)` and verify read-back |
| T-2.2.009 | Duplicate PK raises error | Insert same `activity_date` twice — `throws_ok` |
| T-2.2.010 | Users_registered column exists and is bigint | `col_type_is('stats', 'daily_activity', 'users_registered', 'bigint')` |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310200100_create_daily_activity.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify table exists with 17 columns and correct types
- [ ] 4. Verify PK on `activity_date`
- [ ] 5. Test insert and read-back
- [ ] 6. Run pgTAP tests T-2.2.001 through T-2.2.010

---
