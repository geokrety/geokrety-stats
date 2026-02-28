# GeoKrety Points System: Development Workflow Guide

## Summary

I've created a comprehensive development workflow for the GeoKrety Points System with three main components:

### 1. **DEVELOPMENT_WORKFLOW.md** - The Complete Workflow Guide
A detailed 400+ line guide covering:
- Feature development workflow (5 phases)
- Testing procedures with curl, MCP Playwright, and make
- Docker Compose deployment steps
- Git commit conventions
- Common commands reference
- Success criteria and checklists

**Key Sections:**
- Phase 1: Specification & Documentation
- Phase 2: Implementation
- Phase 3: Testing (curl + MCP Playwright, NOT npx playwright test)
- Phase 4: Docker Deployment
- Phase 5: Git Commits
- Complete testing checklist
- Quick reference for common commands

### 2. **features/ Directory** - Feature Specification Storage
A new directory at project root containing standardized feature documentation:

**Files:**
- `features/README.md` - Directory overview and guidelines
- `features/country-leaderboard.md` - Country rankings feature spec
- `features/breakdown-charts.md` - Statistics charts feature spec
- `features/websocket-user-count.md` - User count feature spec

**Each feature file includes:**
- Overview and goals
- Files modified/created
- Complete API endpoint specifications with curl examples
- Frontend component documentation
- Testing procedures (curl + MCP Playwright examples)
- Database information
- WebSocket message formats
- Deployment notes
- Known limitations
- Future enhancements

### 3. **Updated AGENT.md** - AI Development Rules
Added comprehensive section with:
- Feature documentation directory reference
- Sequential workflow phases
- Testing requirements (explicit: NO npx playwright test, use MCP Playwright browser tools)
- Docker deployment rules
- Git commit conventions
- Critical DO/DON'T rules
- Success criteria
- Quick command reference

---

## 🚀 How to Use This Workflow

### For Implementing a New Feature

```bash
# 1. Create feature spec
touch features/my-feature.md
# (Use template from features/README.md)

# 2. Implement feature
# - Read the spec you created
# - Write backend code
# - Write frontend code
# - Update feature spec with final details

# 3. Test with provided commands (from feature spec)
curl -s http://<hostip>:8080/api/my-endpoint | jq .
# Use MCP Playwright browser tools for UI testing:
# - tool_search_tool_regex with pattern ^mcp_microsoft_pla_browser
# - mcp_microsoft_pla_browser_navigate to http://<hostip>:3000/my-route
# - mcp_microsoft_pla_browser_resize to 1280x1024
# - mcp_microsoft_pla_browser_take_screenshot

# 4. Deploy with docker compose (NOT npm dev or go run)
docker compose down
docker compose build
docker compose up -d

# 5. Commit with conventional format
git add .
git commit -m "feat: add my-feature

- Implement MyFeature.vue component
- Add /api/my-endpoint endpoint
- Include testing procedures in features/my-feature.md"
```

### Key Rules to Remember

✅ **DO:**
- Document features in `features/` directory
- Test API with curl
- Screenshot UI with MCP Playwright browser tools
- Use docker compose for deployment
- Create one commit per feature
- Read feature specs before implementing

❌ **DON'T:**
- Use npx playwright test (use MCP Playwright browser tools instead)
- Start services directly (use docker compose)
- Skip feature documentation
- Make multiple unrelated changes in one commit
- Assume code works without testing

---

## 📚 Documentation Reference

### Master Documents

1. **[DEVELOPMENT_WORKFLOW.md](DEVELOPMENT_WORKFLOW.md)**
   - Detailed step-by-step development process
   - Complete testing procedures
   - Docker commands and examples
   - Common commands quick reference

2. **[AGENT.md](AGENT.md)** (Updated)
   - AI development rules
   - Gamification rules (original content preserved)
   - Feature documentation guidelines
   - Prohibited/required practices

3. **[features/README.md](features/README.md)**
   - Feature directory overview
   - File naming conventions
   - Feature template
   - How AI uses feature files

### Feature Specifications

1. **[features/country-leaderboard.md](features/country-leaderboard.md)**
   - GET /api/users/countries endpoint
   - CountryLeaderboardView.vue component
   - Testing procedures with curl
   - MCP Playwright screenshot examples

