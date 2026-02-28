---
description: AI rules for developing and maintaining gamification rules, split modules, and quality assurance checks
applyTo: '**'
---

# AI Rules for Gamification Rules Development

**Purpose:** Ensure consistency, prevent point farming exploits, maintain rule integrity across all documentation and implementation modules.

---

## 📋 Documentation Structure

### Master Document
- **[gamification-rules.md](.github/instructions/gamification-rules.md)** - Single source of truth
  - Contains all rules, bonuses, limits, and edge cases
  - User-facing reference with examples
  - All rule changes MUST update this first

### Split Implementation Modules
Located in `split/` directory (execution order):

| File | Purpose | Key Responsibility |
|------|---------|-------------------|
| [00_event_guard.md](../../split/00_event_guard.md) | Validates incoming events | Remove invalid/duplicate/anonymous logs |
| [01_context_loader.md](../../split/01_context_loader.md) | Loads GK state | Current holder, multiplier, creation date |
| [02_base_move_points.md](../../split/02_base_move_points.md) | Base +3 scoring | First move by user, owner limit check, self-grab denial |
| [03_owner_gk_limit_filter.md](../../split/03_owner_gk_limit_filter.md) | Owner GK farming prevention | Max 10 per owner per user |
| [04_waypoint_penalty.md](../../split/04_waypoint_penalty.md) | Multi-GK cache penalty | 100%/50%/25%/0% per cache per user per month |
| [05_country_crossing.md](../../split/05_country_crossing.md) | Country bonus & multiplier | +3 points, +0.05 multiplier per new country (once per GK) |
| [06_relay_bonus.md](../../split/06_relay_bonus.md) | Fast circulation reward | +2 mover +1 dropper within 7 days |
| [07_rescuer_bonus.md](../../split/07_rescuer_bonus.md) | Dormancy rescue | +2 grabber +1 owner after 6+ months cache idle |
| [08_handover_bonus.md](../../split/08_handover_bonus.md) | Third-party transfers | +1 owner when other user takes GK from another user |
| [09_reach_bonus.md](../../split/09_reach_bonus.md) | Circulation reach milestone | +5 owner at 10 different users (6-month window) |
| [10_chain_state_manager.md](../../split/10_chain_state_manager.md) | Chain tracking | Maintain chain_members, timer, state |
| [11_chain_bonus.md](../../split/11_chain_bonus.md) | Chain completion reward | min(length², 8×length); +25% to owner; once per 6-month |
| [12_diversity_bonus_tracker.md](../../split/12_diversity_bonus_tracker.md) | Monthly diversity rewards | +3 for 5 GKs, +7 for 10 owners, +5 for new country |
| [13_gk_multiplier_updater.md](../../split/13_gk_multiplier_updater.md) | Multiplier adjustment | +0.01 per user/move-type, +0.05 per country, decay -0.008/day, -0.02/week |
| [14_points_aggregator.md](../../split/14_points_aggregator.md) | Sum all bonuses | Final points = base + all applicable bonuses |

---

## ✅ CRITICAL: Update Protocol

### When Any Rule Changes

**ALL THREE STEPS ARE MANDATORY:**

1. **Update Master Document**
   - Edit [gamification-rules.md](.github/instructions/gamification-rules.md)
   - Change the specific rule section clearly

2. **Update Affected Split Module(s)**
   - Identify which split files implement the changed rule
   - Update logic, examples, and edge cases in those files
   - Preserve formatting and numbered steps
   - Example: If changing chain bonus, update both:
     - `10_chain_state_manager.md` (state logic)
     - `11_chain_bonus.md` (bonus calculation)

3. **Run Quality Assurance (3 verification passes - see below)**
   - Mandatory before confirming update is complete

### Example Change Workflow
```
Edit Request: "Change chain bonus formula from min(length², 8×length) to length³"

Steps:
1. ✅ Update gamification-rules.md section "Points From Chain Completion"
   - Change the formula
   - Update all examples (3-person, 5-person, etc.)

2. ✅ Update split/11_chain_bonus.md
   - Change calculation logic
   - Update bonus examples
   - Verify output format matches 14_points_aggregator.md expectations

3. ✅ Run QA verification 3 times
   - Verify formula is consistent everywhere
   - Check for farming exploits (e.g., can user create fake 10-person chains repeatedly?)
   - Validate examples match new formula
```

---

