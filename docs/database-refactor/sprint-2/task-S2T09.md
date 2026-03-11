---
title: "Task S2T09: Create gk_pictures Counter Trigger + Attach"
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
task: S2T09
step: 2.9
migration: 20260310200800_create_gk_pictures_counter_trigger.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026-03-10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.9
  - 2026-03-10: canonicalized trigger name and logged missing SQL source
---

# Task S2T09: Create gk_pictures Counter Trigger + Attach

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `../00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.9` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.9: Create `gk_pictures` Counter Trigger + Attach

**What this step does:** Creates `geokrety.fn_gk_pictures_counter()` and attaches it as `tr_gk_pictures_after_counter` AFTER INSERT OR UPDATE OR DELETE on `geokrety.gk_pictures`. This trigger maintains `stats.entity_counters_shard` for `gk_pictures` (total) and `gk_pictures_type_0` through `gk_pictures_type_2` (per picture type), and updates `stats.daily_activity` picture columns. Sprint 2 owns shard and daily counters only; Sprint 3 extends the same trigger family contract to `stats.country_daily_stats` once that table exists. Picture types: 0=GEOKRET_AVATAR, 1=GEOKRET_MOVE, 2=USER_AVATAR.

**Migration file name:** `20260310200800_create_gk_pictures_counter_trigger.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_pictures_counter()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  v_old_shard INT;
  v_new_shard INT;
  v_old_date DATE;
  v_new_date DATE;
  v_old_type_entity TEXT;
  v_new_type_entity TEXT;
  v_old_avatar_delta BIGINT := 0;
  v_old_move_delta BIGINT := 0;
  v_old_user_delta BIGINT := 0;
  v_new_avatar_delta BIGINT := 0;
  v_new_move_delta BIGINT := 0;
  v_new_user_delta BIGINT := 0;
BEGIN
  v_old_shard := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN (OLD.id % 16) ELSE NULL END;
  v_new_shard := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN (NEW.id % 16) ELSE NULL END;
  v_old_date := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN OLD.created_on_datetime::date ELSE NULL END;
  v_new_date := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN NEW.created_on_datetime::date ELSE NULL END;
  v_old_type_entity := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN format('gk_pictures_type_%s', OLD.type) ELSE NULL END;
  v_new_type_entity := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN format('gk_pictures_type_%s', NEW.type) ELSE NULL END;

  IF TG_OP IN ('UPDATE', 'DELETE') THEN
    v_old_avatar_delta := CASE WHEN OLD.type = 0 THEN -1 ELSE 0 END;
    v_old_move_delta := CASE WHEN OLD.type = 1 THEN -1 ELSE 0 END;
    v_old_user_delta := CASE WHEN OLD.type = 2 THEN -1 ELSE 0 END;
  END IF;

  IF TG_OP IN ('INSERT', 'UPDATE') THEN
    v_new_avatar_delta := CASE WHEN NEW.type = 0 THEN 1 ELSE 0 END;
    v_new_move_delta := CASE WHEN NEW.type = 1 THEN 1 ELSE 0 END;
    v_new_user_delta := CASE WHEN NEW.type = 2 THEN 1 ELSE 0 END;
  END IF;

  IF TG_OP = 'INSERT' THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_pictures', v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_new_type_entity, v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;

    INSERT INTO stats.daily_activity (
      activity_date,
      pictures_uploaded_total,
      pictures_uploaded_avatar,
      pictures_uploaded_move,
      pictures_uploaded_user
    )
    VALUES (
      v_new_date,
      1,
      v_new_avatar_delta,
      v_new_move_delta,
      v_new_user_delta
    )
    ON CONFLICT (activity_date) DO UPDATE SET
      pictures_uploaded_total = stats.daily_activity.pictures_uploaded_total + 1,
      pictures_uploaded_avatar = stats.daily_activity.pictures_uploaded_avatar + EXCLUDED.pictures_uploaded_avatar,
      pictures_uploaded_move = stats.daily_activity.pictures_uploaded_move + EXCLUDED.pictures_uploaded_move,
      pictures_uploaded_user = stats.daily_activity.pictures_uploaded_user + EXCLUDED.pictures_uploaded_user;

    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_pictures', v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_old_type_entity, v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.daily_activity (
      activity_date,
      pictures_uploaded_total,
      pictures_uploaded_avatar,
      pictures_uploaded_move,
      pictures_uploaded_user
    )
    VALUES (
      v_old_date,
      -1,
      v_old_avatar_delta,
      v_old_move_delta,
      v_old_user_delta
    )
    ON CONFLICT (activity_date) DO UPDATE SET
      pictures_uploaded_total = stats.daily_activity.pictures_uploaded_total - 1,
      pictures_uploaded_avatar = stats.daily_activity.pictures_uploaded_avatar + EXCLUDED.pictures_uploaded_avatar,
      pictures_uploaded_move = stats.daily_activity.pictures_uploaded_move + EXCLUDED.pictures_uploaded_move,
      pictures_uploaded_user = stats.daily_activity.pictures_uploaded_user + EXCLUDED.pictures_uploaded_user;

    RETURN OLD;
  END IF;

  IF (OLD.id, OLD.type, OLD.created_on_datetime::date) = (NEW.id, NEW.type, NEW.created_on_datetime::date) THEN
    RETURN NEW;
  END IF;

  IF OLD.id <> NEW.id THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_pictures', v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_pictures', v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;
  END IF;

  IF OLD.id <> NEW.id OR OLD.type <> NEW.type THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_old_type_entity, v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_new_type_entity, v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;
  END IF;

  INSERT INTO stats.daily_activity (
    activity_date,
    pictures_uploaded_total,
    pictures_uploaded_avatar,
    pictures_uploaded_move,
    pictures_uploaded_user
  )
  VALUES (
    v_old_date,
    -1,
    v_old_avatar_delta,
    v_old_move_delta,
    v_old_user_delta
  )
  ON CONFLICT (activity_date) DO UPDATE SET
    pictures_uploaded_total = stats.daily_activity.pictures_uploaded_total - 1,
    pictures_uploaded_avatar = stats.daily_activity.pictures_uploaded_avatar + EXCLUDED.pictures_uploaded_avatar,
    pictures_uploaded_move = stats.daily_activity.pictures_uploaded_move + EXCLUDED.pictures_uploaded_move,
    pictures_uploaded_user = stats.daily_activity.pictures_uploaded_user + EXCLUDED.pictures_uploaded_user;

  INSERT INTO stats.daily_activity (
    activity_date,
    pictures_uploaded_total,
    pictures_uploaded_avatar,
    pictures_uploaded_move,
    pictures_uploaded_user
  )
  VALUES (
    v_new_date,
    1,
    v_new_avatar_delta,
    v_new_move_delta,
    v_new_user_delta
  )
  ON CONFLICT (activity_date) DO UPDATE SET
    pictures_uploaded_total = stats.daily_activity.pictures_uploaded_total + 1,
    pictures_uploaded_avatar = stats.daily_activity.pictures_uploaded_avatar + EXCLUDED.pictures_uploaded_avatar,
    pictures_uploaded_move = stats.daily_activity.pictures_uploaded_move + EXCLUDED.pictures_uploaded_move,
    pictures_uploaded_user = stats.daily_activity.pictures_uploaded_user + EXCLUDED.pictures_uploaded_user;

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS tr_gk_pictures_after_counter ON geokrety.gk_pictures;
CREATE TRIGGER tr_gk_pictures_after_counter
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_pictures
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_pictures_counter();
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkPicturesCounterTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_pictures_counter()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  v_old_shard INT;
  v_new_shard INT;
  v_old_date DATE;
  v_new_date DATE;
  v_old_type_entity TEXT;
  v_new_type_entity TEXT;
  v_old_avatar_delta BIGINT := 0;
  v_old_move_delta BIGINT := 0;
  v_old_user_delta BIGINT := 0;
  v_new_avatar_delta BIGINT := 0;
  v_new_move_delta BIGINT := 0;
  v_new_user_delta BIGINT := 0;
BEGIN
  v_old_shard := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN (OLD.id % 16) ELSE NULL END;
  v_new_shard := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN (NEW.id % 16) ELSE NULL END;
  v_old_date := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN OLD.created_on_datetime::date ELSE NULL END;
  v_new_date := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN NEW.created_on_datetime::date ELSE NULL END;
  v_old_type_entity := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN format('gk_pictures_type_%s', OLD.type) ELSE NULL END;
  v_new_type_entity := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN format('gk_pictures_type_%s', NEW.type) ELSE NULL END;

  IF TG_OP IN ('UPDATE', 'DELETE') THEN
    v_old_avatar_delta := CASE WHEN OLD.type = 0 THEN -1 ELSE 0 END;
    v_old_move_delta := CASE WHEN OLD.type = 1 THEN -1 ELSE 0 END;
    v_old_user_delta := CASE WHEN OLD.type = 2 THEN -1 ELSE 0 END;
  END IF;

  IF TG_OP IN ('INSERT', 'UPDATE') THEN
    v_new_avatar_delta := CASE WHEN NEW.type = 0 THEN 1 ELSE 0 END;
    v_new_move_delta := CASE WHEN NEW.type = 1 THEN 1 ELSE 0 END;
    v_new_user_delta := CASE WHEN NEW.type = 2 THEN 1 ELSE 0 END;
  END IF;

  IF TG_OP = 'INSERT' THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_pictures', v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_new_type_entity, v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;

    INSERT INTO stats.daily_activity (
      activity_date,
      pictures_uploaded_total,
      pictures_uploaded_avatar,
      pictures_uploaded_move,
      pictures_uploaded_user
    )
    VALUES (
      v_new_date,
      1,
      v_new_avatar_delta,
      v_new_move_delta,
      v_new_user_delta
    )
    ON CONFLICT (activity_date) DO UPDATE SET
      pictures_uploaded_total = stats.daily_activity.pictures_uploaded_total + 1,
      pictures_uploaded_avatar = stats.daily_activity.pictures_uploaded_avatar + EXCLUDED.pictures_uploaded_avatar,
      pictures_uploaded_move = stats.daily_activity.pictures_uploaded_move + EXCLUDED.pictures_uploaded_move,
      pictures_uploaded_user = stats.daily_activity.pictures_uploaded_user + EXCLUDED.pictures_uploaded_user;

    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_pictures', v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_old_type_entity, v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.daily_activity (
      activity_date,
      pictures_uploaded_total,
      pictures_uploaded_avatar,
      pictures_uploaded_move,
      pictures_uploaded_user
    )
    VALUES (
      v_old_date,
      -1,
      v_old_avatar_delta,
      v_old_move_delta,
      v_old_user_delta
    )
    ON CONFLICT (activity_date) DO UPDATE SET
      pictures_uploaded_total = stats.daily_activity.pictures_uploaded_total - 1,
      pictures_uploaded_avatar = stats.daily_activity.pictures_uploaded_avatar + EXCLUDED.pictures_uploaded_avatar,
      pictures_uploaded_move = stats.daily_activity.pictures_uploaded_move + EXCLUDED.pictures_uploaded_move,
      pictures_uploaded_user = stats.daily_activity.pictures_uploaded_user + EXCLUDED.pictures_uploaded_user;

    RETURN OLD;
  END IF;

  IF (OLD.id, OLD.type, OLD.created_on_datetime::date) = (NEW.id, NEW.type, NEW.created_on_datetime::date) THEN
    RETURN NEW;
  END IF;

  IF OLD.id <> NEW.id THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_pictures', v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_pictures', v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;
  END IF;

  IF OLD.id <> NEW.id OR OLD.type <> NEW.type THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_old_type_entity, v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_new_type_entity, v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;
  END IF;

  INSERT INTO stats.daily_activity (
    activity_date,
    pictures_uploaded_total,
    pictures_uploaded_avatar,
    pictures_uploaded_move,
    pictures_uploaded_user
  )
  VALUES (
    v_old_date,
    -1,
    v_old_avatar_delta,
    v_old_move_delta,
    v_old_user_delta
  )
  ON CONFLICT (activity_date) DO UPDATE SET
    pictures_uploaded_total = stats.daily_activity.pictures_uploaded_total - 1,
    pictures_uploaded_avatar = stats.daily_activity.pictures_uploaded_avatar + EXCLUDED.pictures_uploaded_avatar,
    pictures_uploaded_move = stats.daily_activity.pictures_uploaded_move + EXCLUDED.pictures_uploaded_move,
    pictures_uploaded_user = stats.daily_activity.pictures_uploaded_user + EXCLUDED.pictures_uploaded_user;

  INSERT INTO stats.daily_activity (
    activity_date,
    pictures_uploaded_total,
    pictures_uploaded_avatar,
    pictures_uploaded_move,
    pictures_uploaded_user
  )
  VALUES (
    v_new_date,
    1,
    v_new_avatar_delta,
    v_new_move_delta,
    v_new_user_delta
  )
  ON CONFLICT (activity_date) DO UPDATE SET
    pictures_uploaded_total = stats.daily_activity.pictures_uploaded_total + 1,
    pictures_uploaded_avatar = stats.daily_activity.pictures_uploaded_avatar + EXCLUDED.pictures_uploaded_avatar,
    pictures_uploaded_move = stats.daily_activity.pictures_uploaded_move + EXCLUDED.pictures_uploaded_move,
    pictures_uploaded_user = stats.daily_activity.pictures_uploaded_user + EXCLUDED.pictures_uploaded_user;

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS tr_gk_pictures_after_counter ON geokrety.gk_pictures;
CREATE TRIGGER tr_gk_pictures_after_counter
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_pictures
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_pictures_counter();
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TRIGGER IF EXISTS tr_gk_pictures_after_counter ON geokrety.gk_pictures;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_pictures_counter() CASCADE;');
    }
}
```

