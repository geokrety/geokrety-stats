<script setup>
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { ref, onMounted } from 'vue'
import { useLiveStats } from './composables/useWebSocket.js'

const route = useRoute()
const { connected, stats, connectedUsers } = useLiveStats()
const THEME_KEY = 'gk-dashboard-theme'

const navMenu = ref(null)
const currentTheme = ref('light')

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
  applyTheme(currentTheme.value === 'dark' ? 'light' : 'dark')
}

onMounted(() => {
  const htmlTheme = document.documentElement.getAttribute('data-bs-theme')
  currentTheme.value = htmlTheme === 'dark' ? 'dark' : 'light'
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
            class="btn btn-sm btn-outline-secondary theme-toggle-btn me-2"
            :aria-label="`Switch to ${currentTheme === 'dark' ? 'light' : 'dark'} theme`"
            @click="toggleTheme"
          >
            <i class="bi" :class="currentTheme === 'dark' ? 'bi-sun-fill' : 'bi-moon-stars-fill'"></i>
          </button>
          <span v-if="connected" class="text-success">
            <i class="bi bi-wifi me-1"></i>Live
          </span>
          <span v-else class="text-secondary">
            <i class="bi bi-wifi-off me-1"></i>Offline
          </span>
          <template v-if="stats">
            &ensp;|&ensp;
            <span class="text-light">{{ stats.total_moves?.toLocaleString() }} moves</span>
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

