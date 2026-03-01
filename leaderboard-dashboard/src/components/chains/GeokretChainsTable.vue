<script setup>
import { ref, watch, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { fetchList } from '../../composables/useApi.js'
import Pagination from '../Pagination.vue'
import PointsValue from '../PointsValue.vue'

const props = defineProps({
  gkId: { type: [String, Number], required: true },
})
const emit = defineEmits(['meta-updated'])

const chains = ref([])
const meta = ref({})
const page = ref(1)
const sortCol = ref('started')
const sortOrder = ref('desc')
const loading = ref(false)
const error = ref(null)

async function loadChains() {
  if (!props.gkId) return
  loading.value = true
  error.value = null
  try {
    const { items, meta: m } = await fetchList(`/geokrety/${props.gkId}/chains`, {
      page: page.value,
      per_page: 25,
      sort: sortCol.value,
      order: sortOrder.value,
    })
    chains.value = items
    meta.value = m
    emit('meta-updated', m)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

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

onMounted(loadChains)

watch(() => props.gkId, () => {
  chains.value = []
  meta.value = {}
  page.value = 1
  sortCol.value = 'started'
  sortOrder.value = 'desc'
  loadChains()
})

watch([page, sortCol, sortOrder], () => {
  loadChains()
})
</script>

<template>
  <div>
    <div v-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-if="loading && !chains.length" class="text-center py-4">
      <div class="spinner-border spinner-border-sm me-2"></div>Loading chains…
    </div>
    <div v-else>
      <div class="table-responsive border-0 mb-3">
        <table class="table table-sm table-hover mb-0 align-middle">
          <thead class="table-dark">
            <tr>
              <th style="cursor:pointer" @click="toggleSort('chain')" title="Chain identifier">Chain <i class="bi" :class="sortIcon('chain')"></i></th>
              <th style="cursor:pointer" @click="toggleSort('status')" title="Current chain status">Status <i class="bi" :class="sortIcon('status')"></i></th>
              <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('started')" title="Chain start date">Started <i class="bi" :class="sortIcon('started')"></i></th>
              <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('ended')" title="Chain end date">Ended <i class="bi" :class="sortIcon('ended')"></i></th>
              <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('last_active')" title="Last activity date">Last Active <i class="bi" :class="sortIcon('last_active')"></i></th>
              <th class="text-end" style="cursor:pointer" @click="toggleSort('members')" title="Number of chain members">Members <i class="bi" :class="sortIcon('members')"></i></th>
              <th class="text-end" style="cursor:pointer" @click="toggleSort('points')" title="Total chain points">Points <i class="bi" :class="sortIcon('points')"></i></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="chain in chains" :key="chain.chain_id">
              <td><RouterLink :to="`/chains/${chain.chain_id}`" class="fw-semibold">#{{ chain.chain_id }}</RouterLink></td>
              <td>
                <span class="badge" :class="chain.status === 'active' ? 'bg-success' : 'bg-secondary'">{{ chain.status }}</span>
              </td>
              <td class="d-none d-md-table-cell small text-muted">{{ chain.started_at?.slice(0, 10) }}</td>
              <td class="d-none d-md-table-cell small text-muted">{{ chain.ended_at?.slice(0, 10) || '—' }}</td>
              <td class="d-none d-md-table-cell small text-muted">{{ chain.chain_last_active?.slice(0, 10) }}</td>
              <td class="text-end"><span class="text-muted">{{ chain.member_count?.toLocaleString() || '0' }}</span></td>
              <td class="text-end text-primary fw-semibold"><PointsValue :value="chain.chain_points" /></td>
            </tr>
            <tr v-if="!chains.length && !loading">
              <td colspan="7" class="text-center text-muted py-4">No chains found for this GeoKret.</td>
            </tr>
          </tbody>
        </table>
      </div>
      <Pagination v-if="meta.total" :meta="meta" v-model:page="page" />
      <div v-else-if="!loading" class="text-muted text-center small">No chain data available.</div>
    </div>
  </div>
</template>
