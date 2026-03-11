---
title: AMQP Schema
description: Broker configuration used by queue notification flows.
icon: material/rabbit
---

# `amqp` schema

This schema is intentionally small. It holds broker connection configuration that is consumed by queue-aware notification flows.

## Inventory

- `broker`: host, port, vhost, username, and password per broker entry
- live helper functions also exist, including AMQP publish and connection helpers such as `publish`, `exchange_declare`, `disconnect`, and `autonomous_publish`

## Relation to the stats branch

The stats branch does not write directly to `amqp`, but many source-table triggers in `geokrety` emit changes into `notify_queues`, which then bridge toward asynchronous processing. `amqp` is therefore part of the surrounding operational topology even though it is not part of the canonical analytics model.

## Operational notes

- treat access as privileged
- keep credentials out of rendered public documentation or logs
- changes here can affect downstream consumers without altering the database analytics layer itself

## Operator check

- confirm broker rows match the intended environment and credential rotation policy
- verify ownership of helper functions before changing queue-consumer behavior

## See also

- [Notify queues](specs.notify_queues.md)
- [Schema hub](specs.md)
