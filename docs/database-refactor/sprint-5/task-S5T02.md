---
title: "Task S5T02: stats.country_pair_flows Table"
version: 1.0
date_created: 2026-03-10
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 5
tags:
  - analytics
  - country-pair-flows
  - database
  - database-refactor
  - dba
  - specification
  - sprint-5
  - sql
  - stats
  - table
  - task-index
  - task-merge
  - uc7
depends_on:
  - "Sprint 3 country tables"
task: S5T02
step: 5.2
migration: 20260310500100_create_country_pair_flows.php
blocks:
  - S5T07
  - S5T10
changelog:
  - 2026.03.10: created by merge of task-S5T02.dba.md and task-S5T02.specification.md
---

# Task S5T02: stats.country_pair_flows Table

## Sprint Context

- Sprint index: Sprint 5 Task Index
- Tags: database, database-refactor, sprint-5, task-index

## Source

- DBA source: `task-S5T02.dba.md`
- Specification source: `task-S5T02.specification.md`

## Purpose & Scope

Creates `stats.country_pair_flows`, which tracks how often GeoKrety travel **from one country to another** within a calendar month. A row `(2026-03-01, 'PL', 'DE', 47, 12)` means: "in March 2026, 47 qualifying moves crossed from Poland to Germany, involving 12 distinct GeoKrety."

This powers:

- UC7: Country flow diagram / Sankey chart ("Where do GeoKrety travel to/from?")
- Cross-border movement analytics

**"From country"** = country of the previous qualifying move on the same GeoKret.
**"To country"** = country of the current qualifying move.
Same-country transitions (self-loops) are excluded (`from_country <> to_country`).

Monthly aggregate table tracking cross-country GeoKret movements. One row per `(month, from_country, to_country)` triplet. "From" is the country of the GK's previous qualifying move; "To" is the country of the current move. Enables the UC7 Sankey diagram showing where GeoKrety travel internationally.

**Scope:** DDL only. Trigger population in S5T07. Backfill in Sprint 6.

---

## Requirements

| ID      | Description                                                                           | MoSCoW |
| ------- | ------------------------------------------------------------------------------------- | ------ |
| REQ-530 | Table `stats.country_pair_flows` exists                                               | MUST   |
| REQ-531 | 3-part composite PK: `(year_month DATE, from_country CHAR(2), to_country CHAR(2))`    | MUST   |
| REQ-532 | `CHECK (from_country <> to_country)` — no self-loops                                  | MUST   |
| REQ-533 | `CHECK (from_country = UPPER(from_country))` — uppercase invariant                    | MUST   |
| REQ-534 | `CHECK (to_country = UPPER(to_country))` — uppercase invariant                        | MUST   |
| REQ-535 | `move_count BIGINT DEFAULT 0 NOT NULL`                                                | MUST   |
| REQ-536 | `unique_gk_count BIGINT DEFAULT 0 NOT NULL`                                           | MUST   |
| REQ-537 | `year_month` represents the first day of the month (convention, enforced by inserter) | MUST   |
| REQ-538 | Table is empty after DDL creation                                                     | MUST   |
| REQ-539 | `phinx rollback` drops table cleanly                                                  | MUST   |

---

## Acceptance Criteria

| #   | Criterion                       | How to Verify                                     |
| --- | ------------------------------- | ------------------------------------------------- |
| 1   | Table created in `stats` schema | `\d stats.country_pair_flows`                     |
| 2   | 3-part composite PK             | PK visible in description                         |
| 3   | Self-loop rejected by CHECK     | `INSERT ('2026-01-01','PL','PL',0,0)` → exception |
| 4   | Lowercase country code rejected | `INSERT ('2026-01-01','pl','DE',0,0)` → exception |
| 5   | Both count columns default to 0 | INSERT without counts → both = 0                  |
| 6   | Table empty after DDL           | 0 rows                                            |
| 7   | Rollback drops table            | Table absent                                      |

---

## Migration File

**`20260310500100_create_country_pair_flows.php`**

## Full SQL DDL

