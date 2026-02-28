<script setup>
import { ref, watch, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchList } from '../composables/useApi.js'
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

const period   = ref('all')
const page     = ref(1)
const perPage  = ref(25)
const loading  = ref(false)
const error    = ref(null)
const rows     = ref([])
const meta     = ref({})

const { connected, leaderboard: liveTop } = useLeaderboardLive()

async function load() {
  loading.value = true
  error.value   = null
  try {
    const { items, meta: m } = await fetchList('/leaderboard', {
      period: period.value, page: page.value, per_page: perPage.value,
    })
    rows.value = items
    meta.value = m
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([period], () => { page.value = 1; load() })
watch([page], load)

// Live update: merge top-10 into first page results when connected
watch(liveTop, (live) => {
  if (period.value === 'all' && page.value === 1 && live.length) {
    rows.value = live
  }
})

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
            :class="period === p.value ? 'btn-primary' : 'btn-outline-secondary'"
            @click="period = p.value"
          >{{ p.label }}</button>
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
              <th class="text-end">Points</th>
              <th class="text-end">Moves</th>
              <th class="text-end">GKs</th>
              <th class="text-end">Countries</th>
              <th class="text-end">Avg/move</th>
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
              </td>
              <td class="text-end fw-bold text-primary">{{ row.total_points?.toLocaleString() }}</td>
              <td class="text-end">{{ row.move_count?.toLocaleString() }}</td>
              <td class="text-end">{{ row.gk_count?.toLocaleString() }}</td>
              <td class="text-end">{{ row.countries_count?.toLocaleString() }}</td>
              <td class="text-end text-muted small">{{ row.avg_points_per_move?.toFixed(1) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <Pagination v-if="meta.total" :meta="meta" v-model:page="page" class="mt-3" />
  </div>
</template>
