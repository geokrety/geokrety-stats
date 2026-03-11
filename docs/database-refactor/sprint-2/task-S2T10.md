---
title: "Task S2T10: Create gk_users Counter Trigger + Attach"
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
task: S2T10
step: 2.10
migration: 20260310200900_create_gk_users_counter_trigger.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026.03.10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.10
  - 2026.03.10: canonicalized trigger name and registration-date semantics
---

# Task S2T10: Create gk_users Counter Trigger + Attach

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `../00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.10` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.10: Create `gk_users` Counter Trigger + Attach

**What this step does:** Creates `geokrety.fn_gk_users_counter()` and attaches it as `tr_gk_users_activity` AFTER INSERT OR DELETE on `geokrety.gk_users`. This trigger maintains `stats.entity_counters_shard` for `gk_users` (total) and increments `stats.daily_activity.users_registered` for the registration calendar day. Shard = `id % 16`.

**Migration file name:** `20260310200900_create_gk_users_counter_trigger.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_users_counter()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_shard INT;
  v_delta INT;
  v_date DATE;
BEGIN
  IF TG_OP = 'INSERT' THEN
    v_shard := NEW.id % 16;
    v_delta := 1;
    v_date := NEW.joined_on_datetime::date;
  ELSE  -- DELETE
    v_shard := OLD.id % 16;
    v_delta := -1;
    v_date := OLD.joined_on_datetime::date;
  END IF;

  -- Increment/decrement total users shard
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  VALUES ('gk_users', v_shard, v_delta)
  ON CONFLICT (entity, shard)
  DO UPDATE SET cnt = stats.entity_counters_shard.cnt + v_delta;

  -- Update daily activity user registration count
  INSERT INTO stats.daily_activity (activity_date, users_registered)
  VALUES (v_date, v_delta)
  ON CONFLICT (activity_date) DO UPDATE SET
    users_registered = stats.daily_activity.users_registered + v_delta;

  RETURN NULL;
END;
$$;

CREATE TRIGGER tr_gk_users_activity
  AFTER INSERT OR DELETE ON geokrety.gk_users
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_users_counter();
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkUsersCounterTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_users_counter()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_shard INT;
  v_delta INT;
  v_date DATE;
BEGIN
  IF TG_OP = 'INSERT' THEN
    v_shard := NEW.id % 16;
    v_delta := 1;
    v_date := NEW.joined_on_datetime::date;
  ELSE  -- DELETE
    v_shard := OLD.id % 16;
    v_delta := -1;
    v_date := OLD.joined_on_datetime::date;
  END IF;

  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  VALUES ('gk_users', v_shard, v_delta)
  ON CONFLICT (entity, shard)
  DO UPDATE SET cnt = stats.entity_counters_shard.cnt + v_delta;

  INSERT INTO stats.daily_activity (activity_date, users_registered)
  VALUES (v_date, v_delta)
  ON CONFLICT (activity_date) DO UPDATE SET
    users_registered = stats.daily_activity.users_registered + v_delta;

  RETURN NULL;
END;
$$;

CREATE TRIGGER tr_gk_users_activity
  AFTER INSERT OR DELETE ON geokrety.gk_users
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_users_counter();
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TRIGGER IF EXISTS tr_gk_users_activity ON geokrety.gk_users;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_users_counter() CASCADE;');
    }
}
```

#### SQL Usage Examples

```sql
-- Read total users count
SELECT SUM(cnt) AS total_users
FROM stats.entity_counters_shard WHERE entity = 'gk_users';

-- New user registrations per day in the last 30 days
SELECT activity_date, users_registered
FROM stats.daily_activity
WHERE users_registered > 0
  AND activity_date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY activity_date DESC;

-- Peak registration day ever
SELECT activity_date, users_registered
FROM stats.daily_activity
ORDER BY users_registered DESC
LIMIT 5;
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Line chart — User registrations per day (new user growth rate)
  - **Data source:** `SELECT activity_date, users_registered FROM stats.daily_activity WHERE users_registered > 0 ORDER BY activity_date`

- **KPI card:** Total registered users
  - **Data source:** `SELECT SUM(cnt) FROM stats.entity_counters_shard WHERE entity = 'gk_users'`

```
ASCII Sample (User Registration Timeline):
2025-06 |█████████████████████  | 1,243 registrations
2025-05 |██████████████████████ | 1,312 registrations
2025-04 |███████████████████    |   988 registrations
2025-03 |████████████████       |   810 registrations
2025-02 |████████████           |   624 registrations
```

#### TimescaleDB Assessment

**NOT applicable.** This step creates a trigger function, not a table.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.10.001 | Function fn_gk_users_counter exists | `has_function('geokrety', 'fn_gk_users_counter', ARRAY[]::text[])` |
| T-2.10.002 | Function returns trigger | `function_returns('geokrety', 'fn_gk_users_counter', ARRAY[]::text[], 'trigger')` |
| T-2.10.003 | Trigger tr_gk_users_activity exists | `has_trigger('geokrety', 'gk_users', 'tr_gk_users_activity')` |
| T-2.10.004 | INSERT increments gk_users counter | Insert new user, verify `SUM(cnt) WHERE entity='gk_users'` increases by 1 |
| T-2.10.005 | INSERT updates daily_activity.users_registered | Insert user, verify `daily_activity.users_registered` increases by 1 |
| T-2.10.006 | DELETE decrements gk_users counter | Insert then delete user, verify counter returns to prior value |
| T-2.10.007 | Shard selection is id % 16 | Insert user with known id, verify correct shard row updated |

#### Implementation Checklist

- [ ] 1. Verify `stats.entity_counters_shard` and `stats.daily_activity` tables exist (Steps 2.1, 2.2)
- [ ] 2. Create migration file `20260310200900_create_gk_users_counter_trigger.php`
- [ ] 3. Run `phinx migrate`
- [ ] 4. Verify function `geokrety.fn_gk_users_counter` exists
- [ ] 5. Verify trigger `tr_gk_users_activity` exists on `gk_users`
- [ ] 6. Test INSERT → counter and daily_activity.users_registered incremented
- [ ] 7. Test DELETE → counters decremented
- [ ] 8. Run pgTAP tests T-2.10.001 through T-2.10.007

---
