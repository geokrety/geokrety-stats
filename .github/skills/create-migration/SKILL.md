---
name: create-migration
description: 'Create Phinx PHP database migrations with exhaustive pgTAP SQL test suites for the GeoKrety website database. Use when asked to "create a migration", "create a phinx migration", "create a phinx migration with tests", "add a database migration", "add a column", "create a table", "add a trigger", "create a database change", "write db migration tests", or any variant of database schema evolution tasks. Generates migration files following existing patterns, validates PHP syntax, creates comprehensive pgTAP tests targeting ~100% coverage and edge cases, then runs a DBA + critical-thinking review loop before applying. Works together with the phinx and pgtap skills.'
user-invocable: true
---

# Create Migration Skill

## Purpose

Scaffold, validate, review, and apply database migrations for the GeoKrety website with full pgTAP test coverage. This skill orchestrates the entire lifecycle: file creation → syntax check → test creation → expert review loop → apply → verify `down()` reversibility → run tests.

## When to Use This Skill

- User asks to "create a migration", "add a column", "create a table", "add a trigger", "write a database change", "scaffold a phinx migration"
- Any task that involves evolving the GeoKrety website PostgreSQL schema
- When exhaustive regression tests are needed alongside a schema change

## Prerequisites

- The GeoKrety website stack must be running (Docker Compose)
- PHP CLI available (for `php -l` syntax checks)
- `pg_prove` / pgTAP installed and accessible via the `pgtap` skill wrapper
- Phinx accessible via the `phinx` skill wrapper
- Read both sibling skills before starting:
  - `.github/skills/phinx/SKILL.md`
  - `.github/skills/pgtap/SKILL.md`

## Canonical Paths

| Resource | Path |
|----------|------|
| Migrations directory | `/home/kumy/GIT/geokrety-website/website/db/migrations/` |
| Tests directory | `/home/kumy/GIT/geokrety-website/website/db/tests/` |
| Phinx skill wrapper | `/home/kumy/GIT/geokrety-stats/.github/skills/phinx/scripts/phinx.sh` |
| pgTAP skill wrapper | `/home/kumy/GIT/geokrety-stats/.github/skills/pgtap/scripts/pgtap.sh` |

---

## Step-by-Step Workflow

### Step 1 — Inspect Existing Patterns

Before writing anything, read several recent migration files and test files to understand project conventions:

```bash
ls /home/kumy/GIT/geokrety-website/website/db/migrations/ | tail -10
ls /home/kumy/GIT/geokrety-website/website/db/tests/
```

Key patterns to absorb:
- PHP `declare(strict_types=1)` mandatory at the top
- Class extends `Phinx\Migration\AbstractMigration`; class is `final`
- Class name = CamelCase of filename suffix only — strip the timestamp prefix
  - `20260220120000_geokret_loves.php` → class `GeokretLoves`
- Always implement both `up()` **and** `down()` (use `change()` only for trivially-reversible single operations)
- All tables use schema prefix `geokrety.` (e.g. `geokrety.gk_loves`)
- Multi-line SQL uses PHP **single-quoted heredoc** `<<<'EOL' ... EOL` to prevent PHP variable interpolation; do NOT use double-quoted strings for SQL — see Security note in Step 3
- **Trigger naming**: look at nearby existing triggers before choosing a name. Two conventions coexist:
  - Internal counter/logic triggers: `after_<table_suffix>_<description>` (e.g. `after_gk_loves_update_count`)
  - AMQP notification triggers: `after_<NN>_notify_amqp_<topic>` where `NN` is a priority number (e.g. `after_99_notify_amqp_loves`)
  - Match whichever convention existing sibling triggers on the same table use
- Counter columns use `GREATEST(0, count - 1)` pattern in DELETE branches to prevent negative values
- Index names are explicit and descriptive: `idx_<table_suffix>_<columns>`
- FK actions always stated explicitly: `['delete' => 'CASCADE', 'update' => 'NO_ACTION']`

### Step 2 — Generate Timestamp and Filename

Generate a timestamp in format `YYYYMMDDHHmmSS` using current UTC time.

**Collision check** — run this before finalising the filename:

```bash
ls /home/kumy/GIT/geokrety-website/website/db/migrations/ | grep "^$(date -u +%Y%m%d)"
```

