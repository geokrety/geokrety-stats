---
description: 'Instructions for building the GeoKrety Stats API, including OpenAPI design, WebSocket implementation, and operational endpoints.'
applyTo: 'api/**/*.go'
---

# GeoKrety Stats API Instructions

## API Design & Structure

- MUST Use Go with `chi` for routing and `pgx` + `sqlx` for database interactions
- MUST Expose REST endpoints for stats, GeoKrety, users, and countries
- MUST Serve OpenAPI spec and Swagger UI for documentation
- MUST expose only stable view-like contracts, not internal maintenance tables

## Response Format & Data Handling

- MUST Be compliant with JSON REST API and include metadata, relationships, filtering and pagination
- MUST be able to return response as JSON, XML, or CSV based on Accept header
- MUST support GeoJSON for location data and include relevant fields in responses (leaflet integration)
- MUST represent dates as UTC ISO 8601 strings
- MUST include country codes and names in relevant responses for better frontend integration
- MUST Include relevant metadata in responses such as total counts, pagination info, and timestamps for better frontend integration
- MUST include request/query timing statistics in response metadata for monitoring and debugging purposes
- MUST include data_as_of and computed_at fields in responses
- MUST document freshness for materialized-view backed endpoints as part of the contract

## Data Type Marshaling Standards

For all custom type structs and registries (e.g., GeoKret types, move types, status types), implement consistent marshaling/unmarshaling support across multiple data formats:

### JSON Marshaling

- MUST implement `MarshalJSON()` and `UnmarshalJSON()` methods on all type structs
- MUST serialize type labels (not numeric IDs) as the primary JSON representation for human readability in API responses
- MUST support unmarshaling from both string labels and numeric type IDs in requests for flexibility
- MAY include optional metadata fields (e.g., `{id: 0, label: "Traditional"}`) in complex responses
- SHOULD follow the pattern established in `GeokretId` type implementation for consistency

### XML Marshaling

- MUST implement `MarshalXML()` and `UnmarshalXML()` methods on all type structs
- MUST support both element and attribute serialization patterns
- MUST implement `MarshalXMLAttr()` and `UnmarshalXMLAttr()` for XML attribute contexts
- SHOULD serialize type labels as element text or attribute values for XML compatibility

### CSV Marshaling

- MUST implement `MarshalCSV()` and `UnmarshalCSV()` methods or helper functions for bulk export/import operations
- MUST use a standardized CSV format: `ID,Label` (e.g., "0,Traditional" for GeoKret types)
- MUST support unmarshaling from multiple CSV formats: ID-only, label-only, or ID,Label pairs
- SHOULD include CSV header rows in bulk export endpoints that document the schema

### YAML Marshaling

- MUST implement `MarshalYAML()` and `UnmarshalYAML()` methods for configuration and documentation purposes
- MUST support both simple formats (single scalar value) and structured formats (e.g., `{id: 0, label: "Traditional"}`)
- SHOULD enable YAML configuration files to reference types by either ID or label
- MUST maintain backward compatibility with existing YAML-based configurations

### Implementation Guidance

- All type registries MUST use struct-based architectures with receiver methods (not package-level functions)
- Type IDs MUST be defined as exported constants to enable compile-time type checking
- Marshaling errors MUST provide clear error messages indicating which type ID or value was invalid
- All marshaling methods SHOULD be thoroughly unit tested with 100% code coverage
- Type registries SHOULD provide both singleton instances (for convenience) and support constructor injection (for testability)

### Reference Implementation

See [Type-Label Helpers Refactoring Specification](../../tmp/20260322-refactor-types/specification.md) for complete guidance on refactoring type-label helpers with comprehensive marshaling support across JSON, XML, CSV, and YAML formats.

## Performance & Optimization

- MUST be optimized for performance on datasets used for Data Visualisation (ECharts) and mapping (Leaflet.js)
- MUST Consider response caching for expensive endpoints and implement rate limiting if necessary

## Security & Privacy

- MUST Implement graceful shutdown and CORS policies for security
- MUST hide implementation tables such as backfill_progress and job_log from public endpoints
- MUST Never include secret or sensitive information in the API responses

## Database & Querying

- MUST Use prepared statements for all DB queries

## API Usability & Consistency

- MUST Implement sorting, filtering, and pagination for all list endpoints
- MUST Ensure consistent response formats with error handling
- MUST Validate input parameters and return appropriate error messages for invalid requests
- MUST Ensure all endpoints are idempotent and safe to call multiple times
- MUST Ensure proper error handling and status codes for all endpoints (e.g. 400 for bad requests, 500 for server errors)
- MUST include pagination metadata (total count, page size, current page) in list endpoints
- MUST include pagination compatible with infinite scrolling (e.g. cursor-based pagination) for endpoints with potentially large result sets

## Real-Time & WebSocket

- MUST Implement WebSocket for live updates and connection count
- MUST Broadcast WebSocket messages on relevant data changes and periodically broadcast connection count

## Logging & Monitoring

- MUST Use structured logging with zap and respect LOG_LEVEL

## Documentation & Testing

- MUST Document all endpoints and expected responses in OpenAPI spec
- MUST implement Unit tests for all handlers and benchmark tests for critical endpoints
- MUST codecoverage of at least 80% for all handlers
- MUST codecoverage of at least 100% for all critical endpoints (e.g. KPIs, recent moves)

## Must build the code and run the tests
- MUST build the code and run the tests to ensure everything is working correctly before committing any changes
- MUST ensure that all tests pass successfully and that the code is free of errors before committing any changes
- MUST ensure code coverage is at least 80% for all handlers and 100% for critical endpoints before committing any changes

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
  - Country list: /api/v3/countries
  - Country details: /api/v3/countries/{code}
    - GeoKrety in country: /api/v3/countries/{code}/geokrety
    - Country leaderboard: /api/v3/countries/{code}/leaderboard

  - User list: /api/v3/users
  - User details: /api/v3/users/{id}
    - Owned GeoKrety: /api/v3/users/{id}/geokrety-owned
    - Loved GeoKrety: /api/v3/users/{id}/geokrety-loved
    - Watched GeoKrety: /api/v3/users/{id}/geokrety-watched
    - Country history: /api/v3/users/{id}/countries
    - Waypoint history: /api/v3/users/{id}/waypoints

  - GeoKrety list: /api/v3/geokrety
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

# Tips for Implementation

- In order to improve the patching, apply updates directly to the existing files instead of replacing them entirely, the user will have direct feedback on the changes and can easily identify what was added or modified. This also helps to maintain the context of the code and reduces the chances of introducing errors during the patching process.
