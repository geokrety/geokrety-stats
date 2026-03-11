# AGENTS.md

## Project Overview

**geokrety-stats-frontend** is a modern Vue 3 web application that provides a responsive user interface for viewing GeoKrety statistics and movement data. The frontend consumes data from the [geokrety-stats-api](https://github.com/geokrety/geokrety-stats-api) using REST APIs and WebSocket connections for real-time updates.

### Key characteristics:

- **Framework**: Vue 3 with Composition API and TypeScript
- **UI Framework**: shadcn/vue with custom theme system
- **Theme Support**: Dark and light modes with persistent user preferences
- **Real-time Updates**: WebSocket integration for live statistics
- **API Integration**: REST API calls (GET) for initial data load and WebSocket subscriptions for live data
- **Target Browsers**: Modern browsers supporting ES2020+
- **Build Tool**: Vite for fast development and optimized production builds
- **Mapping**: [Leaflet.js](https://leafletjs.com/) (`leaflet` v1.x) for interactive world maps; dark CartoDB tile layer used throughout
- **Data Visualisation**: [ECharts](https://ECharts.org/)
- **Geo data**: [topojson-client](https://github.com/topojson/topojson-client) + [world-atlas](https://github.com/topojson/world-atlas) for country polygons; country numeric ISO 3166-1 → alpha-2 mapping is in `src/data/iso3166.ts`

### Related projects:

- [geokrety-website](https://github.com/geokrety/geokrety-website) - Main GeoKrety platform database

---

## GeoKrety.org — Quick "General" summary

- **Logtypes:** stored as integers. Current mapping:
  - `0`=drop/dropped
  - `1`=grab/grabbed
  - `2`=comment/commented
  - `3`=met/seen
  - `4`=archive/archived
  - `5`=visiting/dipped
- **GeoKrety types:** stored as integers. Current mapping:
  - `0` = Traditional
  - `1` = A book
  - `2` = A human
  - `3` = A coin
  - `4` = KretyPost
  - `5` = Pebble
  - `6` = Car
  - `7` = Playing card
  - `8` = Dog tag / pet
  - `9` = Jigsaw part
  - `10` = Easter egg
- **GKID ↔ integer conversion:** public IDs are `GK` + hexadecimal of internal id (padded to 4 digits); use hex conversion to map between `GK1234` and its integer id.

## Project Structure

```
/
├── public/                   # Static assets
│   ├── favicon.ico
│   └── index.html
├── src/
│   ├── assets/              # Images, fonts, static files
│   │   ├── css/
│   │   │   ├── main.css          # Global styles
│   │   │   └── globals.css        # shadcn/vue base styles
│   │   └── images/
│   ├── components/          # Reusable Vue components
│   │   ├── common/          # Shared components (Header, Footer, etc.)
│   │   ├── statistics/      # Statistics display components
│   │   ├── WorldChoropleth.vue   # Leaflet+D3 world choropleth map
│   │   └── forms/           # Form components (filters, search)
│   ├── composables/         # Reusable composition logic
│   │   ├── useWebSocket.ts  # WebSocket connection management
│   │   ├── useApi.ts        # REST API calls (incl. CountryStats + useCountries)
│   │   ├── useTheme.ts      # Theme switching logic
│   │   └── useFetch.ts      # Data fetching with loading states
│   ├── data/                # Static lookup tables
│   │   └── iso3166.ts       # ISO 3166-1 numeric → alpha-2 country code mapping
│   ├── stores/              # Pinia state management
│   │   ├── stats.ts         # Global statistics state
│   │   ├── countries.ts     # Country statistics state
│   │   └── theme.ts         # Theme preferences
│   ├── views/               # Page components
│   │   ├── HomeView.vue     # Landing page with hero, leaderboard, activity
│   │   ├── CountriesView.vue  # Country ranking with world map, table, cards
│   │   └── AboutView.vue
│   ├── services/            # API services
│   │   ├── api.ts           # REST API client
│   │   └── websocket.ts     # WebSocket service
│   ├── types/               # TypeScript type definitions
│   │   ├── api.ts           # API response types
│   │   └── components.ts    # Component prop types
│   ├── router/              # Vue Router configuration
│   │   └── index.ts
│   ├── App.vue              # Root component
│   └── main.ts              # Application entry point
├── tests/                   # Test files
│   ├── unit/
│   └── integration/
├── .env.example             # Environment variables template
├── .gitignore
├── Dockerfile               # Production container image
├── vite.config.ts           # Vite configuration
├── tsconfig.json            # TypeScript configuration
├── package.json             # Dependencies and scripts
├── pnpm-lock.yaml          # Lock file (if using pnpm)
└── README.md                # User-focused documentation
```

---

## Setup Commands

### Prerequisites

- Node.js 18+ (20+ recommended)
- Package manager: npm, yarn, or pnpm (pnpm preferred for monorepos)
- API server running: `geokrety-stats-api` available at configured URL
- Git for version control

### Local Development Setup

```bash
# Clone the repository
git clone https://github.com/geokrety/geokrety-stats-frontend.git
cd geokrety-stats-frontend

# Install dependencies
pnpm install
# or
npm install

# Copy environment template and configure
cp .env.example .env

# Edit .env with your configuration
# Key variables:
# - VITE_API_URL=http://localhost:3000/api
# - VITE_WS_URL=ws://localhost:3000/ws
```

### Environment Configuration

Create `.env` file with:

```bash
# API Configuration
VITE_API_URL=http://localhost:3000/api
VITE_WS_URL=ws://localhost:3000/ws
VITE_API_TIMEOUT=10000  # milliseconds

# Feature Flags
VITE_ENABLE_DEBUG=false
VITE_ENABLE_THEME_SWITCHER=true

# Analytics (optional)
VITE_ANALYTICS_ENABLED=false
VITE_ANALYTICS_ID=
```

---

## Development Workflow

### Starting the Development Server

```bash
# Start dev server with hot module replacement
pnpm dev
# Server runs at http://localhost:5173

# With custom port
pnpm dev -- --port 3001
```

Note: prefer running the dev server on the host for local development so the OS's inotify can be used (better performance, lower CPU than polling in Docker).

Use the provided Makefile to run the dev server on your host machine:

```bash
# from repo root
make install   # install deps (pnpm preferred)
make dev       # start the dev server on host (uses pnpm or npm)
```

Run containers only for production testing. To start production containers use:

```bash
docker compose --profile prod up --force-recreate
```

### Building for Production

```bash
# Build optimized production bundle
pnpm build

# Preview production build locally
pnpm preview
```

### Development Best Practices

**Component Development:**

- Create new components in `src/components/[feature]/`
- Use `<script setup lang="ts">` syntax
- Define props with `defineProps` and types
- Define emits with `defineEmits`
- Use `<style scoped>` for component styles

- **Prefer `shadcn/vue` primitives:** Use `shadcn/vue` components for common UI elements (cards, buttons, inputs, dialogs) rather than reimplementing them. This ensures consistent theming and accessibility across the app.

- **Install components via MCP:** Use the `shadcn-vue` MCP helper to add components from the official registry. Example commands (run from the frontend repo root):

```bash
# initialize MCP client (if not already done)
pnpm dlx shadcn-vue@latest mcp init --client vscode

# add components
pnpm dlx shadcn-vue@latest mcp add button
pnpm dlx shadcn-vue@latest mcp add card
pnpm dlx shadcn-vue@latest mcp add dialog
```

- **Integrate with theme system:** When using added components, prefer the shadcn tokens and Tailwind classes they provide. Keep theme switches using the project's `useTheme` composable and `src/assets/css/globals.css` variables.

- **Docs & MCP reference:** See https://www.shadcn-vue.com/docs/mcp and https://www.shadcn-vue.com/docs/dark-mode/vite for MCP usage and dark-mode integration.

**State Management:**

- Use Pinia stores in `src/stores/` for global state
- Keep stores focused on single domains
- Place async logic in store actions
- Use getters for computed state values

**Composables:**

- Create reusable logic in `src/composables/`
- Name files as `use[Feature].ts` (e.g., `useFetch.ts`)
- Document parameters and return values with JSDoc
- Clean up side effects in `onUnmounted`

**API Integration:**

- All API calls go through `src/services/api.ts`
- Use REST GET for initial loads
- Use WebSocket for real-time updates
- Handle loading, error, and success states
- Implement proper error messages

**Theme System:**

- Use shadcn/vue theming with CSS variables in `src/assets/css/globals.css`
- Theme values defined in `next.config.js` or `app.config.ts`
- Use `useTheme()` composable for theme switching
- Persist theme preference to localStorage

---

## API Integration

### REST API (GET requests)

Initial page load uses REST API calls:

```typescript
// Example service method
async function getUserStats(userId: number) {
  const response = await fetch(`${API_URL}/users/${userId}/stats`)
  return response.json()
}
```

**Key endpoints:**

- `GET /api/users` - List all users
- `GET /api/users/:id/stats` - User statistics
- `GET /api/geokrety/:id/stats` - Individual geokret statistics
- `GET /api/countries` - Country statistics

### WebSocket Subscription

Real-time updates via WebSocket on active pages:

```typescript
// useWebSocket composable pattern
const { subscribe, unsubscribe, data } = useWebSocket()

onMounted(() => {
  // Subscribe to live updates
  subscribe(`user:${userId}`, (update) => {
    // Handle real-time update
  })
})

onUnmounted(() => {
  // Clean up subscription
  unsubscribe(`user:${userId}`)
})
```

**WebSocket message format:**

```json
{
  "type": "stats_update",
  "path": "user:123",
  "data": {
    "userId": 123,
    "pointsAwarded": 150,
    "timestamp": "2026-03-06T12:34:56Z"
  }
}
```

**Connection Management:**

- Automatic reconnection with exponential backoff
- Heartbeat/ping-pong for connection health
- Message queue for offline handling
- Unsubscribe from channels when leaving pages

---

## Theme System

### Dark and Light Mode

The application supports user-selectable themes using shadcn/vue's built-in theming system.

**Theme Configuration:**

- `src/assets/css/globals.css` - Global styles and CSS custom properties
- Theme configuration in `tailwind.config.js` or `app.config.ts`
- Light and dark mode support via `data-theme` attribute

**Using Themes in Components:**

```vue
<template>
  <div class="rounded-lg border bg-card p-4">
    <h5 class="text-lg font-semibold text-primary">Statistics</h5>
    <p class="text-muted-foreground">Data from API</p>
  </div>
</template>

<style scoped>
/* shadcn/vue uses Tailwind CSS with theme-aware utilities */
/* Colors automatically respond to light/dark mode */
</style>
```

**Switching Themes:**

```typescript
// useTheme composable
const { theme, setTheme, isDark } = useTheme()

// Toggle dark/light
setTheme(isDark.value ? 'light' : 'dark')

// Persistence: theme saved to localStorage and reflected in DOM
```

**shadcn/vue Integration:**

- Use shadcn/vue pre-built components for consistency
- Leverage Tailwind CSS utility classes for responsive design
- Theme switching automatically applies to all shadcn/vue components via CSS variables

---

## Testing Instructions

### Unit Tests

```bash
# Run all unit tests
pnpm test

# Run tests in watch mode
pnpm test:watch

# Run tests with coverage report
pnpm test:coverage

# Run specific test file
pnpm test composables/useTheme.test.ts
```

### Component Tests

```bash
# Test components with Vue Test Utils
pnpm test components/**/*.test.ts

# Test composables
pnpm test composables/**/*.test.ts

# Test stores (Pinia)
pnpm test stores/**/*.test.ts
```

### Integration Tests

```bash
# Run E2E tests (if using Playwright/Cypress)
pnpm test:e2e

# Run in headed mode for debugging
pnpm test:e2e --headed
```

### Test Coverage Goals

- Unit tests: >80% coverage for utilities and composables
- Components: >70% coverage for interactive components
- Critical paths: 100% (theme switching, API calls, WebSocket)
- E2E: Key user flows (loading stats, theme switching, navigation)

---

### Docker Environment Testing

#### Test Development Environment

The development environment runs on **port 5173** with hot module replacement:

```bash
# Start development environment with Docker Compose
docker compose up --force-recreate

# Verify frontend is accessible at http://localhost:5173
# Check logs for errors
docker compose logs -f frontend

# Test WebSocket and API connectivity
# - Open browser DevTools Console
# - Check Network tab for API requests
# - Verify WebSocket connections are established

# Stop development environment
docker compose down
```

#### Test Production Environment

The production environment runs on **port 9090** with Nginx:

```bash
# Start production environment with Docker Compose
docker compose --profile prod up --force-recreate

# Verify build completed successfully
docker images | grep geokrety-stats-frontend

# Verify container is running
docker compose ps

# Test production application
# - Access http://localhost:9090
# - Verify all UI elements load correctly
# - Test theme switching
# - Monitor API calls in DevTools
# - Verify WebSocket connections
# - Check browser console for errors

# View production container logs
docker compose --profile prod logs -f frontend-prod

# Stop production environment
docker compose --profile prod down
```

#### Restore Development Environment

After finishing production testing, restore the development environment:

```bash
# Start development environment again
docker compose up --force-recreate

# Verify dev server is running at http://localhost:5173
# Verify hot module replacement is working
```

### Testing Checklist

**Development Environment Tests:**

- [ ] Docker Compose starts without errors
- [ ] Frontend accessible at http://localhost:5173
- [ ] Hot module replacement works (change component and see instant update)
- [ ] API requests succeed (check Network tab)
- [ ] WebSocket connection established (check Console)
- [ ] Theme switching works (dark/light)
- [ ] No console errors or warnings

**Production Environment Tests:**

- [ ] Docker build completes successfully
- [ ] Container starts and stays running
- [ ] Frontend accessible at http://localhost:9090
- [ ] All static assets load (CSS, JS, images)
- [ ] API requests succeed
- [ ] WebSocket connection established
- [ ] Theme switching works
- [ ] No console errors
- [ ] Performance acceptable (check DevTools Performance tab)
- [ ] Mobile responsive (test at different viewport sizes)

### Performance Benchmarks

```bash
# Measure bundle size
docker compose --profile prod exec frontend-prod ls -lh /usr/share/nginx/html

# Check production image size
docker images geokrety/geokrety-stats-frontend --format "{{.Size}}"

# Monitor runtime resources
docker compose stats
```

---

## Code Style

### Vue 3 / TypeScript Conventions

**File naming:**

- Components: PascalCase (e.g., `UserStats.vue`)
- Composables: camelCase with `use` prefix (e.g., `useWebSocket.ts`)
- Services: camelCase (e.g., `api.ts`)
- Types: PascalCase suffixed with type name (e.g., `User.ts`)

**Script Setup Syntax:**

```vue
<script setup lang="ts">
import { computed, ref } from 'vue'
import type { PropType } from 'vue'

interface Props {
  userId: number
  userName: string
}

const props = withDefaults(defineProps<Props>(), {
  userName: 'Anonymous',
})

const emit = defineEmits<{
  update: [userId: number]
}>()

const isLoading = ref(false)
const stats = computed(() => {
  // computed logic
})
</script>
```

**TypeScript Best Practices:**

```typescript
// Enable strict mode - use types everywhere
interface UserStats {
  userId: number
  totalPoints: number
  movesCount: number
  lastActive: Date
}

// Use type for API responses
type ApiResponse<T> = {
  data: T
  timestamp: string
}

// Generic composables with type parameters
function useFetch<T>(url: string) {
  const data = ref<T | null>(null)
  const error = ref<Error | null>(null)

  return { data, error }
}
```

**Formatting:**

```bash
# Format code with Prettier
pnpm format

# Check formatting without changes
pnpm format:check

# Lint code with ESLint
pnpm lint

# Fix linting issues
pnpm lint:fix
```

**Import Organization:**

```vue
<script setup lang="ts">
// 1. Vue core and plugins
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useStore } from 'pinia'

// 2. External libraries
import { debounce } from 'lodash-es'

// 3. Internal composables
import { useWebSocket } from '@/composables/useWebSocket'
import { useTheme } from '@/composables/useTheme'

// 4. Components
import AppHeader from '@/components/common/AppHeader.vue'
import StatCard from '@/components/statistics/StatCard.vue'

// 5. Types and constants
import type { UserStats } from '@/types/api'
import { API_TIMEOUT } from '@/constants'
</script>
```

### Tailwind CSS / shadcn/vue Conventions

```vue
<script setup lang="ts">
// shadcn/vue components already styled with Tailwind + theme support
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
</script>

<template>
  <!-- 1. Use shadcn/vue components for consistency -->
  <!-- 2. Compose with Tailwind utility classes -->
  <!-- 3. Theme colors respond automatically to light/dark mode -->

  <Card class="w-full">
    <CardHeader>
      <CardTitle class="text-lg">Statistics</CardTitle>
      <CardDescription>Real-time data updates</CardDescription>
    </CardHeader>
    <CardContent class="space-y-2">
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <div class="rounded-lg bg-muted p-4">
          <p class="text-muted-foreground">Total Points</p>
        </div>
      </div>
    </CardContent>
  </Card>
</template>

<style scoped>
/* Custom component styles - extend shadcn/vue components */
:deep(.card) {
  @apply transition-colors duration-200;
}
</style>
```

**Tailwind CSS Best Practices:**

- Use shadcn/vue components as building blocks
- Compose with Tailwind utility classes for customization
- Leverage Tailwind's responsive prefixes (sm:, md:, lg:, xl:)
- Use theme colors via CSS variables that respect light/dark modes
- Keep custom CSS minimal; prefer utility composition

---

## Build and Deployment

### Production Build

```bash
# Build optimized bundle for production
pnpm build

# Output directory: dist/
# All assets are minified and optimized

# Environment-specific builds
VITE_API_URL=https://api.example.com pnpm build
```

### Docker Deployment

**Dockerfile (multi-stage build):**

- **builder stage**: Node.js environment for building Vue application
- **runtime stage**: Lightweight Nginx for serving static files

**Build Targets:**

```bash
# Build development image (with pnpm and source files)
docker compose build

# Build production image (optimized, Nginx only)
docker compose --profile prod build
```

**Starting Environments:**

```bash
# Start development environment (port 5173) — default profile
docker compose up --force-recreate

# Start production environment (port 9090) — requires explicit profile
docker compose --profile prod up --force-recreate

# View logs (development)
docker compose logs -f frontend

# View logs (production)
docker compose --profile prod logs -f frontend-prod

# Stop development environment
docker compose down

# Stop production environment
docker compose --profile prod down

# Rebuild and restart (development)
docker compose up --force-recreate --build

# Rebuild and restart (production)
docker compose --profile prod up --force-recreate --build
```

**Environment Variables:**

- `VITE_API_URL` - Backend API URL (default: http://localhost:3000/api)
- `VITE_WS_URL` - WebSocket URL (default: ws://localhost:3000/ws)
- `VITE_API_TIMEOUT` - API timeout in ms (default: 10000)
- `VITE_ENABLE_DEBUG` - Enable debug logging (default: false)

### CI/CD Pipeline

GitHub Actions workflows in `.github/workflows/`:

- `build.yml` - Build and test on push
- `deploy.yml` - Deploy to staging/production on tag
- `lint.yml` - Linting and formatting checks

**Required checks before merge:**

```bash
pnpm lint
pnpm format:check
pnpm test
pnpm build
```

### Deployment Process

```bash
# 1. Create release tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# 2. Push tag (triggers CI/CD pipeline)
git push origin v1.0.0

# 3. GitHub Actions builds and pushes Docker image
# 4. Deploy via your infrastructure tooling
```

---

## WebSocket Real-time Updates

### Connection Lifecycle

```typescript
// useWebSocket composable manages lifecycle
const { connect, disconnect, subscribe, unsubscribe, isConnected } = useWebSocket()

// Automatic connection on composable creation
onMounted(() => {
  connect() // Usually automatic, but can be manual
})

// Subscribe to specific channels
subscribe('user:123', (message) => {
  // Handle incoming update
})

// Cleanup on unmount
onUnmounted(() => {
  unsubscribe('user:123')
  disconnect() // Only if last subscriber
})
```

### Message Subscriptions

```typescript
// Subscribe to multiple channels
subscribe('user:123', handleUserUpdate)
subscribe('geokrety:GK0001', handleGeokretUpdate)
subscribe('global:stats', handleGlobalUpdate)

// Unsubscribe when done
unsubscribe('user:123')

// Message format
interface WebSocketMessage {
  type: 'stats_update' | 'user_update' | 'geokrety_update'
  path: string
  data: unknown
  timestamp: string
}
```

### Error Handling

```typescript
const { onError, onReconnect } = useWebSocket()

// Handle connection errors
onError((error) => {
  console.error('WebSocket error:', error)
  // Show user-friendly message
})

// Handle reconnection
onReconnect(() => {
  console.log('Reconnected to server')
  // Re-subscribe to channels if needed
})
```

---

## Pull Request Guidelines

### Title Format

```
[component] Brief description of changes

Examples:
[ui] Add dark theme support
[api] Fix WebSocket reconnection
[stats] Fix user statistics display
[perf] Optimize stats rendering with memo
```

### Required Checks Before Submitting

```bash
# 1. Format code
pnpm format

# 2. Lint code
pnpm lint

# 3. Run tests
pnpm test

# 4. Build successfully
pnpm build

# 5. Check no console errors
# - Test in dev server manually
# - Check browser DevTools console
```

### Commit Message Format

Follow conventional commits:

```
[type]([scope]): [description]

[optional body]
[optional footer]

Types: feat, fix, docs, style, refactor, test, chore
Scopes: api, theme, components, websocket, etc.

Examples:
feat(websocket): add automatic reconnection logic
fix(theme): persist dark mode selection
docs(api): document WebSocket message format
```

### Review Expectations

- All tests passing
- > 80% code coverage for new code
- No console warnings or errors
- Responsive design verified (mobile, tablet, desktop)
- Accessibility checks (WCAG 2.1 AA minimum)
- No hardcoded URLs or secrets
- Changes comply with Vue 3 and TypeScript standards

---

## Component Development Guide

### Creating a New Statistics Component

```vue
<!-- UserStats.vue -->
<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import type { UserStats as UserStatsType } from '@/types/api'
import { useWebSocket } from '@/composables/useWebSocket'
import { useFetch } from '@/composables/useFetch'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface Props {
  userId: number
}

const props = defineProps<Props>()

const { data: stats, loading, error } = useFetch<UserStatsType>(`/users/${props.userId}/stats`)

const { subscribe, unsubscribe } = useWebSocket()

onMounted(async () => {
  // Subscribe to real-time updates
  subscribe(`user:${props.userId}`, (message) => {
    if (message.type === 'stats_update') {
      // Update local stats with new data
      Object.assign(stats.value, message.data)
    }
  })
})

onUnmounted(() => {
  unsubscribe(`user:${props.userId}`)
})

const totalPoints = computed(() => stats.value?.totalPoints ?? 0)
</script>

<template>
  <div class="space-y-4">
    <!-- Loading state -->
    <Skeleton v-if="loading" class="h-48 w-full rounded-lg" />

    <!-- Error state -->
    <Alert v-else-if="error" variant="destructive">
      <AlertDescription>{{ error.message }}</AlertDescription>
    </Alert>

    <!-- Stats card -->
    <Card v-else-if="stats" class="w-full">
      <CardHeader>
        <CardTitle>User Statistics</CardTitle>
        <CardDescription>Real-time movement data</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="grid gap-4 md:grid-cols-3">
          <div class="rounded-lg bg-muted p-4">
            <p class="text-sm text-muted-foreground">Total Points</p>
            <p class="text-2xl font-bold">{{ totalPoints }}</p>
          </div>
          <div class="rounded-lg bg-muted p-4">
            <p class="text-sm text-muted-foreground">Moves</p>
            <p class="text-2xl font-bold">{{ stats.movesCount }}</p>
          </div>
          <div class="rounded-lg bg-muted p-4">
            <p class="text-sm text-muted-foreground">Last Active</p>
            <p class="text-sm font-semibold">{{ stats.lastActive }}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
```

---

## Debugging and Troubleshooting

### Browser DevTools

**Vue DevTools:**

```bash
# Chrome Browser Extensions
# - Vue.js devtools (official)
# Inspect components, stores, and performance
```

**Common Issues:**

**API Connection Fails**

```bash
# Check API URL in .env
echo $VITE_API_URL

# Test endpoint manually
curl -i http://localhost:3000/api/users

# Check browser Network tab in DevTools
```

**WebSocket Connection Fails**

```bash
# Check WebSocket URL
echo $VITE_WS_URL

# Enable debug logging
VITE_ENABLE_DEBUG=true pnpm dev

# Check browser Console tab for connection messages
```

**Theme Not Persisting**

```bash
# Check localStorage in DevTools
localStorage.getItem('theme')

# Verify CSS variables are loaded
getComputedStyle(document.documentElement).getPropertyValue('--bs-primary')
```

### Debug Logging

Enable debug output:

```bash
# Set debug environment variable
VITE_ENABLE_DEBUG=true pnpm dev
```

View logs:

```typescript
// In components/composables
if (window.__DEBUG__) {
  console.log('Debug:', message)
}
```

### Performance Profiling

```bash
# Build with source maps for debugging
pnpm build --sourcemap

# Profile in DevTools Performance tab
# - Record performance trace
# - Check component rendering times
# - Monitor network requests
```

---

## Security Considerations

### API Security

- **HTTPS Only**: Ensure WebSocket uses WSS in production
- **CORS**: Configure API CORS headers properly
- **Authentication**: Implement token-based auth if needed
- **Input Validation**: Validate all user input
- **XSS Prevention**: By default, Vue escapes template content

### Secrets Management

```bash
# Environment variables
echo "VITE_API_KEY=secret" >> .env.local  # Not committed

# Public variables must be prefixed with VITE_
VITE_PUBLIC_VALUE=     # Safe in bundle
VITE_SECRET=           # Safe - won't be bundled
```

### Data Security

- Don't store sensitive data in localStorage
- Clear sensitive data on logout
- Use httpOnly cookies for auth tokens
- Validate all API responses
- Sanitize user-generated content

---

## Related Skills

- **[web-design-reviewer](/.github/skills/web-design-reviewer/SKILL.md)** - Visual design inspection and responsive testing
- **[multi-stage-dockerfile](/.github/skills/multi-stage-dockerfile/SKILL.md)** - Docker containerization

---

## Useful Links and References

- **Vue 3 Documentation**: https://vuejs.org/
- **Vue Router**: https://router.vuejs.org/
- **Pinia State Management**: https://pinia.vuejs.org/
- **shadcn/vue**: https://www.shadcn-vue.com/
- **Tailwind CSS**: https://tailwindcss.com/
- **Vite**: https://vitejs.dev/
- **TypeScript Handbook**: https://www.typescriptlang.org/docs/
- **Vue 3 Instructions**: [/.github/instructions/vuejs3.instructions.md](/.github/instructions/vuejs3.instructions.md)
- **geokrety-stats-api**: https://github.com/geokrety/geokrety-stats-api

---

## Additional Notes

### Contributing

1. Create a feature branch: `git checkout -b [component]/[description]`
2. Follow code style guidelines
3. Write tests for new functionality
4. Test theme switching (light/dark)
5. Test WebSocket reconnection scenarios
6. Submit PR with clear description
7. Ensure all CI checks pass

### Performance Optimization Tips

- **Lazy Load Routes**: Use dynamic imports for route components
- **Code Splitting**: Vite handles this automatically
- **Component Memoization**: Use `v-memo` for expensive components
- **Image Optimization**: Use modern formats (WebP) with fallbacks
- **Debounce Search**: Debounce API calls during typing
- **WebSocket Batching**: Batch multiple updates into single message

### Known Limitations

- WebSocket updates only apply to currently visible page
- Theme switching requires page reload (or implement state preservation)
- API calls timeout after 10 seconds (configurable via VITE_API_TIMEOUT)

### Future Enhancements

- [ ] Implement offline support with Service Workers
- [ ] Add PWA manifest for installable app
- [ ] Support for custom theme colors
- [ ] Real-time search with WebSocket
- [ ] Export statistics to CSV/PDF
- [ ] Advanced filtering and sorting options