If another migration was created the same day at the same second, increment the seconds field by 1 until no collision exists (see `20260220130000` and `20260220130100` — 60 s apart, intentional).

Construct:
- **Migration file**: `YYYYMMDDHHMMSS_<snake_case_description>.php`
- **Test file**: `test-<NNN>-<description>.sql` — pick the NNN block matching the primary table being changed (see NNN Block Map below)

**PHP class name collision check** — after choosing the class name, verify uniqueness across all existing migrations:

```bash
grep -rh "^final class " /home/kumy/GIT/geokrety-website/website/db/migrations/ | sort | uniq -d
```

If your intended class name appears in the output, append a date suffix (e.g. `GeokretLovesV2`) and note the deviation in a comment.

### Step 3 — Write the Migration File

Create the `.php` migration file in `/home/kumy/GIT/geokrety-website/website/db/migrations/`.

**Security note**: Never use PHP double-quoted strings for SQL executed via `$this->execute()`. PHP variable interpolation inside double-quoted strings can introduce arbitrary SQL if a variable name appears accidentally in the SQL text. Always use **single-quoted heredoc** (`<<<'EOL'`) for multi-line SQL.

**Template structure:**

```php
<?php

declare(strict_types=1);

use Phinx\Migration\AbstractMigration;

final class <CamelCaseName> extends AbstractMigration {
    public function up(): void {
        // Descriptive comment explaining what this block does and WHY
        // ... Phinx table/column/index/FK operations OR $this->execute(<<<'EOL' ... EOL);
    }

    public function down(): void {
        // Exact reverse of up() — every object created must be dropped/removed
        // Drop order: RLS policies → materialized views → views → triggers →
        //             trigger functions → regular functions → explicit indexes →
        //             columns → tables → sequences
    }
}
```

**`up()` checklist:**
- [ ] `declare(strict_types=1)` present
- [ ] Class is `final` and extends `AbstractMigration`
- [ ] Schema prefix `geokrety.` used on every table reference
- [ ] Multi-line SQL uses `<<<'EOL' ... EOL` single-quoted heredoc
- [ ] Index names are explicit and follow `idx_<table_suffix>_<columns>` convention
- [ ] FK actions explicit — both `delete` and `update` keys present
- [ ] Trigger name follows the appropriate convention for the table (see Step 1)
- [ ] Counter triggers use `GREATEST(0, count - 1)` in DELETE branch
- [ ] Comments explain the *why*, not just the *what*

**`down()` checklist:**
- [ ] Drops every object created in `up()` — no orphans
- [ ] Drop order respected: RLS policies → materialized views → views → triggers → trigger functions → regular functions → explicit indexes → columns → tables → sequences
- [ ] Uses `DROP … IF EXISTS` for resilience against partial failure
- [ ] Does not silently discard data without a comment acknowledging the data loss
- [ ] If a PostgreSQL function signature changes, explicitly `DROP FUNCTION IF EXISTS` the old signature first — `CREATE OR REPLACE FUNCTION` does not replace sibling overloads

**Partial failure note**: PostgreSQL uses transactional DDL — if any statement in `up()` throws inside Phinx's implicit transaction, the entire migration is rolled back and the schema is clean. However, `CREATE INDEX CONCURRENTLY` and `REFRESH MATERIALIZED VIEW CONCURRENTLY` cannot run inside a transaction. If you need these, call `$this->execute('COMMIT;')` before the concurrent operation and `$this->execute('BEGIN;')` after, and document this in a comment.

### Step 4 — PHP Syntax Check

```bash
php -l /home/kumy/GIT/geokrety-website/website/db/migrations/<filename>.php
```

Fix any reported errors before proceeding. Iterate until output is `No syntax errors detected in ...`.

Also verify the class name matches the filename: extract the CamelCase suffix from the filename and confirm it equals the `final class` name.

After `php -l` completes successfully, run the repository hooks to fix formatting and lint issues:

```bash
pre-commit run -a
```

This applies automatic fixes (formatters/linters) before proceeding.

### Step 5 — Write Exhaustive pgTAP Tests

Create the `.sql` test file in `/home/kumy/GIT/geokrety-website/website/db/tests/`.

