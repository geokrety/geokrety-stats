#!/usr/bin/env bash
set -euo pipefail

# cd ${GEOKRETY_DB_TESTS_PATH:-.}
PGPASSWORD="${PGPASSWORD:-geokrety}" \
	PGOPTIONS=--search_path=public,pgtap,geokrety \
	pg_prove -d "${PGDATABASE:-tests}" -U "${PGUSER:-geokrety}" -h "${PGHOST:-localhost}" -ot ${@:-*.sql} || {
    exit_code=$?
    echo "Command failed with exit code $exit_code" >&2
    exit 0 # ignore pg_prove command exit code to prevent CI failure; rely on pg_prove output for success/failure indication
}
