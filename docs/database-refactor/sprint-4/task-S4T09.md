---
title: "Task S4T09: Social Relation Trigger"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 4
tags:
  - database
  - database-refactor
  - dba
  - function
  - gk-related-users
  - social
  - specification
  - sprint-4
  - sql
  - stats
  - task-index
  - task-merge
  - trigger
  - uc2
  - uc3
  - user-related-users
depends_on:
  - S4T06
  - S4T07
task: S4T09
step: 4.9
migration: 20260310400800_create_relation_trigger.php
blocks:
  - S4T10
  - S4T11
changelog:
  - 2026.03.10: created by merge of task-S4T09.dba.md and task-S4T09.specification.md
  - 2026.03.10: logged the stale relation-trigger contract conflict in the merged SQL
---

# Task S4T09: Social Relation Trigger

## Sprint Context

- Sprint index: Sprint 4 Task Index
- Tags: database, database-refactor, sprint-4, task-index

## Source

- DBA source: `task-S4T09.dba.md`
- Specification source: `task-S4T09.specification.md`

## Resolved Decision

- The canonical relation trigger is `tr_gk_moves_after_relations` on `geokrety.gk_moves`, firing `AFTER INSERT OR UPDATE OR DELETE`.
- Canonical relation tables are `stats.gk_related_users(geokrety_id, user_id, interaction_count, first_interaction, last_interaction)` and `stats.user_related_users(user_id, related_user_id, shared_geokrety_count, first_seen_at, last_seen_at)`.
- Exact reconciliation is required: touched GeoKret rows and affected user-pair rows are recomputed from current qualifying `geokrety.gk_moves` state instead of applying fragile per-move decrements.

## Purpose & Scope

Creates the trigger function `geokrety.fn_gk_moves_relations()` and attaches it as trigger `tr_gk_moves_after_relations` on `geokrety.gk_moves` (AFTER INSERT OR UPDATE OR DELETE). This trigger maintains two social-graph tables:

- `stats.gk_related_users` — which users have touched which GeoKrets
- `stats.user_related_users` — which user pairs have co-touched the same GeoKret

**Qualifying moves:** Only `move_type IN (0, 1, 3, 5)` — DROP, GRAB, SEEN, DIP. COMMENT (2) and ARCHIVE (4) do NOT establish social relations.

**Skip conditions:**

- `NEW.author IS NULL` (anonymous): skip entirely — no user to record
- `move_type NOT IN (0,1,3,5)`: skip entirely

**Reconciliation flow:**

- Recompute `stats.gk_related_users` exactly for the touched `geokrety_id` values
- Recompute `stats.user_related_users` exactly for affected users so `shared_geokrety_count` remains distinct-shared-GK based

When a qualifying move (`move_type IN (0,1,3,5)`) is logged by an authenticated user on `geokrety.gk_moves`, this trigger builds and maintains the social graph between users who co-touched the same GeoKret.

It writes to two tables:

- `stats.gk_related_users` — (GeoKret, user) pair with touch counts
- `stats.user_related_users` — (user A, user B) pair when both touched the same GeoKret

**Scope:** Trigger function + attachment. Requires S4T06 (`gk_related_users`) and S4T07 (`user_related_users`).

---

## Requirements

