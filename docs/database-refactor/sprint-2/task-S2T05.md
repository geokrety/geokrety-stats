---
title: "Task S2T05: Create Previous-Move Trigger Function + Attach"
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
task: S2T05
step: 2.5
migration: 20260310200400_create_previous_move_trigger.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026.03.10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.5
  - 2026.03.10: logged master-spec trigger timing conflict for review
---

# Task S2T05: Create Previous-Move Trigger Function + Attach

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `../00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.5` in `02-sprint-2-counters-daily-activity.md`.

## Resolved Decision

- Canonical live contract for this task is `BEFORE INSERT OR UPDATE OR DELETE` on `geokrety.gk_moves`, per Q-012/Q-019 user decision.
- `INSERT` and `UPDATE` recompute `previous_move_id` and `km_distance` for the current row.
- `DELETE` is accepted by the trigger contract and returns `OLD`; downstream trigger families handle counter/date reversal on delete.
- Guardrail: implementation must avoid self-recursive updates and infinite loops when move edits are replayed.

### Step 2.5: Create Previous-Move Trigger Function + Attach

**What this step does:** Creates `geokrety.fn_set_previous_move_id_and_distance()` and attaches it as `tr_gk_moves_before_prev_move` on `geokrety.gk_moves`. This is the **most critical trigger in the entire system**. It sets `NEW.previous_move_id` and computes `NEW.km_distance` before the row is committed.

**Key design decisions:**

1. **BEFORE INSERT/UPDATE/DELETE** — insert and update recompute the row's previous-move linkage and distance; delete is explicitly covered to keep trigger contracts consistent across move-trigger families.
2. **Location-bearing moves only** — only processes `move_type IN (0, 1, 3, 5)` (DROP, GRAB, SEEN, DIP) with non-NULL position.
3. **Primary fast path via `geokrety.gk_geokrety.last_position`** — probe the GK row first, then fall back to ordered lookup in `geokrety.gk_moves` only if no fast-path candidate exists.
4. **Sort by `(moved_on_datetime DESC, id DESC)`** — the fallback lookup uses canonical ordering when two moves share the same timestamp.
5. **If no previous move found** — both `previous_move_id` and `km_distance` are left NULL (not set to 0).

**Migration file name:** `20260310200400_create_previous_move_trigger.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_set_previous_move_id_and_distance()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_prev_move_id BIGINT;
BEGIN
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;

  IF NEW.move_type NOT IN (0, 1, 3, 5) OR NEW.position IS NULL THEN
    NEW.previous_move_id := NULL;
    NEW.km_distance := NULL;
    RETURN NEW;
  END IF;

  SELECT g.last_position INTO v_prev_move_id
  FROM geokrety.gk_geokrety g
  WHERE g.id = NEW.geokret;

  IF v_prev_move_id IS NULL THEN
    SELECT m.id
      INTO v_prev_move_id
    FROM geokrety.gk_moves m
    WHERE m.geokret = NEW.geokret
      AND m.position IS NOT NULL
      AND m.move_type IN (0, 1, 3, 5)
      AND (m.moved_on_datetime < NEW.moved_on_datetime
        OR (m.moved_on_datetime = NEW.moved_on_datetime AND m.id < NEW.id))
    ORDER BY m.moved_on_datetime DESC, m.id DESC
    LIMIT 1;
  END IF;

  NEW.previous_move_id := v_prev_move_id;

  IF NEW.previous_move_id IS NOT NULL THEN
    NEW.km_distance := (
      SELECT (public.ST_Distance(pm.position, NEW.position) / 1000.0)::NUMERIC(8,3)
      FROM geokrety.gk_moves pm
      WHERE pm.id = NEW.previous_move_id
        AND pm.position IS NOT NULL
    );
  ELSE
    NEW.km_distance := NULL;
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER tr_gk_moves_before_prev_move
  BEFORE INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_set_previous_move_id_and_distance();
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreatePreviousMoveTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_set_previous_move_id_and_distance()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_prev_move_id BIGINT;
BEGIN
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;

  IF NEW.move_type NOT IN (0, 1, 3, 5) OR NEW.position IS NULL THEN
    NEW.previous_move_id := NULL;
    NEW.km_distance := NULL;
    RETURN NEW;
  END IF;

  SELECT g.last_position INTO v_prev_move_id
  FROM geokrety.gk_geokrety g
  WHERE g.id = NEW.geokret;

  IF v_prev_move_id IS NULL THEN
    SELECT m.id
      INTO v_prev_move_id
    FROM geokrety.gk_moves m
    WHERE m.geokret = NEW.geokret
      AND m.position IS NOT NULL
      AND m.move_type IN (0, 1, 3, 5)
      AND (m.moved_on_datetime < NEW.moved_on_datetime
        OR (m.moved_on_datetime = NEW.moved_on_datetime AND m.id < NEW.id))
    ORDER BY m.moved_on_datetime DESC, m.id DESC
    LIMIT 1;
  END IF;

  NEW.previous_move_id := v_prev_move_id;

  IF NEW.previous_move_id IS NOT NULL THEN
    NEW.km_distance := (
      SELECT (public.ST_Distance(pm.position, NEW.position) / 1000.0)::NUMERIC(8,3)
      FROM geokrety.gk_moves pm
      WHERE pm.id = NEW.previous_move_id
        AND pm.position IS NOT NULL
    );
  ELSE
    NEW.km_distance := NULL;
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER tr_gk_moves_before_prev_move
  BEFORE INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_set_previous_move_id_and_distance();
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TRIGGER IF EXISTS tr_gk_moves_before_prev_move ON geokrety.gk_moves;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_set_previous_move_id_and_distance() CASCADE;');
    }
}
```

#### SQL Usage Examples

