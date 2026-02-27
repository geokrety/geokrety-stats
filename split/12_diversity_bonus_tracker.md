# Module 12 – Diversity Bonus Tracker

## Responsibility

Tracks the actor's monthly diversity milestones and awards bonuses when those milestones are crossed for the first time in the current calendar month. Diversity bonuses reset monthly (UTC midnight on the 1st) and are independent of GK or chain-specific events.

This module runs after country crossing (module 05) because one of the diversity bonuses depends on knowing whether this move visited a new country.

---

## Input

From context:
- `event.log_type`
- `event.user_id` (actor)
- `event.logged_at`
- `gk_state.owner_id`
- `gk_state.gk_id`

Runtime flags (set by earlier modules):
- `runtime_flags.actor_scored_this_gk` – true if actor earned base points this event
- `runtime_flags.new_country_visited` – true if this move triggered a new country for the GK
- `runtime_flags.new_country_code` – ISO code of the new country (if any)

Monthly state (from module 01):
- `user_state.actor_gks_dropped_this_month` – distinct GKs actor has dropped this month (before this event)
- `user_state.actor_distinct_owners_this_month` – distinct GK owners with whom actor scored this month (before this event)
- `user_state.actor_countries_visited_this_month` – set of countries for which actor already received the diversity country bonus this month

---

## Process

### Diversity Bonus A – "Drop 5 Different GKs This Month" (+3 points)

#### Step A1 – Check Move Type

```
if event.log_type != DROP (0) → skip this bonus
```

#### Step A2 – Check That Points Were Earned (optional: dropping earns points, implying it was a scored drop)

The rule says "Drop 5 different GKs". Whether this counts only scored drops or any drop is open to interpretation, but for anti-gaming purposes: only scored drops (where actor received base points) count.

```
if runtime_flags.actor_scored_this_gk == false → skip this bonus
```

#### Step A3 – Increment Drop Counter

The actor just made a scored drop on a distinct GK. Add this GK to the monthly drop set if not already present.

```
if gk_state.gk_id NOT already counted in actor's drops this month:
    new_drop_count = actor_gks_dropped_this_month + 1
else:
    new_drop_count = actor_gks_dropped_this_month  (already counted)
    skip this bonus (milestone only fires on new additions)
```

#### Step A4 – Check Milestone

```
if actor_gks_dropped_this_month < 5 AND new_drop_count >= 5:
    → Milestone just crossed (exactly the 5th distinct drop this month)
    → Award bonus
```

```
emit award:
{
  recipient_user_id : event.user_id,
  points            : 3,
  reason            : "Diversity bonus: dropped 5 different GKs this month",
  module_source     : "12_diversity_bonus_tracker",
  label             : "diversity_5drops",
  is_owner_reward   : false
}
```

---

### Diversity Bonus B – "Interact with 10 Different GK Owners This Month" (+7 points)

#### Step B1 – Check That Points Were Earned

Any scored move (DROP, GRAB, SEEN) counts as an "interaction":

```
if runtime_flags.actor_scored_this_gk == false → skip this bonus
```

#### Step B2 – Check If This Owner Is New For This Month

```
if gk_state.owner_id already counted in actor's distinct owners this month:
    skip this bonus (owner already counted, no new increment)
```

#### Step B3 – Increment Owner Count

```
new_owner_count = actor_distinct_owners_this_month + 1
```

#### Step B4 – Check Milestone

```
if actor_distinct_owners_this_month < 10 AND new_owner_count >= 10:
    → Milestone just crossed (exactly the 10th distinct owner interaction this month)
    → Award bonus
```

```
emit award:
{
  recipient_user_id : event.user_id,
  points            : 7,
  reason            : "Diversity bonus: interacted with 10 different GK owners this month",
  module_source     : "12_diversity_bonus_tracker",
  label             : "diversity_10owners",
  is_owner_reward   : false
}
```

---

### Diversity Bonus C – "Get a GK to Visit a New Country This Month" (+5 points)

This bonus fires when the actor's move causes a GK to enter a new country AND this is the first time this month the actor has earned this bonus for this specific country.

#### Step C1 – Check New Country Flag

```
if runtime_flags.new_country_visited == false → skip this bonus
  (this move did not trigger a new-country crossing for this GK)
```

#### Step C2 – Check Monthly Per-Country Deduplication

```
country = runtime_flags.new_country_code

if country IN user_state.actor_countries_visited_this_month → skip this bonus
  (actor already earned the diversity country bonus for this country this month)
```

#### Step C3 – Award Diversity Country Bonus

```
emit award:
{
  recipient_user_id : event.user_id,
  points            : 5,
  reason            : "Diversity country bonus: actor brought a GK to new country
                       <country> this month (first time this month for this country)",
  module_source     : "12_diversity_bonus_tracker",
  label             : "diversity_country",
  is_owner_reward   : false
}
```

Side effect: Record `country` in `actor_countries_visited_this_month` so it cannot be earned again this month.

---

## Cumulative Country Bonus Example (from rules)

When an actor moves GK to a new country AND it is the first time this month they earned the country diversity bonus for that country, both stack:

- Module 05 awards: +3 (country_crossing_actor)
- Module 12 awards: +5 (diversity_country)
- Total: +8 points for bringing a GK to a new country for the first time this month

If the actor already earned the diversity bonus for that country this month (via a DIFFERENT GK), they still earn the +3 from module 05 but NOT the +5 from this module.

---

## Output

### Awards Added to Accumulator

Up to 3 awards per event (one for each diversity bonus milestone that is newly crossed):
- +3 for 5th drop milestone
- +7 for 10th owner interaction milestone
- +5 for new country (first time this month) milestone

---

## Notes

- All three diversity bonuses reset at UTC midnight on the 1st of every calendar month. There is no roll-over.
- Each bonus fires exactly once per month per actor (the milestone check ensures it only fires when the counter crosses the threshold, not on every event after it). Once the 5-drop bonus fires in January, no further +3 drop bonuses fire in January regardless of how many more GKs are dropped.
- The interaction count for owner diversity (Bonus B) counts distinct OWNERS, not distinct GKs. A user who moves 5 GKs from the same owner only counts that owner once.
- The country diversity bonus (Bonus C) is per-country per-month. An actor can earn it up to (number of new countries reached by their moves) times per month, each for a different country. The limitation is one award per (actor, country, month), not one award per month total.
