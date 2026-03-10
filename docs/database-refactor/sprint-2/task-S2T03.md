---
title: "Task S2T03: Create stats.daily_active_users Table"
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
task: S2T03
step: 2.3
migration: 20260310200200_create_daily_active_users.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026-03-10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.3
---

# Task S2T03: Create stats.daily_active_users Table

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.3` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.3: Create `stats.daily_active_users` Table

**What this step does:** Creates the `stats.daily_active_users` table that tracks whether a user was active on a given day. A user is considered active when they log at least one move with a non-NULL author. The composite primary key `(activity_date, user_id)` supports efficient per-day user lookups for Daily Active Users (DAU) analytics. This table is populated by the `fn_gk_moves_daily_activity` trigger (Step 2.7) and seeded by `fn_seed_daily_activity` (Step 2.12).

**Migration file name:** `20260310200200_create_daily_active_users.php`

#### Full SQL DDL

```sql
CREATE TABLE stats.daily_active_users (
  activity_date DATE NOT NULL,
  user_id INT NOT NULL,
  PRIMARY KEY (activity_date, user_id)
);

COMMENT ON TABLE stats.daily_active_users IS 'Presence table for users active on a given day; one row per (activity_date, user_id)';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateDailyActiveUsers extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.daily_active_users (
  activity_date DATE NOT NULL,
  user_id INT NOT NULL,
  PRIMARY KEY (activity_date, user_id)
);

COMMENT ON TABLE stats.daily_active_users IS 'Presence table for users active on a given day; one row per (activity_date, user_id)';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.daily_active_users;');
    }
}
```

#### SQL Usage Examples

```sql
-- Daily Active Users (DAU) for the last 30 days
SELECT activity_date, COUNT(*) AS dau
FROM stats.daily_active_users
WHERE activity_date >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY activity_date
ORDER BY activity_date DESC;

-- Monthly Active Users (MAU) for a specific month
SELECT COUNT(DISTINCT user_id) AS mau
FROM stats.daily_active_users
WHERE activity_date BETWEEN '2025-06-01' AND '2025-06-30';

-- Active users today
SELECT user_id
FROM stats.daily_active_users
WHERE activity_date = CURRENT_DATE
LIMIT 10;

-- User activity streak: consecutive days active
SELECT user_id, COUNT(*) AS active_days
FROM stats.daily_active_users
WHERE activity_date >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY user_id
ORDER BY active_days DESC
LIMIT 10;
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Line chart — Daily Active Users over time
  - **X-axis:** `activity_date`
  - **Y-axis:** `COUNT(*)` per day (number of distinct active users)
  - **Data source:** `SELECT activity_date, COUNT(*) AS dau FROM stats.daily_active_users GROUP BY activity_date ORDER BY activity_date`

- **Chart type:** Area chart — DAU vs total moves per day (dual Y-axis)
  - **Data source:** JOIN with `daily_activity` on `activity_date`

```
ASCII Sample (Daily Active Users):
2025-06-15 |██████████████████████████| 1,842 DAU
2025-06-16 |████████████████████      | 1,521 DAU
2025-06-17 |██████████████████████████| 1,934 DAU
2025-06-18 |████████████████          | 1,211 DAU
2025-06-19 |████████████████████████  | 1,788 DAU
2025-06-20 |██████████████████████    | 1,655 DAU
2025-06-21 |████████████              |   892 DAU
                                       (weekend dip)
```

#### TimescaleDB Assessment

**POSSIBLY BENEFICIAL.** Rationale:

- `daily_active_users` grows at `unique_users_per_day × days`. Conservatively, 500–2,000 active users/day × 15 years × 365 days ≈ 2.7M to 11M rows.
- At the upper bound, TimescaleDB hypertable partitioning on `activity_date` would enable chunk-based time-range pruning for common queries like "last 30 days".
- However, at ~2.7M rows, PostgreSQL with the composite PK index handles range queries efficiently.
- **Recommendation:** Deploy as standard table. If row count exceeds 5M or retention policies are needed, convert to hypertable.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.3.001 | daily_active_users table exists | `has_table('stats', 'daily_active_users')` |
| T-2.3.002 | PK is (activity_date, user_id) | `col_is_pk('stats', 'daily_active_users', ARRAY['activity_date', 'user_id'])` |
| T-2.3.003 | activity_date type is date | `col_type_is('stats', 'daily_active_users', 'activity_date', 'date')` |
| T-2.3.004 | user_id type is integer | `col_type_is('stats', 'daily_active_users', 'user_id', 'integer')` |
| T-2.3.005 | table has only canonical presence columns | `columns_are('stats', 'daily_active_users', ARRAY['activity_date', 'user_id'])` |
| T-2.3.006 | Insert succeeds | Insert `('2025-06-15', 42)` and verify read-back |
| T-2.3.007 | Duplicate PK raises error | Insert same `(activity_date, user_id)` twice — `throws_ok` |
| T-2.3.008 | NULL user_id raises error | Insert with `user_id = NULL` — `throws_ok` |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310200200_create_daily_active_users.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify table exists with 2 columns and correct types
- [ ] 4. Verify composite PK on `(activity_date, user_id)`
- [ ] 5. Test insert and duplicate rejection
- [ ] 6. Run pgTAP tests T-2.3.001 through T-2.3.008

---
