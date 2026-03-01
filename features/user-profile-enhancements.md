# User Profile Enhancements

## Overview
Fixes and improvements for user profile and awards pages:
- awards page runtime crash (`availableLabels` undefined)
- user header stat/display alignment and joined year display
- lazy loading of tab content on user page
- sortable/filterable user moves table
- avatar support for user and GeoKret pages
- awards header quick link to chains

## Files Modified
- `leaderboard-dashboard/src/views/UserView.vue`
- `leaderboard-dashboard/src/views/PointAwardsView.vue`
- `leaderboard-dashboard/src/views/GeokretView.vue`
- `leaderboard-dashboard/src/composables/useAvatarUrl.js`
- `leaderboard-api/internal/handlers/users.go`
- `leaderboard-api/internal/handlers/geokrety.go`
- `leaderboard-api/internal/models/models.go`
- `geokrety-stats/migrations/000017_add_avatar_to_mv_user_stats.up.sql`
- `geokrety-stats/migrations/000017_add_avatar_to_mv_user_stats.down.sql`

## API Changes
### `GET /api/v1/users/:id`
- Added `avatar` field in response.

### `GET /api/v1/users/:id/moves`
- Added query params:
  - `sort`: `date|gk|type|country|points`
  - `order`: `asc|desc`
  - `awarding_only`: `true|false`
  - `types`: comma-separated move types (`0..5`)
- Added `gk_avatar` in move rows.
- Points are now calculated from current user awards only.

### `GET /api/v1/geokrety/:id`
- Added `avatar` field in response.

## Database Notes
Added migration to rebuild `geokrety_stats.mv_user_stats` with `avatar` from `geokrety.gk_users`:
- `000017_add_avatar_to_mv_user_stats.up.sql`
- `000017_add_avatar_to_mv_user_stats.down.sql`

## UI Notes
- User header now displays: points, rank, moves, geokrety, countries, avg/move
- Joined date now displays year only
- `Points per Day` card no longer shows `since YYYY-MM-DD`
- Moves tab supports full column sorting and additional filters
- Moves points column moved to far right
- User and GeoKret pages render avatars when available

## Testing
### API (curl)
- `curl -s 'http://<hostip>:8080/api/v1/users/2432' | jq .`
- `curl -s 'http://<hostip>:8080/api/v1/users/2432/moves?sort=points&order=desc&awarding_only=true&types=0,1,3' | jq .`
- `curl -s 'http://<hostip>:8080/api/v1/geokrety/<id>' | jq .`

### UI
- `/users/2432`
- `/users/2432#moves`
- `/users/2432/awards`
- `/geokrety/<id>`

Validate desktop/mobile layouts, sorting controls, filter behavior, and avatar rendering.
