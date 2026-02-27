# Module 07 – Rescuer Bonus

## Responsibility

Awards a bonus when a non-owner user grabs a GeoKret that has been sitting dormant in a cache for 6 or more months. Rewards the grabber for rescuing an abandoned GK and notifies the owner with a smaller bonus to encourage re-engagement.

---

## Input

From context:
- `event.log_type`
- `event.user_id` (actor = potential rescuer)
- `event.logged_at`
- `gk_state.owner_id`
- `gk_state.current_holder` – who holds the GK before this event (must be null for cache dormancy)
- `gk_history.last_cache_entry_at` – UTC timestamp of the most recent event that placed the GK in a cache (last DROP or SEEN while holder was null)

---

## Process

### Step 1 – Check Move Type

Only a GRAB can trigger a rescue. The actor must be physically taking the GK out of a cache.

```
if event.log_type != GRAB (1) → SKIP
```

### Step 2 – Check That the GK Was in a Cache

The GK must currently be sitting in a cache (not in someone's hands). A rescue means removing it from cache dormancy.

```
if gk_state.current_holder is not null → SKIP
  (GK is with a user, not in cache; no dormancy to rescue from)
```

### Step 3 – Check That the Grabber is Not the Owner

Owners cannot rescue their own GKs. The rescuer must be a different user.

```
if event.user_id == gk_state.owner_id → SKIP
```

### Step 4 – Check the 6-Month Dormancy Threshold

```
if gk_history.last_cache_entry_at is null → SKIP
  (no prior cache placement recorded; cannot determine dormancy)

months_dormant = (event.logged_at - gk_history.last_cache_entry_at) in months

if months_dormant < 6 → SKIP
  (GK has been in cache less than 6 months; not qualifying as dormant)
```

The 6-month window is measured precisely from the last time the GK was placed in a cache to the moment of this GRAB.

### Step 5 – Award Rescuer Bonus to Grabber

```
emit award:
{
  recipient_user_id : event.user_id,
  points            : 2,
  reason            : "Rescuer bonus: GK #<gk_id> dormant in cache for
                       <months_dormant> months (grabber)",
  module_source     : "07_rescuer_bonus",
  label             : "rescuer_grabber",
  is_owner_reward   : false
}
```

### Step 6 – Award Rescuer Bonus to Owner

```
emit award:
{
  recipient_user_id : gk_state.owner_id,
  points            : 1,
  reason            : "Rescuer bonus: GK #<gk_id> rescued from cache by user #<user_id>
                       after <months_dormant> months dormancy (owner)",
  module_source     : "07_rescuer_bonus",
  label             : "rescuer_owner",
  is_owner_reward   : true
}
```

---

## Output

### Awards Added to Accumulator

When triggered: 2 awards:
1. +2 to the actor (rescuer / grabber)
2. +1 to the GK owner

When not triggered: nothing added.

---

## Notes

- The `last_cache_entry_at` is set by the most recent DROP or SEEN event (while the GK was already unattended in the cache). It is NOT the timestamp of the GK's initial creation.
- If the GK has never been placed in a cache (e.g., it was only used in person-to-person transfers), `last_cache_entry_at` is null and the module skips. This prevents errors for GKs that were never in a physical cache.
- The "not triggered if current holder equals previous holder" clause from the rules: this is captured by the check `gk_state.current_holder is not null` (step 2). If a user already holds the GK and logs a new move spontaneously after 6 months, the holder is not null and the rescue does not fire. The rescue strictly requires the GK to be in a cache (null holder) immediately before the GRAB.
- The rescue bonus stacks with base move points. The rescuer still earns their +3 × multiplier (from module 02) if this was their first GRAB on this GK, plus this +2 rescue bonus.
- There is no monthly cap or GK cap on rescue bonuses. Each qualifying 6-month dormant rescue triggers independently.
