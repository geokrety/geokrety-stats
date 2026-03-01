//go:build integration

package computers_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fixtureReference struct {
	FixtureName string `json:"fixture_name"`
	Input       struct {
		StartID int64 `json:"start_id"`
		EndID   int64 `json:"end_id"`
	} `json:"input"`
}

func TestReplayFixturesAgainstReferenceJSON(t *testing.T) {
	if os.Getenv("GK_RUN_REPLAY_FIXTURE_TEST") != "1" {
		t.Skip("set GK_RUN_REPLAY_FIXTURE_TEST=1 to run replay fixture integration tests")
	}

	host := envOr("GK_FIXTURE_DB_HOST", "192.168.130.65")
	port := envOr("GK_FIXTURE_DB_PORT", "5432")
	user := envOr("GK_FIXTURE_DB_USER", "geokrety")
	password := envOr("GK_FIXTURE_DB_PASS", "geokrety")

	repoRoot := "/home/kumy/GIT/geokrety-points-system"
	statsRoot := filepath.Join(repoRoot, "geokrety-stats")
	fixturesRoot := filepath.Join(statsRoot, "internal/computers/testdata/replay_fixtures")

	dirs, err := os.ReadDir(fixturesRoot)
	require.NoError(t, err)

	fixtureNames := make([]string, 0, len(dirs))
	for _, entry := range dirs {
		if !entry.IsDir() {
			continue
		}
		expectedPath := filepath.Join(fixturesRoot, entry.Name(), "expected_reference.json")
		if _, err := os.Stat(expectedPath); err == nil {
			fixtureNames = append(fixtureNames, entry.Name())
		}
	}
	sort.Strings(fixtureNames)
	require.NotEmpty(t, fixtureNames, "no replay fixtures found in %s", fixturesRoot)

	for _, fixtureName := range fixtureNames {
		fixtureName := fixtureName
		t.Run(fixtureName, func(t *testing.T) {
			fixtureDir := filepath.Join(fixturesRoot, fixtureName)
			expectedPath := filepath.Join(fixtureDir, "expected_reference.json")

			expectedBytes, err := os.ReadFile(expectedPath)
			require.NoError(t, err)

			var expectedRef fixtureReference
			require.NoError(t, json.Unmarshal(expectedBytes, &expectedRef))

			var expectedAny any
			require.NoError(t, json.Unmarshal(expectedBytes, &expectedAny))

			dbName := fmt.Sprintf("geokrety_replay_%s_%d", sanitizeIdentifier(fixtureName), time.Now().UnixNano())
			t.Cleanup(func() {
				runCmdNoFail(t, statsRoot, psqlEnv(password),
					"dropdb", "-h", host, "-p", port, "-U", user, "--if-exists", dbName)
			})

			runCmd(t, statsRoot, psqlEnv(password),
				"createdb", "-h", host, "-p", port, "-U", user, dbName)

			setupSQL := `
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
`

			runCmd(t, statsRoot, psqlEnv(password),
				"psql", "-h", host, "-p", port, "-U", user, "-d", dbName, "-v", "ON_ERROR_STOP=1", "-c", setupSQL)

			usersCSV := filepath.Join(fixtureDir, "gk_users.csv")
			gksCSV := filepath.Join(fixtureDir, "gk_geokrety.csv")
			movesCSV := filepath.Join(fixtureDir, "gk_moves.csv")

			runCmd(t, statsRoot, psqlEnv(password),
				"psql", "-h", host, "-p", port, "-U", user, "-d", dbName, "-v", "ON_ERROR_STOP=1", "-c",
				fmt.Sprintf("\\copy geokrety.gk_users(id) FROM '%s' WITH CSV HEADER", usersCSV))
			runCmd(t, statsRoot, psqlEnv(password),
				"psql", "-h", host, "-p", port, "-U", user, "-d", dbName, "-v", "ON_ERROR_STOP=1", "-c",
				"UPDATE geokrety.gk_users SET username='u'||id::text")
			runCmd(t, statsRoot, psqlEnv(password),
				"psql", "-h", host, "-p", port, "-U", user, "-d", dbName, "-v", "ON_ERROR_STOP=1", "-c",
				fmt.Sprintf("\\copy geokrety.gk_geokrety(id,type,owner,created_on_datetime,holder) FROM '%s' WITH CSV HEADER", gksCSV))
			runCmd(t, statsRoot, psqlEnv(password),
				"psql", "-h", host, "-p", port, "-U", user, "-d", dbName, "-v", "ON_ERROR_STOP=1", "-c",
				"UPDATE geokrety.gk_geokrety SET name='gk-'||id::text, tracking_code='TK'||id::text")
			runCmd(t, statsRoot, psqlEnv(password),
				"psql", "-h", host, "-p", port, "-U", user, "-d", dbName, "-v", "ON_ERROR_STOP=1", "-c",
				fmt.Sprintf("\\copy geokrety.gk_moves(id,geokret,author,move_type,waypoint,country,lat,lon,moved_on_datetime) FROM '%s' WITH CSV HEADER", movesCSV))

			dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)

			runCmd(t, statsRoot, map[string]string{"GK_STATS_DB_URL": dbURL},
				"./bin/geokrety-stats", "-migration-up")
			runCmd(t, statsRoot, map[string]string{"GK_STATS_DB_URL": dbURL},
				"./bin/geokrety-stats", "-replay",
				"-start-id", fmt.Sprintf("%d", expectedRef.Input.StartID),
				"-end-id", fmt.Sprintf("%d", expectedRef.Input.EndID),
				"-truncate")

			actualAny := computeReference(t, statsRoot, host, port, user, password, dbName, fixtureName)
			require.Equal(t, expectedAny, actualAny, "replay output reference mismatch")
		})
	}
}

func computeReference(t *testing.T, cwd, host, port, user, password, dbName, fixtureName string) any {
	t.Helper()
	query := fmt.Sprintf(`WITH
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
SELECT jsonb_build_object(
  'fixture_name', %s,
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
)::text;`, quoteSQLString(fixtureName))

	cmd := exec.Command("psql", "-h", host, "-p", port, "-U", user, "-d", dbName, "-t", "-A", "-F", "", "-c", query)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "PGPASSWORD="+password, "PSQLRC=/dev/null")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	clean := strings.TrimSpace(string(out))
	var actual any
	require.NoError(t, json.Unmarshal([]byte(clean), &actual), clean)
	return actual
}

func runCmd(t *testing.T, cwd string, env map[string]string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd
	cmd.Env = os.Environ()
	mergedEnv := map[string]string{"PSQLRC": "/dev/null"}
	maps.Copy(mergedEnv, env)
	for key, value := range mergedEnv {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s %v\n%s", name, args, string(out))
}

func runCmdNoFail(t *testing.T, cwd string, env map[string]string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd
	cmd.Env = os.Environ()
	mergedEnv := map[string]string{"PSQLRC": "/dev/null"}
	maps.Copy(mergedEnv, env)
	for key, value := range mergedEnv {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	_, _ = cmd.CombinedOutput()
}

func psqlEnv(password string) map[string]string {
	return map[string]string{
		"PGPASSWORD": password,
		"PSQLRC":     "/dev/null",
	}
}

func envOr(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func sanitizeIdentifier(name string) string {
	var builder strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			builder.WriteRune(ch)
			continue
		}
		if ch >= 'A' && ch <= 'Z' {
			builder.WriteRune(ch + ('a' - 'A'))
			continue
		}
		builder.WriteRune('_')
	}

	if builder.Len() == 0 {
		return "fixture"
	}

	return builder.String()
}

func quoteSQLString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
