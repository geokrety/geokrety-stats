# Feature: User Profile Enhancements

**Status:** In Development
**Date Created:** 2026-03-01
**Last Updated:** 2026-03-01

## Overview
- Stabilize the awards page by guarding against the previously missing `availableLabels` data, adjust the user header layout, and display the joined year instead of the full timestamp.
- Lazy-load each tab on the user profile so that the moves table, awards list, and related content only fetch data when the tab becomes active.
- Make the user moves table fully sortable (date, GeoKret, type, country, points) and filterable (move types, awarding-only) while exposing the current user's move-point total.
- Surface avatars everywhere (users, GeoKrety, move rows and point awards) by building URLs from `gk_pictures` bucket/key pairs and reusing a shared composable.
- Normalize the move-type badges so drops/gras, dips, seen/comment, and archival states follow an ordered palette that matches the chart colors.

## Files Modified/Created
- `leaderboard-dashboard/src/views/UserView.vue`
- `leaderboard-dashboard/src/views/PointAwardsView.vue`
- `leaderboard-dashboard/src/views/GeokretView.vue`
- `leaderboard-dashboard/src/composables/useAvatarUrl.js`
- `leaderboard-dashboard/src/composables/useMoveTypeColors.js`
- `leaderboard-api/internal/handlers/users.go`
- `leaderboard-api/internal/handlers/geokrety.go`
- `leaderboard-api/internal/models/models.go`
- `geokrety-stats/migrations/000017_add_avatar_to_mv_user_stats.up.sql`
- `geokrety-stats/migrations/000017_add_avatar_to_mv_user_stats.down.sql`

## API Endpoints
- `GET /api/v1/users/:id` now returns an `avatar` URL built from the materialized view’s `avatar_bucket`/`avatar_key`. This payload continues to include hourly aggregates (points timelines, diversity stats) and exposes `home_country`, `rank`, and move breakdowns.
- `GET /api/v1/users/:id/moves` supports `sort=date|gk|type|country|points`, `order=asc|desc`, `awarding_only=true|false`, and `types=0,1,3`. Each row now carries `gk_avatar`, move-type names, points contributed this month, and the new `author_avatar` value so the dashboard can render badges with hover tooltips.
- `GET /api/v1/geokrety/:id`/`points/log` leverage the same avatar helper, expose `avatar` for the GeoKret, and include author avatars in the points timeline alongside ordering/filtering by move type and awarding-only rows.

## Frontend Components
- `UserView.vue` renders the refreshed header (points, rank, moves, GeoKrety, countries, avg/move) and defers data loading for the moves and awards tabs. Move rows now display badges through `getMoveTypeBadgeClass`, show avatars via `userAvatarUrl`, and expose newly introduced sorting/filter toggles.
- `PointAwardsView.vue` consumes the enriched API payload, aligns the awards header with chains, and now safely handles missing labels by defaulting to an empty list.
- `GeokretView.vue` pulls the new `avatar` field when available, uses the shared `useAvatarUrl` helper, and shows the author avatar inside the move list.
- `useAvatarUrl.js` now builds URLs from the bucket/key pair, maps the base bucket to the thumbnail counterpart (`users-avatars-thumbnails`, `gk-avatars-thumbnails`), and falls back to CDN icons when metadata is missing.
- `useMoveTypeColors.js` maps each move/action label to the fresh badge palette (green drops, yellow grabs, blue dips, grey seen/comments, dark archive) while preserving tooltip text for clarity.

## Testing Procedures
### API (curl)
- `curl -s 'http://127.0.0.1:8080/api/v1/users/2432' | jq .`
- `curl -s 'http://127.0.0.1:8080/api/v1/users/2432/moves?sort=points&order=desc&awarding_only=true&types=0,1,3' | jq .`
- `curl -s 'http://127.0.0.1:8080/api/v1/geokrety/1001' | jq .`

### UI (MCP Playwright)
1. Load the browser tools via `tool_search_tool_regex` with pattern `^mcp_microsoft_pla_browser`.
2. Navigate to `http://127.0.0.1:3000/users/2432`.
3. Resize to 720x2048 (mobile) and 1280x1024 (desktop), capturing screenshots for each.
4. Verify the user header stays aligned, move rows show avatars + badges, sorting/filter controls work, and awards fail safely if data is missing.
5. Navigate to `http://127.0.0.1:3000/geokrety/1001`, resize, capture, and confirm the GeoKret avatar + move tab avatars load correctly.

## Database
- `geokrety_stats.mv_user_stats` now joins `geokrety.gk_pictures` to expose `avatar_bucket` and `avatar_key`; the new columns fuel `fetchUser` so the handler can derive the avatar URL without joining the picture table on each request.

## Deployment Notes
- Run `./bin/geokrety-stats -migration-up` to refresh the materialized view after editing the migration.
- Rebuild and restart the services with `docker compose down && docker compose build && docker compose up -d` so both the API and dashboard consume the latest avatar helper and move-type palette.
