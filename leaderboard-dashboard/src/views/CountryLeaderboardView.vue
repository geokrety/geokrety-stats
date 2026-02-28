<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { getCountryFlag } from '../composables/useCountryFlags.js'

const countries = ref([])
const loading = ref(false)
const error = ref(null)
const sortBy = ref('points')
const viewMode = ref('cards') // 'cards' or 'table'

const sortOptions = [
  { value: 'points', label: 'By Points' },
  { value: 'moves', label: 'By Moves' },
  { value: 'users', label: 'By Users' },
  { value: 'geokrety', label: 'By GeoKrety' }
]

async function loadCountries() {
  loading.value = true
  error.value = null
  try {
    const response = await fetch('/api/v1/stats/countries')
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const { data } = await response.json()

    // Sort based on selection
    const sorted = (data || []).sort((a, b) => {
      switch (sortBy.value) {
        case 'points':
          return (b.total_points_awarded || 0) - (a.total_points_awarded || 0)
        case 'moves':
          return b.move_count - a.move_count
        case 'users':
          return b.user_count - a.user_count
        case 'geokrety':
          return b.gk_count - a.gk_count
        default:
          return 0
      }
    })
    countries.value = sorted
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadCountries)

const handleSortChange = () => {
  loadCountries()
}

const getMoveTypeIcon = (type) => {
  const icons = {
    drops: '📦',
    grabs: '🎯',
    dips: '💧',
    comments: '💬',
    sees: '👁️'
  }
  return icons[type] || '•'
}
</script>

<template>
  <div>
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-3">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">Countries</li>
      </ol>
    </nav>

    <!-- Header -->
    <div class="card mb-4 shadow-sm">
      <div class="card-body">
        <h2 class="mb-1">🌍 Countries Leaderboard</h2>
        <p class="text-muted mb-0">View statistics by country</p>
      </div>
    </div>

    <!-- Sort Options -->
    <div class="mb-3 d-flex gap-2">
      <span class="text-muted">Sort by:</span>
      <div class="btn-group" role="group">
        <input
          type="radio"
          class="btn-check"
          name="sortBtnradio"
          id="sortPoints"
          value="points"
          v-model="sortBy"
          @change="handleSortChange"
        >
        <label class="btn btn-outline-primary" for="sortPoints">Points</label>

        <input
          type="radio"
          class="btn-check"
          name="sortBtnradio"
          id="sortMoves"
          value="moves"
          v-model="sortBy"
          @change="handleSortChange"
        >
        <label class="btn btn-outline-primary" for="sortMoves">Moves</label>

        <input
          type="radio"
          class="btn-check"
          name="sortBtnradio"
          id="sortUsers"
          value="users"
          v-model="sortBy"
          @change="handleSortChange"
        >
        <label class="btn btn-outline-primary" for="sortUsers">Users</label>

        <input
          type="radio"
          class="btn-check"
          name="sortBtnradio"
          id="sortGeokrety"
          value="geokrety"
          v-model="sortBy"
          @change="handleSortChange"
        >
        <label class="btn btn-outline-primary" for="sortGeokrety">GeoKrety</label>
      </div>
    </div>

    <!-- Loading / Error / Content -->
    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border"></div>
    </div>
    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-else-if="countries.length === 0" class="alert alert-info">No countries data available.</div>
    <div v-else>
      <!-- View Mode Selector -->
      <div class="mb-3">
        <div class="btn-group" role="group">
          <input
            type="radio"
            class="btn-check"
            name="viewMode"
            id="modeCards"
            value="cards"
            v-model="viewMode"
          >
          <label class="btn btn-outline-secondary" for="modeCards">
            <i class="bi bi-card-heading"></i> Cards
          </label>

          <input
            type="radio"
            class="btn-check"
            name="viewMode"
            id="modeTable"
            value="table"
            v-model="viewMode"
          >
          <label class="btn btn-outline-secondary" for="modeTable">
            <i class="bi bi-table"></i> Table
          </label>
        </div>
      </div>

      <!-- Cards View -->
      <div v-if="viewMode === 'cards'" class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-3">
        <div v-for="(country, idx) in countries" :key="country.country" class="col">
          <div class="card h-100 shadow-sm border-0">
            <div class="card-header bg-light d-flex align-items-center gap-2 border-0">
              <span class="fs-3">{{ getCountryFlag(country.country) }}</span>
              <div>
                <div class="fw-bold">{{ country.country.toUpperCase() }}</div>
                <div class="text-muted small">#{{ idx + 1 }}</div>
              </div>
            </div>
            <div class="card-body">
              <!-- Key Stats -->
              <div class="row g-2 mb-3">
                <div class="col-6">
                  <small class="text-muted d-block">Points</small>
                  <div class="fw-bold fs-5 text-success">{{ country.total_points_awarded?.toLocaleString() }}</div>
                </div>
                <div class="col-6">
                  <small class="text-muted d-block">Moves</small>
                  <div class="fw-bold fs-5">{{ country.total_moves?.toLocaleString() }}</div>
                </div>
                <div class="col-6">
                  <small class="text-muted d-block">GeoKrety</small>
                  <div class="fw-bold fs-5">{{ country.unique_gks?.toLocaleString() }}</div>
                </div>
                <div class="col-6">
                  <small class="text-muted d-block">Users</small>
                  <div class="fw-bold fs-5">{{ country.unique_users?.toLocaleString() }}</div>
                </div>
              </div>

              <!-- Move Type Breakdown -->
              <div class="border-top pt-2">
                <small class="text-muted d-block mb-2">Move Types</small>
                <div class="row g-1 small">
                  <div class="col-6">
                    <span class="me-1">{{ getMoveTypeIcon('drops') }}</span>
                    <span class="text-muted">{{ country.drops || 0 }}</span>
                  </div>
                  <div class="col-6">
                    <span class="me-1">{{ getMoveTypeIcon('grabs') }}</span>
                    <span class="text-muted">{{ country.grabs || 0 }}</span>
                  </div>
                  <div class="col-6">
                    <span class="me-1">{{ getMoveTypeIcon('dips') }}</span>
                    <span class="text-muted">{{ country.dips || 0 }}</span>
                  </div>
                  <div class="col-6">
                    <span class="me-1">{{ getMoveTypeIcon('sees') }}</span>
                    <span class="text-muted">{{ country.sees || 0 }}</span>
                  </div>
                  <div class="col-12" v-if="country.comments">
                    <span class="me-1">{{ getMoveTypeIcon('comments') }}</span>
                    <span class="text-muted">{{ country.comments }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Table View -->
      <div v-else-if="viewMode === 'table'" class="table-responsive">
        <table class="table table-hover table-sm align-middle border">
          <thead class="table-light sticky-top">
            <tr>
              <th style="width: 60px">Rank</th>
              <th>Country</th>
              <th class="text-end">Points</th>
              <th class="text-end">Moves</th>
              <th class="text-end">📦</th>
              <th class="text-end">🎯</th>
              <th class="text-end">💧</th>
              <th class="text-end">👁️</th>
              <th class="text-end">GeoKrety</th>
              <th class="text-end">Users</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(country, idx) in countries" :key="country.country">
              <td class="fw-bold">
                <span v-if="idx < 3" class="badge" :class="idx === 0 ? 'bg-warning text-dark' : idx === 1 ? 'bg-secondary' : 'bg-info'">
                  {{ idx + 1 }}
                </span>
                <span v-else>{{ idx + 1 }}</span>
              </td>
              <td>
                <span class="me-2">{{ getCountryFlag(country.country) }}</span>
                <strong>{{ country.country.toUpperCase() }}</strong>
              </td>
              <td class="text-end fw-bold text-success">{{ country.total_points_awarded?.toLocaleString() }}</td>
              <td class="text-end">{{ country.total_moves?.toLocaleString() }}</td>
              <td class="text-end text-muted small">{{ country.drops || 0 }}</td>
              <td class="text-end text-muted small">{{ country.grabs || 0 }}</td>
              <td class="text-end text-muted small">{{ country.dips || 0 }}</td>
              <td class="text-end text-muted small">{{ country.sees || 0 }}</td>
              <td class="text-end">{{ country.unique_gks?.toLocaleString() }}</td>
              <td class="text-end">{{ country.unique_users?.toLocaleString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
