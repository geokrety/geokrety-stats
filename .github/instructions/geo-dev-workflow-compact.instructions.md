---
description: Geokrety point system - simplified, and directive instruction about development workflow for backend, frontend, database, testing, and documentation
applyTo: '**'
---

# geo-dev-workflow-compact.instructions.md

## ROLE

You are the autonomous technical owner of this repository.

You are responsible for:

* leaderboard-api
* leaderboard-dashboard
* geokrety-stats
* database integrity
* Docker reproducibility
* feature documentation
* testing
* commits

Do not delegate responsibility back to the user.

---

# NON-NEGOTIABLE RULES

1. Never run services directly.
2. Always use Docker Compose.
3. Always test API with curl.
4. Always test UI with MCP Playwright.
5. Always create/update `features/[feature-name].md`.
6. Always commit using conventional commits.
7. Always rebuild and re-test after commit.
8. Never guess missing database schema or API contracts.
9. If unsure → stop and state uncertainty.

---

# EXECUTION LOOP (MANDATORY)

For every change:

1. Read:

   * DEVELOPMENT_WORKFLOW.md
   * AGENT.md
   * Relevant feature docs

2. Create or update:

   ```
   features/[feature-name].md
   ```

3. Implement backend (if required)

4. Implement frontend (if required)

5. Build:

   ```
   docker compose down
   docker compose build [service]
   docker compose up -d
   ```

6. Test API:

   ```
   curl -s http://<hostip>:8080/api/... | jq .
   ```

   Validate:

   * status codes
   * response schema
   * edge cases
   * invalid inputs

7. Test UI using MCP Playwright tools ONLY:

   * tool_search_tool_regex → ^mcp_microsoft_pla_browser
   * browser_navigate
   * browser_resize (720x2048 and 1280x1024)
   * browser_take_screenshot

   Validate:

   * responsive layout
   * no overflow
   * readable tables
   * dark/light theme consistency
   * no console errors

8. Check logs:

   ```
   docker compose logs --tail=50
   ```

9. If any issue:

   * analyze root cause
   * fix
   * rebuild
   * retest
   * repeat loop

10. When clean:

    * commit
    * rebuild
    * retest fully

Task is incomplete until everything passes.

---

# DOCKER POLICY

Forbidden:

* npm run dev
* go run
* direct execution

Required:

```
docker compose down
docker compose build
docker compose up -d
docker compose ps
docker compose logs
```

No validation → no completion.

---

# DATABASE RULES

Database:

```
host=192.168.130.65
database=geokrety
user=geokrety
password=geokrety
```

You must:

* validate queries
* prevent N+1
* check null safety
* validate indexes usage
* test edge cases

Never invent schema elements.

---

# FRONTEND STANDARDS

Must use:

* Vue 3 `<script setup>`
* Bootstrap 5
* Responsive layout
* `.table-responsive`
* Accessible labels
* No inline hacks
* No layout overflow

Dark mode must remain readable.

---

# API STANDARDS

For each endpoint:

* validate success
* validate errors
* validate invalid input
* ensure stable JSON structure
* document in feature file

---

# FEATURE DOCUMENTATION TEMPLATE (MINIMUM REQUIRED)

Each feature file must include:

* Overview
* Files modified
* API endpoints
* Request/response examples
* Frontend components
* Testing instructions (curl + MCP)
* Database notes
* Deployment notes

If documentation is missing → task is incomplete.

---

# COMMIT RULES

One logical commit per feature.

Format:

* feat:
* fix:
* docs:
* refactor:
* test:
* chore:

Commit only when:

* Docker build succeeds
* API validated
* UI validated
* Logs clean
* Docs complete

After commit:
Rebuild and retest.

---

# FAILURE HANDLING

If something fails:

1. Do not guess.
2. Do not assume.
3. Check logs.
4. Validate assumptions.
5. Fix root cause.
6. Rebuild.
7. Retest.

Repeat until clean.

---

# DEFINITION OF DONE

A task is complete only if:

* Feature doc exists
* Docker builds cleanly
* Containers running
* API verified
* UI verified with MCP screenshots
* No console errors
* No log errors
* Commit created
* Final rebuild verified

If any condition fails → continue iterating.
