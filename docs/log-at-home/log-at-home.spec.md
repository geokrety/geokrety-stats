---
title: "Log At Home: gk_moves author-home flag"
version: 1.0
date_created: 2026-03-19
last_updated: 2026-03-19
owner: "GeoKrety Community"
---

# Log At Home Specification

## Purpose & Scope

Add a derived boolean flag to `geokrety.gk_moves` that marks whether a move was logged within 50 meters of the move author's configured home location, using geography-based proximity matching for performance and accuracy.

For this specification, the canonical source objects in the current database are:

- movement table: `geokrety.gk_moves`
- author table: `geokrety.gk_users`
- home geography column: `geokrety.gk_users.home_position` (canonical for matching)
- home coordinate columns: `geokrety.gk_users.home_latitude`, `geokrety.gk_users.home_longitude` (used only for normalization)

This specification covers only the database contract:

- one new boolean column on `geokrety.gk_moves`
- one `BEFORE INSERT OR UPDATE` trigger on `geokrety.gk_moves`
- one manual backfill function for historical rows
- one Phinx migration
- one pgTAP test file
- migration apply and rollback verification

This specification does not implement the migration yet.

## Naming Decision

### Recommended column name

`logged_at_author_home`

Reasoning:

- `spotted_at_home` is ambiguous in GeoKrety context because it could mean the GeoKret is at its owner's home, the holder's home, or simply indoors.
- `logged_at_author_home` is explicit that the flag is about the move author and the author's configured home coordinates.
- the name reads naturally as a derived boolean fact on a move row.

### Companion object names

- trigger function: `geokrety.fn_gk_moves_set_logged_at_author_home()`
- trigger: `tr_gk_moves_before_logged_at_author_home`
- backfill function: `stats.fn_backfill_gk_moves_logged_at_author_home(p_period tstzrange DEFAULT NULL, p_batch_size INT DEFAULT 50000)`

This specification makes `logged_at_author_home` the canonical column name for implementation.

## Functional Contract

### Column contract

Add this column to `geokrety.gk_moves`:

- name: `logged_at_author_home`
- type: `boolean`
- nullability: `NOT NULL`
- default: `false`

### Error Handling Contract

This section defines how the trigger and backfill function handle exceptional conditions:

**Trigger Execution Errors:**

- **If `ST_DWithin()` raises an exception** (invalid geography, corrupted coordinates): The trigger propagates the exception and the write is rejected (FAIL-CLOSE). Data integrity is preserved at the cost of availability. The client receives the exception message.

- **If author row lookup fails or returns no home_position**: The flag is set to `false` and the write proceeds (FAIL-SAFE). This represents the conservative case: "no matching author home, treat as not at home."

- **If NULL checks fail unexpectedly**: All NULL conditions explicitly set the flag to `false`. No exception is raised; the trigger is defensive.

**Backfill Function Errors:**

- **If batch update fails mid-transaction**: The entire transaction is rolled back; no rows are modified. The error is returned in the summary text with the SQLSTATE code.

- **If invalid parameters are supplied** (NULL or negative `p_batch_size`): The function raises an exception immediately with a clear error message before any rows are processed:
  ```
  RAISE EXCEPTION 'p_batch_size must be a positive integer (got %)', p_batch_size
    USING HINT = 'Use DEFAULT value for automatic batch sizing or provide p_batch_size > 0';
  ```

### Truth table

`logged_at_author_home = true` only when all of the following are true for the row being written:

- `NEW.author` is not `NULL`
- `NEW.position` is not `NULL`
- the referenced author row exists in `geokrety.gk_users`
- the author has a configured home geography (`home_position IS NOT NULL`)
- `ST_DWithin(NEW.position, gk_users.home_position, 50)` returns `true` (within 50 meters)

In every other case the value must be `false`.

### Comparison rule

The business rule is geography-based proximity matching with a 50-meter tolerance.

- radius search is required using `ST_DWithin`
- tolerance window is fixed at 50 meters
- no country-based fallback
- scalar coordinates are used only for normalization, not for matching

The canonical source of a user's home location is `gk_users.home_position`, and the comparison must use the normalized geography values `NEW.position` and `gk_users.home_position`.

Implementation note:

