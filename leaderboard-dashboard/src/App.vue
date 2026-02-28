<script setup>
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { ref, computed } from 'vue'
import { useLiveStats } from './composables/useWebSocket.js'

const route = useRoute()
const { connected, stats } = useLiveStats()

const isActive = (path) => route.path === path || route.path.startsWith(path + '/')
</script>

<template>
  <!-- Navbar -->
  <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
    <div class="container-fluid">
      <RouterLink class="navbar-brand fw-bold" to="/">
        <i class="bi bi-geo-alt-fill text-warning me-1"></i> GeoKrety Leaderboard
      </RouterLink>
      <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navmenu">
        <span class="navbar-toggler-icon"></span>
      </button>
      <div class="collapse navbar-collapse" id="navmenu">
        <ul class="navbar-nav me-auto">
          <li class="nav-item">
            <RouterLink class="nav-link" :class="{ active: isActive('/') && route.path === '/' }" to="/">
              <i class="bi bi-trophy me-1"></i>Leaderboard
            </RouterLink>
          </li>
          <li class="nav-item">
            <RouterLink class="nav-link" :class="{ active: isActive('/stats') }" to="/stats">
              <i class="bi bi-graph-up me-1"></i>Statistics
            </RouterLink>
          </li>
        </ul>
        <span class="navbar-text small">
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
  <main class="container-xl py-3">
    <RouterView />
  </main>

  <!-- Footer -->
  <footer class="bg-dark text-secondary text-center py-2 small mt-4">
    GeoKrety Points System &mdash; Data refreshes live via WebSocket
  </footer>
</template>

<style>
body { background-color: #f8f9fa; }
.navbar-brand { font-size: 1.1rem; }
</style>

