# Module 04 – Waypoint Penalty

## Responsibility

Applies a scaling penalty to base move points when a user interacts with multiple different GeoKrety at the same location (waypoint or coordinates) within the same calendar month. Prevents "cache farming" where a user drops or picks up many GKs at a single location in rapid succession.

This module modifies existing `"base_move"` award entries in the accumulator. It never adds new awards.

---

## Input

From context:
- `event.waypoint` – the cache/waypoint identifier for this move (may be null)
- `event.coordinates` – latitude/longitude data (may be null, used if waypoint is null)
- `event.user_id` (actor)
- `event.logged_at` (to determine calendar month in UTC)
- `user_state.actor_gks_at_location_this_month` – count of distinct GKs the actor has moved at this location (waypoint or coordinates) in the current calendar month, **not counting this event**
- `runtime_flags.actor_scored_this_gk` – true if base points exist to scale

---

## Process

### Step 1 – Determine Location Identity (Waypoint or Coordinates)

```
if event.waypoint is not null:
    → location_id = event.waypoint (primary identifier)
else if event.coordinates is not null:
    → location_id = event.coordinates (fallback identifier)
else:
    → SKIP (no location data available, penalty not applicable)
```

**Location Hierarchy:**
- **Waypoint** is the primary and most precise location identifier (Official cache/POI)
- **Coordinates** (latitude, longitude) is a fallback when waypoint is null
  - SEEN moves may have coordinates but no waypoint (field observation at unlisted location)
  - DROP and DIP always require coordinates
- **Both null:** Penalty skipped (data integrity issue; should be rare)

### Step 2 – Check if Base Points Exist

```
if runtime_flags.actor_scored_this_gk == false → SKIP (nothing to scale)
```

### Step 3 – Determine Penalty Tier

The count `actor_gks_at_location_this_month` represents how many distinct GKs this actor has already scored at this location (waypoint or coordinates) this month BEFORE the current GK. The current GK is the next one in sequence.

```
count = user_state.actor_gks_at_location_this_month
         (number of previous distinct GKs scored at this location this month)
```

| Count (previous GKs) | This GK's position | Scale factor |
|----------------------|--------------------|--------------|
| 0                    | 1st GK             | 100%         |
| 1                    | 2nd GK             | 50%          |
| 2                    | 3rd GK             | 25%          |
| 3+                   | 4th+ GK            | 0%           |

```
scale = switch count:
    0 → 1.00
    1 → 0.50
    2 → 0.25
    3+ → 0.00
```

### Step 4 – Apply Scale

```
if scale == 1.00 → no change to accumulator
if scale < 1.00 → multiply all "base_move" labeled awards by scale
if scale == 0.00 → set all "base_move" labeled awards to 0
                   set runtime_flags.actor_scored_this_gk = false
```

When scale == 0: the actor effectively earned nothing from this GK at this location this month. This also means the GK should NOT be counted toward the actor's location counter (since 0 points were earned). The counter increment side effect should use the original `actor_scored_this_gk` flag from before this zeroing.

**Important**: The counter should be incremented only when at least some base points were earned (scale > 0). Do not increment the location GK count for 0-point events, otherwise the counter would grow from non-scoring events and distort future calculations.

### Step 5 – Update Waypoint Counter

```
if scale > 0.00:
    → Record that this GK has now been scored at this location in this month
      (side effect: the data store will increment actor_gks_at_location_this_month by 1
       for subsequent events this month)
```

---

## Output

### Accumulator Modifications

- `"base_move"` labeled awards are scaled by the applicable factor.
- If scale results in 0 points, those awards are removed or set to 0.

### Awards Added

None.

---

## Notes

- The penalty resets on the 1st of each calendar month at UTC midnight. On the 1st, every user starts fresh at 0 GKs per location, meaning full points for the first GK at any location.
- "Moved at location" means the actor logged any scored event (base points > 0 after modules 02 and 03) at that location (waypoint or coordinates). Events that were zeroed by module 03 or that naturally produce 0 base points (DIP, ARCHIVED) do not count toward the location GK counter.
- The penalty applies per (actor, location, calendar month) triple. Moving at different locations or in different months is fully independent.
- The penalty does **not** affect bonus awards from other modules (country crossing, relay, rescuer, etc.). It only penalizes the `"base_move"` labeled entries.
- A result of 0.25 × 3.6 = 0.9 points is valid and will be passed to the aggregator for rounding.
- If two GKs are processed simultaneously at the same location for the same user (concurrency scenario), the implementation must serialize or use atomic counters to avoid undercharging the penalty.