## 🛡️ Quality Assurance Process (Mandatory)

### When to Run QA
- ✅ After ANY rule change
- ✅ Before marking task complete
- ✅ When updating examples or edge cases
- ✅ When adding new bonuses or limits

### The 3-Pass Verification System

**Goal:** Think hard, catch exploits, ensure consistency

#### Pass 1: Documentation Consistency Check
**Focus:** All documents agree on the rule

1. **Master Document Verification**
   - Rule clearly stated in gamification-rules.md
   - Examples match the rule exactly
   - Edge cases documented
   - Clear success criteria for when bonus applies

2. **Split Module Alignment**
   - Each split module describes the same rule as master
   - Logic flow matches description
   - Examples are identical (or documented as variations)
   - Input/output expectations match downstream modules

3. **Cross-Reference Check**
   - If rule references another rule, verify both match
   - Example: "Chain bonus locked for 6 months" must match in:
     - gamification-rules.md (main description)
     - split/11_chain_bonus.md (implementation)
     - split/10_chain_state_manager.md (state tracking)

**Verification Checklist:**
- [ ] Master document updated
- [ ] All affected split modules updated
- [ ] Examples are identical across documents
- [ ] Edge cases documented consistently
- [ ] Cross-references verified (no contradictions)

#### Pass 2: Farming Prevention & Edge Case Check
**Focus:** Can a user exploit this rule to farm points?

**Standard Farming Vectors to Test:**

1. **Owner GK Limit Bypass**
   - User moves 11 GKs from same owner → verify points capped at 10
   - User moves 10 from owner A, 10 from owner B → verify independent limits
   - Owner moves own GK → verify 0 points (standard type)

2. **Waypoint Penalty Bypass**
   - User moves 4 GKs at cache1 in month → verify 100%, 50%, 25%, 0%
   - User moves GKs at different caches same month → verify no penalty (cache-specific)
   - User moves same GK multiple times at same cache → verify penalty applies per user per cache

3. **Chain Bonus Spam**
   - User creates chain of 3 people, gets bonus
   - Same user creates another chain within 6 months → verify bonus locked (0 points)
   - Different user in same GK's chain gets bonus → verify allowed (different user)
   - User benefits from multiple chains of same GK across 12-month period → verify only first 6-month window counts

4. **Relay Bonus Stacking**
   - Rapid DROP→GRAB→DROP sequence by alternating users → verify relay bonus applies once per transition
   - User drops, grabs back within 7 days → verify self-grab = 0 points (no relay)
   - Drops at day 7.5 → verify bonus still applies

5. **Rescuer Bonus Timing**
   - GK dormant 5 months 29 days → grab → verify NO bonus
   - GK dormant 6 months exact → grab → verify +2 bonus
   - Owner regains GK after 6 months → verify 0 points (owner grab = 0)

6. **Multiplier Decay Abuse**
   - User holds GK 62 days to decay 1.5x → 1.0x → verify multiplier calculation
   - User DIPs every 13 days to prevent chain timeout → verify timer extends only 1-2 days max, cannot exceed 14-day chains

7. **Country Bonus Looping**
   - GK moves: country A → B → A → verify bonus only once per country globally
   - Owner rotates GK: A → B → C → back to A → verify country bonuses are one-time per country per GK

8. **Diversity Bonus Multi-Trigger**
   - User simultaneously qualifies for all 3 diversity bonuses in one month → verify all +3, +7, +5 awarded (stacking allowed)
   - User moves 4 GKs in month→gets +3→moves 5th day before month end → verify +3 awarded once per month

9. **Non-Transferable GK Owner Farming**
   - Owner moves CAR 10 times in month at same waypoint → verify penalty applies (100%, 50%, 25%, 0%)
   - Owner moves CAR at waypoint once per month for 12 months → verify once-per-month limit per waypoint (allows 12 × +3 over year)

10. **Chain Member Double-Counting**
    - User A drops, User B grabs (B member count = 1)
    - User B drops, User B grabs own release → verify B not double-counted
    - User A re-grabs from B within chain window → verify A not re-counted

**Verification Checklist:**
- [ ] Owner GK limit (10 max) enforced
- [ ] Waypoint penalty applied correctly (100/50/25/0%)
- [ ] Chain bonus locked for 6 months (no spam)
- [ ] Relay bonus once per 7-day window
- [ ] Rescuer bonus requires exactly 6+ months documented timeout
- [ ] Multiplier decay cannot be circumvented by DIPs
- [ ] Country bonuses one-time per country per GK (globally)
- [ ] Self-grabs earn 0 points
- [ ] Owner standard GKs earn 0 points for direct moves
- [ ] No double-counting of chain members