| ID      | Description                                                                                                                                                       | MoSCoW |
| ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| REQ-480 | Function `geokrety.fn_gk_moves_relations()` exists                                                                                                                | MUST   |
| REQ-481 | Trigger `tr_gk_moves_after_relations` AFTER INSERT OR UPDATE OR DELETE on `geokrety.gk_moves`                                                                     | MUST   |
| REQ-482 | Only `move_type IN (0,1,3,5)` qualify — COMMENT (2) and ARCHIVE (4) are skipped entirely                                                                          | MUST   |
| REQ-483 | Anonymous moves (`author IS NULL`) are skipped entirely                                                                                                           | MUST   |
| REQ-484 | On qualifying move reconciliation: exact rows for `stats.gk_related_users (geokrety_id, user_id)` are rebuilt from current qualifying `gk_moves` rows | MUST   |
| REQ-485 | For affected users, `stats.user_related_users` is rebuilt from current `stats.gk_related_users` so both `(A,B)` and `(B,A)` remain in sync | MUST   |
| REQ-486 | `user_related_users` is symmetric: both `(A,B)` and `(B,A)` always exist with equal `shared_geokrety_count`                                                       | MUST   |
| REQ-487 | On DELETE or UPDATE reconciliation: OLD contributions are removed exactly; rows disappear when no qualifying interactions remain                                   | MUST   |
| REQ-488 | `shared_geokrety_count` is distinct-shared-GK based and must not increment or decrement per raw move row                                                           | MUST   |
| REQ-489 | `phinx rollback` drops trigger then function cleanly                                                                                                              | MUST   |

---

## Acceptance Criteria

| #   | Criterion                                                                                | How to Verify                                       |
| --- | ---------------------------------------------------------------------------------------- | --------------------------------------------------- |
| 1   | Function exists in `geokrety` schema                                                     | `\df geokrety.fn_gk_moves_relations`                |
| 2   | Trigger attached to `geokrety.gk_moves`                                                  | `\d geokrety.gk_moves` trigger section              |
| 3   | COMMENT or ARCHIVE move → no rows in `gk_related_users`                                  | Insert type 2 or 4; check table                     |
| 4   | Anonymous move → no rows                                                                 | Insert with `author=NULL`; check table              |
| 5   | Two different users DROP/GRAB same GK → both `(A,B)` and `(B,A)` in `user_related_users` | Insert two qualifying moves; SELECT both directions |
| 6   | Symmetric pair counts match                                                              | Query for asymmetric pairs → 0                      |
| 7   | UPDATE / DELETE reconciliation is exact and row removed at count=0                      | Insert then DELETE/UPDATE; compare to rebuilt source truth |
| 8   | Rollback removes trigger and function cleanly                                            | `phinx rollback`; check `pg_trigger` + `pg_proc`    |

---

## Migration File

**`20260310400800_create_relation_trigger.php`**

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateRelationTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_relations()
  RETURNS TRIGGER LANGUAGE plpgsql SECURITY DEFINER
AS $$
DECLARE
  v_geokrety_ids INT[] := ARRAY[]::INT[];
  v_user_ids INT[] := ARRAY[]::INT[];
