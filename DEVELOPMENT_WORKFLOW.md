# GeoKrety Points System - Development Workflow Guide

**Purpose:** Standardized development, testing, and deployment process for the GeoKrety Points System to ensure quality, consistency, and reproducibility.

**Last Updated:** 2026-02-28
**Version:** 1.0

---

## 📋 Overview

The development workflow follows a structured approach:

1. **Implement Feature** - Write code and create feature documentation
2. **Test Feature** - Verify functionality through multiple testing methods
3. **Document Feature** - Create comprehensive feature specs in dedicated directory
4. **Commit Changes** - Git commit per completed feature
5. **Deploy** - Use Docker Compose for reproducible deployment

---

## 🚀 Feature Development Workflow

### Phase 1: Implementation

#### 1.1 Create Feature Documentation First
**Location:** `features/[feature-name].md`

Before writing code, document:
- Feature overview and goals
- What endpoints are needed
- What components are needed
- How data flows through the system
- Expected API responses
- UI mockups if applicable

**Benefits:**
- AI can automatically read and understand the feature
- Clarifies requirements before coding
- Serves as specification during implementation
- Simplifies testing and deployment guidance

#### 1.2 Implement Backend (if needed)
```bash
# Edit relevant Go files
# For leaderboard-api:
cd leaderboard-api/
# Modify internal/handlers/*.go
# Modify internal/models/*.go
# etc.
```

Ensure:
- Code follows existing patterns
- Proper error handling
- Database queries optimized
- New endpoints documented in feature file

#### 1.3 Implement Frontend (if needed)
```bash
# For leaderboard-dashboard:
cd leaderboard-dashboard/src/
# Create/modify Vue components
# Update router/index.js if adding routes
# Update composables if needed
```

Ensure:
- Vue 3 Composition API with `<script setup>`
- Responsive design (Bootstrap 5)
- Proper error handling
- Components properly typed

#### 1.4 Update Feature Documentation
Add to `features/[feature-name].md`:
- Actual API endpoint specifications
- Component file locations
- Code examples
- Integration notes
- Any deviations from initial plan

---

### Phase 2: Testing

#### 2.1 Test REST API Endpoints (curl)

**Quick API Test Template:**
```bash
# Test basic endpoint
curl -s http://<hostip>:8080/api/endpoint | jq .

# Test with query parameters
curl -s "http://<hostip>:8080/api/endpoint?param=value" | jq .

# Test POST/PUT with JSON
curl -X POST http://<hostip>:8080/api/endpoint \
  -H 'Content-Type: application/json' \
  -d '{"key": "value"}' | jq .

# Test with headers
curl -s -H 'Authorization: Bearer token' http://<hostip>:8080/api/endpoint | jq .
```

**Important:**
- Use `jq` for pretty-printing JSON
- Always test against running API (`docker compose up`)
- Test error cases (invalid input, missing params, etc.)
- Verify response format matches documentation

#### 2.2 Test Binary CLI Tool (geokrety-stats)

**Build and Run:**
```bash
cd /home/kumy/GIT/geokrety-points-system/geokrety-stats

# Build using Makefile
make build

# Get binary help and options
./bin/geokrety-stats --help

# Run with specific options
./bin/geokrety-stats --config=path/to/config --dry-run
```

**What to verify:**
- Binary builds without errors
- Help output is clear
- All flags work as documented
- Output format is correct
- Error handling works for edge cases

#### 2.3 Test Frontend UI (Visual Testing with MCP Playwright)

**DO NOT use npx playwright test.** Use MCP Playwright browser tools for screenshots and interaction:

**Steps:**
1. Load MCP Playwright tools: Use `tool_search_tool_regex` with pattern `^mcp_microsoft_pla_browser`
2. Navigate to page: `mcp_microsoft_pla_browser_navigate` with URL `http://<hostip>:3000/users/4559`
3. Resize viewport (optional): `mcp_microsoft_pla_browser_resize` with width/height (e.g., 720x2048 for mobile, 1280x1024 for desktop)
4. Take screenshot: `mcp_microsoft_pla_browser_take_screenshot`
5. Interact if needed: `browser_click`, `browser_fill_form`, `browser_evaluate`, etc.

**Available MCP Playwright Tools:**
- `browser_navigate` - Navigate to URL
- `browser_resize` - Change viewport size
- `browser_take_screenshot` - Capture page
- `browser_click` - Click elements
- `browser_fill_form` - Fill input fields
- `browser_evaluate` - Run JavaScript
- `browser_wait_for` - Wait for elements
- `browser_snapshot` - Get page state

