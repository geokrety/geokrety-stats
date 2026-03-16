
# GeoKrety Stats API — Design Notes

This repository provides the backend API for the GeoKrety Stats frontend. It is a Go HTTP server that sits between the existing PostgreSQL database (schema & migrations maintained in the `geokrety-website` repo) and the frontend dashboard. The API exposes REST endpoints for the dashboard, a WebSocket endpoint for live updates, and serves OpenAPI/Swagger documentation.

Goals
- Serve dashboard data required by `geokrety-stats-frontend` (see `src/composables/useApi.ts`).
- Provide a real-time WebSocket channel to push updates and report the number of connected clients.
- Expose machine-readable API docs (OpenAPI/Swagger) and human-friendly Swagger UI at `/docs`.
- Provide health, logging, and metrics endpoints for operations.

Recommended tech choices (implementers may change if justified):
- HTTP router: `chi` for lightweight routing.
- DB access: `pgx` + `sqlx` for typed queries. Prefer prepared statements.
- WebSocket: `nhooyr/websocket` for safe handling.
- OpenAPI/Swagger: `swaggo/swag` to generate docs from code comments, serve Swagger UI at `/docs`.
- Logging: `uber-go/zap` (structured logging) with `LOG_LEVEL` env var.
- Graceful shutdown: use signal handling with context timeouts.

Environment / Configuration
- `DATABASE_URL` or `PGHOST, PGPORT, PGUSER, PGPASSWORD, PGDATABASE`
- `PORT` (default `7415`)
- `LOG_LEVEL` (debug|info|warn|error)
- `ENABLE_SWAGGER` (true/false)
- `WS_BROADCAST_INTERVAL` (ms) — for periodic status broadcasts (optional)

REST Endpoints (suggested basics)
- `GET /openapi.yaml` — raw OpenAPI spec
- `GET /docs` — Swagger UI (serve only in non-production or when `ENABLE_SWAGGER=true`)
- `GET /health` — simple health check
- `GET /metrics` — Prometheus metrics (if enabled)

WebSocket
- Endpoint: `ws://<host>/ws` (or `wss://` in production)
- Subscriptions: simple topic model (e.g. path `user:123`, `global`, `countries`) or single broadcast channel depending on needs
- Message envelope (JSON):
	{
		"type": "stats_update" | "move" | "conn_count" | "heartbeat",
		"path": "global" | "user:123" | "country:PL",
		"data": { ... }
	}
- Connection counting: server maintains an atomic counter of active WS connections and broadcasts a `conn_count` message periodically or on connect/disconnect to subscribers.

Live update flow (example)
- Backend receives new data (push or DB trigger) or a scheduled poll retrieves updated aggregates.
- API pushes update to connected websocket clients matching subscription.
- Frontend listens and refreshes the relevant UI (the frontend composable should merge or refetch as needed).

OpenAPI / Swagger
- Keep an `openapi.yaml` or generate from Go comments (swaggo). Serve a Swagger UI at `/docs` and expose raw spec at `/openapi.yaml`.

Logging & Levels
- Use structured logs.
- Controlled by `LOG_LEVEL` (debug|info|warn|error).
- Log important lifecycle events: startup, DB connection, schema version (if known), incoming requests (optionally with request id), websocket connect/disconnect, errors.

Operational considerations
- Database connections: use a connection pool tuned to environment (avoid exhausting DB connections used by other apps). Prefer read-only replicas for heavy dashboard loads if available.
- Rate limiting and caching: implement response caching for expensive aggregates and consider rate limits for endpoints that trigger heavy queries.
- Security: enable CORS only for permitted origins; prefer `wss` in production.
- Migrations: database schema/migrations are maintained in `geokrety-website/website/db/migrations`. This service should assume the database schema is up-to-date.
- Metrics: expose `/metrics` for Prometheus to scrape. Count requests, errors, ws_connections, and broadcast events.

## Docker Compose (local development)

Use the included `docker compose.yml` to start a small local stack (API, PostgreSQL and MQTT broker).

- Build and start the stack (force recreate):

```bash
cd geokrety-stats-api
docker compose up --force-recreate -d
```

- Stop the stack:

```bash
cd geokrety-stats-api
docker compose down
```

- `Makefile` helpers: fmt lint vet build run test tidy docker-build up down clean
  - `make build`
  - `make run`
  - `make test`

```bash
# Start the stack (uses docker compose --force-recreate -d)
make up

# Stop the stack
make down
```