BEGIN
  IF TG_OP IN ('DELETE', 'UPDATE')
     AND OLD.author IS NOT NULL
     AND OLD.move_type IN (0,1,3,5) THEN
    v_geokrety_ids := array_append(v_geokrety_ids, OLD.geokret_id);
    v_user_ids := array_append(v_user_ids, OLD.author);
  END IF;

  IF TG_OP IN ('INSERT', 'UPDATE')
     AND NEW.author IS NOT NULL
     AND NEW.move_type IN (0,1,3,5) THEN
    v_geokrety_ids := array_append(v_geokrety_ids, NEW.geokret_id);
    v_user_ids := array_append(v_user_ids, NEW.author);
  END IF;

  IF cardinality(v_geokrety_ids) = 0 THEN
    RETURN COALESCE(NEW, OLD);
  END IF;

  SELECT array_agg(DISTINCT geokrety_id)
  INTO v_geokrety_ids
  FROM unnest(v_geokrety_ids) AS t(geokrety_id);

  SELECT array_agg(DISTINCT user_id)
  INTO v_user_ids
  FROM (
    SELECT unnest(v_user_ids) AS user_id
    UNION
    SELECT m.author
    FROM geokrety.gk_moves m
    WHERE m.author IS NOT NULL
      AND m.move_type IN (0,1,3,5)
      AND m.geokret_id = ANY(v_geokrety_ids)
  ) AS affected_users;

  DELETE FROM stats.gk_related_users
  WHERE geokrety_id = ANY(v_geokrety_ids);

  INSERT INTO stats.gk_related_users (
    geokrety_id,
    user_id,
    interaction_count,
    first_interaction,
    last_interaction
  )
  SELECT
    m.geokret_id,
    m.author,
    COUNT(*)::BIGINT,
    MIN(m.moved_on_datetime),
    MAX(m.moved_on_datetime)
  FROM geokrety.gk_moves m
  WHERE m.author IS NOT NULL
    AND m.move_type IN (0,1,3,5)
    AND m.geokret_id = ANY(v_geokrety_ids)
  GROUP BY m.geokret_id, m.author;

  IF v_user_ids IS NOT NULL AND cardinality(v_user_ids) > 0 THEN
    DELETE FROM stats.user_related_users
    WHERE user_id = ANY(v_user_ids)
       OR related_user_id = ANY(v_user_ids);

    INSERT INTO stats.user_related_users (
      user_id,
      related_user_id,
      shared_geokrety_count,
      first_seen_at,
      last_seen_at
    )
    SELECT
      a.user_id,
      b.user_id AS related_user_id,
      COUNT(DISTINCT a.geokrety_id)::BIGINT,
      MIN(LEAST(a.first_interaction, b.first_interaction)),
      MAX(GREATEST(a.last_interaction, b.last_interaction))
    FROM stats.gk_related_users a
    JOIN stats.gk_related_users b
      ON a.geokrety_id = b.geokrety_id
     AND a.user_id <> b.user_id
    WHERE a.user_id = ANY(v_user_ids)
       OR b.user_id = ANY(v_user_ids)
    GROUP BY a.user_id, b.user_id;
  END IF;

  RETURN COALESCE(NEW, OLD);
END;
$$;

CREATE TRIGGER tr_gk_moves_after_relations
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_relations();
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP TRIGGER IF EXISTS tr_gk_moves_after_relations ON geokrety.gk_moves;
DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_relations();
SQL
        );
    }
}
```

## SQL Usage Examples

```sql
-- Verify trigger attached
SELECT tgname FROM pg_trigger
WHERE tgrelid = 'geokrety.gk_moves'::regclass
  AND tgname = 'tr_gk_moves_after_relations';

-- Check social graph for GK 123: who touched it?
SELECT user_id, interaction_count FROM stats.gk_related_users WHERE geokrety_id = 123;

-- Who are user 9876's social connections via shared GKs?
SELECT related_user_id, shared_geokrety_count
FROM stats.user_related_users
WHERE user_id = 9876
ORDER BY shared_geokrety_count DESC LIMIT 20;

