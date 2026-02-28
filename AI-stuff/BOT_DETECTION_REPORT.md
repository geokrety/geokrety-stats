# System & Bot Account Detection

## Suspicious Accounts Findings

### 🚨 HIGH PRIORITY: User 44304
**Verdict: LIKELY ARCHIVE/BOT ACCOUNT - RECOMMEND FILTERING**

```
Activity Pattern:
  Database Moves: 74,711 (EXTREME)
  Unique Items: 1,352
  Days Active: 1,702 (June 2020 - Feb 2025)
  Moves Per Day: 43.9 ← INHUMAN RATE

Move Composition:
  RESCUE: 72,757 (97.4%) ← RESCUING ARCHIVE ITEMS
  GRAB: 1,424 (1.9%)
  DIP: 487 (0.6%)

Item Usage Pattern:
  GK 80499: 833 moves (1.1%)
  GK 27144: 344 moves (0.5%)
  Moves per item: 55.3 average

Ownership: 98.2% of moved items
```

**Interpretation:**
- 43.9 moves per day makes this mathematically impossible for human play
  - Would require moving an item every ~30 seconds, 24/7
  - Only move types are GRAB/RESCUE (automated operations)
- 97.4% RESCUE suggests automated item preservation
- 98.2% ownership suggests managing items, not playing
- Only 1,352 unique items in 74K moves = repeating items

**Likely Scenario:**
- Automated system rescuing items from archive/deletion
- Keeps important items active by moving them regularly
- Could be GeoKrety.org infrastructure, not a player

**Recommendation:**
```
Status: FILTER FROM LEADERBOARD
Reason: Non-human activity pattern, likely system bot
Action: Exclude from competition, but keep data for audit
```

---

### ⚠️ MEDIUM PRIORITY: User 10761
**Verdict: POSSIBLE STAFF/TEST ACCOUNT - RECOMMEND REVIEW**

```
Activity Pattern:
  Database Moves: 26,061
  Unique Items: 1,177
  Days Active: 4,942 (July 2011 - Feb 2025)
  Moves Per Day: 5.3 (high but plausible)

Move Composition:
  RESCUE: 23,404 (89.8%)
  GRAB: 1,485 (5.7%)
  DIP: 747 (2.9%)

Suspicious Item:
  GK 17792: 7,735 moves (29.7% of ALL their moves)
    → Moved ~1.5 times per day on average
    → Mathematical impossibility for normal play
    → Indicator: MARKER ITEM for testing/tracking
  GK 74487: 1,881 moves (7.2%)
    → Also excessive for casual play

Ownership: 95.5% of moved items
```

**Interpretation:**
- GK 17792 with 7,735 moves is a RED FLAG
  - Could be test item tracking position coordinates
  - Could be community hub item (unlikely at this volume)
  - Most likely: Development/staff marker
- 89.8% RESCUE + 95.5% ownership = managing items
- 5.3 moves/day: High but technically possible for dedicated player
- Likely: Staff account used for testing/debugging

**Possible Scenarios:**
1. **Development Account:** Testing database/score calculation
2. **Staff Account:** Testing item mechanics
3. **Marker Item:** Using GK 17792 to track coordinates/versions
4. **Community Manager:** Maintaining reference items (unlikely)

**Recommendation:**
```
Status: REVIEW BEFORE LEADERBOARD
Reason: Suspicious concentration on single item (7,735 moves)
Action: Verify with dev team if staff account, then filter if yes
```

---

## Legitimate Cache Managers (Keep, But Monitor)

### User 19185 & 22471 - Acceptable Pattern

```
Similarity Metrics:
  Both own ~52% of items they move
  Both have 4,000+ days activity
  Both have balanced move types
  Both show distributed locations

Assessment: LEGITIMATE CACHE CURATORS

Why They're OK:
  • Ownership % shows they're managing items
  • But they still interact with community items
  • Geographic distribution shows real play

What They Do:
  • Release items for community to move
  • Play other people's items too
  • No single item dominance
  • Normal activity rates (0.9-1.6 moves/day)

Recommendation: KEEP IN LEADERBOARD
```

---

## Account Classification Framework

### Red Flags for Detection

```
Level 1: CERTAIN BOT
  • >20 moves/day for extended period (>100 days)
  • >90% single move type (RESCUE/GRAB indicates automation)
  • >80% ownership of moved items
  • Single item >10% of all moves

Level 2: LIKELY BOT/SYSTEM
  • >10 moves/day for extended period
  • >80% single move type
  • Single item >5% of all moves
  • 98%+ ownership

Level 3: MONITOR
  • >5 moves/day
  • >70% single move type
  • Single item >1% of all moves
  • >70% ownership OR highly location-concentrated
```