```sql
CREATE TABLE stats.country_pair_flows (
  year_month        DATE     NOT NULL,
  from_country      CHAR(2)  NOT NULL,
  to_country        CHAR(2)  NOT NULL,
  move_count        BIGINT   NOT NULL DEFAULT 0,
  unique_gk_count   BIGINT   NOT NULL DEFAULT 0,
  PRIMARY KEY (year_month, from_country, to_country),
  CONSTRAINT chk_cpf_different_countries
    CHECK (from_country <> to_country),
  CONSTRAINT chk_cpf_from_country_upper
    CHECK (from_country = UPPER(from_country)),
  CONSTRAINT chk_cpf_to_country_upper
    CHECK (to_country = UPPER(to_country))
);

COMMENT ON TABLE stats.country_pair_flows
  IS 'Monthly GK flow counts between country pairs; powers UC7 Sankey/flow chart';
COMMENT ON COLUMN stats.country_pair_flows.year_month
  IS 'First day of the month (e.g. 2026-03-01)';
COMMENT ON COLUMN stats.country_pair_flows.from_country
  IS 'ISO 3166-1 alpha-2 country code of origin (uppercase)';
COMMENT ON COLUMN stats.country_pair_flows.to_country
  IS 'ISO 3166-1 alpha-2 country code of destination (uppercase)';
COMMENT ON COLUMN stats.country_pair_flows.move_count
  IS 'Number of qualifying moves this month that crossed from→to';
COMMENT ON COLUMN stats.country_pair_flows.unique_gk_count
  IS 'Distinct GeoKrety that crossed from→to at least once this month';
```

## Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateCountryPairFlows extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.country_pair_flows (
  year_month        DATE     NOT NULL,
  from_country      CHAR(2)  NOT NULL,
  to_country        CHAR(2)  NOT NULL,
  move_count        BIGINT   NOT NULL DEFAULT 0,
  unique_gk_count   BIGINT   NOT NULL DEFAULT 0,
  PRIMARY KEY (year_month, from_country, to_country),
  CONSTRAINT chk_cpf_different_countries CHECK (from_country <> to_country),
  CONSTRAINT chk_cpf_from_country_upper  CHECK (from_country = UPPER(from_country)),
  CONSTRAINT chk_cpf_to_country_upper    CHECK (to_country = UPPER(to_country))
);

COMMENT ON TABLE stats.country_pair_flows
  IS 'Monthly GK flow counts between country pairs; powers UC7 Sankey/flow chart';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.country_pair_flows;');
    }
}
```

## Data Contract

| Column            | Type      | Nullable | Default | Description                            |
| ----------------- | --------- | -------- | ------- | -------------------------------------- |
| `year_month`      | `DATE`    | NOT NULL | —       | **PK (part 1)** — First day of month   |
| `from_country`    | `CHAR(2)` | NOT NULL | —       | **PK (part 2)** — Origin ISO code      |
| `to_country`      | `CHAR(2)` | NOT NULL | —       | **PK (part 3)** — Destination ISO code |
| `move_count`      | `BIGINT`  | NOT NULL | `0`     | Qualifying moves crossing this pair    |
| `unique_gk_count` | `BIGINT`  | NOT NULL | `0`     | Distinct GeoKrety crossing this pair   |

**Constraints:**

- `chk_cpf_different_countries`: `from_country <> to_country` (no self-loops)
- `chk_cpf_from_country_upper`: country codes stored uppercase
- `chk_cpf_to_country_upper`: country codes stored uppercase
- `year_month` = first day of month (enforced by application/trigger logic using `DATE_TRUNC('month', move_date)`)

## SQL Usage Examples

```sql
-- UC7: Top flows in last 12 months
SELECT from_country, to_country, SUM(move_count) AS total
FROM stats.country_pair_flows
WHERE year_month >= DATE_TRUNC('month', NOW() - INTERVAL '12 months')
GROUP BY from_country, to_country
ORDER BY total DESC
LIMIT 50;

-- UC7: All flows for a specific country
SELECT
  CASE WHEN from_country = 'PL' THEN to_country ELSE from_country END AS partner,
  CASE WHEN from_country = 'PL' THEN 'outflow' ELSE 'inflow' END AS direction,
  SUM(move_count) AS total
FROM stats.country_pair_flows
WHERE (from_country = 'PL' OR to_country = 'PL')
  AND year_month >= NOW() - INTERVAL '12 months'
GROUP BY 1, 2
ORDER BY total DESC;

