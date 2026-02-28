<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import Pagination from '../components/Pagination.vue'

const route = useRoute()
const chainId = ref(route.params.id)

const chain = ref(null)
const members = ref([])
const membersMeta = ref({})
const membersPage = ref(1)
const moves = ref([])
const movesMeta = ref({})
const movesPage = ref(1)
const loading = ref(false)
const error = ref(null)

async function loadDetail() {
  chain.value = await fetchOne(`/chains/${chainId.value}`)
}

async function loadMembers() {
  const { items, meta } = await fetchList(`/chains/${chainId.value}/members`, { page: membersPage.value, per_page: 25 })
  members.value = items
  membersMeta.value = meta
}

async function loadMoves() {
  const { items, meta } = await fetchList(`/chains/${chainId.value}/moves`, { page: movesPage.value, per_page: 25 })
  moves.value = items
  movesMeta.value = meta
}

async function loadAll() {
  loading.value = true
  error.value = null
  try {
    await loadDetail()
    await loadMembers()
    await loadMoves()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)
watch(membersPage, loadMembers)
watch(movesPage, loadMoves)
watch(() => route.params.id, async (id) => {
  chainId.value = id
  membersPage.value = 1
  movesPage.value = 1
  await loadAll()
})
</script>

<template>
  <div>
    <nav aria-label="breadcrumb" class="mb-2" v-if="chain">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item"><RouterLink :to="`/geokrety/${chain.gk_id}`">{{ chain.gk_hex_id || chain.gk_name }}</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">Chain #{{ chain.chain_id }}</li>
      </ol>
    </nav>

    <div v-if="loading && !chain" class="text-center py-5"><div class="spinner-border"></div></div>
    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>

    <div v-else-if="chain">
      <div class="card mb-3 shadow-sm">
        <div class="card-body d-flex align-items-center justify-content-between flex-wrap gap-2">
          <div>
            <h4 class="mb-1"><i class="bi bi-diagram-3 me-2"></i>Chain #{{ chain.chain_id }}</h4>
            <div class="text-muted small">
              GeoKret:
              <RouterLink :to="`/geokrety/${chain.gk_id}`">{{ chain.gk_hex_id || chain.gk_name || `GK #${chain.gk_id}` }}</RouterLink>
              · Started {{ chain.started_at?.slice(0, 10) }}
              · Last active {{ chain.chain_last_active?.slice(0, 10) }}
            </div>
          </div>
          <div class="text-end">
            <div><span class="badge" :class="chain.status === 'active' ? 'bg-success' : 'bg-secondary'">{{ chain.status }}</span></div>
            <div class="small text-muted mt-1">Members: {{ chain.member_count?.toLocaleString() }} · Points: {{ chain.chain_points?.toFixed(2) }}</div>
          </div>
        </div>
        <div class="card-footer d-flex gap-2 flex-wrap">
          <RouterLink :to="`/geokrety/${chain.gk_id}/chains`" class="btn btn-sm btn-outline-secondary">All chains for this GeoKret</RouterLink>
        </div>
      </div>

      <div class="row g-3">
        <div class="col-12 col-xl-5">
          <div class="card shadow-sm h-100">
            <div class="card-header"><b>Members</b></div>
            <div class="table-responsive">
              <table class="table table-sm table-hover mb-0">
                <thead class="table-light">
                  <tr>
                    <th>#</th>
                    <th>User</th>
                    <th class="d-none d-md-table-cell">Joined</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!members.length">
                    <td colspan="3" class="text-center text-muted py-3">No members.</td>
                  </tr>
                  <tr v-for="member in members" :key="member.user_id">
                    <td>{{ member.position }}</td>
                    <td>
                      <RouterLink :to="`/users/${member.user_id}`">{{ member.username }}</RouterLink>
                      <RouterLink :to="`/users/${member.user_id}/chains`" class="btn btn-xs btn-outline-secondary ms-2 py-0 px-1" style="font-size:0.75rem">chains</RouterLink>
                    </td>
                    <td class="d-none d-md-table-cell small text-muted">{{ member.joined_at?.slice(0, 10) || '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="card-body py-2" v-if="membersMeta.total">
              <Pagination :meta="membersMeta" v-model:page="membersPage" />
            </div>
          </div>
        </div>

        <div class="col-12 col-xl-7">
          <div class="card shadow-sm h-100">
            <div class="card-header"><b>Moves in chain window</b></div>
            <div class="table-responsive">
              <table class="table table-sm table-hover mb-0 align-middle">
                <thead class="table-light">
                  <tr>
                    <th>Date</th>
                    <th>User</th>
                    <th class="d-none d-md-table-cell">Type</th>
                    <th class="text-end">Chain pts</th>
                    <th class="text-end">Links</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!moves.length">
                    <td colspan="5" class="text-center text-muted py-3">No moves.</td>
                  </tr>
                  <tr v-for="move in moves" :key="move.move_id">
                    <td class="small text-muted">{{ move.moved_on?.slice(0, 10) || '—' }}</td>
                    <td>
                      <RouterLink v-if="move.author_id" :to="`/users/${move.author_id}`">{{ move.author_username || `User #${move.author_id}` }}</RouterLink>
                      <span v-else class="text-muted">—</span>
                    </td>
                    <td class="d-none d-md-table-cell"><span class="badge bg-light text-dark border">{{ move.type_name }}</span></td>
                    <td class="text-end fw-semibold text-primary">{{ move.chain_points?.toFixed(2) }}</td>
                    <td class="text-end">
                      <RouterLink :to="`/moves/${move.move_id}/chains`" class="btn btn-xs btn-outline-secondary py-0 px-1" style="font-size:0.75rem">move</RouterLink>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="card-body py-2" v-if="movesMeta.total">
              <Pagination :meta="movesMeta" v-model:page="movesPage" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