**Test plan strategy**: Use `SELECT * FROM no_plan();` during initial drafting so you are not blocked by a count mismatch. Only once the test file is complete, replace with `SELECT plan(N);` where N is the exact assertion count. Count with:

```bash
grep -cP '^\s*SELECT (lives_ok|throws_ok|is|isnt|ok|has_|col_|index_|row_eq|results_eq|set_eq|bag_eq|matches|imatches|alike|unalike|passes|fail|cmp_ok|throws_like|like)' test-NNN-description.sql
```

Each `SELECT` returning a single pgTAP assertion = 1 plan unit. SQL inside `DO $$ ... $$` blocks does **not** auto-count.

**Template structure:**

```sql
-- Start transaction and plan the tests.
BEGIN;
SELECT plan(<N>);   -- Replace N once file is complete; use no_plan() during drafting

-- \set declarations for reusable values
\set nice '\'0101000020E6100000F6285C8FC2F51C405C8FC2F528DC4540\''

-- Setup: minimal fixture data
-- IMPORTANT: always use explicit `id` values — never rely on DEFAULT serial.
-- Sequences advance even inside rolled-back transactions; using DEFAULT
-- causes non-deterministic ID values in subsequent test runs.
INSERT INTO "gk_users" ("id", "username", "registration_ip") VALUES (1, 'test 1', '127.0.0.1');
-- ... other fixtures

-- ── GROUP 1: Schema existence ──────────────────────────────────────────────
-- For new tables:
SELECT has_table('geokrety', 'gk_<new_table>', 'table gk_<new_table> exists');
-- For new columns on existing tables:
SELECT has_column('geokrety', 'gk_<table>', '<column>', 'column <column> exists');
SELECT col_not_null('geokrety', 'gk_<table>', '<column>', '<column> is NOT NULL');
SELECT col_default_is('geokrety', 'gk_<table>', '<column>', '<default>', '<column> default');
-- For indexes:
SELECT has_index('geokrety', 'gk_<table>'::name, 'idx_<name>', ARRAY['col1', 'col2']);
-- For triggers (schema, table, trigger_name):
SELECT has_trigger('geokrety', 'gk_<table>', '<trigger_name>', '<trigger_name> trigger exists');
-- For functions (schema, function_name):
SELECT has_function('geokrety', '<function_name>', '<function_name> function exists');

-- ── GROUP 2: Happy-path inserts ────────────────────────────────────────────
SELECT lives_ok($$INSERT INTO ...$$, 'description of what is being tested');

-- ── GROUP 3: Constraint violations ────────────────────────────────────────
SELECT throws_ok($$INSERT INTO ... duplicate ...$$, '23505', NULL, 'UNIQUE violation');
SELECT throws_ok($$INSERT INTO ... null not_null_col ...$$, '23502', NULL, 'NOT NULL violation');
SELECT throws_ok($$INSERT INTO ... FK nonexistent ...$$, '23503', NULL, 'FK violation');
SELECT throws_ok($$INSERT INTO ... CHECK violation ...$$, '23514', NULL, 'CHECK violation');
SELECT throws_ok($$INSERT INTO ... too long string ...$$, '22001', NULL, 'value too long');

-- ── GROUP 4: Trigger / counter behavior ───────────────────────────────────
-- Cover ALL branches of every trigger function:
-- INSERT branch: insert → verify side-effect (e.g. counter = 1)
-- UPDATE branch — watched column changes: verify side-effect fires
-- UPDATE branch — unwatched column changes: verify no side-effect (no-op)
-- UPDATE branch — NULL→value and value→NULL (if function is NULL-sensitive)
-- DELETE branch: delete → verify counter decremented
--   → also verify GREATEST(0) floor: delete when count=0, confirm still 0

-- ── GROUP 5: CASCADE / ON DELETE behavior ─────────────────────────────────
-- Delete parent → verify FK-cascaded child is gone (0 rows)
-- Insert with nonexistent FK parent → verify 23503

-- ── GROUP 6: Edge cases ───────────────────────────────────────────────────
-- NULL in nullable columns; empty strings; boundary timestamps;
-- max-length varchar; duplicate op idempotency where applicable

-- NOTE: down() reversibility is verified by the Step 7b rollback cycle,
-- NOT by this transaction's ROLLBACK. This ROLLBACK only cleans up test data.

-- Finish the tests and clean up.
SELECT * FROM finish();
ROLLBACK;
```

