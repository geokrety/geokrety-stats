---
description: Gamification rules for Geokrety - live points, multiplier system, and circulation incentives
applyTo: '**'
---

# Geokrety Gamification Rules

**Core Philosophy:** Reward circulation and real hand-to-hand movement. Encourage quick turnover and diverse interaction patterns.

---

---

## 📋 Reference: Move Types & Geokrety Types

### Move Types (LogType)

- `LOG_TYPE_DROPPED (0)`
  GK placed in cache.
  - Updates last position
  - Requires coordinates
  - Counts kilometers
  - User touched
  - Considered alive

- `LOG_TYPE_GRABBED (1)`
  GK taken from cache or another user.
  - Updates last position
  - Does NOT count kilometers
  - User touched
  - Considered alive

- `LOG_TYPE_COMMENT (2)`
  Comment only. No state change.
  - Does NOT update position
  - Does NOT count kilometers
  - Not considered alive
  - Editable

- `LOG_TYPE_SEEN (3)`
  GK observed in cache (no ownership change).
  - Updates last position
  - Coordinates optional (but supported)
  - Counts kilometers
  - User touched
  - Considered alive
  - Theoretically in cache

- `LOG_TYPE_ARCHIVED (4)`
  GK archived / retired from active circulation.
  - Updates last position
  - Does NOT count kilometers
  - Not considered alive
  - Not editable

- `LOG_TYPE_DIPPED (5)`
  GK moved within same holder (inventory/cache internal move).
  - Updates last position
  - Requires coordinates
  - Counts kilometers
  - User touched
  - Considered alive

---

### LogType Behavioral Groups (Derived Logic)

- **Alive Logs**
  `DROPPED, GRABBED, SEEN, DIPPED`

- **Require Coordinates**
  `DROPPED, SEEN, DIPPED`

- **Optional Coordinates**
  `SEEN`

- **Count Kilometers**
  `DROPPED, SEEN, DIPPED`

- **Theoretically In Cache**
  `DROPPED, SEEN`

- **Update Last Position**
  `DROPPED, GRABBED, SEEN, ARCHIVED, DIPPED`

- **User Touched GK**
  `DROPPED, GRABBED, SEEN, DIPPED`

- **Editable Logs**
  `DROPPED, GRABBED, COMMENT, SEEN, DIPPED`

---

### Geokrety Types

- `GEOKRETY_TYPE_TRADITIONAL (0)` – Standard transferable GK
- `GEOKRETY_TYPE_BOOK_CD_DVD (1)` – Media object
- `GEOKRETY_TYPE_HUMAN (2)` – Person traveling
- `GEOKRETY_TYPE_COIN (3)` – Coin object
- `GEOKRETY_TYPE_KRETYPOST (4)` – Postal-type GK
- `GEOKRETY_TYPE_PEBBLE (5)` – Painted pebble
- `GEOKRETY_TYPE_CAR (6)` – Non-transferable, vehicle-based
- `GEOKRETY_TYPE_PLAYING_CARD (7)` – Card collectible
- `GEOKRETY_TYPE_DOG_TAG (8)` – Non-transferable personal item
- `GEOKRETY_TYPE_JIGSAW (9)` – Puzzle piece
- `GEOKRETY_TYPE_EASTER_EGG (10)` – Admin-only special type

---

### Type Categories

- **Standard Transferable Types**
  `TRADITIONAL, BOOK_CD_DVD, HUMAN, COIN, KRETYPOST, PEBBLE, PLAYING_CARD, JIGSAW`

- **Non-Transferable Types**
  `CAR, DOG_TAG, EASTER_EGG`
  - Single owner
  - No ownership handover intended
  - Reward movement instead of transfers

- **Admin-Only Types**
  `EASTER_EGG`

- **Types Supporting "Missing" Status**
  `TRADITIONAL, BOOK_CD_DVD, COIN, KRETYPOST, PEBBLE, PLAYING_CARD, JIGSAW`

---

### Picture Types

