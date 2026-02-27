# Module 05 – Country Crossing Bonus

## Responsibility

Detects when a GeoKret enters a country it has never visited before (excluding its home country) and awards points to both the actor and the GK owner. Also sets the `new_country_visited` runtime flag for use by module 12 (diversity bonus tracker).

---

## Input

From context:
- `event.log_type`
- `event.country` – ISO country code of this event (may be null)
- `event.user_id` (actor)
- `event.logged_at`
- `gk_state.owner_id`
- `gk_state.gk_type`
- `gk_state.countries_visited` – set of all countries this GK has been to before this event
- `gk_state.home_country` – the GK's first country (no bonus generated for home country)

---

## Process

### Step 1 – Check Move Type Eligibility

Country bonuses only apply when the GK physically arrives somewhere. Eligible log types:

```
if event.log_type not in {DROP (0), DIP (5), SEEN (3)} → SKIP
```

GRAB does not carry a "destination country" in the same sense; it is the act of picking up. The country bonus is earned when the GK is **placed** or **observed** at a location.

### Step 2 – Check Country Data

```
if event.country is null → SKIP (no geographic data for this event)
```

### Step 3 – Check Home Country

```
if event.country == gk_state.home_country → SKIP (no bonus for home country)
```

The home country is the GK's baseline; it is exempt from bonuses to avoid rewarding the creation drop itself.

### Step 4 – Check if Country is New

```
if event.country IN gk_state.countries_visited → SKIP (already visited, one-time reward only)
```

The bonus is awarded globally once per GK per country, regardless of who carried it there. No repeat bonuses for revisiting the same country.

### Step 5 – New Country Confirmed

Set the runtime flag for downstream modules:
```
context.runtime_flags.new_country_visited = true
context.runtime_flags.new_country_code = event.country
```

Record the country as visited in the data store (side effect):
```
→ Add event.country to gk's countries_visited set
```

### Step 6 – Determine Actor Award

The actor (who moved the GK to this new country) always earns a points bonus, with one exception based on GK type when the actor is also the owner:

**Standard GK (types 0–7):**
```
if event.user_id == gk_state.owner_id:
    actor_award = 2    ← owner moving their own standard GK to new country
else:
    actor_award = 3    ← non-owner moving GK to new country
```

**Non-Transferable GK (types 8–10):**
```
if event.user_id == gk_state.owner_id:
    actor_award = 4    ← owner moving their own non-transferable GK to new country
                         (once per user/country/gk_type combination, all time)
else:
    actor_award = 3    ← non-owner moving non-transferable GK to new country
```

For non-transferable GK where actor is the owner and award is +4: this is a one-time award per (actor, country, gk_type) for all time. Check if the owner has already received a +4 non-transferable country bonus for the same (country, gk_type) combination. If yes, skip the award.

### Step 7 – Award Actor Points

```
emit award:
{
  recipient_user_id : event.user_id,
  points            : actor_award,
  reason            : "GK #<gk_id> reached new country <country> (actor bonus)",
  module_source     : "05_country_crossing",
  label             : "country_crossing_actor",
  is_owner_reward   : false
}
```

### Step 8 – Award Owner Points (when actor is NOT the owner)

When a non-owner moves the GK to a new country, the GK owner earns +3 regardless of GK type. The +4 non-transferable bonus (step 6) is exclusively for when the **owner themselves** moves their non-transferable GK to a new country.

```
if event.user_id != gk_state.owner_id:
    emit award:
    {
      recipient_user_id : gk_state.owner_id,
      points            : 3,
      reason            : "GK #<gk_id> reached new country <country> (owner bonus)",
      module_source     : "05_country_crossing",
      label             : "country_crossing_owner",
      is_owner_reward   : true
    }
```

When the actor IS the owner (step 6 applies), no separate owner award is emitted here. The owner's award is already captured in the actor award (either +2 for standard GK or +4 for non-transferable GK).

---

## Output

### Awards Added to Accumulator

Up to two awards per event:
1. Actor country crossing award (+2, +3, or +4 depending on conditions)
2. Owner country bonus (+3, only if actor is not the owner)

### Runtime Flags Written to Context

```
context.runtime_flags.new_country_visited = true  (if new country detected)
context.runtime_flags.new_country_code = "<ISO>"   (if new country detected)
```

If no new country was detected, these flags remain false/null for module 12.

---

## Notes

- Country detection is based on the move's country field. If the same GK visits country X twice (two different events), only the first triggers the bonus. The `countries_visited` set in module 01 loads the pre-event set, and this module's side effect adds the new country to it.
- The home country rule prevents the initial creation drop (which naturally has a country) from generating a bonus. The creator does not need to earn points for simply placing their new GK in their own country.
- For standard GK owners moving to new countries: the owner gets a reduced +2 (not +3) because the rule specifically says "If actor is GK owner → owner receives +2". This encourages distributing GKs internationally via others rather than personally transporting them.
- The diversity country bonus (+5 for actor reaching a new country per month) is handled separately in module 12 and stacks on top of these awards.
