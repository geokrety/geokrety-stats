<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import { countryName } from '../composables/useCountry.js'
import { getMoveTypeTooltip } from '../composables/useMoveTypeColors.js'
import PointsValue from '../components/PointsValue.vue'

const countriesRaw = ref([])
const loading = ref(false)
const error = ref(null)
const sortBy = ref('points')
const viewMode = ref('cards') // 'cards' or 'table'

// Hash handling
const updateHash = () => {
  window.location.hash = viewMode.value
}

const loadHash = () => {
  const hash = window.location.hash.replace('#', '')
  if (hash === 'cards' || hash === 'table') {
    viewMode.value = hash
  }
}

watch(viewMode, updateHash)

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

// Computed property for sorted countries (now handle sort on server side)
const countries = computed(() => countriesRaw.value)

async function loadCountries() {
  loading.value = true
  error.value = null
  try {
    const response = await fetch(`/api/v1/stats/countries?sort=${sortBy.value}`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const { data } = await response.json()
    countriesRaw.value = data || []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadHash()
  loadCountries()
})
watch(sortBy, loadCountries)

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

function headerSort(field) {
  if (sortBy.value === field) return
  sortBy.value = field
}

function sortIcon(field) {
  return sortBy.value === field ? 'bi-sort-down-alt text-primary' : 'bi-sort-down text-muted'
}

</script>

<template>
  <div class="country-leaderboard">
    <style scoped>
    .country-card {
      transition: transform 0.2s, box-shadow 0.2s;
    }
    .country-card:hover {
      transform: translateY(-5px);
      box-shadow: 0 10px 20px rgba(0,0,0,0.1) !important;
    }
    .ls-1 { letter-spacing: 0.5px; }
    .x-small { font-size: 0.72rem; }
    </style>
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
      <div v-if="viewMode === 'cards'" class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-4">
        <div v-for="(country, idx) in countries" :key="country.country" class="col">
          <div class="card h-100 shadow border-0 overflow-hidden country-card">
            <div class="card-header bg-primary bg-opacity-10 d-flex flex-column align-items-center py-3 border-0">
              <span class="display-4 mb-1">{{ getCountryFlag(country.country) }}</span>
              <RouterLink :to="`/country/${country.country}`" class="text-decoration-none text-dark" style="cursor: pointer">
                <h4 class="fw-bold mb-0">{{ countryName(country.country) }}</h4>
              </RouterLink>
              <div class="badge bg-secondary mt-1">Rank #{{ idx + 1 }}</div>
            </div>
            <div class="card-body">
              <!-- Key Stats -->
              <div class="row g-3 text-center mb-4">
                <div class="col-6 border-end">
                  <small class="text-muted d-block text-uppercase small ls-1">Points</small>
                  <div class="fw-bold fs-4 text-primary"><PointsValue :value="country.total_points_awarded" :digits="0" /></div>
                </div>
                <div class="col-6">
                  <small class="text-muted d-block text-uppercase small ls-1">Avg/Move</small>
                  <div class="fw-bold fs-5 text-info"><PointsValue :value="country.avg_points_per_move" /></div>
                </div>
                <div class="col-4 border-end">
                  <small class="text-muted d-block x-small">Moves</small>
                  <div class="fw-semibold">{{ formatInt(country.total_moves) }}</div>
                </div>
                <div class="col-4 border-end">
                  <small class="text-muted d-block x-small">GeoKrety</small>
                  <div class="fw-semibold">{{ formatInt(country.unique_gks) }}</div>
                </div>
                <div class="col-4">
                  <small class="text-muted d-block x-small">Users</small>
                  <div class="fw-semibold">{{ formatInt(country.unique_users) }}</div>
                </div>
              </div>

              <!-- Move Type Breakdown -->
              <div class="bg-light rounded p-2 border">
                <div class="d-flex justify-content-between flex-wrap gap-2 px-1">
                  <div :title="getMoveTypeTooltip('drop')" data-bs-toggle="tooltip">
                    <span class="me-1">{{ getMoveTypeIcon('drops') }}</span>
                    <span class="fw-bold">{{ formatInt(country.drops) }}</span>
                  </div>
                  <div :title="getMoveTypeTooltip('grab')" data-bs-toggle="tooltip">
                    <span class="me-1">{{ getMoveTypeIcon('grabs') }}</span>
                    <span class="fw-bold">{{ formatInt(country.grabs) }}</span>
                  </div>
                  <div :title="getMoveTypeTooltip('dip')" data-bs-toggle="tooltip">
                    <span class="me-1">{{ getMoveTypeIcon('dips') }}</span>
                    <span class="fw-bold">{{ formatInt(country.dips) }}</span>
                  </div>
                  <div :title="getMoveTypeTooltip('seen')" data-bs-toggle="tooltip">
                    <span class="me-1">{{ getMoveTypeIcon('sees') }}</span>
                    <span class="fw-bold">{{ formatInt(country.seen) }}</span>
                  </div>
                  <div :title="getMoveTypeTooltip('loves')" data-bs-toggle="tooltip">
                    <span class="me-1">❤️</span>
                    <span class="fw-bold text-danger">{{ formatInt(country.total_loves) }}</span>
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
              <th style="width: 60px" data-bs-toggle="tooltip" title="Rank position based on current sort">#</th>
              <th data-bs-toggle="tooltip" title="Country name">Country</th>
              <th class="text-end" style="cursor:pointer" @click="headerSort('points')" data-bs-toggle="tooltip" title="Total points earned in this country (sorted from API)">Points <i class="bi" :class="sortIcon('points')"></i></th>
              <th class="text-end d-none d-md-table-cell text-nowrap" style="cursor:pointer" @click="headerSort('avg_points')" data-bs-toggle="tooltip" title="Average points earned per move in this country (sorted from API)">Avg/Move <i class="bi" :class="sortIcon('avg_points')"></i></th>
              <th class="text-end d-none d-sm-table-cell" style="cursor:pointer" @click="headerSort('moves')" data-bs-toggle="tooltip" title="Total number of recorded moves in this country (sorted from API)">Moves <i class="bi" :class="sortIcon('moves')"></i></th>
              <th class="text-end d-none d-lg-table-cell" style="cursor:pointer" @click="headerSort('drops')" data-bs-toggle="tooltip" :title="`${getMoveTypeTooltip('drop')} (sorted from API)`">🌳 <i class="bi" :class="sortIcon('drops')"></i></th>
              <th class="text-end d-none d-lg-table-cell" style="cursor:pointer" @click="headerSort('grabs')" data-bs-toggle="tooltip" :title="`${getMoveTypeTooltip('grab')} (sorted from API)`">🚀 <i class="bi" :class="sortIcon('grabs')"></i></th>
              <th class="text-end d-none d-lg-table-cell" style="cursor:pointer" @click="headerSort('dips')" data-bs-toggle="tooltip" :title="`${getMoveTypeTooltip('dip')} (sorted from API)`">🥾 <i class="bi" :class="sortIcon('dips')"></i></th>
              <th class="text-end d-none d-lg-table-cell" style="cursor:pointer" @click="headerSort('sees')" data-bs-toggle="tooltip" :title="`${getMoveTypeTooltip('seen')} (sorted from API)`">👀 <i class="bi" :class="sortIcon('sees')"></i></th>
              <th class="text-end d-none d-md-table-cell" style="cursor:pointer" @click="headerSort('gks')" data-bs-toggle="tooltip" title="Number of distinct GeoKrety that visited this country (sorted from API)">GeoKrety <i class="bi" :class="sortIcon('gks')"></i></th>
              <th class="text-end d-none d-md-table-cell" style="cursor:pointer" @click="headerSort('users')" data-bs-toggle="tooltip" title="Number of distinct users who made moves in this country (sorted from API)">Users <i class="bi" :class="sortIcon('users')"></i></th>
              <th class="text-end" style="cursor:pointer" @click="headerSort('loves')" data-bs-toggle="tooltip" title="Total loves given to GeoKrety that visited this country (sorted from API)">❤️ <i class="bi" :class="sortIcon('loves')"></i></th>
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
              <td class="text-end fw-bold text-success"><PointsValue :value="country.total_points_awarded" :digits="0" /></td>
              <td class="text-end fw-bold text-info small d-none d-md-table-cell"><PointsValue :value="country.avg_points_per_move" /></td>
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