**Test isolation caveats:**
1. **Sequences**: Always use explicit `id` values in `INSERT` fixtures. PostgreSQL sequences advance even in rolled-back transactions; relying on `DEFAULT` causes non-deterministic ID values across test runs.
2. **`pg_notify()` / AMQP triggers**: If the migration adds a trigger that calls `pg_notify()` or `amqp_notify_id()`, those notifications fire even inside rolled-back transactions. To suppress them during tests that are not testing the AMQP path:
   ```sql
   ALTER TABLE geokrety.gk_<table> DISABLE TRIGGER after_99_notify_amqp_<name>;
   -- ... your tests ...
   ALTER TABLE geokrety.gk_<table> ENABLE TRIGGER after_99_notify_amqp_<name>;
   ```

**Coverage targets (aim to cover all that apply):**

| Category | What to test |
|----------|-------------|
| Schema shape | `has_table`, `has_column`, `col_not_null`, `col_default_is`, `has_index`, `has_fk`, `has_trigger`, `has_function` |
| Insert happy path | `lives_ok` for every valid combination |
| Insert violations | `throws_ok` for NOT NULL (`23502`), UNIQUE (`23505`), FK (`23503`), CHECK (`23514`), value too long (`22001`) |
| Trigger — INSERT branch | Side-effect verified (e.g. counter incremented) |
| Trigger — UPDATE branch (watched col changes) | Side-effect verified |
| Trigger — UPDATE branch (unwatched col changes) | No side-effect (no-op branch) |
| Trigger — UPDATE branch (NULL ↔ value) | NULL-sensitive logic tested both ways |
| Trigger — DELETE branch | Counter decremented; `GREATEST(0,…)` floor confirmed |
| CASCADE on delete | Delete parent → child gone (`0` rows) |
| Default values | Column defaults applied when omitted from INSERT |
| Edge cases | Empty strings, max-length, NULL in nullable columns, boundary timestamps |
| ⚠️ `down()` inverse | Verified by Step 7b rollback cycle, NOT this transaction |

### Step 6 — Expert Review Loop

Use critical-thinking loop skill `.github/skills/critical-loop/SKILL.md` to run the migration and test files through multiple expert reviews (DBA, critical-thinking, quality-engineer) before applying. This iterative review process catches issues early and ensures the migration is robust, performant, and well-tested before it touches production data.

#### Old process (for reference):

Invoke the following agent loop **before** running the migration. **Maximum 5 full rounds**; if consensus is not reached by then, record remaining concerns in `99-OPEN-QUESTIONS.md` and proceed. The **human user has final authority** on any unresolved concern. At the end of each round, check the updated files again as the user may have added `TODO` comments to guide the next round of review, you must read those comments and verify they are addressed in the next round. Also re-reader this `SKILL.md` documentation after each round to ensure all best practices are followed as there may have been live updates to the documentation.

##### Round 1 — DBA Review

Ask the `dba` agent to review:
1. The migration file for correctness, safety, performance impact, and reversibility
2. The test file for coverage completeness
3. Missing indexes; locking implications (consider `CREATE INDEX CONCURRENTLY` for large tables)

##### Round 2 — Critical Thinking Review

Ask the `critical-thinking` agent to challenge:
1. Every assumption in the migration design
2. Whether `down()` is a true inverse (including data loss scenarios)
3. Whether all trigger branches and constraint types are covered in tests
4. Naming, conventions, security (`$this->execute()` SQL injection surface)

##### Round 3 — Quality Engineer Review

Ask the `quality-engineer` agent to assess:
1. Test completeness — all trigger branches, all relevant error codes
2. Missing edge cases (NULL handling, empty sets, boundary values, concurrent writes)
3. Correct `SELECT plan(N)` count
4. Test isolation (sequences, `pg_notify()` side-effects on external systems)

##### Round 4 — Check for User Comments

1. Check if the user has added any comments or `TODO` notes in the migration or test files after the previous rounds of review.
2. If yes, verify that the concerns raised in those comments are addressed in the next round of review.

##### Convergence

Pass unresolved concerns back through agents up to the 5-round cap. If any concern remains, file it in `99-OPEN-QUESTIONS.md`, add a cross-reference comment in the relevant file, and proceed — open questions do **not** block migration apply but must be reviewed before the next production deployment.

