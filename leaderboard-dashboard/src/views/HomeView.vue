<script setup>
import { ref, watch, onMounted, computed } from 'vue'
import { RouterLink, useRouter, useRoute } from 'vue-router'
import { fetchList, fetchOne } from '../composables/useApi.js'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import { useLeaderboardLive } from '../composables/useWebSocket.js'
import Pagination from '../components/Pagination.vue'

const PERIODS = [
  { value: 'all',     label: 'All Time' },
  { value: 'year',    label: 'This Year' },
  { value: '3months', label: 'Last 3 Months' },
  { value: 'month',   label: 'This Month' },
  { value: 'week',    label: 'This Week' },
  { value: 'today',   label: 'Today' },
]

const router = useRouter()
const route  = useRoute()

const period    = ref(route.query.period || 'all')
const yearValue = ref(route.query.year   || '')
const page      = ref(Number(route.query.page) || 1)
const perPage   = ref(100)
const sortCol   = ref(route.query.sort   || 'points') // default to points as it's the leaderboard
const sortOrder = ref(route.query.order  || 'desc')
const loading   = ref(false)
const error     = ref(null)
const rows      = ref([])
const meta      = ref({})
const availableYears = ref([])

const { connected, leaderboard: liveTop } = useLeaderboardLive()

// Effective period sent to the API
const effectivePeriod = computed(() => yearValue.value || period.value)

// Is this a period where the API returns points in `points_period` instead of `total_points`?
const isPeriodMode = computed(() =>
  yearValue.value ||
  ['today', 'week', 'month', '3months'].includes(period.value)
)

async function fetchYears() {
  try {
    const data = await fetchOne('/stats/periods')
    availableYears.value = (data.years || []).map(String).sort((a, b) => Number(b) - Number(a))
  } catch (e) {
    // non-critical
  }
}

async function load() {
  loading.value = true
  error.value   = null
  try {
    const { items, meta: m } = await fetchList('/leaderboard', {
      period: effectivePeriod.value,
      page: page.value,
      per_page: perPage.value,
      sort: sortCol.value,
      order: sortOrder.value,
    })
    rows.value = items
    meta.value = m
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
  // Sync URL
  const q = {}
  if (yearValue.value)  q.year   = yearValue.value
  if (period.value !== 'all') q.period = period.value
  if (page.value > 1)   q.page   = page.value
  if (sortCol.value !== 'points') q.sort = sortCol.value
  if (sortOrder.value !== 'desc') q.order = sortOrder.value
  router.replace({ query: q })
}

onMounted(() => { fetchYears(); load() })
watch([period, yearValue, sortCol, sortOrder], () => { page.value = 1; load() })
watch([page], load)

function selectPeriod(p) { period.value = p; yearValue.value = '' }
function selectYear(y)   { yearValue.value = y; period.value = '' }

function medalClass(rank) {
  if (rank === 1) return 'text-warning fw-bold'
  if (rank === 2) return 'text-secondary fw-bold'
  if (rank === 3) return 'text-danger fw-bold'
  return ''
}

// Helper: get the displayed points value for a row
function displayPoints(row) {
  // For period-based views, the API returns `points_period` (not total_points)
  if (isPeriodMode.value && row.points_period !== undefined) return row.points_period
  return row.total_points
}

function sortedRows() {
  return rows.value
}

function toggleSort(col) {
  if (sortCol.value === col) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
    return
  }
  sortCol.value = col
  sortOrder.value = 'desc'
}
</script>

