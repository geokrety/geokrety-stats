# Module 11 – Chain Bonus

## Responsibility

Awards chain completion bonuses when a movement chain ends. Triggered either inline (by module 10 setting `chain_ended = true`) or via an asynchronous background job (chain timeout due to 14-day inactivity).

Computes the bonus amount per chain participant and the owner's share, respects per-user anti-farming limits, and emits award records.

---

## Input

From context (inline trigger):
- `chain_state.chain_ended` – flag set by module 10
- `chain_state.ended_chain_members` – ordered list of distinct user_ids (as corrected in module 10 notes)
- `chain_state.ended_chain_id` – ID of the chain that ended
- `gk_state.owner_id`
- `gk_state.gk_id`

For background trigger: same data passed directly from the persisted chain record.

Per-user anti-farming check (for each chain member):
- For each user in `ended_chain_members`: lookup whether they have received a chain bonus for this specific GK within the last 6 months.

---

## Process

### Step 1 – Check If Chain Ended

```
if chain_state.chain_ended == false → SKIP (chain still active)
```

### Step 2 – Compute Chain Length

```
chain_length = count of distinct user_ids in ended_chain_members
  (anonymous users, deleted users are excluded from count)
```

### Step 3 – Check Minimum Length for Bonus

```
if chain_length < 3 → SKIP (no bonus for chains shorter than 3 unique users)
```

### Step 4 – Compute Bonus Per Participant

```
bonus_per_user = min(chain_length², 8 × chain_length)
```

Examples:
| Chain length | chain_length² | 8 × chain_length | bonus_per_user |
|---|---|---|---|
| 3 | 9 | 24 | 9 |
| 4 | 16 | 32 | 16 |
| 5 | 25 | 40 | 25 |
| 8 | 64 | 64 | 64 |
| 9 | 81 | 72 | 72 |
| 10 | 100 | 80 | 80 |
| 12 | 144 | 96 | 96 |

The formula grows quadratically until chain_length = 8 (where both sides equal 64), then the linear cap kicks in, preventing runaway bonuses for very long chains.

### Step 5 – Apply Anti-Farming Filter per User

For each user in `ended_chain_members`:

```
check: did this user receive a chain bonus on GK #<gk_id> within the last 6 months?

if yes: this_user_award = 0 (anti-farming rule; they already earned a chain bonus recently)
if no:  this_user_award = bonus_per_user
```

This is evaluated independently per user. Some users in the chain may have farmed it recently and receive 0; others may be fresh participants and receive the full bonus.

### Step 6 – Award Bonus to Each Chain Participant

For each user in `ended_chain_members` where `this_user_award > 0`:

```
emit award:
{
  recipient_user_id : user_id,
  points            : bonus_per_user,
  reason            : "Chain bonus: GK #<gk_id>, chain of <chain_length> users
                       (chain #<chain_id>)",
  module_source     : "11_chain_bonus",
  label             : "chain_participant",
  is_owner_reward   : false
}
```

### Step 7 – Compute Owner's Chain Share

The owner receives an additional 25% royalty on top of all participant awards, regardless of whether the owner was also a chain participant themselves.

```
total_distributed = sum of this_user_award for ALL non-owner participants in chain
  (exclude the owner's own participant award from the base, since the 25% is a royalty
   for their GK circulating through OTHER people's hands)

owner_chain_award = 0.25 × total_distributed
```

**Rationale**: The 25% share is the owner's reward for creating a popular GK that others want to carry. The rules example says: "Chain of 4 distributes 64 points total (4×16) → Owner gets 16 points". This implies the 4 participants are the 4 chain members (which may or may not include the owner). The cleanest interpretation consistent with the example is:
- If owner IS in the chain (e.g., as the initial dropper), they earn the participant bonus like anyone else, PLUS 25% of the points distributed to the other (non-owner) participants
- If owner is NOT in the chain, they earn 25% of the total distributed to all participants

In practice, when the owner starts the chain with a DROP:
```
other_participant_points = sum of this_user_award for participants WHERE user_id != owner_id
participant_award_for_owner = this_user_award for owner (may be 0 if owner was anti-farmed)
owner_chain_award = 0.25 × other_participant_points
```

If `owner_chain_award == 0` (because all other participants were farmed or chain length < 3): no 25% award.

### Step 8 – Award Owner's Chain Share

```
if owner_chain_award > 0:
    emit award:
    {
      recipient_user_id : gk_state.owner_id,
      points            : owner_chain_award,
      reason            : "Chain bonus (owner share): GK #<gk_id>, chain of
                           <chain_length> users, 25% of distributed points
                           (chain #<chain_id>)",
      module_source     : "11_chain_bonus",
      label             : "chain_owner_share",
      is_owner_reward   : true
    }
```

### Step 9 – Record Chain Completion

Mark the chain as closed in the data store. For each user who received a non-zero chain bonus, record: `(user_id, gk_id, chain_completion_timestamp)` for future anti-farming lookups.

---

## Output

### Awards Added to Accumulator

When triggered (chain_length >= 3):
- Up to N awards for chain participants (one per non-farmed member)
- 1 award for the GK owner (25% share of distributed total), if > 0

When not triggered: nothing.

---

## Notes

- The owner's 25% share can be fractional (e.g., 3 users × 9 points each = 27 total → owner gets 6.75). The aggregator handles rounding.
- The anti-farming check (step 5) means a user who participates in chain after chain on the same GK within 6 months only earns the bonus once per 6-month period. After 6 months, they're eligible again.
- The anti-farming rule applies to the PARTICIPANT bonus, not to the OWNER's 25% share. The owner always receives 25% of whatever was actually distributed, even if many participants are anti-farming zeroed. If ALL participants are zeroed (total_distributed = 0), the owner gets nothing.
- Background job processing: when a background timeout scanner detects an expired chain, it should call this module with the chain data, using the chain's last known GK state and membership list. The pipeline then runs from module 11 onwards (no event guard or base scoring is re-run; only the chain bonus and diversity/aggregator modules fire).
- Chain bonuses are NOT subject to the waypoint penalty (module 04) or owner GK limit (module 03). They are event-independent bonuses for circulation behavior, not per-event scoring.
