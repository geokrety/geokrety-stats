---
title: Audit Schema
description: Request and action logging surfaces for operator and security review.
icon: material/file-document-alert
---

# `audit` schema

The `audit` schema contains append-heavy logging tables that are operationally important for traceability and abuse review.

## Inventory

- `actions_logs`: general event log with author, IP, context JSON, and session fields
- `posts`: request payload capture with route, errors, user agent, and timestamps

## Live footprint

- `actions_logs`: roughly `871k` estimated rows and `213 MB`
- `posts`: roughly `178k` estimated rows and `126 MB`

## Relation to stats

These tables do not directly feed `stats`, but they matter for:

- backfill incident analysis
- abuse or bot investigations affecting traffic-driven metrics
- forensic review when asynchronous scoring or maintenance jobs appear inconsistent with application activity

## Operational notes

- this schema is log-like and retention should be treated explicitly
- API exposure should be internal-only, if exposed at all
- JSON payload columns are useful for operator tooling but should not be part of public analytics contracts

## Operator check

- define and review retention windows for `actions_logs` and `posts`
- monitor table growth because both tables are already material in size

## See also

- [Schema hub](specs.md)
- [Geokrety schema](specs.geokrety.md)
