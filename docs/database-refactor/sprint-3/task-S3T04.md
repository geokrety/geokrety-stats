---
title: "Task S3T04: Create stats.gk_country_history Table"
version: 1.0
date_created: 2026-03-08
last_updated: 2026-03-10
owner: "GeoKrety Community"
sprint: 3
tags:
  - country
  - database
  - geography
  - migration
  - postgresql
  - schema
  - sprint-3
  - stats
  - task-merge
  - traversal
depends_on: [1, 2]
task: S3T04
step: 3.4
migration: 20260310300300_create_gk_country_history.php
blocks: [5, 6]
changelog:
  - 2026-03-10: created by merge of 03-sprint-3-country-geography.md step 3.4
  - 2026-03-10: added an explicit btree_gist prerequisite guard to the DDL and migration
---

# Task S3T04: Create stats.gk_country_history Table

## Master-Spec Alignment

The normative contract for this sprint is [00-SPEC-DRAFT-v1.obsolete.md](00-SPEC-DRAFT-v1.obsolete.md), Sections 5.3, 5.4, 8.4, 9.2, and 11.

- `stats.country_daily_stats.unique_users` and `unique_gks` are exact online-maintained values, not approximate placeholders.
- `INSERT`, `UPDATE`, and `DELETE` handling for `stats.gk_countries_visited`, `stats.user_countries`, and `stats.gk_country_history` must maintain exact state. When earliest/latest rows are invalidated, affected rows must be recomputed from remaining qualifying moves.
- Snapshot functions seed and verify canonical state; they do not compensate for knowingly inexact live maintenance.
- Any lower text that still describes `unique_users` or `unique_gks` as approximate is obsolete and superseded by this alignment block.

## Source

- Generated from sprint document step `3.4` in `03-sprint-3-country-geography.md`.

### Step 3.4: Create `stats.gk_country_history` Table

**What this step does:** Creates the `stats.gk_country_history` table that maintains temporal intervals of GK presence in countries. Each row records when a GK arrived in a country and when it departed (NULL if still present). A GiST exclusion constraint prevents overlapping intervals for the same GK, ensuring data integrity. This table enables "current country" lookups, dwell-time analytics, and country transition timelines without scanning the full move history.

**Important:** This table requires the `btree_gist` extension (enabled in Sprint 1, Step 1.7) for the exclusion constraint.

**Migration file name:** `20260310300300_create_gk_country_history.php`

#### Full SQL DDL

```sql
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE stats.gk_country_history (
  id BIGSERIAL PRIMARY KEY,
  geokrety_id INT NOT NULL,
  country_code CHAR(2) NOT NULL,
  arrived_at TIMESTAMPTZ NOT NULL,
  departed_at TIMESTAMPTZ,
  move_id BIGINT NOT NULL,
  CONSTRAINT gk_country_history_excl
    EXCLUDE USING gist (
      geokrety_id WITH =,
      tstzrange(arrived_at, COALESCE(departed_at, 'infinity')) WITH &&
    )
);

COMMENT ON TABLE stats.gk_country_history IS 'Temporal intervals of GK presence in countries; exclusion constraint prevents overlapping intervals per GK';
COMMENT ON COLUMN stats.gk_country_history.departed_at IS 'NULL means the GK is currently in this country (open interval)';
COMMENT ON COLUMN stats.gk_country_history.move_id IS 'Move ID that caused the GK to arrive in this country';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkCountryHistory extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE stats.gk_country_history (
  id BIGSERIAL PRIMARY KEY,
  geokrety_id INT NOT NULL,
  country_code CHAR(2) NOT NULL,
  arrived_at TIMESTAMPTZ NOT NULL,
  departed_at TIMESTAMPTZ,
  move_id BIGINT NOT NULL,
  CONSTRAINT gk_country_history_excl
    EXCLUDE USING gist (
      geokrety_id WITH =,
      tstzrange(arrived_at, COALESCE(departed_at, 'infinity')) WITH &&
    )
);

COMMENT ON TABLE stats.gk_country_history IS 'Temporal intervals of GK presence in countries; exclusion constraint prevents overlapping intervals per GK';
COMMENT ON COLUMN stats.gk_country_history.departed_at IS 'NULL means the GK is currently in this country (open interval)';
COMMENT ON COLUMN stats.gk_country_history.move_id IS 'Move ID that caused the GK to arrive in this country';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.gk_country_history;');
    }
}
```

#### SQL Usage Examples