- the default and recommended implementation path is to use `ST_DWithin(NEW.position, gk_users.home_position, 50)` for all matching
- scalar coordinates (`lat`, `lon`) are only used for normalization of incoming position data, not for matching
- the normalized geography approach is chosen for performance and accuracy over scalar coordinate equality
- migration comments and pgTAP test intent must document that the internal predicate uses geography proximity (≤50m) for matching

### Definition of configured home location

In this specification, `configured home location` means the author's `home_position` is not `NULL` after the existing `gk_users` home-normalization trigger logic has run.

The `home_position` geography is the system-normalized representation of the author's home coordinates. Scalar coordinates (`home_latitude`, `home_longitude`) are automatically synchronized with `home_position` by the existing schema.

If that normalization clears the home location, `logged_at_author_home` must evaluate to `false`.

### Position Nullability Clarification

**Confirmed:** The column `geokrety.gk_moves.position` is declared `NOT NULL` in the current schema. It cannot store NULL values due to the schema constraint.

**Why the truth table checks for NULL position:** The condition `NEW.position IS NULL → false` in the truth table is included as a defensive programming practice and forward-compatibility guard. Even though the schema constraint prevents NULL, the trigger function must include this check explicitly to:

1. Fail safely if the schema constraint is ever removed in the future
2. Document intent clearly for maintainers
3. Prevent silent misbehavior if position normalization somehow produces unexpected values

**In practice:** The NULL check will never succeed during normal operation because the `before_20_gis_updates` trigger normalizes position before this trigger executes. However, the defensive check remains in the specification for robustness.

### Position-only writes

`geokrety.gk_moves` already supports synchronization between `lat` / `lon` and `position`. **The new migration must run after the `before_20_gis_updates` trigger to ensure `position` is fully normalized before this feature computes `logged_at_author_home`.**

This feature must therefore be correct for all of these write paths:

- insert or update with `lat` and `lon`
- insert or update with `position` only
- insert or update with both `position` and scalar coordinates

When `position` and scalar coordinates disagree in the incoming statement, `logged_at_author_home` must be computed from the final normalized `NEW.position` value after the existing GIS synchronization logic has run.

The new trigger must execute after the existing `before_20_gis_updates` trigger on `geokrety.gk_moves` so that the fully normalized `NEW.position` value is available before `logged_at_author_home` is derived. The chosen trigger name must preserve that ordering.

## Trigger Contract

### Trigger shape

Create a PostgreSQL trigger on `geokrety.gk_moves` with this contract:

- timing: `BEFORE`
- events: `INSERT OR UPDATE`
- granularity: `FOR EACH ROW`
- trigger name: `tr_gk_moves_before_logged_at_author_home`
- function name: `geokrety.fn_gk_moves_set_logged_at_author_home()`

### Trigger Execution Order Requirement

**CRITICAL:** This trigger must execute **AFTER** the existing `before_20_gis_updates` trigger on `geokrety.gk_moves`.

**Why this matters:** The `before_20_gis_updates` trigger (line 5911 of geokrety-schema.sql) synchronizes `position` geography from scalar `lat`/`lon` values. This feature **must** use the normalized `position` geography for matching. If this trigger executes before normalization completes, it will compute the derived boolean from non-normalized (or partially normalized) data.

**How execution order is guaranteed:** PostgreSQL executes BEFORE triggers in alphabetical order by trigger name. The trigger name `tr_gk_moves_before_logged_at_author_home` sorts alphabetically **after** all `before_*` triggers (since 't' > 'b'), ensuring this trigger runs last among BEFORE triggers.

**Verification:** To confirm ordering after migration apply, run:
```sql
SELECT tgname FROM pg_trigger
WHERE tgrelid = 'geokrety.gk_moves'::regclass AND NOT tgisinternal
ORDER BY tgforenabled DESC, tgname ASC;
```

`tr_gk_moves_before_logged_at_author_home` must appear after `before_20_gis_updates` in the output.

### Trigger behavior

The trigger function must set `NEW.logged_at_author_home` on every qualifying write before the row is stored.

The value is fully derived. Any caller-supplied value for `logged_at_author_home` must be ignored and overwritten by the trigger result.

Minimum recomputation inputs:

- `NEW.author`
- `NEW.position`

On `UPDATE`, the function may skip recomputation only when `NEW.author` and `NEW.position` are unchanged and `NEW.logged_at_author_home = OLD.logged_at_author_home`; otherwise it must recompute and overwrite the stored value.

