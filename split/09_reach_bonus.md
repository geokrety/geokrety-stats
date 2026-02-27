# Module 09 – Reach Bonus

## Responsibility

Awards the GK owner a one-time bonus each time their GeoKret reaches the milestone of **10 distinct users** within a rolling 6-month window. Rewards owners for having actively circulating GKs that reach many people.

---

## Input

From context:
- `event.log_type`
- `event.user_id` (actor)
- `event.logged_at`
- `gk_state.owner_id`
- `gk_state.gk_type`
- `gk_history.distinct_users_6m` – list of distinct user_ids who have scored a move on this GK in the 6 months prior to `event.logged_at` (loaded BEFORE this event's actor is counted)
- `runtime_flags.actor_scored_this_gk` – true if the actor just earned base points for this move

---

## Process

### Step 1 – Check Move Produced Points

The reach bonus is only triggered by a move that actually earned points. If the actor didn't score, they don't count toward the 10-user milestone.

```
if runtime_flags.actor_scored_this_gk == false → SKIP
```

### Step 2 – Check That Actor Is Not the Owner

The owner themselves is not counted in the "distinct users who interacted with the GK". The milestone measures how many OTHER people engaged with the GK.

```
if event.user_id == gk_state.owner_id → SKIP
```

### Step 3 – Check If Actor Is Already in the 6-Month Window

```
if event.user_id IN gk_history.distinct_users_6m → SKIP
  (actor already counted in the rolling window; this move doesn't add a new person)
```

### Step 4 – Compute New Distinct Count

The actor is a new user in the 6-month window:

```
new_count = count(gk_history.distinct_users_6m) + 1
  (already loaded list does NOT include actor yet; +1 for actor)
```

### Step 5 – Check Milestone

The reach bonus fires exactly when the count crosses the 10-user threshold:

```
previous_count = count(gk_history.distinct_users_6m)

if previous_count < 10 AND new_count >= 10:
    → milestone just reached, proceed to award
else:
    → SKIP (already beyond milestone, or not yet reached)
```

This "just crossed" logic ensures the bonus is awarded exactly once per time the GK reaches the 10-user threshold. After reaching 10, the rolling window naturally shrinks as old entries age out. If the count later drops below 10 and rises again to 10, the bonus fires again.

### Step 6 – Award Reach Bonus to Owner

```
emit award:
{
  recipient_user_id : gk_state.owner_id,
  points            : 5,
  reason            : "Reach bonus: GK #<gk_id> reached 10 distinct users
                       in the last 6 months (owner)",
  module_source     : "09_reach_bonus",
  label             : "reach_owner",
  is_owner_reward   : true
}
```

---

## Output

### Awards Added to Accumulator

When triggered: 1 award:
- +5 to the GK owner

When not triggered: nothing.

---

## Notes

- The 6-month window is **rolling**, not calendar-based. At any point in time, only moves logged within the last 6 months (from the current event's timestamp) count toward the distinct-user list. Old events age out naturally.
- This means the reach bonus can fire **multiple times** over the lifetime of a GK:
  - First time: 10th person in any 6-month window
  - Later: if the GK goes quiet and the rolling window drops below 10, then picks up again – the 10th threshold can be crossed again
- There is no total cap on how many reach bonuses an owner can receive for a single GK. An actively circulating GK can trigger this bonus repeatedly over months and years.
- The `gk_history.distinct_users_6m` list loaded in module 01 does NOT include the current event's actor. This module performs the "+1" check locally to detect the threshold crossing.
- Applies to both standard and non-transferable GK types. The owner benefits from both.
- The distinct-user counter only counts users who earned at least some base points (i.e., `actor_scored_this_gk == true`). DIP events, ARCHIVED events, and events that were zeroed by module 03 do not count a person toward the reach milestone.
