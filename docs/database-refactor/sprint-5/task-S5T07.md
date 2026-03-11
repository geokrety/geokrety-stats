---
title: "Task S5T07: Milestone Detection Trigger"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 5
tags:
  - analytics
  - database
  - database-refactor
  - dba
  - function
  - milestones
  - specification
  - sprint-5
  - sql
  - stats
  - task-index
  - task-merge
  - trigger
depends_on:
  - S5T03
  - S2
  - S4T06
task: S5T07
step: 5.7
migration: 20260310500600_create_milestone_trigger.php
blocks:
  - S5T10
changelog:
  - 2026.03.10: created by merge of task-S5T07.dba.md and task-S5T07.specification.md
  - 2026.03.10: resolved Q-033 by narrowing S5T07 to milestone detection only
---

# Task S5T07: Milestone Detection Trigger

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T07.dba.md`
- Specification source: `task-S5T07.specification.md`

## Resolved Decision

- S5T07 canonically owns the live milestone trigger only.
- Hourly activity and country-pair flows are batch/manual concerns owned by S5T09.
- Live first-finder detection is owned by S5T08.
- The canonical migration name remains `20260310500600_create_milestone_trigger.php`.

## Purpose & Scope

Creates the live milestone detection trigger function and attaches it to `geokrety.gk_moves` so qualifying inserts append milestone facts into `stats.gk_milestone_events`.

**Scope:**

- km milestones derived from qualifying kilometer-counting moves
- distinct-user milestones derived from canonical `stats.gk_related_users`
- one-time insert semantics into `stats.gk_milestone_events`

**Out of scope:**

- first-finder detection
- hourly activity rollups
- country-pair flow rollups

## Requirements

| ID      | Description                                                                           | MoSCoW |
| ------- | ------------------------------------------------------------------------------------- | ------ |
| REQ-620 | Function `geokrety.fn_gk_moves_milestones()` exists                                   | MUST   |
| REQ-621 | Trigger `tr_gk_moves_after_milestones` fires `AFTER INSERT` on `geokrety.gk_moves`   | MUST   |
| REQ-622 | Kilometer-counting moves evaluate canonical km milestones at 100, 1000, and 10000    | MUST   |
| REQ-623 | Distinct-user milestones evaluate canonical thresholds using `stats.gk_related_users` | MUST   |
| REQ-624 | Milestone inserts use one-time-event semantics via `ON CONFLICT DO NOTHING`           | MUST   |
| REQ-625 | Trigger logic is limited to live milestone detection and append-only event creation   | MUST   |
| REQ-626 | `phinx rollback` drops trigger then function cleanly                                  | MUST   |

## Acceptance Criteria

| #   | Criterion                                            | How to Verify                                   |
| --- | ---------------------------------------------------- | ----------------------------------------------- |
| 1   | Function created in `geokrety` schema                | `\df geokrety.fn_gk_moves_milestones`           |
| 2   | Trigger attached to `geokrety.gk_moves`              | `\d geokrety.gk_moves` trigger list            |
| 3   | Crossing 100 km inserts `km_100` once                | Insert qualifying data twice; row count stays 1 |
| 4   | Crossing user threshold inserts user milestone once  | Verify one row for the touched GK/event type    |
| 5   | Non-qualifying inserts do not create milestone rows  | Verify no new event                             |
| 6   | Rollback removes trigger and function                | `phinx rollback`                                |

## Migration File

**`20260310500600_create_milestone_trigger.php`**

## Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_milestones()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
BEGIN
  -- On live INSERT only:
  -- 1. Evaluate kilometer milestones from canonical km-counting move history.
  -- 2. Evaluate user-reach milestones from canonical `stats.gk_related_users`.
  -- 3. Append milestone rows into `stats.gk_milestone_events` with
  --    `ON CONFLICT DO NOTHING` semantics.
  RETURN NEW;
END;
$$;

CREATE TRIGGER tr_gk_moves_after_milestones
  AFTER INSERT ON geokrety.gk_moves
  FOR EACH ROW
  EXECUTE FUNCTION geokrety.fn_gk_moves_milestones();
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateMilestoneTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_milestones()
  RETURNS TRIGGER LANGUAGE plpgsql SECURITY DEFINER
AS $$
BEGIN
  RETURN NEW;
END;
$$;

CREATE TRIGGER tr_gk_moves_after_milestones
  AFTER INSERT ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_milestones();
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP TRIGGER IF EXISTS tr_gk_moves_after_milestones ON geokrety.gk_moves;
DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_milestones();
SQL
        );
    }
}
```

The placeholder function body above must be replaced by the canonical milestone checks required by REQ-622 through REQ-625.

## Canonical Notes

- Kilometer milestones are derived only from the canonical km-counting move types.
- User-count milestones read the already-maintained `stats.gk_related_users` table instead of rebuilding relationships inside this trigger.
- Any merged text assigning first-finder, hourly, or country-pair ownership to this task is obsolete.

## pgTAP Unit Tests

| Test ID   | Assertion                                          | Pass Condition |
| --------- | -------------------------------------------------- | -------------- |
| T-5.7.001 | Function exists                                    | `has_function()` |
| T-5.7.002 | Trigger exists                                     | `has_trigger()` |
| T-5.7.003 | `km_100` milestone inserted when threshold crossed | exact match     |
| T-5.7.004 | Re-check does not duplicate milestone rows         | exact match     |
| T-5.7.005 | User-threshold milestone inserted once             | exact match     |
| T-5.7.006 | Rollback removes trigger and function              | pass            |

## Implementation Checklist

- [ ] 1. Create `20260310500600_create_milestone_trigger.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Confirm `tr_gk_moves_after_milestones` in `\d geokrety.gk_moves`
- [ ] 4. Verify km milestones at 100, 1000, and 10000
- [ ] 5. Verify user-count milestone insertion using seeded `stats.gk_related_users`
- [ ] 6. Run pgTAP T-5.7.001 through T-5.7.006
- [ ] 7. `phinx rollback` — trigger and function dropped cleanly

## Canonical Alignment

- S5T07 is the live milestone trigger task named in the Sprint 5 index.
- First-finder trigger logic belongs to S5T08.
- Batch/manual analytics rollups belong to S5T09.

## Agent Loop Log

- 2026-03-10T19:55:00Z — `dba`: removed hourly, country-pair, and first-finder trigger scope from S5T07 so the task matches the Sprint 5 migration plan.
- 2026-03-10T19:55:00Z — `critical-thinking`: separated live milestone detection from other analytics concerns to avoid overlapping trigger ownership.
- 2026-03-10T19:55:00Z — `specification`: canonized the migration name and narrowed the acceptance surface to milestone events only.

## Resolution

Q-033 is resolved by canonizing S5T07 as the milestone detection trigger task only.