### Step 7 — Apply Migration

```bash
.github/skills/phinx/scripts/phinx.sh migrate --count=1
```

Verify the output confirms the migration ran successfully. If it fails, diagnose the root cause — do **not** retry without a fix. Check `phinx.sh status` to confirm the migration is still pending before re-running.

### Step 7b — Verify `down()` Reversibility

```bash
# Roll back the migration
.github/skills/phinx/scripts/phinx.sh rollback

# Verify the schema is clean — no orphaned objects remain
# (inspect with: psql -c "\dt geokrety.*" and "\df geokrety.*")

# Re-apply the migration
.github/skills/phinx/scripts/phinx.sh migrate --count=1
```

If rollback fails, the `down()` method has a bug — fix it before proceeding. Common causes: wrong drop order, missing `IF EXISTS`, dropping a column that was never added.

### Step 8 — Run Tests

Before running the tests, ensure the test database schema is in sync with the latest migration. If your migration added new tables/columns, the test schema must be updated to avoid false failures due to missing objects. Only copy the schema after the migration has been re-applied in Step 7b to ensure the test schema matches the migrated schema exactly.
```bash
/home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh
```

Then run the pgTAP tests:
```bash
.github/skills/pgtap/scripts/pgtap.sh
```

**If tests fail:**

1. Diagnose: schema mismatch? wrong plan count? trigger logic bug? assertion typo?
2. Classify the failure surface before editing: rerun the smallest focused subset that covers the migration first, then compare against the full-suite failures. If only unrelated dirty-worktree files fail, record that explicitly and do not fold those fixes into the current migration task.
3. If the migration needs changes:
   ```bash
   .github/skills/phinx/scripts/phinx.sh rollback
   # edit migration file; re-run php -l check
   .github/skills/phinx/scripts/phinx.sh migrate --count=1
   .github/skills/phinx/scripts/phinx.sh rollback   # re-verify down()
   .github/skills/phinx/scripts/phinx.sh migrate --count=1
   /home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh
   ```
4. If only the tests are wrong (migration is correct), edit the test file and re-run.
5. **Plan count mismatch**: recount with the grep command from Step 5 and update `plan(N)`.
6. Iterate until all owned tests pass with no plan mismatch, then run the full suite and report any unrelated failures separately.

### Step 9 — Update Schema Test

For every new object created in the migration, update the relevant schema test file:

| Object created | Action required |
|----------------|-----------------|
| New table | Add `SELECT has_table('gk_<name>');` to `test-10-schema.sql`; increment its `plan(N)` |
| New schema | Add `SELECT has_schema('<name>');` to `test-10-schema.sql` |
| New trigger | Assert in your dedicated test file with `has_trigger(schema, table, name, msg)` |
| New function | Assert in your dedicated test file with `has_function(schema, name, msg)` |
| New materialized view | Assert in your dedicated test file with `has_materialized_view(schema, name, msg)` |

### Step 10 — Suggest improvement to the process or documentation if you encountered any friction or ambiguity during this workflow.

Once all operation is complete, reflect on the process and suggest any improvements to this documentation or the workflow itself. Consider:
- Were there any points of confusion or ambiguity in the documentation?
- Did you encounter any friction or blockers during the workflow?
- Is there any step that could be streamlined or clarified for future users?
- Are there any additional best practices or tips that could be added to the documentation based on your experience?

Update this `SKILL.md` file with your suggestions to help future users have a smoother experience. Your insights are valuable for continuous improvement of our processes!

Process improvement added 2026-03-16:
- When the repository is already dirty, Step 8 should explicitly distinguish between `owned` validation failures and unrelated full-suite regressions. Running the focused pgTAP subset before the full suite makes that boundary obvious and avoids accidental scope creep during migration work.

---

## PHP Migration Patterns Reference

### Add Column to Existing Table

```php
// up():
$this->table('geokrety.gk_geokrety')
    ->addColumn('new_column', 'boolean', [
        'default' => false,
        'null'    => false,
        // NOTE: 'after' is a MySQL-specific Phinx option.
        // PostgreSQL ignores it silently — do NOT use 'after' here.
        // Columns are always appended in PostgreSQL.
    ])
    ->update();

// down():
$this->table('geokrety.gk_geokrety')->removeColumn('new_column')->update();
```