- `PICTURE_GEOKRET_AVATAR (0)` – Main Geokret image
- `PICTURE_GEOKRET_MOVE (1)` – Image attached to a move/log
- `PICTURE_USER_AVATAR (2)` – User profile avatar

---

## 🎯 Core Design Principles

- ✅ Reward change of hands (handovers between users)
- ✅ Reward box drops
- ✅ Reward discovery and new interactions
- ✅ Discourage inventory hoarding
- ✅ Discourage ping-pong between same users
- ✅ Prevent self-farming (limits on own GKs)
- ✅ Dynamic point system (not lifetime ranking)


---

## ⚙️ Base System Rules

### When Points Are Awarded
- ✅ **Logged user moves only** - Anonymous moves = 0 points
- ✅ **Once per action** - Each move logged reward once, no duplicates
- ✅ **Points awarded immediately** - User gains points on log creation (before GK multiplier updates)
- ✅ **Chain-based rewards** - Chain bonuses awarded when chain ends (see Chain section)

### GK Multiplier System

Every Geokret starts with a **multiplier of 1.0x** that rises and falls based on its history.

#### Multiplier Increases

**On Move Logged (each user contributes once per move type):**

- First DROP logged by a user on this GK: +0.01 (per user, so multiple users each add +0.01)
- First GRAB logged by a user on this GK: +0.01 (per user, so multiple users each add +0.01)
- First SEEN logged by a user on this GK: +0.01 (per user, so multiple users each add +0.01)
- First DIP logged by a user on this GK: +0.01 (per user, so multiple users each add +0.01)
- *Note: COMMENTs do not increase multiplier*
- **Example:** When User A drops, multiplier +0.01. When User B grabs, multiplier +0.01. When User C drops, multiplier +0.01. = Total +0.03

**On Country Crossing:**
- When GK reaches a country for the first time: +0.05 (counts once globally per GK, regardless of actor)

**Total possible:** Multiple users add +0.01 each for first drop/grab/seen, plus +0.05 per new country

#### Multiplier Decreases

**Time in Holder's Hands:**
- Every day GK stays with same holder: -0.008/day
- (Encouraged to drop, not carry. A GK with 1.5x multiplier returns to 1.0x in ~62 days)

**Time in Cache:**
- Every week GK sits in same cache: -0.02/week
- (Encourages picking up abandoned GKs. A GK with 1.5x multiplier returns to 1.0x in ~6 months)

**Floor:** Multiplier never goes below 1.0x
**Ceiling:** Multiplier capped at 2.0x (never exceeds)

#### Recomputation Timing
- Multiplier recalculated **after log is registered**
- User points calculated **before multiplier updates** (so they benefit from old multiplier)

---

## 👥 User Points System

### Base Move Scoring

**Per Move (Per User, Per GK):**

**Waypoint Requirement (Anti-Farming Rule):**
- DROP, SEEN, and DIP moves **require a waypoint** to earn points
- If waypoint is null: Those moves earn **0 points** regardless of whether coordinates exist
- **Design:** Ensures points are only awarded for verified locations (official caches/POIs); prevents gaming via unregistered coordinates
- **Exception:** GRAB from inventory does not require waypoint (no physical location involved)

**For Regular Users (not GK owner):**
- First GRAB/DROP/SEEN by user on this GK: **+3 base points** (DROP/SEEN must have waypoint; GRAB unrestricted)
  - Multiplied by GK's current multiplier
  - Example: Drop non-own GK worth +3 base, GK has 1.2x multiplier = 3 × 1.2 = 3.6 points
  - Example: Drop without waypoint = 0 points (waypoint required)
- Further moves by same user on same GK: **0 points**
- DIPs (same user/cache): **0 points** (requires waypoint to earn points)

**For GK Owner (moving own GK):**
- **Standard Types**
  - ALL moves (DROP, GRAB, SEEN, DIP): **0 points**
  - Owner earns points ONLY through bonuses from other players' actions (see Owner-Specific Rules)
