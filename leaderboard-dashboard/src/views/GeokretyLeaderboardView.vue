<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Pagination from '../components/Pagination.vue'

const page = ref(1)
const perPage = ref(50)
const loading = ref(false)
const error = ref(null)
const geokrety = ref([])
const meta = ref({})
const sortBy = ref('points')

const sortOptions = [
  { value: 'points', label: 'Points Generated' },
  { value: 'moves', label: 'Total Moves' },
  { value: 'users', label: 'Distinct Users' },
  { value: 'countries', label: 'Countries' },
]

async function loadGeokrety() {
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams({
      page: page.value,
      per_page: perPage.value,
    })
    const response = await fetch(`http://localhost:8080/api/v1/geokrety?${params}`, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    })
    if (!response.ok) throw new Error(`API error: ${response.status}`)
    const json = await response.json()
    geokrety.value = json.data || []
    meta.value = json.meta || {}
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadGeokrety)

watch(page, loadGeokrety)

const handleSort = () => {
  page.value = 1
  loadGeokrety()
}

const getGkTypeColor = (typeName) => {
  if (!typeName) return 'bg-secondary'
  const lower = typeName.toLowerCase()
  if (lower.includes('traditional') || lower.includes('car')) return 'bg-primary'
  if (lower.includes('moving') || lower.includes('human')) return 'bg-success'
  if (lower.includes('evo')) return 'bg-info'
  return 'bg-secondary'
}

const getGkTypeIcon = (typeName) => {
  if (!typeName) return '❓'
  const lower = typeName.toLowerCase()
  if (lower.includes('traditional') || lower.includes('car')) return '🚗'
  if (lower.includes('moving') || lower.includes('human')) return '👤'
  if (lower.includes('evo')) return '🎮'
  return '❓'
}

const formatInt = (value) => {
  return value ? parseInt(value).toLocaleString() : '0'
}

const formatFloat = (value, decimals = 2) => {
  return value ? parseFloat(value).toFixed(decimals) : '0.00'
}
</script>

<template>
  <div>
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-3">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">GeoKrety</li>
      </ol>
    </nav>

    <!-- Header -->
    <div class="card mb-4 shadow-sm">
      <div class="card-body">
        <h2 class="mb-1">🎁 GeoKrety Database</h2>
        <p class="text-muted mb-0">Browse all GeoKrety items and their movement statistics</p>
      </div>
    </div>

    <!-- Controls -->
    <div class="mb-3 d-flex gap-2 align-items-center flex-wrap">
      <span class="text-muted">Sort by:</span>
      <div class="btn-group btn-group-sm" role="group">
        <template v-for="opt in sortOptions" :key="opt.value">
          <input
            type="radio"
            class="btn-check"
            name="sortBtnradio"
            :id="'sort' + opt.value"
            :value="opt.value"
            v-model="sortBy"
            @change="handleSort"
          >
          <label class="btn btn-outline-primary" :for="'sort' + opt.value">{{ opt.label }}</label>
        </template>
      </div>
      <span class="text-muted ms-auto small">Showing {{ geokrety.length }} of {{ meta.total || 0 }} GeoKrety</span>
    </div>

    <!-- Loading / Error / Content -->
    <div v-if="loading && !geokrety.length" class="text-center py-5">
      <div class="spinner-border"></div>
    </div>
    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-else-if="geokrety.length === 0" class="alert alert-info">No GeoKrety found.</div>
    <div v-else class="table-responsive">
      <table class="table table-hover table-sm align-middle border">
        <thead class="table-light sticky-top">
          <tr>
            <th style="width: 80px">ID</th>
            <th>Name</th>
            <th>Type</th>
            <th>Owner</th>
            <th class="text-end">Moves</th>
            <th class="text-end">Users</th>
            <th class="text-end">Countries</th>
            <th class="text-end">Points</th>
            <th class="text-end">Multiplier</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="gk in geokrety" :key="gk.gk_id" class="table-row-link">
            <td>
              <RouterLink :to="`/geokrety/${gk.gk_id}`" class="text-decoration-none fw-bold text-primary">
                {{ formatInt(gk.gk_id) }}
              </RouterLink>
            </td>
            <td>
              <RouterLink :to="`/geokrety/${gk.gk_id}`" class="text-decoration-none fw-semibold d-block">
                {{ gk.gk_name || `GK #${gk.gk_id}` }}
              </RouterLink>
              <small v-if="gk.tracking_code" class="text-muted">{{ gk.tracking_code }}</small>
            </td>
            <td>
              <span :class="['badge', getGkTypeColor(gk.gk_type_name)]">
                {{ getGkTypeIcon(gk.gk_type_name) }} {{ gk.gk_type_name || 'Unknown' }}
              </span>
            </td>
            <td>
              <RouterLink v-if="gk.owner_id" :to="`/users/${gk.owner_id}`" class="text-decoration-none">
                {{ gk.owner_username || 'Unknown' }}
              </RouterLink>
              <span v-else class="text-muted">—</span>
            </td>
            <td class="text-end fw-bold">{{ formatInt(gk.total_moves) }}</td>
            <td class="text-end">{{ formatInt(gk.distinct_users) }}</td>
            <td class="text-end">{{ formatInt(gk.countries_count) }}</td>
            <td class="text-end fw-bold text-success">{{ formatFloat(gk.total_points_generated) }}</td>
            <td class="text-end">{{ formatFloat(gk.current_multiplier, 2) }}x</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <Pagination v-if="meta.total" :meta="meta" :page="page" @update:page="page = $event" class="mt-3" />
  </div>
</template>

<style scoped>
.table-row-link tbody tr {
  cursor: pointer;
}

.table-row-link tbody tr:hover {
  background-color: rgba(0, 123, 255, 0.05);
}
</style>