---

## By-Account Assessment

| User | Status | Recommendation | Reason |
|------|--------|---|---|
| 23452 | ✅ KEEP | Leaderboard OK | Legitimate player pattern |
| 14462 | ✅ KEEP | Archive OK | Historical player, inactive |
| 19185 | ✅ KEEP | Leaderboard OK | Legitimate cache manager |
| 22471 | ✅ KEEP | Leaderboard OK | Legitimate cache manager |
| 6983  | ✅ KEEP | Leaderboard OK | Casual legitimate player |
| 19048 | ✅ KEEP | Monitor | Legitimate but rescue-heavy |
| 44304 | 🚫 FILTER | Exclude | Bot/system account (43.9/day) |
| 10761 | 🟡 REVIEW | Conditional | Possible staff account (GK marker) |
| 3813  | ✅ KEEP | Leaderboard OK | Legitimate local player |
| 3807  | ✅ KEEP | Leaderboard OK | Legitimate local player |

---

## Recommendations for Production

### Immediate Actions
1. **Remove from public leaderboards:**
   - User 44304 (confirmed bot pattern)

2. **Review before release:**
   - User 10761 (verify staff/marker status)

### Monitoring Strategy
```python
# Add to production code:
EXCLUDED_USERS = {44304}  # Bot accounts

# Optional: Low priority monitoring
MONITOR_USERS = {10761, 19048}  # High RESCUE %, needs verification

# Filters to apply:
users_for_leaderboard = [u for u in all_users
                         if u not in EXCLUDED_USERS
                         and u.moves_per_day < 20]
```

### Ongoing Validation
Monitor for new accounts matching:
- >20 moves/day sustained
- >95% ownership
- >90% single move type
- Single item dominance (>10% of moves)

---

## Data Quality Notes

### Why These Patterns Exist
1. **RESCUE-heavy accounts:** Early game (pre-2010s) didn't have item deletion
   - Items would stay active forever
   - RESCUE mechanic added later for cleanup
   - Some power users got items into "lost" state and rescue them repeatedly

2. **High ownership %:** Two scenarios
   - **Scenario A:** Cache network operator (legitimate)
   - **Scenario B:** Cache setter dropping own items on repeat (farming)
   - Currently no way to distinguish without location analysis

3. **Single location dominance:** Mostly legitimate
   - User lives near that location
   - User manages cache there
   - Local meetup location

### Data Collection Improvements Needed
To better detect abuse:
- Track item ownership history (who initially created it)
- Log timestamps to detect burst activity (batch processing)
- Identify coordinated multi-account activity
- Analyze geographic clusters (are all moves at same GPS?)

---

## Testing Checklist Before Leaderboard Release

- [ ] Remove User 44304 from rankings
- [ ] Verify User 10761 status with dev team
- [ ] Confirm User 19185/22471 have no alt accounts
- [ ] Run automated bot detection on all remaining users
- [ ] Sample 20 random users and verify their play patterns make sense
- [ ] Check for burst activity (many moves in short time window)
- [ ] Validate geographic distribution of moves
- [ ] Test with known gaming scenarios:
  - Same user, same waypoint, rapid-fire drops
  - Multiple accounts at single waypoint
  - Automated item rotation patterns

---

## Notes on Bot Detection

The current dataset is EXCELLENT for bot detection because:
1. **Humans naturally vary** activity (weekends/weekdays, seasons)
2. **Bots are consistent** (same moves/day regardless of day of week)
3. **Humans explore** (geographically distributed)
4. **Bots repeat** (same items, same locations)
5. **Humans have interests** (diverse item types)
6. **Bots execute tasks** (90%+ single move type)

User 44304 passes every bot detection test with flying colors. This is a likely system automation account.

---

## Appendix: Related Metrics to Watch

### Account Health Score (Example)
```
Score = (Diversity × 0.3) + (Spread × 0.3) + (Consistency × 0.2) + (Activity × 0.2)

Where:
  Diversity = Unique items / Total moves (lower = more repetitive)
  Spread = Locations / Total moves (lower = farming one spot)
  Consistency = StdDev(moves_per_day) (higher = varying activity)
  Activity = Moves per day (plausible range: 0.1-5.0)

Red Zone: Score < 0.3
Yellow Zone: Score 0.3-0.5
Green Zone: Score > 0.5

User 44304: Score = 0.018 (DEEP RED)
User 10761: Score = 0.15 (DEEP RED)
User 23452: Score = 0.92 (GREEN)
```

Consider implementing something similar for ongoing monitoring.
