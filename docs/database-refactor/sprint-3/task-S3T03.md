---
title: "Task S3T03: Create stats.user_countries Table"
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
task: S3T03
step: 3.3
migration: 20260310300200_create_user_countries.php
blocks: [5, 6]
changelog:
  - 2026-03-10: created by merge of 03-sprint-3-country-geography.md step 3.3
  - 2026-03-10: restored the missing Phinx migration and SQL usage examples sections
---

# Task S3T03: Create stats.user_countries Table

## Master-Spec Alignment

The normative contract for this sprint is [00-SPEC-DRAFT-v1.obsolete.md](../00-SPEC-DRAFT-v1.obsolete.md), Sections 5.3, 5.4, 8.4, 9.2, and 11.

- `stats.country_daily_stats.unique_users` and `unique_gks` are exact online-maintained values, not approximate placeholders.
- `INSERT`, `UPDATE`, and `DELETE` handling for `stats.gk_countries_visited`, `stats.user_countries`, and `stats.gk_country_history` must maintain exact state. When earliest/latest rows are invalidated, affected rows must be recomputed from remaining qualifying moves.
- Snapshot functions seed and verify canonical state; they do not compensate for knowingly inexact live maintenance.
- Any lower text that still describes `unique_users` or `unique_gks` as approximate is obsolete and superseded by this alignment block.

## Source

- Generated from sprint document step `3.3` in `03-sprint-3-country-geography.md`.

### Step 3.3: Create `stats.user_countries` Table

**What this step does:** Creates the `stats.user_countries` table that tracks which countries each user has interacted in, with move counts and first/last visit timestamps. This table supports user country coverage maps, diversity bonus tracking (gamification: +5 points for visiting a new country), and user geography leaderboards.

**Migration file name:** `20260310300200_create_user_countries.php`

#### Full SQL DDL

```sql
CREATE TABLE stats.user_countries (
  user_id INT NOT NULL,
  country_code CHAR(2) NOT NULL,
  move_count BIGINT NOT NULL DEFAULT 0,
  first_visit TIMESTAMPTZ NOT NULL,
  last_visit TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (user_id, country_code)
);

COMMENT ON TABLE stats.user_countries IS 'Tracks which countries each user has interacted in, with move counts and visit timestamps';
COMMENT ON COLUMN stats.user_countries.first_visit IS 'Timestamp of first move by this user in this country';
COMMENT ON COLUMN stats.user_countries.last_visit IS 'Timestamp of most recent move by this user in this country';
```

#### Full Phinx Migration PHP Code

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class CreateUserCountries extends AbstractMigration
{
  public function up(): void
  {
    $this->execute(<<<'SQL'
CREATE TABLE stats.user_countries (
  user_id INT NOT NULL,
  country_code CHAR(2) NOT NULL,
  move_count BIGINT NOT NULL DEFAULT 0,
  first_visit TIMESTAMPTZ NOT NULL,
  last_visit TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (user_id, country_code)
);

COMMENT ON TABLE stats.user_countries IS 'Tracks which countries each user has interacted in, with move counts and visit timestamps';
COMMENT ON COLUMN stats.user_countries.first_visit IS 'Timestamp of first move by this user in this country';
COMMENT ON COLUMN stats.user_countries.last_visit IS 'Timestamp of most recent move by this user in this country';
SQL
    );
  }

  public function down(): void
  {
    $this->execute(<<<'SQL'
DROP TABLE IF EXISTS stats.user_countries CASCADE;
SQL
    );
  }
}
```

#### SQL Usage Examples

```sql
-- Read the countries visited by a specific user
SELECT country_code, move_count, first_visit, last_visit
FROM stats.user_countries
WHERE user_id = 42
ORDER BY last_visit DESC;

-- Top 10 most geographically diverse users
SELECT user_id, COUNT(*) AS countries, SUM(move_count) AS total_moves
FROM stats.user_countries
GROUP BY user_id
ORDER BY countries DESC
LIMIT 10;

-- Users who have been active in both Poland and Czech Republic
SELECT uc1.user_id
FROM stats.user_countries uc1
JOIN stats.user_countries uc2 ON uc1.user_id = uc2.user_id
WHERE uc1.country_code = 'PL' AND uc2.country_code = 'CZ';

-- Users with 5+ distinct countries overall
SELECT user_id, COUNT(*) AS countries_visited
FROM stats.user_countries
GROUP BY user_id
HAVING COUNT(*) >= 5;
```

#### Graph/Visualization Specification

**Unlocked visualizations:**

- **Chart type:** User choropleth map — highlight countries visited by a specific user
- **Data source:** `SELECT country_code, move_count FROM stats.user_countries WHERE user_id = ?`
- **Color scale:** Sequential blue, intensity proportional to `move_count`

- **Chart type:** Bar chart — top users by country diversity
- **X-axis:** Username
- **Y-axis:** Country count

```
ASCII Sample (User Country Map):
User #42: PL(1520) DE(342) CZ(188) FR(95) SK(67) AT(34) ...
Total: 12 countries, 2246 moves
```

#### TimescaleDB Assessment

**NOT recommended.** This is a lookup/dimension table keyed by `(user_id, country_code)`. No suitable time column for hypertable partitioning. Expected row count: `unique_users × avg_countries_per_user` (~50K–200K rows). Standard PostgreSQL handles this efficiently.

#### pgTAP Unit Tests

| Test ID   | Test Name                       | Assertion                                                                           |
| --------- | ------------------------------- | ----------------------------------------------------------------------------------- |
| T-3.3.001 | user_countries table exists     | `has_table('stats', 'user_countries')`                                              |
| T-3.3.002 | PK is (user_id, country_code)   | `col_is_pk('stats', 'user_countries', ARRAY['user_id', 'country_code'])`            |
| T-3.3.003 | user_id type is integer         | `col_type_is('stats', 'user_countries', 'user_id', 'integer')`                      |
| T-3.3.004 | country_code type is char(2)    | `col_type_is('stats', 'user_countries', 'country_code', 'character(2)')`            |
| T-3.3.005 | move_count default is 0         | `col_default_is('stats', 'user_countries', 'move_count', '0')`                      |
| T-3.3.006 | first_visit type is timestamptz | `col_type_is('stats', 'user_countries', 'first_visit', 'timestamp with time zone')` |
| T-3.3.007 | last_visit type is timestamptz  | `col_type_is('stats', 'user_countries', 'last_visit', 'timestamp with time zone')`  |
| T-3.3.008 | Insert and read-back succeeds   | Insert `(42, 'PL', 5, now(), now())` and verify                                     |
| T-3.3.009 | Duplicate PK raises error       | Insert same `(user_id, country_code)` twice — `throws_ok`                           |

#### Implementation Checklist

- [ ] 1. Create migration file `20260310300200_create_user_countries.php`
- [ ] 2. Run `phinx migrate`
- [ ] 3. Verify table exists with 5 columns and correct types
- [ ] 4. Verify composite PK on `(user_id, country_code)`
- [ ] 5. Test insert and duplicate rejection
- [ ] 6. Run pgTAP tests T-3.3.001 through T-3.3.009

---
