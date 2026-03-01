# Leaderboard Dashboard – Development Setup

## Quick Start

### Production Mode (Default)
```bash
cd /home/kumy/GIT/geokrety-points-system

# Build and run production dashboard (optimized, nginx-based)
docker compose up -d leaderboard-dashboard

# Access: http://localhost:3000
```

### Development Mode (New)
```bash
cd /home/kumy/GIT/geokrety-points-system

# Build and run dev dashboard with Vite (hot reload enabled)
docker compose --profile dev up -d leaderboard-dashboard-dev

# Access: http://localhost:3000
# Changes to src/ files will automatically reload in the browser
```

## Architecture

### Multistage Build (docker-bake.hcl)

The build is now configured with three stages:

1. **dependencies** – Base node image with npm packages
2. **builder** – Builds production assets (npm run build)
3. **dev** – Node with Vite dev server (npm run dev)
4. **prod** – Nginx serving built assets + reverse proxy

### Docker Targets

- **BUILD_TARGET=prod** – Optimized nginx production build (default)
- **BUILD_TARGET=dev** – Vite dev server with hot module reload

## Volume Mounts (Dev Mode)

In dev mode, the following directories are mounted from the host:

```yaml
volumes:
  - ./leaderboard-dashboard/src:/app/src                       # Auto-reload on file changes
  - ./leaderboard-dashboard/public:/app/public                 # Static assets
  - ./leaderboard-dashboard/index.html:/app/index.html         # Main HTML
  - ./leaderboard-dashboard/vite.config.js:/app/vite.config.js # Vite config
  - ./leaderboard-dashboard/package.json:/app/package.json     # Dependencies
  - /app/node_modules                                          # Anonymous volume (container version)
```

The anonymous volume `/app/node_modules` ensures the container's installed node_modules are not overwritten by the host's version (which might not exist or be outdated).

## Commands

### Build Production Image
```bash
docker compose build leaderboard-dashboard
docker compose build --profile prod leaderboard-dashboard
```

### Build Dev Image
```bash
docker compose build --profile dev leaderboard-dashboard-dev
```

### Use docker-bake for Multistage Builds
```bash
cd leaderboard-dashboard

# Build both prod and dev images
docker buildx bake

# Build only prod
docker buildx bake prod

# Build only dev
docker buildx bake dev

# Build specific tags
docker buildx bake --set "*.tags=myregistry/dashboard:custom"
```

## Hot Reload Configuration

Dev mode uses these HMR (Hot Module Reload) settings:

```
VITE_HMR_HOST: localhost
VITE_HMR_PORT: 5173
VITE_HMR_PROTOCOL: ws
```

The Vite dev server runs on **port 3000** with HMR on **port 5173**.

## Switching Profiles

### Print active profiles
```bash
docker compose config | grep -A5 profiles
```

### Run with specific profile
```bash
# Dev mode
docker compose --profile dev up -d leaderboard-dashboard-dev

# Prod mode (default)
docker compose up -d leaderboard-dashboard
```

### Stop dev container
```bash
docker compose --profile dev down leaderboard-dashboard-dev
```

## Troubleshooting

### Changes not reflecting in browser
1. Ensure the file was saved
2. Check browser console for errors
3. Verify volume mount: `docker volume inspect <container_name>`
4. Restart the container: `docker compose restart leaderboard-dashboard-dev`

### Port already in use
If port 3000 is taken, update docker-compose.yml:
```yaml
ports:
  - "3001:3000"  # Host:Container
  - "5174:5173"  # HMR port
```

### Node modules issues
If package.json changes, rebuild to refresh node_modules:
```bash
docker compose --profile dev down leaderboard-dashboard-dev
docker compose --profile dev build --no-cache leaderboard-dashboard-dev
docker compose --profile dev up -d leaderboard-dashboard-dev
```

## Performance Notes

- **Dev mode** is designed for development with hot reload; it's slower and uses more resources
- **Prod mode** is optimized for performance and production deployment
- Dev builds do NOT include nginx or the production build optimizations
- For CI/CD, always use the prod target