**Verification Checklist:**
- [ ] Layout is responsive (test with 720x2048 for mobile, 1280x1024 for desktop)
- [ ] All text is visible and readable
- [ ] Images/icons render correctly
- [ ] Buttons and links are clickable
- [ ] Colors match design system
- [ ] No JavaScript errors in console
- [ ] Data displays correctly
- [ ] Navigation works

#### 2.4 Integration Testing

**Complete Integration Test:**
```bash
# 1. Start fresh containers
cd /home/kumy/GIT/geokrety-points-system
docker compose down
docker compose build [service-name]
docker compose up -d

# 2. Wait for services to be ready
sleep 5

# 3. Test API endpoint
curl -s http://<hostip>:8080/api/health | jq .

# 4. Take UI screenshot with MCP Playwright
# Load tools: tool_search_tool_regex with pattern ^mcp_microsoft_pla_browser
# Navigate: mcp_microsoft_pla_browser_navigate to http://<hostip>:3000/path/to/feature
# Resize: mcp_microsoft_pla_browser_resize to 1280x1024
# Screenshot: mcp_microsoft_pla_browser_take_screenshot

# 5. Verify with curl
curl -s http://<hostip>:8080/api/feature-endpoint | jq .

# 6. Check logs for errors
docker compose logs [service-name] | tail -20
```

**Test Scenarios:**
- Happy path (expected input → expected output)
- Edge cases (boundary values, empty responses)
- Error handling (invalid input, missing data)
- Performance (response times reasonable)

---

### Phase 3: Docker Compose Workflow

#### 3.1 DO NOT Start Services Directly

❌ **WRONG:**
```bash
# DO NOT do this
npm run dev
go run ./cmd/api
```

✅ **CORRECT:**
```bash
# Always use Docker Compose
docker compose build [service-name]
docker compose up -d
```

#### 3.2 Docker Compose Commands

**Clean Build and Deploy:**
```bash
# Stop and remove all containers/networks
docker compose down

# Rebuild specific service with fresh image
docker compose build leaderboard-api
docker compose build leaderboard-dashboard

# Start all services in background
docker compose up -d

# Verify all services are running
docker compose ps
```

**Useful Commands:**
```bash
# View logs for a service
docker compose logs -f leaderboard-api

# View logs for all services
docker compose logs -f

# Stop all services
docker compose stop

# Start stopped services
docker compose start

# Rebuild and restart single service
docker compose up -d --build leaderboard-dashboard

# Remove volumes (WARNING: deletes data)
docker compose down -v
```

#### 3.3 Service Ports

| Service | Port | URL |
|---------|------|-----|
| leaderboard-api | 8080 | http://<hostip>:8080 |
| leaderboard-dashboard | 3000 | http://<hostip>:3000 |

---

### Phase 4: Feature Documentation

#### 4.1 Create Feature File in `features/` Directory

**Naming Convention:** `features/[feature-name].md`

**Template:**
```markdown
# Feature: [Feature Name]

**Status:** In Development / Testing / Complete
**Date Created:** YYYY-MM-DD
**Last Updated:** YYYY-MM-DD

## Overview
Brief description of what this feature does.

## Files Modified/Created
- `path/to/file.go`
- `path/to/component.vue`
- `migrations/00000X_something.sql`

## API Endpoints

### Endpoint 1: List Items
```
GET /api/items
```

**Query Parameters:**
- `limit` (int, optional, default: 50)
- `offset` (int, optional, default: 0)

**Response:**
```json
[
  {
    "id": 1,
    "name": "Item 1",
    "value": 100
  }
]
```

**Error Cases:**
- 400: Invalid parameter
- 500: Database error

### Endpoint 2: Get Item Details
```
GET /api/items/:id
```

**Response:**
```json
{
  "id": 1,
  "name": "Item 1",
  "details": {...}
}
```

## Frontend Components

### Component: ItemList.vue
**Location:** `leaderboard-dashboard/src/views/ItemList.vue`

**Props:**
- `items` (Array) - List of items to display
- `loading` (Boolean) - Show loading state

**Events:**
- `select` - Emitted when item is clicked

**Features:**
- Responsive table layout
- Pagination
- Sort by column

### Component: ItemDetail.vue
**Location:** `leaderboard-dashboard/src/views/ItemDetail.vue`

**Features:**
- Display full item details
- Edit capability
- Delete with confirmation

## Testing

### API Testing
```bash
# Test list endpoint
curl -s http://<hostip>:8080/api/items | jq .

