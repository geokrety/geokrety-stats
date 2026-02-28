<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import LineChart from '../components/LineChart.vue'
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

async function load() {
  loading.value = true
  error.value   = null
  try {
    gk.value        = await fetchOne(`/geokrety/${gkId.value}`)
    const tl        = await fetchList(`/geokrety/${gkId.value}/points/timeline`)
    timeline.value  = tl.items
    const co        = await fetchList(`/geokrety/${gkId.value}/countries`, { per_page: 100 })
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
    <!-- GK Header -->
    <div class="card mb-4 shadow-sm">
      <div class="card-body d-flex align-items-center gap-4 flex-wrap">
        <div class="fs-1">🎯</div>
        <div class="flex-grow-1">
          <h3 class="mb-1">{{ gk.name }}</h3>
          <p class="mb-0 text-muted small">
            Tracking: <code>{{ gk.tracking_code }}</code>
            &ensp;|&ensp; Type: {{ gk.gk_type }}
            <span v-if="gk.missing" class="badge bg-danger ms-2">Missing</span>
          </p>
          <p class="mb-0 text-muted small">
            Owner:
            <RouterLink v-if="gk.owner_id" :to="`/users/${gk.owner_id}`">{{ gk.owner_username }}</RouterLink>
            &ensp;|&ensp;
            Holder:
            <RouterLink v-if="gk.holder_id" :to="`/users/${gk.holder_id}`">{{ gk.holder_username }}</RouterLink>
          </p>
        </div>
        <div class="row g-3 text-center">
          <div class="col">
            <div class="fw-bold text-primary fs-4">{{ gk.total_points_generated?.toLocaleString() }}</div>
            <div class="text-muted small">Points Generated</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.total_moves?.toLocaleString() }}</div>
            <div class="text-muted small">Total Moves</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.distance?.toLocaleString() }} km</div>
            <div class="text-muted small">Distance</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.countries_count?.toLocaleString() }}</div>
            <div class="text-muted small">Countries</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.current_multiplier?.toFixed(2) }}x</div>
            <div class="text-muted small">Multiplier</div>
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

    <!-- Overview -->
    <div v-if="activeTab === 'overview'">
      <div class="card mb-4 shadow-sm">
        <div class="card-header"><b>Points Generated per Day</b></div>
        <div class="card-body">
          <LineChart v-if="timeline.length" :data="timeline" x-key="day" y-key="points" color="#198754" />
          <p v-else class="text-muted text-center py-3">No timeline data.</p>
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
                <td><span class="badge bg-secondary">{{ m.type_name }}</span></td>
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
      <div class="row row-cols-2 row-cols-md-4 row-cols-lg-6 g-2">
        <div v-for="c in countries" :key="c.country" class="col">
          <div class="card text-center p-2 shadow-sm h-100">
            <div class="fw-semibold">{{ c.country }}</div>
            <div class="text-muted small">{{ c.move_count }} moves</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