2. **[features/breakdown-charts.md](features/breakdown-charts.md)**
   - GET /api/stats/breakdown endpoint
   - Four D3.js chart components
   - Testing all chart types
   - Data aggregation details

3. **[features/websocket-user-count.md](features/websocket-user-count.md)**
   - WebSocket /ws endpoint
   - connected_users message type
   - Frontend composable usage
   - Connection lifecycle

---

## 🧪 Testing Quick Reference

### API Testing with curl
```bash
# Basic endpoint test
curl -s http://<hostip>:8080/api/endpoint | jq .

# With query parameters
curl -s "http://<hostip>:8080/api/endpoint?param=value" | jq .

# Pretty print first result
curl -s http://<hostip>:8080/api/endpoint | jq '.[0]'
```

### UI Testing with MCP Playwright

**Load MCP Playwright tools first:**
```
tool_search_tool_regex with pattern: ^mcp_microsoft_pla_browser
```

**Desktop screenshot (1280px):**
1. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://<hostip>:3000/route`
2. Resize: `mcp_microsoft_pla_browser_resize` to 1280x1024
3. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

**Mobile screenshot (720px):**
1. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://<hostip>:3000/route`
2. Resize: `mcp_microsoft_pla_browser_resize` to 720x2048
3. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

### Binary Testing
```bash
# Build geokrety-stats
cd geokrety-stats && make build

# View help
./bin/geokrety-stats --help

# Run with options
./bin/geokrety-stats --flag=value --another=option
```

### Service Management
```bash
# Start all services
cd /home/kumy/GIT/geokrety-points-system
docker compose down
docker compose build
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f [service-name]

# Stop services
docker compose stop
```

---

## 🎯 Feature Documentation Template

Every feature file should include:

```markdown
# Feature: [Name]

**Status:** In Development / Testing / Complete
**Date Created:** YYYY-MM-DD
**Last Updated:** YYYY-MM-DD

## Overview
What does this do? Why was it built?

## Files Modified/Created
- Backend files
- Frontend files
- Database files

## API Endpoints
- GET /api/... with examples
- POST /api/... with examples

## Frontend Components
- Component name and location
- Props, events, usage

## Testing Procedures
- curl examples for API
- MCP Playwright examples for UI

## Database
- Views used
- Functions called

## Deployment Notes
- Prerequisites
- Build steps
- Verification

## Known Issues / Limitations
- List any issues

## Future Enhancements
- Planned improvements
```

See [features/country-leaderboard.md](features/country-leaderboard.md) for a complete example.

---

## 🔄 Git Commit Conventions

**Format:** `type: description`

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `refactor:` - Code restructuring
- `test:` - Testing
- `chore:` - Build, dependencies

**Examples:**
```bash
git commit -m "feat: add country leaderboard page

- Implement CountryLeaderboard.vue
- Add /api/users/countries endpoint
- Create features/country-leaderboard.md"

git commit -m "docs: document breakdown charts feature

- Create features/breakdown-charts.md
- Include API endpoint specification
- Add testing procedures"

git commit -m "fix: correct WebSocket message format

- Fix connected_users payload structure
- Update composable to handle count
- Test with MCP Playwright screenshots"
```

---

## 📊 Project Structure

```
/home/kumy/GIT/geokrety-points-system/
├── AGENT.md                           # Updated: AI rules (now includes dev workflow)
├── DEVELOPMENT_WORKFLOW.md            # NEW: Complete workflow guide
├── LEADERBOARD_FEATURES.md            # Feature summary document
│
├── features/                          # NEW: Feature documentation directory
│   ├── README.md                      # Directory guide
│   ├── country-leaderboard.md         # Feature spec
│   ├── breakdown-charts.md            # Feature spec
│   └── websocket-user-count.md        # Feature spec
│
├── leaderboard-api/                   # Go backend
│   └── internal/handlers/
├── leaderboard-dashboard/             # Vue frontend
│   └── src/
├── geokrety-stats/                    # CLI tool
│   └── Makefile                       # Build instructions
│
├── split/                             # Gamification rules (original)
└── migrations/                        # Database migrations
```