```sql
-- Current country for GK #1
SELECT country_code, arrived_at
FROM stats.gk_country_history
WHERE geokrety_id = 1 AND departed_at IS NULL;

-- Full country timeline for GK #1
SELECT country_code, arrived_at, departed_at,
       COALESCE(departed_at, now()) - arrived_at AS dwell_time
FROM stats.gk_country_history
WHERE geokrety_id = 1
ORDER BY arrived_at ASC;

-- All GKs currently in Poland
SELECT geokrety_id, arrived_at
FROM stats.gk_country_history
WHERE country_code = 'PL' AND departed_at IS NULL
ORDER BY arrived_at DESC;

-- Average dwell time per country (closed intervals only)
SELECT country_code,
       AVG(EXTRACT(EPOCH FROM (departed_at - arrived_at)) / 86400)::NUMERIC(8,1) AS avg_days
FROM stats.gk_country_history
WHERE departed_at IS NOT NULL
GROUP BY country_code
ORDER BY avg_days DESC
LIMIT 10;

-- Verify exclusion constraint: overlapping intervals rejected
-- This should FAIL with exclusion violation:
-- INSERT INTO stats.gk_country_history (geokrety_id, country_code, arrived_at, departed_at, move_id)
-- VALUES (1, 'PL', '2025-01-01', '2025-06-01', 100);
-- INSERT INTO stats.gk_country_history (geokrety_id, country_code, arrived_at, departed_at, move_id)
-- VALUES (1, 'DE', '2025-03-01', '2025-09-01', 200);  -- overlaps!
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** GK timeline — horizontal bar chart showing country intervals
- **X-axis:** Time (arrived_at to departed_at or now)
- **Y-axis:** Country code
- **Color:** One color per country

- **Chart type:** Country dwell-time histogram
- **X-axis:** Dwell time buckets (days)
- **Y-axis:** Count of intervals

```
ASCII Sample (GK #1 Country Timeline):
PL   |████████████|          |████████████████████████|
DE               |██████████|
CZ                                                     |██████████|
     Jan 2025   Mar 2025   Jun 2025   Sep 2025   Dec 2025
```

#### TimescaleDB Assessment

**POSSIBLE but NOT recommended now.** Rationale:

- `gk_country_history` has a time dimension (`arrived_at`) and grows proportionally to country transitions (~2 rows per GK per cross-border move). Estimated volume: 500K–2M rows.
- HyperTable conversion would support time-range pruning and automatic retention policies.
- **However**, the GiST exclusion constraint (`gk_country_history_excl`) is not compatible with TimescaleDB hypertables, as exclusion constraints cannot span chunks. This is a blocking limitation.
- **Recommendation:** Use standard PostgreSQL table. The exclusion constraint is more valuable than hypertable benefits at this scale.

#### pgTAP Unit Tests

| Test ID   | Test Name                                             | Assertion                                                                                                                            |
| --------- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| T-3.4.001 | gk_country_history table exists                       | `has_table('stats', 'gk_country_history')`                                                                                           |
| T-3.4.002 | PK is id                                              | `col_is_pk('stats', 'gk_country_history', 'id')`                                                                                     |
| T-3.4.003 | id column is bigserial                                | `col_type_is('stats', 'gk_country_history', 'id', 'bigint')`                                                                         |
| T-3.4.004 | geokrety_id type is integer                           | `col_type_is('stats', 'gk_country_history', 'geokrety_id', 'integer')`                                                               |
| T-3.4.005 | country_code type is char(2)                          | `col_type_is('stats', 'gk_country_history', 'country_code', 'character(2)')`                                                         |
| T-3.4.006 | arrived_at is NOT NULL                                | `col_not_null('stats', 'gk_country_history', 'arrived_at')`                                                                          |
| T-3.4.007 | departed_at is nullable                               | `col_is_null('stats', 'gk_country_history', 'departed_at')`                                                                          |
| T-3.4.008 | move_id is NOT NULL                                   | `col_not_null('stats', 'gk_country_history', 'move_id')`                                                                             |
| T-3.4.009 | Exclusion constraint exists                           | `SELECT COUNT(*) = 1 FROM pg_constraint WHERE conname = 'gk_country_history_excl' AND contype = 'x'`                                 |
| T-3.4.010 | Non-overlapping insert succeeds                       | Insert `(1, 'PL', '2025-01-01', '2025-06-01', 100)` then `(1, 'DE', '2025-06-01', NULL, 200)` — `lives_ok`                           |
| T-3.4.011 | Overlapping insert fails                              | Insert `(1, 'PL', '2025-01-01', '2025-06-01', 100)` then `(1, 'CZ', '2025-03-01', '2025-09-01', 300)` — `throws_ok`                  |
| T-3.4.012 | Different GKs can overlap                             | Insert `(1, 'PL', '2025-01-01', NULL, 100)` and `(2, 'PL', '2025-01-01', NULL, 200)` — `lives_ok`                                    |
| T-3.4.013 | Open interval (departed_at NULL) blocks later overlap | Insert `(1, 'PL', '2025-01-01', NULL, 100)` then `(1, 'DE', '2025-06-01', NULL, 200)` — `throws_ok` (overlapping with open interval) |

#### Implementation Checklist

- [ ] 1. Verify `btree_gist` extension is enabled (Sprint 1, Step 1.7)
- [ ] 2. Create migration file `20260310300300_create_gk_country_history.php`
- [ ] 3. Run `phinx migrate`
- [ ] 4. Verify table exists with 6 columns and correct types
- [ ] 5. Verify exclusion constraint `gk_country_history_excl` exists
- [ ] 6. Test non-overlapping intervals succeed
- [ ] 7. Test overlapping intervals are rejected
- [ ] 8. Test different GKs can have overlapping intervals
- [ ] 9. Run pgTAP tests T-3.4.001 through T-3.4.013

---