### Trigger ownership boundaries

This feature must be maintained automatically only from `gk_moves` writes.

Out of scope for live trigger maintenance:

- fan-out updates from `gk_users` when a user changes their home location
- retroactive recomputation of historical moves during user-profile edits

If an author changes their home location later, historical rows are corrected by the dedicated backfill function, not by a trigger on `gk_users`.

## Backfill Function Contract

### Purpose

Historical `gk_moves` rows need a repair path so existing data can be populated after the new column and trigger are deployed.

### Canonical function

Create a manual backfill function:

`stats.fn_backfill_gk_moves_logged_at_author_home(p_period tstzrange DEFAULT NULL, p_batch_size INT DEFAULT 50000)`

The function returns `text`.

**Return Value Format:**

The function must return a single-line text summary with the following format:
```
Processed <total_processed> rows; <rows_changed> rows updated;
<batch_count> batches completed; <scope_description>.
```

**Example return values:**
```
'Processed 1000000 rows; 15234 rows updated; 20 batches completed; full-history scope.'
'Processed 50000 rows; 8765 rows updated; 1 batch completed; period-scoped from 2025-01-01 to 2025-03-19.'
'Processed 0 rows; 0 rows updated; 0 batches completed; empty period scope (no rows in range).'
```

### Parameter semantics

- `p_period` scopes candidate rows by `gk_moves.moved_on_datetime`
- `NULL` means full-history scope
- non-`NULL` ranges use standard PostgreSQL `tstzrange` semantics
- unbounded lower or upper bounds are allowed
- the implementation must treat the range exactly as passed and not silently widen it
- `p_batch_size` must be a positive integer
- `NULL`, zero, or negative `p_batch_size` values must raise an error:
  ```
  RAISE EXCEPTION 'p_batch_size must be a positive integer (got %)', p_batch_size
    USING HINT = 'Use DEFAULT value for automatic batch sizing or provide p_batch_size > 0';
  ```

### Required behavior

- operates on existing rows in `geokrety.gk_moves`
- recomputes `logged_at_author_home` from current `author`, `position`, and `home_position` using `ST_DWithin(position, home_position, 50)`
- accepts optional `p_period` scoped on `gk_moves.moved_on_datetime`
- accepts `p_batch_size` for batched execution on large datasets
- is idempotent: repeated calls produce the same final values
- processes rows in deterministic ascending `gk_moves.id` order inside each scoped run
- returns a deterministic summary text including, at minimum, scoped row count, changed row count, batch count, and whether the run was full-history or period-scoped
- updates only rows whose derived value differs from the stored value

### Operational constraints

The implementation must avoid a backfill strategy that needlessly amplifies unrelated `gk_moves` trigger work across millions of rows.

The implementation should therefore:

- prefer set-based batched updates over row-by-row loops
- update only rows whose derived boolean changes
- prefer a dedicated backfill-safe trigger guard or equivalent scoped mechanism if unrelated `gk_moves` maintenance triggers would otherwise turn the repair into a full analytics replay
- not rely on globally disabling triggers as the primary operational plan

### Integration requirement

The backfill method MUST be integrated into script `/home/kumy/GIT/geokrety-stats/docs/database-refactor/run_snapshot_backfill.py` with a new command option `--backfill-logged-at-author-home` that calls this function with appropriate parameters for a full-history run.

- The script must be updated to allow users to run the backfill function with a single command, without needing to call the function manually from psql.
- The script must handle any necessary setup or teardown for the backfill function, such as logging or progress reporting, using existing infrastructure where possible.
- The script must be tested to ensure that the backfill function is called correctly and that the expected summary output is produced.
- The script must maintain backward compatibility with existing commands and not require any changes to run other backfill operations.
- Documentation for the new command option must be added to the script's help output.
- The integration must be implemented in a way that does not introduce new dependencies or significantly increase the complexity of the script.

## Requirements

