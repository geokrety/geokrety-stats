# Module 02 – Base Move Points

## Responsibility

Calculates the raw base points earned by the **actor** (the user performing the move) for the current event. This is the primary point-earning mechanism for non-owner users.

Base points are always multiplied by the GK's current multiplier (the value before this event, loaded in module 01).

The result is tagged with label `"base_move"` in the accumulator so that downstream penalty modules (03, 04) can identify and modify it.

---

## Input

From context:
- `event.log_type`
- `event.logged_at`
- `event.user_id` (actor)
- `gk_state.owner_id`
- `gk_state.gk_type`
- `gk_state.created_at`
- `gk_state.current_multiplier`
- `gk_state.current_holder` – needed for self-grab validation
- `event.waypoint`
- `event.logged_at` (for first-finder window check)
- `user_state.actor_move_history_on_gk` – set of log_types actor has already logged on this GK

---

## Process

### Step 1 – Check Log Type Eligibility for Base Points

ARCHIVED (4) always produces 0 base points regardless of any other condition:

```
if event.log_type == ARCHIVED (4) → base_points = 0, SKIP remaining steps
```

DIP (5) produces 0 points **except** for the owner of a non-transferable GK (handled in step 6). For all other actors, skip early:

```
if event.log_type == DIP (5) AND NOT (is_owner AND gk_state.gk_type in {8,9,10}):
    base_points = 0, SKIP remaining steps
```

Note: `is_owner` is computed here provisionally as `event.user_id == gk_state.owner_id`. Steps 2–6 will also use this value.

### Step 1b – Self-Grab Validation (Prevent Circumvention)

A self-grab (GRAB where the grabber already holds the GK) is not a valid transfer. It earns 0 points and should be treated as a DIP.

```
if event.log_type == GRAB (1):
    if event.user_id == gk_state.current_holder:
        → base_points = 0, SKIP remaining steps
           (self-grab, no holder change; treat as internal move)
```

This prevents circumvention where users grab their own GK then immediately drop it to earn points instead of logging a DIP.

### Step 2 – Waypoint Requirement for Location-Based Moves (Anti-Farming Rule)

Moves that involve specific locations (DROP, SEEN, DIP) **require a waypoint** to earn points:

```
if event.log_type in {DROP (0), SEEN (3), DIP (5)}:
    if event.waypoint is null:
        → base_points = 0, SKIP remaining steps
           (waypoint required; unverified locations do not earn points)
```

**Scope:**
- **Restricted moves** (require waypoint): DROP, SEEN, DIP
- **Unrestricted moves** (no waypoint required): GRAB (inventory-based, no physical location)
- **Design:** Ensures points are earned only at verified locations (official caches/POIs); prevents gaming via unregistered coordinates

### Step 3 – Owner vs Non-Owner

Determine whether the actor is the GK owner:

```
is_owner = (event.user_id == gk_state.owner_id)
```

### Step 4 – Owner / Standard GK → Always 0

```
if is_owner AND gk_state.gk_type in {0..7} (standard types):
    base_points = 0
    SKIP remaining steps
```

Owners of standard GKs earn no points from their own direct moves. All owner income comes from bonuses triggered by other players (modules 05–11).

### Step 5 – Non-Owner OR Non-Transferable Owner: First-Move Check

Check if this is the actor's first time logging this specific log_type on this GK:

```
is_first_move = event.log_type NOT IN user_state.actor_move_history_on_gk
```

```
if NOT is_first_move:
    base_points = 0
    SKIP remaining steps
```

Rationale: a user earns base points at most once per log_type per GK. Subsequent moves of the same type on the same GK earn nothing (prevents repeated-interaction farming).

### Step 6 – First Finder Window (for NEW GKs only)

The First Finder rule is an **eligibility filter**, not a bonus. It applies only for brand-new GKs.

Window definition: the GK qualifies as "new" if:
```
(event.logged_at - gk_state.created_at) < 7 days (168 hours)
```
Moves on day 8 or later of GK existence do NOT qualify (rules state: "Hard Cutoff: Moves on day 8+ do NOT qualify for First Finder credits"). Day 7 (hours 144–167 since creation) is still within the window; hour 168 is not.