# Test get endpoint
curl -s http://<hostip>:8080/api/items/1 | jq .
```

### UI Testing

Use MCP Playwright browser tools for UI testing:

**Screenshot list view:**
1. Load tools: `tool_search_tool_regex` with pattern `^mcp_microsoft_pla_browser`
2. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://<hostip>:3000/items`
3. Resize: `mcp_microsoft_pla_browser_resize` to 1280x1024
4. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

**Screenshot detail view:**
1. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://<hostip>:3000/items/1`
2. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

## WebSocket Messages (if applicable)

**Message Type:** `items_update`
```json
{
  "type": "items_update",
  "payload": {
    "items": [...],
    "timestamp": "2026-02-28T10:30:00Z"
  }
}
```

## Database
- View: `mv_item_stats`
- Materialized view refresh: `geokrety_stats.refresh_leaderboard_views()`

## Known Issues / Limitations
- None currently

## Future Enhancements
- Search functionality
- Export to CSV
- Advanced filtering

## Deployment Notes
- Requires `docker compose build && docker compose up -d`
- No manual restarts needed
- See DEVELOPMENT_WORKFLOW.md for testing procedures
```

#### 4.2 Documentation Standards

**All feature files must include:**
- Clear overview of functionality
- Complete file listing
- All API endpoints with request/response examples
- All UI components with usage instructions
- Testing procedures (curl examples)
- Database views/functions used (if any)
- Known limitations
- Deployment instructions

**Documentation serves as:**
- AI context for future modifications
- Onboarding guide for new developers
- Integration specification
- Testing checklist
- Deployment runbook

---

### Phase 5: Git Commits

#### 5.1 Commit Per Feature

**Strategy:** One logical commit per completed feature

```bash
# Stage all changes for this feature
git add .

# Commit with conventional commit message
git commit -m "feat: add country leaderboard page

- Implement CountryLeaderboard.vue component
- Add /api/users/countries endpoint
- Add /countries route
- Display top 50 countries by points
- Include responsive table design"
```

#### 5.2 Conventional Commit Format

**Format:** `<type>: <short description>`

**Types:**
- `feat: ` - New feature
- `fix: ` - Bug fix
- `docs: ` - Documentation
- `refactor: ` - Code refactoring without changing functionality
- `test: ` - Testing
- `chore: ` - Build, dependencies, etc.

**Example Commits:**
```bash
# Feature commit
git commit -m "feat: add breakdown charts to statistics page

- Create StatsBreakdown.vue with 4 D3.js charts
- Add /api/stats/breakdown endpoint
- Include points, moves, cost, event distributions
- Make all charts responsive with tooltips"

# Documentation commit
git commit -m "docs: add feature documentation for charts

- Create features/breakdown-charts.md
- Document API endpoints and response format
- Include testing procedures"

# Fix commit
git commit -m "fix: correct WebSocket user count broadcast

- Send connected_users every 10 seconds
- Fix message format in frontend
- Verify count updates in footer"
```

---

## 📁 Directory Structure for Features

```
/home/kumy/GIT/geokrety-points-system/
├── features/                          # NEW: Feature documentation
│   ├── country-leaderboard.md         # Feature: Country rankings
│   ├── breakdown-charts.md            # Feature: Statistics visualizations
│   ├── websocket-user-count.md        # Feature: Connected users display
│   ├── [future-feature].md
│   └── README.md                      # Guide to features/ directory
├── leaderboard-api/
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── handlers/
│   │   ├── models/
│   │   └── websocket/
│   └── ...
├── leaderboard-dashboard/
│   ├── src/
│   │   ├── views/
│   │   ├── components/
│   │   ├── composables/
│   │   └── router/
│   └── ...
├── geokrety-stats/
│   ├── internal/
│   ├── cmd/
│   └── Makefile
├── AGENT.md                           # Updated with development workflow
├── DEVELOPMENT_WORKFLOW.md            # THIS FILE
└── ...
```

---

## 🔍 Testing Checklist (Per Feature)

Before committing, verify:

- [ ] **Code Quality**
  - [ ] Code follows project style/patterns
  - [ ] Error handling implemented
  - [ ] No debug logs left in code
  - [ ] Comments for non-obvious logic

- [ ] **API Testing**
  - [ ] All endpoints respond with correct status codes
  - [ ] Response format matches documentation
  - [ ] Query parameters work correctly
  - [ ] Error cases handled gracefully
  - [ ] Examples in curl work exactly as documented

- [ ] **Binary Testing** (if applicable)
  - [ ] Builds with `make build`
  - [ ] Help output is clear
  - [ ] All flags documented
  - [ ] Runs without errors

- [ ] **UI Testing**
  - [ ] Screenshots taken with MCP Playwright
  - [ ] Layout is responsive (test multiple widths)
  - [ ] No JavaScript errors
  - [ ] Data displays correctly
  - [ ] Styling matches design system

- [ ] **Documentation**
  - [ ] Feature file exists in `features/`
  - [ ] All endpoints documented
  - [ ] All components documented
  - [ ] Testing procedures included
  - [ ] Examples are copy-paste ready

- [ ] **Deployment**
  - [ ] Docker images build successfully
  - [ ] `docker compose up -d` starts without errors
  - [ ] All services running: `docker compose ps`
  - [ ] No errors in logs: `docker compose logs`

---

## 🛠️ Common Commands Reference

### Build & Deploy
```bash
# Full rebuild and restart
cd /home/kumy/GIT/geokrety-points-system
docker compose down
docker compose build leaderboard-api leaderboard-dashboard
docker compose up -d

