<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import { idToGkId } from '../composables/useGkId.js'
import LineChart from '../components/LineChart.vue'
import Pagination from '../components/Pagination.vue'

const route = useRoute()
const userId = ref(route.params.id)

const user        = ref(null)
const timeline    = ref([])
const countries   = ref([])
const moves       = ref([])
const breakdown   = ref([])
const movePage    = ref(1)
const moveMeta    = ref({})
const loading     = ref(false)
const error       = ref(null)
const activeTab   = ref('overview')

async function load() {
  loading.value = true
  error.value   = null
  try {
    user.value      = await fetchOne(`/users/${userId.value}`)
    const tl        = await fetchList(`/users/${userId.value}/points/timeline`)
    timeline.value  = tl.items
    const co        = await fetchList(`/users/${userId.value}/countries`, { per_page: 100 })
    countries.value = co.items
    const bd        = await fetchList(`/users/${userId.value}/points/breakdown`)
    breakdown.value = bd.items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadMoves() {
  const { items, meta } = await fetchList(`/users/${userId.value}/moves`, {
    page: movePage.value, per_page: 25,
  })
  moves.value    = items
  moveMeta.value = meta
}

onMounted(() => { load(); loadMoves() })
watch(movePage, loadMoves)
watch(() => route.params.id, (id) => { userId.value = id; load(); loadMoves() })
</script>

<template>
  <div v-if="loading && !user" class="text-center py-5">
    <div class="spinner-border"></div>
  </div>
  <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
  <div v-else-if="user">
    <!-- User header -->
    <div class="card mb-4 shadow-sm">
      <div class="card-body d-flex align-items-center gap-4 flex-wrap">
        <div class="fs-1">👤</div>
        <div class="flex-grow-1">
          <h3 class="mb-1">{{ user.username }}</h3>
          <p class="text-muted mb-0 small">User #{{ user.user_id }}</p>
        </div>
        <div class="row g-3 text-center">
          <div class="col">
            <div class="fw-bold text-primary fs-4">{{ user.total_points?.toLocaleString() }}</div>
            <div class="text-muted small">Total Points</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ user.rank?.toLocaleString() }}</div>
            <div class="text-muted small">Rank</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ user.total_moves?.toLocaleString() }}</div>
            <div class="text-muted small">Moves</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ user.gk_count?.toLocaleString() }}</div>
            <div class="text-muted small">GeoKrety</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ user.countries_count?.toLocaleString() }}</div>
            <div class="text-muted small">Countries</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <ul class="nav nav-tabs mb-3">
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'overview' }" @click="activeTab = 'overview'">
          Overview
        </button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'moves' }" @click="activeTab = 'moves'">
          Moves
        </button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'countries' }" @click="activeTab = 'countries'">
          Countries
        </button>
      </li>
    </ul>

    <!-- Overview tab -->
    <div v-if="activeTab === 'overview'">
      <!-- Points timeline chart -->
      <div class="card mb-4 shadow-sm">
        <div class="card-header"><b>Points per Day</b> (last 90 days)</div>
        <div class="card-body">
          <LineChart v-if="timeline.length" :data="timeline" x-key="day" y-key="points" color="#0d6efd" />
          <p v-else class="text-muted text-center py-3">No timeline data.</p>
        </div>
      </div>
      <!-- Points breakdown table -->
      <div class="card shadow-sm">
        <div class="card-header"><b>Points Breakdown</b></div>
        <div class="table-responsive">
          <table class="table table-sm mb-0">
            <thead class="table-light">
              <tr><th>Source</th><th class="text-end">Points</th><th class="text-end">Count</th></tr>
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

    <!-- Moves tab -->
    <div v-if="activeTab === 'moves'">
      <div class="card shadow-sm">
        <div class="table-responsive">
          <table class="table table-hover table-sm mb-0 align-middle">
            <thead class="table-dark">
              <tr>
                <th>Date</th>
                <th>GeoKret</th>
                <th>Type</th>
                <th class="text-end">Points</th>
                <th>Country</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in moves" :key="m.move_id">
                <td class="small text-muted">{{ m.moved_on?.slice(0, 10) }}</td>
                <td>
                  <RouterLink :to="`/geokrety/${m.gk_id}`">
                    <span v-if="m.gk_name">{{ m.gk_name }}</span>
                    <code class="text-muted small">{{ idToGkId(m.gk_id) }}</code>
                  </RouterLink>
                </td>
                <td><span class="badge bg-secondary">{{ m.type_name }}</span></td>
                <td class="text-end fw-semibold text-primary">{{ m.points?.toLocaleString() }}</td>
                <td>{{ m.country }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="moveMeta.total" :meta="moveMeta" v-model:page="movePage" class="mt-3" />
    </div>

    <!-- Countries tab -->
    <div v-if="activeTab === 'countries'">
      <div class="row row-cols-2 row-cols-md-4 row-cols-lg-6 g-2">
        <div v-for="c in countries" :key="c.country" class="col">
          <div class="card text-center p-2 shadow-sm h-100">
            <div class="fw-semibold">{{ c.country }}</div>
            <div class="text-muted small">{{ c.move_count || c.moves || '—' }} moves</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
