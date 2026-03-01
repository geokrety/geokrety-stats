<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, RouterLink, useRouter } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import { idToGkId } from '../composables/useGkId.js'
import { getMoveTypeBadgeClass } from '../composables/useMoveTypeColors.js'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import { userAvatarUrl, gkAvatarUrl } from '../composables/useAvatarUrl.js'
import LineChart from '../components/LineChart.vue'
import WorldMap from '../components/WorldMap.vue'
import Pagination from '../components/Pagination.vue'
import RelatedUsersTab from '../components/RelatedUsersTab.vue'
import PointsBreakdownChart from '../components/PointsBreakdownChart.vue'
import GkTypeBadge from '../components/GkTypeBadge.vue'

const route  = useRoute()
const router = useRouter()
const userId = ref(route.params.id)

const user        = ref(null)
const timeline    = ref([])
const countries   = ref([])
const moves       = ref([])
const geokrety    = ref([])
const breakdown   = ref([])
const movePage    = ref(1)
const moveMeta    = ref({})
const gkPage      = ref(1)
const gkMeta      = ref({})
const moveSortCol = ref('date')
const moveSortOrder = ref('desc')
const moveAwardingOnly = ref(false)
const selectedMoveTypes = ref([])
const loading     = ref(false)
const error       = ref(null)
const activeTab   = ref('overview')
const loadedTabs  = ref({
  overview: false,
  moves: false,
  geokrety: false,
  countries: false,
  'related-users': false,
})

const moveTypeOptions = [
  { value: 0, label: 'Drop' },
  { value: 1, label: 'Grab' },
  { value: 2, label: 'Comment' },
  { value: 3, label: 'Seen' },
  { value: 4, label: 'Archived' },
  { value: 5, label: 'Dip' },
]

const today = new Date().toISOString().slice(0, 10)

const joinedYear = computed(() => {
  if (!user.value?.joined_at) return null
  return user.value.joined_at.slice(0, 4)
})

const displayTotalPoints = computed(() => {
  const points = Number(user.value?.total_points || 0)
  if (points > 0) return points
  const fallback = Number(user.value?.pts_base || 0)
    + Number(user.value?.pts_relay || 0)
    + Number(user.value?.pts_rescuer || 0)
    + Number(user.value?.pts_chain || 0)
    + Number(user.value?.pts_country || 0)
    + Number(user.value?.pts_diversity || 0)
    + Number(user.value?.pts_handover || 0)
    + Number(user.value?.pts_reach || 0)
  return fallback
})

const avgPointsPerMove = computed(() => {
  const movesCount = Number(user.value?.total_moves || 0)
  if (!movesCount) return 0
  return displayTotalPoints.value / movesCount
})

const selectedMoveTypeLabels = computed(() => {
  if (!selectedMoveTypes.value.length) return 'All types'
  return moveTypeOptions
    .filter((opt) => selectedMoveTypes.value.includes(opt.value))
    .map((opt) => opt.label)
    .join(', ')
})

