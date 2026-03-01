<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import { getMoveTypeBadgeClass } from '../composables/useMoveTypeColors.js'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import { gkAvatarUrl, userAvatarUrl } from '../composables/useAvatarUrl.js'
import Pagination from '../components/Pagination.vue'
import PointsValue from '../components/PointsValue.vue'
import GkTypeBadge from '../components/GkTypeBadge.vue'
import MoveTypeBadge from '../components/MoveTypeBadge.vue'
import VTooltip from '../components/VTooltip.vue'

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
      <!-- Hero Header inspired by Geokret/User Detail -->
      <div class="row mb-4 g-3 align-items-stretch">
        <div class="col-12 col-lg-8">
          <div class="card h-100 shadow-sm border-0 bg-dark text-white overflow-hidden hero-card">
            <div class="card-body p-4 position-relative d-flex flex-column justify-content-center">
              <div class="d-flex align-items-center gap-4 flex-wrap flex-md-nowrap">
                <div class="flex-shrink-0 hero-avatar-container">
                  <img :src="gkAvatarUrl(chain.gk_avatar_key)"
                       class="rounded hero-avatar"
                       style="width: 100px; height: 100px; object-fit: cover; border: 3px solid rgba(255,255,255,0.2)"
                       @error="e => e.target.src = '/gk-default.png'" />
                </div>
                <div class="flex-grow-1">
                  <div class="d-flex align-items-center gap-2 mb-2 flex-wrap">
                    <h2 class="mb-0 fw-bold">Chain #{{ chain.chain_id }}</h2>
                    <span class="badge" :class="chain.status === 'active' ? 'bg-success' : 'bg-secondary'">{{ chain.status }}</span>
                  </div>
                  <div class="fs-5 mb-3 opacity-75 d-flex align-items-center gap-2 flex-wrap">
                    <i class="bi bi-box-seam"></i>
                    <RouterLink :to="`/geokrety/${chain.gk_id}`" class="text-white text-decoration-none hover-underline">
                      {{ chain.gk_name || `GK #${chain.gk_id}` }}
                    </RouterLink>
                    <span class="opacity-50 mx-1">·</span>
                    <GkTypeBadge :type="chain.gk_type" />
                  </div>
                  <div class="d-flex gap-4 flex-wrap opacity-75 small">
                    <div class="d-flex align-items-center gap-2" title="Chain start date">
                      <i class="bi bi-calendar-check text-primary"></i> Started: <b>{{ chain.started_at?.slice(0, 10) }}</b>
                    </div>
                    <div class="d-flex align-items-center gap-2" title="Last activity recorded in this chain">
                      <i class="bi bi-clock-history text-info"></i> Last Active: <b>{{ chain.chain_last_active?.slice(0, 10) }}</b>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-4">
          <div class="card h-100 shadow-sm border-0 bg-primary text-white text-center d-flex flex-column justify-content-center p-3 overflow-hidden">
            <div class="position-absolute top-0 end-0 p-3 opacity-25">
              <i class="bi bi-diagram-3" style="font-size: 5rem; transform: rotate(-15deg); display: block;"></i>
            </div>
            <div class="position-relative">
              <div class="display-5 fw-bold mb-0">
                <PointsValue :value="chain.chain_points" />
              </div>
              <div class="text-uppercase small opacity-75 ls-1 fw-semibold mb-3">Total Chain Points</div>
              <div class="d-flex justify-content-center gap-3 border-top border-white border-opacity-25 pt-3">
                <div class="text-center px-2">
                  <div class="h4 mb-0 fw-bold">{{ chain.member_count?.toLocaleString() || 0 }}</div>
                  <div class="x-small text-uppercase opacity-75">Members</div>
                </div>
                <div class="text-center px-2" v-if="chain.moves_count">
                  <div class="h4 mb-0 fw-bold">{{ chain.moves_count }}</div>
                  <div class="x-small text-uppercase opacity-75">Moves</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="mb-4">
        <div class="d-flex align-items-center gap-2 mb-2">
          <h5 class="mb-0 fw-bold">Chain Description</h5>
          <hr class="flex-grow-1 opacity-25" />
        </div>
        <p class="text-muted mb-0 lead-sm p-3 bg-light rounded border-start border-primary border-4">
          This movement chain tracks a sequence of <b>{{ chain.member_count }}</b> unique participants
          moving <b>{{ chain.gk_name || 'this GeoKret' }}</b>. It has accumulated a total of <b>{{ chain.chain_points }}</b> points
          since it began on <b>{{ chain.started_at?.slice(0, 10) }}</b>.
        </p>
      </div>

      <div class="row g-4">
        <div class="col-12 col-xl-5">
          <div class="card shadow-sm border-0 overflow-hidden h-100">
            <div class="card-header bg-white py-3 d-flex align-items-center justify-content-between">
              <h6 class="mb-0 fw-bold"><i class="bi bi-people me-2 text-primary"></i>Chain Members</h6>
              <span class="badge bg-primary rounded-pill">{{ membersMeta.total || 0 }}</span>
            </div>
            <div class="card-body py-2" v-if="membersMeta.total">
              This card displays the members of the chain along with their positions and join dates
            </div>
            <div class="table-responsive">
              <table class="table table-hover mb-0 align-middle">
                <thead class="table-light">
                  <tr>
                    <th class="ps-3" style="cursor:pointer" @click="toggleMembersSort('position')">
                      <v-tooltip text="Position of the user in the chain sequence">
                        <template #activator="{ props }">
                          <span v-bind="props">
                            # <i class="bi" :class="sortIcon(membersSort, 'position', membersOrder)"></i>
                          </span>
                        </template>
                      </v-tooltip>
                    </th>
                    <th style="cursor:pointer" @click="toggleMembersSort('user')">
                      <v-tooltip text="The username of the chain participant">
                        <template #activator="{ props }">
                          <span v-bind="props">
                            User <i class="bi" :class="sortIcon(membersSort, 'user', membersOrder)"></i>
                          </span>
                        </template>
                      </v-tooltip>
                    </th>
                    <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleMembersSort('joined')">
                      <v-tooltip text="The date this user joined the chain">
                        <template #activator="{ props }">
                          <span v-bind="props">
                            Joined <i class="bi" :class="sortIcon(membersSort, 'joined', membersOrder)"></i>
                          </span>
                        </template>
                      </v-tooltip>
                    </th>
                  </tr>
                </thead>
                <tbody class="border-top-0">
                  <tr v-if="!members.length">
                    <td colspan="3" class="text-center text-muted py-5">
                      <i class="bi bi-people display-4 opacity-25 d-block mb-2"></i>
                      No members found for this chain.
                    </td>
                  </tr>
                  <tr v-for="member in members" :key="member.user_id">
                    <td class="ps-3 fw-bold text-muted">{{ member.position }}</td>
                    <td>
                      <div class="d-flex align-items-center gap-2">
                        <img :src="userAvatarUrl(member.user_avatar_key)" width="24" height="24" class="rounded-circle border" @error="e => e.target.src = '/user-default.png'" />
                        <RouterLink :to="`/users/${member.user_id}`" class="fw-semibold text-decoration-none">{{ member.username }}</RouterLink>
                      </div>
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
          <div class="card shadow-sm border-0 overflow-hidden h-100">
            <div class="card-header bg-white py-3 d-flex align-items-center justify-content-between">
              <h6 class="mb-0 fw-bold"><i class="bi bi-activity me-2 text-info"></i>Moves History</h6>
              <span class="badge bg-info text-white rounded-pill">{{ movesMeta.total || 0 }}</span>
            </div>
            <div class="table-responsive">
              <table class="table table-hover mb-0 align-middle">
                <thead class="table-light">
                  <tr>
                    <th class="ps-3" style="cursor:pointer" @click="toggleMovesSort('date')">
                      <v-tooltip text="Date the move was performed">
                        <template #activator="{ props }">
                          <span v-bind="props">
                            Date <i class="bi" :class="sortIcon(movesSort, 'date', movesOrder)"></i>
                          </span>
                        </template>
                      </v-tooltip>
                    </th>
                    <th style="cursor:pointer" @click="toggleMovesSort('user')">
                      <v-tooltip text="The user who logged this move">
                        <template #activator="{ props }">
                          <span v-bind="props">
                            User <i class="bi" :class="sortIcon(movesSort, 'user', movesOrder)"></i>
                          </span>
                        </template>
                      </v-tooltip>
                    </th>
                    <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleMovesSort('type')">
                      <v-tooltip text="The type of the log (e.g., Drop, Grab, Seen)">
                        <template #activator="{ props }">
                          <span v-bind="props">
                            Type <i class="bi" :class="sortIcon(movesSort, 'type', movesOrder)"></i>
                          </span>
                        </template>
                      </v-tooltip>
                    </th>
                    <th class="text-end pe-3" style="cursor:pointer" @click="toggleMovesSort('chain_points')">
                      <v-tooltip text="Points awarded specifically for this chain completion">
                        <template #activator="{ props }">
                          <span v-bind="props">
                            Chain pts <i class="bi" :class="sortIcon(movesSort, 'chain_points', movesOrder)"></i>
                          </span>
                        </template>
                      </v-tooltip>
                    </th>
                  </tr>
                </thead>
                <tbody class="border-top-0">
                  <tr v-if="!moves.length">
                    <td colspan="4" class="text-center text-muted py-5">
                      <i class="bi bi-journal-text display-4 opacity-25 d-block mb-2"></i>
                      No moves recorded in this chain window.
                    </td>
                  </tr>
                  <tr v-for="move in moves" :key="move.move_id">
                    <td class="ps-3 small text-muted">{{ move.moved_on?.slice(0, 10) || '—' }}</td>
                    <td>
                      <div class="d-flex align-items-center gap-2">
                        <img :src="userAvatarUrl(move.author_avatar_key)" width="20" height="20" class="rounded-circle border" v-if="move.author_id" @error="e => e.target.src = '/user-default.png'" />
                        <RouterLink v-if="move.author_id" :to="`/users/${move.author_id}`" class="text-decoration-none small fw-semibold">{{ move.author_username || `User #${move.author_id}` }}</RouterLink>
                        <span v-else class="text-muted small">—</span>
                      </div>
                    </td>
                    <td class="d-none d-md-table-cell">
                      <MoveTypeBadge :type="move.log_type" :text="move.type_name" />
                    </td>
                    <td class="text-end pe-3 fw-bold text-primary"><PointsValue :value="move.chain_points" /></td>
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