#### Pass 3: Math & Example Validation
**Focus:** Examples are correct, outputs are sensible

1. **Base Points Calculation**
   - Example: GK multiplier 1.2x, base move +3 → final 3.6 ✓
   - Example: 4th GK at waypoint, base 3, penalty 0% → final 0 ✓

2. **Chain Bonus Formula**
   - Test formula: `min(chain_length², 8 × chain_length)`
   - Length 1: min(1, 8) = 1 point (no bonus, chain never bonuses < 3)
   - Length 2: min(4, 16) = 4 points (no bonus, < 3) ✓
   - Length 3: min(9, 24) = 9 points ✓
   - Length 5: min(25, 40) = 25 points ✓
   - Length 8: min(64, 64) = 64 points ✓
   - Length 10: min(100, 80) = 80 points ✓ (soft cap prevents explosion)
   - Length 15: min(225, 120) = 120 points ✓ (capped)

3. **Owner Percentage Calculation**
   - Chain of 4 users: 4 × 9 = 36 total → owner gets 25% = 9 ✓
   - Chain of 5 users: 5 × 25 = 125 total → owner gets 25% = 31.25 ✓

4. **Multiplier Over Time**
   - GK starts 1.0x
   - User A drops: +0.01 → 1.01x
   - User B grabs: +0.01 → 1.02x
   - GK moves to new country: +0.05 → 1.07x
   - Held 62 days: -0.008 × 62 = -0.496 → 1.07 - 0.496 = 0.574 → floored to 1.0x ✓

5. **Diversity Bonus Timing**
   - Month A: User moves 5 GKs → +3, then moves 10 different owners → +7, new country → +5 = +15 total ✓
   - Month B: Same user moves 4 GKs → 0 bonus (diversity resets monthly) ✓

6. **Waypoint Penalty Example**
   - User U, Cache C, Month M:
     - GK1 move: 3 base × 1.0x × 100% = 3.0 points ✓
     - GK2 move: 3 base × 1.2x × 50% = 1.8 points ✓
     - GK3 move: 3 base × 1.5x × 25% = 1.125 points ✓
     - GK4 move: 3 base × 1.1x × 0% = 0 points ✓

