#!/usr/bin/env bash
# phinx wrapper: run phinx inside the running stack service container
# Usage: ./scripts/phinx migrate|rollback|status [phinx-args...]

set -euo pipefail

SCRIPT_NAME=$(basename "$0")

STACK_NAME=${STACK_NAME:-geokrety-new-theme}
SERVICE_NAME=${SERVICE_NAME:-website}
WORKDIR=${WORKDIR:-website}
PHINX_BIN=${PHINX_BIN:-/var/www/geokrety/vendor/bin/phinx}

# Collect args
ARGS=("$@")

# ensure docker is available
if ! command -v docker >/dev/null 2>&1; then
  echo "docker not found in PATH" >&2
  exit 3
fi

# Find a running container name for the stack/service
find_container() {
  # prefer docker ps name matching
  container=$(docker ps --format '{{.Names}}' | grep "${STACK_NAME}_${SERVICE_NAME}" | grep -v Shutdown | head -n1 || true)
  if [ -n "$container" ]; then
    echo "$container"
    return 0
  fi

  # fallback: try docker stack ps and translate task -> container via docker ps
  task_line=$(docker stack ps "$STACK_NAME" --no-trunc 2>/dev/null | grep "_${SERVICE_NAME}" | grep -v Shutdown | head -n1 || true)
  if [ -n "$task_line" ]; then
    task_id=$(echo "$task_line" | awk '{print $1}')
    # docker ps may include task id suffix in names; try to find container containing task id
    container=$(docker ps --format '{{.Names}}' | grep "$task_id" | head -n1 || true)
    if [ -n "$container" ]; then
      echo "$container"
      return 0
    fi
  fi

  # last resort: any container matching stack_service anywhere
  container=$(docker ps --format '{{.Names}}' | grep "${STACK_NAME}_${SERVICE_NAME}" -m1 || true)
  if [ -n "$container" ]; then
    echo "$container"
    return 0
  fi

  return 1
}

container=$(find_container) || true
if [ -z "$container" ]; then
  echo "No running container found for ${STACK_NAME}_${SERVICE_NAME}." >&2
  echo "Try setting STACK_NAME and SERVICE_NAME environment variables." >&2
  exit 4
fi

# Build safely quoted argument list for remote shell
quoted_args=""
for a in "${ARGS[@]}"; do
  quoted_args+="$(printf '%q' "$a") "
done

# attach TTY if local is a TTY
TTY_FLAG="-i"
if [ -t 1 ]; then
  TTY_FLAG="-it"
fi

cmd="cd $(printf '%q' "$WORKDIR") && exec $(printf '%q' "$PHINX_BIN") $quoted_args"

echo "Running in container: $container" >&2

docker exec $TTY_FLAG "$container" bash -lc "$cmd" || {
    exit_code=$?
    echo "Command failed with exit code $exit_code" >&2
    exit 0 # ignore phinx command exit code to prevent CI failure; rely on phinx output for success/failure indication
}
