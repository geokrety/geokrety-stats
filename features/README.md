# Features Directory

This directory contains comprehensive documentation for all features implemented in the GeoKrety Points System.

## Purpose

Each feature file serves as:
1. **AI Context** - Provides specification for development
2. **Integration Spec** - Documents API endpoints and component structure
3. **Testing Guide** - Includes curl examples and verification procedures
4. **Deployment Runbook** - Lists files modified and deployment steps
5. **Future Reference** - Enables maintenance and extensions

## File Naming Convention

Use kebab-case (lowercase with hyphens):
```
features/country-leaderboard.md
features/breakdown-charts.md
features/websocket-user-count.md
features/[future-feature].md
```

## Feature File Template

Every feature file should include:

```markdown
# Feature: [Feature Name]

**Status:** In Development / Testing / Complete
**Date Created:** YYYY-MM-DD
**Last Updated:** YYYY-MM-DD

## Overview
What does this feature do? Why was it built?

## Files Modified/Created
- List all files changed
- Include backend, frontend, database files

## API Endpoints
Document each endpoint:
- HTTP method and path
- Query/path parameters
- Request/response examples
- Error cases

## Frontend Components
Document each component:
- Location (path)
- Props
- Events
- Usage

## Testing Procedures
Include:
- curl examples for API testing
- MCP Playwright browser tool commands
- Expected results

## Database
- Any views created or modified
- Any functions used
- Migration files

## Deployment Notes
- Any special considerations
- Service dependencies
- Known issues
```

## Existing Features

### Country Leaderboard
**File:** [country-leaderboard.md](country-leaderboard.md)
**Status:** Complete
**Description:** Display top 50 countries ranked by total points

### Breakdown Charts
**File:** [breakdown-charts.md](breakdown-charts.md)
**Status:** Complete
**Description:** Interactive charts showing points, moves, costs, and events

### WebSocket Connected Users
**File:** [websocket-user-count.md](websocket-user-count.md)
**Status:** Complete
**Description:** Real-time display of active dashboard users

### Chain Details Navigation
**File:** [chain-details-navigation.md](chain-details-navigation.md)
**Status:** Complete
**Description:** Chain detail pages and cross-links from users, GeoKrety, and moves views

## How AI Uses These Files

The AI assistant will:

1. **Before Implementation**
   - Read the feature spec to understand requirements
   - Use API endpoints as specification
   - Follow component structure

2. **During Implementation**
   - Reference curl examples for testing
   - Use MCP Playwright browser tools for visual verification
   - Follow documented patterns

3. **After Implementation**
   - Update feature file with final details
   - Verify all endpoints work as documented
   - Ensure all tests pass

4. **For Maintenance**
   - Use feature files as source of truth
   - Understand dependencies between features
   - Make backward-compatible changes

## Adding New Features

1. **Create feature file** in `features/[name].md`
2. **Add to this README** with brief description
3. **Implement feature** following DEVELOPMENT_WORKFLOW.md
4. **Update feature file** with implementation details
5. **Git commit** with conventional commit format
6. **Deploy** using docker compose

## Quick Reference: curl Commands

### API Testing
```bash
# Basic request
curl -s http://<hostip>:8080/api/path | jq .

# With query parameters
curl -s "http://<hostip>:8080/api/path?param=value" | jq .

# POST with JSON
curl -X POST http://<hostip>:8080/api/path \
  -H 'Content-Type: application/json' \
  -d '{"key": "value"}' | jq .
```

### UI Screenshot with MCP Playwright

**Load MCP Playwright tools first:**
Use `tool_search_tool_regex` with pattern: `^mcp_microsoft_pla_browser`

**Mobile view (720px wide):**
1. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://<hostip>:3000/route`
2. Resize: `mcp_microsoft_pla_browser_resize` to 720x2048
3. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

**Desktop view (1280px wide):**
1. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://<hostip>:3000/route`
2. Resize: `mcp_microsoft_pla_browser_resize` to 1280x1024
3. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

## Docker Compose Quick Reference

```bash
# Rebuild and restart all services
docker compose down && docker compose build && docker compose up -d

# Rebuild specific service
docker compose build leaderboard-api && docker compose up -d leaderboard-api

# View logs
docker compose logs -f [service-name]

# Check status
docker compose ps
```

## Important Rules

✅ **DO:**
- Document all API endpoints in feature files
- Use curl for API testing
- Use MCP Playwright browser tools for UI screenshots
- Use docker compose for deployment (not direct npm/go run)
- Use make build for geokrety-stats binary
- Create one git commit per feature
- Keep feature files updated with implementation details

❌ **DON'T:**
- Use npx playwright test (use MCP Playwright browser tools instead)
- Start frontend/backend directly (npm dev, go run)
- Forget to document feature before/after implementation
- Make multiple unrelated changes in one commit
- Skip testing procedures

## File Organization

```
features/
├── README.md                          # This file
├── country-leaderboard.md             # Feature spec
├── breakdown-charts.md                # Feature spec
├── websocket-user-count.md            # Feature spec
└── [future-feature].md                # Feature spec
```

All feature files are checked by AI when developing new features or maintaining existing ones.

---

**Last Updated:** 2026-02-28
**Maintained By:** Development Team
