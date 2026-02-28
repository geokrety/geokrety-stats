<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import { idToGkId } from '../composables/useGkId.js'
import { getMoveTypeBadgeClass, getGkTypeBadgeClass } from '../composables/useMoveTypeColors.js'
import LineChart from '../components/LineChart.vue'
import WorldMap from '../components/WorldMap.vue'
import Pagination from '../components/Pagination.vue'

const route   = useRoute()
const gkId    = ref(route.params.id)
const gk      = ref(null)
const timeline = ref([])
const countries = ref([])
const moves    = ref([])
const movePage = ref(1)
const moveMeta = ref({})
const loading  = ref(false)
const error    = ref(null)
const activeTab = ref('overview')

const today = new Date().toISOString().slice(0, 10)

const chartStartDate = computed(() => {
  if (gk.value?.first_move_at) return gk.value.first_move_at.slice(0, 10)
  if (gk.value?.born_at)       return gk.value.born_at.slice(0, 10)
  if (gk.value?.created_at)    return gk.value.created_at.slice(0, 10)
  return null
})

async function load() {
  loading.value = true
  error.value   = null
  try {
    gk.value        = await fetchOne(`/geokrety/${gkId.value}`)
    const tl        = await fetchList(`/geokrety/${gkId.value}/points/timeline`, { per_page: 3650 })
    timeline.value  = tl.items
    const co        = await fetchList(`/geokrety/${gkId.value}/countries`, { per_page: 300 })
    countries.value = co.items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadMoves() {
  const { items, meta } = await fetchList(`/geokrety/${gkId.value}/moves`, {
    page: movePage.value, per_page: 25,
  })
  moves.value    = items
  moveMeta.value = meta
}

onMounted(() => { load(); loadMoves() })
watch(movePage, loadMoves)
watch(() => route.params.id, (id) => { gkId.value = id; load(); loadMoves() })
</script>

<template>
  <div v-if="loading && !gk" class="text-center py-5">
    <div class="spinner-border"></div>
  </div>
  <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
  <div v-else-if="gk">
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-3">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Leaderboard</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">{{ gk.gk_name }}</li>
      </ol>
    </nav>

    <!-- GK Header -->
    <div class="card mb-4 shadow-sm">
      <div class="card-body d-flex align-items-center gap-4 flex-wrap">
        <div class="fs-1">🎯</div>
        <div class="flex-grow-1">
          <div class="d-flex align-items-center gap-2 flex-wrap">
            <h3 class="mb-0">{{ gk.gk_name }}</h3>
            <span class="badge bg-dark" style="font-size: 0.8rem">{{ idToGkId(gk.gk_id) }}</span>
          </div>
          <p class="mb-0 text-muted small mt-1">
            Type: <span :class="`badge ${getGkTypeBadgeClass(gk.gk_type_name)}`">{{ gk.gk_type_name || 'unknown' }}</span>
            <span v-if="gk.missing" class="badge bg-danger ms-2">Missing</span>
          </p>
          <p class="mb-0 text-muted small">
            Owner:
            <RouterLink v-if="gk.owner_id" :to="`/users/${gk.owner_id}`">{{ gk.owner_username }}</RouterLink>
            <span v-else>—</span>
            &ensp;|&ensp;
            Holder:
            <RouterLink v-if="gk.holder_id" :to="`/users/${gk.holder_id}`">{{ gk.holder_username }}</RouterLink>
            <span v-else>—</span>
          </p>
        </div>
        <div class="row g-3 text-center">
          <div class="col">
            <div class="fw-bold text-success fs-4">{{ gk.total_points_generated?.toLocaleString() }}</div>
            <div class="text-muted small">Points Generated</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.total_moves?.toLocaleString() }}</div>
            <div class="text-muted small">Total Moves</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.distance_km?.toLocaleString() }} km</div>
            <div class="text-muted small">Distance</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.countries_count?.toLocaleString() }}</div>
            <div class="text-muted small">Countries</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.current_multiplier?.toFixed(2) }}×</div>
            <div class="text-muted small" title="Points multiplier applied to moves with this GeoKret">Multiplier</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <ul class="nav nav-tabs mb-3">
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'overview' }" @click="activeTab = 'overview'">
          <i class="bi bi-bar-chart-line me-1"></i>Overview
        </button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'moves' }" @click="activeTab = 'moves'">
          <i class="bi bi-list-ul me-1"></i>Moves
        </button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'countries' }" @click="activeTab = 'countries'">
          <i class="bi bi-globe me-1"></i>Countries
          <span v-if="countries.length" class="badge bg-secondary ms-1">{{ countries.length }}</span>
        </button>
      </li>
    </ul>

    <!-- Overview -->
    <div v-if="activeTab === 'overview'">
      <div class="card mb-4 shadow-sm">
        <div class="card-header d-flex justify-content-between align-items-center">
          <b>Points Generated per Day</b>
          <span class="text-muted small" v-if="chartStartDate">since {{ chartStartDate }}</span>
        </div>
        <div class="card-body">
          <LineChart
            v-if="timeline.length"
            :data="timeline"
            x-key="day"
            y-key="points"
            color="#198754"
            :height="220"
            :startDate="chartStartDate"
            :endDate="today"
            :showRangeButtons="true"
          />
          <p v-else class="text-muted text-center py-3">No timeline data.</p>
        </div>
      </div>
      <!-- GK stats mini summary -->
      <div class="row g-3">
        <div class="col-md-4">
          <div class="card shadow-sm h-100">
            <div class="card-body">
              <h6 class="card-title text-muted">Move breakdown</h6>
              <ul class="list-unstyled mb-0 small">
                <li><span class="fw-semibold">{{ gk.total_drops?.toLocaleString() }}</span> drops</li>
                <li><span class="fw-semibold">{{ gk.total_grabs?.toLocaleString() }}</span> grabs</li>
                <li><span class="fw-semibold">{{ gk.total_seen?.toLocaleString() }}</span> seen</li>
                <li><span class="fw-semibold">{{ gk.total_dips?.toLocaleString() }}</span> dips</li>
              </ul>
            </div>
          </div>
        </div>
        <div class="col-md-4">
          <div class="card shadow-sm h-100">
            <div class="card-body">
              <h6 class="card-title text-muted">Reach</h6>
              <ul class="list-unstyled mb-0 small">
                <li><span class="fw-semibold">{{ gk.distinct_users?.toLocaleString() }}</span> distinct users</li>
                <li><span class="fw-semibold">{{ gk.distinct_caches?.toLocaleString() }}</span> distinct waypoints</li>
                <li><span class="fw-semibold">{{ gk.users_awarded?.toLocaleString() }}</span> users awarded</li>
              </ul>
            </div>
          </div>
        </div>
        <div class="col-md-4">
          <div class="card shadow-sm h-100">
            <div class="card-body">
              <h6 class="card-title text-muted">Dates</h6>
              <ul class="list-unstyled mb-0 small">
                <li v-if="gk.born_at">Born: {{ gk.born_at?.slice(0, 10) }}</li>
                <li v-if="gk.first_move_at">First move: {{ gk.first_move_at?.slice(0, 10) }}</li>
                <li v-if="gk.last_move_at">Last move: {{ gk.last_move_at?.slice(0, 10) }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Moves -->
    <div v-if="activeTab === 'moves'">
      <div class="card shadow-sm">
        <div class="table-responsive">
          <table class="table table-hover table-sm mb-0 align-middle">
            <thead class="table-dark">
              <tr>
                <th>Date</th>
                <th>User</th>
                <th>Type</th>
                <th class="text-end">Points</th>
                <th>Country</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in moves" :key="m.move_id">
                <td class="small text-muted">{{ m.moved_on?.slice(0, 10) }}</td>
                <td>
                  <RouterLink :to="`/users/${m.author_id}`">{{ m.author_username }}</RouterLink>
                </td>
                <td><span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`">{{ m.type_name }}</span></td>
                <td class="text-end fw-semibold text-success">{{ m.points?.toLocaleString() }}</td>
                <td>{{ m.country }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="moveMeta.total" :meta="moveMeta" v-model:page="movePage" class="mt-3" />
    </div>

    <!-- Countries -->
    <div v-if="activeTab === 'countries'">
      <div class="card shadow-sm mb-3">
        <div class="card-header"><b>Countries visited</b></div>
        <div class="card-body p-2">
          <WorldMap v-if="countries.length" :countries="countries" :height="380" />
          <p v-else class="text-muted text-center py-3">No countries data.</p>
        </div>
      </div>
      <div class="row row-cols-2 row-cols-md-4 row-cols-lg-6 g-2">
        <div v-for="c in countries" :key="c.country" class="col">
          <div class="card text-center p-2 shadow-sm h-100">
            <div class="fw-semibold">{{ c.country }}</div>
            <div class="text-muted small">{{ c.move_count?.toLocaleString() }} moves</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
