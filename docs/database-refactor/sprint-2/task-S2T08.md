---
title: "Task S2T08: Create gk_geokrety Counter Trigger + Attach"
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
task: S2T08
step: 2.8
migration: 20260310200700_create_gk_geokrety_counter_trigger.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026-03-10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.8
  - 2026-03-10: canonicalized trigger name and creation-date semantics
---

# Task S2T08: Create gk_geokrety Counter Trigger + Attach

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.8` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.8: Create `gk_geokrety` Counter Trigger + Attach

**What this step does:** Creates `geokrety.fn_gk_geokrety_counter()` and attaches it as `tr_gk_geokrety_counters` AFTER INSERT OR DELETE on `geokrety.gk_geokrety`. This trigger maintains `stats.entity_counters_shard` for `gk_geokrety` (total) and `gk_geokrety_type_0` through `gk_geokrety_type_10` (per GK type), and increments `stats.daily_activity.gk_created` for the GeoKret creation calendar day. Shard = `id % 16`.

**Migration file name:** `20260310200700_create_gk_geokrety_counter_trigger.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_geokrety_counter()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_shard INT;
  v_gk_type INT;
  v_delta INT;
  v_date DATE;
BEGIN
  IF TG_OP = 'INSERT' THEN
    v_shard := NEW.id % 16;
    v_gk_type := NEW.type;
    v_delta := 1;
    v_date := NEW.created_on_datetime::date;
  ELSE  -- DELETE
    v_shard := OLD.id % 16;
    v_gk_type := OLD.type;
    v_delta := -1;
    v_date := OLD.created_on_datetime::date;
  END IF;

  -- Increment/decrement total geokrety shard
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  VALUES ('gk_geokrety', v_shard, v_delta)
  ON CONFLICT (entity, shard)
  DO UPDATE SET cnt = stats.entity_counters_shard.cnt + v_delta;

  -- Increment/decrement per-type shard (types 0..10 only)
  IF v_gk_type BETWEEN 0 AND 10 THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_geokrety_type_' || v_gk_type::text, v_shard, v_delta)
    ON CONFLICT (entity, shard)
    DO UPDATE SET cnt = stats.entity_counters_shard.cnt + v_delta;
  END IF;

  -- Update daily activity GeoKrety creation count
  INSERT INTO stats.daily_activity (activity_date, gk_created)
  VALUES (v_date, v_delta)
  ON CONFLICT (activity_date) DO UPDATE SET
    gk_created = stats.daily_activity.gk_created + v_delta;

  RETURN NULL;
END;
$$;

CREATE TRIGGER tr_gk_geokrety_counters
  AFTER INSERT OR DELETE ON geokrety.gk_geokrety
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_geokrety_counter();
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkGeokretyCounterTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_geokrety_counter()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_shard INT;
  v_gk_type INT;
  v_delta INT;
  v_date DATE;
BEGIN
  IF TG_OP = 'INSERT' THEN
    v_shard := NEW.id % 16;
    v_gk_type := NEW.type;
    v_delta := 1;
    v_date := NEW.created_on_datetime::date;
  ELSE  -- DELETE
    v_shard := OLD.id % 16;
    v_gk_type := OLD.type;
    v_delta := -1;
    v_date := OLD.created_on_datetime::date;
  END IF;

  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  VALUES ('gk_geokrety', v_shard, v_delta)
  ON CONFLICT (entity, shard)
  DO UPDATE SET cnt = stats.entity_counters_shard.cnt + v_delta;

  IF v_gk_type BETWEEN 0 AND 10 THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_geokrety_type_' || v_gk_type::text, v_shard, v_delta)
    ON CONFLICT (entity, shard)
    DO UPDATE SET cnt = stats.entity_counters_shard.cnt + v_delta;
  END IF;

  INSERT INTO stats.daily_activity (activity_date, gk_created)
  VALUES (v_date, v_delta)
  ON CONFLICT (activity_date) DO UPDATE SET
    gk_created = stats.daily_activity.gk_created + v_delta;

  RETURN NULL;
END;
$$;

CREATE TRIGGER tr_gk_geokrety_counters
  AFTER INSERT OR DELETE ON geokrety.gk_geokrety
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_geokrety_counter();
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TRIGGER IF EXISTS tr_gk_geokrety_counters ON geokrety.gk_geokrety;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_geokrety_counter() CASCADE;');
    }
}
```

#### SQL Usage Examples

```sql
-- Read total GeoKrety count
SELECT SUM(cnt) AS total_geokrety
FROM stats.entity_counters_shard WHERE entity = 'gk_geokrety';

-- Read GeoKrety type distribution
SELECT entity, SUM(cnt) AS total
FROM stats.entity_counters_shard
WHERE entity LIKE 'gk_geokrety_type_%'
GROUP BY entity
ORDER BY entity;

-- Check how many new GeoKrety were created today
SELECT gk_created FROM stats.daily_activity
WHERE activity_date = CURRENT_DATE;
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Donut chart — GeoKrety type distribution
  - **Data source:** `SELECT entity, SUM(cnt) FROM stats.entity_counters_shard WHERE entity LIKE 'gk_geokrety_type_%' GROUP BY entity`

- **Chart type:** Line chart — New GeoKrety created per day
  - **Data source:** `SELECT activity_date, gk_created FROM stats.daily_activity ORDER BY activity_date`

```
ASCII Sample (GK Type Distribution):
Traditional(0) |████████████████████████████| 68%
Book/DVD(1)    |████                        |  9%
Coin(3)        |███                         |  7%
Human(2)       |██                          |  5%
KretyPost(4)   |██                          |  4%
Other          |███                         |  7%
```

#### TimescaleDB Assessment

**NOT applicable.** This step creates a trigger function, not a table.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.8.001 | Function fn_gk_geokrety_counter exists | `has_function('geokrety', 'fn_gk_geokrety_counter', ARRAY[]::text[])` |
| T-2.8.002 | Function returns trigger | `function_returns('geokrety', 'fn_gk_geokrety_counter', ARRAY[]::text[], 'trigger')` |
| T-2.8.003 | Trigger tr_gk_geokrety_counters exists | `has_trigger('geokrety', 'gk_geokrety', 'tr_gk_geokrety_counters')` |
| T-2.8.004 | INSERT increments gk_geokrety | Insert new GK, verify `SUM(cnt) WHERE entity='gk_geokrety'` increases by 1 |
| T-2.8.005 | INSERT increments correct type shard | Insert GK of type 0, verify `SUM(cnt) WHERE entity='gk_geokrety_type_0'` increases |
| T-2.8.006 | INSERT updates daily_activity.gk_created | Insert GK, verify `daily_activity.gk_created` increases by 1 |
| T-2.8.007 | DELETE decrements gk_geokrety | Insert then delete GK, verify counter returns to prior value |
| T-2.8.008 | All 11 type entities (0..10) are updated correctly | Insert one GK of each type 0-10, verify all type counters incremented |

#### Implementation Checklist

- [ ] 1. Verify `stats.entity_counters_shard` and `stats.daily_activity` tables exist (Steps 2.1, 2.2)
- [ ] 2. Create migration file `20260310200700_create_gk_geokrety_counter_trigger.php`
- [ ] 3. Run `phinx migrate`
- [ ] 4. Verify function `geokrety.fn_gk_geokrety_counter` exists
- [ ] 5. Verify trigger `tr_gk_geokrety_counters` exists on `gk_geokrety`
- [ ] 6. Test INSERT → both shard counters and daily_activity.gk_created increment
- [ ] 7. Test DELETE → counters decrement
- [ ] 8. Run pgTAP tests T-2.8.001 through T-2.8.008

---