If the GK is **7 or more days old** (>= 168 hours) at the time of this event, the First Finder consideration does not apply at all – standard first-move logic continues normally.

If the GK is **less than 7 days old (< 168 hours)** at the time of this event:
- The actor qualifies as a First Finder, and the standard +3 base points apply (with multiplier)
- This is NOT a separate bonus – it is simply normal base move scoring; the same +3 base applies
- For non-owners: +3 × multiplier (same as normal)
- For owner of standard type: still 0 (owner rule takes precedence)
- For owner of non-transferable type: +3 × multiplier (follows the non-transferable owner rule below)

There is no additional bonus for being a First Finder beyond the normal +3 × multiplier.

### Step 7 – Non-Transferable GK Owner Monthly Limit

This step only applies when:
- `is_owner == true`
- `gk_state.gk_type in {8, 9, 10}` (non-transferable: CAR, DOG_TAG, EASTER_EGG)

The owner of a non-transferable GK can earn +3 base per move, but only **once per (gk_type, waypoint, calendar month)** combination.

Check:
```
already_scored_this_combo = has actor already earned base points for any GK of type
  gk_state.gk_type, at waypoint event.waypoint, during current calendar month?
```

```
if already_scored_this_combo:
    base_points = 0
    SKIP remaining steps
```

If `event.waypoint` is null, treat null as a unique waypoint value per event (no deduplication possible without a waypoint).

### Step 8 – Compute Base Points

All remaining cases earn the standard base amount:
```
base_points = 3 × gk_state.current_multiplier
```

This applies to:
- Non-owner users: first DROP/GRAB/SEEN on this GK (DROP and SEEN must have waypoint per Step 2)
- GRAB moves with waypoint or without (inventory-based, no waypoint required)
- Owners of non-transferable GKs: first DROP/GRAB/SEEN/DIP of this GK type at this waypoint this month (all location moves must have waypoint per Step 2)

---

## Output

### Awards Added to Accumulator

If `base_points > 0`:

```
{
  recipient_user_id : event.user_id,
  points            : 3 × gk_state.current_multiplier,
  reason            : "First [DROP|GRAB|SEEN] of GK #<gk_id> by user #<user_id>
                       (multiplier: <multiplier>x)",
  module_source     : "02_base_move_points",
  label             : "base_move",
  is_owner_reward   : false
}
```

### Runtime Flags Written to Context

```
context.runtime_flags.base_points_awarded = base_points (raw, before penalty)
context.runtime_flags.actor_scored_this_gk = (base_points > 0)
```

If `base_points == 0`:
- No award is added
- `actor_scored_this_gk` is set to `false`

---

## Notes

- The multiplier used here is `gk_state.current_multiplier` – the value **before** this event. Module 13 will update the multiplier after all scoring is complete.
- Points may be fractional (e.g., 3 × 1.25 = 3.75). The aggregator (module 14) is responsible for rounding decisions.
- The "First Finder" rule produces no separate line item. It is simply the normal base +3 × multiplier that happens to be the actor's first-ever interaction with a new GK. The timing window (< 168 hours / less than 7 full days) is what qualifies the move; the points formula is identical.
- DIP for non-transferable GK owner: step 1 allows the DIP to fall through (it is NOT zeroed early) because non-transferable owners DO earn points for DIPs. Step 6 then applies the monthly limit check (once per gk_type per waypoint per month). The DIP must also pass the first-move check (step 4) — owner earns nothing on a repeated DIP for the same GK if they already logged a move of the same type before.
- DIP for non-owner: zeroed in step 1. DIPs are internal inventory moves; non-owners observing or dipping a GK earn nothing from that action.
- DIP for owner of standard GK: also zeroed in step 1 (since is_owner is true but gk_type is NOT non-transferable, the condition `NOT (is_owner AND gk_type in {8,9,10})` evaluates to true, so base_points = 0).
