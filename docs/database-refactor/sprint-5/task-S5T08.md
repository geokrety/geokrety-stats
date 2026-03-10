---
title: "Task S5T08: First-Finder Detection Trigger"
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
  - first-finder
  - function
  - specification
  - sprint-5
  - sql
  - stats
  - task-index
  - task-merge
  - trigger
depends_on:
  - S5T04
  - S5T07
task: S5T08
step: 5.8
migration: 20260310500700_create_first_finder_trigger.php
blocks:
  - S5T09
changelog:
  - 2026-03-10: created by merge of task-S5T08.dba.md and task-S5T08.specification.md
  - 2026-03-10: resolved Q-034 by canonizing the live first-finder trigger scope
---

# Task S5T08: First-Finder Detection Trigger

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T08.dba.md`
- Specification source: `task-S5T08.specification.md`

## Resolved Decision

- S5T08 canonically owns the live first-finder trigger function plus trigger attachment.
- The canonical migration name remains `20260310500700_create_first_finder_trigger.php`.
- The live trigger remains the canonical task identity, but all execution paths must share the same normative helper logic.
- The trigger-versus-function conflict is resolved in favor of a live trigger surface because Sprint 5 schedules first-finder as an attached trigger, not as a free-standing utility task.

## Purpose & Scope

Creates the live first-finder detection trigger function and attaches it to `geokrety.gk_moves` so qualifying inserts append one canonical row to `stats.first_finder_events`.

**Scope:**

- live `AFTER INSERT` detection on `geokrety.gk_moves`
- canonical non-owner and anonymous guard clauses
- 168-hour cutoff from GK creation time
- one-row-per-GK semantics in `stats.first_finder_events`

**Out of scope:**

- milestone detection
- hourly activity or country-pair aggregation
- standalone backfill orchestration

## Requirements

| ID      | Description                                                                                     | MoSCoW |
| ------- | ----------------------------------------------------------------------------------------------- | ------ |
| REQ-640 | Function `geokrety.fn_gk_moves_first_finder()` exists                                           | MUST   |
| REQ-641 | Trigger `tr_gk_moves_after_first_finder` fires `AFTER INSERT` on `geokrety.gk_moves`           | MUST   |
| REQ-642 | Anonymous moves do not create first-finder rows                                                 | MUST   |
| REQ-643 | Owner moves do not create first-finder rows                                                     | MUST   |
| REQ-644 | Only qualifying move types can create first-finder rows                                         | MUST   |
| REQ-645 | Candidate moves later than 168 hours after GK creation are rejected                             | MUST   |
| REQ-646 | First qualifying non-owner move inserts one canonical row in `stats.first_finder_events`       | MUST   |
| REQ-647 | Inserts use idempotent one-row-per-GK semantics via `ON CONFLICT (gk_id) DO NOTHING`           | MUST   |
| REQ-648 | `phinx rollback` drops trigger then function cleanly                                            | MUST   |

## Acceptance Criteria

| #   | Criterion                                             | How to Verify                                  |
| --- | ----------------------------------------------------- | ---------------------------------------------- |
| 1   | Function created in `geokrety` schema                 | `\df geokrety.fn_gk_moves_first_finder`        |
| 2   | Trigger attached to `geokrety.gk_moves`               | `\d geokrety.gk_moves` trigger list           |
| 3   | First qualifying non-owner move creates one row       | Verify insert into `stats.first_finder_events` |
| 4   | Owner and anonymous moves create no rows              | Verify no insert                               |
| 5   | Moves after 168 hours create no rows                  | Verify no insert                               |
| 6   | Repeated qualifying moves do not duplicate rows       | Verify row count stays 1                       |
| 7   | Rollback removes trigger and function                 | `phinx rollback`                               |

## Migration File

**`20260310500700_create_first_finder_trigger.php`**

## Full SQL DDL

```sql
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_first_finder()
  RETURNS TRIGGER
  LANGUAGE plpgsql
  SECURITY DEFINER
AS $$
BEGIN
  -- On live INSERT only:
  -- 1. Reject anonymous, owner, and non-qualifying move types.
  -- 2. Reject candidates later than 168 hours after GK creation.
  -- 3. Insert one canonical row into `stats.first_finder_events` with
  --    `ON CONFLICT (gk_id) DO NOTHING` semantics.
  RETURN NEW;
