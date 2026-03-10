---
title: "Task S5T05: gk_loves Activity Trigger"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 5
tags:
  - country-stats
  - daily-activity
  - database
  - database-refactor
  - dba
  - gk_loves
  - specification
  - sprint-5
  - sql
  - stats
  - task-index
  - task-merge
  - trigger
depends_on:
  - "Sprint 1 foundation"
  - "Sprint 2 daily activity"
  - "Sprint 3 country stats"
task: S5T05
step: 5.5
migration: 20260310500400_create_gk_loves_counter_trigger.php
blocks:
  - S5T06
changelog:
  - 2026-03-10: created by merge of task-S5T05.dba.md and task-S5T05.specification.md
  - 2026-03-10: resolved Q-031 by restoring the canonical loves-trigger scope
---

# Task S5T05: gk_loves Activity Trigger

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T05.dba.md`
- Specification source: `task-S5T05.specification.md`

## Resolved Decision

- S5T05 is canonically the `gk_loves` activity trigger task from the Sprint 5 index.
- The stale `stats.outbox_events` draft was merge residue from the event-bridge work and is not part of this task.
- Love attribution uses the loved GK current country context when available; otherwise only shard and daily counters are updated.

## Purpose & Scope

Defines the trigger function and trigger attachment on `geokrety.gk_loves` that maintain love-related counters and rollups across insert, update, and delete operations.

**Scope:**

- `stats.entity_counters_shard` for entity `gk_loves`
- `stats.daily_activity.loves_count`
- `stats.country_daily_stats.loves_count` when GK country context is resolvable

**Out of scope:**

- Event bridge or queue emission mechanics
- `stats.outbox_events`
- `pg_notify`

## Requirements

| ID      | Description                                                                                             | MoSCoW |
| ------- | ------------------------------------------------------------------------------------------------------- | ------ |
| REQ-580 | Function `geokrety.fn_gk_loves_activity()` exists                                                       | MUST   |
| REQ-581 | Trigger `tr_gk_loves_activity` is attached to `geokrety.gk_loves`                                      | MUST   |
| REQ-582 | The trigger fires `AFTER INSERT OR UPDATE OR DELETE` on `geokrety.gk_loves`                            | MUST   |
| REQ-583 | Love changes update `stats.entity_counters_shard` for entity `gk_loves`                                | MUST   |
| REQ-584 | Love changes update `stats.daily_activity.loves_count`                                                  | MUST   |
| REQ-585 | Love changes update `stats.country_daily_stats.loves_count` using GK country context when available    | MUST   |
| REQ-586 | When GK country context is unavailable, country love rollups are skipped without affecting daily totals | MUST   |
| REQ-587 | `phinx rollback` drops the trigger and function cleanly                                                 | MUST   |

## Acceptance Criteria

| #   | Criterion                                                       | How to Verify                                          |
| --- | --------------------------------------------------------------- | ------------------------------------------------------ |
| 1   | Function created in `geokrety` schema                           | `\df geokrety.fn_gk_loves_activity`                   |
| 2   | Trigger attached to `geokrety.gk_loves`                         | `\d geokrety.gk_loves` trigger list                   |
| 3   | Love insert/update/delete adjusts the `gk_loves` shard counter  | Verify `SUM(cnt)` deltas for the entity                |
| 4   | Love insert/update/delete adjusts `daily_activity.loves_count`  | Verify touched day buckets                             |
| 5   | Country rollup uses GK country context when available           | Verify touched country/day buckets                     |
| 6   | Missing country context skips only the country rollup           | Verify daily counters change without country mutation  |
| 7   | Rollback drops trigger and function                             | Both absent after `phinx rollback`                     |

## Migration File

**`20260310500400_create_gk_loves_counter_trigger.php`**

## Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_loves_activity()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
BEGIN
  -- Maintain `stats.entity_counters_shard` for entity `gk_loves`.
  -- Maintain `stats.daily_activity.loves_count` for the touched date bucket(s).
  -- Maintain `stats.country_daily_stats.loves_count` only when the loved GK
  -- current country context can be resolved from the canonical GK location path.
  -- UPDATE reconciles OLD then NEW exactly for touched date/country buckets.
  RETURN COALESCE(NEW, OLD);
END;
$$;

CREATE TRIGGER tr_gk_loves_activity
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_loves
  FOR EACH ROW
  EXECUTE FUNCTION geokrety.fn_gk_loves_activity();
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkLovesCounterTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_loves_activity()
  RETURNS TRIGGER LANGUAGE plpgsql SECURITY DEFINER
AS $$
BEGIN
  RETURN COALESCE(NEW, OLD);
END;
$$;

CREATE TRIGGER tr_gk_loves_activity
  AFTER INSERT OR UPDATE OR DELETE ON geokrety.gk_loves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_loves_activity();
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP TRIGGER IF EXISTS tr_gk_loves_activity ON geokrety.gk_loves;
DROP FUNCTION IF EXISTS geokrety.fn_gk_loves_activity();
SQL
        );
    }
}
```

