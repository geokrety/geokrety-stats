# Module 13 – GK Multiplier Updater

## Responsibility

Updates the GeoKret's multiplier value AFTER all point scoring is complete. This ordering is critical: all modules 02–12 use the multiplier value that existed BEFORE this event; this module writes the new value for use by future events.

This module produces no point awards. It is a state mutation module only.

---

## Input

From context:
- `event.log_type`
- `event.user_id`
- `event.logged_at`
- `gk_state.current_multiplier` – the multiplier value before this event
- `gk_state.gk_id`
- `gk_state.current_holder` – who will be the holder AFTER this event
- `gk_state.last_holder_change_at` – timestamp when the current holder acquired the GK
  (used to compute ongoing time decay since last update)
- `runtime_flags.new_country_visited` – true if this event triggered a new country
- `user_state.actor_move_history_on_gk` – which log_types this user has previously logged on this GK
  (before this event; used to determine if this is the user's first time for this type)

---

## Process

### Step 1 – Apply Time Decay First

Decay has been accumulating since the last multiplier update. Compute how much time has passed since the last update and apply the appropriate decay.

**In-holder decay** (GK is being carried by a user):
```
if gk_state.current_holder is not null:
    days_in_hands = (event.logged_at - gk_state.last_multiplier_update_at) in days
    decay = days_in_hands × 0.008
    current_multiplier -= decay
```

**In-cache decay** (GK is sitting in a cache):
```
if gk_state.current_holder is null:
    weeks_in_cache = (event.logged_at - gk_state.last_multiplier_update_at) in weeks
    decay = weeks_in_cache × 0.02
    current_multiplier -= decay
```

Apply floor:
```
current_multiplier = max(current_multiplier, 1.0)
```

### Step 2 – First Move Type Increase (per user per move type)

```
if event.log_type in {DROP(0), GRAB(1), SEEN(3), DIP(5)}:
    if event.log_type NOT IN user_state.actor_move_history_on_gk:
        current_multiplier += 0.01
        → Mark this log_type as now "seen" for this user on this GK
          (side effect: update actor_move_history_on_gk)
```

Note: COMMENT (2) and ARCHIVED (4) do NOT increase the multiplier.

### Step 3 – Country Crossing Increase

```
if runtime_flags.new_country_visited == true:
    current_multiplier += 0.05
```

This +0.05 applies once per new country per GK, matching module 05's detection.

### Step 4 – Apply Ceiling

```
current_multiplier = min(current_multiplier, 2.0)
```

The multiplier cannot exceed 2.0x under any circumstances.

### Step 5 – Apply Floor (again, for safety)

```
current_multiplier = max(current_multiplier, 1.0)
```

### Step 6 – Persist New Multiplier

Write the new multiplier value to the GK record:
```
gk_state.current_multiplier = current_multiplier (new value)
gk_state.last_multiplier_update_at = event.logged_at
```

---

## Output

### Accumulator

No awards added.

### Side Effects

- `gk_state.current_multiplier` updated in data store
- `gk_state.last_multiplier_update_at` updated
- `actor_move_history_on_gk` updated in data store for this user (if first-move increase applied)

---

## Notes

- **Critical ordering**: All other modules use `gk_state.current_multiplier` as it was at the START of the event (loaded in module 01). This module is deliberately last in the sequence so that the old multiplier is used for all scoring decisions in this pipeline run.
- The time decay calculation requires knowing when the multiplier was last updated (`last_multiplier_update_at`). This field must be stored alongside the multiplier and updated here each time.
- Decay is calculated on a continuous basis at event time: a GK with no moves for 3 months gets all 3 months of decay applied at once when the next event fires.
- **Edge case: between-event decay**: The multiplier shown to users between events should account for ongoing decay even without a triggering event. The last known `current_multiplier` and `last_multiplier_update_at` allow computing the "live" multiplier at any query time by applying the same decay formula on the fly, without writing to the database. The database value is only updated on actual events (this module).
- The first-move tracking (step 2) records `(user_id, gk_id, log_type)` tuples. This data is shared with module 02 for base points eligibility (module 01 loads it; both module 02 and module 13 use it). Module 13 is responsible for the permanent write of the new log_type for this user.
- The country crossing increase in step 3 uses the same flag set by module 05 so that both the multiplier update and the country bonus award are driven by the same detection. This prevents any discrepancy between "module 05 awarded country points" and "module 13 increased multiplier for country".
