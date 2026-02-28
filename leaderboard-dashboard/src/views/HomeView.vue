<script setup>
import { ref, watch, onMounted, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchList } from '../composables/useApi.js'
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

const period    = ref('all')
const yearValue = ref('')       // e.g. '2023'
const page      = ref(1)
const perPage   = ref(25)
const loading   = ref(false)
const error     = ref(null)
const rows      = ref([])
const meta      = ref({})
const availableYears = ref([])

const { connected, leaderboard: liveTop } = useLeaderboardLive()

// Effective period sent to the API
const effectivePeriod = computed(() => yearValue.value || period.value)

async function fetchYears() {
  try {
    const data = await fetchList('/stats/periods', { per_page: 100 })
    // data.items may be [{year: 2023, months: [...]}, ...]
    const items = data.items || []
    availableYears.value = items.map(i => String(i.year || i)).filter(Boolean).sort((a, b) => b - a)
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
}

onMounted(() => { fetchYears(); load() })
watch([period, yearValue], () => { page.value = 1; load() })
watch([page], load)

// Live update: merge top-10 into first page results when connected
watch(liveTop, (live) => {
  if (effectivePeriod.value === 'all' && page.value === 1 && live.length) {
    rows.value = live
  }
})

function selectPeriod(p) { period.value = p; yearValue.value = '' }
function selectYear(y)   { yearValue.value = y; period.value = '' }

function medalClass(rank) {
  if (rank === 1) return 'text-warning fw-bold'
  if (rank === 2) return 'text-secondary fw-bold'
  if (rank === 3) return 'text-danger fw-bold'
  return ''
}
</script>

<template>
  <div>
    <!-- Header row -->
    <div class="d-flex align-items-center justify-content-between mb-3 flex-wrap gap-2">
      <h4 class="mb-0"><i class="bi bi-trophy-fill text-warning me-2"></i>Leaderboard</h4>
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
              <th style="width:60px">#</th>
              <th>User</th>
              <th class="text-end" title="Total accumulated points for this period">Points</th>
              <th class="text-end" title="Number of moves logged in this period">Moves</th>
              <th class="text-end" title="Number of distinct GeoKrety interacted with">GKs</th>
              <th class="text-end" title="Number of unique countries visited">Countries</th>
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
            <tr v-for="row in rows" :key="row.user_id">
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
                <span v-if="row.home_country" class="text-muted small ms-1" :title="`Country: ${row.home_country}`">
                  {{ getCountryFlag(row.home_country) }} {{ row.home_country.toUpperCase() }}
                </span>
              </td>
              <td class="text-end fw-bold text-primary">{{ row.total_points?.toLocaleString() }}</td>
              <td class="text-end">{{ row.move_count?.toLocaleString() }}</td>
              <td class="text-end">{{ row.gk_count?.toLocaleString() }}</td>
              <td class="text-end">{{ row.countries_count?.toLocaleString() ?? '—' }}</td>
              <td class="text-end text-muted small">
                <span :title="`${row.total_points?.toLocaleString()} pts ÷ ${row.move_count?.toLocaleString()} moves`">
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
