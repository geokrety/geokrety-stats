---
name: pgtap
description: 'Run pgTAP SQL tests using the repository wrapper. Use when asked to "run pgtap tests", "execute db tests", "run tests for migration test-NNN-*.sql", or "run pg_prove on tests directory". Works together with the create-migration skill to execute the exhaustive tests authored for a migration.'
user-invocable: true
---

# PGTAP Skill
## Purpose
- Provide a clear, minimal contract for running the project's pgtap tests via the wrapper script.

## Wrapper script (where it lives)
- The wrapper script for this skill is located at `./scripts/pgtap.sh` (relative to this `SKILL.md`). Invoke it from the repository root using the explicit path to avoid ambiguity:

  ```bash
  ./.github/skills/pgtap/scripts/pgtap.sh [args...]
  ```

## Contract / behavior
- The wrapper runs `pg_prove` with the repository defaults and search_path required for the tests. By default it runs the equivalent of:

  ```bash
  PGPASSWORD=geokrety \
    PGOPTIONS=--search_path=public,pgtap,geokrety \
    pg_prove -d tests -U geokrety -h localhost -ot ${WORKSPACE:-.}/tests/test*.sql
  ```

- Behavior details:
  - If the environment variable `GEOKRETY_DB_TESTS_PATH` is defined, the wrapper will run tests from that path instead of the default `./tests` directory.
  - The wrapper respects common Postgres environment overrides: `PGPASSWORD`, `PGDATABASE`, `PGUSER`, `PGHOST` can be set by callers to change credentials/target.
  - Any additional args passed to the wrapper are forwarded to `pg_prove` (shell expansion applies to the final test glob).

## Usage examples (explicit)
- Run tests from repo root (default path):

  ```bash
  ./.github/skills/pgtap/scripts/pgtap.sh
  ```

- Run tests from a custom path (exported env):

  ```bash
  GEOKRETY_DB_TESTS_PATH=/path/to/tests ./.github/skills/pgtap/scripts/pgtap.sh
  ```

- Override DB credentials for a single run:

  ```bash
  PGPASSWORD=secret PGUSER=otheruser ./.github/skills/pgtap/scripts/pgtap.sh
  ```

## Notes for automation and AI actors
- Always prefer the explicit `./.github/skills/pgtap/scripts/pgtap.sh` path in automation to avoid invoking a system-wide `pg_prove` or other wrappers.
- This wrapper intentionally implements the repository's canonical `pg_prove` invocation; automation should not reimplement command-line flags unless explicitly required.
 - `direnv` may load and export environment variables (for example `GEOKRETY_DB_TESTS_PATH` or `PGPASSWORD`) when entering the repository; this is expected. Run the wrapper immediately after `direnv` has loaded the environment so the exported variables are available to the test run.

## Commands supported
- `run` (default): call `./.github/skills/pgtap/scripts/pgtap.sh` with optional extra args to pass through to `pg_prove`.

Examples above are copy-paste ready and unambiguous about the wrapper location.

## Integration with create-migration skill
- When a new migration is created via the `create-migration` skill, the pgTAP tests are authored as part of that workflow (Step 5) and run here (Step 8).
- This wrapper runs **all** test files matching `test*.sql`; if you want to run only the tests for the new migration, pass the specific file path:
  ```bash
  GEOKRETY_DB_TESTS_PATH=/home/kumy/GIT/geokrety-website/website/db/tests \
    ./.github/skills/pgtap/scripts/pgtap.sh -- tests/test-NNN-description.sql
  ```
- See `.github/skills/create-migration/SKILL.md` for the full test authoring workflow including:
  - `SELECT plan(N)` counting strategy
  - Test isolation caveats (sequences, `pg_notify()` side-effects)
  - Coverage targets (trigger branches, constraints, edge cases)
  - pgTAP assertion quick reference and PostgreSQL error codes

## pgTAP test file conventions in this project
- All test files live in `/home/kumy/GIT/geokrety-website/website/db/tests/`
- Naming: `test-NNN-description.sql` (NNN = numeric block, see NNN Block Map in create-migration skill)
- Every test file wraps all operations in `BEGIN; ... ROLLBACK;` for isolation
- Fixtures always use explicit `id` values — never `DEFAULT` serial — to avoid sequence leakage across test runs
- `SELECT plan(N)` is required in committed test files; `no_plan()` is acceptable only during authoring
