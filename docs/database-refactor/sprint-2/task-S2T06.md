---
title: "Task S2T06: Create gk_moves Sharded Counter Trigger + Attach"
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
task: S2T06
step: 2.6
migration: 20260310200500_create_gk_moves_counter_trigger.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026-03-10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.6
  - 2026-03-10: marked missing canonical SQL source as blocking open question
---

# Task S2T06: Create gk_moves Sharded Counter Trigger + Attach

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.6` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.6: Create `gk_moves` Sharded Counter Trigger + Attach

**What this step does:** Creates `geokrety.fn_gk_moves_sharded_counter()` and attaches it as `tr_gk_moves_after_sharded_counters` AFTER INSERT OR UPDATE OR DELETE on `geokrety.gk_moves`. This trigger maintains `stats.entity_counters_shard` for `gk_moves` (total) and `gk_moves_type_0` through `gk_moves_type_5` (per type). INSERT/DELETE adjust totals directly; UPDATE must reverse the old typed contribution and apply the new one when the counted values change.

**Migration file name:** `20260310200500_create_gk_moves_counter_trigger.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_sharded_counter()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  v_old_shard INT;
  v_new_shard INT;
  v_old_type_entity TEXT;
  v_new_type_entity TEXT;
BEGIN
  v_old_shard := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN (OLD.id % 16) ELSE NULL END;
  v_new_shard := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN (NEW.id % 16) ELSE NULL END;
  v_old_type_entity := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN format('gk_moves_type_%s', OLD.move_type) ELSE NULL END;
  v_new_type_entity := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN format('gk_moves_type_%s', NEW.move_type) ELSE NULL END;

  IF TG_OP = 'INSERT' THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_moves', v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_new_type_entity, v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;

    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_moves', v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_old_type_entity, v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    RETURN OLD;
  END IF;

  IF (OLD.id, OLD.move_type) = (NEW.id, NEW.move_type) THEN
    RETURN NEW;
  END IF;

  IF OLD.id <> NEW.id THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_moves', v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_moves', v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;
  END IF;

  IF OLD.id <> NEW.id OR OLD.move_type <> NEW.move_type THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_old_type_entity, v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_new_type_entity, v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;
  END IF;

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS tr_gk_moves_after_sharded_counters ON geokrety.gk_moves;
CREATE TRIGGER tr_gk_moves_after_sharded_counters
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_sharded_counter();
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkMovesCounterTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_sharded_counter()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  v_old_shard INT;
  v_new_shard INT;
  v_old_type_entity TEXT;
  v_new_type_entity TEXT;
BEGIN
  v_old_shard := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN (OLD.id % 16) ELSE NULL END;
  v_new_shard := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN (NEW.id % 16) ELSE NULL END;
  v_old_type_entity := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN format('gk_moves_type_%s', OLD.move_type) ELSE NULL END;
  v_new_type_entity := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN format('gk_moves_type_%s', NEW.move_type) ELSE NULL END;

  IF TG_OP = 'INSERT' THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_moves', v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_new_type_entity, v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;

    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_moves', v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_old_type_entity, v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    RETURN OLD;
  END IF;

  IF (OLD.id, OLD.move_type) = (NEW.id, NEW.move_type) THEN
    RETURN NEW;
  END IF;

  IF OLD.id <> NEW.id THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_moves', v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES ('gk_moves', v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;
  END IF;

  IF OLD.id <> NEW.id OR OLD.move_type <> NEW.move_type THEN
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_old_type_entity, v_old_shard, -1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt - 1;

    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    VALUES (v_new_type_entity, v_new_shard, 1)
    ON CONFLICT (entity, shard) DO UPDATE SET cnt = stats.entity_counters_shard.cnt + 1;
  END IF;

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS tr_gk_moves_after_sharded_counters ON geokrety.gk_moves;
CREATE TRIGGER tr_gk_moves_after_sharded_counters
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_sharded_counter();
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TRIGGER IF EXISTS tr_gk_moves_after_sharded_counters ON geokrety.gk_moves;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_sharded_counter() CASCADE;');
    }
}
```

#### SQL Usage Examples

```sql
-- Verify trigger exists
SELECT tgname FROM pg_trigger t
JOIN pg_class c ON c.oid = t.tgrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'geokrety' AND c.relname = 'gk_moves'
  AND t.tgname = 'tr_gk_moves_after_sharded_counters';