- **Non-Transferable** Types
  - ALL moves (DROP, GRAB, SEEN, DIP): **+3 base points** (normal, multiplied by multiplier)
  - DROP, SEEN, DIP **require waypoint** (0 points if null); GRAB unrestricted
  - **Limit:** Only once per GK type per waypoint per calendar month; further moves that month earn 0 points
  - These GKs cannot legitimately change owners, so rewarding owner for circulation is allowed

### Multi-GK Waypoint Penalty

**Rule:** When user moves multiple GeoKrety at the same **location** (waypoint or coordinates) per calendar month:

- **1st GK moved at location:** 100% of points earned
- **2nd GK moved at location:** 50% of points earned
- **3rd GK moved at location:** 25% of points earned
- **4th+ GK moved at location:** 0 points
- **Resets:** Per calendar month (resets at UTC midnight on 1st of each month)
- **Location Identification:**
  - Primary: **Waypoint** (if provided) - unambiguous cache/POI identifier
  - Fallback: **Coordinates** (if waypoint is null) - latitude/longitude location
  - Skip penalty: Only if **both waypoint AND coordinates are null** (rare edge case)

**Design:** Prevents rapid multi-drop farming at a single cache. User earns full points for first move, progressively reduced for subsequent moves at same location. Only 3 GKs per user per location per month earn points. Location is identified by waypoint first, then coordinates as secondary identifier.

**Technical Notes:**
- **Move Type Support**: DROP (always has waypoint/coordinates), DIP (always has waypoint/coordinates), and SEEN (may have waypoint OR coordinates) can all contribute to the location counter. GRAB from inventory has neither and is not penalized.
- **SEEN Special Case**: SEEN can be logged with coordinates but without a waypoint (field observation at non-listed location); location tracking still applies via coordinates fallback.
- **Penalty Scope**: Only affects `base_move` awards. Other bonuses (relay, rescuer, country crossing, etc.) are not penalized by location frequency.

### Points From Chain Completion

**Chain Bonus Formula:** `bonus_per_user = min(chain_length², 8 × chain_length)`

When a movement chain ends (3+ different users in sequence):
- Each user in chain receives: **bonus points capped to prevent explosion**
  - Formula ensures growth up to 8-person chains, then soft cap applies
  - Example: Chain of 3 users → Each gets +9 bonus (min(9, 24) = 9)
  - Example: Chain of 5 users → Each gets +25 bonus (min(25, 40) = 25)
  - Example: Chain of 8 users → Each gets +64 bonus (min(64, 64) = 64)
  - Example: Chain of 10 users → Each gets +80 bonus (min(100, 80) = 80, prevents explosion)
- GK Owner receives: **25% of total chain points** awarded to all participants
  - Example: Chain of 4 distributes 64 points total (4×16) → Owner gets 16 points (25% of 64)

**Design:** Quadratic growth incentivizes longer circulation chains without dominating the point system.

---

## 🎯 User Bonuses

### Diversity Bonuses (Resets monthly)

- **+3 points** - Drop 5 different GKs in one month
- **+7 points** - Interact with 10 different GK owners in one month
- **+5 points** - Get a GK to visit new country (per user, per country, once per month only; can stack with country crossing +3)

**Bonus Stacking:** Diversity country bonus +5 is cumulative with Movement country bonus +3 when both apply:
Example:
- user move GK1 in a new country A, it receive +3 (movement) + new country for the GK bonus +3
- same user move GK1 in a the same country A -> receive +3 (movement)
- same user move GK1 in a new country B -> receive +3 (movement) + new country for the GK bonus +3
- same user move GK2 in a new country C -> receive +3 (movement) + new country for the GK bonus +3
- same user move GK2 in a new country B -> receive +3 (movement) (user already visited this country)
- same user move GK1 in a new country D -> receive +3 (movement) + new country for the GK bonus +3 + Diversity country bonus +5

### Special Movement Bonuses

#### Relay Bonus
- When GK moves within **7 days** of your drop (new user grabs):
  - **Mover receives:** +2 extra points
  - **Previous dropper receives:** +1 extra point
- Design: Encourages fast circulation and leaving GKs for others

