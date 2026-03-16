---
description: 'Instructions for building the GeoKrety Stats API, including OpenAPI design, WebSocket implementation, and operational endpoints.'
applyTo: 'api/**/*.go'
---

# GeoKrety Stats API Instructions

- MUST Use Go with `chi` for routing and `pgx` + `sqlx`
- MUST Be compliant with JSON REST API and include metadata, relationships, filtering and pagination if needed
- MUST support GeoJSON for location data and include relevant fields in responses (leaflet integration)
- MUST be optimized for performance on datasets used for Data Visualisation (ECharts) and mapping (Leaflet.js)
- MUST Expose REST endpoints for stats, GeoKrety, users, and countries
- MUST Implement WebSocket for live updates and connection count
- MUST Serve OpenAPI spec and Swagger UI for documentation
- MUST Use structured logging with `zap` and respect `LOG_LEVEL`
- MUST Implement graceful shutdown and CORS policies for security
- MUST Use prepared statements for all DB queries
- MUST sorting, filtering, and pagination for all list endpoints
- MUST Consistent response formats with error handling
- MUST Broadcast WebSocket messages on relevant data changes and periodically broadcast connection count
- MUST Document all endpoints and expected responses in OpenAPI spec
- MUST Validate input parameters and return appropriate error messages for invalid requests
- MUST Ensure all endpoints are idempotent and safe to call multiple times
- MUST Consider response caching for expensive endpoints and implement rate limiting if necessary
- MUST Ensure proper error handling and status codes for all endpoints (e.g. 400 for bad requests, 500 for server errors)
- MUST include country codes and names in relevant responses for better frontend integration
- MUST Include relevant metadata in responses such as total counts, pagination info, and timestamps for better frontend integration
- MUST include request/query timing statistics in response metadata for monitoring and debugging purposes
- MUST expose only stable view-like contracts, not internal maintenance tables
- MUST represent dates as UTC ISO 8601 strings
- MUST document freshness for materialized-view backed endpoints as part of the contract
- MUST include `data_as_of` and `computed_at` fields in responses
- MUST hide implementation tables such as `backfill_progress` and `job_log` from public endpoints
- MUST implement Unit tests for all handlers and benchmark tests for critical endpoints
- MUST codecoverage of at least 80% for all handlers
- MUST codecoverage of at least 100% for all critical endpoints (e.g. KPIs, recent moves)

## API Design

- REST endpoints:
  - OpenAPI spec and Swagger UI: /openapi.yaml, /docs
  - Health and metrics endpoints: /health, /metrics

- REST endpoints for Stats:
  - Global stats: /api/v3/stats/kpis
  - Countries stats: /api/v3/stats/countries
  - Leaderboard: /api/v3/stats/leaderboard

- REST endpoints for General activity:
  - GeoKrety Recent moves: /api/v3/geokrety/recent-moves
  - Recently born GeoKrety: /api/v3/geokrety/recent-born
  - Recently loved GeoKrety: /api/v3/geokrety/recent-loved
  - Recently watched GeoKrety: /api/v3/geokrety/recent-watched
  - Recently active countries: /api/v3/countries/recent-active
  - Recently active waypoints: /api/v3/waypoints/recent-active
  - Recently registered users: /api/v3/users/recent-registered
  - Recently active users: /api/v3/users/recent-active

- REST endpoints for GeoKrety and users activity:
  - Country details: /api/v3/countries/{code}
    - GeoKrety in country: /api/v3/countries/{code}/geokrety
    - Country leaderboard: /api/v3/countries/{code}/leaderboard

  - User details: /api/v3/users/{id}
    - Owned GeoKrety: /api/v3/users/{id}/geokrety-owned
    - Loved GeoKrety: /api/v3/users/{id}/geokrety-loved
    - Watched GeoKrety: /api/v3/users/{id}/geokrety-watched
    - Country history: /api/v3/users/{id}/countries
    - Waypoint history: /api/v3/users/{id}/waypoints

  - GeoKrety details: /api/v3/geokrety/{id}
    - Move history: /api/v3/geokrety/{id}/moves
    - Current location: /api/v3/geokrety/{id}/location
    - Country history: /api/v3/geokrety/{id}/countries
    - Waypoint history: /api/v3/geokrety/{id}/waypoints
    - Loved by users: /api/v3/geokrety/{id}/loved-by
    - Watched by users: /api/v3/geokrety/{id}/watched-by

- WebSocket for live updates and connection count

## Database Schema and Tables

- Never include secret or sensitive information in the API responses such as:
  - `geokrety.gk_geokrety.tracking_code`
  - `geokrety.gk_users.*secid*`
  - `geokrety.gk_users.*email*`
  - `geokrety.gk_users.password`
  - `geokrety.gk_users.*ip*`
  - `geokrety.gk_users.home_latitude/longitude`
  - `geokrety.gk_users.home_position`
  - `geokrety.gk_users.terms_of_use_datetime`
  - `geokrety.gk_users.last_mail_datetime`
  - `geokrety.gk_users.last_login_datetime`
  - `geokrety.gk_users.updated_on_datetime`
  - `geokrety.gk_users.list_unsubscribe_token`

- Use existing tables:
  - `geokrety.gk_moves`
  - `geokrety.gk_geokrety`
  - `geokrety.gk_users`
  - `geokrety.gk_waypoints`
  - `geokrety.gk_loves`
  - `geokrety.gk_watched`
  - `geokrety.gk_pictures`
  - `stats.continent_reference`
  - `stats.country_daily_stats`
  - `stats.country_pair_flows`
  - `stats.daily_active_users`
  - `stats.daily_activity`
  - `stats.daily_entity_counts`
  - `stats.entity_counters_shard`
  - `stats.first_finder_events`
  - `stats.gk_cache_visits`
  - `stats.gk_countries_visited`
  - `stats.gk_country_history`
  - `stats.gk_milestone_events`
  - `stats.gk_related_users`
  - `stats.hourly_activity`
  - `stats.mv_country_month_rollup`
  - `stats.mv_global_kpi`
  - `stats.mv_top_caches_global`
  - `stats.user_cache_visits`
  - `stats.user_countries`
  - `stats.user_related_users`
  - `stats.waypoints`
