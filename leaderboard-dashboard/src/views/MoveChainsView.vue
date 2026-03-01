<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchList } from '../composables/useApi.js'
import PointsValue from '../components/PointsValue.vue'

const route = useRoute()
const moveId = ref(route.params.id)
const chains = ref([])
const sortCol = ref('started')
const sortOrder = ref('desc')
const loading = ref(false)
const error = ref(null)

async function loadChains() {
  loading.value = true
  error.value = null
  try {
    const { items } = await fetchList(`/moves/${moveId.value}/chains`, {
      sort: sortCol.value,
      order: sortOrder.value,
    })
    chains.value = items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadChains)
watch([sortCol, sortOrder], loadChains)
watch(() => route.params.id, (id) => {
  moveId.value = id
  loadChains()
})

function toggleSort(col) {
  if (sortCol.value === col) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
    return
  }
  sortCol.value = col
  sortOrder.value = col === 'gk' || col === 'status' ? 'asc' : 'desc'
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
        <li class="breadcrumb-item active" aria-current="page">Move #{{ moveId }} Chains</li>
      </ol>
    </nav>

    <div class="d-flex align-items-center justify-content-between flex-wrap gap-2 mb-3">
      <h4 class="mb-0"><i class="bi bi-link-45deg me-2"></i>Chains linked to move #{{ moveId }}</h4>
    </div>

    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <div class="card shadow-sm">
      <div class="table-responsive">
        <table class="table table-sm table-hover mb-0 align-middle">
          <thead class="table-dark">
            <tr>
              <th style="cursor:pointer" @click="toggleSort('chain')" title="Chain identifier">Chain <i class="bi" :class="sortIcon('chain')"></i></th>
              <th style="cursor:pointer" @click="toggleSort('gk')" title="GeoKret linked to this chain">GeoKret <i class="bi" :class="sortIcon('gk')"></i></th>
              <th style="cursor:pointer" @click="toggleSort('status')" title="Current chain status">Status <i class="bi" :class="sortIcon('status')"></i></th>
              <th class="text-end" style="cursor:pointer" @click="toggleSort('move_chain_points')" title="Points contributed by this move to the chain">Move Chain Points <i class="bi" :class="sortIcon('move_chain_points')"></i></th>
              <th class="text-end d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('points')" title="Total points of the whole chain">Total Chain Points <i class="bi" :class="sortIcon('points')"></i></th>
              <th class="text-end d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('members')" title="Number of chain members">Members <i class="bi" :class="sortIcon('members')"></i></th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading && !chains.length">
              <td colspan="6" class="text-center py-4"><div class="spinner-border spinner-border-sm me-2"></div>Loading…</td>
            </tr>
            <tr v-else-if="!chains.length">
              <td colspan="6" class="text-center text-muted py-4">No chain links found for this move.</td>
            </tr>
            <tr v-for="chain in chains" :key="chain.chain_id">
              <td><RouterLink :to="`/chains/${chain.chain_id}`" class="fw-semibold">#{{ chain.chain_id }}</RouterLink></td>
              <td>
                <RouterLink :to="`/geokrety/${chain.gk_id}`">{{ chain.gk_hex_id || chain.gk_name || `GK #${chain.gk_id}` }}</RouterLink>
              </td>
              <td><span class="badge" :class="chain.status === 'active' ? 'bg-success' : 'bg-secondary'">{{ chain.status }}</span></td>
              <td class="text-end text-primary fw-semibold"><PointsValue :value="chain.move_chain_points" /></td>
              <td class="text-end d-none d-md-table-cell"><PointsValue :value="chain.chain_points" /></td>
              <td class="text-end d-none d-md-table-cell">{{ chain.member_count?.toLocaleString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
