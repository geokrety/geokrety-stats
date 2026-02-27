# Module 08 – Handover Bonus

## Responsibility

Awards the GK owner a small bonus every time their standard-type GeoKret changes hands between non-owner users. Rewards owners for creating popular GKs that circulate among many people.

---

## Input

From context:
- `event.log_type`
- `event.user_id` (actor = the new holder)
- `gk_state.owner_id`
- `gk_state.gk_type`
- `gk_state.current_holder` – who holds the GK immediately BEFORE this event (loaded in module 01 as the "previous holder" in practical terms, this is the holder state before the new GRAB)
- `gk_state.previous_holder` – the holder immediately before `current_holder` (for chain context)

---

## Process

### Step 1 – Check Move Type

Handovers only happen at GRAB events. The actor is taking the GK from wherever it is.

```
if event.log_type != GRAB (1) → SKIP
```

### Step 2 – Check GK Type

The handover bonus applies to **standard GKs** only (types 0–7).

Non-transferable GKs (types 8–10) are designed for a single owner; "handover" is not a meaningful concept for them. The owner of a non-transferable GK already earns points via base move scoring and other mechanisms.

```
if gk_state.gk_type not in {0..7} → SKIP
```

### Step 3 – Check That the New Holder is Not the Owner

The point of the handover bonus is rewarding change-of-hands BETWEEN other users. The owner picking up their own GK is not a handover.

```
if event.user_id == gk_state.owner_id → SKIP
```

### Step 4 – Check That the GK Was Previously Held by Someone (Not in Cache)

A handover requires a previous non-null holder. If the GK was sitting in a cache (current_holder == null), this is a grab from cache, not a handover from person to person.

```
if gk_state.current_holder is null → SKIP
  (GK was in cache; this is a cache-grab, not a human handover)
```

Note: `gk_state.current_holder` here is the holder BEFORE this GRAB event (i.e., the person who had it before). This must be the holder at the time module 01 loaded state.

### Step 5 – Check That the Previous Holder is Not the Owner

The rule specifies "Another user takes your GK **from another user**", meaning both the previous and new holders are non-owner users. When the owner physically hands the GK to another user, that is an owner→user transfer, not a user→user handover, so no bonus fires.

```
if gk_state.current_holder == gk_state.owner_id → SKIP
  (GK was with owner; this is owner-to-user transfer, not user-to-user handover)
```

So the handover bonus fires only when ALL of the following are true:
- GRAB event ✅ (step 1)
- GK is standard type ✅ (step 2)
- New holder (actor) is NOT the owner ✅ (step 3)
- Previous holder is NOT null (GK was in someone's hands) ✅ (step 4)
- Previous holder is NOT the owner (true user-to-user handover) ✅ (step 5)

### Step 6 – Award Handover Bonus to Owner

```
emit award:
{
  recipient_user_id : gk_state.owner_id,
  points            : 1,
  reason            : "Handover bonus: GK #<gk_id> passed from user #<prev_holder>
                       to user #<actor> (owner bonus)",
  module_source     : "08_handover_bonus",
  label             : "handover_owner",
  is_owner_reward   : true
}
```

---

## Output

### Awards Added to Accumulator

When triggered: 1 award:
- +1 to the GK owner

When not triggered: nothing.

---

## Notes

- The handover bonus can trigger many times as a GK circulates among users. There is no monthly cap or total limit on handover bonuses.
- This bonus is one of the key mechanisms rewarding owners with passive income as their GK circulates. Combined with the reach bonus (module 09) and chain bonus (module 11), active GKs earn their owners meaningful points.
- The distinction between a cache grab and a user-to-user handover is important: a cache grab means `current_holder = null`, which is skipped in step 4. A user-to-user handover means `current_holder = some non-owner user_id`, which proceeds to step 6.
- Non-transferable GKs (types 8–10) are excluded because they have a different reward model. Their owner earns base points directly for every move.
