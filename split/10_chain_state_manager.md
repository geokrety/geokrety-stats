# Module 10 – Chain State Manager

## Responsibility

Maintains the state of the movement chain for this GeoKret. A chain tracks sequences of hand-to-hand exchanges to eventually reward long-circulation patterns. This module does NOT award points itself; it only updates the chain and sets a flag if the chain has ended. The actual chain bonus award is in module 11.

A separate background job must also call module 11 asynchronously for chains that end due to inactivity timeout (when no new event arrives for 14 days). This module handles only inline chain endings.

---

## Input

From context:
- `event.log_type`
- `event.user_id` (actor)
- `event.logged_at`
- `gk_state.gk_id`
- `chain_state.active_chain_id` – null if no active chain
- `chain_state.chain_members` – ordered list of distinct user_ids in the active chain
- `chain_state.chain_last_active` – UTC timestamp of last chain-activity event
- `chain_state.holder_acquired_at` – when the current holder first took possession (for DIP timer)

---

## Process

### Step 1 – Check for Expired Chain Before Processing

Even before handling the new event, check if the existing chain has already expired due to inactivity:

```
if chain_state.active_chain_id is not null:
    days_since_active = (event.logged_at - chain_state.chain_last_active) in days
    if days_since_active >= 14:
        → The chain expired BEFORE this new event arrived
        → Set chain_state.chain_ended = true and finalize current chain
           (carry this signal to module 11 to award bonuses for the expired chain)
        → Start fresh: clear chain_state to begin a new potential chain below
```

This ensures that if a long gap occurred before this event, the old chain gets its bonus and a new chain can start from this event.

### Step 2 – Handle ARCHIVED Event

```
if event.log_type == ARCHIVED (4):
    if chain_state.active_chain_id is not null:
        → Set chain_state.chain_ended = true
        → Finalize chain
    → No new chain is started for ARCHIVED events
    → STOP (no further chain processing)
```

### Step 3 – Handle COMMENT Event

```
if event.log_type == COMMENT (2):
    → COMMENT has no effect on chain state whatsoever (does not reset timer, does not break chain)
    → STOP
```

This case should not reach module 10 due to the event guard, but is listed for completeness.

### Step 4 – Handle DIP Event

DIP is a special case: it extends the chain timer slightly but cannot extend it beyond 14 days from when the current holder first acquired the GK.

```
if event.log_type == DIP (5):
    if chain_state.active_chain_id is null:
        → No active chain to extend; SKIP
    else:
        max_allowed_deadline = chain_state.holder_acquired_at + 14 days
        current_deadline = chain_state.chain_last_active + 14 days

        new_deadline = min(current_deadline + 1 day, max_allowed_deadline)

        → Update chain_state.chain_last_active to: new_deadline - 14 days
          (i.e., set last_active such that the new 14-day window ends at new_deadline)
        → Do NOT add the DIP actor to chain_members
        → STOP
```

Rationale: DIP can push the deadline by up to 1 day, but the total window from when the holder took the GK cannot exceed 14 days. This prevents infinite chain extension via repeated DIPs.

### Step 5 – Handle GRAB, DROP, SEEN Events (Full Timer Reset)

For GRAB (1), DROP (0), and SEEN (3):

```
→ Fully reset the 14-day countdown: chain_state.chain_last_active = event.logged_at
```

### Step 6 – Update Chain Members (GRAB and DROP)

Both GRAB and DROP events add the actor to the chain member list, because a chain requires at least one person who initiated by dropping (A) and another who picked up (B). The rules show: `A:DROP → B:GRAB` means both A (dropper) and B (grabber) are counted in the chain.

```
if event.log_type in {GRAB (1), DROP (0)}:
    if event.user_id NOT IN chain_state.chain_members:
        → Append event.user_id to chain_state.chain_members
    (if already in members: no change; same user making additional moves)
```

SEEN (3) and DIP (5) do NOT add users to the chain member list. They only affect the timer.

Chain inclusion semantics:
- DROP: the dropper is the person who "puts the GK up for grabs"; they are a chain participant
- GRAB: the grabber joins the chain as a new possessor
- SEEN: observing does not constitute taking possession → no chain membership
- DIP: an internal move by the current holder who is already in the chain (or not) → no new membership

### Step 7 – Ensure Chain Exists

If no active chain exists after step 1 (none was active, or one just ended), start a new one:

```
if chain_state.active_chain_id is null:
    → Create a new chain record:
        active_chain_id = <new unique ID>
        chain_members = []
        chain_last_active = event.logged_at
        holder_acquired_at = event.logged_at (if GRAB; otherwise use current event time)
    → Then apply step 6 (add actor to chain if applicable)
```

Note: When a new chain is created and step 6 executes:
- If this is a DROP event: dropper is added to chain_members
- If this is a GRAB event with holder change: grabber is added to chain_members
- If this is a GRAB self-grab: no member added (treated as DIP)
- If this is a SEEN event: no member added, only timer reset

### Step 8 – Persist Updated Chain State

Save the updated `chain_state` to the data store:
- Updated `chain_members`
- Updated `chain_last_active`
- Updated `holder_acquired_at` (set on new GRAB)

---

## Output

### Accumulator

No awards added. Chain bonuses are the responsibility of module 11.

### Runtime Flags Written to Context

```
context.chain_state.chain_ended = true   (if an existing chain just ended in steps 1 or 2)
context.chain_state.ended_chain_id = <ID of the chain that ended>
context.chain_state.ended_chain_members = <list of distinct user_ids in the ended chain>
```

These flags are read by module 11 to determine whether to compute and award chain bonuses.

---

## Notes

- **Chain Membership**: Both DROP and GRAB actors (when representing holder changes) are added to `chain_members`. Each user is counted once per chain regardless of how many times they move the GK. SEEN events do NOT add members; DIPs do NOT add members.
  - DROP: Dropper joins chain as a participant
  - GRAB (holder change): Grabber joins chain as a participant
  - GRAB (self-grab, grabber == current_holder): Treated as DIP, no chain member
  - SEEN: Timer reset only, no membership
  - DIP: Timer extension only, no membership

- **Chain Length Formula**: Defined as count of distinct users in `chain_members`. Example: A:DROP → B:GRAB → B:DROP → C:GRAB = [A, B, C] = length 3

- **Background Timeout Job**: Chains that expire due to 14+ days of inactivity (no event fires) must be processed by a separate job that calls module 11 to award bonuses. This module handles only inline endings (ARCHIVE events or timeouts detected when a new event arrives).

- **Anonymous Logs**: When users are deleted/banned post-event, their logs become anonymous (user_id = NULL). These entries do not count toward chain members on recomputation. The chain continues with remaining countable users.