-- Verify symmetry: for each (A→B), (B→A) must exist
SELECT COUNT(*) AS asymmetric_pairs
FROM stats.user_related_users a
WHERE NOT EXISTS (
  SELECT 1 FROM stats.user_related_users b
  WHERE b.user_id = a.related_user_id AND b.related_user_id = a.user_id
);
-- Must return 0
```

## pgTAP Unit Tests

| Test ID   | Assertion                                                                    | Expected |
| --------- | ---------------------------------------------------------------------------- | -------- |
| T-4.9.001 | Function `geokrety.fn_gk_moves_relations()` exists                           | pass     |
| T-4.9.002 | Trigger `tr_gk_moves_after_relations` exists on `geokrety.gk_moves`          | pass     |
| T-4.9.003 | INSERT DROP by user A → row in `gk_related_users (geokrety_id, A)`           | pass     |
| T-4.9.004 | INSERT DROP by user B on same GK → row `gk_related_users (geokrety_id, B)`   | pass     |
| T-4.9.005 | After A and B on same GK → `user_related_users` has both `(A,B)` and `(B,A)` | pass     |
| T-4.9.006 | COMMENT move (type 2) → no `gk_related_users` row                            | pass     |
| T-4.9.007 | Anonymous move (author=NULL) → no `gk_related_users` row                     | pass     |
| T-4.9.008 | Same user twice on same GK → `interaction_count = 2` in `gk_related_users`   | pass     |
| T-4.9.009 | Symmetric check: `(A,B)` and `(B,A)` have equal `shared_geokrety_count`      | pass     |
| T-4.9.010 | UPDATE / DELETE reconciliation preserves exact relation rows                  | pass     |
| T-4.9.011 | `phinx rollback` drops trigger and function                                  | pass     |

| Test ID   | Scenario                                                               | Pass Condition                         |
| --------- | ---------------------------------------------------------------------- | -------------------------------------- |
| T-4.9.001 | Function `fn_gk_moves_relations` exists                                | pgTAP `has_function()`                 |
| T-4.9.002 | Trigger exists on `geokrety.gk_moves`                                  | pgTAP `has_trigger()`                  |
| T-4.9.003 | User A DROP on GK1 → `gk_related_users(GK1, A)` created                | 1 row, `interaction_count=1`           |
| T-4.9.004 | User B DROP on GK1 → `gk_related_users(GK1, B)` created                | 1 row, `interaction_count=1`           |
| T-4.9.005 | Both A and B on GK1 → both `(A,B)` and `(B,A)` in `user_related_users` | 2 rows, `shared_geokrety_count=1` each |
| T-4.9.006 | COMMENT (type 2) → no `gk_related_users` row                           | 0 rows                                 |
| T-4.9.007 | Anonymous (author=NULL) → no `gk_related_users` row                    | 0 rows                                 |
| T-4.9.008 | Same user twice on GK1 → `interaction_count = 2`                       | 1 row with count=2                     |
| T-4.9.009 | Symmetry: `shared_geokrety_count(A→B) == shared_geokrety_count(B→A)`   | Equality check                         |
| T-4.9.010 | UPDATE / DELETE reconciliation exact                                   | Row counts match rebuilt source truth  |
| T-4.9.011 | `phinx rollback` drops trigger + function                              | Both absent after rollback             |

---

## Implementation Checklist

- [ ] 1. Write `20260310400800_create_relation_trigger.php` with `up()` + `down()`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Trigger `tr_gk_moves_after_relations` present in `\d geokrety.gk_moves`
- [ ] 4. Function listed in `\df geokrety.fn_gk_moves_relations`
- [ ] 5. Insert moves by two users on same GK; confirm `user_related_users` symmetry
- [ ] 6. Test COMMENT move → no relation rows
- [ ] 7. Test anonymous move → no relation rows
- [ ] 8. Test `interaction_count` increment on repeated move by same user
- [ ] 9. Symmetry check query returns 0 asymmetric pairs
- [ ] 10. UPDATE / DELETE reconciliation rebuilds exact relation rows for touched GeoKrety and affected users
- [ ] 11. Run pgTAP T-4.9.001 through T-4.9.011 — all pass
- [ ] 12. `phinx rollback` — trigger and function dropped

- [ ] 1. Write `20260310400800_create_relation_trigger.php` with `up()` + `down()`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Trigger present; function exists
- [ ] 4. Simulate 2-user DROP scenario; verify both directions in `user_related_users`
- [ ] 5. Test COMMENT and ARCHIVE moves → no relation rows
- [ ] 6. Test anonymous move → no relation rows
- [ ] 7. Symmetry check query returns 0 asymmetric rows
- [ ] 8. DELETE scenario tested; row removed at count 0
- [ ] 9. Run pgTAP T-4.9.001 through T-4.9.011 — all pass
- [ ] 10. `phinx rollback` — clean removal

## Full SQL DDL — Trigger Function

The canonical function body is the same one embedded in the migration block above under `20260310400800_create_relation_trigger.php`. This duplicate lower DDL section was intentionally collapsed to avoid keeping stale merged SQL that used obsolete relation columns and incremental per-move pair logic.

## Full SQL DDL — Trigger Attachment

```sql
CREATE TRIGGER tr_gk_moves_after_relations
  AFTER INSERT OR UPDATE OR DELETE
  ON geokrety.gk_moves
  FOR EACH ROW
  EXECUTE FUNCTION geokrety.fn_gk_moves_relations();

