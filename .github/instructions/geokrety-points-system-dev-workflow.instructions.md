---
description: Geokrety point system - development workflow instructions for backend, frontend, database, testing, and documentation
applyTo: '**'
---

# geokrety-points-system-dev-workflow.instructions.md

## Scope

The agent is responsible for:

* Managing **leaderboard-api**
* Managing **leaderboard-dashboard**
* Managing **geokrety-stats**
* Managing feature documentation
* Managing Docker-based integration
* Managing database interactions
* Managing commits
* Following DEVELOPMENT_WORKFLOW.md strictly

The agent owns the full lifecycle until completion.

---

# Execution Model (Mandatory)

You MUST operate autonomously and iteratively.

For every feature, bugfix, refactor, or improvement:

1. Read:

   * `DEVELOPMENT_WORKFLOW.md`
   * `AGENT.md`
   * Relevant files in `features/`
2. Create or update a feature spec in `features/[feature-name].md`
3. Implement backend if required
4. Implement frontend if required
5. Update documentation
6. Build via Docker Compose (never local direct execution)
7. Test API with curl
8. Test UI with MCP Playwright tools
9. Validate database behavior
10. Check logs
11. Iterate until no issues remain
12. Commit using conventional commit format
13. Rebuild and re-test
14. Final verification

If something fails:

* Analyze root cause
* Fix
* Rebuild
* Retest
* Repeat until resolved

Stopping before full validation is not allowed.

---

# Docker Rules (Strict)

Never run services directly.

❌ Forbidden:

* npm run dev
* go run
* node index.js

✅ Required:

```
docker compose down
docker compose build [service]
docker compose up -d
```

Before finalizing:

```
docker compose down
docker compose build
docker compose up -d
docker compose ps
docker compose logs --tail=50
```

If logs contain errors → fix → rebuild → retest.

---

# Database Access

Database credentials:

```
host="192.168.130.65"
database="geokrety"
user="geokrety"
password="geokrety"
```

You must:

* Validate queries
* Ensure indexes are used
* Avoid N+1
* Confirm expected results
* Test edge cases
* Validate null handling
* Validate constraints

If schema changes are required:

* Document them in feature file
* Ensure compatibility
* Test migrations

---

# API Testing Requirements

For each endpoint:

* Test happy path
* Test invalid parameters
* Test missing parameters
* Test boundary values
* Validate HTTP codes
* Validate JSON structure
* Ensure consistent response schema

Use:

```
curl -s http://<hostip>:8080/api/... | jq .
```

Never assume endpoint correctness without testing.

---

# UI Testing Requirements (MCP Playwright Mandatory)

Do NOT use:

* npx playwright test

You MUST use MCP Playwright browser tools:

Workflow:

1. Load tools via tool_search_tool_regex using pattern:
   ^mcp_microsoft_pla_browser

2. Navigate:
   mcp_microsoft_pla_browser_navigate

3. Resize:

   * Mobile: 720x2048
   * Desktop: 1280x1024

4. Screenshot:
   mcp_microsoft_pla_browser_take_screenshot

5. Validate:

   * Responsive layout
   * No overflow
   * No hidden elements
   * No broken tables
   * No JS console errors
   * Accessibility basics
   * Dark/light theme consistency

Repeat until layout is correct.

---

# Frontend Standards

Frontend must:

* Use Vue 3 with `<script setup>`
* Use Bootstrap 5
* Be fully responsive
* Avoid layout overflow
* Use semantic HTML
* Use accessible labels
* Avoid inline styles
* Use proper component separation

Tables must:

* Use `.table`
* Be wrapped in `.table-responsive`
* Be readable in dark mode
* Avoid overflow on mobile

---

# Feature Documentation (Mandatory)

Before implementation:
Create:

```
features/[feature-name].md
```

It must include:

* Overview
* Files touched
* API endpoints
* Request/response examples
* Frontend components
* Testing instructions (curl + MCP)
* Database notes
* Deployment notes
* Known limitations

Documentation is not optional.

---

# Iteration Loop

You must internally execute this loop:

```
implement
build
deploy
test API
test UI
test DB
check logs
fix issues
repeat
```

Continue until:

* No Docker errors
* No API errors
* No UI issues
* No layout issues
* No accessibility blockers
* No console errors
* No failing scenarios

Only then commit.

---

# Commit Rules

One logical commit per feature.

Format:

```
feat: short description
fix: short description
docs: short description
refactor: short description
test: short description
chore: short description
```

Commit must:

* Represent complete feature
* Include documentation
* Include working code
* Pass tests
* Build successfully via Docker

After commit:
Rebuild and verify again.

---

# Error Handling Policy

If something fails:

1. Do not guess.
2. Analyze.
3. Check logs.
4. Validate assumptions.
5. Fix root cause.
6. Re-test.

If uncertain:
Admit uncertainty and request clarification.

Never invent database fields, endpoints, or behaviors.

---

# Quality Standard

The result must:

* Build cleanly
* Run cleanly
* Have no obvious technical debt
* Be production-safe
* Be documented
* Be reproducible via Docker
* Be testable via curl
* Be visually validated via MCP

---

# Responsibility Clause

You are responsible for managing:

* Backend
* Frontend
* CLI tool
* Docker
* Database
* Documentation
* Testing
* Commits

Do not delegate responsibility to the user.

Complete the loop fully.

---

# Completion Definition

A task is complete only when:

* Feature documentation exists
* Docker build succeeds
* Containers run without error
* API tested and validated
* UI tested via MCP screenshots
* Database verified
* Logs clean
* Commit performed
* Final rebuild successful

If any condition fails → the task is not complete.
