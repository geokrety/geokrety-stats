<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import Pagination from '../components/Pagination.vue'

const route = useRoute()
const gkId = ref(route.params.id)

const gk = ref(null)
const chains = ref([])
const meta = ref({})
const page = ref(1)
const sortCol = ref('started')
const sortOrder = ref('desc')
const loading = ref(false)
const error = ref(null)

async function loadGK() {
  gk.value = await fetchOne(`/geokrety/${gkId.value}`)
}

async function loadChains() {
  loading.value = true
  error.value = null
  try {
    const { items, meta: m } = await fetchList(`/geokrety/${gkId.value}/chains`, {
      page: page.value,
      per_page: 25,
      sort: sortCol.value,
      order: sortOrder.value,
    })
    chains.value = items
    meta.value = m
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadGK()
  await loadChains()
})

watch([page, sortCol, sortOrder], loadChains)
watch(() => route.params.id, async (id) => {
  gkId.value = id
  page.value = 1
  await loadGK()
  await loadChains()
})

function toggleSort(col) {
  if (sortCol.value === col) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
    return
  }
  sortCol.value = col
  sortOrder.value = col === 'status' ? 'asc' : 'desc'
}

function sortIcon(col) {
  if (sortCol.value !== col) return 'bi-sort-down'
  return sortOrder.value === 'asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt'
}
</script>

<template>
  <div>
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item"><RouterLink to="/geokrety">GeoKrety</RouterLink></li>
        <li class="breadcrumb-item"><RouterLink :to="`/geokrety/${gkId}`">{{ gk?.gk_name || `GK #${gkId}` }}</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">Chains</li>
      </ol>
    </nav>

    <div class="d-flex align-items-center justify-content-between flex-wrap gap-2 mb-3">
      <h4 class="mb-0"><i class="bi bi-link-45deg me-2"></i>GeoKret Chains</h4>
      <RouterLink :to="`/geokrety/${gkId}`" class="btn btn-sm btn-outline-secondary">Back to GeoKret</RouterLink>
    </div>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <div class="card shadow-sm">
      <div class="table-responsive">
        <table class="table table-sm table-hover mb-0 align-middle">
          <thead class="table-dark">
            <tr>
              <th style="cursor:pointer" @click="toggleSort('chain')">Chain <i class="bi" :class="sortIcon('chain')"></i></th>
              <th style="cursor:pointer" @click="toggleSort('status')">Status <i class="bi" :class="sortIcon('status')"></i></th>
              <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('started')">Started <i class="bi" :class="sortIcon('started')"></i></th>
              <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('ended')">Ended <i class="bi" :class="sortIcon('ended')"></i></th>
              <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('last_active')">Last Active <i class="bi" :class="sortIcon('last_active')"></i></th>
              <th class="text-end" style="cursor:pointer" @click="toggleSort('members')">Members <i class="bi" :class="sortIcon('members')"></i></th>
              <th class="text-end" style="cursor:pointer" @click="toggleSort('points')">Points <i class="bi" :class="sortIcon('points')"></i></th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading && !chains.length">
              <td colspan="7" class="text-center py-4"><div class="spinner-border spinner-border-sm me-2"></div>Loading…</td>
            </tr>
            <tr v-else-if="!chains.length">
              <td colspan="7" class="text-center text-muted py-4">No chains found for this GeoKret.</td>
            </tr>
            <tr v-for="chain in chains" :key="chain.chain_id">
              <td><RouterLink :to="`/chains/${chain.chain_id}`" class="fw-semibold">#{{ chain.chain_id }}</RouterLink></td>
              <td><span class="badge" :class="chain.status === 'active' ? 'bg-success' : 'bg-secondary'">{{ chain.status }}</span></td>
              <td class="d-none d-md-table-cell small text-muted">{{ chain.started_at?.slice(0, 10) }}</td>
              <td class="d-none d-md-table-cell small text-muted">{{ chain.ended_at?.slice(0, 10) || '—' }}</td>
              <td class="d-none d-md-table-cell small text-muted">{{ chain.chain_last_active?.slice(0, 10) }}</td>
              <td class="text-end">{{ chain.member_count?.toLocaleString() }}</td>
              <td class="text-end text-primary fw-semibold">{{ chain.chain_points?.toFixed(2) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <Pagination v-if="meta.total" :meta="meta" v-model:page="page" class="mt-3" />
  </div>
</template>
