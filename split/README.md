# Points System – Module Architecture

## Overview

Every point calculation is triggered by a **log event**: a user action on a GeoKret (drop, grab, seen, dip, comment, archive). The event flows through a **linear chain of modules**, each responsible for a single, isolated concern.

No module knows about the others. Each module receives the same shared **context** (event + pre-loaded state) and an **accumulator** (the growing list of point awards). A module may:
- **Append** new awards to the accumulator
- **Modify** existing awards already in the accumulator (apply scaling or zeroing)
- **Update** mutable fields in the context (e.g., flag that a new country was visited, so downstream modules can use it)
- **Halt** the pipeline entirely (only the event guard does this)

The final output of the pipeline is a structured list of point awards, each with a recipient, an amount, and a human-readable reason.

---

## Context Object

The context object is assembled by module 01 and flows read-mostly through the rest of the pipeline. Key fields:

```
event:
  log_id          – unique ID of this log entry
  user_id         – user performing the action (null = anonymous)
  gk_id           – geokret being moved
  log_type        – DROP(0), GRAB(1), COMMENT(2), SEEN(3), ARCHIVED(4), DIP(5)
  waypoint        – cache/place identifier (may be null)
  country         – ISO country code of this move (may be null)
  logged_at       – UTC timestamp of this log

gk_state:
  gk_type         – integer 0-10 (0-7 = standard, 8-10 = non-transferable)
  owner_id        – user_id of the GeoKret owner
  created_at      – UTC timestamp of GK creation
  current_multiplier – float, current multiplier value BEFORE this event
  current_holder  – user_id currently holding the GK (null = in cache)
  previous_holder – user_id who held it before (null = was in cache)
  countries_visited – set of ISO country codes GK has visited (including home country)
  home_country    – first country the GK was ever seen in

gk_history:
  last_drop_at    – UTC timestamp of last DROP log
  last_drop_user  – user_id who last dropped it
  last_seen_at    – UTC timestamp of last SEEN log (while in cache)
  last_cache_entry_at – UTC timestamp of last time GK was placed in cache (DROP/SEEN)
  distinct_users_6m   – list of distinct user_ids who moved this GK in the last 6 months

user_state:
  actor_move_history_on_gk – set of log_types the actor has previously logged on THIS gk
                              (used for first-move tracking)
  actor_gks_per_owner_count – count of distinct GKs from this GK's owner that the actor
                               has already earned points from (globally, all time)
  actor_gks_at_waypoint_this_month – count of distinct GKs the actor has moved at THIS
                                      waypoint in the current calendar month
  actor_countries_visited_this_month – set of countries actor has already gotten the
                                        diversity country bonus for this month
  actor_gks_dropped_this_month       – count of distinct GKs the actor has dropped this month
  actor_distinct_owners_this_month   – count of distinct GK owners the actor has interacted
                                        with (scored move) this month

chain_state:
  active_chain_id   – ID of the current active chain (null if no chain)
  chain_members     – ordered list of distinct user_ids in the chain
  chain_last_active – UTC timestamp of last chain-extending activity
  holder_acquired_at – UTC timestamp when current holder acquired the GK (for DIP extension)
  chain_ended       – boolean flag set to true by module 10 if the chain ends this event

runtime_flags (mutable, written by modules, read by later modules):
  new_country_visited  – boolean: true if this move entered a NEW country for this GK
  base_points_awarded  – float: base points awarded to the actor (set by module 02)
  base_points_label    – string: label used to tag base award entries for penalty application
  actor_scored_this_gk – boolean: true if actor earned any base points this event
```

---

## Pipeline Execution Order

| Step | Module File                   | Concern                                              |
|------|-------------------------------|------------------------------------------------------|
| 00   | `00_event_guard.md`           | Reject non-scoreable events before any work          |
| 01   | `01_context_loader.md`        | Load all state needed by downstream modules          |
| 02   | `02_base_move_points.md`      | Compute raw base points for the actor                |
| 03   | `03_owner_gk_limit_filter.md` | Zero base points if actor exceeded owner-GK limit    |
| 04   | `04_waypoint_penalty.md`      | Scale base points by multi-GK waypoint penalty       |
| 05   | `05_country_crossing.md`      | Country bonus to actor + owner; set new_country flag |
| 06   | `06_relay_bonus.md`           | Fast-circulation relay bonus to mover + dropper      |
| 07   | `07_rescuer_bonus.md`         | Dormancy rescue bonus to grabber + owner             |
| 08   | `08_handover_bonus.md`        | Handover bonus to owner when GK changes hands        |
| 09   | `09_reach_bonus.md`           | Reach milestone bonus to owner (10 users/6 months)   |
| 10   | `10_chain_state_manager.md`   | Update chain; detect chain end; set chain_ended flag |
| 11   | `11_chain_bonus.md`           | Award chain completion bonus if chain just ended     |
| 12   | `12_diversity_bonus_tracker.md` | Monthly diversity milestone bonuses to actor       |
| 13   | `13_gk_multiplier_updater.md` | Update GK multiplier (runs AFTER all scoring)        |
| 14   | `14_points_aggregator.md`     | Collect, validate, and emit final award list         |

---

## Award Record Format

Each award emitted into the accumulator has the following shape:

```
{
  recipient_user_id : integer   – who receives the points
  points            : float     – amount (may be fractional; aggregator rounds)
  reason            : string    – human-readable description
  module_source     : string    – which module generated this award (for debugging)
  label             : string    – internal tag for pipeline targeting (e.g. "base_move")
  is_owner_reward   : boolean   – true if this is an owner-specific bonus
}
```

Negative points are never emitted.  A module may reduce an existing award to 0 (e.g., penalty, limit), but never to negative.

---

## Special Cases: Chain Timeout

The chain state manager (module 10) handles **inline** chain ending triggered by the ARCHIVE event or by a DIP/GRAB that arrives after the 14-day window has lapsed. However, chains can also end **asynchronously** due to pure inactivity (no event fires on the GK). To handle this, a background **chain timeout checker** must exist as a standalone job that:
1. Periodically scans active chains for those where `now - chain_last_active > 14 days`
2. Calls module 11 directly with the chain state to compute and award the bonuses

This is documented in `10_chain_state_manager.md` and `11_chain_bonus.md`.
