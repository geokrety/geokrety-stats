# Log At Home Implementation Log

## Summary

- Goal: implement the `logged_at_author_home` database contract from the log-at-home specification.
- Status: complete.

## Progress

- [x] Read the specification and repository migration workflow.
- [x] Chose migration file `20260320065055_add_logged_at_author_home_to_gk_moves.php`.
- [x] Chose focused pgTAP slot `test-260-gk-moves-logged-at-author-home.sql`.
- [x] Drafted the Phinx migration for the new column, trigger function, trigger, and manual backfill function.
- [x] Drafted focused pgTAP coverage for schema assertions, trigger behavior, boundary distances, caller override handling, and manual backfill behavior.
- [x] Drafted `run_snapshot_backfill.py` integration for a full-history `--backfill-logged-at-author-home` command.
- [x] Updated the backfill implementation so each function call repairs at most one batch and the CLI commits after every batch.
- [x] Run the required review loop (`dba` -> `critical-thinking` -> `quality-engineer`) and address findings.
- [x] Validate the migration with `php -l`; `python -m py_compile` also passed for the CLI integration.
- [x] Apply the migration, verify rollback, re-apply, sync schema to `tests`, and run focused pgTAP.
- [x] Verify the CLI integration against `PGDATABASE=tests` and record the summary output.
- [x] Verify caller-managed batching end to end with a controlled `PGDATABASE=tests` fixture and `--batch-size 2`.

## Notes

- The historical repair remains manual by design; the migration creates the function but does not execute it.
- The dedicated pgTAP file disables `after_99_notify_amqp_moves` during fixtures to avoid notification side effects.
- Direct SQL calls to the backfill function are ordinary `UPDATE`s on `geokrety.gk_moves`.
- The standalone CLI path intentionally acquires its own advisory lock, runs each batch inside `SET LOCAL session_replication_role = replica`, and commits after every batch so large runs avoid one giant top-level transaction and suppress unrelated trigger side effects.
- The review loop caught and resolved an embedded-SQL CTE syntax error before the rollback and re-apply verification run.
- `pre-commit run -a` was attempted in both repositories, but the PHP Coding Standards Fixer hook did not return a final status in this environment even though the owned files remained syntactically clean.

## Validation

- `php -l website/db/migrations/20260320065055_add_logged_at_author_home_to_gk_moves.php` passed.
- `phinx migrate --count=1` passed; the column, trigger, and functions were verified present.
- `phinx rollback` passed; the column, trigger, and functions were verified absent.
- `phinx migrate --count=1` passed again after rollback verification.
- `/home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh` ran successfully.
- Focused pgTAP passed: `test-260-gk-moves-logged-at-author-home.sql` (`43` tests, `PASS`).
- Controlled CLI validation passed on `PGDATABASE=tests`: `python docs/database-refactor/run_snapshot_backfill.py --backfill-logged-at-author-home --batch-size 2` emitted two committed batch lines and the final summary `Processed 3 rows; 3 rows updated; 2 batches completed; full-history scope.`