#### Rescuer Bonus
- When a **non-owner** user grabs a GK that has been **dormant in a cache for 6+ months** (measured from last DROP or SEEN event, where the grab changes the holder from NULL to a user):
  - **Grabber receives:** +2 points
  - **GK Owner receives:** +1 point
- **Not triggered:** If the current holder is the same as the previous holder (e.g., user spontaneously logs a move on their own inventory after 6 months dormancy)
- Design: Encourages rescuing abandoned GKs and keeping owner engaged

#### First Finder Bonus
- First person to **grab, drop, dip, or see** a brand-new GK **within 7 days of GK creation date**: **These points are added in base move scoring**.
  - First GRAB: Earns +3 base (non-owner only; owner grab = 0 per owner rules)
  - First DROP: Earns +3 base (non-owner, first move only; owner gets 0)
  - First SEEN: Earns +3 base (first move only) - no separate bonus
  - **Hard Cutoff:** Moves on day 8+ of GK existence do NOT qualify for First Finder credits
- Design: Encourages early engagement (discovery/pickup/observation) with freshly created GKs; rewards are included in standard base move points

---

## 👑 Owner-Specific Rules

### Owner's Direct Move Points (when owner moves it)

| Action | Points |
|--------|--------|
| Drop own GK | 0 |
| Grab own GK | 0 |
| See own GK | 0 |
| Dip own GK | 0 |

**Design:** Owner earns NO points from moving their own GKs. All owner points come exclusively from bonus triggers when OTHER players interact with their GKs (handover bonus, country bonus, rescue bonus, chain bonus, reach bonus).

### Points When Others Move Owner's GK

| Event | Points | Notes |
|-------|--------|-------|
| Another user takes your GK from another user | +1 | New holder (holder change from previous value and new is not owner) |
| Your GK enters new country (actor moves it) | +3 | Except the first country it sees, mother-country |
| Your GK reaches 10 different users (6-month window) | +5 | Measures circulation/reach |
| Your GK rescued from 6+ months cache dormancy | +1 | Rescue event (holder was NULL in cache) |

### Owner GK Limit (Prevent Farming)

- A user can earn points from **maximum 10 GKs per GK owner** (globally, not per month)
- Example: Owner A has 100 GKs. User B can earn points on at most 10 of them. Further moves = 0 points
- This limit applies per owner, not globally (User B can earn from 10 of Owner A's GKs AND 10 of Owner C's GKs)
- Design: Prevents users from farming all GKs from prolific owners

---

## 🔗 Movement Chains

**All bonuses stack additively with base points**

### What is a Chain?

A chain is a sequence of **hand-to-hand exchanges** between distinct users. Chain length = **count of unique users** in the chain.

**Chain Membership:**
- Both **DROP actors** and **GRAB actors** add themselves to chain members (each counted once per chain)
- GRAB with holder change (new user): adds to chain
- GRAB self-grab (grabber == current_holder): treated as DIP, no chain membership, timer extended
- DROP: dropper adds to chain
- SEEN: adds to chain if move log contains location data (coordinates and/or waypoint); counted once per user per chain
- DIP: adds to chain (counted once per user per chain); extends timer by ≤1 day

**Chain Pattern:**
- `A:DROP → B:GRAB → B:DROP → C:GRAB → ...`
- User A drops: A joins chain [A]
- User B grabs: B joins chain [A, B]
- User B drops: B already in chain, no new member [A, B]
- User C grabs: C joins chain [A, B, C]
- No additional members if same user moves multiple times

### Move Types & Chain Impact

| Move Type | Adds User to Chain? | Restarts 14-Day Timer? | Effect |
|---|---|---|---|
| **GRAB** (holder change) | ✅ Yes | ✅ Yes | New user joins chain, becomes new holder. **Self-grabs (grabber == current_holder) earn 0 pts and extend timer like DIP** |
| **DROP** (release to cache) | ✅ Yes | ✅ Yes | Dropper joins chain, passes GK forward |
| **SEEN** (observation) | ✅ Yes* | ✅ Yes | Adds to chain if move log contains location data (coordinates/waypoint); resets timer. *No chain membership if location data missing |
| **DIP** (internal move) | ✅ Yes | ✅ Yes (+1-2 days max) | Internal move by current holder; adds to chain but extends timer less than DROP/GRAB/SEEN (1-2 days max instead of full reset) |
| **COMMENT** | ❌ No | ❌ No | Meta only, no chain impact |
| **ARCHIVE** | - | - | **Ends chain immediately** |

