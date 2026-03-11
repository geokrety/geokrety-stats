# Documentation Development and Deployment

This documentation site is built with [Zensical](https://zensical.org/), a modern static site generator with Material theme support.

## Local Development

### Initial Setup

```bash
# Initialize Python virtual environment
make init

# Install dependencies (Zensical)
make install
```

### Build and Preview

```bash
# Build the site
make build

# Serve with live reload (for editing docs)
make serve

# Build and serve with a clickable OSC 8 terminal link
make preview
```

### Available Make Targets

- `make help` - Show all available targets
- `make init` - Initialize Python venv
- `make install` - Install Zensical
- `make build` - Build static site to `site/`
- `make serve` - Start live-reload server (localhost:8000)
- `make preview` - Build and serve with clickable link
- `make clean` - Remove build artifacts

## CI/CD Behavior

### On Pull Request
- **Workflow**: `.github/workflows/zensical-pr.yml`
- **Trigger**: Any PR targeting `main` with changes to `docs/`, `zensical.toml`, or workflow files
- **Action**: Builds site and uploads `site/` directory as artifact `zensical-site` for reviewers to download and inspect
- **Retention**: 30 days

### On Push to Main
- **Workflow**: `.github/workflows/zensical-deploy.yml`
- **Trigger**: Push to `main` with changes to `docs/`, `zensical.toml`, or deployment workflow
- **Action**: Builds site and automatically publishes to GitHub Pages
- **URL**: https://geokrety.github.io/geokrety-stats/

## Configuration

- **Config file**: `zensical.toml` (site metadata, navigation, theme settings)
- **Docs root**: `docs/` directory
- **Build output**: `site/` directory
- **Site name**: GeoKrety Stats specs

## Site Structure

```
docs/
├── index.md                           # Home page
├── database-refactor/                 # Database refactoring docs
│   ├── 00-SPRINT-INDEX.md            # Overview
│   ├── 01-sprint-1-foundation.md
│   └── sprint-2/ through sprint-6/
├── gamification/                      # Gamification spec
│   └── 00-SPEC-DRAFT-v1.md
└── MIGRATION_TO_ZENSICAL.md          # Migration guide
```

## Troubleshooting

### Local build issues?

If `make build` fails with `Markdown.meta` errors:
1. Update Zensical: `pip install --upgrade zensical`
2. Check [Zensical GitHub Issues](https://github.com/zensical/zensical/issues)
3. See [MIGRATION_TO_ZENSICAL.md](docs/MIGRATION_TO_ZENSICAL.md) for more details

### Need to rollback?

To revert to MkDocs:
```bash
git checkout HEAD -- docs/ Makefile .github/workflows/
pip install mkdocs mkdocs-material
make build
```

## Learn More

- [Zensical Documentation](https://zensical.org/docs/)
- [Migration notes](docs/MIGRATION_TO_ZENSICAL.md)