| ID | Requirement |
| --- | --- |
| REQ-LAH-001 | Add `geokrety.gk_moves.logged_at_author_home BOOLEAN NOT NULL DEFAULT false`. |
| REQ-LAH-002 | Populate the column automatically with a PostgreSQL `BEFORE INSERT OR UPDATE` trigger on `geokrety.gk_moves`. |
| REQ-LAH-003 | Use the move author's home geography from `geokrety.gk_users.home_position`. |
| REQ-LAH-004 | Set the column to `true` only when move position is within 50 meters of the author's home geography (`ST_DWithin(NEW.position, gk_users.home_position, 50) = true`). |
| REQ-LAH-005 | Set the column to `false` when author, move position, or author home position is missing. |
| REQ-LAH-006 | Use geography proximity matching (≤50m) as the canonical matching rule; do not use country-based fallback. |
| REQ-LAH-006A | The derived value must be correct for writes that enter through the existing `position` normalization path, not only direct `lat` / `lon` writes. |
| REQ-LAH-006B | When `position` and scalar coordinates disagree on input, the derived value must use the post-normalization `NEW.position` value produced by `before_20_gis_updates`. |
| REQ-LAH-006C | The implementation must use geography proximity (`ST_DWithin`) for all matching; migration comments and pgTAP test intent must document that the internal predicate uses geography proximity (≤50m). |
| REQ-LAH-007 | Provide a dedicated manual backfill function following the existing `stats.fn_backfill_*` naming family. |
| REQ-LAH-008 | The backfill function must support optional period-scoped execution and configurable batch size. |
| REQ-LAH-009 | The backfill function must be idempotent. |
| REQ-LAH-009A | The backfill function must reject `NULL`, zero, or negative batch sizes. |
| REQ-LAH-010 | Deliver the change as a new Phinx migration in `/home/kumy/GIT/geokrety-website/website/db/migrations/`. |
| REQ-LAH-011 | Deliver exhaustive pgTAP tests in `/home/kumy/GIT/geokrety-website/website/db/tests/`. |
| REQ-LAH-012 | Validate migration apply, rollback, and pgTAP coverage before considering the change complete. |
| REQ-LAH-013 | Live maintenance is intentionally limited to `gk_moves` writes; user home-profile edits are reconciled by backfill, not fan-out triggers. |
| REQ-LAH-014 | Any user-supplied `logged_at_author_home` value on INSERT or UPDATE must be overwritten by the trigger-computed result. |
| REQ-LAH-015 | The migration must create the backfill function but must not execute the historical backfill automatically during `phinx migrate`. |

## Acceptance Criteria

### AC-LAH-001: Matching insert

Given a user with `home_position` set at coordinates (latitude = 48.8, longitude = 2.3)
When a new `gk_moves` row is inserted for that author with `position` within 50 meters of the home geography
Then `logged_at_author_home = true`

### AC-LAH-002: Non-matching insert

Given a user with configured home geography
When a new `gk_moves` row is inserted for that author with `position` more than 50 meters away from the home geography
Then `logged_at_author_home = false`

### AC-LAH-003: Missing author

Given a move with `author IS NULL`
When the row is inserted or updated
Then `logged_at_author_home = false`

### AC-LAH-004: Missing user home

Given an author whose `home_position` is `NULL`
When a move is inserted or updated
Then `logged_at_author_home = false`

### AC-LAH-005: Matching update

Given an existing move with `logged_at_author_home = false`
When `author`, `lat`, or `lon` is updated so the normalized `position` is now within 50 meters of the author's home geography
Then `logged_at_author_home` is recomputed to `true`

### AC-LAH-006: Non-matching update

Given an existing move with `logged_at_author_home = true`
When `author`, `lat`, or `lon` is updated so the normalized `position` is now more than 50 meters from the author's home geography
Then `logged_at_author_home` is recomputed to `false`

### AC-LAH-006A: Position-only insert

Given a user with configured home geography
When a new `gk_moves` row is inserted using `position` only and the normalized position is within 50 meters of the author's home geography
Then `logged_at_author_home = true`

### AC-LAH-006B: Position-only update

Given an existing move with `logged_at_author_home = false`
When the move `position` is updated and the normalized position is now within 50 meters of the author's home geography
Then `logged_at_author_home` is recomputed to `true`

### AC-LAH-006C: Caller override is ignored

Given an insert or update that explicitly supplies an incorrect `logged_at_author_home` value
When the row is written
Then the trigger overwrites that supplied value with the correct derived result

### AC-LAH-006D: Mixed position and scalar input uses normalized result

Given an insert or update that supplies both `position` and scalar coordinates with conflicting values
When the existing GIS synchronization path normalizes the row
Then `logged_at_author_home` is computed from the final normalized `NEW.position` value using `ST_DWithin`, not from the raw conflicting input

