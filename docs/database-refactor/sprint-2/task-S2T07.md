---
title: "Task S2T07: Create gk_moves Daily Activity Trigger + Attach"
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
task: S2T07
step: 2.7
migration: 20260310200600_create_gk_moves_daily_trigger.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026-03-10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.7
---

# Task S2T07: Create gk_moves Daily Activity Trigger + Attach

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.7` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.7: Create `gk_moves` Daily Activity Trigger + Attach

**What this step does:** Creates `geokrety.fn_gk_moves_daily_activity()` and attaches it as `tr_gk_moves_after_daily_activity` AFTER INSERT OR UPDATE OR DELETE on `geokrety.gk_moves`. This trigger maintains two tables: `stats.daily_activity` (move type counters and km_contributed per day) and `stats.daily_active_users` (presence-only per-user per-day activity). On INSERT, it upserts `daily_activity` and records user presence with `ON CONFLICT DO NOTHING`. On UPDATE, it reverses the old daily contribution and applies the new one. On DELETE, it decrements `daily_activity` counters; it does not delete `daily_active_users` rows.

**Migration file name:** `20260310200600_create_gk_moves_daily_trigger.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_daily_activity()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_old_date DATE;
  v_new_date DATE;
BEGIN
  v_old_date := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN OLD.moved_on_datetime::date ELSE NULL END;
  v_new_date := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN NEW.moved_on_datetime::date ELSE NULL END;

  IF TG_OP = 'UPDATE'
     AND OLD.moved_on_datetime IS NOT DISTINCT FROM NEW.moved_on_datetime
     AND OLD.move_type IS NOT DISTINCT FROM NEW.move_type
     AND OLD.km_distance IS NOT DISTINCT FROM NEW.km_distance
     AND OLD.author IS NOT DISTINCT FROM NEW.author THEN
    RETURN NULL;
  END IF;

  IF TG_OP IN ('UPDATE', 'DELETE') THEN
    INSERT INTO stats.daily_activity (
      activity_date, total_moves,
      drops, grabs, comments, sees, archives, dips,
      km_contributed
    ) VALUES (
      v_old_date, -1,
      CASE WHEN OLD.move_type = 0 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 1 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 2 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 3 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 4 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 5 THEN -1 ELSE 0 END,
      COALESCE(OLD.km_distance, 0) * -1
    )
    ON CONFLICT (activity_date) DO UPDATE SET
      total_moves = stats.daily_activity.total_moves - 1,
      drops       = stats.daily_activity.drops    + CASE WHEN OLD.move_type = 0 THEN -1 ELSE 0 END,
      grabs       = stats.daily_activity.grabs    + CASE WHEN OLD.move_type = 1 THEN -1 ELSE 0 END,
      comments    = stats.daily_activity.comments + CASE WHEN OLD.move_type = 2 THEN -1 ELSE 0 END,
      sees        = stats.daily_activity.sees     + CASE WHEN OLD.move_type = 3 THEN -1 ELSE 0 END,
      archives    = stats.daily_activity.archives + CASE WHEN OLD.move_type = 4 THEN -1 ELSE 0 END,
      dips        = stats.daily_activity.dips     + CASE WHEN OLD.move_type = 5 THEN -1 ELSE 0 END,
      km_contributed = stats.daily_activity.km_contributed - COALESCE(OLD.km_distance, 0);
  END IF;

  IF TG_OP IN ('INSERT', 'UPDATE') THEN
    INSERT INTO stats.daily_activity (
      activity_date, total_moves,
      drops, grabs, comments, sees, archives, dips,
      km_contributed
    ) VALUES (
      v_new_date, 1,
      CASE WHEN NEW.move_type = 0 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 1 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 2 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 3 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 4 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 5 THEN 1 ELSE 0 END,
      COALESCE(NEW.km_distance, 0)
    )
    ON CONFLICT (activity_date) DO UPDATE SET
      total_moves = stats.daily_activity.total_moves + 1,
      drops       = stats.daily_activity.drops    + CASE WHEN NEW.move_type = 0 THEN 1 ELSE 0 END,
      grabs       = stats.daily_activity.grabs    + CASE WHEN NEW.move_type = 1 THEN 1 ELSE 0 END,
      comments    = stats.daily_activity.comments + CASE WHEN NEW.move_type = 2 THEN 1 ELSE 0 END,
      sees        = stats.daily_activity.sees     + CASE WHEN NEW.move_type = 3 THEN 1 ELSE 0 END,
      archives    = stats.daily_activity.archives + CASE WHEN NEW.move_type = 4 THEN 1 ELSE 0 END,
      dips        = stats.daily_activity.dips     + CASE WHEN NEW.move_type = 5 THEN 1 ELSE 0 END,
      km_contributed = stats.daily_activity.km_contributed + COALESCE(NEW.km_distance, 0);
  END IF;

  IF TG_OP IN ('INSERT', 'UPDATE') AND NEW.author IS NOT NULL THEN
    INSERT INTO stats.daily_active_users (activity_date, user_id)
    VALUES (v_new_date, NEW.author)
    ON CONFLICT (activity_date, user_id)
    DO NOTHING;
  END IF;

  RETURN NULL;
