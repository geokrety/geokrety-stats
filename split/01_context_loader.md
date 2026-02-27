# Module 01 – Context Loader

## Responsibility

Loads all persistent state required by downstream modules in one place. No other module should independently query the data store. By centralizing data access here, the rest of the pipeline operates on a consistent, point-in-time snapshot.

---

## Input

- `event.user_id`  – actor performing the action
- `event.gk_id`    – the GeoKret being acted upon
- `event.log_type` – for context-specific loading decisions
- `event.waypoint` – cache identifier (may be null for grab-from-user)
- `event.country`  – ISO country code of this move (may be null for grabs)
- `event.logged_at` – UTC timestamp

---

## Process

Load the following state groups. All timestamps are UTC.

### 1 – GK Core State

From the GeoKret record:
- `gk_state.gk_type` – integer 0–10 (0–7 = standard; 8–10 = non-transferable)
- `gk_state.owner_id` – the GK's owner user_id
- `gk_state.created_at` – creation timestamp of the GK
- `gk_state.current_multiplier` – the multiplier value **before this event** (module 13 will update it after all scoring)

### 2 – GK Holder State

From the most recent non-COMMENT log on the GK, determine:
- `gk_state.current_holder` – user currently holding the GK (null = in cache)
- `gk_state.previous_holder` – user who held it immediately before current (null if was also in cache)

"In cache" means the GK was last placed in a cache by DROP, SEEN (while unattended), or similar; `current_holder = null`.

### 3 – GK Geographic History

- `gk_state.countries_visited` – full set of ISO country codes the GK has passed through (from all prior DROP/DIP/SEEN moves that had a country). Used to detect new-country crossings.
- `gk_state.home_country` – first country recorded for this GK (first DROP/SEEN/DIP that had a country). This country never generates country bonuses.

### 4 – GK Last-Activity Timestamps

These are needed by relay, rescuer, and multiplier modules:
- `gk_history.last_drop_at` – UTC timestamp of the most recent DROP log on this GK
- `gk_history.last_drop_user` – user_id who made that DROP
- `gk_history.last_cache_entry_at` – UTC timestamp of the most recent event that put the GK into a cache (DROP or SEEN while holder was null). Used for dormancy detection.
- `gk_state.last_multiplier_update_at` – UTC timestamp of the last time the GK multiplier was updated. Used by module 13 to compute elapsed time for time-decay. Must be loaded alongside `current_multiplier`.

### 5 – GK Distinct Users (6-Month Window)

- `gk_history.distinct_users_6m` – list of distinct user_ids who have recorded a scored move on this GK in the 6 months prior to `event.logged_at`.

Used by module 09 (reach bonus) to detect the 10-users milestone.

### 6 – Actor Move History on This GK

- `user_state.actor_move_history_on_gk` – the set of log_types (DROP, GRAB, SEEN, DIP) the acting user has previously logged on this specific GK.

Used by module 02 to determine first-move eligibility and by module 13 for multiplier first-move tracking.

### 7 – Actor Owner-GK Interaction Count

- `user_state.actor_gks_per_owner_count` – count of distinct GK IDs owned by `gk_state.owner_id` on which the actor has ever earned base points (globally, all time, not per month).

Used by module 03 (owner GK limit filter).

### 8 – Actor Waypoint Activity This Month

- `user_state.actor_gks_at_waypoint_this_month` – count of distinct GK IDs the actor has moved at `event.waypoint` during the current calendar month (UTC).

Used by module 04 (waypoint penalty). Load as 0 if `event.waypoint` is null.

### 9 – Actor Monthly Diversity State

Load current calendar month counters for the actor:
- `user_state.actor_countries_visited_this_month` – set of country codes for which the actor has already received the diversity country bonus this month
- `user_state.actor_gks_dropped_this_month` – count of distinct GKs the actor has dropped this month (for the 5-drops diversity bonus)
- `user_state.actor_distinct_owners_this_month` – count of distinct GK owner user_ids the actor has had a scored interaction with this month (for the 10-owners diversity bonus)

Used by module 12 (diversity bonus tracker).

### 10 – Active Chain State

- `chain_state.active_chain_id` – ID of any currently active chain on this GK (null if none)
- `chain_state.chain_members` – ordered list of distinct user_ids in the chain so far
- `chain_state.chain_last_active` – UTC timestamp of last event that reset or extended the chain timer
- `chain_state.holder_acquired_at` – UTC timestamp when the current holder took possession of the GK (for DIP extension calculation)

---

## Output

A fully populated context object (as described in README.md) is attached to the pipeline.
No entries are added to the awards accumulator.

---

## Notes

- All data loaded here must reflect the **state just before this event is processed**. This is critical for the multiplier (must use old value) and for the distinct-users list (must not yet include this event's actor).
- If any critical data is missing (e.g., GK does not exist, owner not found), the pipeline should halt with an error. This is a data integrity failure, not a scoring decision.
- The `event.country` field may be null for GRAB events (grabbing from a user who is travelling may have no fixed location). In that case, country crossing bonus (module 05) will simply be skipped.