# Rebuild single service
docker compose build leaderboard-dashboard
docker compose up -d leaderboard-dashboard

# Quick restart (no rebuild)
docker compose restart
```

### Testing
```bash
# API test
curl -s http://<hostip>:8080/api/endpoint | jq .

# UI screenshot with MCP Playwright
# Load tools: tool_search_tool_regex with pattern ^mcp_microsoft_pla_browser
# Navigate: mcp_microsoft_pla_browser_navigate to http://<hostip>:3000/route
# Resize: mcp_microsoft_pla_browser_resize to 1280x1024
# Screenshot: mcp_microsoft_pla_browser_take_screenshot

# Binary build and run
cd geokrety-stats && make build && ./bin/geokrety-stats --help

# Check service status
docker compose ps
docker compose logs -f [service-name]
```

### Git
```bash
# Commit feature
git add .
git commit -m "feat: describe your feature"

# View feature documentation
cat features/[feature-name].md

# View development guide
cat DEVELOPMENT_WORKFLOW.md
```

---

## 📚 AI Context for Feature Development

When working on new features, the AI will:

1. **Read feature documentation** from `features/` directory
   - This provides context about the API spec
   - Clarifies component structure
   - Shows expected behavior

2. **Follow this workflow** (from AGENT.md + this file)
   - Implements feature according to spec
   - Tests with curl (not npx playwright test)
   - Uses MCP Playwright browser tools for screenshots
   - Tests binaries with make build
   - Rebuilds via docker compose

3. **Create/update feature documentation**
   - Before or immediately after implementation
   - Ensures specifications are preserved
   - Helps with testing and QA

4. **Commit per feature**
   - Clean git history
   - Logical grouping of changes
   - Conventional commit format

---

## ✅ Workflow Summary

```
┌─────────────────────────────────────────────────────────┐
│ 1. Create feature spec in features/[name].md            │
├─────────────────────────────────────────────────────────┤
│ 2. Implement frontend/backend                           │
├─────────────────────────────────────────────────────────┤
│ 3. Update feature spec with final details               │
├─────────────────────────────────────────────────────────┤
│ 4. Test with:                                           │
│    • curl for API endpoints                             │
│    • MCP Playwright browser tools for UI screenshots    │
│    • make build for binaries                            │
│    • docker compose for integration                     │
├─────────────────────────────────────────────────────────┤
│ 5. Git commit with conventional message                 │
├─────────────────────────────────────────────────────────┤
│ 6. Deploy with docker compose build/up                  │
└─────────────────────────────────────────────────────────┘
```

---

## 📞 Quick Reference

**Need to test API?**
```bash
curl -s http://<hostip>:8080/api/endpoint | jq .
```

**Need UI screenshot?**

Use MCP Playwright browser tools:
1. Load tools: `tool_search_tool_regex` with pattern `^mcp_microsoft_pla_browser`
2. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://<hostip>:3000/path`
3. Resize: `mcp_microsoft_pla_browser_resize` to 1280x1024
4. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

**Need to rebuild?**
```bash
cd /home/kumy/GIT/geokrety-points-system && docker compose down && \
docker compose build && docker compose up -d
```

**Need to check status?**
```bash
docker compose ps && docker compose logs -f
```

---

**See also:**
- [AGENT.md](AGENT.md) - AI development rules
- [gamification-rules.md](.github/instructions/gamification-rules.md) - Points system rules
- [LEADERBOARD_FEATURES.md](LEADERBOARD_FEATURES.md) - Specific leaderboard feature docs
