<script setup>
import { ref, onMounted } from 'vue'
import { fetchOne, fetchList } from '../composables/useApi.js'
import LineChart from '../components/LineChart.vue'
import BarChart from '../components/BarChart.vue'

const stats      = ref(null)
const daily      = ref([])
const countries  = ref([])
const breakdown  = ref([])
const loading    = ref(false)
const error      = ref(null)

onMounted(async () => {
  loading.value = true
  error.value   = null
  try {
    stats.value     = await fetchOne('/stats')
    const da        = await fetchList('/stats/activity/daily', { days: 90 })
    daily.value     = da.items
    const co        = await fetchList('/stats/countries')
    countries.value = co.items.slice(0, 20)
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
              <div class="fs-3 fw-bold text-primary">{{ stats.total_moves?.toLocaleString() }}</div>
              <div class="text-muted small">Total Moves</div>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-3 fw-bold text-success">{{ stats.total_points_awarded?.toLocaleString() }}</div>
              <div class="text-muted small">Points Awarded</div>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-3 fw-bold">{{ stats.total_users?.toLocaleString() }}</div>
              <div class="text-muted small">Users</div>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-3 fw-bold text-warning">{{ stats.scored_users?.toLocaleString() }}</div>
              <div class="text-muted small">Scored Users</div>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-3 fw-bold text-info">{{ stats.total_gks?.toLocaleString() }}</div>
              <div class="text-muted small">GeoKrety</div>
            </div>
          </div>
        </div>
        <div class="col">
          <div class="card text-center shadow-sm h-100">
            <div class="card-body py-3">
              <div class="fs-3 fw-bold text-danger">{{ stats.countries_reached?.toLocaleString() }}</div>
              <div class="text-muted small">Countries</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Daily activity chart -->
      <div class="card mb-4 shadow-sm">
        <div class="card-header"><b>Daily Activity</b> (last 90 days)</div>
        <div class="card-body">
          <LineChart v-if="daily.length" :data="daily" x-key="day" y-key="moves" color="#0dcaf0" :height="220" />
          <p v-else class="text-muted text-center py-3">No data.</p>
        </div>
      </div>

      <div class="row g-4">
        <!-- Top countries -->
        <div class="col-lg-6">
          <div class="card shadow-sm h-100">
            <div class="card-header"><b>Top 20 Countries by Moves</b></div>
            <div class="card-body p-2">
              <BarChart v-if="countries.length" :data="countries" x-key="country" y-key="move_count" color="#ffc107" :height="320" />
              <p v-else class="text-muted text-center py-3">No data.</p>
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
                    <th>Source</th>
                    <th class="text-end">Points</th>
                    <th class="text-end">Count</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="b in breakdown" :key="b.source">
                    <td>{{ b.source }}</td>
                    <td class="text-end">{{ b.total_points?.toLocaleString() }}</td>
                    <td class="text-end">{{ b.count?.toLocaleString() }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
