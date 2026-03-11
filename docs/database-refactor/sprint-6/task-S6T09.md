---
title: "Task S6T09: Performance Validation & Quality Gates"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 6
tags:
  - database
  - database-refactor
  - dba
  - explain
  - final
  - performance
  - quality-gate
  - specification
  - sprint-6
  - stats
  - task-index
  - task-merge
  - validation
depends_on:
  - S6T07
  - S6T08
task: S6T09
step: 6.9
migration: "(no migration)"
changelog:
  - 2026.03.10: created by merge of task-S6T09.dba.md and task-S6T09.specification.md
  - 2026.03.10: resolved Q-041 by splitting DB validation scope from rollout sign-off
---

# Task S6T09: Performance Validation & Quality Gates

## Sprint Context

- Sprint index: Sprint 6 Task Index
- Tags: database, database-refactor, sprint-6, task-index

## Source

- DBA source: `task-S6T09.dba.md`
- Specification source: `task-S6T09.specification.md`

## Resolved Decision

- S6T09 owns database-side performance validation and quality gates only.
- It does not own broader rollout sign-off, application-release approval, or infrastructure go-live coordination.
- All checks target canonical Sprint 4-6 stats objects and views only.

## Purpose & Scope

Defines the final database validation gate before the stats refactor can be considered technically ready.

In scope:

- EXPLAIN and latency checks for canonical UC views
- live trigger overhead checks on canonical write paths
- full backfill/orchestration duration checks
- canonical reconciliation success checks
- optional materialized-view freshness checks when S6T07 is implemented

Out of scope:

- API rollout approval
- RabbitMQ consumer rollout coordination
- cron or scheduler production enablement approvals
- dashboard product sign-off
- organizational go/no-go process

## Canonical Quality Gates

| Gate | Focus | Owner |
| ---- | ----- | ----- |
| QG-1 | canonical UC view latency | S6T09 |
| QG-2 | representative heavy-query plans | S6T09 |
| QG-3 | optional MV freshness and latency where implemented | S6T09 |
| QG-4 | live trigger overhead budget | S6T09 |
| QG-5 | full orchestration runtime window | S6T09 |
| QG-6 | no disallowed sequential full scans on canonical stats read paths | S6T09 |
| QG-7 | canonical reconciliation passes with zero mismatches | S6T09 |

## Rollout Sign-off Items Not Owned Here

These belong to a separate rollout activity outside S6T09:

- API endpoint contract verification against deployed application code
- message-bus consumer deployment and production queue validation
- scheduler activation for refresh or decay jobs
- monitoring, alerting, and incident-handbook approval
- production cutover approval and stakeholder sign-off

## Requirements

| ID      | Description                                                                  | MoSCoW |
| ------- | ---------------------------------------------------------------------------- | ------ |
| REQ-790 | Canonical UC views meet defined latency targets                              | MUST   |
| REQ-791 | Representative heavy queries use acceptable plans and avoid disallowed full scans | MUST |
| REQ-792 | Optional MVs, if implemented, meet freshness and latency targets             | MUST   |
| REQ-793 | Live trigger overhead stays within accepted DB-side budget                   | MUST   |
| REQ-794 | Full canonical orchestration completes within the target runtime window      | MUST   |
| REQ-795 | Reconciliation succeeds with zero mismatches                                | MUST   |
| REQ-796 | Task text clearly separates DB validation from rollout sign-off              | MUST   |

## Acceptance Criteria

| #   | Criterion                                         | How to Verify |
| --- | ------------------------------------------------- | ------------- |
| 1   | Canonical view targets are named consistently     | Inspect gate list |
| 2   | Rollout sign-off items are explicitly out of scope | Inspect out-of-scope list |
| 3   | Validation checks focus on canonical DB objects   | Review SQL and EXPLAIN targets |
| 4   | No stale noncanonical views or outbox tables drive the gates | Search task body |

## Validation Examples

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM stats.v_uc1_country_activity LIMIT 50;

EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM stats.v_uc2_user_network LIMIT 50;

EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM stats.v_uc15_distance_records LIMIT 50;

SELECT *
FROM stats.job_log
WHERE job_name = 'fn_run_all_snapshots'
ORDER BY completed_at DESC
LIMIT 1;
```

## Canonical Notes

- If S6T07 materialized views are not implemented, MV-specific gates are skipped rather than replaced with unrelated rollout checks.
- This task validates the database specification; it does not authorize production deployment on its own.
- Any merged references to noncanonical views like `v_uc1_global_kpi` or stale outbox-driven hot paths are obsolete.

## Implementation Checklist

- [ ] 1. Run EXPLAIN and latency checks against canonical S6T06 views
- [ ] 2. Validate trigger overhead on representative insert path
- [ ] 3. Confirm orchestration runtime meets target window
- [ ] 4. Confirm reconciliation finishes with zero mismatches
- [ ] 5. Record any follow-up rollout items outside this task

## Agent Loop Log

- 2026-03-10T21:05:00Z — `dba`: removed rollout-operational ownership from S6T09 and kept only database validation gates.
- 2026-03-10T21:05:00Z — `critical-thinking`: separated technical readiness evidence from organizational go-live approval to avoid false closure.
- 2026-03-10T21:05:00Z — `specification`: canonized the boundary between Sprint 6 performance validation and external rollout sign-off.

## Resolution

Q-041 is resolved by limiting S6T09 to database-side performance validation and quality gates only.
