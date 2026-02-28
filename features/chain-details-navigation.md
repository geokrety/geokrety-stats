# Feature: Chain Details Navigation

**Status:** Complete
**Date Created:** 2026-02-28
**Last Updated:** 2026-02-28

## Overview
Adds chain-focused subpages and navigation paths so users can inspect chain state and switch context from users, GeoKrety, and moves.

## Files Modified/Created
- leaderboard-api/cmd/api/main.go
- leaderboard-api/internal/handlers/chains.go
- leaderboard-api/internal/handlers/users.go
- leaderboard-api/internal/handlers/geokrety.go
- leaderboard-api/internal/models/models.go
- leaderboard-dashboard/src/router/index.js
- leaderboard-dashboard/src/views/ChainDetailView.vue
- leaderboard-dashboard/src/views/UserChainsView.vue
- leaderboard-dashboard/src/views/GeokretChainsView.vue
- leaderboard-dashboard/src/views/MoveChainsView.vue
- leaderboard-dashboard/src/views/HomeView.vue
- leaderboard-dashboard/src/views/GeokretyLeaderboardView.vue
- leaderboard-dashboard/src/views/UserView.vue
- leaderboard-dashboard/src/views/GeokretView.vue

## API Endpoints
- GET /api/v1/chains/:id
  - Returns chain metadata (status, timestamps, points, member count, GK context).
- GET /api/v1/chains/:id/members
  - Returns paginated chain members ordered by position.
- GET /api/v1/chains/:id/moves
  - Returns paginated moves in the chain time window with chain-specific points per move.
- GET /api/v1/users/:id/chains
  - Returns paginated chains where a user is a member.
- GET /api/v1/geokrety/:id/chains
  - Returns paginated chains for a GeoKret.
- GET /api/v1/moves/:id/chains
  - Returns chains linked to a move via user_points_log.chain_id.

## Frontend Components
- src/views/UserChainsView.vue
  - User-centric chain list page.
- src/views/GeokretChainsView.vue
  - GeoKret-centric chain list page.
- src/views/MoveChainsView.vue
  - Move-centric chain list page.
- src/views/ChainDetailView.vue
  - Chain detail with member and move tables.

## Testing Procedures
### API (curl)
```bash
curl -s http://<hostip>:8080/api/v1/users/1/chains | jq .
curl -s http://<hostip>:8080/api/v1/geokrety/1/chains | jq .
curl -s http://<hostip>:8080/api/v1/moves/1/chains | jq .
curl -s http://<hostip>:8080/api/v1/chains/1 | jq .
curl -s http://<hostip>:8080/api/v1/chains/1/members | jq .
curl -s http://<hostip>:8080/api/v1/chains/1/moves | jq .
```

### UI
- Open routes:
  - /users/:id/chains
  - /geokrety/:id/chains
  - /moves/:id/chains
  - /chains/:id
- Verify links from:
  - Home users table (user -> chains)
  - GeoKrety leaderboard table (GK -> chains)
  - User and GeoKret move tables (move -> chains)

## Database
- Reads from:
  - geokrety_stats.gk_chains
  - geokrety_stats.gk_chain_members
  - geokrety_stats.gk_chain_completions
  - geokrety_stats.user_points_log
  - geokrety.gk_moves
  - geokrety.gk_geokrety
  - geokrety.gk_users
- No migration required for this feature.

## Deployment Notes
- Rebuild both services after pulling changes:
  - leaderboard-api
  - leaderboard-dashboard
- Feature depends on existing geokrety_stats chain tables populated by replay/live processing.