### Chain Timeout: 14-Day Inactivity Rule

**Chain officially ends when:**
- **14 consecutive days** pass with no GRAB, DROP, SEEN, or DIP by any user
- GK is archived or deleted by owner

**Timer behavior:**
- GRAB, DROP, SEEN: Fully reset the 14-day countdown
- DIP: Extends remaining time by small amount (1-2 days max), contributing less to chain timer extension than active moves, but total countdown never exceeds 14 days from holder's initial acquisition
- COMMENT: Does not affect timer

**Bonus awarded when chain ends (if chain_length >= 3):**
- Each user in chain receives: `min(chain_length², 8 × chain_length)` bonus points
  - **Anti-Farming Rule:** A player can earn chain bonus from **only ONE chain per GK per 6-month period**. Further chain bonuses within 6 months = 0 points (prevents repetitive chain farming)
- GK Owner receives: 25% of total chain points distributed

**No bonus if chain_length < 3** (insufficient circulation)

### What Breaks a Chain?

- **14+ days of inactivity** - GK sits untouched, chain dies
- **Owner archives GK** - Manual action, chain ends
- **Same user re-grabbing** back their own released GK - Allowed but doesn't add chain length (user already counted)
- **User deletion/banning** - When a user is deleted/banned, their logs become anonymous (user set to NULL). Anonymous logs don't earn points and don't count toward chain length. Chain continues with remaining counted users (those still in the system).

### Current Examples

```
Example 1: Successful 3-person chain (ends naturally)
Feb 1:  A:DROP (starts chain)
Feb 3:  B:GRAB (B joins, timer resets)
Feb 6:  B:DROP (B still chain member, timer resets)
Feb 8:  C:GRAB (C joins, chain now = A, B, C, timer resets)
Feb 24: No moves for 16 days → CHAIN ENDS (Feb 22 = 14-day mark)

Bonus per user: min(3², 24) = min(9, 24) = 9 points
Owner receives: 25% × (3 × 9) = 6.75 points

---

Example 2: Chain with DIPs and SEENs
Feb 1:  A:DROP (chain member: A)
Feb 2:  B:SEEN with location (B joins chain: A, B, timer resets)
Feb 3:  B:GRAB (B already in chain, no new member, timer resets)
Feb 5:  B:DIP 1 (internal move, B still counted once, timer +1-2 days max)
Feb 6:  B:DIP 2 (internal move, B still counted once, timer +1-2 days max)
Feb 8:  B:DROP (B still counted once, timer resets)
Feb 10: C:GRAB (C joins chain: A, B, C, timer resets)
Feb 25: No moves for 15 days → CHAIN ENDS (Feb 24 = 14-day mark)

Chain length: 3 (A, B, C) - B counted once despite multiple moves (GRAB, DIPs, DROP)
Bonus per user: min(3², 24) = 9 points each

---

Example 3: User re-captures their GK (allowed, doesn't reset)
Feb 1:  A:DROP
Feb 3:  B:GRAB (chain: A, B)
Feb 5:  B:DROP
Feb 7:  A:GRAB (A takes it back - allowed but A not re-counted)
Chain still = A, B (length 2, not 3)
Unless C grabs by Feb 21 (14-day mark), chain dies with no bonus (length < 3)
```

---

## 🛡️ GK Type-Specific Rules

### Standard GKs (TRADITIONAL → PEBBLE types)

- Standard points system applies
- Handovers with change-of-hands bonus
- Multiplier system applies normally
- Hoarding penalties apply

### Non-Transferable GKs (CAR, DOG_TAG, EASTER_EGG, etc.)

