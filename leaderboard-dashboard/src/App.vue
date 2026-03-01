<script setup>
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { ref, onMounted, watch } from 'vue'
import { useLiveStats } from './composables/useWebSocket.js'
import VTooltip from './components/VTooltip.vue'

const route = useRoute()
const { connected, enabled, stats, connectedUsers, lastUpdate, toggleEnabled } = useLiveStats()
const THEME_KEY = 'gk-dashboard-theme'

const navMenu = ref(null)
const currentTheme = ref('light')
const pulseEffect = ref(false)
const previousMoveCount = ref(null)

// Trigger visual pulse only when move count actually changes
watch(() => stats.value?.total_moves, (newVal) => {
  if (newVal !== null && newVal !== previousMoveCount.value) {
    previousMoveCount.value = newVal
    pulseEffect.value = true
    setTimeout(() => { pulseEffect.value = false }, 1000)
  }
})

const isActive = (path) => route.path === path || route.path.startsWith(path + '/')

const closeMenu = () => {
  const collapseElement = document.getElementById('navmenu')
  if (collapseElement && collapseElement.classList.contains('show')) {
    const bsCollapse = window.bootstrap.Collapse.getInstance(collapseElement) || new window.bootstrap.Collapse(collapseElement)
    bsCollapse.hide()
  }
}

const applyTheme = (theme) => {
  currentTheme.value = theme
  document.documentElement.setAttribute('data-bs-theme', theme)
  localStorage.setItem(THEME_KEY, theme)
}

const toggleTheme = () => {
  const nextTheme = currentTheme.value === 'dark' ? 'light' : 'dark'
  applyTheme(nextTheme)
}

onMounted(() => {
  const savedTheme = localStorage.getItem(THEME_KEY)
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  const initialTheme = savedTheme || (prefersDark ? 'dark' : 'light')
  applyTheme(initialTheme)
})
</script>

<template>
  <!-- Navbar -->
  <nav class="navbar app-navbar navbar-expand-lg sticky-top shadow-sm" :class="currentTheme === 'dark' ? 'navbar-dark' : 'navbar-light'">
    <div class="container-fluid px-2 px-md-4">
      <RouterLink class="navbar-brand fw-bold" to="/" @click="closeMenu">
        <i class="bi bi-geo-alt-fill text-warning me-1"></i> GeoKrety Leaderboard
      </RouterLink>
      <button
        class="navbar-toggler border-0"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#navmenu"
        aria-controls="navmenu"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navmenu" ref="navMenu">
        <ul class="navbar-nav me-auto">
          <li class="nav-item">
            <RouterLink class="nav-link px-3" :class="{ active: isActive('/') && route.path === '/' }" to="/" @click="closeMenu">
              <i class="bi bi-trophy me-1"></i>Leaderboard
            </RouterLink>
          </li>
          <li class="nav-item">
            <RouterLink class="nav-link px-3" :class="{ active: isActive('/geokrety') && route.path.startsWith('/geokrety') && !route.params.id }" to="/geokrety" @click="closeMenu">
              <i class="bi bi-gift me-1"></i>GeoKrety
            </RouterLink>
          </li>
          <li class="nav-item">
            <RouterLink class="nav-link px-3" :class="{ active: isActive('/countries') }" to="/countries" @click="closeMenu">
              <i class="bi bi-globe me-1"></i>Countries
            </RouterLink>
          </li>
          <li class="nav-item">
            <RouterLink class="nav-link px-3" :class="{ active: isActive('/stats') }" to="/stats" @click="closeMenu">
              <i class="bi bi-graph-up me-1"></i>Statistics
            </RouterLink>
          </li>
        </ul>
        <span class="navbar-text small px-3">
          <button
            type="button"
            class="btn btn-sm theme-toggle-btn me-2"
            :class="currentTheme === 'dark' ? 'btn-outline-warning' : 'btn-outline-secondary'"
            :title="`Switch to ${currentTheme === 'dark' ? 'light' : 'dark'} theme`"
            @click="toggleTheme"
          >
            <i class="bi" :class="currentTheme === 'dark' ? 'bi-sun-fill' : 'bi-moon-stars-fill'"></i>
          </button>

          <v-tooltip :text="`Real-time sync: ${enabled ? 'Enabled' : 'Disabled'}. Click to toggle.`">
            <template #activator="{ props }">
              <span
                v-bind="props"
                class="me-2 cursor-pointer"
                style="cursor: pointer;"
                @click="toggleEnabled"
              >
                <span v-if="connected" class="text-success fw-bold">
                  <i class="bi bi-broadcast me-1 pulse-wifi"></i>Live
                </span>
                <span v-else-if="!enabled" class="text-muted opacity-50">
                  <i class="bi bi-broadcast me-1"></i>Offline
                </span>
                <span v-else class="text-secondary small">
                  <i class="bi bi-wifi-off me-1"></i>Connecting...
                </span>
              </span>
            </template>
          </v-tooltip>

          <template v-if="stats">
            &ensp;|&ensp;
            <v-tooltip>
              <template #activator="{ props }">
                <span
                  v-bind="props"
                  class="navbar-stats-value"
                  :class="{'stats-update-pulse': pulseEffect}"
                >
                  <i class="bi bi-activity text-info me-1"></i>
                  <span class="fw-semibold">{{ stats.total_moves?.toLocaleString() }}</span>
                  <span class="d-none d-md-inline">&nbsp;moves</span>
                </span>
              </template>
              <div class="text-start p-2" style="min-width: 200px">
                <div class="d-flex justify-content-between align-items-center mb-1 pb-1 border-bottom border-white border-opacity-25">
                  <span class="small fw-bold"><i class="bi bi-speedometer2 me-1"></i>Site Real-time Stats</span>
                </div>
                <div class="d-flex justify-content-between my-2">
                   <span class="small opacity-75">Users Connected:</span>
                   <span class="badge bg-success p-1 ms-3">{{ connectedUsers }}</span>
                </div>
                <div class="d-flex justify-content-between my-2">
                   <span class="small opacity-75">Total Points:</span>
                   <span class="small fw-bold ms-3">{{ stats.total_points_awarded?.toLocaleString() }}</span>
                </div>
                <div class="d-flex justify-content-between my-2">
                   <span class="small opacity-75">GKs Tracked:</span>
                   <span class="small fw-bold ms-3">{{ stats.total_gks?.toLocaleString() }}</span>
                </div>
                <div class="d-flex justify-content-between mt-1 pt-1 opacity-50 x-small fst-italic border-top border-white border-opacity-25">
                   Last update: {{ new Date(lastUpdate).toLocaleTimeString() }}
                </div>
              </div>
            </v-tooltip>
          </template>
        </span>
      </div>
    </div>
  </nav>

  <!-- Main content -->
  <main class="app-main container-fluid py-3 px-1 px-md-4">
    <RouterView />
  </main>

  <!-- Footer -->
  <footer class="app-footer text-center py-2 small mt-4">
    GeoKrety Points System &mdash;
    <span v-if="connected" class="text-success">
      <i class="bi bi-people-fill me-1"></i>{{ connectedUsers }} user{{ connectedUsers !== 1 ? 's' : '' }} online
    </span>
    <span v-else class="text-secondary">
      Data refreshes when connection restored
    </span>
  </footer>
</template>

