---
title: Notify Queues Schema
description: Database outbox and PostgreSQL NOTIFY bridge for downstream processing.
icon: material/message-arrow-right
---

# `notify_queues` schema

`notify_queues` is the database outbox bridge between transactional changes and downstream consumers.

## Inventory

- `geokrety_changes`: queue table holding channel, action, payload, timestamps, and error JSON
- trigger `channel_notify`: emits PostgreSQL `NOTIFY` on insert
- live handoff functions include `channel_notify()`, `amqp_notify_id()`, `amqp_notify_gkid()`, and helper entry points such as `new_handle`

## Relation to stats

The canonical `stats` schema is maintained synchronously through database triggers, but adjacent systems such as message-driven consumers or future API caches may still rely on the outbox stream produced here.

## Operational notes

- keep this schema small and monitored
- errors in queue processing do not automatically imply `stats` drift, because the branch’s main analytics maintenance is synchronous
- this schema is still a useful event bridge for non-core derivatives and notifications

## Operator check

- monitor unprocessed and errored rows in `geokrety_changes`
- verify trigger-produced queue rows are consumed at the expected rate

## See also

- [AMQP](specs.amqp.md)
- [Schema hub](specs.md)
