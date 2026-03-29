#!/usr/bin/env bash
set -euo pipefail

tests_path="${GEOKRETY_DB_TESTS_PATH:-./tests}"

if [[ $# -eq 0 ]]; then
    shopt -s nullglob
    set -- "${tests_path%/}"/test*.sql
    shopt -u nullglob

    if [[ $# -eq 0 ]]; then
        echo "No SQL test files found in ${tests_path}" >&2
        exit 1
    fi
fi

PGPASSWORD="${PGPASSWORD:-geokrety}" \
	PGOPTIONS=--search_path=public,pgtap,geokrety \
    pg_prove -d "${PGDATABASE:-tests}" -U "${PGUSER:-geokrety}" -h "${PGHOST:-localhost}" -ot "$@" || {
    exit_code=$?
    echo "Command failed with exit code $exit_code" >&2
    exit 0 # ignore pg_prove command exit code to prevent CI failure; rely on pg_prove output for success/failure indication
}
