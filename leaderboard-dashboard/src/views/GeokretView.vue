<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import { idToGkId } from '../composables/useGkId.js'
import { getMoveTypeBadgeClass } from '../composables/useMoveTypeColors.js'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import { gkAvatarUrl, userAvatarUrl } from '../composables/useAvatarUrl.js'
import { waypointExternalUrl, displayWaypoint, waypointTooltip } from '../composables/useWaypoint.js'
import GkTypeBadge from '../components/GkTypeBadge.vue'
import LineChart from '../components/LineChart.vue'
import WorldMap from '../components/WorldMap.vue'
import Pagination from '../components/Pagination.vue'
import RelatedUsersTab from '../components/RelatedUsersTab.vue'
import MoveTypeBreakdown from '../components/MoveTypeBreakdown.vue'

const route = useRoute()
const gkId = ref(route.params.id)

const gk = ref(null)
const timeline = ref([])
const countries = ref([])
const moves = ref([])
const movePage = ref(1)
const moveMeta = ref({})
const moveSortCol = ref('date')
const moveSortOrder = ref('desc')
const moveAwardingOnly = ref(false)
const selectedMoveTypes = ref([])

const pointsLog = ref([])
const pointsLogPage = ref(1)
const pointsLogMeta = ref({})
const pointsSortCol = ref('date')
const pointsSortOrder = ref('desc')
const pointsAwardingOnly = ref(false)
const selectedPointsMoveTypes = ref([])

