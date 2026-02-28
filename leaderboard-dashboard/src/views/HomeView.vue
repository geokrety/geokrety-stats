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
const perPage   = ref(25)
const sortCol   = ref(route.query.sort   || 'rank')  // rank|points|moves
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
      period: effectivePeriod.value, page: page.value, per_page: perPage.value,
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
  router.replace({ query: q })
}

onMounted(() => { fetchYears(); load() })
watch([period, yearValue], () => { page.value = 1; load() })
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
  if (isPeriodMode.value && row.points_period) return row.points_period
  return row.total_points
}

const sortOptions = [
  { key: 'rank',     label: 'Rank',        title: 'Sort by overall rank' },
  { key: 'points',   label: 'Points',      title: 'Sort by total points' },
  { key: 'moves',    label: 'Moves',       title: 'Sort by number of moves' },
  { key: 'gks',      label: 'GKs',         title: 'Sort by number of GeoKrety interacted with' },
  { key: 'countries',label: 'Countries',   title: 'Sort by countries visited' },
]

function sortedRows() {
  if (sortCol.value === 'rank' || !rows.value.length) return rows.value
  return [...rows.value].sort((a, b) => {
    switch (sortCol.value) {
      case 'points':    return (displayPoints(b) || 0) - (displayPoints(a) || 0)
      case 'moves':     return (b.move_count || 0) - (a.move_count || 0)
      case 'gks':       return (b.gk_count || 0) - (a.gk_count || 0)
      case 'countries': return (b.countries_count || 0) - (a.countries_count || 0)
      default:          return (a.rank || 0) - (b.rank || 0)
    }
  })
}
</script>

<template>
  <div>
    <!-- Header card -->
    <div class="card mb-4 shadow-sm">
      <div class="card-body">
        <h2 class="mb-1"><i class="bi bi-trophy-fill text-warning me-2"></i>Leaderboard</h2>
        <p class="text-muted mb-0">User rankings by points earned through GeoKrety interactions</p>
      </div>
    </div>

    <!-- Controls row -->
    <div class="d-flex align-items-center justify-content-end mb-3 flex-wrap gap-2">
      <div class="d-flex gap-2 align-items-center flex-wrap">
        <!-- Live badge -->
        <span v-if="connected" class="badge bg-success"><i class="bi bi-broadcast me-1"></i>Live</span>
        <!-- Period selector -->
        <div class="btn-group btn-group-sm" role="group">
          <button
            v-for="p in PERIODS" :key="p.value"
            class="btn"
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
            class="btn btn-sm"
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

    <!-- Error -->
    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <!-- Table -->
    <div class="card shadow-sm">
      <div class="table-responsive">
        <table class="table table-hover mb-0 align-middle">
          <thead class="table-dark">
            <tr>
              <th style="width:60px" title="Rank position for the selected time period">#</th>
              <th title="Username (click to view profile)">User</th>
              <th class="text-end" style="cursor:pointer" :class="sortCol==='points' ? 'text-warning' : ''" @click="sortCol='points'" :title="'Total accumulated points for this period — click to sort'">
                Points <i class="bi" :class="sortCol==='points' ? 'bi-sort-down-alt' : 'bi-sort-down'"></i>
              </th>
              <th class="text-end" style="cursor:pointer" :class="sortCol==='moves' ? 'text-warning' : ''" @click="sortCol='moves'" :title="'Number of moves logged in this period — click to sort'">
                Moves <i class="bi" :class="sortCol==='moves' ? 'bi-sort-down-alt' : 'bi-sort-down'"></i>
              </th>
              <th class="text-end" style="cursor:pointer" :class="sortCol==='gks' ? 'text-warning' : ''" @click="sortCol='gks'" :title="'Number of distinct GeoKrety interacted with — click to sort'">
                GKs <i class="bi" :class="sortCol==='gks' ? 'bi-sort-down-alt' : 'bi-sort-down'"></i>
              </th>
              <th class="text-end" style="cursor:pointer" :class="sortCol==='countries' ? 'text-warning' : ''" @click="sortCol='countries'" :title="'Number of unique countries visited — click to sort'">
                Countries <i class="bi" :class="sortCol==='countries' ? 'bi-sort-down-alt' : 'bi-sort-down'"></i>
              </th>
              <th class="text-end" title="Average points earned per logged move (total_points ÷ total_moves)">
                Avg/move <i class="bi bi-info-circle text-secondary" style="font-size:0.75rem"></i>
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
            <tr v-for="row in sortedRows()" :key="row.user_id">
              <td :class="medalClass(row.rank)">
                <span v-if="row.rank === 1">🥇</span>
                <span v-else-if="row.rank === 2">🥈</span>
                <span v-else-if="row.rank === 3">🥉</span>
                <span v-else>{{ row.rank }}</span>
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
