# GeoKrety Points System - Runbook

## Quick Start

### Prerequisites
- Go 1.23.5+
- PostgreSQL 12+
- RabbitMQ 3.8+

### Installation

```bash
cd /home/kumy/GIT/geokrety-points-system

# Build
make build
# Output: bin/geokrety-stats

# Run tests
make test
# Output: 8 tests passing (4 EventGuard + 4 PointsAggregator aggregator)

# View available targets
make help
```

## Running in Daemon Mode (Real-time AMQP)

```bash
# Set environment (example)
export GK_STATS_DB_URL="postgresql://gk_user:password@db.internal:5432/geokrety"
export GK_STATS_AMQP_URL="amqp://gk_user:password@rabbit.internal:5672/"
export GK_STATS_LOG_LEVEL="info"

# Start daemon (subscribes to geokrety exchange, processes moves in real-time)
make run
# Or directly:
./bin/geokrety-stats
```

**Output**: Logs each processed move with final award count and duration
```
{"level":"info","move_id":12345,"duration_ms":45,"awards":3,"time":"...","message":"move processed"}
```

## Replay Historical Data

### Full Year
```bash
make run_2017
# Expands to: ./bin/geokrety-stats -replay -year 2017
```

### Specific ID Range
```bash
./bin/geokrety-stats -replay -start-id 10000 -end-id 50000
```

### Specific Date Range
```bash
./bin/geokrety-stats -replay -start-date 2020-01-01 -end-date 2020-12-31
```

### Refresh Stats (Truncate + Replay)
```bash
./bin/geokrety-stats -replay -year 2017 -truncate
# Clears geokrety_stats first, then replays 2017
```

### With Batch Delay (avoid DB overload)
```bash
export GK_STATS_REPLAY_BATCH_DELAY="100ms"
make run_2017
```

## Database Management

### Initialize Schema
Runs automatically on first daemon/replay start:
```bash
./bin/geokrety-stats -config /etc/geokrety/config.yaml
# Applies migrations/000001_create_stats_schema.up.sql
```

### Check Database
```bash
psql postgresql://user:pass@host:5432/geokrety

# View processed moves
SELECT COUNT(*) FROM geokrety_stats.processed_events;

# View total points awarded
SELECT SUM(points) FROM geokrety_stats.user_points_log;

# View active chains
SELECT COUNT(*) FROM geokrety_stats.gk_chains WHERE ended_at IS NULL;
```

### Reset Stats
```bash
# Option 1: Truncate via replay
./bin/geokrety-stats -replay -year 2017 -truncate

# Option 2: Manual truncate (drop + recreate schema)
psql -c "DROP SCHEMA geokrety_stats CASCADE;"
# Next run will recreate via migrations
```

## Configuration

### Environment Variables

**Database**:
```bash
GK_STATS_DB_URL="postgresql://user:pass@host:5432/db"
GK_STATS_DB_POOL_SIZE="25"
GK_STATS_DB_QUERY_TIMEOUT="30s"
```

**AMQP**:
```bash
GK_STATS_AMQP_URL="amqp://user:pass@host:5672/"
GK_STATS_AMQP_PREFETCH="10"
GK_STATS_AMQP_COMPRESSION="true"
```

**Logging**:
```bash
GK_STATS_LOG_LEVEL="info"  # debug, info, warn, error
GK_STATS_LOG_FORMAT="json" # json or text
```

**Replay**:
```bash
GK_STATS_REPLAY_BATCH_SIZE="1000"  # IDs per batch
GK_STATS_REPLAY_BATCH_DELAY="0ms"  # Delay between batches
```

**Stats/Scoring** (via config file or env):
```bash
# See internal/config/config.go for all StatsConfig fields
# Examples:
GK_STATS_BASE_POINTS="3"
GK_STATS_MULTIPLIER_MIN="1.0"
GK_STATS_MULTIPLIER_MAX="2.0"
GK_STATS_CHAIN_TIMEOUT_DAYS="60"
```

### Config File

Create `/etc/geokrety/config.yaml`:
```yaml
database:
  url: postgresql://user:pass@localhost:5432/geokrety
  pool_size: 25

amqp:
  url: amqp://guest:guest@localhost:5672/
  prefetch: 10
  compression: true

logging:
  level: info
  format: json

stats:
  base_points: 3
  multiplier_min: 1.0
  multiplier_max: 2.0
  relay_bonus_points: 5
  chain_timeout_days: 60
  chain_antifarming_months: 3

maintenance:
  enabled: true
  chain_expiry_enabled: true
```

Then run with:
```bash
./bin/geokrety-stats -config /etc/geokrety/config.yaml
```

## Monitoring & Logs

### Daemon Logs
```bash
# Follow real-time logs
./bin/geokrety-stats | grep "move_id"

# Check for errors
./bin/geokrety-stats 2>&1 | grep ERROR
```

### Key Log Levels
- **ERROR**: Move processing failed (e.g., DB error, invalid event)
- **WARN**: Skipped event (e.g., duplicate, non-scoreable)
- **INFO**: Successfully processed move (default)
- **DEBUG**: Detailed trace (verbose, affects performance)