### Create Table with FK, Index, and Counter Trigger

```php
// up():
$table = $this->table('geokrety.gk_loves', ['id' => true, 'primary_key' => ['id']]);
$table->addColumn('user', 'integer', ['null' => false])
    ->addColumn('geokret', 'integer', ['null' => false])
    ->addColumn('created_on_datetime', 'datetime', [
        'null' => false, 'default' => 'CURRENT_TIMESTAMP', 'timezone' => true,
    ])
    ->addIndex(['user', 'geokret'], ['unique' => true, 'name' => 'idx_gk_loves_user_geokret'])
    ->addIndex(['geokret'], ['name' => 'idx_gk_loves_geokret'])
    ->addForeignKey('user', 'geokrety.gk_users', 'id', ['delete' => 'CASCADE', 'update' => 'NO_ACTION'])
    ->addForeignKey('geokret', 'geokrety.gk_geokrety', 'id', ['delete' => 'CASCADE', 'update' => 'NO_ACTION'])
    ->create();

// Trigger function — note: single-quoted heredoc is mandatory
$this->execute(<<<'EOL'
CREATE OR REPLACE FUNCTION geokrety.gk_loves_update_count()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    VOLATILE
    COST 100
AS $BODY$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE geokrety.gk_geokrety
        SET loves_count = loves_count + 1
        WHERE id = NEW.geokret;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE geokrety.gk_geokrety
        SET loves_count = GREATEST(0, loves_count - 1)
        WHERE id = OLD.geokret;
    END IF;
    RETURN NULL;
END;
$BODY$;
EOL);

// Trigger — naming convention: after_<table_suffix>_<description> for internal logic triggers
$this->execute(<<<'EOL'
CREATE TRIGGER after_gk_loves_update_count
    AFTER INSERT OR DELETE
    ON geokrety.gk_loves
    FOR EACH ROW
    EXECUTE FUNCTION geokrety.gk_loves_update_count();
EOL);

// down():
// Drop order: triggers → functions → columns → tables
$this->execute('DROP TRIGGER IF EXISTS after_gk_loves_update_count ON geokrety.gk_loves');
$this->execute('DROP FUNCTION IF EXISTS geokrety.gk_loves_update_count()');
$this->table('geokrety.gk_geokrety')->removeColumn('loves_count')->update();
$this->table('geokrety.gk_loves')->drop()->save();
```

### Execute Raw SQL with UPDATE branch trigger (single-quoted heredoc)

```php
// Use <<<'EOL' (single-quoted) to prevent PHP variable interpolation.
$this->execute(<<<'EOL'
CREATE OR REPLACE FUNCTION geokrety.my_trigger_function()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    VOLATILE AS
$BODY$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- handle insert
    ELSIF TG_OP = 'UPDATE' AND NEW.watched_col IS DISTINCT FROM OLD.watched_col THEN
        -- Only fire when the watched column actually changes
        -- IS DISTINCT FROM handles NULLs correctly unlike <>
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE geokrety.gk_some_table
        SET counter = GREATEST(0, counter - 1)
        WHERE id = OLD.related_id;
    END IF;
    RETURN NULL;
END;
$BODY$;
EOL);
```

### Insert Reference Data (single-quoted heredoc mandatory)

```php
// Use <<<'EOL' — the INSERT contains literal SQL strings that could
// accidentally shadow PHP variables in a double-quoted string.
$this->execute(<<<'EOL'
INSERT INTO geokrety.gk_users_settings_parameters
    (name, type, "default", description, created_on_datetime, updated_on_datetime)
VALUES
    ('SETTING_NAME', 'bool', 'true', 'Description here', NOW(), NOW());
EOL);
```

### Materialized View

```php
// up():
$this->execute(<<<'EOL'
CREATE MATERIALIZED VIEW geokrety.gk_my_stats AS
SELECT ...;
EOL);
$this->execute(<<<'EOL'
CREATE UNIQUE INDEX idx_gk_my_stats_key ON geokrety.gk_my_stats (key_column);
EOL);
$this->execute("COMMENT ON MATERIALIZED VIEW geokrety.gk_my_stats IS 'Description here.';");
$this->execute('REFRESH MATERIALIZED VIEW geokrety.gk_my_stats;');

// down():
// The index is dropped automatically with the materialized view
$this->execute('DROP MATERIALIZED VIEW IF EXISTS geokrety.gk_my_stats;');
```