```sql
-- Verify trigger exists on gk_moves
SELECT tgname, tgtype, tgenabled
FROM pg_trigger t
JOIN pg_class c ON c.oid = t.tgrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'geokrety' AND c.relname = 'gk_moves'
  AND t.tgname = 'tr_gk_moves_before_prev_move';

-- Verify function exists
SELECT proname, pronamespace::regnamespace
FROM pg_proc
WHERE proname = 'fn_set_previous_move_id_and_distance';

-- Check km_distance was set on a recently inserted move
SELECT id, geokret, move_type, previous_move_id, km_distance
FROM geokrety.gk_moves
WHERE km_distance IS NOT NULL
ORDER BY id DESC
LIMIT 10;

-- Find the longest single move ever recorded
SELECT id, geokret, moved_on_datetime, km_distance
FROM geokrety.gk_moves
WHERE km_distance IS NOT NULL
ORDER BY km_distance DESC
LIMIT 5;

-- Verify COMMENT moves have km_distance = NULL (not set by trigger)
SELECT COUNT(*) FROM geokrety.gk_moves
WHERE move_type = 2 AND km_distance IS NOT NULL;
-- Expected: 0
```

#### Graph/Visualization Specification

**Unlocked visualizations:** This trigger does not directly unlock a chart, but it is the prerequisite for ALL distance-based visualizations:

- **KM contributed timeline** (Step 2.2 daily_activity)
- **Country km statistics** (Sprint 3 country_daily_stats.km_contributed)
- **GK total km traveled** (Sprint 5 user/GK stats)
- **Gamification distance rewards** (Sprint 4 points system)

```
ASCII Sample (distance data quality check):
GK #1  move 100→101: Warsaw→Berlin:    573.2 km
GK #1  move 101→102: Berlin→Prague:    280.4 km
GK #1  move 102→103: Prague→Vienna:    291.7 km
GK #42 move 200→201: Paris→London:     340.1 km
       move 201→202: London→Dublin:    462.3 km
       (COMMENT 203: no distance)
```

#### TimescaleDB Assessment

**NOT applicable.** This step creates a trigger function, not a table.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.5.001 | Function fn_set_previous_move_id_and_distance exists | `has_function('geokrety', 'fn_set_previous_move_id_and_distance', ARRAY[]::text[])` |
| T-2.5.002 | Function returns trigger | `function_returns('geokrety', 'fn_set_previous_move_id_and_distance', ARRAY[]::text[], 'trigger')` |
| T-2.5.003 | Trigger tr_gk_moves_before_prev_move exists | `has_trigger('geokrety', 'gk_moves', 'tr_gk_moves_before_prev_move')` |
| T-2.5.004 | DROP with prior DROP sets previous_move_id | Insert DROP then INSERT second DROP for same GK, verify `previous_move_id` is set to first move id |
| T-2.5.005 | DROP with prior DROP computes km_distance > 0 | Second DROP for different position — verify `km_distance > 0` |
| T-2.5.006 | COMMENT move leaves previous_move_id NULL | Insert COMMENT (type 2), verify `previous_move_id IS NULL` |
| T-2.5.007 | ARCHIVE move leaves previous_move_id NULL | Insert ARCHIVE (type 4), verify `previous_move_id IS NULL` |
| T-2.5.008 | Move with NULL position leaves distance NULL | Insert DROP with `position = NULL`, verify `km_distance IS NULL` |
| T-2.5.009 | First move for a GK has NULL previous_move_id | Insert first DROP for new GK, verify `previous_move_id IS NULL` |
| T-2.5.010 | Distance is rounded to 3 decimal places | Insert two closely spaced points, verify `km_distance` scale is 3 |
| T-2.5.011 | UPDATE recomputes previous_move_id and km_distance | Update moved_on_datetime/position for a location-bearing move, verify linkage and distance are recalculated |
| T-2.5.012 | DELETE path returns OLD without recursion | Delete a move and verify trigger execution completes without recursive trigger-depth failures |

#### Implementation Checklist

- [ ] 1. Verify `geokrety.gk_moves.previous_move_id` and `km_distance` columns exist (Sprint 1, Step 1.5)
- [ ] 2. Create migration file `20260310200400_create_previous_move_trigger.php`
- [ ] 3. Run `phinx migrate`
- [ ] 4. Verify function `geokrety.fn_set_previous_move_id_and_distance()` exists
- [ ] 5. Verify trigger `tr_gk_moves_before_prev_move` exists on `gk_moves`
- [ ] 6. Test: INSERT drop after prior drop → `previous_move_id` set, `km_distance > 0`
- [ ] 7. Test: COMMENT/ARCHIVE moves → `previous_move_id` and `km_distance` remain NULL
- [ ] 8. Test: Move with NULL position → `km_distance` remains NULL
- [ ] 9. Test: First move for a GK → `previous_move_id` remains NULL
- [ ] 10. Test: UPDATE recalculates `previous_move_id` and `km_distance`
- [ ] 11. Test: DELETE path executes safely (no infinite trigger loop)
- [ ] 12. Run pgTAP tests T-2.5.001 through T-2.5.012

## Agent Loop Log

- 2026-03-10T17:56:26Z - Loop 1 - `dba`: Accepted `BEFORE INSERT/UPDATE/DELETE` contract and required explicit `TG_OP = 'DELETE'` branch with loop-safety guardrail language.
- 2026-03-10T17:56:26Z - Loop 1 - `critical-thinking`: Confirmed this resolves Q-012/Q-019 contradiction; flagged that successor-row recomputation details remain an implementation hardening topic, not a contract blocker for this task.
- 2026-03-10T17:56:26Z - Loop 1 - `specification`: Updated canonical text, SQL/Phinx trigger event list, tests, and checklist to the agreed contract; consensus reached in loop 1.

---
