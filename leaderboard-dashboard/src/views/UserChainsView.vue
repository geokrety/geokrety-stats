<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import Pagination from '../components/Pagination.vue'

const route = useRoute()
const userId = ref(route.params.id)

const user = ref(null)
const chains = ref([])
const meta = ref({})
const page = ref(1)
const loading = ref(false)
const error = ref(null)

async function loadUser() {
  user.value = await fetchOne(`/users/${userId.value}`)
}

async function loadChains() {
  loading.value = true
  error.value = null
  try {
    const { items, meta: m } = await fetchList(`/users/${userId.value}/chains`, { page: page.value, per_page: 25 })
    chains.value = items
    meta.value = m
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadUser()
  await loadChains()
})

watch(page, loadChains)
watch(() => route.params.id, async (id) => {
  userId.value = id
  page.value = 1
  await loadUser()
  await loadChains()
})
</script>

<template>
  <div>
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item"><RouterLink :to="`/users/${userId}`">{{ user?.username || `User #${userId}` }}</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">Chains</li>
      </ol>
    </nav>

    <div class="d-flex align-items-center justify-content-between flex-wrap gap-2 mb-3">
      <h4 class="mb-0"><i class="bi bi-link-45deg me-2"></i>User Chains</h4>
      <RouterLink :to="`/users/${userId}`" class="btn btn-sm btn-outline-secondary">Back to user</RouterLink>
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
              <th class="d-none d-md-table-cell">Started</th>
              <th class="d-none d-md-table-cell">Last Active</th>
              <th class="text-end">Members</th>
              <th class="text-end">Points</th>
              <th class="text-end">Completed</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading && !chains.length">
              <td colspan="8" class="text-center py-4"><div class="spinner-border spinner-border-sm me-2"></div>Loading…</td>
            </tr>
            <tr v-else-if="!chains.length">
              <td colspan="8" class="text-center text-muted py-4">No chains found for this user.</td>
            </tr>
            <tr v-for="chain in chains" :key="chain.chain_id">
              <td><RouterLink :to="`/chains/${chain.chain_id}`" class="fw-semibold">#{{ chain.chain_id }}</RouterLink></td>
              <td>
                <RouterLink :to="`/geokrety/${chain.gk_id}`">{{ chain.gk_hex_id || chain.gk_name || `GK #${chain.gk_id}` }}</RouterLink>
              </td>
              <td><span class="badge" :class="chain.status === 'active' ? 'bg-success' : 'bg-secondary'">{{ chain.status }}</span></td>
              <td class="d-none d-md-table-cell small text-muted">{{ chain.started_at?.slice(0, 10) }}</td>
              <td class="d-none d-md-table-cell small text-muted">{{ chain.chain_last_active?.slice(0, 10) }}</td>
              <td class="text-end">{{ chain.member_count?.toLocaleString() }}</td>
              <td class="text-end text-primary fw-semibold">{{ chain.chain_points?.toFixed(2) }}</td>
              <td class="text-end">
                <span class="badge" :class="chain.has_user_completion ? 'bg-success' : 'bg-light text-dark border'">{{ chain.has_user_completion ? 'yes' : 'no' }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <Pagination v-if="meta.total" :meta="meta" v-model:page="page" class="mt-3" />
  </div>
</template>
