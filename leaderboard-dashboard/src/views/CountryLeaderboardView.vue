<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchList } from '../composables/useApi.js'
import { getCountryFlag } from '../composables/useCountryFlags.js'

const countries = ref([])
const loading = ref(false)
const error = ref(null)
const sortBy = ref('points')

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
    // For now, fetch the basic countries endpoint
    // This would ideally have a /stats/countries endpoint with points data
    const { items } = await fetchList('/stats/countries', { per_page: 500 })

    // Sort based on selection
    const sorted = [...items].sort((a, b) => {
      switch (sortBy.value) {
        case 'points':
          return (b.points || 0) - (a.points || 0)
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
      <div class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-3">
        <div v-for="(country, idx) in countries" :key="country.country" class="col">
          <div class="card h-100 shadow-sm">
            <div class="card-header bg-light d-flex align-items-center gap-2">
              <span class="fs-3">{{ getCountryFlag(country.country) }}</span>
              <div>
                <div class="fw-bold">{{ country.country.toUpperCase() }}</div>
                <div class="text-muted small">{{ idx + 1 }}th</div>
              </div>
            </div>
            <div class="card-body">
              <div class="row g-3">
                <div class="col-6">
                  <div class="text-muted small">Moves</div>
                  <div class="fw-bold fs-5">{{ country.move_count?.toLocaleString() }}</div>
                </div>
                <div class="col-6">
                  <div class="text-muted small">GeoKrety</div>
                  <div class="fw-bold fs-5">{{ country.gk_count?.toLocaleString() }}</div>
                </div>
                <div class="col-6">
                  <div class="text-muted small">Users</div>
                  <div class="fw-bold fs-5">{{ country.user_count?.toLocaleString() }}</div>
                </div>
                <div class="col-6">
                  <div class="text-muted small">Points</div>
                  <div class="fw-bold fs-5 text-success">{{ (country.points || 0).toLocaleString() }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
