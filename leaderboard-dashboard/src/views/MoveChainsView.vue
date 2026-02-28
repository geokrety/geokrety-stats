<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchList } from '../composables/useApi.js'

const route = useRoute()
const moveId = ref(route.params.id)
const chains = ref([])
const loading = ref(false)
const error = ref(null)

async function loadChains() {
  loading.value = true
  error.value = null
  try {
    const { items } = await fetchList(`/moves/${moveId.value}/chains`)
    chains.value = items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadChains)
watch(() => route.params.id, (id) => {
  moveId.value = id
  loadChains()
})
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
              <th>Chain</th>
              <th>GeoKret</th>
              <th>Status</th>
              <th class="text-end">Move Chain Points</th>
              <th class="text-end d-none d-md-table-cell">Total Chain Points</th>
              <th class="text-end d-none d-md-table-cell">Members</th>
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
                <RouterLink :to="`/geokrety/${chain.gk_id}/chains`" class="btn btn-xs btn-outline-secondary ms-2 py-0 px-1" style="font-size:0.75rem">all chains</RouterLink>
              </td>
              <td><span class="badge" :class="chain.status === 'active' ? 'bg-success' : 'bg-secondary'">{{ chain.status }}</span></td>
              <td class="text-end text-primary fw-semibold">{{ chain.move_chain_points?.toFixed(2) }}</td>
              <td class="text-end d-none d-md-table-cell">{{ chain.chain_points?.toFixed(2) }}</td>
              <td class="text-end d-none d-md-table-cell">{{ chain.member_count?.toLocaleString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