#### SQL Usage Examples

```sql
-- Read total pictures count
SELECT SUM(cnt) AS total_pictures
FROM stats.entity_counters_shard WHERE entity = 'gk_pictures';

-- Picture type breakdown
SELECT entity, SUM(cnt) AS total
FROM stats.entity_counters_shard
WHERE entity LIKE 'gk_pictures_type_%'
GROUP BY entity
ORDER BY entity;

-- Pictures uploaded today (all types)
SELECT pictures_uploaded_total, pictures_uploaded_avatar,
       pictures_uploaded_move, pictures_uploaded_user
FROM stats.daily_activity
WHERE activity_date = CURRENT_DATE;
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Donut chart — Picture type distribution
  - **Data source:** `SELECT entity, SUM(cnt) FROM stats.entity_counters_shard WHERE entity LIKE 'gk_pictures_type_%' GROUP BY entity`

- **Chart type:** Bar chart — Pictures uploaded per day (stacked by type)
  - **Data source:** `SELECT activity_date, pictures_uploaded_avatar, pictures_uploaded_move, pictures_uploaded_user FROM stats.daily_activity ORDER BY activity_date`

```
ASCII Sample (Picture Type Donut):
GK Avatar (0) |████████████████████| 47%
Move (1)      |████████████████    | 38%
User Avatar(2)|█████               | 15%
```

#### TimescaleDB Assessment

**NOT applicable.** This step creates a trigger function, not a table.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.9.001 | Function fn_gk_pictures_counter exists | `has_function('geokrety', 'fn_gk_pictures_counter', ARRAY[]::text[])` |
| T-2.9.002 | Function returns trigger | `function_returns('geokrety', 'fn_gk_pictures_counter', ARRAY[]::text[], 'trigger')` |
| T-2.9.003 | Trigger tr_gk_pictures_after_counter exists | `has_trigger('geokrety', 'gk_pictures', 'tr_gk_pictures_after_counter')` |
| T-2.9.004 | INSERT type 0 increments gk_pictures and gk_pictures_type_0 | Insert avatar picture, verify both counters incremented |
| T-2.9.005 | INSERT type 1 increments pictures_uploaded_move | Insert move picture, verify `daily_activity.pictures_uploaded_move` incremented |
| T-2.9.006 | INSERT type 2 increments pictures_uploaded_user | Insert user picture, verify `daily_activity.pictures_uploaded_user` incremented |
| T-2.9.007 | INSERT increments pictures_uploaded_total | Insert any picture, verify `daily_activity.pictures_uploaded_total` incremented |
| T-2.9.008 | DELETE decrements counters | Insert then delete picture, verify counters return to prior values |
| T-2.9.009 | UPDATE reclassifies type and daily totals correctly | Update picture type and/or created date, verify old bucket decremented and new bucket incremented |

#### Implementation Checklist

- [ ] 1. Verify `stats.entity_counters_shard` and `stats.daily_activity` tables exist (Steps 2.1, 2.2)
- [ ] 2. Create migration file `20260310200800_create_gk_pictures_counter_trigger.php`
- [ ] 3. Run `phinx migrate`
- [ ] 4. Verify function `geokrety.fn_gk_pictures_counter` exists
- [ ] 5. Verify trigger `tr_gk_pictures_after_counter` exists on `gk_pictures`
- [ ] 6. Test all three picture types (0, 1, 2) update correct columns
- [ ] 7. Test DELETE → counters decrement
- [ ] 8. Test: UPDATE reclassifies type/date contributions exactly
- [ ] 9. Run pgTAP tests T-2.9.001 through T-2.9.009

## Agent Loop Log

- 2026-03-10T18:06:46Z - Loop 1 - `dba`: Recommended reverse-old/apply-new daily rollup math and preserving `id % 16` shard movement only when picture identifiers change.
- 2026-03-10T18:06:46Z - Loop 1 - `critical-thinking`: No blocking contradiction; clarified that Sprint 2 owns shard/daily counters while country attribution remains a Sprint 3 extension.
- 2026-03-10T18:06:46Z - Loop 1 - `specification`: Inserted full SQL and Phinx bodies, normalized trigger naming to `tr_gk_pictures_after_counter`, and updated tests/checklist to concrete UPDATE behavior.

## Resolution

- Blocking SQL/Phinx gap for this task has been resolved in this file; see Q-021 reference update in `99-OPEN-QUESTIONS.md`.

---