async function loadUser() {
  loading.value = true
  error.value   = null
  try {
    user.value = await fetchOne(`/users/${userId.value}`)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadOverview() {
  if (loadedTabs.value.overview) return
  const [tl, bd] = await Promise.all([
    fetchList(`/users/${userId.value}/points/timeline`, { per_page: 3650 }),
    fetchList(`/users/${userId.value}/points/breakdown`),
  ])
  timeline.value = tl.items
  breakdown.value = bd.items
  loadedTabs.value.overview = true
}

async function loadMoves() {
  const params = {
    page: movePage.value,
    per_page: 25,
    sort: moveSortCol.value,
    order: moveSortOrder.value,
    awarding_only: moveAwardingOnly.value,
  }
  if (selectedMoveTypes.value.length) {
    params.types = selectedMoveTypes.value.join(',')
  }

  const { items, meta } = await fetchList(`/users/${userId.value}/moves`, params)
  moves.value    = items
  moveMeta.value = meta
  loadedTabs.value.moves = true
}

async function loadCountries() {
  if (loadedTabs.value.countries) return
  const co = await fetchList(`/users/${userId.value}/countries`, { per_page: 300 })
  countries.value = co.items
  loadedTabs.value.countries = true
}

async function loadGeokrety() {
  const { items, meta } = await fetchList(`/users/${userId.value}/geokrety`, {
    page: gkPage.value,
    per_page: 25,
  })
  geokrety.value = items
  gkMeta.value = meta
  loadedTabs.value.geokrety = true
}

async function loadActiveTabData(tab) {
  try {
    if (tab === 'overview') await loadOverview()
    if (tab === 'moves') await loadMoves()
    if (tab === 'geokrety') await loadGeokrety()
    if (tab === 'countries') await loadCountries()
    if (tab === 'related-users') loadedTabs.value['related-users'] = true
  } catch (e) {
    error.value = e.message
  }
}

function toggleMoveSort(col) {
  if (moveSortCol.value === col) {
    moveSortOrder.value = moveSortOrder.value === 'asc' ? 'desc' : 'asc'
    return
  }
  moveSortCol.value = col
  moveSortOrder.value = col === 'gk' || col === 'country' ? 'asc' : 'desc'
}

function moveSortIcon(col) {
  if (moveSortCol.value !== col) return 'bi-sort-down'
  return moveSortOrder.value === 'asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt'
}

function resetMoveTypeFilter() {
  selectedMoveTypes.value = []
}

function toggleMoveType(type) {
  if (selectedMoveTypes.value.includes(type)) {
    selectedMoveTypes.value = selectedMoveTypes.value.filter((t) => t !== type)
    return
  }
  selectedMoveTypes.value = [...selectedMoveTypes.value, type].sort((a, b) => a - b)
}

onMounted(() => {
  // Read tab from URL hash
  const hash = window.location.hash.slice(1)
  if (hash && ['overview', 'moves', 'geokrety', 'countries', 'related-users'].includes(hash)) {
    activeTab.value = hash
  }
  loadUser().then(() => loadActiveTabData(activeTab.value))
})
watch(movePage, () => {
  if (activeTab.value === 'moves') loadMoves()
})
watch(gkPage, () => {
  if (activeTab.value === 'geokrety') loadGeokrety()
})
watch([moveSortCol, moveSortOrder, moveAwardingOnly, selectedMoveTypes], () => {
  movePage.value = 1
  if (activeTab.value === 'moves') loadMoves()
}, { deep: true })
watch(() => route.params.id, async (id) => {
  userId.value = id
  activeTab.value = 'overview'
  timeline.value = []
  countries.value = []
  moves.value = []
  breakdown.value = []
  moveMeta.value = {}
  geokrety.value = []
  gkMeta.value = {}
  movePage.value = 1
  gkPage.value = 1
  moveSortCol.value = 'date'
  moveSortOrder.value = 'desc'
  moveAwardingOnly.value = false
  selectedMoveTypes.value = []
  loadedTabs.value = { overview: false, moves: false, geokrety: false, countries: false, 'related-users': false }
  await loadUser()
  await loadActiveTabData('overview')
})
watch(activeTab, (tab) => {
  // Update URL hash when tab changes
  window.location.hash = tab
  loadActiveTabData(tab)
})
</script>

<template>
  <div v-if="loading && !user" class="text-center py-5">
    <div class="spinner-border"></div>
  </div>
  <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
  <div v-else-if="user">
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">{{ user.username }}</li>
      </ol>
    </nav>

    <!-- User header -->
    <div class="card mb-4 shadow-sm">
      <div class="card-body">
        <div class="row align-items-center g-3">
          <div class="col-auto">
            <img
              v-if="userAvatarUrl(user.avatar)"
              :src="userAvatarUrl(user.avatar)"
              :alt="`${user.username} avatar`"
              class="user-avatar"
            />
            <div v-else class="user-avatar-placeholder">👤</div>
          </div>
          <div class="col">
            <h3 class="mb-1 text-break">{{ user.username }}</h3>
            <p class="text-muted mb-0 small text-nowrap">
              User #{{ user.user_id }}
              <span v-if="joinedYear" class="d-none d-sm-inline"> &mdash; joined {{ joinedYear }}</span>
            </p>
          </div>
          <div class="col-12 col-xl-7 mt-xl-0 mt-3 border-top pt-3 pt-xl-0 border-xl-top-0">
            <div class="row row-cols-2 row-cols-md-3 row-cols-xl-6 g-2 text-center justify-content-center">
              <div class="col">
                <div class="fw-bold text-primary fs-5">{{ displayTotalPoints.toLocaleString(undefined, { maximumFractionDigits: 2 }) }}</div>
                <div class="text-muted small">Points</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ user.rank_all_time?.toLocaleString() || '—' }}</div>
                <div class="text-muted small">Rank</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ user.total_moves?.toLocaleString() }}</div>
                <div class="text-muted small">Moves</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ user.distinct_gks?.toLocaleString() }}</div>
                <div class="text-muted small">GeoKrety</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ user.countries_count?.toLocaleString() }}</div>
                <div class="text-muted small">Countries</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ avgPointsPerMove.toFixed(2) }}</div>
                <div class="text-muted small">Avg/Move</div>
              </div>
            </div>
            <div class="text-center mt-2">
              <RouterLink :to="`/users/${userId}/chains`" class="btn btn-sm btn-outline-secondary">
                <i class="bi bi-link-45deg me-1"></i>Chains
              </RouterLink>
              <RouterLink :to="`/users/${userId}/awards`" class="btn btn-sm btn-outline-primary ms-2">
                <i class="bi bi-award me-1"></i>Awards
              </RouterLink>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <ul class="nav nav-tabs mb-2">
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'overview' }" @click="activeTab = 'overview'">
          <i class="bi bi-bar-chart-line me-1"></i>Overview
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'moves' }" @click="activeTab = 'moves'">
          <i class="bi bi-list-ul me-1"></i>Moves
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'countries' }" @click="activeTab = 'countries'">
          <i class="bi bi-globe me-1"></i>Countries
          <span v-if="user?.countries_count" class="badge bg-secondary ms-1">{{ user.countries_count }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'geokrety' }" @click="activeTab = 'geokrety'">
          <i class="bi bi-box-seam me-1"></i>GeoKrety
          <span v-if="gkMeta.total" class="badge bg-secondary ms-1">{{ gkMeta.total.toLocaleString() }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'related-users' }" @click="activeTab = 'related-users'">
          <i class="bi bi-people me-1"></i>Related Users
        </button>
      </li>
    </ul>

    <!-- Overview tab -->
    <div v-if="activeTab === 'overview'">
      <!-- Points timeline chart -->
      <div class="card mb-4 shadow-sm">
        <div class="card-header d-flex justify-content-between align-items-center">
          <b>Points per Day</b>
        </div>
        <div class="card-body">
          <LineChart
            v-if="timeline.length"
            :data="timeline"
            x-key="day"
            y-key="points"
            color="#0d6efd"
            :height="220"
            :startDate="user?.joined_at?.slice(0, 10)"
            :endDate="today"
            :showRangeButtons="true"
          />
          <p v-else class="text-muted text-center py-3">No timeline data.</p>
        </div>
      </div>
      <!-- Points breakdown chart -->
      <div class="card mb-4 shadow-sm" v-if="breakdown.length">
        <div class="card-header"><b>Points by Bonus Type</b></div>
        <div class="card-body">
          <PointsBreakdownChart :data="breakdown" :height="300" />
        </div>
      </div>
      <!-- Points breakdown table -->
      <div class="card shadow-sm">
        <div class="card-header"><b>Points Breakdown</b></div>
        <div class="table-responsive border-0 mb-0">
          <table class="table table-sm table-hover mb-0 align-middle">
            <thead class="table-light">
              <tr>
                <th title="Activity or bonus type that awarded points to this user">Source</th>
                <th class="text-end" title="Total points earned from this source">Points</th>
                <th class="text-end d-none d-sm-table-cell" title="Number of times this reward was earned by the user">Count</th>
                <th class="text-end" style="width: 50px" title="Actions (View details)"></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="b in breakdown" :key="b.source" @click="$router.push(`/users/${userId}/awards?label=${encodeURIComponent(b.source)}`)" style="cursor: pointer">
                <td class="fw-medium">{{ b.source }}</td>
                <td class="text-end fw-bold text-success">{{ b.points?.toLocaleString() }}</td>
                <td class="text-end d-none d-sm-table-cell text-muted">{{ b.count?.toLocaleString() }}</td>
                <td class="text-end">
                  <RouterLink
                    :to="`/users/${userId}/awards?label=${encodeURIComponent(b.source)}`"
                    class="btn btn-sm btn-outline-secondary py-0 px-1"
                    style="font-size:0.75rem"
                    title="View award details"
                  ><i class="bi bi-eye"></i></RouterLink>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="card-footer text-end">
          <RouterLink :to="`/users/${userId}/awards`" class="btn btn-sm btn-outline-primary shadow-sm">
            <i class="bi bi-list-stars me-1"></i>View all point awards
          </RouterLink>
        </div>
      </div>
    </div>

    <!-- Moves tab -->
    <div v-if="activeTab === 'moves'">
      <div class="d-flex flex-wrap gap-2 align-items-center mb-2">
        <button
          type="button"
          class="btn btn-sm"
          :class="moveAwardingOnly ? 'btn-primary' : 'btn-outline-secondary'"
          @click="moveAwardingOnly = !moveAwardingOnly"
        >
          <i class="bi bi-coin me-1"></i>Only awarding points
        </button>

        <div class="dropdown">
          <button
            class="btn btn-sm btn-outline-secondary dropdown-toggle"
            type="button"
            data-bs-toggle="dropdown"
            data-bs-auto-close="outside"
            aria-expanded="false"
          >
            <i class="bi bi-funnel me-1"></i>{{ selectedMoveTypeLabels }}
          </button>
          <div class="dropdown-menu p-2" style="min-width: 220px;">
            <div class="d-flex justify-content-between align-items-center mb-2">
              <strong class="small">Move types</strong>
              <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="resetMoveTypeFilter">All</button>
            </div>
            <div v-for="opt in moveTypeOptions" :key="opt.value" class="form-check">
              <input
                class="form-check-input"
                type="checkbox"
                :id="`move-type-${opt.value}`"
                :checked="selectedMoveTypes.includes(opt.value)"
                @change="toggleMoveType(opt.value)"
              />
              <label class="form-check-label small" :for="`move-type-${opt.value}`">{{ opt.label }}</label>
            </div>
          </div>
        </div>
      </div>

      <div class="card shadow-sm border-0">
        <div class="table-responsive border-0 mb-0">
          <table class="table table-hover table-sm mb-0 align-middle border">
            <thead class="table-dark">
              <tr>
                <th class="ps-3" style="cursor:pointer" @click="toggleMoveSort('date')" :class="moveSortCol==='date' ? 'text-warning' : ''" title="Date the user logged the move">Date <i class="bi" :class="moveSortIcon('date')"></i></th>
                <th style="cursor:pointer" @click="toggleMoveSort('gk')" :class="moveSortCol==='gk' ? 'text-warning' : ''" title="GeoKret that was moved">GeoKret <i class="bi" :class="moveSortIcon('gk')"></i></th>
                <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleMoveSort('type')" :class="moveSortCol==='type' ? 'text-warning' : ''" title="Type of activity logged (Drop, Grab, Dip, etc.)">Type <i class="bi" :class="moveSortIcon('type')"></i></th>
                <th class="d-none d-sm-table-cell pe-3" style="cursor:pointer" @click="toggleMoveSort('country')" :class="moveSortCol==='country' ? 'text-warning' : ''" title="Country where the activity took place">Country <i class="bi" :class="moveSortIcon('country')"></i></th>
                <th class="text-end" style="cursor:pointer" @click="toggleMoveSort('points')" :class="moveSortCol==='points' ? 'text-warning' : ''" title="Total points earned by the user for this move">Points <i class="bi" :class="moveSortIcon('points')"></i></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in moves" :key="m.move_id" @click="$router.push(`/geokrety/${m.gk_id}`)" style="cursor: pointer">
                <td class="small text-muted ps-3">{{ m.moved_on?.slice(0, 10) }}</td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <img
                      v-if="gkAvatarUrl(m.gk_avatar)"
                      :src="gkAvatarUrl(m.gk_avatar)"
                      :alt="`${m.gk_name || idToGkId(m.gk_id)} avatar`"
                      class="gk-thumb"
                    />
                    <div class="fw-bold text-truncate" style="max-width: 150px">
                      {{ m.gk_name || idToGkId(m.gk_id) }}
                    </div>
                  </div>
                  <div class="d-md-none small">
                    <span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`" style="font-size:0.7rem">{{ m.type_name }}</span>
                  </div>
                </td>
                <td class="d-none d-md-table-cell">
                  <span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`">{{ m.type_name }}</span>
                </td>
                <td class="d-none d-sm-table-cell pe-3">
                  <span v-if="m.country" :title="`Country: ${m.country}`" class="text-nowrap small text-muted">
                    {{ getCountryFlag(m.country) }} {{ m.country.toUpperCase() }}
                  </span>
                </td>
                <td class="text-end fw-bold text-primary">{{ m.points !== null && m.points !== undefined ? m.points.toLocaleString() : '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="moveMeta.total" :meta="moveMeta" v-model:page="movePage" class="mt-3" />
    </div>

    <!-- GeoKrety tab -->
    <div v-if="activeTab === 'geokrety'">
      <div class="card shadow-sm border-0">
        <div class="table-responsive border-0 mb-0">
          <table class="table table-hover table-sm mb-0 align-middle border">
            <thead class="table-dark">
              <tr>
                <th class="ps-3">GeoKret</th>
                <th class="d-none d-md-table-cell">Type</th>
                <th class="d-none d-sm-table-cell">Last interaction</th>
                <th class="text-end">Points Generated</th>
                <th class="text-end pe-3">Multiplier</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="g in geokrety" :key="g.gk_id" @click="$router.push(`/geokrety/${g.gk_id}`)" style="cursor: pointer">
                <td class="ps-3">
                  <div class="d-flex align-items-center gap-2">
                    <img
                      v-if="gkAvatarUrl(g.avatar)"
                      :src="gkAvatarUrl(g.avatar)"
                      :alt="`${g.gk_name || idToGkId(g.gk_id)} avatar`"
                      class="gk-thumb"
                    />
                    <div>
                      <RouterLink :to="`/geokrety/${g.gk_id}`" class="fw-bold text-decoration-none" @click.stop>
                        {{ g.gk_name || idToGkId(g.gk_id) }}
                      </RouterLink>
                      <div class="small text-muted">#{{ g.gk_id }}</div>
                    </div>
                  </div>
                </td>
                <td class="d-none d-md-table-cell">
                  <GkTypeBadge :gk-type="g.gk_type" />
                </td>
                <td class="d-none d-sm-table-cell small text-muted">
                  {{ g.last_interaction ? String(g.last_interaction).slice(0, 10) : '—' }}
                </td>
                <td class="text-end fw-semibold text-primary">{{ Number(g.total_points_generated || 0).toLocaleString() }}</td>
                <td class="text-end pe-3">{{ g.current_multiplier ? Number(g.current_multiplier).toFixed(2) : '1.00' }}×</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="gkMeta.total" :meta="gkMeta" v-model:page="gkPage" class="mt-3" />
    </div>

    <!-- Countries tab -->
    <div v-if="activeTab === 'countries'">
      <div class="card shadow-sm mb-2">
        <div class="card-header"><b>Countries visited</b></div>
        <div class="card-body p-2">
          <WorldMap v-if="countries.length" :countries="countries" :height="380" />
          <p v-else class="text-muted text-center py-3">No countries data.</p>
        </div>
      </div>
      <!-- Country list below map -->
      <div class="row row-cols-2 row-cols-md-4 row-cols-lg-6 g-2">
        <div v-for="c in countries" :key="c.country" class="col">
          <div class="card text-center p-2 shadow-sm h-100">
            <div class="fw-semibold">
              <span class="fs-3">{{ getCountryFlag(c.country) }}</span><br/>
              {{ c.country.toUpperCase() }}
            </div>
            <div class="text-muted small">{{ (c.move_count || c.moves || 0).toLocaleString() }} moves</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Related Users tab -->
    <div v-if="activeTab === 'related-users'">
      <RelatedUsersTab
        :endpoint="`/users/${userId}/related-users`"
        title="Users who moved same GeoKrety"
      />
    </div>
  </div>
</template>

<style scoped>
.user-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--bs-border-color);
}

.user-avatar-placeholder {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  background: var(--bs-light);
}

.gk-thumb {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--bs-border-color);
  flex-shrink: 0;
}
</style>
