<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { getCountryFlag } from '../composables/useCountryFlags.js'

const route = useRoute()
const country = ref(route.params.country?.toUpperCase())
const countryData = ref(null)
const loading = ref(false)
const error = ref(null)

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

async function loadCountryData() {
  loading.value = true
  error.value = null
  try {
    const response = await fetch(`/api/v1/stats/countries`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const { data } = await response.json()

    // Find this country in the data
    const found = data?.find(c => c.country.toUpperCase() === country.value)
    if (found) {
      countryData.value = found
    } else {
      error.value = `Country ${country.value} not found`
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadCountryData)
</script>

<template>
  <div>
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-3">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item"><RouterLink to="/countries">Countries</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">{{ country }}</li>
      </ol>
    </nav>

    <!-- Loading / Error / Content -->
    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border"></div>
    </div>
    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-else-if="!countryData" class="alert alert-info">Country data not available.</div>
    <div v-else>
      <!-- Header -->
      <div class="card mb-4 shadow-sm">
        <div class="card-body">
          <div class="d-flex align-items-center gap-3">
            <span class="fs-1">{{ getCountryFlag(country) }}</span>
            <div>
              <h1 class="mb-1">{{ country }} - {{ countryData?.country_name || country }}</h1>
              <p class="text-muted mb-0">Country activity and statistics</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Key Statistics -->
      <div class="row g-3 mb-4">
        <div class="col-12 col-md-6 col-lg-4">
          <div class="card shadow-sm border-0">
            <div class="card-body">
              <div class="text-muted small mb-2" title="Points awarded for all moves involving GeoKrety that visited this country">Total Points</div>
              <div class="fs-3 fw-bold text-success">{{ formatInt(countryData.total_points_awarded) }}</div>
              <div class="text-muted small mt-2">from {{ formatInt(countryData.total_moves) }} moves</div>
            </div>
          </div>
        </div>

        <div class="col-12 col-md-6 col-lg-4">
          <div class="card shadow-sm border-0">
            <div class="card-body">
              <div class="text-muted small mb-2" title="Average points awarded by GeoKrety that visited this country">Avg Points per Move</div>
              <div class="fs-3 fw-bold text-info">{{ formatFloat(countryData.avg_points_per_move, 4) }}</div>
              <div class="text-muted small mt-2">based on {{ formatInt(countryData.total_moves) }} total moves</div>
            </div>
          </div>
        </div>

        <div class="col-12 col-md-6 col-lg-4">
          <div class="card shadow-sm border-0">
            <div class="card-body">
              <div class="text-muted small mb-2" title="Number of distinct users who made moves involving GeoKrety that visited this country">Active Participants</div>
              <div class="fs-3 fw-bold text-primary">{{ formatInt(countryData.unique_users) }}</div>
              <div class="text-muted small mt-2">{{ formatInt(countryData.unique_gks) }} unique GeoKrety involved</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Move Type Breakdown -->
      <div class="card shadow-sm mb-4">
        <div class="card-header bg-light">
          <h5 class="mb-0">Move Type Breakdown</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-6 col-md-4 col-lg-2 text-center mb-3">
              <div class="fs-2 mb-2">📦</div>
              <div class="text-muted small">Drops</div>
              <div class="fs-5 fw-bold">{{ formatInt(countryData.drops) }}</div>
            </div>
            <div class="col-6 col-md-4 col-lg-2 text-center mb-3">
              <div class="fs-2 mb-2">🎯</div>
              <div class="text-muted small">Grabs</div>
              <div class="fs-5 fw-bold">{{ formatInt(countryData.grabs) }}</div>
            </div>
            <div class="col-6 col-md-4 col-lg-2 text-center mb-3">
              <div class="fs-2 mb-2">💧</div>
              <div class="text-muted small">DIPs</div>
              <div class="fs-5 fw-bold">{{ formatInt(countryData.dips) }}</div>
            </div>
            <div class="col-6 col-md-4 col-lg-2 text-center mb-3">
              <div class="fs-2 mb-2">👁️</div>
              <div class="text-muted small">Seen</div>
              <div class="fs-5 fw-bold">{{ formatInt(countryData.seen) }}</div>
            </div>
            <div class="col-6 col-md-4 col-lg-2 text-center mb-3">
              <div class="fs-2 mb-2">❤️</div>
              <div class="text-muted small">Loves</div>
              <div class="fs-5 fw-bold text-danger">{{ formatInt(countryData.total_loves) }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Additional Stats -->
      <div class="card shadow-sm">
        <div class="card-header bg-light">
          <h5 class="mb-0">Summary</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <div class="mb-3">
                <div class="text-muted small">Unique GeoKrety</div>
                <div class="fs-5 fw-bold">{{ formatInt(countryData.unique_gks) }}</div>
              </div>
            </div>
            <div class="col-md-6">
              <div class="mb-3">
                <div class="text-muted small">Unique Users</div>
                <div class="fs-5 fw-bold">{{ formatInt(countryData.unique_users) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
a {
  color: inherit;
  text-decoration: none;
}

a:hover {
  color: #0d6efd;
}
</style>