-- Read current total move count
SELECT SUM(cnt) AS total_moves
FROM stats.entity_counters_shard
WHERE entity = 'gk_moves';

-- Read per-type breakdown
SELECT entity, SUM(cnt) AS total
FROM stats.entity_counters_shard
WHERE entity LIKE 'gk_moves_type_%'
GROUP BY entity
ORDER BY entity;

-- Verify shard distribution is uniform after many inserts
SELECT shard, cnt FROM stats.entity_counters_shard
WHERE entity = 'gk_moves'
ORDER BY shard;
```

#### Graph/Visualization Specification

**Unlocked visualizations:** This trigger maintains the real-time total for the KPI Move Counter card. When a move is inserted, the live total updates instantly without scanning `gk_moves`.

```
ASCII Sample (KPI Card data source):
Total Moves: 6,931,442
  └─ shard 0:   433,215
  └─ shard 1:   433,840
  └─ ...
  └─ shard 15:  432,441
     SUM = 6,931,442 ✓

Move Type Donut:
drop:    2,634,748 (38%)
grab:    2,356,690 (34%)
dip:       554,515  (8%)
seen:      554,515  (8%)
comment:   554,515  (8%)
archive:   277,459  (4%)
```

#### TimescaleDB Assessment

**NOT applicable.** This step creates a trigger function, not a table.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.6.001 | Function fn_gk_moves_sharded_counter exists | `has_function('geokrety', 'fn_gk_moves_sharded_counter', ARRAY[]::text[])` |
| T-2.6.002 | Function returns trigger | `function_returns('geokrety', 'fn_gk_moves_sharded_counter', ARRAY[]::text[], 'trigger')` |
| T-2.6.003 | Trigger tr_gk_moves_after_sharded_counters exists | `has_trigger('geokrety', 'gk_moves', 'tr_gk_moves_after_sharded_counters')` |
| T-2.6.004 | INSERT increments gk_moves counter | Insert one DROP, verify `SUM(cnt) WHERE entity='gk_moves'` increases by 1 |
| T-2.6.005 | INSERT increments gk_moves_type_0 counter | Insert DROP (type 0), verify `SUM(cnt) WHERE entity='gk_moves_type_0'` increases by 1 |
| T-2.6.006 | DELETE decrements gk_moves counter | Insert then delete one move, verify counter returns to prior value |
| T-2.6.007 | DELETE decrements correct type counter | Insert DROP then delete it, verify `gk_moves_type_0` returns to prior value |
| T-2.6.008 | Shard selection is id % 16 | Insert move with known id, verify shard row updated is `id % 16` |
| T-2.6.009 | All 6 move types increment correct shard entity | Insert one of each type 0-5, verify each `gk_moves_type_N` incremented |
| T-2.6.010 | UPDATE reverses old type and applies new type | Update `move_type` for one row and verify old type decremented and new type incremented |

#### Implementation Checklist

- [ ] 1. Verify `stats.entity_counters_shard` table exists (Step 2.1)
- [ ] 2. Create migration file `20260310200500_create_gk_moves_counter_trigger.php`
- [ ] 3. Run `phinx migrate`
- [ ] 4. Verify function `geokrety.fn_gk_moves_sharded_counter` exists
- [ ] 5. Verify trigger `tr_gk_moves_after_sharded_counters` exists on `gk_moves`
- [ ] 6. Test: INSERT → counter incremented
- [ ] 7. Test: DELETE → counter decremented
- [ ] 8. Test: Correct shard row updated based on `id % 16`
- [ ] 9. Test: UPDATE reverses old typed contribution and applies new typed contribution
- [ ] 10. Run pgTAP tests T-2.6.001 through T-2.6.010

## Agent Loop Log

- 2026-03-10T17:56:26Z - Loop 1 - `dba`: Recommended explicit reverse-old/apply-new math for UPDATE and shard migration support when `id` changes.
- 2026-03-10T17:56:26Z - Loop 1 - `critical-thinking`: No blocking contradiction; accepted this as canonical replacement for removed insert/delete-only draft.
- 2026-03-10T17:56:26Z - Loop 1 - `specification`: Inserted full SQL DDL and Phinx migration code and updated tests/checklist to concrete UPDATE behavior.

## Resolution

- Blocking SQL/Phinx gap for this task has been resolved in this file; see Q-020 reference update in `99-OPEN-QUESTIONS.md`.

---
