#!/usr/bin/env bash

set -euo pipefail

export PSQLRC=/dev/null

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FIXTURES_ROOT="$ROOT_DIR/internal/computers/testdata/replay_fixtures"

SRC_DB_HOST="${GK_SOURCE_DB_HOST:-192.168.130.65}"
SRC_DB_PORT="${GK_SOURCE_DB_PORT:-5432}"
SRC_DB_USER="${GK_SOURCE_DB_USER:-geokrety}"
SRC_DB_PASS="${GK_SOURCE_DB_PASS:-geokrety}"
SRC_DB_NAME="${GK_SOURCE_DB_NAME:-geokrety}"

TMP_DB_HOST="${GK_FIXTURE_DB_HOST:-192.168.130.65}"
TMP_DB_PORT="${GK_FIXTURE_DB_PORT:-5432}"
TMP_DB_USER="${GK_FIXTURE_DB_USER:-geokrety}"
TMP_DB_PASS="${GK_FIXTURE_DB_PASS:-geokrety}"

fixture_names() {
  cat <<'EOF'
realistic_100
chain_30
small_15
all_rules_70
rescuer_20
EOF
}

fixture_move_selector_sql() {
  local fixture="$1"
  case "$fixture" in
    realistic_100)
      cat <<'SQL'
SELECT id
FROM geokrety.gk_moves
WHERE moved_on_datetime >= '2017-02-24'
  AND moved_on_datetime < '2017-02-25'
ORDER BY id
LIMIT 100
SQL
      ;;
    chain_30)
      cat <<'SQL'
SELECT id
FROM geokrety.gk_moves
WHERE geokret = 11165
ORDER BY moved_on_datetime, id
LIMIT 30
SQL
      ;;
    small_15)
      cat <<'SQL'
SELECT id
FROM geokrety.gk_moves
WHERE moved_on_datetime >= '2016-01-15'
  AND moved_on_datetime < '2016-01-20'
ORDER BY id
LIMIT 15
SQL
      ;;
    all_rules_70)
      cat <<'SQL'
WITH module_moves AS (
  SELECT upl.module_source,
         upl.move_id,
         m.geokret,
         ROW_NUMBER() OVER (PARTITION BY upl.module_source, m.geokret ORDER BY upl.move_id DESC) AS rn_gk
  FROM geokrety_stats.user_points_log upl
  JOIN geokrety.gk_moves m ON m.id = upl.move_id
  WHERE upl.module_source IN (
      '05_country_crossing',
      '06_relay_bonus',
      '07_rescuer_bonus',
      '08_handover_bonus',
      '09_reach_bonus',
      '11_chain_bonus',
      '12_diversity_bonus_tracker'
  )
    AND upl.move_id IS NOT NULL
),
dedup AS (
  SELECT module_source,
         move_id,
         geokret,
         ROW_NUMBER() OVER (PARTITION BY module_source ORDER BY move_id DESC) AS rn_mod
  FROM module_moves
  WHERE rn_gk = 1
),
seed_gk AS (
  SELECT move_id, geokret
  FROM dedup
  WHERE rn_mod <= 1
),
realistic_base AS (
  SELECT id AS move_id, geokret
  FROM geokrety.gk_moves
  WHERE moved_on_datetime >= '2017-02-24'
    AND moved_on_datetime < '2017-02-25'
  ORDER BY id
  LIMIT 8
),
seed_all AS (
  SELECT move_id, geokret FROM seed_gk
  UNION
  SELECT move_id, geokret FROM realistic_base
),
context_ids AS (
  SELECT DISTINCT x.id
  FROM seed_all s
  JOIN LATERAL (
    SELECT id
    FROM geokrety.gk_moves x
    WHERE x.geokret = s.geokret
      AND x.id <= s.move_id
    ORDER BY x.id DESC
    LIMIT 7
  ) x ON TRUE
)
SELECT id
FROM context_ids
ORDER BY id
SQL
      ;;
    rescuer_20)
      cat <<'SQL'