---

## pgTAP Assertions Quick Reference

| Assertion | Signature | Use case |
|-----------|-----------|----------|
| `lives_ok` | `lives_ok($$SQL$$, 'msg')` | SQL must succeed |
| `throws_ok` | `throws_ok($$SQL$$, errcode, errmsg, 'msg')` | SQL must throw specific error |
| `is` | `is(actual, expected, 'msg')` | Equality check |
| `isnt` | `isnt(actual, expected, 'msg')` | Not-equal check |
| `ok` | `ok(boolean, 'msg')` | Generic truth check |
| `like` | `like(value, pattern, 'msg')` | Pattern match (e.g. EXPLAIN plans) |
| `has_table` | `has_table(schema, table, 'msg')` | Table exists |
| `has_column` | `has_column(schema, table, col, 'msg')` | Column exists |
| `col_not_null` | `col_not_null(schema, table, col, 'msg')` | Column is NOT NULL |
| `col_default_is` | `col_default_is(schema, table, col, default, 'msg')` | Column default |
| `col_type_is` | `col_type_is(schema, table, col, type, 'msg')` | Column type |
| `has_index` | `has_index(schema, table::name, idx, ARRAY[cols])` | Index exists |
| `has_fk` | `has_fk(schema, table, fk_name, 'msg')` | Foreign key exists |
| `has_trigger` | `has_trigger(schema, table, trigger_name, 'msg')` | Trigger exists |
| `has_function` | `has_function(schema, func_name, 'msg')` | Function exists |
| `has_view` | `has_view(schema, view_name, 'msg')` | Regular view exists (**NOT** materialized views) |
| `has_materialized_view` | `has_materialized_view(schema, matview_name, 'msg')` | Materialized view exists |
| `has_sequence` | `has_sequence(schema, seq_name, 'msg')` | Sequence exists |
| `has_schema` | `has_schema(schema_name, 'msg')` | Schema exists |

## PostgreSQL Error Code Reference

| Code | Condition | When to test |
|------|-----------|--------------|
| `'23502'` | NOT NULL violation | Any NOT NULL column |
| `'23503'` | FK violation | Any FK column — referencing nonexistent parent |
| `'23505'` | UNIQUE violation | Any UNIQUE constraint or unique index |
| `'23514'` | CHECK constraint violation | Any CHECK constraint |
| `'23P01'` | Exclusion constraint violation | Any EXCLUDE constraint |
| `'22001'` | Value too long | Any `varchar(N)` column |
| `'22P02'` | Invalid input syntax | Enum or domain violations |

---

## NNN Block Map for Test Files

| NNN range | Domain |
|-----------|--------|
| `00` | Init / pgTAP sanity |
| `10–29` | Schema shape, functions, sessions |
| `30–49` | Moves, move distance, move archive |
| `50–69` | GeoKrety, pictures, collectibles |
| `70–89` | Users, user accounts, email |
| `90–109` | GeoKret counters, last log/position |
| `110–199` | Account/email activation, token flows |
| `200–299` | Audit, sharded counters, password flows |
| `300+` | Awards, misc |

**Conflict check** — before finalising the test filename:

```bash
ls /home/kumy/GIT/geokrety-website/website/db/tests/test-NNN-*.sql
```

If the NNN slot is taken, append a hyphen qualifier (e.g. `test-52-geokret-loves.sql` alongside `test-52-geokret-count-distance.sql`) — never reuse the exact same filename.

---

## Troubleshooting

