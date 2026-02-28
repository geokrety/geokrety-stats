<script setup>
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import { getMoveTypeTooltip } from '../composables/useMoveTypeColors.js'

const countriesRaw = ref([])
const loading = ref(false)
const error = ref(null)
const sortBy = ref('points')
const viewMode = ref('cards') // 'cards' or 'table'

const sortOptions = [
  { value: 'points', label: 'Points' },
  { value: 'avg_points', label: 'Avg Points/Move' },
  { value: 'moves', label: 'Total Moves' },
  { value: 'users', label: 'Active Users' },
  { value: 'gks', label: 'GeoKrety Count' },
  { value: 'grabs', label: 'Grabs' },
  { value: 'drops', label: 'Drops' },
  { value: 'dips', label: 'DIPs' },
  { value: 'loves', label: '❤️ Loves' }
]

// Computed property for sorted countries (avoids re-fetching API)
const countries = computed(() => {
  if (!countriesRaw.value.length) return []

  return [...countriesRaw.value].sort((a, b) => {
    switch (sortBy.value) {
      case 'points':
        return (b.total_points_awarded || 0) - (a.total_points_awarded || 0)
      case 'avg_points':
        return (b.avg_points_per_move || 0) - (a.avg_points_per_move || 0)
      case 'moves':
        return (b.total_moves || 0) - (a.total_moves || 0)
      case 'users':
        return (b.unique_users || 0) - (a.unique_users || 0)
      case 'gks':
        return (b.unique_gks || 0) - (a.unique_gks || 0)
      case 'grabs':
        return (b.grabs || 0) - (a.grabs || 0)
      case 'drops':
        return (b.drops || 0) - (a.drops || 0)
      case 'dips':
        return (b.dips || 0) - (a.dips || 0)
      case 'loves':
        return (b.total_loves || 0) - (a.total_loves || 0)
      default:
        return 0
    }
  })
})

async function loadCountries() {
  loading.value = true
  error.value = null
  try {
    const response = await fetch('/api/v1/stats/countries')
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const { data } = await response.json()
    countriesRaw.value = data || []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadCountries)

const getMoveTypeIcon = (type) => {
  const icons = {
    drops: '🌳',
    grabs: '🚀',
    dips: '🥾',
    comments: '💬',
    sees: '👀'
  }
  return icons[type] || '•'
}

const formatInt = (num) => {
  if (!num) return '0'
  return Math.round(num).toLocaleString()
}

const formatFloat = (num, decimals = 2) => {
  if (!num) return '0'
  return (Math.round(num * Math.pow(10, decimals)) / Math.pow(10, decimals)).toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  })
}

</script>