END;
$$;

CREATE TRIGGER tr_gk_moves_after_daily_activity
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_daily_activity();
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkMovesDailyTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_daily_activity()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_old_date DATE;
  v_new_date DATE;
BEGIN
  v_old_date := CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN OLD.moved_on_datetime::date ELSE NULL END;
  v_new_date := CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN NEW.moved_on_datetime::date ELSE NULL END;

  IF TG_OP = 'UPDATE'
     AND OLD.moved_on_datetime IS NOT DISTINCT FROM NEW.moved_on_datetime
     AND OLD.move_type IS NOT DISTINCT FROM NEW.move_type
     AND OLD.km_distance IS NOT DISTINCT FROM NEW.km_distance
     AND OLD.author IS NOT DISTINCT FROM NEW.author THEN
    RETURN NULL;
  END IF;

  IF TG_OP IN ('UPDATE', 'DELETE') THEN
    INSERT INTO stats.daily_activity (
      activity_date, total_moves,
      drops, grabs, comments, sees, archives, dips,
      km_contributed
    ) VALUES (
      v_old_date, -1,
      CASE WHEN OLD.move_type = 0 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 1 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 2 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 3 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 4 THEN -1 ELSE 0 END,
      CASE WHEN OLD.move_type = 5 THEN -1 ELSE 0 END,
      COALESCE(OLD.km_distance, 0) * -1
    )
    ON CONFLICT (activity_date) DO UPDATE SET
      total_moves = stats.daily_activity.total_moves - 1,
      drops       = stats.daily_activity.drops    + CASE WHEN OLD.move_type = 0 THEN -1 ELSE 0 END,
      grabs       = stats.daily_activity.grabs    + CASE WHEN OLD.move_type = 1 THEN -1 ELSE 0 END,
      comments    = stats.daily_activity.comments + CASE WHEN OLD.move_type = 2 THEN -1 ELSE 0 END,
      sees        = stats.daily_activity.sees     + CASE WHEN OLD.move_type = 3 THEN -1 ELSE 0 END,
      archives    = stats.daily_activity.archives + CASE WHEN OLD.move_type = 4 THEN -1 ELSE 0 END,
      dips        = stats.daily_activity.dips     + CASE WHEN OLD.move_type = 5 THEN -1 ELSE 0 END,
      km_contributed = stats.daily_activity.km_contributed - COALESCE(OLD.km_distance, 0);
  END IF;

  IF TG_OP IN ('INSERT', 'UPDATE') THEN
    INSERT INTO stats.daily_activity (
      activity_date, total_moves,
      drops, grabs, comments, sees, archives, dips,
      km_contributed
    ) VALUES (
      v_new_date, 1,
      CASE WHEN NEW.move_type = 0 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 1 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 2 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 3 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 4 THEN 1 ELSE 0 END,
      CASE WHEN NEW.move_type = 5 THEN 1 ELSE 0 END,
      COALESCE(NEW.km_distance, 0)
    )
    ON CONFLICT (activity_date) DO UPDATE SET
      total_moves = stats.daily_activity.total_moves + 1,
      drops       = stats.daily_activity.drops    + CASE WHEN NEW.move_type = 0 THEN 1 ELSE 0 END,
      grabs       = stats.daily_activity.grabs    + CASE WHEN NEW.move_type = 1 THEN 1 ELSE 0 END,
      comments    = stats.daily_activity.comments + CASE WHEN NEW.move_type = 2 THEN 1 ELSE 0 END,
      sees        = stats.daily_activity.sees     + CASE WHEN NEW.move_type = 3 THEN 1 ELSE 0 END,
      archives    = stats.daily_activity.archives + CASE WHEN NEW.move_type = 4 THEN 1 ELSE 0 END,
      dips        = stats.daily_activity.dips     + CASE WHEN NEW.move_type = 5 THEN 1 ELSE 0 END,
      km_contributed = stats.daily_activity.km_contributed + COALESCE(NEW.km_distance, 0);
  END IF;

  IF TG_OP IN ('INSERT', 'UPDATE') AND NEW.author IS NOT NULL THEN
    INSERT INTO stats.daily_active_users (activity_date, user_id)
    VALUES (v_new_date, NEW.author)
    ON CONFLICT (activity_date, user_id)
    DO NOTHING;
  END IF;

  RETURN NULL;