<template>
  <div>
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item active" aria-current="page">Home</li>
      </ol>
    </nav>

    <!-- Header card -->
    <div class="card mb-4 shadow-sm">
      <div class="card-body">
        <h2 class="mb-1"><i class="bi bi-trophy-fill text-warning me-2"></i>Leaderboard</h2>
        <p class="text-muted mb-0">User rankings by points earned through GeoKrety interactions</p>
      </div>
    </div>

    <!-- Controls row -->
    <div class="d-flex align-items-center justify-content-end mb-2">
      <div class="container-fluid p-0">
        <div class="row align-items-center g-2">
          <div v-if="connected" class="col-auto">
            <span class="badge bg-success"><i class="bi bi-broadcast me-1"></i>Live</span>
          </div>
          <div class="col">
            <div class="d-flex flex-wrap gap-2 justify-content-end">
              <!-- Period selector -->
              <div class="btn-group btn-group-sm overflow-auto d-flex" role="group">
                <button
                  v-for="p in PERIODS" :key="p.value"
                  class="btn flex-fill text-nowrap"
                  :class="(period === p.value && !yearValue) ? 'btn-primary' : 'btn-outline-secondary'"
                  @click="selectPeriod(p.value)"
                >
                  <span v-if="p.value === 'year'">📅 Year</span>
                  <span v-else>{{ p.label }}</span>
                </button>
              </div>
              <!-- Year dropdown selector -->
              <div v-if="availableYears.length" class="dropdown">
                <button
                  class="btn btn-sm w-100"
                  :class="yearValue ? 'btn-info' : 'btn-outline-secondary'"
                  type="button"
                  data-bs-toggle="dropdown"
                  aria-expanded="false"
                >
                  {{ yearValue ? `${yearValue} 📅` : 'Select Year...' }}
                </button>
                <ul class="dropdown-menu dropdown-menu-end" style="max-height: 300px; overflow-y: auto;">
                  <li v-for="y in availableYears" :key="y">
                    <a href="#" class="dropdown-item" :class="yearValue === y ? 'active' : ''" @click.prevent="selectYear(y)">
                      {{ y }}
                    </a>
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Error -->
    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <!-- Table -->
    <div class="card shadow-sm">
      <div class="table-responsive border-0 mb-0">
        <table class="table table-hover mb-0 align-middle">
          <thead class="table-dark">
            <tr>
              <th style="width:60px" title="Rank position for the selected time period">#</th>
              <th title="Username (click to view profile)">User</th>
              <th class="text-end" style="cursor:pointer" :class="sortCol==='points' ? 'text-warning' : ''" @click="toggleSort('points')" :title="'Total accumulated points for this period — click to sort'">
                Points <i class="bi" :class="sortCol==='points' ? (sortOrder==='asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt') : 'bi-sort-down'"></i>
              </th>
              <th class="text-end" style="cursor:pointer" :class="sortCol==='moves' ? 'text-warning' : ''" @click="toggleSort('moves')" :title="'Number of moves logged in this period — click to sort'">
                Moves <i class="bi" :class="sortCol==='moves' ? (sortOrder==='asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt') : 'bi-sort-down'"></i>
              </th>
              <th class="text-end" style="cursor:pointer" :class="sortCol==='gks' ? 'text-warning' : ''" @click="toggleSort('gks')" :title="'Number of distinct GeoKrety interacted with — click to sort'">
                GKs <i class="bi" :class="sortCol==='gks' ? (sortOrder==='asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt') : 'bi-sort-down'"></i>
              </th>
              <th class="text-end" style="cursor:pointer" :class="sortCol==='countries' ? 'text-warning' : ''" @click="toggleSort('countries')" :title="'Number of unique countries visited — click to sort'">
                Countries <i class="bi" :class="sortCol==='countries' ? (sortOrder==='asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt') : 'bi-sort-down'"></i>
              </th>
              <th class="text-end" style="cursor:pointer" :class="sortCol==='avg_points' ? 'text-warning' : ''" @click="toggleSort('avg_points')" title="Average points earned per logged move (total_points ÷ total_moves) — click to sort">
                Avg/move <i class="bi" :class="sortCol==='avg_points' ? (sortOrder==='asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt') : 'bi-sort-down'"></i>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading && !rows.length">
              <td colspan="7" class="text-center py-4">
                <div class="spinner-border spinner-border-sm me-2"></div>Loading…
              </td>
            </tr>
            <tr v-else-if="!rows.length && !loading">
              <td colspan="7" class="text-center text-secondary py-4">No data for this period.</td>
            </tr>
            <tr v-for="(row, index) in rows" :key="row.user_id">
              <td :class="medalClass(row.rank)">
                <span v-if="(meta.page-1)*meta.per_page + index + 1 === 1">🥇</span>
                <span v-else-if="(meta.page-1)*meta.per_page + index + 1 === 2">🥈</span>
                <span v-else-if="(meta.page-1)*meta.per_page + index + 1 === 3">🥉</span>
                <span v-else>{{ (meta.page-1)*meta.per_page + index + 1 }}</span>
              </td>
              <td>
                <RouterLink :to="`/users/${row.user_id}`" class="text-decoration-none fw-semibold">
                  {{ row.username }}
                </RouterLink>
                <span v-if="row.home_country" class="text-muted small ms-1" :title="`Home country: ${row.home_country.toUpperCase()}`">
                  {{ getCountryFlag(row.home_country) }} {{ row.home_country.toUpperCase() }}
                </span>
              </td>
              <td class="text-end fw-bold text-primary">{{ displayPoints(row)?.toLocaleString() }}</td>
              <td class="text-end">{{ row.move_count?.toLocaleString() }}</td>
              <td class="text-end">{{ row.gk_count?.toLocaleString() }}</td>
              <td class="text-end">{{ row.countries_count?.toLocaleString() ?? '—' }}</td>
              <td class="text-end text-muted small">
                <span :title="`${displayPoints(row)?.toLocaleString()} pts ÷ ${row.move_count?.toLocaleString()} moves`">
                  {{ row.avg_points_per_move?.toFixed(1) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <Pagination v-if="meta.total" :meta="meta" v-model:page="page" class="mt-3" />
  </div>
</template>
