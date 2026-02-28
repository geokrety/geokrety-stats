<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const waypoint = ref(route.params.waypoint || '')
const loading = ref(true)
const error = ref(null)
const mapContainer = ref(null)
const map = ref(null)

const loadLeaflet = () => {
  return new Promise((resolve, reject) => {
    if (window.L) {
      resolve()
      return
    }

    // Load Leaflet CSS
    const link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css'
    link.integrity = 'sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY='
    link.crossOrigin = ''
    document.head.appendChild(link)

    // Load Leaflet JS
    const script = document.createElement('script')
    script.src = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.js'
    script.integrity = 'sha256-20nQCchB9co0qIjJZRGuk2/Z9VM+kNiyxNV1lvTlZBo='
    script.crossOrigin = ''
    script.onload = resolve
    script.onerror = reject
    document.head.appendChild(script)
  })
}

const initMap = async () => {
  loading.value = true
  error.value = null

  try {
    await loadLeaflet()

    // Initialize map centered on Europe (default position)
    map.value = window.L.map(mapContainer.value).setView([50.0, 14.0], 5)

    // Add OpenStreetMap tiles
    window.L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      maxZoom: 19
    }).addTo(map.value)

    // If waypoint is provided, try to geocode it
    if (waypoint.value) {
      await searchWaypoint(waypoint.value)
    }

    loading.value = false
  } catch (e) {
    error.value = 'Failed to load map: ' + e.message
    loading.value = false
  }
}

const searchWaypoint = async (wp) => {
  if (!wp || !map.value) return

  loading.value = true
  error.value = null

  try {
    // Use Nominatim geocoding to find the waypoint location
    const response = await fetch(
      `https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(wp)}&format=json&limit=1`,
      {
        headers: {
          'User-Agent': 'GeoKrety-Leaderboard/1.0'
        }
      }
    )

    if (!response.ok) {
      throw new Error('Geocoding service unavailable')
    }

    const results = await response.json()

    if (results && results.length > 0) {
      const { lat, lon, display_name } = results[0]
      const latLng = [parseFloat(lat), parseFloat(lon)]

      // Center map on location
      map.value.setView(latLng, 13)

      // Add marker
      window.L.marker(latLng).addTo(map.value)
        .bindPopup(`<b>${wp}</b><br>${display_name}`)
        .openPopup()

      error.value = null
    } else {
      error.value = `Waypoint "${wp}" not found. Showing default view.`
    }
  } catch (e) {
    error.value = 'Failed to search waypoint: ' + e.message
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  if (waypoint.value) {
    router.push(`/map/${encodeURIComponent(waypoint.value)}`)
    searchWaypoint(waypoint.value)
  }
}

onMounted(initMap)
</script>

<template>
  <div>
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">Waypoint Map</li>
      </ol>
    </nav>

    <!-- Header -->
    <div class="card mb-3 shadow-sm">
      <div class="card-body">
        <h4 class="mb-2">🗺️ Waypoint Map Viewer</h4>
        <p class="text-muted mb-2">Enter a waypoint code (e.g., GC1A2B3, OK1234) to view on map</p>

        <!-- Search Form -->
        <div class="input-group">
          <input
            v-model="waypoint"
            type="text"
            class="form-control"
            placeholder="Enter waypoint code..."
            @keyup.enter="handleSearch"
          >
          <button class="btn btn-primary" type="button" @click="handleSearch">
            <i class="bi bi-search"></i> Search
          </button>
          <a
            v-if="waypoint"
            :href="`https://geokrety.org/go2geo/${encodeURIComponent(waypoint)}`"
            target="_blank"
            rel="noopener"
            class="btn btn-outline-secondary"
            title="Open on GeoKrety.org"
          >
            <i class="bi bi-box-arrow-up-right"></i> GeoKrety.org
          </a>
        </div>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="alert alert-warning alert-dismissible fade show" role="alert">
      <i class="bi bi-exclamation-triangle me-2"></i>{{ error }}
      <button type="button" class="btn-close" @click="error = null"></button>
    </div>

    <!-- Map Container -->
    <div class="card shadow-sm">
      <div class="card-body p-0 position-relative">
        <div v-if="loading && !map" class="position-absolute top-50 start-50 translate-middle" style="z-index: 1000;">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Loading map...</span>
          </div>
        </div>
        <div ref="mapContainer" style="height: 600px; width: 100%;"></div>
      </div>
    </div>

    <!-- Help Text -->
    <div class="mt-3 text-muted small">
      <p class="mb-1">
        <i class="bi bi-info-circle me-1"></i>
        This map uses OpenStreetMap and Nominatim geocoding to locate waypoints.
        Not all waypoint codes may be found in the geocoding database.
      </p>
      <p class="mb-0">
        For the most accurate results, visit <a href="https://geokrety.org" target="_blank">GeoKrety.org</a>
        or the original geocaching platform.
      </p>
    </div>
  </div>
</template>

<style scoped>
/* Ensure map fills container properly */
.card-body {
  border-radius: 0.25rem;
  overflow: hidden;
}
</style>