### AC-LAH-007: Historical repair

Given historical `gk_moves` rows created before the migration
When `stats.fn_backfill_gk_moves_logged_at_author_home(...)` is run
Then all in-scope rows are updated to the same values that the live trigger would produce from current source data

### AC-LAH-008: Idempotent repair

Given a dataset already fully reconciled
When the backfill function is run again with the same scope
Then the final stored values remain unchanged and the function completes without error

### AC-LAH-009: User home edits do not fan out automatically

Given existing `gk_moves` rows for a user
When that user's home coordinates are changed in `geokrety.gk_users`
Then existing move rows do not change automatically
And a later backfill run reconciles them to the new derived truth table

### AC-LAH-010: Migration does not auto-run historical repair

Given the migration is applied on a database that already contains historical `gk_moves` rows
When `phinx migrate` completes
Then the schema objects exist
And historical rows remain untouched until the manual backfill function is executed explicitly

## Migration Deliverables

### Migration file

Create one new Phinx migration in:

`/home/kumy/GIT/geokrety-website/website/db/migrations/`

Recommended filename suffix:

- `<timestamp>_add_logged_at_author_home_to_gk_moves.php`

The exact timestamp is assigned during implementation.

### Objects created by the migration

- column `geokrety.gk_moves.logged_at_author_home`
- function `geokrety.fn_gk_moves_set_logged_at_author_home()`
- trigger `tr_gk_moves_before_logged_at_author_home`
- function `stats.fn_backfill_gk_moves_logged_at_author_home(tstzrange, int)`

### Rollback contract

`down()` must remove all objects added by `up()` in safe dependency order:

- drop trigger
- drop trigger function
- drop backfill function
- drop column

### Migration Dependencies

This migration depends on the following pre-existing database objects:

**Required Triggers:**
- `geokrety.before_20_gis_updates` trigger on `geokrety.gk_moves` must exist and be functional
  - This trigger (defined in geokrety-schema.sql, line 5911) synchronizes `position` geography from `lat`/`lon` scalar coordinates
  - Failure to execute this trigger before our trigger will result in incorrect matching logic

**Required Versions:**
- PostgreSQL ≥ 10 (supports BEFORE INSERT OR UPDATE trigger syntax)
- PostGIS ≥ 2.4 (supports `ST_DWithin` on geography type with meter-based distance)

**Validation Check (must be included in migration `up()`):**

```sql
-- Validate required trigger exists
DO $$
BEGIN
  IF NOT EXISTS(
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'before_20_gis_updates'
    AND tgrelid = 'geokrety.gk_moves'::regclass
  ) THEN
    RAISE EXCEPTION 'Required trigger before_20_gis_updates not found on geokrety.gk_moves. Check that geokrety-schema migrations have been applied.';
  END IF;
END $$;
```

**Rollback Safety:**
- If other migrations depend on `logged_at_author_home` column existing, rollback will fail
- Future migrations should NOT depend on this column; instead, they should check for its existence before using it
- The backfill function created by this migration should be dropped explicitly in `down()` to avoid orphaned function definitions

## pgTAP Test Plan

Create one new pgTAP test file in:

`/home/kumy/GIT/geokrety-website/website/db/tests/`

Recommended filename suffix:

- `test-XXX-gk-moves-logged-at-author-home.sql`

The exact numeric block should be chosen during implementation to match the existing `gk_moves` test range.

### Required schema assertions

| Test ID | Assertion | Expected |
| --- | --- | --- |
| T-LAH-001 | `has_column('geokrety', 'gk_moves', 'logged_at_author_home')` | pass |
| T-LAH-002 | `col_not_null('geokrety', 'gk_moves', 'logged_at_author_home')` | pass |
| T-LAH-003 | `col_default_is('geokrety', 'gk_moves', 'logged_at_author_home', 'false')` | pass |
| T-LAH-004 | `has_function('geokrety', 'fn_gk_moves_set_logged_at_author_home', ARRAY[]::text[])` | pass |
| T-LAH-005 | `function_returns('geokrety', 'fn_gk_moves_set_logged_at_author_home', ARRAY[]::text[], 'trigger')` | pass |
| T-LAH-006 | `has_trigger('geokrety', 'gk_moves', 'tr_gk_moves_before_logged_at_author_home')` | pass |
| T-LAH-007 | `has_function('stats', 'fn_backfill_gk_moves_logged_at_author_home', ARRAY['tstzrange','integer'])` | pass |
| T-LAH-008 | `function_returns('stats', 'fn_backfill_gk_moves_logged_at_author_home', ARRAY['tstzrange','integer'], 'text')` | pass |

