---
title: "Task S3T02: Create stats.gk_countries_visited Table"
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
task: S3T02
step: 3.2
migration: 20260310300100_create_gk_countries_visited.php
blocks: [5, 6]
changelog:
  - 2026-03-10: created by merge of 03-sprint-3-country-geography.md step 3.2
---

# Task S3T02: Create stats.gk_countries_visited Table

## Master-Spec Alignment

The normative contract for this sprint is [00-SPEC-DRAFT-v1.obsolete.md](00-SPEC-DRAFT-v1.obsolete.md), Sections 5.3, 5.4, 8.4, 9.2, and 11.

- `stats.country_daily_stats.unique_users` and `unique_gks` are exact online-maintained values, not approximate placeholders.
- `INSERT`, `UPDATE`, and `DELETE` handling for `stats.gk_countries_visited`, `stats.user_countries`, and `stats.gk_country_history` must maintain exact state. When earliest/latest rows are invalidated, affected rows must be recomputed from remaining qualifying moves.
- Snapshot functions seed and verify canonical state; they do not compensate for knowingly inexact live maintenance.
- Any lower text that still describes `unique_users` or `unique_gks` as approximate is obsolete and superseded by this alignment block.

## Source

- Generated from sprint document step `3.2` in `03-sprint-3-country-geography.md`.

### Step 3.2: Create `stats.gk_countries_visited` Table

**What this step does:** Creates the `stats.gk_countries_visited` table that tracks which countries each GeoKret has visited, including first-visit metadata and move counts. This table enables "countries visited" badges, GK travel maps, and country-crossing detection for the gamification multiplier system (+0.05 per new country).

**Migration file name:** `20260310300100_create_gk_countries_visited.php`

#### Full SQL DDL

```sql
CREATE TABLE stats.gk_countries_visited (
  geokrety_id INT NOT NULL,
  country_code CHAR(2) NOT NULL,
  first_visited_at TIMESTAMPTZ NOT NULL,
  first_move_id BIGINT NOT NULL,
  move_count INT NOT NULL DEFAULT 1,
  PRIMARY KEY (geokrety_id, country_code)
);

COMMENT ON TABLE stats.gk_countries_visited IS 'Tracks which countries each GK has visited, with first-visit metadata and move counts';
COMMENT ON COLUMN stats.gk_countries_visited.first_move_id IS 'ID of the first move that placed this GK in this country';
COMMENT ON COLUMN stats.gk_countries_visited.move_count IS 'Total number of moves by this GK in this country';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateGkCountriesVisited extends AbstractMigration
{
    public function up(): void
    {
        $this->execute(<<<'SQL'
CREATE TABLE stats.gk_countries_visited (
  geokrety_id INT NOT NULL,
  country_code CHAR(2) NOT NULL,
  first_visited_at TIMESTAMPTZ NOT NULL,
  first_move_id BIGINT NOT NULL,
  move_count INT NOT NULL DEFAULT 1,
  PRIMARY KEY (geokrety_id, country_code)
);

COMMENT ON TABLE stats.gk_countries_visited IS 'Tracks which countries each GK has visited, with first-visit metadata and move counts';
COMMENT ON COLUMN stats.gk_countries_visited.first_move_id IS 'ID of the first move that placed this GK in this country';
COMMENT ON COLUMN stats.gk_countries_visited.move_count IS 'Total number of moves by this GK in this country';
SQL
        );
    }

    public function down(): void
    {
        $this->execute('DROP TABLE IF EXISTS stats.gk_countries_visited;');
    }
}
```

#### SQL Usage Examples

```sql
-- How many countries has GK #1 visited?
SELECT COUNT(*) AS countries_visited
FROM stats.gk_countries_visited
WHERE geokrety_id = 1;

-- Travel history for GK #1 in country visit order
SELECT country_code, first_visited_at, first_move_id, move_count
FROM stats.gk_countries_visited
WHERE geokrety_id = 1
ORDER BY first_visited_at ASC;

-- Top 10 most-traveled GKs by country count
SELECT geokrety_id, COUNT(*) AS countries
FROM stats.gk_countries_visited
GROUP BY geokrety_id
ORDER BY countries DESC
LIMIT 10;

-- GKs that have visited both Poland and Germany
SELECT v1.geokrety_id
FROM stats.gk_countries_visited v1
JOIN stats.gk_countries_visited v2 ON v1.geokrety_id = v2.geokrety_id
WHERE v1.country_code = 'PL' AND v2.country_code = 'DE';
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** Bar chart — top GKs by country count
- **X-axis:** GK tracking code (GK + hex ID)
- **Y-axis:** `COUNT(*)` of countries visited

- **Chart type:** Route map — GK travel path across countries in order of `first_visited_at`
- **Data source:** `SELECT country_code, first_visited_at FROM stats.gk_countries_visited WHERE geokrety_id = ? ORDER BY first_visited_at`

```
ASCII Sample (GK Travel Path):
GK0001: PL → DE → CZ → AT → CH → FR → ES  (7 countries)
GK0042: PL → SK → HU → RO → BG → GR        (6 countries)
```

Route-map visualizations should present the ordered unique-country visitation sequence derived from `stats.gk_countries_visited`; repeated later returns do not create duplicate country rows in this table.

#### TimescaleDB Assessment

**NOT recommended.** This is a lookup/dimension table keyed by `(geokrety_id, country_code)`. There is no time column suitable for hypertable partitioning. Row count grows proportionally to `unique_GKs × unique_countries_per_GK` (~200K–500K rows), well within standard PostgreSQL capabilities.

#### pgTAP Unit Tests

| Test ID   | Test Name                            | Assertion                                                                                      |
| --------- | ------------------------------------ | ---------------------------------------------------------------------------------------------- |
| T-3.2.001 | gk_countries_visited table exists    | `has_table('stats', 'gk_countries_visited')`                                                   |
| T-3.2.002 | PK is (geokrety_id, country_code)    | `col_is_pk('stats', 'gk_countries_visited', ARRAY['geokrety_id', 'country_code'])`             |
| T-3.2.003 | geokrety_id type is integer          | `col_type_is('stats', 'gk_countries_visited', 'geokrety_id', 'integer')`                       |
| T-3.2.004 | country_code type is char(2)         | `col_type_is('stats', 'gk_countries_visited', 'country_code', 'character(2)')`                 |
| T-3.2.005 | first_visited_at type is timestamptz | `col_type_is('stats', 'gk_countries_visited', 'first_visited_at', 'timestamp with time zone')` |
| T-3.2.006 | first_move_id is NOT NULL            | `col_not_null('stats', 'gk_countries_visited', 'first_move_id')`                               |
| T-3.2.007 | move_count default is 1              | `col_default_is('stats', 'gk_countries_visited', 'move_count', '1')`                           |
| T-3.2.008 | Insert and read-back succeeds        | Insert `(1, 'PL', now(), 100, 1)` and verify                                                   |
| T-3.2.009 | Duplicate PK raises error            | Insert same `(geokrety_id, country_code)` twice — `throws_ok`                                  |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310300100_create_gk_countries_visited.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify table exists with 5 columns and correct types
- [ ] 4. Verify composite PK on `(geokrety_id, country_code)`
- [ ] 5. Test insert and duplicate rejection
- [ ] 6. Run pgTAP tests T-3.2.001 through T-3.2.009

---
