# Module 03 – Owner GK Limit Filter

## Responsibility

Enforces the anti-farming rule that limits how many different GKs from a single owner a user can earn base points from. Once a user has earned base points on **10 distinct GKs** owned by a specific owner, any base move points earned on additional GKs from that same owner are zeroed out.

This is a filter that modifies existing base_move awards in the accumulator; it never adds new awards.

---

## Input

From context:
- `event.user_id` (actor)
- `gk_state.owner_id`
- `gk_state.gk_id` (or `event.gk_id`)
- `user_state.actor_gks_per_owner_count` – count of distinct GKs from this owner on which the actor has previously earned base points (all time, globally)
- `runtime_flags.actor_scored_this_gk` – set by module 02; true if base points were just awarded

---

## Process

### Step 1 – Check if This Is a Non-Owner Move

This limit only applies to non-owner actors interacting with someone else's GK.

```
if event.user_id == gk_state.owner_id → SKIP (owners aren't limited by this rule)
```

### Step 2 – Check if Actor Even Earned Points This Event

```
if runtime_flags.actor_scored_this_gk == false → SKIP (no base points to zero out)
```

Rationale: if module 02 didn't award base points (e.g., not a first move, wrong log type), there is nothing to filter here.

### Step 3 – Check if This GK Was Already Counted for This Actor-Owner Pair

Look up whether the actor has already earned base points on THIS specific GK from this owner in a previous event.

```
gk_already_counted = has actor already earned base points on gk_id from owner_id before?
```

```
if gk_already_counted:
    SKIP (this GK is already in the actor's 10, no need to check limit again)
```

Rationale: the 10-GK limit tracks distinct GKs. If this GK was counted before, adding a new move on the same GK does not consume another slot.

### Step 4 – Check Limit for New GK

This is the first time the actor is earning base points on this GK from this owner. Check the running count:

```
if user_state.actor_gks_per_owner_count >= 10:
    → Zero out all awards with label "base_move" in the accumulator
    → Set runtime_flags.actor_scored_this_gk = false
    → Emit a "limit reached" note (no points; for logging/debugging only)
    → STOP
```

```
if user_state.actor_gks_per_owner_count < 10:
    → Base points are NOT zeroed; they remain in the accumulator
    → The data store must record that this GK now counts toward the actor's
      10-GK-per-owner limit (this is a side effect: increment the counter)
    → STOP
```

---

## Output

### Accumulator Modifications

- If limit was exceeded: all `"base_move"` labeled awards in the accumulator are set to 0 points (or removed).
- If limit was NOT exceeded (and this is a first-earn on this GK): no change to accumulator; side effect of incrementing the actor-owner counter is recorded.

### Awards Added

None. This module only removes or zeroes existing awards.

---

## Notes

- This limit is **global** (not per month, not per year). Once a user has earned points on 10 GKs from owner X, they can never earn base points on a new GK from owner X again.
- The limit does **not** affect owner-directed bonuses. Modules 05–11 may still award points TO the owner regardless of this filter. The filter only affects what the ACTOR earns.
- The limit does **not** affect the relay bonus (module 06), rescuer bonus (module 07), or country crossing actor bonus (module 05). Those are event-triggered bonuses, not base move points. They carry different labels and are not affected by this zeroing step.
- A practical scenario: User Bob has moved 10 different GKs belonging to Alice. If Bob moves an 11th Alice-GK, he earns 0 base points. But if that 11th move results in a country crossing, Bob still earns the country crossing point from module 05 (it is labeled "country_crossing", not "base_move"). Similarly, if Bob drops Alice's GK and another user grabs it quickly, the relay bonus to Bob still applies.
- The anti-farming intent of this rule is focused on the base repetitive behavior (moving the same owner's GKs repeatedly), not on rare event bonuses.
