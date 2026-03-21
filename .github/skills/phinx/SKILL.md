---
name: phinx
description: 'Manage Phinx DB migrations up and down. Use when asked to "run phinx migrate", "rollback last migration", "check migration status", "apply one migration", or "run phinx with --count=1", "migrate up", "migrate down". Works together with the create-migration skill for creating migration files and with the pgtap skill for running tests.'
user-invocable: true
---

# Phinx Migration Skill
## Purpose
- Provide clear, unambiguous instructions for invoking the Phinx wrapper shipped with this skill.

## Wrapper script (where it lives)
- The wrapper script for this skill is located in the same directory at `./scripts/phinx.sh` (relative to this `SKILL.md`). To avoid ambiguity, use the explicit repository-root path when invoking it from automation or CI:

- From the repository root:

```bash
cd ${workspace} # ensure you're at the repo root
./.github/skills/phinx/scripts/phinx.sh <command> [args...]
```

- Or, if you `cd` into the skill directory first:

```bash
cd ${workspace}/.github/skills/phinx/scripts/ # ensure you're at the repo root
./phinx.sh <command> [args...]
```

## Contract / behavior
- The wrapper forwards arguments to the real Phinx executable inside the project's running stack container, using the project's `phinx.php` configuration (if present). It returns the underlying command's exit status.
- If the container or Phinx binary is unavailable the wrapper prints a short helpful message and exits non-zero.

## Usage examples (explicit)
- Check status (from repo root):

```bash
./.github/skills/phinx/scripts/phinx.sh status
```

- Apply all pending migrations (from repo root):

```bash
./.github/skills/phinx/scripts/phinx.sh migrate
```

- Roll back the last migration (from repo root):

```bash
./.github/skills/phinx/scripts/phinx.sh rollback
```

## Notes for automation and AI actors
- Always prefer the explicit `./.github/skills/phinx/scripts/phinx.sh` path in automation to avoid invoking a different `phinx` binary that might be installed system-wide.

## Migration policy (default behavior)
- This repository applies migrations one-at-a-time by default. Automated runs (including AI actors and CI) MUST invoke `migrate` with `--count=1` unless the user explicitly requests applying multiple migrations at once.

- Examples (repo root) — default, safe one-by-one migration:

```bash
./.github/skills/phinx/scripts/phinx.sh migrate --count=1
```

- If a user explicitly requests batching multiple migrations, they must state the desired `--count` or set no `--count` at all to migrate all pending migrations; automation will then run exactly what the user specified.

## Commands supported
- `help`: run `./.github/skills/phinx/scripts/phinx.sh` to show wrapper help
- `migrate`: `./.github/skills/phinx/scripts/phinx.sh migrate` — apply pending migrations
- `rollback`: `./.github/skills/phinx/scripts/phinx.sh rollback` — roll back last migration
- `status`: `./.github/skills/phinx/scripts/phinx.sh status` — show migration status

Examples above are copy-paste ready and unambiguous about the wrapper location.

## Integration with create-migration skill
- When creating a new migration via the `create-migration` skill, the workflow calls:
  1. `./.github/skills/phinx/scripts/phinx.sh migrate --count=1` — apply the new migration
  2. `./.github/skills/phinx/scripts/phinx.sh rollback` — verify `down()` reversibility (Step 7b)
  3. `./.github/skills/phinx/scripts/phinx.sh migrate --count=1` — re-apply after down() check
- If `up()` fails, check status with `status` before retrying — do not retry the same broken migration.
- If `rollback` fails after migration, the `down()` method has a bug and must be fixed before tests can run.

## Generating comprehensive pgTAP tests alongside migrations
- When applying a migration, also ensure the companion pgTAP test file is created per the `create-migration` skill workflow.
- See `.github/skills/create-migration/SKILL.md` for the full lifecycle including test authoring.
- See `.github/skills/pgtap/SKILL.md` for running those tests.
