<script setup>
import { ref, onMounted, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import LineChart from '../components/LineChart.vue'

const stats      = ref(null)
const daily      = ref([])
const countries  = ref([])
const breakdown  = ref([])
const loading    = ref(false)
const error      = ref(null)
const chartKey   = ref('moves')   // currently displayed metric for daily chart
const countrySort = ref('total_moves')  // column for countries table sort
const countrySortDir = ref('desc')

const dailyMetrics = [
  { key: 'moves',    label: 'All Moves', color: '#0dcaf0' },
  { key: 'drops',    label: 'Drops',     color: '#dc3545' },
  { key: 'grabs',    label: 'Grabs',     color: '#ffc107' },
  { key: 'dips',     label: 'Dips',      color: '#0d6efd' },
  { key: 'seen',     label: 'Seen',      color: '#6c757d' },
]

const currentMetric = computed(() => dailyMetrics.find(m => m.key === chartKey.value) || dailyMetrics[0])

const sortedCountries = computed(() => {
  const arr = [...countries.value]
  return arr.sort((a, b) => {
    const va = Number(a[countrySort.value] || 0)
    const vb = Number(b[countrySort.value] || 0)
    return countrySortDir.value === 'desc' ? vb - va : va - vb
  })
})

function toggleCountrySort(col) {
  if (countrySort.value === col) {
    countrySortDir.value = countrySortDir.value === 'desc' ? 'asc' : 'desc'
  } else {
    countrySort.value = col
    countrySortDir.value = 'desc'
  }
}

onMounted(async () => {
  loading.value = true
  error.value   = null
  try {
    stats.value     = await fetchOne('/stats')
    const da        = await fetchList('/stats/activity/daily', { days: 90 })
    daily.value     = da.items
    const co        = await fetchList('/stats/countries', { per_page: 50 })
    countries.value = co.items
    const bd        = await fetchList('/stats/points/breakdown')
    breakdown.value = bd.items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <nav aria-label="breadcrumb" class="mb-3">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">Statistics</li>
      </ol>
    </nav>

    <h4 class="mb-4"><i class="bi bi-graph-up-arrow text-info me-2"></i>Global Statistics</h4>

    <div v-if="loading && !stats" class="text-center py-5">
      <div class="spinner-border"></div>
    </div>
    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-else>
      <!-- KPI Cards -->
      <div class="row row-cols-2 row-cols-md-3 row-cols-lg-6 g-3 mb-4" v-if="stats">
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-2 fw-bold text-primary">
                <span style="display: inline-block; white-space: nowrap;">{{ stats.total_moves?.toLocaleString() }}</span>
              </div>
              <div class="text-muted small">Total Moves</div>
              <small class="d-block text-muted mt-1">All moves recorded in the system</small>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-2 fw-bold text-success">
                <span style="display: inline-block; white-space: nowrap;">{{ stats.total_points_awarded?.toLocaleString() }}</span>
              </div>
              <div class="text-muted small">Points Awarded</div>
              <small class="d-block text-muted mt-1">Total gamification points</small>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-2 fw-bold">
                <span style="display: inline-block; white-space: nowrap;">{{ stats.total_users?.toLocaleString() }}</span>
              </div>
              <div class="text-muted small">Total Users</div>
              <small class="d-block text-muted mt-1">Registered users on platform</small>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-2 fw-bold text-warning">
                <span style="display: inline-block; white-space: nowrap;">{{ stats.scored_users?.toLocaleString() }}</span>
              </div>
              <div class="text-muted small">Scored Users</div>
              <small class="d-block text-muted mt-1">Users who earned points</small>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-2 fw-bold text-info">
                <span style="display: inline-block; white-space: nowrap;">{{ stats.total_gks?.toLocaleString() }}</span>
              </div>
              <div class="text-muted small">GeoKrety</div>
              <small class="d-block text-muted mt-1">Total GeoKrety in circulation</small>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-2 fw-bold text-danger">
                <span style="display: inline-block; white-space: nowrap;">{{ stats.countries_reached?.toLocaleString() }}</span>
              </div>
              <div class="text-muted small">Countries</div>
              <small class="d-block text-muted mt-1">Countries with GeoKrety</small>
            </div>
          </div>
        </div>
      </div>

      <!-- Daily activity chart -->
      <div class="card mb-4 shadow-sm">
        <div class="card-header">
          <div class="d-flex justify-content-between align-items-center flex-wrap gap-2">
            <b>Daily Activity (Last 90 Days)</b>
            <!-- Move type filter buttons -->
            <div class="btn-group btn-group-sm" role="group" aria-label="Chart metric">
              <template v-for="m in dailyMetrics" :key="m.key">
                <input type="radio" class="btn-check" name="dailyMetric" :id="'dm-'+m.key" :value="m.key" v-model="chartKey">
                <label class="btn btn-outline-secondary" :for="'dm-'+m.key">{{ m.label }}</label>
              </template>
            </div>
          </div>
        </div>
        <div class="card-body">
          <LineChart v-if="daily.length" :data="daily" x-key="day" :y-key="chartKey" :color="currentMetric.color" :height="220" />
          <p v-else class="text-muted text-center py-3">No data.</p>
          <hr class="my-3" />
          <div class="row g-3 text-center small">
            <div class="col-6 col-md-3">
              <div class="text-muted">Active Days</div>
              <div class="fw-bold">{{ daily.length }}</div>
            </div>
            <div class="col-6 col-md-3">
              <div class="text-muted">Avg/Day</div>
              <div class="fw-bold">{{ (daily.reduce((s, d) => s + (d[chartKey] || 0), 0) / (daily.length || 1)).toFixed(0) }}</div>
            </div>
            <div class="col-6 col-md-3">
              <div class="text-muted">Peak Day</div>
              <div class="fw-bold">{{ Math.max(...daily.map(d => d[chartKey] || 0)).toLocaleString() }}</div>
            </div>
            <div class="col-6 col-md-3">
              <div class="text-muted">Total Period</div>
              <div class="fw-bold">{{ daily.reduce((s, d) => s + (d[chartKey] || 0), 0).toLocaleString() }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="row g-4">
        <!-- Top countries -->
        <div class="col-lg-6">
          <div class="card shadow-sm h-100">
            <div class="card-header"><b>Top 20 Countries by Moves</b></div>
            <div class="table-responsive" style="max-height:370px; overflow-y:auto">
              <table class="table table-sm table-hover mb-0">
                <thead class="table-light sticky-top">
                  <tr>
                    <th>Country</th>
                    <th class="text-end" style="cursor:pointer" @click="toggleCountrySort('total_moves')" title="Sort by moves">
                      Moves <span v-if="countrySort==='total_moves'">{{ countrySortDir==='desc' ? '▼' : '▲' }}</span>
                    </th>
                    <th class="text-end" style="cursor:pointer" @click="toggleCountrySort('total_points_awarded')" title="Sort by points">
                      Points <span v-if="countrySort==='total_points_awarded'">{{ countrySortDir==='desc' ? '▼' : '▲' }}</span>
                    </th>
                    <th class="text-end" style="cursor:pointer" @click="toggleCountrySort('unique_users')" title="Sort by users">
                      Users <span v-if="countrySort==='unique_users'">{{ countrySortDir==='desc' ? '▼' : '▲' }}</span>
                    </th>
                    <th class="text-end" style="cursor:pointer" @click="toggleCountrySort('drops')" title="Drops">
                      Drops <span v-if="countrySort==='drops'">{{ countrySortDir==='desc' ? '▼' : '▲' }}</span>
                    </th>
                    <th class="text-end" style="cursor:pointer" @click="toggleCountrySort('dips')" title="Dips">
                      Dips <span v-if="countrySort==='dips'">{{ countrySortDir==='desc' ? '▼' : '▲' }}</span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="c in sortedCountries" :key="c.country">
                    <td class="fw-semibold">{{ c.country?.toUpperCase() }}</td>
                    <td class="text-end">{{ Number(c.total_moves||0).toLocaleString() }}</td>
                    <td class="text-end text-success">{{ Number(c.total_points_awarded||0).toLocaleString() }}</td>
                    <td class="text-end">{{ Number(c.unique_users||0).toLocaleString() }}</td>
                    <td class="text-end text-danger small">{{ Number(c.drops||0).toLocaleString() }}</td>
                    <td class="text-end text-primary small">{{ Number(c.dips||0).toLocaleString() }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
        <!-- Points breakdown -->
        <div class="col-lg-6">
          <div class="card shadow-sm h-100">
            <div class="card-header"><b>Points Breakdown by Type</b></div>
            <div class="table-responsive" style="max-height:360px; overflow-y:auto">
              <table class="table table-sm mb-0">
                <thead class="table-light sticky-top">
                  <tr>
                    <th>Reward Type</th>
                    <th class="text-end">Points</th>
                    <th class="text-end">Count</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="b in breakdown" :key="b.label">
                    <td>
                      <span class="badge bg-light text-dark">{{ b.label.replace(/_/g, ' ') }}</span>
                    </td>
                    <td class="text-end">{{ b.points?.toLocaleString() }}</td>
                    <td class="text-end">{{ b.count?.toLocaleString() }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <small class="d-block text-muted text-center mt-2">Distribution of awarded points by reward type</small>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
