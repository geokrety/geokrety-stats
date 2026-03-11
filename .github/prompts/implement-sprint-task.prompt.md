---
agent: 'agent'
tools: [vscode/askQuestions, execute/runInTerminal, read/problems, read/readFile, agent, edit/createFile, edit/editFiles, search/changes, search/codebase, search/fileSearch, search/listDirectory, search/textSearch, postgresql-mcp/pgsql_connect, postgresql-mcp/pgsql_disconnect, postgresql-mcp/pgsql_list_databases, postgresql-mcp/pgsql_query]
description: 'request an end-to-end migration implementation for a specific task'
---

Please run the `create-migration` skill to implement the migration described for the sprint task (given by user instructions) in `./docs/database-refactor/sprint-x/task-SxTyy.md`, where `x` is the sprint number and `yy` is the task number.

Requirements:
- Implement the Phinx migration PHP file according to the task specification and project conventions.
- Use skill `create-migration` to implement exhaustive pgTAP tests covering schema, constraints, trigger behavior, and edge cases; place tests in `/home/kumy/GIT/geokrety-website/website/db/tests/` using the proper NNN slot.
- Update `/home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh` if the migration adds data that must be copied into the test DB for the tests to run.
- Maintain and update `/home/kumy/GIT/geokrety-stats/docs/database-refactor/99-IMPLEMENTATION.md` with advancement entries and checkboxes as work proceeds.
- Apply the migration with the phinx wrapper, copy schema to tests, and run pgTAP. Iterate until all tests pass or a non-trivial open question remains.
- If the task completes successfully, automatically continue to the next task in the given sprint directory `./docs/database-refactor/sprint-x/` and repeat.

Review loop (must be executed before applying migrations):
1) Invoke the `dba` agent for a safety/performance/reversibility review of the migration file and tests.
2) Invoke the `critical-thinking` agent to challenge assumptions and `down()` reversibility.
3) Invoke the `quality-engineer` agent to verify test coverage and plan counts.
If any agent raises an unresolved concern, log it in `99-OPEN-QUESTIONS.md` and present it to the human operator. Do not apply the migration when there are blocking unresolved safety issues.

Execution commands (do not run outside repo wrappers):
`.github/skills/phinx/scripts/phinx.sh migrate --count=1`
`/home/kumy/GIT/geokrety-website/website/db/tests-copy-schema-geokrety-to-tests.sh`
`.github/skills/pgtap/scripts/pgtap.sh`

Return format: Provide a concise execution summary that includes:
- Files added/modified with workspace-relative paths.
- `99-IMPLEMENTATION.md` entries added or updated and their checklist status.
- Migration applied? (yes/no) and rollback verified? (yes/no)
- pgTAP result summary (pass/fail) and failing assertions (if any).
- Next suggested action.
