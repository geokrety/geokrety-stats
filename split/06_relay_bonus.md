# Module 06 – Relay Bonus

## Responsibility

Awards a time-based circulation bonus when a GeoKret is quickly picked up by a new user shortly after being dropped. Rewards both the person who picks it up (mover) and the person who dropped it (previous dropper), incentivizing fast turnover.

---

## Input

From context:
- `event.log_type`
- `event.user_id` (actor = the new mover)
- `event.logged_at`
- `gk_history.last_drop_at` – UTC timestamp of the most recent DROP on this GK
- `gk_history.last_drop_user` – user_id who made that DROP
- `gk_state.current_holder` – who currently holds the GK (before this event)

---

## Process

### Step 1 – Check Move Type

The relay bonus only triggers on a GRAB event. The actor is picking up the GK.

```
if event.log_type != GRAB (1) → SKIP
```

### Step 2 – Check That a Previous Drop Exists

```
if gk_history.last_drop_at is null → SKIP (GK has never been dropped; no dropper to reward)
```

### Step 3 – Check That GK Was in Cache (Not in Someone's Hands)

The relay bonus is about cache-to-cache rapid circulation. The GK must be in a cache when it is grabbed (not being grabbed directly from another person's inventory).

```
if gk_state.current_holder is not null → SKIP
  (GK is currently held by a user, not sitting in a cache;
   this is a person-to-person grab, not a relay from a cache)
```

### Step 4 – Check That the Mover is Different from the Previous Dropper

Relay bonus should not be awarded if the same person drops and immediately grabs their own GK back.

```
if event.user_id == gk_history.last_drop_user → SKIP
```

### Step 5 – Check the 7-Day Window

The GRAB must occur within 7 days (168 hours) of the last DROP.

```
days_since_drop = (event.logged_at - gk_history.last_drop_at) in days

if days_since_drop > 7 → SKIP
```

### Step 6 – Award Relay Bonus to Mover

```
emit award:
{
  recipient_user_id : event.user_id,
  points            : 2,
  reason            : "Relay bonus: GK #<gk_id> grabbed within 7 days of last drop (mover)",
  module_source     : "06_relay_bonus",
  label             : "relay_mover",
  is_owner_reward   : false
}
```

### Step 7 – Award Relay Bonus to Previous Dropper

```
emit award:
{
  recipient_user_id : gk_history.last_drop_user,
  points            : 1,
  reason            : "Relay bonus: GK #<gk_id> grabbed within 7 days (previous dropper)",
  module_source     : "06_relay_bonus",
  label             : "relay_dropper",
  is_owner_reward   : false
}
```

---

## Output

### Awards Added to Accumulator

When triggered: up to 2 awards:
1. +2 to the actor (the new grabber)
2. +1 to the previous dropper

When not triggered: nothing added.

---

## Notes

- The previous dropper (`gk_history.last_drop_user`) may be the GK owner or a non-owner. The +1 dropper bonus is awarded regardless of whether the dropper is the owner. Owners normally earn 0 base points, but event-triggered bonuses like this relay reward are not base points – they are separate bonus lines.
- There is no cap on how many relay bonuses a user can earn per month or per GK. Each qualifying event triggers its own relay bonus independently.
- If the GK was dropped and grabbed multiple times within 7-day windows in quick succession, each qualifying grab triggers its own relay bonus.
- The mover also gets the standard base +3 (from module 02) if this is their first GRAB on this GK. The relay +2 stacks additively on top of that.
- The previous dropper's +1 is purely additive and does not depend on whether the dropper earned any points from their original drop event.
