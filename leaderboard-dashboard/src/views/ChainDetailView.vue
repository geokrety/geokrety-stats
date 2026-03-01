<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import Pagination from '../components/Pagination.vue'
import PointsValue from '../components/PointsValue.vue'

const route = useRoute()
const chainId = ref(route.params.id)

const chain = ref(null)
const members = ref([])
const membersMeta = ref({})
const membersPage = ref(1)
const membersSort = ref('position')
const membersOrder = ref('asc')
const moves = ref([])
const movesMeta = ref({})
const movesPage = ref(1)
const movesSort = ref('date')
const movesOrder = ref('desc')
const loading = ref(false)
const error = ref(null)

async function loadDetail() {
  chain.value = await fetchOne(`/chains/${chainId.value}`)
}

async function loadMembers() {
  const { items, meta } = await fetchList(`/chains/${chainId.value}/members`, {
    page: membersPage.value,
    per_page: 25,
    sort: membersSort.value,
    order: membersOrder.value,
  })
  members.value = items
  membersMeta.value = meta
}

async function loadMoves() {
  const { items, meta } = await fetchList(`/chains/${chainId.value}/moves`, {
    page: movesPage.value,
    per_page: 25,
    sort: movesSort.value,
    order: movesOrder.value,
  })
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
watch([membersPage, membersSort, membersOrder], loadMembers)
watch([movesPage, movesSort, movesOrder], loadMoves)
watch(() => route.params.id, async (id) => {
  chainId.value = id
  membersPage.value = 1
  movesPage.value = 1
  await loadAll()
})

function toggleMembersSort(col) {
  if (membersSort.value === col) {
    membersOrder.value = membersOrder.value === 'asc' ? 'desc' : 'asc'
    return
  }
  membersSort.value = col
  membersOrder.value = col === 'user' ? 'asc' : 'desc'
}

function toggleMovesSort(col) {
  if (movesSort.value === col) {
    movesOrder.value = movesOrder.value === 'asc' ? 'desc' : 'asc'
    return
  }
  movesSort.value = col
  movesOrder.value = col === 'user' ? 'asc' : 'desc'
}

function sortIcon(activeCol, col, order) {
  if (activeCol !== col) return 'bi-sort-down'
  return order === 'asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt'
}
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
            <div class="small text-muted mt-1">Members: {{ chain.member_count?.toLocaleString() }} · Points: <PointsValue :value="chain.chain_points" /></div>
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
                    <th style="cursor:pointer" @click="toggleMembersSort('position')" title="Position of the user in chain order"># <i class="bi" :class="sortIcon(membersSort, 'position', membersOrder)"></i></th>
                    <th style="cursor:pointer" @click="toggleMembersSort('user')" title="Chain member username">User <i class="bi" :class="sortIcon(membersSort, 'user', membersOrder)"></i></th>
                    <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleMembersSort('joined')" title="Date user joined the chain">Joined <i class="bi" :class="sortIcon(membersSort, 'joined', membersOrder)"></i></th>
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
                    <th style="cursor:pointer" @click="toggleMovesSort('date')" title="Move date">Date <i class="bi" :class="sortIcon(movesSort, 'date', movesOrder)"></i></th>
                    <th style="cursor:pointer" @click="toggleMovesSort('user')" title="User who performed the move">User <i class="bi" :class="sortIcon(movesSort, 'user', movesOrder)"></i></th>
                    <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleMovesSort('type')" title="Move type">Type <i class="bi" :class="sortIcon(movesSort, 'type', movesOrder)"></i></th>
                    <th class="text-end" style="cursor:pointer" @click="toggleMovesSort('chain_points')" title="Points awarded for this move in chain context">Chain pts <i class="bi" :class="sortIcon(movesSort, 'chain_points', movesOrder)"></i></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!moves.length">
                    <td colspan="4" class="text-center text-muted py-3">No moves.</td>
                  </tr>
                  <tr v-for="move in moves" :key="move.move_id">
                    <td class="small text-muted">{{ move.moved_on?.slice(0, 10) || '—' }}</td>
                    <td>
                      <RouterLink v-if="move.author_id" :to="`/users/${move.author_id}`">{{ move.author_username || `User #${move.author_id}` }}</RouterLink>
                      <span v-else class="text-muted">—</span>
                    </td>
                    <td class="d-none d-md-table-cell"><span class="badge bg-light text-dark border">{{ move.type_name }}</span></td>
                    <td class="text-end fw-semibold text-primary"><PointsValue :value="move.chain_points" /></td>
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