const loading = ref(false)
const error = ref(null)
const activeTab = ref('overview')
const loadedTabs = ref({
  overview: false,
  moves: false,
  countries: false,
  'related-users': false,
  points: false,
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

const chartStartDate = computed(() => {
  if (gk.value?.first_move_at) return gk.value.first_move_at.slice(0, 10)
  if (gk.value?.born_at) return gk.value.born_at.slice(0, 10)
  if (gk.value?.created_at) return gk.value.created_at.slice(0, 10)
  return null
})

const moveTypeFilterLabel = computed(() => {
  if (!selectedMoveTypes.value.length) return 'All types'
  return moveTypeOptions
    .filter((opt) => selectedMoveTypes.value.includes(opt.value))
    .map((opt) => opt.label)
    .join(', ')
})

const pointsTypeFilterLabel = computed(() => {
  if (!selectedPointsMoveTypes.value.length) return 'All types'
  return moveTypeOptions
    .filter((opt) => selectedPointsMoveTypes.value.includes(opt.value))
    .map((opt) => opt.label)
    .join(', ')
})

const lifetimeText = computed(() => {
  const start = gk.value?.born_at || gk.value?.created_at || gk.value?.first_move_at
  if (!start) return '—'
  return formatRelativeDuration(start)
})

const inactiveText = computed(() => {
  if (!gk.value?.last_move_at) return '—'
  return formatRelativeDuration(gk.value.last_move_at)
})

function formatRelativeDuration(inputDate) {
  const dt = new Date(inputDate)
  if (Number.isNaN(dt.getTime())) return '—'
  const diffMs = Date.now() - dt.getTime()
  if (diffMs <= 0) return '0 days'

  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  if (days < 30) return `${days} day${days === 1 ? '' : 's'}`

  const months = Math.floor(days / 30)
  if (months < 24) return `${months} month${months === 1 ? '' : 's'}`

  const years = Math.floor(months / 12)
  return `${years} year${years === 1 ? '' : 's'}`
}

async function loadGeoKret() {
  loading.value = true
  error.value = null
  try {
    gk.value = await fetchOne(`/geokrety/${gkId.value}`)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadOverview() {
  if (loadedTabs.value.overview) return
  const tl = await fetchList(`/geokrety/${gkId.value}/points/timeline`, { per_page: 3650 })
  timeline.value = tl.items
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

  const { items, meta } = await fetchList(`/geokrety/${gkId.value}/moves`, params)
  moves.value = items
  moveMeta.value = meta
  loadedTabs.value.moves = true
}

async function loadCountries() {
  if (loadedTabs.value.countries) return
  const co = await fetchList(`/geokrety/${gkId.value}/countries`, { per_page: 300 })
  countries.value = co.items
  loadedTabs.value.countries = true
}

async function loadPointsLog() {
  const params = {
    page: pointsLogPage.value,
    per_page: 25,
    sort: pointsSortCol.value,
    order: pointsSortOrder.value,
    awarding_only: pointsAwardingOnly.value,
  }
  if (selectedPointsMoveTypes.value.length) {
    params.types = selectedPointsMoveTypes.value.join(',')
  }

  const { items, meta } = await fetchList(`/geokrety/${gkId.value}/points/log`, params)
  pointsLog.value = items
  pointsLogMeta.value = meta
  loadedTabs.value.points = true
}

async function loadActiveTabData(tab) {
  try {
    if (tab === 'overview') await loadOverview()
    if (tab === 'moves') await loadMoves()
    if (tab === 'countries') await loadCountries()
    if (tab === 'points') await loadPointsLog()
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
  moveSortOrder.value = col === 'author' || col === 'country' || col === 'waypoint' ? 'asc' : 'desc'
}

function moveSortIcon(col) {
  if (moveSortCol.value !== col) return 'bi-sort-down'
  return moveSortOrder.value === 'asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt'
}

function togglePointsSort(col) {
  if (pointsSortCol.value === col) {
    pointsSortOrder.value = pointsSortOrder.value === 'asc' ? 'desc' : 'asc'
    return
  }
  pointsSortCol.value = col
  pointsSortOrder.value = col === 'user' || col === 'label' || col === 'country' ? 'asc' : 'desc'
}

function pointsSortIcon(col) {
  if (pointsSortCol.value !== col) return 'bi-sort-down'
  return pointsSortOrder.value === 'asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt'
}

function toggleMoveType(type) {
  if (selectedMoveTypes.value.includes(type)) {
    selectedMoveTypes.value = selectedMoveTypes.value.filter((t) => t !== type)
    return
  }
  selectedMoveTypes.value = [...selectedMoveTypes.value, type].sort((a, b) => a - b)
}

function resetMoveTypeFilter() {
  selectedMoveTypes.value = []
}

function togglePointsMoveType(type) {
  if (selectedPointsMoveTypes.value.includes(type)) {
    selectedPointsMoveTypes.value = selectedPointsMoveTypes.value.filter((t) => t !== type)
    return
  }
  selectedPointsMoveTypes.value = [...selectedPointsMoveTypes.value, type].sort((a, b) => a - b)
}

function resetPointsTypeFilter() {
  selectedPointsMoveTypes.value = []
}

onMounted(async () => {
  const hash = window.location.hash.slice(1)
  if (hash && ['overview', 'moves', 'countries', 'related-users', 'points'].includes(hash)) {
    activeTab.value = hash
  }
  await loadGeoKret()
  await loadActiveTabData(activeTab.value)
})

watch(movePage, () => {
  if (activeTab.value === 'moves') loadMoves()
})
watch(pointsLogPage, () => {
  if (activeTab.value === 'points') loadPointsLog()
})
watch([moveSortCol, moveSortOrder, moveAwardingOnly, selectedMoveTypes], () => {
  movePage.value = 1
  if (activeTab.value === 'moves') loadMoves()
}, { deep: true })
watch([pointsSortCol, pointsSortOrder, pointsAwardingOnly, selectedPointsMoveTypes], () => {
  pointsLogPage.value = 1
  if (activeTab.value === 'points') loadPointsLog()
}, { deep: true })
watch(() => route.params.id, async (id) => {
  gkId.value = id
  activeTab.value = 'overview'
  timeline.value = []
  countries.value = []
  moves.value = []
  moveMeta.value = {}
  pointsLog.value = []
  pointsLogMeta.value = {}
  movePage.value = 1
  pointsLogPage.value = 1
  moveSortCol.value = 'date'
  moveSortOrder.value = 'desc'
  pointsSortCol.value = 'date'
  pointsSortOrder.value = 'desc'
  moveAwardingOnly.value = false
  pointsAwardingOnly.value = false
  selectedMoveTypes.value = []
  selectedPointsMoveTypes.value = []
  loadedTabs.value = { overview: false, moves: false, countries: false, 'related-users': false, points: false }
  await loadGeoKret()
  await loadActiveTabData('overview')
})
watch(activeTab, (tab) => {
  window.location.hash = tab
  loadActiveTabData(tab)
})
</script>

<template>
  <div v-if="loading && !gk" class="text-center py-5">
    <div class="spinner-border"></div>
  </div>
  <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
  <div v-else-if="gk">
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item"><RouterLink to="/geokrety">GeoKrety</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">{{ gk.gk_name }}</li>
      </ol>
    </nav>

    <div class="card mb-4 shadow-sm">
      <div class="card-body">
        <div class="row align-items-center g-3">
          <div class="col-auto">
            <img
              v-if="gkAvatarUrl(gk.avatar)"
              :src="gkAvatarUrl(gk.avatar)"
              :alt="`${gk.gk_name} avatar`"
              class="gk-avatar"
            />
            <div v-else class="fs-1">🐢</div>
          </div>

          <div class="col">
            <div class="d-flex align-items-center gap-2 flex-wrap">
              <h3 class="mb-0 text-break">{{ gk.gk_name }}</h3>
              <span class="badge bg-dark" style="font-size: 0.8rem">{{ gk.gk_hex_id || idToGkId(gk.gk_id) }}</span>
              <span v-if="gk.missing" class="badge bg-danger">⚠️ Missing</span>
              <span v-if="gk.is_non_collectible" class="badge bg-warning text-dark" title="Non-transferable (sealed) GeoKret">🔒 Sealed</span>
              <span v-if="gk.is_parked" class="badge bg-info text-dark" title="Parked GeoKret">🅿️ Parked</span>
            </div>

            <p class="mb-0 text-muted small mt-1">
              Type: <GkTypeBadge :gk-type="gk.gk_type" :type-name="gk.gk_type_name" />
              <span v-if="gk.loves_count" class="ms-2">❤️ {{ gk.loves_count.toLocaleString() }} loves</span>
            </p>

            <p class="mb-0 text-muted small">
              Owner:
              <RouterLink v-if="gk.owner_id" :to="`/users/${gk.owner_id}`">{{ gk.owner_username }}</RouterLink>
              <span v-else>—</span>
              <span v-if="gk.owner_home_country" class="ms-1" :title="`Home country: ${gk.owner_home_country.toUpperCase()}`">
                {{ getCountryFlag(gk.owner_home_country) }}
              </span>
            </p>

            <p v-if="gk.in_cache" class="mb-0 text-muted small">
              In cache:
              <span class="badge bg-success ms-1">🏦 In Cache</span>
              <span v-if="gk.cache_country" class="ms-1" :title="`Cache location: ${gk.cache_country.toUpperCase()}`">
                {{ getCountryFlag(gk.cache_country) }}
              </span>
            </p>
            <p v-else class="mb-0 text-muted small">
              Holder:
              <RouterLink v-if="gk.holder_id" :to="`/users/${gk.holder_id}`">{{ gk.holder_username }}</RouterLink>
              <span v-else>—</span>
              <span v-if="gk.holder_home_country" class="ms-1" :title="`Home country: ${gk.holder_home_country.toUpperCase()}`">
                {{ getCountryFlag(gk.holder_home_country) }}
              </span>
            </p>
          </div>

          <div class="col-12 col-xl-7 mt-xl-0 mt-3 border-top pt-3 border-xl-top-0 pt-xl-0">
            <div class="row g-2 text-center justify-content-center">
              <div class="col-4 col-sm-auto mb-2 px-1">
                <div class="fw-bold text-success fs-5">{{ gk.total_points_generated?.toLocaleString() }}</div>
                <div class="text-muted" style="font-size: 0.65rem">Points</div>
              </div>
              <div class="col-4 col-sm-auto mb-2 px-1">
                <div class="fw-bold fs-5">{{ gk.total_moves?.toLocaleString() }}</div>
                <div class="text-muted" style="font-size: 0.65rem">Moves</div>
              </div>
              <div class="col-4 col-sm-auto mb-2 px-1">
                <div class="fw-bold fs-5">{{ gk.distance_km?.toLocaleString() }} km</div>
                <div class="text-muted" style="font-size: 0.65rem">Dist.</div>
              </div>
              <div class="col-4 col-sm-auto mb-2 px-1">
                <div class="fw-bold fs-5">{{ gk.countries_count?.toLocaleString() }}</div>
                <div class="text-muted" style="font-size: 0.65rem">Countries</div>
              </div>
              <div class="col-4 col-sm-auto mb-2 px-1">
                <div class="fw-bold fs-5">{{ gk.distinct_caches?.toLocaleString() }}</div>
                <div class="text-muted" style="font-size: 0.65rem">Places</div>
              </div>
              <div class="col-4 col-sm-auto mb-2 px-1">
                <div class="fw-bold fs-5">{{ gk.current_multiplier?.toFixed(2) }}×</div>
                <div class="text-muted" style="font-size: 0.65rem">Mult.</div>
              </div>
            </div>
            <div class="text-center mt-2">
              <RouterLink :to="`/geokrety/${gkId}/chains`" class="btn btn-sm btn-outline-secondary">
                <i class="bi bi-link-45deg me-1"></i>Chains
              </RouterLink>
              <button type="button" class="btn btn-sm btn-outline-primary ms-2" @click="activeTab = 'points'">
                <i class="bi bi-award me-1"></i>Awards
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ul class="nav nav-tabs mb-2">
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'overview' }" @click="activeTab = 'overview'">
          <i class="bi bi-bar-chart-line me-1"></i>Overview
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'moves' }" @click="activeTab = 'moves'">
          <i class="bi bi-list-ul me-1"></i>Moves
          <span v-if="moveMeta.total" class="badge bg-secondary ms-1">{{ moveMeta.total.toLocaleString() }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'countries' }" @click="activeTab = 'countries'">
          <i class="bi bi-globe me-1"></i>Countries
          <span v-if="countries.length" class="badge bg-secondary ms-1">{{ countries.length }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'related-users' }" @click="activeTab = 'related-users'">
          <i class="bi bi-people me-1"></i>Movers
          <span v-if="gk?.distinct_users" class="badge bg-secondary ms-1">{{ gk.distinct_users.toLocaleString() }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'points' }" @click="activeTab = 'points'">
          <i class="bi bi-coin me-1"></i>Points Log
          <span v-if="pointsLogMeta.total" class="badge bg-success ms-1">{{ pointsLogMeta.total.toLocaleString() }}</span>
        </button>
      </li>
    </ul>

    <div v-if="activeTab === 'overview'">
      <div class="card mb-4 shadow-sm">
        <div class="card-header d-flex justify-content-between align-items-center bg-light">
          <b>Points Generated per Day</b>
          <span class="text-muted small" v-if="chartStartDate">since {{ chartStartDate }}</span>
        </div>
        <div class="card-body">
          <LineChart
            v-if="timeline.length"
            :data="timeline"
            x-key="day"
            y-key="points"
            color="#198754"
            :height="220"
            :startDate="chartStartDate"
            :endDate="today"
            :showRangeButtons="true"
          />
          <p v-else class="text-muted text-center py-3">No timeline data.</p>
        </div>
      </div>

      <MoveTypeBreakdown
        title="Move Type Breakdown"
        :drops="gk.total_drops"
        :grabs="gk.total_grabs"
        :dips="gk.total_dips"
        :seen="gk.total_seen"
        :comments="gk.total_comments"
        :loves="gk.loves_count"
      />

      <div class="row g-3">
        <div class="col-12 col-lg-6">
          <div class="card shadow-sm h-100">
            <div class="card-header bg-light">
              <h5 class="mb-0">Reach</h5>
            </div>
            <div class="card-body">
              <ul class="list-group list-group-flush small">
                <li class="list-group-item d-flex justify-content-between px-0">
                  <span>People who interacted</span>
                  <span class="fw-semibold">{{ gk.distinct_users?.toLocaleString() }}</span>
                </li>
                <li class="list-group-item d-flex justify-content-between px-0">
                  <span>Different places visited</span>
                  <span class="fw-semibold">{{ gk.distinct_caches?.toLocaleString() }}</span>
                </li>
                <li class="list-group-item d-flex justify-content-between px-0">
                  <span>Countries reached</span>
                  <span class="fw-semibold">{{ gk.countries_count?.toLocaleString() }}</span>
                </li>
                <li class="list-group-item d-flex justify-content-between px-0">
                  <span>Users awarded</span>
                  <span class="fw-semibold">{{ gk.users_awarded?.toLocaleString() }}</span>
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="card shadow-sm h-100">
            <div class="card-header bg-light">
              <h5 class="mb-0">Dates</h5>
            </div>
            <div class="card-body">
              <ul class="list-group list-group-flush small">
                <li class="list-group-item d-flex justify-content-between px-0" v-if="gk.born_at || gk.created_at">
                  <span>Created</span>
                  <span class="fw-semibold">{{ (gk.born_at || gk.created_at)?.slice(0, 10) }}</span>
                </li>
                <li class="list-group-item d-flex justify-content-between px-0" v-if="gk.first_move_at">
                  <span>First move</span>
                  <span class="fw-semibold">{{ gk.first_move_at?.slice(0, 10) }}</span>
                </li>
                <li class="list-group-item d-flex justify-content-between px-0" v-if="gk.last_move_at">
                  <span>Last move</span>
                  <span class="fw-semibold">{{ gk.last_move_at?.slice(0, 10) }}</span>
                </li>
                <li class="list-group-item d-flex justify-content-between px-0">
                  <span>Lifetime / age</span>
                  <span class="fw-semibold">{{ lifetimeText }}</span>
                </li>
                <li class="list-group-item d-flex justify-content-between px-0">
                  <span>Inactive for</span>
                  <span class="fw-semibold">{{ inactiveText }}</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>

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
            <i class="bi bi-funnel me-1"></i>{{ moveTypeFilterLabel }}
          </button>
          <div class="dropdown-menu p-2" style="min-width: 220px;">
            <div class="d-flex justify-content-between align-items-center mb-2">
              <strong class="small">Move types</strong>
              <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="resetMoveTypeFilter">All</button>
            </div>
            <div v-for="opt in moveTypeOptions" :key="`move-opt-${opt.value}`" class="form-check">
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
                <th class="ps-3" style="cursor:pointer" @click="toggleMoveSort('date')" :class="moveSortCol==='date' ? 'text-warning' : ''">Date <i class="bi" :class="moveSortIcon('date')"></i></th>
                <th style="cursor:pointer" @click="toggleMoveSort('author')" :class="moveSortCol==='author' ? 'text-warning' : ''">Author <i class="bi" :class="moveSortIcon('author')"></i></th>
                <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleMoveSort('type')" :class="moveSortCol==='type' ? 'text-warning' : ''">Type <i class="bi" :class="moveSortIcon('type')"></i></th>
                <th class="d-none d-sm-table-cell" style="cursor:pointer" @click="toggleMoveSort('country')" :class="moveSortCol==='country' ? 'text-warning' : ''">Country <i class="bi" :class="moveSortIcon('country')"></i></th>
                <th class="d-none d-lg-table-cell" style="cursor:pointer" @click="toggleMoveSort('waypoint')" :class="moveSortCol==='waypoint' ? 'text-warning' : ''">Waypoint <i class="bi" :class="moveSortIcon('waypoint')"></i></th>
                <th class="text-end pe-3" style="cursor:pointer" @click="toggleMoveSort('points')" :class="moveSortCol==='points' ? 'text-warning' : ''">Points <i class="bi" :class="moveSortIcon('points')"></i></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in moves" :key="m.move_id" @click="m.author_id && $router.push(`/users/${m.author_id}`)" style="cursor: pointer">
                <td class="small text-muted ps-3">{{ m.moved_on?.slice(0, 10) }}</td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <img
                      v-if="userAvatarUrl(m.author_avatar)"
                      :src="userAvatarUrl(m.author_avatar)"
                      :alt="`${m.author_username || m.author_id || 'Unknown'} avatar`"
                      class="author-avatar"
                    />
                    <div>
                      <div class="fw-bold text-truncate" style="max-width: 140px">{{ m.author_username || 'Anonymous' }}</div>
                      <div class="d-md-none small mt-1">
                        <span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`" style="font-size:0.7rem">{{ m.type_name }}</span>
                      </div>
                    </div>
                  </div>
                </td>
                <td class="d-none d-md-table-cell">
                  <span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`">{{ m.type_name }}</span>
                </td>
                <td class="d-none d-sm-table-cell">
                  <span v-if="m.country" :title="`Country: ${m.country}`" class="text-nowrap small text-muted">
                    {{ getCountryFlag(m.country) }} {{ m.country.toUpperCase() }}
                  </span>
                  <span v-else class="text-muted small">—</span>
                </td>
                <td class="d-none d-lg-table-cell">
                  <a v-if="m.waypoint"
                     :href="waypointExternalUrl(m.waypoint)"
                     target="_blank"
                     rel="noopener"
                     @click.stop
                     class="text-decoration-none font-monospace small"
                     :title="waypointTooltip(m.waypoint)">
                    {{ displayWaypoint(m.waypoint) }}
                  </a>
                  <span v-else class="text-muted small">—</span>
                </td>
                <td class="text-end fw-bold pe-3" :class="m.points > 0 ? 'text-success' : 'text-muted'">
                  {{ m.points !== null && m.points !== undefined ? m.points.toLocaleString() : '—' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="moveMeta.total" :meta="moveMeta" v-model:page="movePage" class="mt-3" />
    </div>

    <div v-if="activeTab === 'countries'">
      <div class="card shadow-sm mb-2">
        <div class="card-header bg-light"><b>Countries visited</b></div>
        <div class="card-body p-2">
          <WorldMap v-if="countries.length" :countries="countries" :height="380" />
          <p v-else class="text-muted text-center py-3">No countries data.</p>
        </div>
      </div>
      <div class="row row-cols-2 row-cols-md-4 row-cols-lg-6 g-2">
        <div v-for="c in countries" :key="c.country" class="col">
          <div class="card text-center p-2 shadow-sm h-100">
            <div class="fw-semibold">
              <span class="fs-3">{{ getCountryFlag(c.country) }}</span><br/>
              {{ c.country.toUpperCase() }}
            </div>
            <div class="text-muted small">{{ c.move_count?.toLocaleString() }} moves</div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'related-users'">
      <RelatedUsersTab
        :endpoint="`/geokrety/${gkId}/related-users`"
        title="Users who moved this GeoKret"
      />
    </div>

    <div v-if="activeTab === 'points'">
      <div class="d-flex flex-wrap gap-2 align-items-center mb-2">
        <button
          type="button"
          class="btn btn-sm"
          :class="pointsAwardingOnly ? 'btn-primary' : 'btn-outline-secondary'"
          @click="pointsAwardingOnly = !pointsAwardingOnly"
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
            <i class="bi bi-funnel me-1"></i>{{ pointsTypeFilterLabel }}
          </button>
          <div class="dropdown-menu p-2" style="min-width: 220px;">
            <div class="d-flex justify-content-between align-items-center mb-2">
              <strong class="small">Move types</strong>
              <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="resetPointsTypeFilter">All</button>
            </div>
            <div v-for="opt in moveTypeOptions" :key="`points-opt-${opt.value}`" class="form-check">
              <input
                class="form-check-input"
                type="checkbox"
                :id="`points-type-${opt.value}`"
                :checked="selectedPointsMoveTypes.includes(opt.value)"
                @change="togglePointsMoveType(opt.value)"
              />
              <label class="form-check-label small" :for="`points-type-${opt.value}`">{{ opt.label }}</label>
            </div>
          </div>
        </div>
      </div>

      <div v-if="pointsLog.length === 0" class="text-center text-muted py-5">
        <i class="bi bi-inbox fs-1 d-block mb-2"></i>
        No points recorded for this GeoKret yet.
      </div>
      <div v-else class="card shadow-sm border-0">
        <div class="table-responsive border-0 mb-0">
          <table class="table table-hover table-sm mb-0 align-middle border">
            <thead class="table-dark">
              <tr>
                <th class="ps-3" style="cursor:pointer" @click="togglePointsSort('date')" :class="pointsSortCol==='date' ? 'text-warning' : ''">Date <i class="bi" :class="pointsSortIcon('date')"></i></th>
                <th style="cursor:pointer" @click="togglePointsSort('user')" :class="pointsSortCol==='user' ? 'text-warning' : ''">User <i class="bi" :class="pointsSortIcon('user')"></i></th>
                <th class="d-none d-md-table-cell" style="cursor:pointer" @click="togglePointsSort('label')" :class="pointsSortCol==='label' ? 'text-warning' : ''">Reward <i class="bi" :class="pointsSortIcon('label')"></i></th>
                <th class="d-none d-sm-table-cell" style="cursor:pointer" @click="togglePointsSort('type')" :class="pointsSortCol==='type' ? 'text-warning' : ''">Type <i class="bi" :class="pointsSortIcon('type')"></i></th>
                <th class="d-none d-sm-table-cell" style="cursor:pointer" @click="togglePointsSort('country')" :class="pointsSortCol==='country' ? 'text-warning' : ''">Country <i class="bi" :class="pointsSortIcon('country')"></i></th>
                <th class="text-end pe-3" style="cursor:pointer" @click="togglePointsSort('points')" :class="pointsSortCol==='points' ? 'text-warning' : ''">Points <i class="bi" :class="pointsSortIcon('points')"></i></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in pointsLog" :key="p.id" @click="$router.push(`/users/${p.user_id}`)" style="cursor: pointer">
                <td class="small text-muted text-nowrap ps-3">{{ p.awarded_at?.slice(0, 10) }}</td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <img
                      v-if="userAvatarUrl(p.author_avatar)"
                      :src="userAvatarUrl(p.author_avatar)"
                      :alt="`${p.username || p.user_id} avatar`"
                      class="author-avatar"
                    />
                    <div>
                      <div class="fw-bold text-truncate" style="max-width: 150px">{{ p.username || p.user_id }}</div>
                      <div class="d-md-none small mt-1">
                        <span class="badge bg-light text-dark border overflow-hidden text-truncate" style="max-width: 160px">
                          {{ (p.label || '—').replace(/_/g, ' ') }}
                        </span>
                      </div>
                    </div>
                  </div>
                </td>
                <td class="d-none d-md-table-cell">
                  <span class="badge bg-light text-dark border" :title="p.reason || ''">
                    {{ (p.label || '—').replace(/_/g, ' ') }}
                  </span>
                </td>
                <td class="d-none d-sm-table-cell">
                  <span :class="`badge ${getMoveTypeBadgeClass(p.type_name)}`">{{ p.type_name }}</span>
                </td>
                <td class="d-none d-sm-table-cell small text-muted">
                  <span v-if="p.country">{{ getCountryFlag(p.country) }} {{ p.country.toUpperCase() }}</span>
                  <span v-else>—</span>
                </td>
                <td class="text-end fw-bold pe-3" :class="p.points >= 0 ? 'text-success' : 'text-danger'">
                  {{ p.points >= 0 ? '+' : '' }}{{ p.points?.toLocaleString() }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="pointsLogMeta.total > 0" :meta="pointsLogMeta" v-model:page="pointsLogPage" class="mt-3" />
    </div>
  </div>
</template>

<style scoped>
.gk-avatar {
  width: 68px;
  height: 68px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--bs-border-color);
}

.author-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--bs-border-color);
  flex-shrink: 0;
}
</style>