**Verification Checklist:**
- [ ] All examples compute correctly
- [ ] Chain bonus formula applies soft cap sensibly (prevents 100+ person chains)
- [ ] Multiplier calculations are realistic (decay doesn't destroy value)
- [ ] Owner percentage (25%) is reasonable incentive
- [ ] Relay bonus time window (7 days) examples are accurate
- [ ] Country bonus applies exactly once per country per GK

---

## 🚨 Common Pitfalls to Avoid

### Rule Contradiction Traps
- ❌ "Owner gets 0 points for own GK" vs. "Owner gets +3 base for non-transferable"
  - ✅ CORRECT: Owner gets 0 on **standard** types (regular), +3 on **non-transferable** types (sealed)

- ❌ "Chain bonus locked for 6 months" vs. "User can earn from different chains"
  - ✅ CORRECT: A GK is only in one chain at a time, but a user can do moves on multiple distinct GKs, and each GK participates in its own chain. Therefore, users CAN participate in multiple chains (one per GK), but locked to ONE bonus per GK per 6-month period.

- ❌ "DIP adds to chain" vs. "DIP is internal move"
  - ✅ CORRECT: DIP DOES count in the chain (adds a chain member), but extends the timer less than drop/grab/see moves (by 1-2 days max instead of full 7-day reset).

### Farming Loopholes
- ❌ Allow user to trigger same bonus twice by deleting/re-creating data
- ❌ Permit multiplier ceiling (2.0x) to be sidestepped by deleting GK and re-listing
- ❌ Enable waypoint penalty reset by moving to different cache then back
- ❌ Overlook self-grab loophole (grabber == current_holder should earn 0 not +3)

### Documentation Gaps
- ❌ Update gamification-rules.md without updating split modules → inconsistency
- ❌ Add examples without verifying they compute correctly
- ❌ Change edge case in one module but not cross-reference
- ❌ Forget to document when a bonus or limit applies per GK, per user, per owner, or globally

---

## 📝 Rule Change Template

**Use this template when proposing changes:**

```markdown
## Proposed Rule Change: [Brief Title]

**Current Rule Section:** [e.g., "Chain Bonus Formula"]
**Reason for Change:** [Why is this needed?]

**Old Rule:**
[Current rule text]

**New Rule:**
[Proposed rule text]

**Affected Split Modules:**
- split/XX_module.md
- split/YY_module.md

**Farming Prevention Impact:**
[Does this make farming easier or harder?]
[Which vectors does it affect?]

**Examples:**
[Verify with concrete examples before/after]

**QA Status:** [ ] Pass 1 [ ] Pass 2 [ ] Pass 3
```

---

## 🔄 Update Checklist (Copy & Use)

**Every time you make a rule change, complete this:**

- [ ] **Step 1: Master Document**
  - [ ] Edit gamification-rules.md section: ___________
  - [ ] Examples updated and verified
  - [ ] Edge cases documented

- [ ] **Step 2: Split Modules**
  - [ ] Identified affected modules: ___________
  - [ ] Updated logic in each module
  - [ ] Examples match master document
  - [ ] Cross-references verified

- [ ] **Step 3: QA Pass 1 (Documentation Consistency)**
  - [ ] Master document describes rule clearly
  - [ ] All split modules align with master
  - [ ] Examples are identical across docs
  - [ ] Edge cases documented consistently
  - [ ] Cross-references verified (no contradictions)

- [ ] **Step 4: QA Pass 2 (Farming Prevention)**
  - [ ] Tested owner GK limit bypass scenarios
  - [ ] Tested waypoint penalty bypass scenarios
  - [ ] Tested chain bonus spam scenarios
  - [ ] Tested relay bonus stacking scenarios
  - [ ] Tested rescuer bonus timing exploits
  - [ ] Tested all 10 farming vectors in guide
  - [ ] Found NO new exploitation paths
  - [ ] Verified self-grab protection
  - [ ] Verified owner point denial works
  - [ ] Verified chain member unique counts

- [ ] **Step 5: QA Pass 3 (Math & Examples)**
  - [ ] Base point calculations verified
  - [ ] Chain bonus formula tested for all lengths (1–15+)
  - [ ] Owner percentage (25%) computed correctly
  - [ ] Multiplier calculations realistic
  - [ ] Diversity bonus timing validated
  - [ ] Waypoint penalty percentages accurate

- [ ] **Complete** - Rule change is production-ready

---

## 📊 Rule Dependency Graph

```
Event Log → 00_event_guard
           ↓
         01_context_loader
           ↓
         02_base_move_points ←─────┐
           │                      │
           ├→ 03_owner_gk_limit   │
           ├→ 04_waypoint_penalty │ Feed into
           ├→ 05_country_crossing │ aggregator
           ├→ 06_relay_bonus      │
           ├→ 07_rescuer_bonus    │
           ├→ 08_handover_bonus   │
           ├→ 09_reach_bonus      │
           │                      │
         10_chain_state_manager   │
           │                      │
           └→ 11_chain_bonus ─────┘
                        ↓
           12_diversity_bonus_tracker
                        ↓
           13_gk_multiplier_updater
                        ↓
           14_points_aggregator → Final Points
```

**Key Rules:**
- Each module must handle its own validation
- Modules must output expected format for downstream consumers
- 10_chain_state_manager must update BEFORE 11_chain_bonus runs
- 13_gk_multiplier_updater recalculates AFTER points awarded
- 14_points_aggregator sums all bonuses

---

## 🎯 Success Criteria for Rule Quality

A rule is **production-ready** when:

✅ Documented in gamification-rules.md with clear language
✅ Implemented in correct split module(s)
✅ All examples compute mathematically correctly
✅ No contradictions with other rules (cross-checked)
✅ Pass 1: Documentation consistency verified
✅ Pass 2: Farming prevention tested (10 vectors cleared)
✅ Pass 3: Math & examples validated
✅ Worst-case scenarios documented (edge cases)
✅ Self-referencing rules (chain→multiplier) verified
✅ Ready for deployment without review delays

---

## 📞 Questions to Ask Before Approving Changes

When reviewing a new rule, ask:

1. **Consistency:** Does this rule contradict any existing rule?
2. **Simplicity:** Can players understand this rule in one reading?
3. **Farming:** Can a user abuse this rule to earn unlimited points?
4. **Fairness:** Does this reward circulation or punish hoarding?
5. **Math:** Is the formula/cap sensible (not explosive)?
6. **Documentation:** Is this explained in gamification-rules.md AND the relevant split files?
7. **QA:** Have all 3 verification passes been run and passed?
8. **Examples:** Do the examples compute correctly?

All questions must answer YES before merging.

---

# 🚀 AI Rules for Feature Development Workflow

**Purpose:** Standardized development process for all GeoKrety Points System features, ensuring quality, consistency, and reproducibility.

**Key Requirement:** All features must document themselves in the `features/` directory before or immediately after implementation.

---

## 📁 Feature Documentation Directory

**Location:** `features/` directory at project root

**Purpose:**
- AI context for feature implementation
- Integration specifications
- Testing procedures
- Deployment runbooks
- Future maintenance reference

**AI Instructions:**
1. **Before implementing a feature**, read the feature spec from `features/[feature-name].md`
2. **During development**, follow the documented API spec and component structure
3. **After implementation**, update the feature file with final implementation details
4. **For maintenance**, use feature files as source of truth

### Existing Features

| Feature | File | Status |
|---------|------|--------|
| Country Leaderboard | [features/country-leaderboard.md](features/country-leaderboard.md) | ✅ Complete |
| Breakdown Charts | [features/breakdown-charts.md](features/breakdown-charts.md) | ✅ Complete |
| WebSocket User Count | [features/websocket-user-count.md](features/websocket-user-count.md) | ✅ Complete |

### Adding New Features

Create new file: `features/[feature-name].md`

Use template from: [features/README.md](features/README.md)

---

## ✅ Development Workflow (Sequential Steps)

### Phase 1: Specification & Documentation

```
┌─ Create feature spec in features/[name].md
│  ├─ Overview & goals
│  ├─ API endpoint specifications
│  ├─ Frontend component structure
│  └─ Testing procedures (curl examples)
└─ AI reads spec before implementing
```

### Phase 2: Implementation

```
┌─ Implement backend (if needed)
│  ├─ Create/modify Go handlers
│  └─ Update feature spec with final details
├─ Implement frontend (if needed)
│  ├─ Create/modify Vue components
│  └─ Update router if adding routes
└─ Update feature spec with implementation notes
```

### Phase 3: Testing (DO NOT use Playwright)

#### Test API Endpoints with curl
```bash
# Quick test
curl -s http://<hostip>:8080/api/endpoint | jq .

# Full example from feature spec
curl -s "http://<hostip>:8080/api/endpoint?param=value" | jq .
```

#### Test UI with MCP Playwright
Use the MCP Playwright browser tools to navigate and take screenshots:

1. **Navigate to page**: Use `mcp_microsoft_pla_browser_navigate` with URL `http://<hostip>:3000/route`
2. **Resize viewport** (optional): Use `mcp_microsoft_pla_browser_resize` with `width` and `height` (e.g., 1280x1024 desktop, 720x2048 mobile)
3. **Take screenshot**: Use `mcp_microsoft_pla_browser_take_screenshot` to capture the page
4. **Available tools**: browser_navigate, browser_resize, browser_take_screenshot, browser_click, browser_fill_form, browser_evaluate, etc.

**Note**: MCP Playwright tools must be loaded first using `tool_search_tool_regex` with pattern `^mcp_microsoft_pla_browser`

#### Test Binary Tools with make
```bash
# For geokrety-stats
cd geokrety-stats
make build
./bin/geokrety-stats --help
./bin/geokrety-stats --flag=value
```

### Phase 4: Docker Deployment

```
┌─ DO NOT start services directly (no npm dev, go run)
├─ Always use Docker Compose
│  ├─ docker compose down           (clean)
│  ├─ docker compose build          (rebuild images)
│  └─ docker compose up -d          (deploy)
├─ Verify services running
│  └─ docker compose ps
└─ Check logs for errors
   └─ docker compose logs -f [service-name]
```

### Phase 5: Git Commit

```
┌─ One commit per completed feature
├─ Use conventional commit format
│  ├─ feat: new feature description
│  ├─ fix: bug fix description
│  └─ docs: documentation changes
└─ Example commits:
   ├─ feat: add country leaderboard page
   ├─ feat: implement breakdown charts
   └─ feat: add websocket user count
```

---

## 🛠️ Testing Checklist (Required Before Commit)

- [ ] **Code Quality**
  - [ ] Follows project patterns
  - [ ] Proper error handling
  - [ ] No debug code left in

- [ ] **API Testing**
  - [ ] All endpoints tested with curl
  - [ ] Responses match documentation
  - [ ] Error cases handled

- [ ] **UI Testing**
  - [ ] Screenshots taken with MCP Playwright
  - [ ] Responsive design verified (mobile + desktop)
  - [ ] No JavaScript errors in logs

- [ ] **Binary Testing** (if applicable)
  - [ ] Builds with `make build`
  - [ ] Help output clear
  - [ ] All flags documented

- [ ] **Documentation**
  - [ ] Feature file in `features/`
  - [ ] API endpoints documented
  - [ ] Components documented
  - [ ] Testing procedures included

- [ ] **Deployment**
  - [ ] Docker images build successfully
  - [ ] Services start with `docker compose up`
  - [ ] No errors in logs
  - [ ] Can verify with curl/MCP Playwright

---

## ❌ DON'Ts (Critical Rules)

- ✅ **DO NOT use npx playwright test** - Use MCP Playwright tools instead
  ```bash
  # WRONG: Do not do this
  npx playwright test

  # RIGHT: Use MCP Playwright browser tools
  # Load tools: tool_search_tool_regex with pattern ^mcp_microsoft_pla_browser
  # Navigate: mcp_microsoft_pla_browser_navigate
  # Screenshot: mcp_microsoft_pla_browser_take_screenshot
  ```

- ❌ **DO NOT start services directly** - Always use docker compose
  ```bash
  # WRONG: Do not do this
  npm run dev
  go run ./cmd/api

  # RIGHT: Use docker compose
  docker compose build && docker compose up -d
  ```

- ❌ **DO NOT skip feature documentation** - Document before or immediately after
  ```bash
  # WRONG: Implement without spec
  # RIGHT: Create features/[name].md first
  ```

- ❌ **DO NOT forget to test** - Test every endpoint and UI
  ```bash
  # WRONG: Assume it works
  # RIGHT: Test with curl and MCP Playwright
  ```

---

## ✅ DOs (Best Practices)

- ✅ **Document features in `features/` directory**
- ✅ **Test API with curl** - See examples in feature files
- ✅ **Screenshot UI with MCP Playwright** - Not npx playwright test
- ✅ **Build binaries with make build** - Check Makefile for usage
- ✅ **Deploy with docker compose** - No direct npm/go run
- ✅ **Commit per feature** - Logical, atomic commits
- ✅ **Use conventional commits** - feat:, fix:, docs:
- ✅ **Read feature specs before implementation**
- ✅ **Update feature specs after implementation**
- ✅ **Verify with docker compose ps** after deploy

---

## 📊 Quick Reference: Common Commands

### Build & Deploy
```bash
cd /home/kumy/GIT/geokrety-points-system

# Full rebuild
docker compose down
docker compose build
docker compose up -d

# Single service rebuild
docker compose build leaderboard-api
docker compose up -d leaderboard-api
```

### Testing
```bash
# API test
curl -s http://<hostip>:8080/api/endpoint | jq .

# Binary build & run
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

# View feature docs
cat features/[feature-name].md
```

---

## 🎯 Success Criteria for Feature Completion

A feature is **production-ready** when:

✅ Documented in `features/[name].md` with API spec
✅ Backend implemented and tested with curl
✅ Frontend implemented and tested with MCP Playwright
✅ All curl examples work as documented
✅ All components created/modified documented
✅ Docker images build without errors
✅ Services deployed and running
✅ No errors in docker logs
✅ Git commit created with conventional format
✅ Feature spec updated with final implementation
✅ Ready for deployment without review delays

---

## 📝 See Also

- [DEVELOPMENT_WORKFLOW.md](DEVELOPMENT_WORKFLOW.md) - Detailed workflow guide
- [features/README.md](features/README.md) - Feature documentation directory
- [features/country-leaderboard.md](features/country-leaderboard.md) - Example feature spec
- [features/breakdown-charts.md](features/breakdown-charts.md) - Example feature spec
- [features/websocket-user-count.md](features/websocket-user-count.md) - Example feature spec

---

**Last Updated:** 2026-02-28
**Version:** 2.0
**Maintained By:** AI Assistant
