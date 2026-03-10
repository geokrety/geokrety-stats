---
title: "Task S2T11: Create Entity Counter Snapshot Function"
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
task: S2T11
step: 2.11
migration: 20260310201000_create_entity_counter_snapshot.php
blocks: [3, 4, 5, 6]
changelog:
  - 2026-03-10: created by merge of 02-sprint-2-counters-daily-activity.md step 2.11
  - 2026-03-10: logged shard-distribution ambiguity for snapshot seeding
---

# Task S2T11: Create Entity Counter Snapshot Function

## Master-Spec Alignment

The normative contract for this sprint is `00-SPRINT-INDEX.md` plus the canonical Sprint 2 task set in `docs/database-refactor/sprint-2/`. `00-SPEC-DRAFT-v1.obsolete.md` is legacy context only and is not authoritative.

- Canonical previous-move function name: `geokrety.fn_set_previous_move_id_and_distance()`.
- Canonical `stats.daily_activity` column name: `gk_created`, not `new_geokrety_count`.
- Canonical `stats.daily_entity_counts` column name: `cnt`, not `cumulative_count`.
- Canonical `stats.daily_active_users` contract is presence-only: `(activity_date, user_id)` with no per-user `move_count` column.
- The previous-move function must probe `geokrety.gk_geokrety.last_position` first, then fall back to ordered lookup in `geokrety.gk_moves`.
- Any lower sections that still use legacy names such as `fn_gk_moves_set_previous_move`, `new_geokrety_count`, `cumulative_count`, or `distance_km` are obsolete draft text and are superseded by this alignment block.

## Source

- Generated from sprint document step `2.11` in `02-sprint-2-counters-daily-activity.md`.

### Step 2.11: Create Entity Counter Snapshot Function

**What this step does:** Creates `stats.fn_snapshot_entity_counters()` — an idempotent function that seeds `stats.entity_counters_shard` from current source table counts. This function is run once during the Sprint 6 historical backfill, after all historical moves, GKs, pictures, users, and loves have been loaded but before the live counters start. It is safe to re-run at any time: it deletes and re-seeds all shard rows. Canonical contract: the exact aggregate `SUM(cnt)` per entity matters; the snapshot does not need to preserve the live `id % 16` distribution, so it may concentrate totals into shard 0 and leave the remaining shards at 0.

**Migration file name:** `20260310201000_create_entity_counter_snapshot.php`

#### Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION stats.fn_snapshot_entity_counters()
RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE
  v_total BIGINT;
  v_move_type INT;
  v_gk_type INT;
  v_pic_type INT;
BEGIN
  -- ============================================================
  -- gk_moves total
  -- ============================================================
  SELECT COUNT(*) INTO v_total FROM geokrety.gk_moves;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_moves';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_moves', s, CASE WHEN s = 0 THEN v_total ELSE 0 END
  FROM generate_series(0, 15) s;

  -- gk_moves per move type (0..5)
  FOR v_move_type IN 0..5 LOOP
    SELECT COUNT(*) INTO v_total FROM geokrety.gk_moves WHERE move_type = v_move_type;
    DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_moves_type_' || v_move_type;
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    SELECT 'gk_moves_type_' || v_move_type, s, CASE WHEN s = 0 THEN v_total ELSE 0 END
    FROM generate_series(0, 15) s;
  END LOOP;

  -- ============================================================
  -- gk_geokrety total
  -- ============================================================
  SELECT COUNT(*) INTO v_total FROM geokrety.gk_geokrety;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_geokrety';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_geokrety', s, CASE WHEN s = 0 THEN v_total ELSE 0 END
  FROM generate_series(0, 15) s;

  -- gk_geokrety per type (0..10)
  FOR v_gk_type IN 0..10 LOOP
    SELECT COUNT(*) INTO v_total FROM geokrety.gk_geokrety WHERE type = v_gk_type;
    DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_geokrety_type_' || v_gk_type;
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    SELECT 'gk_geokrety_type_' || v_gk_type, s, CASE WHEN s = 0 THEN v_total ELSE 0 END
    FROM generate_series(0, 15) s;
  END LOOP;

  -- ============================================================
  -- gk_pictures total
  -- ============================================================
  SELECT COUNT(*) INTO v_total FROM geokrety.gk_pictures;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_pictures';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_pictures', s, CASE WHEN s = 0 THEN v_total ELSE 0 END
  FROM generate_series(0, 15) s;

  -- gk_pictures per type (0..2)
  FOR v_pic_type IN 0..2 LOOP
    SELECT COUNT(*) INTO v_total FROM geokrety.gk_pictures WHERE type = v_pic_type;
    DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_pictures_type_' || v_pic_type;
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    SELECT 'gk_pictures_type_' || v_pic_type, s, CASE WHEN s = 0 THEN v_total ELSE 0 END
    FROM generate_series(0, 15) s;
  END LOOP;

  -- ============================================================
  -- gk_users total
  -- ============================================================
  SELECT COUNT(*) INTO v_total FROM geokrety.gk_users;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_users';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_users', s, CASE WHEN s = 0 THEN v_total ELSE 0 END
  FROM generate_series(0, 15) s;

  -- ============================================================
  -- gk_loves total
  -- ============================================================
  SELECT COUNT(*) INTO v_total FROM geokrety.gk_loves;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_loves';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_loves', s, CASE WHEN s = 0 THEN v_total ELSE 0 END
  FROM generate_series(0, 15) s;

  RAISE NOTICE 'Entity counter snapshot completed — all 25 entities refreshed';