### Required behavior assertions

These tests must cover AC-LAH-001 through AC-LAH-009 plus batch-size validation.

| Test ID | Scenario | Expected |
| --- | --- | --- |
| T-LAH-010 | Insert move with position within 50 meters of home | `logged_at_author_home = true` |
| T-LAH-011 | Insert move with position more than 50 meters from home | `logged_at_author_home = false` |
| T-LAH-012 | Insert move for author with `NULL` home position | `logged_at_author_home = false` |
| T-LAH-013 | Insert move with `author IS NULL` | `logged_at_author_home = false` |
| T-LAH-014 | Update position from >50m away to ≤50m away from home | value flips to `true` |
| T-LAH-015 | Update position from ≤50m away to >50m away from home | value flips to `false` |
| T-LAH-016 | Update author from matching user to non-matching user | value recomputed correctly |
| T-LAH-017 | Insert using `position` only and matching normalized coordinates | `logged_at_author_home = true` |
| T-LAH-018 | Update using `position` only from non-match to match | value flips to `true` |
| T-LAH-019 | Caller supplies incorrect boolean on insert or update | stored value is trigger-derived, not caller-derived |
| T-LAH-020 | Updating `gk_users` home coordinates does not auto-update historical moves | move rows remain unchanged before backfill |
| T-LAH-021 | Mixed `position` plus scalar input with conflicting values uses normalized coordinates | stored boolean matches normalized row state |
| T-LAH-022 | Backfill repairs pre-existing rows with stale or default values | rows match expected truth table |
| T-LAH-023 | Backfill is idempotent on second run | no further value changes |
| T-LAH-024 | Backfill rejects `NULL`, zero, and negative batch sizes | error raised |
| T-LAH-025 | Applying the migration does not auto-run the historical backfill | pre-existing rows keep default or prior values until manual repair |
| T-LAH-026 | Insert move with non-existent author row (orphaned FK or deleted user) | `logged_at_author_home = false` or FK constraint violation prevents insert |

### Fixture notes

- use explicit `id` values in test fixtures
- use real `home_position` geography objects and corresponding `position` values on moves, compatible with the existing user-home trigger contract
- include test cases with positions at exactly 0m, exactly 50m, just under 50m, and just over 50m from home to validate the boundary
- include at least one author with no home position and one move with missing position
- for T-LAH-026 (non-existent author): test with an author_id that references a deleted user row, or rely on FK constraint to prevent the insert; either outcome validates the requirement
- pgTAP assertions must verify geography-based proximity matching using `ST_DWithin` semantics

## Verification Workflow

Implementation must follow the repository migration workflow and use the create-migration skill.

Minimum verification sequence:

1. Create the migration and pgTAP test file.
2. Run PHP syntax validation on the new migration.
3. Apply the migration with the Phinx workflow.
4. Verify the new column, trigger, and functions exist.
5. Roll back the migration.
6. Verify the new objects are removed.
7. Re-apply the migration.
8. Copy the database `geokrety` to `tests` using script `/home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh`.
9. Run the focused pgTAP test file.

Use the repository-standard Phinx and pgTAP workflow from the database migration docs and the existing migration skill guidance. Do not substitute ad hoc SQL spot checks for the focused pgTAP file.

## Out Of Scope

- API exposure of the new flag
- frontend usage of the new flag
- analytics aggregation based on the new flag
- automatic recomputation of all historical moves whenever a user edits their home profile
- fuzzy distance or radius-based "near home" detection

## Implementation Notes For Later Work

- Use the `create-migration` skill when implementation starts.
- Use the existing GeoKrety database migration conventions from the docs and recent Phinx migrations.
- Use pgTAP, not ad hoc SQL checks, for the behavioral contract above.
- When the implementation is finished, run the requested review loop in this order using agents: `specification` -> `quality-engineer` -> `system-architect`, `refactoring-expert`, `performance-engineer` and iterate until full resolution.