WITH seed AS (
  SELECT upl.move_id,
         m.geokret
  FROM geokrety_stats.user_points_log upl
  JOIN geokrety.gk_moves m ON m.id = upl.move_id
  WHERE upl.module_source = '07_rescuer_bonus'
    AND upl.move_id IS NOT NULL
  ORDER BY upl.move_id DESC
  LIMIT 1
),
context_ids AS (
  SELECT DISTINCT x.id
  FROM seed s
  JOIN LATERAL (
    SELECT id
    FROM geokrety.gk_moves x
    WHERE x.geokret = s.geokret
      AND x.id <= s.move_id
    ORDER BY x.id DESC
    LIMIT 20
  ) x ON TRUE
)
SELECT id
FROM context_ids
ORDER BY id
SQL
      ;;
    *)
      echo "Unknown fixture: $fixture" >&2
      exit 1
      ;;
  esac
}

fixture_label() {
  local fixture="$1"
  case "$fixture" in
    realistic_100) echo "Realistic one-day activity sample" ;;
    chain_30) echo "Single-GK chain-heavy sample" ;;
    small_15) echo "Small mixed sample" ;;
    all_rules_70) echo "Curated cross-rule sample from production module hits" ;;
    rescuer_20) echo "Focused rescuer-bonus sample" ;;
    *) echo "$fixture" ;;
  esac
}

extract_fixture() {
  local fixture="$1"
  local fixture_dir="$FIXTURES_ROOT/$fixture"
  mkdir -p "$fixture_dir"

  local selector
  selector="$(fixture_move_selector_sql "$fixture")"

  PGPASSWORD="$SRC_DB_PASS" psql \
    -h "$SRC_DB_HOST" -p "$SRC_DB_PORT" -U "$SRC_DB_USER" -d "$SRC_DB_NAME" \
    -v ON_ERROR_STOP=1 -t -A -F '' \
    -c "COPY (WITH selected_moves AS ( $selector ) SELECT m.id, m.geokret, m.author, m.move_type, m.waypoint, m.country, m.lat, m.lon, m.moved_on_datetime FROM geokrety.gk_moves m JOIN selected_moves sm ON sm.id = m.id ORDER BY m.moved_on_datetime, m.id) TO STDOUT WITH CSV HEADER" > "$fixture_dir/gk_moves.csv"

  PGPASSWORD="$SRC_DB_PASS" psql \
    -h "$SRC_DB_HOST" -p "$SRC_DB_PORT" -U "$SRC_DB_USER" -d "$SRC_DB_NAME" \
    -v ON_ERROR_STOP=1 -t -A -F '' \
    -c "COPY (WITH selected_moves AS ( $selector ) SELECT DISTINCT g.id, g.type, g.owner, g.created_on_datetime, g.holder FROM geokrety.gk_geokrety g JOIN geokrety.gk_moves m ON m.geokret = g.id JOIN selected_moves sm ON sm.id = m.id ORDER BY g.id) TO STDOUT WITH CSV HEADER" > "$fixture_dir/gk_geokrety.csv"

  PGPASSWORD="$SRC_DB_PASS" psql \
    -h "$SRC_DB_HOST" -p "$SRC_DB_PORT" -U "$SRC_DB_USER" -d "$SRC_DB_NAME" \
    -v ON_ERROR_STOP=1 -t -A -F '' \
    -c "COPY (WITH selected_moves AS ( $selector ), ids AS ( SELECT DISTINCT m.author AS id FROM geokrety.gk_moves m JOIN selected_moves sm ON sm.id = m.id WHERE m.author IS NOT NULL UNION SELECT DISTINCT g.owner AS id FROM geokrety.gk_geokrety g JOIN geokrety.gk_moves m ON m.geokret = g.id JOIN selected_moves sm ON sm.id = m.id WHERE g.owner IS NOT NULL ) SELECT id FROM ids ORDER BY id) TO STDOUT WITH CSV HEADER" > "$fixture_dir/gk_users.csv"

  local metadata_file="$fixture_dir/fixture_metadata.json"
  PGPASSWORD="$SRC_DB_PASS" psql \
    -h "$SRC_DB_HOST" -p "$SRC_DB_PORT" -U "$SRC_DB_USER" -d "$SRC_DB_NAME" \
    -t -A -F '' -v ON_ERROR_STOP=1 <<SQL > "$metadata_file"
WITH selected_moves AS (
  $selector
), selected AS (
  SELECT m.*
  FROM geokrety.gk_moves m
  JOIN selected_moves sm ON sm.id = m.id
)
SELECT jsonb_pretty(jsonb_build_object(
  'fixture_name', '$fixture',
  'label', '$(fixture_label "$fixture")',
  'source_db', '$SRC_DB_NAME',
  'move_count', (SELECT COUNT(*) FROM selected),
  'gk_count', (SELECT COUNT(DISTINCT geokret) FROM selected),
  'author_count', (SELECT COUNT(DISTINCT author) FROM selected WHERE author IS NOT NULL),
  'move_type_count', (SELECT COUNT(DISTINCT move_type) FROM selected),
  'start_id', (SELECT MIN(id) FROM selected),
  'end_id', (SELECT MAX(id) FROM selected),
  'start_ts', (SELECT MIN(moved_on_datetime) FROM selected),
  'end_ts', (SELECT MAX(moved_on_datetime) FROM selected)
));
SQL

  sed -i '/^Pager usage is off\.$/d' "$metadata_file"
}