- **Cannot change owner legitimately** (sealed to one owner)
- **Owner earns +3 base per move** (DROP, GRAB, SEEN, DIP) - multiplied by multiplier
  - **Limit:** Only once per GK type per waypoint per calendar month; further moves that month earn 0 points
- Multiplier system applies normally
- Reward movement over hoarding (since they can't circulate to other owners)

---

## 🌍 Country Rules

### Country Crossing Rules

**Country bonus counted once globally per GK per country (one-time only, starting from 2nd country)**

**Home Country (First Country):**
- When GK is created/dropped in its mother-country: NO multiplier increase, NO country bonus
- This is the baseline; bonuses start from second distinct country visit

**When GK reaches NEW country (via DROP/DIPPED/SEEN move, starting from 2nd country):**
- GK multiplier increases: **+0.05**
- for **Standard Types**
  - **The actor (who moved it to new country) receives:** +3 points
    - **Exception:** If actor is GK owner → owner receives +2 to their points (for Non-transferable GK - +4 points once per user/country/gk_type)
- for **Non-transferable Types**
  - **The actor (who moved it to new country) receives:** +3 points
    - **Exception:** If actor is GK owner → owner receives +4 to their points (once per user/country/gk_type for all time)
- All future moves within that country: 0 country bonus (already visited, one-time reward per country)

**Design:** Encourages international circulation and single-time country exploration; prevents gaming same country repeatedly. Home country doesn't generate country bonuses to keep initial creation incentives focused on drops/circulation

---

## 📐 System Logic Summary

**GK Multiplier System:**
- ✅ Starts at 1.0x per GK
- ✅ Increases on first move types by each user (+0.01 each: drop, grab, seen, dip per user)
- ✅ Increases on country crossing (+0.05 once per country per GK)
- ✅ Decreases over time (-0.008/day in hands, -0.02/week in cache; 1.5x→1.0x in ~62 days or ~6 months)
- ✅ Minimum floor: 1.0x per GK (never goes lower)

**User Points Calculation:**
- ✅ Base +3 per first move by non-owner (multiplied by GK's multiplier)
- ✅ Owner gets 0 points for direct moves on **Standard type** GKs (only earns from bonuses)
- ✅ Owner gets +3 base points for direct moves on **Non-Transferable type** GKs (once per GK type per waypoint per month; these cannot transfer, so owner movement is rewarded)
- ✅ Chain bonus: +min(chain_length², 8×chain_length) per user when chain ends; owner gets 25% of total
- ✅ Relay bonus: +2 mover, +1 previous dropper (within 7 days)
- ✅ Rescuer bonus: +2 grabber, +1 owner (6+ months dormancy in cache, non-owner grab)
- ✅ Handover bonus: +1 to owner when other user takes GK from another user
- ✅ Country crossing: +3 to actor moving GK to new country (once per country); owner may receive +2 or +4 depending on GK type
- ✅ Reach bonus: +5 to owner when GK reaches 10 different users (6-month window)
- ✅ Diversity bonuses reset monthly and are independent
- ✅ DIPs earn 0 points (prevents farming)
- ✅ Multi-GK waypoint penalty: 100% → 50% → 25% → 0% (per cache per user per month)
- ✅ Owner GK limit: Max 10 GKs per owner per user (prevents farming single owner)
- ✅ First Finder: +3 base (non-owner) within 7 days; owner earns 0 on first drop

**Owner Incentives (Standard Type GKs only):**
- ✅ Owner gets 0 points for direct moves, incentivizing them to release GKs for others to circulate
- ✅ Owner gets +1 when others hand over their GK (per handover event)
- ✅ Owner gets +2 or +4 when GK reaches new country (depending on GK type)
- ✅ Owner gets +5 when GK reaches 10 different users (circulation reward)
- ✅ Owner gets +1 when GK rescued from 6+ months dormancy
- ✅ Owner gets 25% of total chain points as incentive for quality GKs
- ✅ Encourages releasing and maintaining circulation over hoarding