The placeholder function body above must be replaced by the canonical counter-maintenance logic required by REQ-583 through REQ-586.

## Canonical Notes

- `gk_loves` is assumed to exist per Q-001 and Q-007.
- Country attribution follows Q-005: use the loved GK current country context if resolvable; otherwise skip only the country rollup.
- Event-bridge concerns are handled separately in S5T06.

## pgTAP Unit Tests

| Test ID   | Assertion                                                        | Pass Condition |
| --------- | ---------------------------------------------------------------- | -------------- |
| T-5.5.001 | Function exists                                                  | `has_function()` |
| T-5.5.002 | Trigger exists                                                   | `has_trigger()` |
| T-5.5.003 | `gk_loves` shard counter updates across INSERT/UPDATE/DELETE     | exact match    |
| T-5.5.004 | `daily_activity.loves_count` updates across INSERT/UPDATE/DELETE | exact match    |
| T-5.5.005 | Country rollup follows GK country context rules                  | exact match    |
| T-5.5.006 | Rollback removes trigger and function                            | pass           |

## Implementation Checklist

- [ ] 1. Write `20260310500400_create_gk_loves_counter_trigger.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Verify trigger and function are attached correctly
- [ ] 4. Test insert, update, and delete behavior for shard, daily, and country love rollups
- [ ] 5. Run pgTAP T-5.5.001 through T-5.5.006 — all pass
- [ ] 6. `phinx rollback` — trigger and function gone

## Canonical Alignment

- The canonical requirement is a `gk_loves` trigger family that updates `stats.entity_counters_shard`, `stats.daily_activity`, and `stats.country_daily_stats` where country attribution is available.
- Loves remain part of both daily and country rollups.
- `stats.outbox_events` is not part of the canonical S5T05 scope.

## Agent Loop Log

- 2026-03-10T19:20:00Z — `dba`: removed the merged outbox-table identity from S5T05 and restored the loves-trigger contract from the sprint index.
- 2026-03-10T19:20:00Z — `critical-thinking`: separated loves-counter maintenance from move-event bridge concerns so the task boundaries match the Sprint 5 plan.
- 2026-03-10T19:20:00Z — `specification`: aligned migration naming, scope, and attribution rules with Q-001, Q-005, Q-007, and the Sprint 5 index.

## Resolution

Q-031 is resolved by canonizing S5T05 as the `gk_loves` activity trigger task only.

## Objects Created

| Object Type | Name                     | Schema              |
| ----------- | ------------------------ | ------------------- |
| Function    | `fn_gk_loves_activity()` | `geokrety`          |
| Trigger     | `tr_gk_loves_activity`   | `geokrety.gk_loves` |

## Country Attribution Rule

Country love rollups use the GeoKret current country context available at the time of the love event. When the loved GK has no resolvable country context, the trigger still updates shard and daily counters but skips `stats.country_daily_stats`.

## Security Note

The trigger should run with only the privileges required to update the canonical stats tables.