setup_temp_db() {
  local db_name="$1"
  PGPASSWORD="$TMP_DB_PASS" createdb -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" "$db_name"

  PGPASSWORD="$TMP_DB_PASS" psql \
    -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" -d "$db_name" \
    -v ON_ERROR_STOP=1 <<'SQL'
CREATE SCHEMA geokrety;
CREATE TABLE geokrety.gk_users (
  id BIGINT PRIMARY KEY,
  username TEXT,
  home_country TEXT,
  home_latitude DOUBLE PRECISION,
  home_longitude DOUBLE PRECISION,
  joined_on_datetime TIMESTAMPTZ,
  last_login_datetime TIMESTAMPTZ
);
CREATE TABLE geokrety.gk_geokrety (
  id BIGINT PRIMARY KEY,
  name TEXT,
  tracking_code TEXT,
  type INT NOT NULL,
  owner BIGINT,
  created_on_datetime TIMESTAMPTZ NOT NULL,
  born_on_datetime TIMESTAMPTZ,
  holder BIGINT,
  missing BOOLEAN DEFAULT FALSE,
  distance DOUBLE PRECISION DEFAULT 0,
  caches_count INT DEFAULT 0,
  non_collectible TIMESTAMPTZ,
  parked TIMESTAMPTZ,
  loves_count INT DEFAULT 0
);
CREATE TABLE geokrety.gk_moves (
  id BIGINT PRIMARY KEY,
  geokret BIGINT NOT NULL,
  author BIGINT,
  move_type INT NOT NULL,
  waypoint TEXT,
  country TEXT,
  lat DOUBLE PRECISION,
  lon DOUBLE PRECISION,
  distance DOUBLE PRECISION DEFAULT 0,
  moved_on_datetime TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_gk_moves_geokret_datetime ON geokrety.gk_moves(geokret, moved_on_datetime);
CREATE INDEX idx_gk_moves_datetime ON geokrety.gk_moves(moved_on_datetime);
CREATE INDEX idx_gk_moves_author ON geokrety.gk_moves(author);
SQL
}

load_fixture_to_temp_db() {
  local fixture="$1"
  local db_name="$2"
  local fixture_dir="$FIXTURES_ROOT/$fixture"

  PGPASSWORD="$TMP_DB_PASS" psql \
    -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" -d "$db_name" -v ON_ERROR_STOP=1 \
    -c "\\copy geokrety.gk_users(id) FROM '$fixture_dir/gk_users.csv' WITH CSV HEADER"

  PGPASSWORD="$TMP_DB_PASS" psql \
    -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" -d "$db_name" -v ON_ERROR_STOP=1 \
    -c "UPDATE geokrety.gk_users SET username='u'||id::text"

  PGPASSWORD="$TMP_DB_PASS" psql \
    -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" -d "$db_name" -v ON_ERROR_STOP=1 \
    -c "\\copy geokrety.gk_geokrety(id,type,owner,created_on_datetime,holder) FROM '$fixture_dir/gk_geokrety.csv' WITH CSV HEADER"

  PGPASSWORD="$TMP_DB_PASS" psql \
    -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" -d "$db_name" -v ON_ERROR_STOP=1 \
    -c "UPDATE geokrety.gk_geokrety SET name='gk-'||id::text, tracking_code='TK'||id::text"

  PGPASSWORD="$TMP_DB_PASS" psql \
    -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" -d "$db_name" -v ON_ERROR_STOP=1 \
    -c "\\copy geokrety.gk_moves(id,geokret,author,move_type,waypoint,country,lat,lon,moved_on_datetime) FROM '$fixture_dir/gk_moves.csv' WITH CSV HEADER"
}

refresh_reference() {
  local fixture="$1"
  local fixture_dir="$FIXTURES_ROOT/$fixture"
  local metadata_file="$fixture_dir/fixture_metadata.json"

  local start_id
  local end_id
  start_id="$(python3 -c "import json;print(json.load(open('$metadata_file'))['start_id'])")"
  end_id="$(python3 -c "import json;print(json.load(open('$metadata_file'))['end_id'])")"

  local db_name="geokrety_fixture_${fixture}_$RANDOM$RANDOM"
  trap 'PGPASSWORD="$TMP_DB_PASS" dropdb -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" --if-exists "$db_name" >/dev/null 2>&1 || true' EXIT

  setup_temp_db "$db_name"
  load_fixture_to_temp_db "$fixture" "$db_name"

  local db_url
  db_url="postgres://$TMP_DB_USER:$TMP_DB_PASS@$TMP_DB_HOST:$TMP_DB_PORT/$db_name?sslmode=disable"

  GK_STATS_DB_URL="$db_url" "$ROOT_DIR/bin/geokrety-stats" -migration-up >/dev/null
  GK_STATS_DB_URL="$db_url" "$ROOT_DIR/bin/geokrety-stats" -replay -start-id "$start_id" -end-id "$end_id" -truncate >/dev/null

  local reference_file="$fixture_dir/expected_reference.json"

  PGPASSWORD="$TMP_DB_PASS" psql \
    -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" -d "$db_name" \
    -t -A -F '' -v ON_ERROR_STOP=1 <<SQL > "$reference_file"
WITH
input_stats AS (
  SELECT
    COUNT(*)::bigint AS move_count,
    COUNT(DISTINCT geokret)::bigint AS gk_count,
    (COUNT(DISTINCT author) FILTER (WHERE author IS NOT NULL))::bigint AS author_count,
    MIN(id)::bigint AS start_id,
    MAX(id)::bigint AS end_id
  FROM geokrety.gk_moves
),
input_move_types AS (
  SELECT COALESCE(move_type::text, 'NULL') AS key, COUNT(*)::bigint AS count
  FROM geokrety.gk_moves
  GROUP BY 1
  ORDER BY 1
),
pipeline_results AS (
  SELECT COALESCE(pipeline_result, 'NULL') AS key, COUNT(*)::bigint AS count
  FROM geokrety_stats.processed_events
  GROUP BY 1
  ORDER BY 1
),
module_breakdown AS (
  SELECT COALESCE(module_source, 'NULL') AS module_source,
         COUNT(*)::bigint AS rows,
         ROUND(SUM(points)::numeric, 3) AS points_sum
  FROM geokrety_stats.user_points_log
  GROUP BY 1
  ORDER BY 1
),
top_users AS (
  SELECT user_id,
         ROUND(total_points::numeric, 3) AS total_points
  FROM geokrety_stats.user_points_totals
  ORDER BY total_points DESC, user_id ASC
  LIMIT 15
),
sample_log AS (
  SELECT user_id,
         ROUND(points::numeric, 3) AS points,
         COALESCE(module_source, '') AS module_source,
         COALESCE(label, '') AS label,
         COALESCE(move_id, 0)::bigint AS move_id,
         COALESCE(gk_id, 0)::bigint AS gk_id,
         COALESCE(chain_id, 0)::bigint AS chain_id,
         is_owner_reward
  FROM geokrety_stats.user_points_log
  ORDER BY COALESCE(move_id, 0), user_id, module_source, label, points, COALESCE(gk_id, 0), COALESCE(chain_id, 0), is_owner_reward
  LIMIT 40
),
chain_status AS (
  SELECT status, COUNT(*)::bigint AS count
  FROM geokrety_stats.gk_chains
  GROUP BY 1
  ORDER BY 1
)
SELECT jsonb_pretty(jsonb_build_object(
  'fixture_name', '$fixture',
  'input', jsonb_build_object(
    'move_count', (SELECT move_count FROM input_stats),
    'gk_count', (SELECT gk_count FROM input_stats),
    'author_count', (SELECT author_count FROM input_stats),
    'start_id', (SELECT start_id FROM input_stats),
    'end_id', (SELECT end_id FROM input_stats),
    'move_type_counts', (SELECT COALESCE(jsonb_agg(jsonb_build_object('move_type', key, 'count', count) ORDER BY key), '[]'::jsonb) FROM input_move_types)
  ),
  'output', jsonb_build_object(
    'processed_events', jsonb_build_object(
      'rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.processed_events),
      'pipeline_result_counts', (SELECT COALESCE(jsonb_agg(jsonb_build_object('result', key, 'count', count) ORDER BY key), '[]'::jsonb) FROM pipeline_results)
    ),
    'user_points', jsonb_build_object(
      'log_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.user_points_log),
      'totals_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.user_points_totals),
      'sum_total_points', COALESCE((SELECT ROUND(SUM(total_points)::numeric, 3) FROM geokrety_stats.user_points_totals), 0),
      'sum_log_points', COALESCE((SELECT ROUND(SUM(points)::numeric, 3) FROM geokrety_stats.user_points_log), 0),
      'owner_reward_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.user_points_log WHERE is_owner_reward),
      'module_breakdown', (SELECT COALESCE(jsonb_agg(jsonb_build_object('module_source', module_source, 'rows', rows, 'points_sum', points_sum) ORDER BY module_source), '[]'::jsonb) FROM module_breakdown),
      'top_users', (SELECT COALESCE(jsonb_agg(jsonb_build_object('user_id', user_id, 'total_points', total_points) ORDER BY total_points DESC, user_id), '[]'::jsonb) FROM top_users),
      'sample_log', (SELECT COALESCE(jsonb_agg(to_jsonb(sample_log) ORDER BY move_id, user_id, module_source, label, points, gk_id, chain_id, is_owner_reward), '[]'::jsonb) FROM sample_log)
    ),
    'chains', jsonb_build_object(
      'chains_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.gk_chains),
      'chain_members_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.gk_chain_members),
      'chain_completions_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.gk_chain_completions),
      'status_counts', (SELECT COALESCE(jsonb_agg(jsonb_build_object('status', status, 'count', count) ORDER BY status), '[]'::jsonb) FROM chain_status)
    ),
    'state_tables', jsonb_build_object(
      'gk_multiplier_state_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.gk_multiplier_state),
      'gk_countries_visited_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.gk_countries_visited),
      'user_owner_gk_counts_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.user_owner_gk_counts),
      'user_waypoint_monthly_counts_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.user_waypoint_monthly_counts),
      'user_monthly_diversity_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.user_monthly_diversity),
      'user_monthly_diversity_owners_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.user_monthly_diversity_owners),
      'user_monthly_diversity_drops_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.user_monthly_diversity_drops),
      'user_monthly_diversity_countries_rows', (SELECT COUNT(*)::bigint FROM geokrety_stats.user_monthly_diversity_countries)
    )
  )
));
SQL

  sed -i '/^Pager usage is off\.$/d' "$reference_file"

  PGPASSWORD="$TMP_DB_PASS" dropdb -h "$TMP_DB_HOST" -p "$TMP_DB_PORT" -U "$TMP_DB_USER" --if-exists "$db_name" >/dev/null
  trap - EXIT
}

extract_all() {
  while IFS= read -r fixture; do
    [ -z "$fixture" ] && continue
    echo "Extracting fixture: $fixture"
    extract_fixture "$fixture"
  done < <(fixture_names)
}

refresh_all() {
  while IFS= read -r fixture; do
    [ -z "$fixture" ] && continue
    echo "Refreshing reference: $fixture"
    refresh_reference "$fixture"
  done < <(fixture_names)
}

usage() {
  cat <<'EOF'
Usage:
  scripts/replay_fixtures.sh extract-all
  scripts/replay_fixtures.sh refresh-all
  scripts/replay_fixtures.sh refresh-one <fixture>

Fixtures:
  realistic_100
  chain_30
  small_15
  all_rules_70
  rescuer_20
EOF
}

main() {
  local cmd="${1:-}"
  case "$cmd" in
    extract-all)
      extract_all
      ;;
    refresh-all)
      refresh_all
      ;;
    refresh-one)
      local fixture="${2:-}"
      [ -z "$fixture" ] && { usage; exit 1; }
      refresh_reference "$fixture"
      ;;
    *)
      usage
      exit 1
      ;;
  esac
}

main "$@"
