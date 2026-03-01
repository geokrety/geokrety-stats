# Replay Fixtures (Reference JSON)

This directory contains small replay datasets and expected reference outputs used by
`TestReplayFixturesAgainstReferenceJSON`.

Each fixture directory contains:

- `gk_moves.csv`
- `gk_geokrety.csv`
- `gk_users.csv`
- `fixture_metadata.json`
- `expected_reference.json`

## Available fixtures

- `realistic_100`: one-day realistic sample (~100 moves)
- `all_rules_70`: curated multi-rule sample from real module hits
- `chain_30`: chain/reach-heavy sample
- `rescuer_20`: rescuer-focused sample
- `small_15`: very small mixed sample

## Refresh workflow

From `geokrety-stats/`:

1. Re-extract fixture inputs from source DB:

	```bash
	make fixtures_extract
	```

2. Rebuild all expected references:

	```bash
	make fixtures_refresh_refs
	```

3. Refresh one fixture only:

	```bash
	make fixtures_refresh_one FIXTURE=all_rules_70
	```

4. Run replay fixture integration tests:

	```bash
	make fixtures_test
	```

## DB environment overrides

Use env vars if default DB connection is not desired:

- Source extraction DB:
  - `GK_SOURCE_DB_HOST`, `GK_SOURCE_DB_PORT`, `GK_SOURCE_DB_USER`, `GK_SOURCE_DB_PASS`, `GK_SOURCE_DB_NAME`
- Temporary replay DB:
  - `GK_FIXTURE_DB_HOST`, `GK_FIXTURE_DB_PORT`, `GK_FIXTURE_DB_USER`, `GK_FIXTURE_DB_PASS`