### Metrics to Monitor
```sql
-- Moves processed today
SELECT COUNT(*) FROM geokrety_stats.processed_events
WHERE inserted_at > NOW() - INTERVAL '1 day';

-- Points awarded today
SELECT SUM(points) FROM geokrety_stats.user_points_log
WHERE inserted_at > NOW() - INTERVAL '1 day';

-- Queue of moves waiting to be processed
-- (if using AMQP, check RabbitMQ admin panel)

-- Chains active but approaching expiry
SELECT COUNT(*) FROM geokrety_stats.gk_chains
WHERE ended_at IS NULL AND last_active_at < NOW() - INTERVAL '50 days';
```

## Maintenance Tasks

### Chain Expiry (Automatic)
- **Schedule**: Every hour (via cron)
- **Trigger**: Chains inactive > 60 days (configurable)
- **Action**: Awards timeout bonus to chain starter, marks chain ended

**View expired chains**:
```sql
SELECT id, gk_id, ended_at FROM geokrety_stats.gk_chains
WHERE ended_at IS NOT NULL
ORDER BY ended_at DESC LIMIT 10;
```

### Recalculate User Scores
Not yet implemented; requires replay:
```bash
./bin/geokrety-stats -replay -year 2017 -truncate
```

### Database Backup
```bash
pg_dump postgresql://user:pass@host:5432/geokrety \
  --schema=geokrety_stats \
  > backup_stats_$(date +%Y%m%d).sql
```

## Troubleshooting

### Daemon won't start
```bash
# Check database connection
./bin/geokrety-stats -config /etc/geokrety/config.yaml

# Enable debug logging
export GK_STATS_LOG_LEVEL="debug"
./bin/geokrety-stats

# Check error in stdout/stderr
```

### AMQP connection errors
```bash
# Verify RabbitMQ is running
nc -zv rabbit.internal 5672

# Check credentials
export GK_STATS_AMQP_URL="amqp://guest:guest@rabbit.internal:5672/"

# Verify exchange exists
rabbitmqctl list_exchanges | grep geokrety

# (Daemon will auto-reconnect every 2-5 seconds)
```

### High DB latency
```bash
# Check pool size
export GK_STATS_DB_POOL_SIZE="50"

# Check slow queries
SELECT * FROM pg_stat_statements WHERE mean_time > 100;

# Verify indexes exist
\d+ geokrety_stats.processed_events
```

### Replay seems stuck
```bash
# Check batch progress (move IDs loaded)
export GK_STATS_REPLAY_BATCH_SIZE="100"  # Reduce batch size
export GK_STATS_LOG_LEVEL="debug"
./bin/geokrety-stats -replay -year 2017
```

## Testing

### Run All Tests
```bash
make test
# Output: 8 tests pass (2 test files: EventGuard, PointsAggregator)
```

### Run Specific Test
```bash
/usr/local/go/bin/go test ./internal/computers -v -run TestEventGuard
```

### Run with Coverage
```bash
/usr/local/go/bin/go test ./... -cover
```

### Future: Full Test Suite
Expected ~18-20 tests total:
- ✅ 4 tests: EventGuard (computer 00)
- ✅ 4 tests: PointsAggregator (computer 14)
- ⏳ 1 test each: Computers 01-13 (planned)
- ⏳ 3+ integration tests (all 15 computers chained)

## Performance Targets

- **Normal**: Process 100+ moves/sec (single instance)
- **Replay**: Process 50+ moves/sec (DB-bound)
- **Latency**: Move → awarded ≤ 1 second (daemon mode)

## Deployment

### Docker
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk add --no-cache ca-certificates postgresql-client
COPY --from=builder /app/bin/geokrety-stats /usr/local/bin/
ENV GK_STATS_LOG_LEVEL=info
CMD ["geokrety-stats"]
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: geokrety-stats
spec:
  replicas: 1
  selector:
    matchLabels:
      app: geokrety-stats
  template:
    metadata:
      labels:
        app: geokrety-stats
    spec:
      containers:
      - name: stats
        image: geokrety/geokrety-stats:latest
        env:
        - name: GK_STATS_DB_URL
          valueFrom:
            secretKeyRef:
              name: geokrety-secrets
              key: db-url
        - name: GK_STATS_AMQP_URL
          valueFrom:
            secretKeyRef:
              name: geokrety-secrets
              key: amqp-url
        livenessProbe:
          exec:
            command: ["wget", "--spider", "http://<hostip>:8080/health"]
          initialDelaySeconds: 10
          periodSeconds: 30
```

## Support

- **Logs**: Check stdout/stderr with `GK_STATS_LOG_LEVEL=debug`
- **Database**: See [ARCHITECTURE.md](ARCHITECTURE.md) for schema details
- **Code**: See [ARCHITECTURE.md](ARCHITECTURE.md) for design patterns
- **Issues**: Review replay mode (`-replay -year YYYY`) to verify data consistency

---

**Last Updated**: 2026-02-27
**Status**: Production-ready (8/18+ tests passing)