| Problem | Solution |
|---------|---------|
| PHP syntax error | Fix heredoc, missing semicolon, or class name mismatch; re-run `php -l` |
| Class name mismatch | Class must be CamelCase of filename suffix only (strip timestamp prefix) |
| Plan count mismatch | Use the `grep -cP` command from Step 5 to recount; update `plan(N)` |
| `down()` fails | Check drop order: RLS policies → mat. views → views → triggers → functions → cols → tables → sequences |
| FK violation on rollback | `down()` must drop dependent objects before parent tables |
| Counter goes negative | Use `GREATEST(0, count - 1)` in trigger DELETE branch |
| Sequence IDs non-deterministic | Use explicit `id` values in all test `INSERT` fixtures |
| Spurious AMQP notifications | Disable the AMQP trigger for the test block (see Step 5 isolation caveats) |
| `after` column option has no effect | `'after'` is MySQL-only; PostgreSQL always appends columns — remove it |
| Partial `up()` failure | PG transactional DDL auto-rolls back; check `phinx.sh status` to confirm, fix, retry |
| Class name collision | Append a version suffix (e.g. `V2`) and note why; verify with `grep -rh "^final class "` |
| Function signature changed but old callers still hit the legacy overload | Explicitly `DROP FUNCTION IF EXISTS schema.fn_name(old_arg_types...)` before creating the new signature; PostgreSQL keeps overloads side-by-side |

---

## Open Questions File

When concerns cannot be resolved within the review loop (max 7 rounds), append to `99-OPEN-QUESTIONS.md` **located alongside the migration file** at:

```
/home/kumy/GIT/geokrety-stats/docs/database-refactor/99-OPEN-QUESTIONS.md
```

```markdown
## OPEN QUESTION — <YYYY-MM-DD> — <short title>

**Migration:** `<filename>.php`
**Raised by:** <agent name>, review round <N>
**Question:** <full description>
**Impact:** <what could go wrong if left open>
**Proposed approaches:** <options if any>
**Status:** OPEN — does not block current migration; must be reviewed before next production deployment
```

Also add a cross-reference in the migration or test file:
```sql
-- OPEN QUESTION: see 99-OPEN-QUESTIONS.md — <short title>
```

---

## Integration with Sibling Skills

This skill **calls** and **depends on**:

- **`phinx` skill** (`.github/skills/phinx/SKILL.md`): used for `migrate --count=1`, `rollback`, and `status` in Steps 7, 7b, and 8.
- **`pgtap` skill** (`.github/skills/pgtap/SKILL.md`): used for running the test suite in Step 8.

Always obey the migration policy defined in the `phinx` skill: apply one migration at a time (`--count=1`) unless the user explicitly requests otherwise.

---

## Additional automation requirements

- **Maintain `99-IMPLEMENTATION.md`:** Every migration task executed via this skill MUST update `docs/database-refactor/99-IMPLEMENTATION.md` with an advancement summary following the existing format. Mark checklist items completed as you progress.
- **Update test schema copy script when needed:** If the migration requires new reference data or schema objects to be present in the `test` DB, update `/home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh` so the test database receives any necessary data copies before running pgTAP.
    - **Prefer idempotent helpers over raw dumps:** When available, use the database's initialization/snapshot helpers instead of copying raw table dumps into `tests` (example helpers: `stats.fn_snapshot_entity_counters()`, `stats.fn_seed_daily_activity()`). This produces the canonical, "zeroed" reference state and avoids stale or environment-specific data.
    - **Execution order:** Run `stats.fn_snapshot_entity_counters()` before `stats.fn_seed_daily_activity()` so shard totals exist before daily aggregates are computed.
    - **Trigger side-effects:** If seeding causes `pg_notify()` or AMQP notifications, disable the relevant AMQP triggers around the seed call in the tests script and re-enable them afterwards.
- **Execution commands:** Use the repository wrappers for all runtime steps to ensure environment parity:
    - Apply / rollback migrations: `/home/kumy/GIT/geokrety-stats/.github/skills/phinx/scripts/phinx.sh migrate --count=1` and `/home/kumy/GIT/geokrety-stats/.github/skills/phinx/scripts/phinx.sh rollback`
    - Copy schema to tests: `/home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh`
    - Run pgTAP: `/home/kumy/GIT/geokrety-stats/.github/skills/pgtap/scripts/pgtap.sh`
- **Iterate until green:** Run `phinx migrate`, update `99-IMPLEMENTATION.md`, copy schema to tests, run pgTAP, fix migration or tests as required, repeat the apply/rollback verification cycle until tests pass.
- **Continue work automatically when trivial:** If a task completes quickly (migration + tests pass without open questions), proceed to the next unimplemented task listed under `docs/database-refactor/sprint-2/` and repeat the same workflow, unless the user says to stop.
