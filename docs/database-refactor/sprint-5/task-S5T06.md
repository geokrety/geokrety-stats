---
title: "Task S5T06: Points Event Bridge Trigger"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 5
tags:
  - database
  - database-refactor
  - dba
  - event-bridge
  - function
  - rabbitmq
  - specification
  - sprint-5
  - sql
  - task-index
  - task-merge
  - trigger
depends_on:
  - S5T05
task: S5T06
step: 5.6
migration: 20260310500500_create_amqp_event_trigger.php
blocks:
  - S5T10
changelog:
  - 2026-03-10: created by merge of task-S5T06.dba.md and task-S5T06.specification.md
  - 2026-03-10: resolved Q-032 by restoring the canonical transport-neutral bridge contract
---

# Task S5T06: Points Event Bridge Trigger

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T06.dba.md`
- Specification source: `task-S5T06.specification.md`

## Resolved Decision

- S5T06 is canonically the move-event emission trigger for `points-awarder`.
- The bridge remains transport-neutral in the spec even though the current deployment target is RabbitMQ.
- The canonical live scope is `AFTER INSERT` on `geokrety.gk_moves` only, emitting one compact event reference per inserted move.
- `stats.outbox_events` and `pg_notify` are non-canonical candidate implementations, not Sprint 5 schema requirements.

## Purpose & Scope

After each insert on `geokrety.gk_moves`, this trigger atomically emits the canonical scoring event to the approved bridge implementation. The bridge contract is intentionally minimal so the consumer can re-read canonical move data from the database.

**Scope:**

- bridge trigger function
- trigger attachment on `geokrety.gk_moves`
- minimal payload contract for asynchronous scoring

**Out of scope:**

- required `stats.outbox_events` schema
- required `pg_notify`
- update/delete event emission
- transport-specific implementation details

## Requirements

| ID      | Description                                                                                   | MoSCoW |
| ------- | --------------------------------------------------------------------------------------------- | ------ |
| REQ-600 | Canonical bridge function for move-event emission exists                                      | MUST   |
| REQ-601 | Trigger `tr_gk_moves_emit_points_event` fires `AFTER INSERT` on `geokrety.gk_moves` only     | MUST   |
| REQ-602 | INSERT emits one compact scoring-event reference to the approved bridge                       | MUST   |
| REQ-603 | Bridge write is in the same transaction as the originating insert                             | MUST   |
| REQ-604 | The canonical minimal payload is `type + id`, sufficient for `points-awarder` to re-read data | MUST  |
| REQ-605 | Anonymous moves are allowed; no actor field is required in the canonical minimal payload      | MUST   |
| REQ-610 | `phinx rollback` drops trigger then function cleanly                                          | MUST   |

## Acceptance Criteria

| #   | Criterion                                              | How to Verify                           |
| --- | ------------------------------------------------------ | --------------------------------------- |
| 1   | Canonical bridge function created in `geokrety` schema | `\df geokrety.*emit*points*event*`      |
| 2   | Trigger attached to `geokrety.gk_moves`                | `\d geokrety.gk_moves` trigger list    |
| 3   | INSERT emits one canonical scoring event               | Check approved bridge after test insert |
| 4   | Payload contains canonical `type + id` reference       | Inspect emitted payload                 |
| 5   | Anonymous move remains emit-eligible                   | Test with `author = NULL` move          |
| 6   | Rollback removes trigger and function                  | `phinx rollback`                        |

## Migration File

**`20260310500500_create_amqp_event_trigger.php`**

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateAmqpEventTrigger extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE OR REPLACE FUNCTION geokrety.fn_gk_moves_emit_points_event()
  RETURNS TRIGGER LANGUAGE plpgsql SECURITY DEFINER
AS $$
BEGIN
  -- Bridge implementation remains abstract. Canonical contract: enqueue one
  -- compact move-event reference atomically with the inserted move row.
  RETURN NEW;
END;
$$;

CREATE TRIGGER tr_gk_moves_emit_points_event
  AFTER INSERT ON geokrety.gk_moves
  FOR EACH ROW EXECUTE FUNCTION geokrety.fn_gk_moves_emit_points_event();
SQL
        );
    }

    public function down(): void
    {
        $this->execute(<<<'SQL'
DROP TRIGGER IF EXISTS tr_gk_moves_emit_points_event ON geokrety.gk_moves;
DROP FUNCTION IF EXISTS geokrety.fn_gk_moves_emit_points_event();
SQL
        );
    }
}
```