END;
$$;

CREATE TRIGGER tr_gk_moves_after_daily_activity
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_daily_activity();
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TRIGGER IF EXISTS tr_gk_moves_after_daily_activity ON geokrety.gk_moves;');
        $this->execute('DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_daily_activity() CASCADE;');
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
  AND t.tgname = 'tr_gk_moves_after_daily_activity';

-- Check daily activity was updated for today
SELECT * FROM stats.daily_activity WHERE activity_date = CURRENT_DATE;

-- Verify daily_active_users was updated for today
SELECT COUNT(*) AS active_users_today
FROM stats.daily_active_users
WHERE activity_date = CURRENT_DATE;

-- Confirm km_contributed accumulates correctly after insert
SELECT activity_date, km_contributed
FROM stats.daily_activity
WHERE activity_date = CURRENT_DATE;
```

#### Graph/Visualization Specification

No new visualization unlocked at this step. The trigger populates data for the charts defined in Steps 2.2 (daily_activity stacked area chart) and 2.3 (daily_active_users chart).

#### TimescaleDB Assessment

**NOT applicable.** This step creates a trigger function, not a table.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.7.001 | Function fn_gk_moves_daily_activity exists | `has_function('geokrety', 'fn_gk_moves_daily_activity', ARRAY[]::text[])` |
| T-2.7.002 | Function returns trigger | `function_returns('geokrety', 'fn_gk_moves_daily_activity', ARRAY[]::text[], 'trigger')` |
| T-2.7.003 | Trigger tr_gk_moves_after_daily_activity exists | `has_trigger('geokrety', 'gk_moves', 'tr_gk_moves_after_daily_activity')` |
| T-2.7.004 | INSERT DROP increments drops and total_moves | Insert DROP, verify `daily_activity.drops = 1, total_moves = 1` |
| T-2.7.005 | INSERT GRAB increments grabs | Insert GRAB (type 1), verify `daily_activity.grabs = 1` |
| T-2.7.006 | INSERT updates km_contributed | Insert move with `km_distance = 150.500`, verify `daily_activity.km_contributed = 150.500` |
| T-2.7.007 | INSERT with author updates daily_active_users | Insert move with `author = 42`, verify `daily_active_users` row for `(today, 42)` |
| T-2.7.008 | INSERT with NULL author skips daily_active_users | Insert anonymous move, verify no new row in `daily_active_users` |
| T-2.7.009 | Second move by same user keeps one presence row | Insert two moves by user 42 on same day, verify one `(activity_date, user_id)` row |
| T-2.7.010 | UPDATE moves contribution between days exactly | Update `moved_on_datetime`, verify old day decremented and new day incremented |
| T-2.7.011 | UPDATE changes move-type counters exactly | Update move type, verify old type decremented and new type incremented |
| T-2.7.012 | DELETE decrements total_moves | Insert then delete move, verify `total_moves` returns to prior value |
| T-2.7.013 | DELETE decrements correct type column | Insert DROP then delete, verify `drops` returns to prior value |
| T-2.7.014 | DELETE does not remove daily_active_users row | Insert by user 42 then delete, verify `daily_active_users` row still exists |

#### Implementation Checklist

- [ ] 1. Verify `stats.daily_activity` table exists (Step 2.2)
- [ ] 2. Verify `stats.daily_active_users` table exists (Step 2.3)
- [ ] 3. Create migration file `20260310200600_create_gk_moves_daily_trigger.php`
- [ ] 4. Run `phinx migrate`
- [ ] 5. Verify function `geokrety.fn_gk_moves_daily_activity` exists
- [ ] 6. Verify trigger `tr_gk_moves_after_daily_activity` exists on `gk_moves`
- [ ] 7. Test INSERT DROP → daily_activity.drops incremented
- [ ] 8. Test INSERT with author → daily_active_users row created
- [ ] 9. Test INSERT with NULL author → no daily_active_users row
- [ ] 10. Test UPDATE → daily_activity old/new buckets reconciled exactly
- [ ] 11. Test DELETE → daily_activity counters decremented
- [ ] 12. Run pgTAP tests T-2.7.001 through T-2.7.014

---
