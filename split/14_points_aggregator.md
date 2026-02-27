# Module 14 – Points Aggregator

## Responsibility

Final module in the pipeline. Collects all award records accumulated during the pipeline run, validates them, resolves any edge cases, and emits the definitive list of point awards to be persisted and applied to user balances.

---

## Input

- The complete awards accumulator: list of all award records emitted by modules 02–12
- Context (for validation reference):
  - `event.log_id`
  - `event.user_id`
  - `gk_state.gk_id`
  - `gk_state.owner_id`

---

## Process

### Step 1 – Validate All Awards

For each award in the accumulator:

```
assert award.recipient_user_id is not null
assert award.points >= 0           (no negative awards)
assert award.module_source is set  (for auditability)
assert award.label is set
```

Any award failing validation is discarded and an error is logged. Discards should be rare and indicate a programming error in upstream modules.

### Step 2 – Remove Zero-Point Awards

```
remove all awards where points == 0
```

Zero-point awards were possibly zeroed out by module 03 (owner limit) or module 04 (waypoint penalty). They are removed here to produce a clean final list.

### Step 3 – Consolidate Multiple Awards to Same Recipient from Same Module

In rare cases (e.g., a user participating as both actor and chain participant), the same module might emit multiple awards to the same recipient. Merge those into a single line item:

```
for each (recipient_user_id, module_source) pair with duplicate entries:
    merge into single award: sum the points, concatenate reasons
```

Different modules may legitimately emit separate awards to the same recipient (e.g., +3 base from module 02 AND +2 relay from module 06 both go to the actor). These are kept as separate entries for clarity and debugging; only same-module duplicates are merged.

### Step 4 – Round Fractional Points

Base move points can be fractional due to multiplier multiplication (e.g., 3 × 1.25 = 3.75). Some chain calculations may produce fractions (e.g., 6.75 for an owner's chain share).

```
for each award:
    award.points = round(award.points, 2)  ← keep up to 2 decimal places, OR
    award.points = ceil(award.points)       ← always round up to nearest integer
```

The rounding strategy (bank rounding, ceiling, floor, nearest integer) is a product decision. This module applies whichever rounding policy is configured. The default recommendation is **round half-up to nearest integer** to keep balances simple for users.

### Step 5 – Build Per-Recipient Summary

Group awards by recipient for the final output:

```
for each unique recipient_user_id in the award list:
    total_points = sum of all award.points for this recipient
    award_breakdown = list of {points, reason, label, module_source}
```

### Step 6 – Emit Final Output

The final output is a list of per-recipient summaries with full breakdowns:

```
[
  {
    recipient_user_id : integer,
    total_points      : float (rounded),
    event_log_id      : event.log_id,
    gk_id             : gk_state.gk_id,
    awards            : [
      {
        points        : float,
        reason        : string (human-readable),
        label         : string (machine-readable category),
        module_source : string,
        is_owner_reward: boolean
      },
      ...
    ]
  },
  ...
]
```

This output is used to:
1. **Persist to the database**: write individual award records linked to `event.log_id`
2. **Update user point balances**: apply `total_points` to each recipient's balance
3. **Notify users**: optionally surface the breakdown to users (why they got points)

---

## Output

### Final Award List

Structured list as described in Step 6. Empty list if no points are awarded (e.g., event was a COMMENT that passed guard, or all awards were zeroed).

### Awards Added to Accumulator

None (this is the terminal module; no further modules run after it).

---

## Notes

- The `event_log_id` links each award record to the triggering event. This enables auditing, recalculation, and dispute resolution.
- Award records should be **append-only** in the database. Never update or delete historical award records; only write new ones. If a recalculation is needed (e.g., bug fix), mark old records as superseded and insert fresh ones with a recalculation reference.
- The breakdown (list of individual award items) is valuable for transparency: users can see exactly why they received each point. The UI can display: "+3 base move", "+2 relay bonus", "+5 chain bonus", etc.
- The owner's chain share may be fractional (e.g., 6.75). Since this is a monetary-like situation, floor or round-half-up are common strategies. The key is consistency.
- If the same event results in both a base move award AND a relay bonus for the same user, both appear separately in the breakdown but are summed in `total_points`.
- A COMMENT event (if it somehow reaches the aggregator) will produce an empty award list (all awards zeroed or absent). This is correct behavior.
- The aggregator does NOT apply additional filtering or business logic. Any filtering should have been done in earlier modules (03, 04). The aggregator is purely a collector and formatter.