END;
$$;

COMMENT ON FUNCTION stats.fn_snapshot_entity_counters() IS 'Seeds entity_counters_shard from current source table counts; idempotent; run once during Sprint 6 backfill';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateEntityCounterSnapshot extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION stats.fn_snapshot_entity_counters()
RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE
  v_total BIGINT;
  v_move_type INT;
  v_gk_type INT;
  v_pic_type INT;
BEGIN
  SELECT COUNT(*) INTO v_total FROM geokrety.gk_moves;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_moves';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_moves', s, CASE WHEN s = 0 THEN v_total ELSE 0 END FROM generate_series(0, 15) s;

  FOR v_move_type IN 0..5 LOOP
    SELECT COUNT(*) INTO v_total FROM geokrety.gk_moves WHERE move_type = v_move_type;
    DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_moves_type_' || v_move_type;
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    SELECT 'gk_moves_type_' || v_move_type, s, CASE WHEN s = 0 THEN v_total ELSE 0 END FROM generate_series(0, 15) s;
  END LOOP;

  SELECT COUNT(*) INTO v_total FROM geokrety.gk_geokrety;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_geokrety';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_geokrety', s, CASE WHEN s = 0 THEN v_total ELSE 0 END FROM generate_series(0, 15) s;

  FOR v_gk_type IN 0..10 LOOP
    SELECT COUNT(*) INTO v_total FROM geokrety.gk_geokrety WHERE type = v_gk_type;
    DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_geokrety_type_' || v_gk_type;
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    SELECT 'gk_geokrety_type_' || v_gk_type, s, CASE WHEN s = 0 THEN v_total ELSE 0 END FROM generate_series(0, 15) s;
  END LOOP;

  SELECT COUNT(*) INTO v_total FROM geokrety.gk_pictures;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_pictures';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_pictures', s, CASE WHEN s = 0 THEN v_total ELSE 0 END FROM generate_series(0, 15) s;

  FOR v_pic_type IN 0..2 LOOP
    SELECT COUNT(*) INTO v_total FROM geokrety.gk_pictures WHERE type = v_pic_type;
    DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_pictures_type_' || v_pic_type;
    INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
    SELECT 'gk_pictures_type_' || v_pic_type, s, CASE WHEN s = 0 THEN v_total ELSE 0 END FROM generate_series(0, 15) s;
  END LOOP;

  SELECT COUNT(*) INTO v_total FROM geokrety.gk_users;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_users';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_users', s, CASE WHEN s = 0 THEN v_total ELSE 0 END FROM generate_series(0, 15) s;

  SELECT COUNT(*) INTO v_total FROM geokrety.gk_loves;
  DELETE FROM stats.entity_counters_shard WHERE entity = 'gk_loves';
  INSERT INTO stats.entity_counters_shard (entity, shard, cnt)
  SELECT 'gk_loves', s, CASE WHEN s = 0 THEN v_total ELSE 0 END FROM generate_series(0, 15) s;

  RAISE NOTICE 'Entity counter snapshot completed — all 25 entities refreshed';
END;
$$;

COMMENT ON FUNCTION stats.fn_snapshot_entity_counters() IS 'Seeds entity_counters_shard from current source table counts; idempotent; run once during Sprint 6 backfill';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP FUNCTION IF EXISTS stats.fn_snapshot_entity_counters();');
    }
}
```

#### SQL Usage Examples

```sql
-- Run full snapshot (typically called once during backfill)
SELECT stats.fn_snapshot_entity_counters();