---

## ⚡ Speed Tips

**For Fast Implementation:**

1. Copy feature template from `features/README.md`
2. Fill in API specification first
3. Implement backend using spec
4. Implement frontend using spec
5. Test with curl examples from feature file
6. Test with MCP Playwright browser tools
7. Deploy with docker compose
8. Commit using conventional format

**All testing commands are already written in the feature files!**

---

## 🎓 Example: Complete Feature Development

### Request: Add new feature "User Activity Log"

**Step 1: Create feature spec** (features/user-activity-log.md)
```markdown
# Feature: User Activity Log
- Overview: Show recent actions by users
- API: GET /api/users/{id}/activity
- Component: UserActivityView.vue
- Testing: Provide curl and MCP Playwright examples
```

**Step 2: Implement backend**
- Create handler in leaderboard-api
- Update feature spec with final endpoint details

**Step 3: Implement frontend**
- Create Vue component
- Add route to router
- Style with Bootstrap

**Step 4: Test**
```bash
# From feature spec curl example
curl -s http://<hostip>:8080/api/users/123/activity | jq .

# From feature spec MCP Playwright example
# Load tools: tool_search_tool_regex with pattern ^mcp_microsoft_pla_browser
# Navigate: mcp_microsoft_pla_browser_navigate to http://<hostip>:3000/users/123/activity
# Resize: mcp_microsoft_pla_browser_resize to 1280x1024
# Screenshot: mcp_microsoft_pla_browser_take_screenshot
```

**Step 5: Deploy**
```bash
docker compose down && docker compose build && docker compose up -d
```

**Step 6: Commit**
```bash
git commit -m "feat: add user activity log page

- Implement UserActivityView.vue
- Add GET /api/users/{id}/activity endpoint
- Document in features/user-activity-log.md"
```

---

## 🚨 Important Reminders

### Testing Requirements

**BEFORE committing feature:**
- [ ] All curl API tests pass
- [ ] MCP Playwright screenshots look correct
- [ ] Docker build succeeds
- [ ] Services start without errors
- [ ] No errors in docker logs
- [ ] Feature file is complete and up-to-date
- [ ] Commit message follows conventions

### Prohibited Practices

❌ **DO NOT:**
- Use npx playwright test (→ use MCP Playwright browser tools instead)
- Run services directly (→ use docker compose)
- Skip feature documentation (→ create features/[name].md first)
- Test without curl/screenshots (→ follow testing procedures)
- Commit without testing (→ verify each step)

### Required Practices

✅ **MUST:**
- Document features in features/ directory
- Test with curl (examples in feature specs)
- Screenshot with MCP Playwright browser tools
- Deploy with docker compose
- Create one feature per commit
- Use conventional commit format

---

## 📞 Quick Links

| Document | Purpose |
|----------|---------|
| [DEVELOPMENT_WORKFLOW.md](DEVELOPMENT_WORKFLOW.md) | Complete workflow guide |
| [AGENT.md](AGENT.md) | AI development rules |
| [features/README.md](features/README.md) | Feature directory guide |
| [features/country-leaderboard.md](features/country-leaderboard.md) | Example feature spec |
| [features/breakdown-charts.md](features/breakdown-charts.md) | Example feature spec |
| [features/websocket-user-count.md](features/websocket-user-count.md) | Example feature spec |

---

## 🎉 You're Ready!

The development workflow is now fully documented and automated:

1. ✅ **AI knows how to develop features** (AGENT.md + DEVELOPMENT_WORKFLOW.md)
2. ✅ **Features are pre-documented before coding** (features/ directory)
3. ✅ **Testing is guided by feature specs** (curl + MCP Playwright examples)
4. ✅ **Deployment is standardized** (docker compose commands)
5. ✅ **Git history is clean** (one feature per commit)

**Next steps:**
- For new features, create `features/[name].md` first
- Follow the workflow in DEVELOPMENT_WORKFLOW.md
- Use curl and MCP Playwright browser tools for testing
- Deploy with docker compose (not direct npm/go)
- Commit per feature with conventional format

---

**Created:** 2026-02-28
**Last Updated:** 2026-02-28
**Status:** Ready for Production
