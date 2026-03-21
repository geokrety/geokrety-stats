#!/usr/bin/env python3
from typing import Any

import yaml  # type: ignore[import-untyped]


# Read the modular files
def load_yaml(path: str) -> Any:
    with open(path, encoding='utf-8') as input_file:
        return yaml.safe_load(input_file)


# Read individual path files
system_paths = load_yaml('openapi/paths/system.yaml')
stats_paths = load_yaml('openapi/paths/stats.yaml')
geokrety_paths = load_yaml('openapi/paths/geokrety.yaml')
countries_paths = load_yaml('openapi/paths/countries.yaml')
waypoints_paths = load_yaml('openapi/paths/waypoints.yaml')
users_paths = load_yaml('openapi/paths/users.yaml')
pictures_paths = load_yaml('openapi/paths/pictures.yaml')

parameters = load_yaml('openapi/components/parameters.yaml')
responses = load_yaml('openapi/components/responses.yaml')
schemas = load_yaml('openapi/components/schemas.yaml')

# Merge all paths
all_paths: dict[str, Any] = {}
all_paths.update(system_paths or {})
all_paths.update(stats_paths or {})
all_paths.update(geokrety_paths or {})
all_paths.update(countries_paths or {})
all_paths.update(waypoints_paths or {})
all_paths.update(users_paths or {})
all_paths.update(pictures_paths or {})

# Create the full OpenAPI spec
spec = {
    'openapi': '3.0.3',
    'info': {
        'title': 'GeoKrety Stats API',
        'version': '3.1.0',
        'description': '''Read-only API for GeoKrety statistics,
    activity feeds,
entity exploration, and visualization workloads.

Every `/api/v3/...` JSON endpoint returns a response envelope with `data`
and `meta`. List endpoints support `limit` and `offset` pagination.
Search endpoints require a `q` query parameter of at least 2 characters.

## GeoKret Public Identifiers (GKID)

GeoKrety endpoints accept public identifiers in three forms:
- **Numeric decimal**: `1`, `123`, `65535` (direct GKID number)
- **Bare hexadecimal**: `0001`, `00FF`, `FFFF` (hex without the `GK` prefix)
- **GK-prefixed hexadecimal**: `GK0001`, `GK00FF`, `GKFFFF`
  (padded hex representation)

Parsing is case-insensitive. Digit-only inputs without a leading zero are
treated as decimal, while zero-padded digit-only inputs are treated as
hexadecimal.

All forms resolve to the same GeoKret. Examples:
- `/api/v3/geokrety/1` and `/api/v3/geokrety/GK0001` fetch the same GeoKret
- `/api/v3/geokrety/00FF` and `/api/v3/geokrety/GK00FF` fetch the same GeoKret
- `/api/v3/geokrety/255` and `/api/v3/geokrety/GK00FF` are equivalent'''
    },
    'servers': [
        {
            'url': 'http://192.168.130.65:7415',
            'description': 'Local network development server'
        },
        {
            'url': 'http://localhost:7415',
            'description': 'Local development server'
        }
    ],
    'tags': [
        {
            'name': 'System',
            'description': (
                'Service metadata, health checks, metrics, '
                'and websocket entrypoint'
            ),
        },
        {
            'name': 'Stats',
            'description': 'Global graph and leaderboard datasets',
        },
        {
            'name': 'Geokrety',
            'description': (
                'GeoKret entity details, feeds, graphs, '
                'and map payloads'
            ),
        },
        {
            'name': 'Countries',
            'description': (
                'Country-level aggregates and currently '
                'spotted GeoKrety'
            ),
        },
        {
            'name': 'Waypoints',
            'description': (
                'Waypoint search, details, and spotted '
                'or historical GeoKrety'
            ),
        },
        {
            'name': 'Users',
            'description': (
                'User details, activity, relationships, '
                'collections, and graphs'
            ),
        },
        {'name': 'Pictures', 'description': 'Read-only picture metadata'}
    ],
    'paths': all_paths,
    'components': {
        'parameters': parameters,
        'responses': responses,
        'schemas': schemas
    }
}

# Write the merged spec
with open('openapi.yaml', 'w', encoding='utf-8') as handle:
    yaml.dump(
        spec,
        handle,
        sort_keys=False,
        allow_unicode=True,
        default_flow_style=False,
    )

print("✓ OpenAPI spec rebuilt successfully")