-- Verify total moves matches source table after snapshot
SELECT SUM(cnt) AS shard_total,
       (SELECT COUNT(*) FROM geokrety.gk_moves) AS source_total
FROM stats.entity_counters_shard
WHERE entity = 'gk_moves';
-- Both columns should be equal

-- Verify all 25 entities were refreshed
SELECT COUNT(DISTINCT entity) AS entity_count
FROM stats.entity_counters_shard;
-- Expected: 25

-- Spot-check type totals
SELECT entity, SUM(cnt) AS total
FROM stats.entity_counters_shard
WHERE entity IN ('gk_geokrety_type_0', 'gk_geokrety_type_3')
GROUP BY entity;
```

#### Graph/Visualization Specification

No new visualization unlocked at this step. This function populates data for all counter-based KPI cards defined in Step 2.1.

#### TimescaleDB Assessment

**NOT applicable.** This step creates a utility function, not a table.

#### pgTAP Unit Tests

| Test ID | Test Name | Assertion |
| --- | --- | --- |
| T-2.11.001 | Function fn_snapshot_entity_counters exists | `has_function('stats', 'fn_snapshot_entity_counters', ARRAY[]::text[])` |
| T-2.11.002 | Function returns void | `function_returns('stats', 'fn_snapshot_entity_counters', ARRAY[]::text[], 'void')` |
| T-2.11.003 | Function executes without error | `lives_ok($$ SELECT stats.fn_snapshot_entity_counters(); $$)` |
| T-2.11.004 | 25 entities present after snapshot | `SELECT is(COUNT(DISTINCT entity)::INT, 25) FROM stats.entity_counters_shard` |
| T-2.11.005 | 400 rows present after snapshot (25 × 16) | `SELECT is(COUNT(*)::INT, 400) FROM stats.entity_counters_shard` |
| T-2.11.006 | gk_moves total matches source | `SELECT is(SUM(cnt), (SELECT COUNT(*) FROM geokrety.gk_moves)) FROM stats.entity_counters_shard WHERE entity = 'gk_moves'` |
| T-2.11.007 | Function is idempotent | Run twice and verify counts identical |
| T-2.11.008 | gk_geokrety total matches source | `SELECT is(SUM(cnt), (SELECT COUNT(*) FROM geokrety.gk_geokrety)) FROM stats.entity_counters_shard WHERE entity = 'gk_geokrety'` |
| T-2.11.009 | gk_users total matches source | `SELECT is(SUM(cnt), (SELECT COUNT(*) FROM geokrety.gk_users)) FROM stats.entity_counters_shard WHERE entity = 'gk_users'` |
| T-2.11.010 | RAISE NOTICE is emitted | Capture NOTICE output and verify 'Entity counter snapshot completed' appears |
| T-2.11.011 | Snapshot may concentrate totals in shard 0 | Verify aggregate totals match source even if shards 1..15 remain 0 after reseed |

#### Implementation Checklist

- [ ] 1. Verify `stats.entity_counters_shard` table exists (Step 2.1)
- [ ] 2. Create migration file `20260310201000_create_entity_counter_snapshot.php`
- [ ] 3. Run `phinx migrate`
- [ ] 4. Verify function `stats.fn_snapshot_entity_counters` exists
- [ ] 5. Test function executes without error
- [ ] 6. Verify entity count totals match source tables after execution
- [ ] 7. Test idempotency: run twice, verify results identical
- [ ] 8. Verify aggregate correctness is accepted even when shard distribution is not preserved
- [ ] 9. Run pgTAP tests T-2.11.001 through T-2.11.011

## Agent Loop Log

- 2026-03-10T18:09:40Z - Loop 1 - `dba`: Accepted shard-0 concentration for snapshot runs because backfill correctness depends on aggregate totals, not preservation of live modulo distribution.
- 2026-03-10T18:09:40Z - Loop 1 - `critical-thinking`: No blocking objection; clarified that reseed simplicity and idempotency outweigh synthetic shard-balance fidelity during backfill.
- 2026-03-10T18:09:40Z - Loop 1 - `specification`: Removed the conflict/open-question wording, made aggregate-total correctness canonical, and updated tests/checklist accordingly.

## Resolution

- Snapshot seeding contract is canonicalized: total sums across shards are authoritative; preserving historical shard placement is not required. See Q-022 reference update in `99-OPEN-QUESTIONS.md`.

---