-- Monthly trend for PL→DE
SELECT year_month, move_count, unique_gk_count
FROM stats.country_pair_flows
WHERE from_country = 'PL' AND to_country = 'DE'
ORDER BY year_month DESC;
```

## Graph / Visualization Specification

**UC7: Sankey diagram (country flows)**

```
Data query:
SELECT from_country, to_country, SUM(move_count) AS total_flows
FROM stats.country_pair_flows
WHERE year_month >= DATE_TRUNC('month', NOW() - INTERVAL '12 months')
GROUP BY from_country, to_country
ORDER BY total_flows DESC
LIMIT 100;
```

```
ASCII Example:
Germany (DE) ──────────── 1,240 flows ──────────── Poland (PL)
Germany (DE) ────────────   890 flows ──────────── Czech (CZ)
Poland  (PL) ────────────   670 flows ──────────── Germany (DE)
```

## TimescaleDB Assessment

**CANDIDATE for hypertable** (time dimension: `year_month`). One row per month per country pair → low cardinality growth. Chunk by 6-month intervals.

## pgTAP Unit Tests

| Test ID   | Assertion                                                                                   | Expected  |
| --------- | ------------------------------------------------------------------------------------------- | --------- |
| T-5.2.001 | `has_table('stats', 'country_pair_flows')`                                                  | pass      |
| T-5.2.002 | `col_is_pk('stats', 'country_pair_flows', ARRAY['year_month','from_country','to_country'])` | pass      |
| T-5.2.003 | `col_type_is('stats', 'country_pair_flows', 'from_country', 'character')`                   | pass      |
| T-5.2.004 | `col_type_is('stats', 'country_pair_flows', 'move_count', 'bigint')`                        | pass      |
| T-5.2.005 | Same-country INSERT (`from_country = to_country`) → CHECK violation                         | exception |
| T-5.2.006 | Lowercase country INSERT → CHECK violation                                                  | exception |
| T-5.2.007 | Table is empty after creation                                                               | pass      |
| T-5.2.008 | `phinx rollback` drops table                                                                | pass      |

| Test ID   | Assertion                 | Pass Condition   |
| --------- | ------------------------- | ---------------- |
| T-5.2.001 | Table exists              | `has_table()`    |
| T-5.2.002 | 3-part PK                 | `col_is_pk()`    |
| T-5.2.003 | `from_country` is CHAR(2) | `col_type_is()`  |
| T-5.2.004 | `move_count` is BIGINT    | `col_type_is()`  |
| T-5.2.005 | Self-loop CHECK violation | Exception raised |
| T-5.2.006 | Lowercase CHECK violation | Exception raised |
| T-5.2.007 | Table empty               | `is_empty()`     |
| T-5.2.008 | Rollback drops table      | `hasnt_table()`  |

---

## Implementation Checklist

- [ ] 1. Create `20260310500100_create_country_pair_flows.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. `\d stats.country_pair_flows` — 5 columns, 3-part PK, 3 CHECKs
- [ ] 4. Test self-loop violation: `('2026-01-01','PL','PL',0,0)` → exception
- [ ] 5. Test lowercase: `('2026-01-01','pl','DE',0,0)` → exception
- [ ] 6. Run pgTAP T-5.2.001 through T-5.2.008 — all pass
- [ ] 7. `phinx rollback` — table gone

- [ ] 1. Write `20260310500100_create_country_pair_flows.php`
- [ ] 2. `phinx migrate` — no errors
- [ ] 3. 5 columns, 3-part PK, 3 CHECK constraints
- [ ] 4. Verify self-loop and lowercase violations
- [ ] 5. Run pgTAP T-5.2.001 through T-5.2.008 — all pass
- [ ] 6. `phinx rollback` — table gone

## Table Created

```
stats.country_pair_flows (year_month, from_country, to_country, move_count, unique_gk_count)
```

| Column            | Type    | Constraints                                      |
| ----------------- | ------- | ------------------------------------------------ |
| `year_month`      | DATE    | PK (part 1), NOT NULL                            |
| `from_country`    | CHAR(2) | PK (part 2), NOT NULL, uppercase, ≠ to_country   |
| `to_country`      | CHAR(2) | PK (part 3), NOT NULL, uppercase, ≠ from_country |
| `move_count`      | BIGINT  | NOT NULL, DEFAULT 0                              |
| `unique_gk_count` | BIGINT  | NOT NULL, DEFAULT 0                              |

---
