# User + GeoKret Detail Enhancements

## Overview
This feature improves user and GeoKret detail pages with stronger navigation, richer tab content, and consistent UX. It also aligns GeoKret avatar sourcing in `mv_gk_stats` with the `mv_user_stats` approach (`bucket/key` from `gk_pictures`).

## Files modified
- `leaderboard-dashboard/src/views/UserView.vue`
- `leaderboard-dashboard/src/views/GeokretView.vue`
- `leaderboard-dashboard/src/views/CountryDetailView.vue`
- `leaderboard-dashboard/src/components/MoveTypeBreakdown.vue`
- `leaderboard-api/internal/handlers/geokrety.go`
- `leaderboard-api/internal/models/models.go`
- `geokrety-stats/migrations/000018_add_avatar_to_mv_gk_stats.up.sql`
- `geokrety-stats/migrations/000018_add_avatar_to_mv_gk_stats.down.sql`

## API endpoints
- `GET /api/v1/geokrety/:id`
  - now reads avatar fields from `mv_gk_stats.avatar_bucket/avatar_key`
  - now includes `total_comments`
- `GET /api/v1/geokrety/:id/moves`
  - supports sorting by `waypoint`
  - existing filters continue: `awarding_only`, `types`
- `GET /api/v1/geokrety/:id/points/log`
  - supports sorting by `user`, `country`, `waypoint`
  - existing filters continue: `awarding_only`, `types`
- `GET /api/v1/users/:id/geokrety`
  - used by new user tab listing interacted GeoKrety

## Request/response examples
### GeoKret moves with filters/sort
```bash
curl -s "http://192.168.130.65:8080/api/v1/geokrety/11128/moves?sort=author&order=asc&awarding_only=true&types=0,1,3" | jq .
```

### GeoKret points log with filters/sort
```bash
curl -s "http://192.168.130.65:8080/api/v1/geokrety/11128/points/log?sort=user&order=asc&awarding_only=true&types=0,1,3,5" | jq .
```

### User interacted GeoKrety
```bash
curl -s "http://192.168.130.65:8080/api/v1/users/2432/geokrety?page=1&per_page=25" | jq .
```

## Frontend components
- `UserView.vue`
  - adds header awards button
  - widens desktop header stats column
  - adds lazy-loaded `GeoKrety` tab with linked table
- `GeokretView.vue`
  - adds header awards button (`Points Log` focus)
  - desktop header width and in-cache line layout improvements
  - lazy tab loading for all tabs
  - reusable move breakdown section
  - richer `Reach` and `Dates` cards (lifetime/inactive durations)
  - moves/points tables now include avatars, backend sorting headers, awarding-only and type filters
- `MoveTypeBreakdown.vue`
  - shared panel for exact move-type breakdown UX
- `CountryDetailView.vue`
  - now reuses `MoveTypeBreakdown.vue`

## Testing instructions (curl + MCP)
### API (curl)
1. Validate GeoKret detail payload includes avatar URL and `total_comments`:
```bash
curl -s "http://192.168.130.65:8080/api/v1/geokrety/11128" | jq .
```
2. Validate moves table sort/filter combinations:
```bash
curl -s "http://192.168.130.65:8080/api/v1/geokrety/11128/moves?sort=waypoint&order=asc&awarding_only=true&types=0,3,5" | jq .
```
3. Validate points log sort/filter combinations:
```bash
curl -s "http://192.168.130.65:8080/api/v1/geokrety/11128/points/log?sort=country&order=asc&awarding_only=true&types=0,1,3,5" | jq .
```
4. Validate user interacted GeoKrety tab source:
```bash
curl -s "http://192.168.130.65:8080/api/v1/users/2432/geokrety" | jq .
```

### UI (MCP Playwright)
1. Open `http://192.168.130.65:3000/users/2432` and verify:
   - header has `Chains` + `Awards`
   - `GeoKrety` tab exists and loads data on activation
2. Open `http://192.168.130.65:3000/geokrety/11128#overview` and verify:
   - header has `Chains` + `Awards`
   - in-cache state appears on its own line
   - move breakdown panel matches country page style
   - reach/dates cards show lifetime/inactive durations
3. Open `#moves` and `#points` tabs and verify:
   - author avatars render
   - all header sorts are clickable and reflected in network calls
   - filters (`Only awarding points`, type multiselect) apply correctly
4. Capture screenshots at `1280x1024` and `720x2048`.

## Database notes
- Added migration `000018_add_avatar_to_mv_gk_stats`.
- `mv_gk_stats` now sources avatar from `gk_pictures` (`avatar_bucket` + `avatar_key`) and exposes `total_comments`.
- Down migration restores pre-avatar-column shape.

## Deployment notes
1. Rebuild and restart services via Docker Compose.
2. Confirm migrations are applied in API/stats startup logs.
3. Validate target pages and endpoint responses.

## Known limitations
- GeoKret awards access is exposed in-page via `Points Log` tab activation rather than a dedicated `/geokrety/:id/awards` route.
