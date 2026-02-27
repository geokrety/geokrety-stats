# Module 00 – Event Guard

## Responsibility

First gate in the pipeline. Determines whether the incoming log event is eligible for any point calculation. If the event is rejected here, the entire pipeline stops immediately and no points are awarded.

This module exists to avoid wasted computation and to enforce the rule that only certain events for authenticated users can ever produce points.

---

## Input

- `event.user_id` – the user who logged this action (null if anonymous)
- `event.log_type` – the type of log action recorded

---

## Process

Evaluate each condition in order. If any condition is not met, **halt the pipeline** with reason noted.

### Condition 1 – Authenticated User Only

```
if event.user_id is null → REJECT ("Anonymous moves earn 0 points")
```

Anonymous moves (user not logged in, or user account later deleted/banned) do not score any points, period.

### Condition 2 – Scoreable Log Type Only

Only these log types may ever produce points:

| log_type | name     | Scoreable? |
|----------|----------|------------|
| 0        | DROP     | ✅ yes     |
| 1        | GRAB     | ✅ yes     |
| 2        | COMMENT  | ❌ no      |
| 3        | SEEN     | ✅ yes     |
| 4        | ARCHIVED | ✅ yes (special: ends chain, may trigger chain bonus; no base points) |
| 5        | DIP      | ✅ yes (special: chain timer extension; no base points) |

```
if event.log_type not in {0, 1, 3, 4, 5} → REJECT ("Log type is not scoreable")
```

COMMENT (type 2) is entirely non-scoreable and is rejected here. No module downstream ever awards points for comments.

ARCHIVED (type 4) is passed through so that module 10 (chain state manager) can detect and end an active chain. No base points are awarded for ARCHIVED events.

DIP (type 5) is passed through so that module 10 can extend the chain timer. No base points are awarded for DIPs.

### Condition 3 – No Duplicate Processing

```
if this log_id has already been processed → REJECT ("Duplicate event, already scored")
```

Each log entry must be scored exactly once. If the pipeline receives the same log_id twice (e.g., due to a retry or requeue), it is rejected.

---

## Output

- **PASS** – event flows to module 01 with no awards added
- **HALT** – pipeline terminated, accumulator remains empty, reason is logged for observability

---

## Notes

- This module does **not** load any GK or user state; it operates only on the raw event fields.
- The rejection of COMMENT and ARCHIVED for point purposes does not mean those logs are invalid; they are stored as usual in the log history. The multiplier update (module 13) and chain manager (module 10) still need ARCHIVED to function – but COMMENT is fully ignored by every module.
- User deletion/banning retroactively sets logs to `user_id = null`. If the pipeline is ever re-run for historical recalculation, those entries will be rejected here.