COMMENT ON TRIGGER tr_gk_moves_after_relations ON geokrety.gk_moves
  IS 'Fires on INSERT/UPDATE/DELETE to maintain stats.gk_related_users and stats.user_related_users';
```

## Qualifying Move Types Reference

| move_type | Name    | Qualifies | Reason                             |
| --------- | ------- | --------- | ---------------------------------- |
| 0         | DROP    | ✅ YES    | Physical presence confirmed        |
| 1         | GRAB    | ✅ YES    | Change of hands                    |
| 2         | COMMENT | ❌ NO     | No physical presence               |
| 3         | SEEN    | ✅ YES    | Physical observation               |
| 4         | ARCHIVE | ❌ NO     | Retirement, not social interaction |
| 5         | DIP     | ✅ YES    | In-hand movement                   |

## Master-Spec Alignment

This task is governed by [../00-SPEC-DRAFT-v1.obsolete.md](../00-SPEC-DRAFT-v1.obsolete.md), Sections 5.6 and 8.4.

- Canonical table names and columns are `stats.gk_related_users(geokrety_id, user_id, interaction_count, first_interaction, last_interaction)` and `stats.user_related_users(user_id, related_user_id, shared_geokrety_count, first_seen_at, last_seen_at)`.
- `tr_gk_moves_after_relations` must be documented as `AFTER INSERT OR UPDATE OR DELETE` with exact `OLD` removal plus `NEW` application semantics.
- `shared_geokrety_count` counts distinct shared GeoKrety for a user pair, not raw move rows.
- Any lower draft text that uses names such as `geokret_id`, `touch_count`, or decrements shared counts per move row is obsolete and superseded by this alignment block.

## Agent Loop Log

- 2026-03-10T18:40:00Z — `dba`: replaced stale incremental relation logic with exact recomputation for touched GeoKrety and affected user pairs using canonical column names.
- 2026-03-10T18:40:00Z — `critical-thinking`: distinct shared-GK counts cannot be maintained safely by raw per-move decrement logic, so the spec now rebuilds the affected relation surface exactly.
- 2026-03-10T18:40:00Z — `specification`: aligned trigger events, requirements, tests, and checklist with Q-028 and the Sprint 4 table contracts from S4T06/S4T07.

## Resolution

Q-028 is resolved by canonizing exact INSERT / UPDATE / DELETE reconciliation for relation tables in this task.

## Objects Created

| Object Type | Name                               | Owning Schema          |
| ----------- | ---------------------------------- | ---------------------- |
| Function    | `geokrety.fn_gk_moves_relations()` | `geokrety`             |
| Trigger     | `tr_gk_moves_after_relations`      | on `geokrety.gk_moves` |

## Side-effects (rows modified in `stats.*`)

| Table                      | Operation                     | Condition                                       |
| -------------------------- | ----------------------------- | ----------------------------------------------- |
| `stats.gk_related_users`   | INSERT/UPDATE/DELETE reconciliation | `author IS NOT NULL AND move_type IN (0,1,3,5)` |
| `stats.user_related_users` | INSERT/UPDATE/DELETE reconciliation | When distinct shared-GK membership changes      |

---

## Edge Cases to Test

| Edge Case                         | Expected Behavior                                            |
| --------------------------------- | ------------------------------------------------------------ |
| user_id == related_user_id insert | Prevented by CHECK constraint in S4T07; trigger skips pair   |
| Three users on same GK            | 6 pairs created: (A,B), (B,A), (A,C), (C,A), (B,C), (C,B)    |
| Same user on GK 3 times           | `interaction_count = 3`; no new `user_related_users` for same user |

---