<template>
  <div>
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-2">
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
    <div class="mb-3 d-flex gap-2 flex-wrap align-items-center">
      <span class="text-muted small fw-semibold">Sort by:</span>
      <div class="btn-group flex-wrap" role="group">
        <template v-for="opt in sortOptions" :key="opt.value">
          <input
            type="radio"
            class="btn-check"
            name="sortBtnradio"
            :id="'sort' + opt.value"
            :value="opt.value"
            v-model="sortBy"
          >
          <label class="btn btn-outline-primary btn-sm" :for="'sort' + opt.value">{{ opt.label }}</label>
        </template>
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
      <div class="mb-2">
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
              <div style="flex: 1">
                <RouterLink :to="`/country/${country.country}`" class="text-decoration-none" style="cursor: pointer">
                  <div class="fw-bold">{{ country.country.toUpperCase() }}</div>
                </RouterLink>
                <div class="text-muted small">#{{ idx + 1 }}</div>
              </div>
            </div>
            <div class="card-body">
              <!-- Key Stats -->
              <div class="row g-2 mb-2">
                <div class="col-6">
                  <small class="text-muted d-block">Points</small>
                  <div class="fw-bold fs-5 text-success">{{ formatInt(country.total_points_awarded) }}</div>
                </div>
                <div class="col-6">
                  <small class="text-muted d-block">Avg/Move</small>
                  <div class="fw-bold fs-6 text-info">{{ formatFloat(country.avg_points_per_move, 4) }}</div>
                </div>
                <div class="col-6">
                  <small class="text-muted d-block">Moves</small>
                  <div class="fw-bold fs-5">{{ formatInt(country.total_moves) }}</div>
                </div>
                <div class="col-6">
                  <small class="text-muted d-block">GeoKrety</small>
                  <div class="fw-bold fs-5">{{ formatInt(country.unique_gks) }}</div>
                </div>
                <div class="col-6">
                  <small class="text-muted d-block">Users</small>
                  <div class="fw-bold fs-5">{{ formatInt(country.unique_users) }}</div>
                </div>
                <div class="col-6">
                  <small class="text-muted d-block" title="Total loves given to GeoKrety that visited this country">❤️ Loves</small>
                  <div class="fw-bold fs-5 text-danger">{{ formatInt(country.total_loves) }}</div>
                </div>
              </div>

              <!-- Move Type Breakdown -->
              <div class="border-top pt-2">
                <small class="text-muted d-block mb-2">Move Types</small>
                <div class="row g-1 small">
                  <div class="col-6" :title="getMoveTypeTooltip('drop')">
                    <span class="me-1">{{ getMoveTypeIcon('drops') }}</span>
                    <span class="text-muted">{{ formatInt(country.drops) }}</span>
                  </div>
                  <div class="col-6" :title="getMoveTypeTooltip('grab')">
                    <span class="me-1">{{ getMoveTypeIcon('grabs') }}</span>
                    <span class="text-muted">{{ formatInt(country.grabs) }}</span>
                  </div>
                  <div class="col-6" :title="getMoveTypeTooltip('dip')">
                    <span class="me-1">{{ getMoveTypeIcon('dips') }}</span>
                    <span class="text-muted">{{ formatInt(country.dips) }}</span>
                  </div>
                  <div class="col-6" :title="getMoveTypeTooltip('seen')">
                    <span class="me-1">{{ getMoveTypeIcon('sees') }}</span>
                    <span class="text-muted">{{ formatInt(country.seen) }}</span>
                  </div>
                  <div class="col-12" v-if="country.comments" :title="getMoveTypeTooltip('comments')">
                    <span class="me-1">{{ getMoveTypeIcon('comments') }}</span>
                    <span class="text-muted">{{ formatInt(country.comments) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Table View -->
      <div v-else-if="viewMode === 'table'" class="table-responsive border-0 mb-0">
        <table class="table table-hover table-sm align-middle border">
          <thead class="table-light sticky-top">
            <tr>
              <th style="width: 60px" title="Rank position based on current sort">#</th>
              <th title="Country name">Country</th>
              <th class="text-end" title="Total points earned in this country">Points</th>
              <th class="text-end d-none d-md-table-cell text-nowrap" title="Average points earned per move in this country">Avg/Move</th>
              <th class="text-end d-none d-sm-table-cell" title="Total number of recorded moves in this country">Moves</th>
              <th class="text-end d-none d-lg-table-cell" :title="getMoveTypeTooltip('drop')">🌳</th>
              <th class="text-end d-none d-lg-table-cell" :title="getMoveTypeTooltip('grab')">🚀</th>
              <th class="text-end d-none d-lg-table-cell" :title="getMoveTypeTooltip('dip')">🥾</th>
              <th class="text-end d-none d-lg-table-cell" :title="getMoveTypeTooltip('seen')">👀</th>
              <th class="text-end d-none d-md-table-cell" title="Number of distinct GeoKrety that visited this country">GeoKrety</th>
              <th class="text-end d-none d-md-table-cell" title="Number of distinct users who made moves in this country">Users</th>
              <th class="text-end" title="Total loves given to GeoKrety that visited this country">❤️</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(country, idx) in countries" :key="country.country" @click="$router.push(`/country/${country.country}`)" style="cursor: pointer">
              <td class="fw-bold">
                <span v-if="idx < 3" class="badge" :class="idx === 0 ? 'bg-warning text-dark' : idx === 1 ? 'bg-secondary' : 'bg-info'">
                  {{ idx + 1 }}
                </span>
                <span v-else class="text-muted small">{{ idx + 1 }}</span>
              </td>
              <td>
                <div class="d-flex align-items-center gap-2">
                  <span class="fs-5">{{ getCountryFlag(country.country) }}</span>
                  <strong class="text-truncate" style="max-width: 120px">{{ country.country.toUpperCase() }}</strong>
                </div>
              </td>
              <td class="text-end fw-bold text-success">{{ formatInt(country.total_points_awarded) }}</td>
              <td class="text-end fw-bold text-info small d-none d-md-table-cell">{{ formatFloat(country.avg_points_per_move, 3) }}</td>
              <td class="text-end d-none d-sm-table-cell">{{ formatInt(country.total_moves) }}</td>
              <td class="text-end text-muted small d-none d-lg-table-cell">{{ formatInt(country.drops) }}</td>
              <td class="text-end text-muted small d-none d-lg-table-cell">{{ formatInt(country.grabs) }}</td>
              <td class="text-end text-muted small d-none d-lg-table-cell">{{ formatInt(country.dips) }}</td>
              <td class="text-end text-muted small d-none d-lg-table-cell">{{ formatInt(country.seen) }}</td>
              <td class="text-end d-none d-md-table-cell">{{ formatInt(country.unique_gks) }}</td>
              <td class="text-end d-none d-md-table-cell">{{ formatInt(country.unique_users) }}</td>
              <td class="text-end text-danger fw-semibold">{{ formatInt(country.total_loves) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