The placeholder function body above must be replaced by the approved bridge write while preserving the canonical trigger name, `AFTER INSERT` scope, and minimal payload contract.

## Canonical Payload Contract

```json
{
  "type": "gk_move_created",
  "id": 9999
}
```

`points-awarder` re-reads the move and related records from the database using this reference.

## SQL Usage Examples

```sql
SELECT tgname
FROM pg_trigger
WHERE tgrelid = 'geokrety.gk_moves'::regclass
  AND tgname = 'tr_gk_moves_emit_points_event';

INSERT INTO geokrety.gk_moves (...) VALUES (...);
```

## pgTAP Unit Tests

| Test ID   | Assertion                                       | Pass Condition |
| --------- | ----------------------------------------------- | -------------- |
| T-5.6.001 | Function exists                                 | `has_function()` |
| T-5.6.002 | Trigger exists                                  | `has_trigger()` |
| T-5.6.003 | INSERT move emits one canonical event           | 1 emitted record |
| T-5.6.004 | Payload is the canonical `type + id` reference  | exact match     |
| T-5.6.005 | Anonymous move remains emit-eligible            | pass            |
| T-5.6.006 | Rollback drops trigger and function             | pass            |

## Implementation Checklist

- [ ] 1. Write `20260310500500_create_amqp_event_trigger.php` using the canonical bridge function name
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. Trigger `tr_gk_moves_emit_points_event` present in `\d geokrety.gk_moves`
- [ ] 4. Insert a move and verify one bridge write is emitted in the same transaction
- [ ] 5. Verify emitted payload is the canonical `type + id` reference
- [ ] 6. Verify anonymous moves remain emit-eligible
- [ ] 7. Run pgTAP T-5.6.001 through T-5.6.006 — all pass
- [ ] 8. `phinx rollback` — trigger and function dropped

## Canonical Alignment

- Canonical trigger family name: `tr_gk_moves_emit_points_event`.
- Canonical live scope: `AFTER INSERT` on `geokrety.gk_moves` only.
- The canonical minimal payload is a compact move-event reference, not a denormalized envelope.
- Any implementation using `stats.outbox_events`, `pg_notify`, or other transport-specific mechanics is optional and non-canonical at the Sprint 5 spec layer.

## Agent Loop Log

- 2026-03-10T19:20:00Z — `dba`: removed the concrete outbox-table requirement from the canonical S5T06 contract while preserving the Sprint 5 move-event trigger boundary.
- 2026-03-10T19:20:00Z — `critical-thinking`: reconciled Q-002 and Q-014 by keeping the bridge abstract in-spec while using the RabbitMQ `type + id` pattern as the minimal payload contract.
- 2026-03-10T19:20:00Z — `specification`: aligned trigger name, trigger scope, payload, and rollback semantics with the sprint index.

## Resolution

Q-032 is resolved by canonizing S5T06 as a transport-neutral `AFTER INSERT` move-event bridge with a minimal `type + id` payload contract.

## Objects Created

| Object Type | Name                              | Owning Schema         |
| ----------- | --------------------------------- | --------------------- |
| Function    | `fn_gk_moves_emit_points_event()` | `geokrety`            |
| Trigger     | `tr_gk_moves_emit_points_event`   | `geokrety.gk_moves`   |

## Side-effects

| Target                  | Operation | Condition                           |
| ----------------------- | --------- | ----------------------------------- |
| approved bridge backend | INSERT    | Every INSERT on `geokrety.gk_moves` |