END;
$$;

CREATE TRIGGER tr_gk_moves_after_first_finder
  AFTER INSERT ON geokrety.gk_moves
  FOR EACH ROW
  EXECUTE FUNCTION geokrety.fn_gk_moves_first_finder();
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateFirstFinderTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_first_finder()
  RETURNS TRIGGER LANGUAGE plpgsql SECURITY DEFINER
AS $$
BEGIN
  RETURN NEW;
END;
$$;

CREATE TRIGGER tr_gk_moves_after_first_finder
  AFTER INSERT ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_first_finder();
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP TRIGGER IF EXISTS tr_gk_moves_after_first_finder ON geokrety.gk_moves;
DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_first_finder();
SQL
        );
    }
}
```

The placeholder function body above must be replaced by the canonical guard clauses and insert logic required by REQ-642 through REQ-647.

## Canonical Notes

- The canonical live surface is the attached trigger family in `geokrety`.
- The live trigger must call the shared normative helper `stats.fn_detect_first_finder(...)` so live and replay paths use the same eligibility rules.
- The helper does not change the task identity, migration name, or acceptance surface; it standardizes shared first-finder behavior across execution paths.
- Any merged text treating S5T08 as helper-only is obsolete.

## pgTAP Unit Tests

| Test ID   | Assertion                                            | Pass Condition |
| --------- | ---------------------------------------------------- | -------------- |
| T-5.8.001 | Function exists                                      | `has_function()` |
| T-5.8.002 | Trigger exists                                       | `has_trigger()` |
| T-5.8.003 | Non-owner qualifying move inserts one row            | exact match     |
| T-5.8.004 | Owner move inserts no row                            | exact match     |
| T-5.8.005 | Anonymous move inserts no row                        | exact match     |
| T-5.8.006 | Older-than-168-hours move inserts no row             | exact match     |
| T-5.8.007 | Repeated qualifying inserts do not duplicate rows    | exact match     |
| T-5.8.008 | Rollback removes trigger and function                | pass            |

## Implementation Checklist

- [ ] 1. Create `20260310500700_create_first_finder_trigger.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Confirm `tr_gk_moves_after_first_finder` in `\d geokrety.gk_moves`
- [ ] 4. Verify non-owner qualifying move insertion
- [ ] 5. Verify owner and anonymous guards
- [ ] 6. Verify 168-hour cutoff enforcement
- [ ] 7. Run pgTAP T-5.8.001 through T-5.8.008
- [ ] 8. `phinx rollback` — trigger and function dropped cleanly

## Canonical Alignment

- S5T08 is the live first-finder trigger task named in the Sprint 5 index.
- S5T07 handles milestones only.
- S5T09 handles batch/manual hourly and country-pair rollups only.

## Agent Loop Log

- 2026-03-10T19:55:00Z — `dba`: removed helper-only identity from S5T08 and restored the live trigger contract scheduled in Sprint 5.
- 2026-03-10T19:55:00Z — `critical-thinking`: resolved the trigger-versus-function conflict by separating canonical trigger ownership from optional internal helper reuse.
- 2026-03-10T19:55:00Z — `specification`: canonized the migration name, trigger scope, and first-finder guard rules.

## Resolution

Q-034 is resolved by canonizing S5T08 as the live first-finder trigger task.

## AC-5.8.5 — Idempotency

**Given** `first_finder_events` already has a row for `gk_id`
**When** function called again (any user, any qualifying type)
**Then** no second row inserted; original row unchanged

## AC-5.8.6 — Missing GK Guard

**When** function called with `p_gk_id` that has no row in `geokrety.gk_geokrety`
**Then** function returns without error, no row inserted

## AC-5.8.7 — Hours Calculation

**Given** GK created at T0, move at T0 + 25 hours
**When** function runs
**Then** `hours_since_creation = 25`

## AC-5.8.8 — 168-Hour Cutoff Enforced

**Given** GK created more than 168 hours before the candidate move
**When** function runs
**Then** no first-finder row is inserted
