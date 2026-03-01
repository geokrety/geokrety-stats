<script setup>
import { ref, computed, onMounted, watch, defineAsyncComponent } from 'vue'
import { useRoute, RouterLink, useRouter } from 'vue-router'
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
import PointsValue from '../components/PointsValue.vue'
import AwardingOnlyToggle from '../components/AwardingOnlyToggle.vue'
import MoveTypeFilterDropdown from '../components/MoveTypeFilterDropdown.vue'

const route = useRoute()
const router = useRouter()

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
  chains: false,
})

const chainsMeta = ref({})
const GeokretChainsTable = defineAsyncComponent(() => import('../components/chains/GeokretChainsTable.vue'))

const moveTypeOptions = [
  { value: 0, label: 'Drop' },
  { value: 1, label: 'Grab' },
  { value: 2, label: 'Comment' },
  { value: 3, label: 'Seen' },
  { value: 4, label: 'Archived' },
  { value: 5, label: 'Dip' },
]

const validTabs = ['overview', 'moves', 'countries', 'related-users', 'points', 'chains']
const today = new Date().toISOString().slice(0, 10)

const gkMedal = computed(() => {
  const rank = Number(gk.value?.rank_all_time || 0)
  if (rank === 1) return '🥇'
  if (rank === 2) return '🥈'
  if (rank === 3) return '🥉'
  return ''
})

const chartStartDate = computed(() => {
  if (gk.value?.first_move_at) return gk.value.first_move_at.slice(0, 10)
  if (gk.value?.born_at) return gk.value.born_at.slice(0, 10)
  if (gk.value?.created_at) return gk.value.created_at.slice(0, 10)
  return null
})

const avgPointsPerMove = computed(() => {
  const totalMoves = Number(gk.value?.total_moves || 0)
  if (!totalMoves) return 0
  return Number(gk.value?.total_points_generated || 0) / totalMoves
})

const tabCounts = computed(() => ({
  overview: loadedTabs.value.overview ? '1' : '…',
  moves: loadedTabs.value.moves ? String(moveMeta.value.total ?? 0) : '…',
  countries: loadedTabs.value.countries ? String(countries.value.length || 0) : String(gk.value?.countries_count ?? 0),
  'related-users': loadedTabs.value['related-users'] ? String(gk.value?.distinct_users ?? 0) : '…',
  points: loadedTabs.value.points ? String(pointsLogMeta.value.total ?? 0) : '…',
  chains: loadedTabs.value.chains ? String(chainsMeta.value.total ?? 0) : '…',
}))

const statusText = computed(() => {
  if (!gk.value) return '—'
  if (gk.value.missing) return 'Missing'
  if (gk.value.in_cache) return 'In Cache'
  if (gk.value.holder_username) return `Held by ${gk.value.holder_username}`
  return 'Unknown'
})

function parseHashTab(hashValue) {
  const raw = (hashValue || '').replace(/^#/, '')
  if (validTabs.includes(raw)) return raw
  return 'overview'
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

function handleChainsMeta(metaData) {
  chainsMeta.value = metaData || {}
  loadedTabs.value.chains = true
}

function toggleSort(colRef, orderRef, col, ascDefaults = []) {
  if (colRef.value === col) {
    orderRef.value = orderRef.value === 'asc' ? 'desc' : 'asc'
    return
  }
  colRef.value = col
  orderRef.value = ascDefaults.includes(col) ? 'asc' : 'desc'
}

function sortIcon(activeCol, activeOrder, col) {
  if (activeCol !== col) return 'bi-sort-down'
  return activeOrder === 'asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt'
}

onMounted(async () => {
  activeTab.value = parseHashTab(route.hash || window.location.hash)
  await loadGeoKret()
  await loadActiveTabData(activeTab.value)
})

watch(() => route.hash, (hash) => {
  const tab = parseHashTab(hash)
  if (activeTab.value !== tab) activeTab.value = tab
  loadActiveTabData(tab)
})

watch(activeTab, (tab) => {
  const nextHash = `#${tab}`
  if (route.hash !== nextHash) {
    router.push({ path: `/geokrety/${gkId.value}`, hash: nextHash })
  }
  loadActiveTabData(tab)
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
  activeTab.value = parseHashTab(route.hash)

  timeline.value = []
  countries.value = []
  moves.value = []
  pointsLog.value = []
  moveMeta.value = {}
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

  loadedTabs.value = { overview: false, moves: false, countries: false, 'related-users': false, points: false, chains: false }
  chainsMeta.value = {}

  await loadGeoKret()
  await loadActiveTabData(activeTab.value)
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
            <img v-if="gkAvatarUrl(gk.avatar)" :src="gkAvatarUrl(gk.avatar)" :alt="`${gk.gk_name} avatar`" class="gk-avatar" />
            <div v-else class="fs-1">🐢</div>
          </div>

          <div class="col">
            <div class="d-flex align-items-center gap-2 flex-wrap">
              <h3 class="mb-0 text-break d-flex align-items-center gap-2">
                <span v-if="gkMedal" class="display-6 lh-1">{{ gkMedal }}</span>
                <span>{{ gk.gk_name }}</span>
              </h3>
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
              <span v-if="gk.owner_home_country" class="ms-1" :title="`Home country: ${gk.owner_home_country.toUpperCase()}`">{{ getCountryFlag(gk.owner_home_country) }}</span>
            </p>

            <p class="mb-0 text-muted small">
              Status:
              <span class="badge bg-secondary ms-1">{{ statusText }}</span>
              <span v-if="gk.in_cache && gk.cache_country" class="ms-1" :title="`Cache location: ${gk.cache_country.toUpperCase()}`">{{ getCountryFlag(gk.cache_country) }}</span>
              <span v-else-if="!gk.in_cache && gk.holder_home_country" class="ms-1" :title="`Holder country: ${gk.holder_home_country.toUpperCase()}`">{{ getCountryFlag(gk.holder_home_country) }}</span>
            </p>
          </div>

          <div class="col-12 col-xl-7 mt-xl-0 mt-3 border-top pt-3 pt-xl-0 border-xl-top-0">
            <div class="row row-cols-2 row-cols-md-3 row-cols-xl-6 g-2 text-center justify-content-center">
              <div class="col">
                <div class="fw-bold text-success fs-5"><PointsValue :value="gk.total_points_generated" /></div>
                <div class="text-muted small">Points Generated</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ gk.total_moves?.toLocaleString() }}</div>
                <div class="text-muted small">Moves</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ gk.distance_km?.toLocaleString() }} km</div>
                <div class="text-muted small">Distance</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ gk.countries_count?.toLocaleString() }}</div>
                <div class="text-muted small">Countries</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ gk.distinct_caches?.toLocaleString() }}</div>
                <div class="text-muted small">Places</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5">{{ Number(gk.current_multiplier || 1).toFixed(2) }}×</div>
                <div class="text-muted small">Multiplier</div>
              </div>
            </div>
            <div class="text-center mt-2">
              <RouterLink :to="`/geokrety/${gkId}/chains`" class="btn btn-sm btn-outline-secondary">
                <i class="bi bi-link-45deg me-1"></i>Chains
              </RouterLink>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ul class="nav nav-tabs mb-2">
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'overview' }" title="Overview panels and charts" @click="activeTab = 'overview'">
          <i class="bi bi-bar-chart-line me-1"></i>Overview <span class="badge bg-secondary ms-1">{{ tabCounts.overview }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'moves' }" title="Chronological move history" @click="activeTab = 'moves'">
          <i class="bi bi-list-ul me-1"></i>Moves <span class="badge bg-secondary ms-1">{{ tabCounts.moves }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'countries' }" title="Countries visited by this GeoKret" @click="activeTab = 'countries'">
          <i class="bi bi-globe me-1"></i>Countries <span class="badge bg-secondary ms-1">{{ tabCounts.countries }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'related-users' }" title="Users who moved this GeoKret" @click="activeTab = 'related-users'">
          <i class="bi bi-people me-1"></i>Movers <span class="badge bg-secondary ms-1">{{ tabCounts['related-users'] }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'points' }" title="Detailed points log for this GeoKret" @click="activeTab = 'points'">
          <i class="bi bi-coin me-1"></i>Points Log <span class="badge bg-secondary ms-1">{{ tabCounts.points }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'chains' }" title="Chains this GeoKret joined" @click="activeTab = 'chains'">
          <i class="bi bi-link-45deg me-1"></i>Chains <span class="badge bg-secondary ms-1">{{ tabCounts.chains }}</span>
        </button>
      </li>
    </ul>

    <div v-if="activeTab === 'overview'">
      <MoveTypeBreakdown
        title="Move Type Breakdown"
        :drops="gk.total_drops"
        :grabs="gk.total_grabs"
        :dips="gk.total_dips"
        :seen="gk.total_seen"
        :comments="gk.total_comments"
        :loves="gk.loves_count"
      />

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

      <div class="row g-3">
        <div class="col-12 col-lg-6">
          <div class="card shadow-sm h-100">
            <div class="card-header bg-light"><h5 class="mb-0">Reach</h5></div>
            <div class="card-body">
              <ul class="list-group list-group-flush small">
                <li class="list-group-item d-flex justify-content-between px-0"><span>People who interacted</span><span class="fw-semibold">{{ gk.distinct_users?.toLocaleString() }}</span></li>
                <li class="list-group-item d-flex justify-content-between px-0"><span>Different places visited</span><span class="fw-semibold">{{ gk.distinct_caches?.toLocaleString() }}</span></li>
                <li class="list-group-item d-flex justify-content-between px-0"><span>Countries reached</span><span class="fw-semibold">{{ gk.countries_count?.toLocaleString() }}</span></li>
                <li class="list-group-item d-flex justify-content-between px-0"><span>Users awarded</span><span class="fw-semibold">{{ gk.users_awarded?.toLocaleString() }}</span></li>
                <li class="list-group-item d-flex justify-content-between px-0"><span>Avg/move</span><span class="fw-semibold"><PointsValue :value="avgPointsPerMove" /></span></li>
              </ul>
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="card shadow-sm h-100">
            <div class="card-header bg-light"><h5 class="mb-0">Dates</h5></div>
            <div class="card-body">
              <ul class="list-group list-group-flush small">
                <li class="list-group-item d-flex justify-content-between px-0" v-if="gk.born_at || gk.created_at"><span>Created</span><span class="fw-semibold">{{ (gk.born_at || gk.created_at)?.slice(0, 10) }}</span></li>
                <li class="list-group-item d-flex justify-content-between px-0" v-if="gk.first_move_at"><span>First move</span><span class="fw-semibold">{{ gk.first_move_at?.slice(0, 10) }}</span></li>
                <li class="list-group-item d-flex justify-content-between px-0" v-if="gk.last_move_at"><span>Last move</span><span class="fw-semibold">{{ gk.last_move_at?.slice(0, 10) }}</span></li>
                <li class="list-group-item d-flex justify-content-between px-0"><span>Lifetime / age</span><span class="fw-semibold">TODO</span></li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'moves'">
      <div class="d-flex flex-wrap gap-2 align-items-center mb-2">
        <AwardingOnlyToggle v-model="moveAwardingOnly" />
        <MoveTypeFilterDropdown v-model="selectedMoveTypes" :options="moveTypeOptions" id-prefix="gk-move-type" />
      </div>

      <div class="card shadow-sm border-0">
        <div class="table-responsive border-0 mb-0">
          <table class="table table-hover table-sm mb-0 align-middle border">
            <thead class="table-dark">
              <tr>
                <th class="ps-3" style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'date')" :class="moveSortCol==='date' ? 'text-warning' : ''" title="Move date">Date <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'date')"></i></th>
                <th style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'author', ['author'])" :class="moveSortCol==='author' ? 'text-warning' : ''" title="Move author">Author <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'author')"></i></th>
                <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'type')" :class="moveSortCol==='type' ? 'text-warning' : ''" title="Move type">Type <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'type')"></i></th>
                <th class="d-none d-sm-table-cell" style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'country', ['country'])" :class="moveSortCol==='country' ? 'text-warning' : ''" title="Country">Country <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'country')"></i></th>
                <th class="d-none d-lg-table-cell" style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'waypoint', ['waypoint'])" :class="moveSortCol==='waypoint' ? 'text-warning' : ''" title="Waypoint identifier">Waypoint <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'waypoint')"></i></th>
                <th class="text-end pe-3" style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'points')" :class="moveSortCol==='points' ? 'text-warning' : ''" title="Points awarded for this move">Points <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'points')"></i></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in moves" :key="m.move_id" @click="m.author_id && $router.push(`/users/${m.author_id}`)" style="cursor: pointer">
                <td class="small text-muted ps-3">{{ m.moved_on?.slice(0, 10) }}</td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <img v-if="userAvatarUrl(m.author_avatar)" :src="userAvatarUrl(m.author_avatar)" :alt="`${m.author_username || m.author_id || 'Unknown'} avatar`" class="author-avatar" />
                    <div>
                      <div class="fw-bold text-truncate" style="max-width: 140px">{{ m.author_username || 'Anonymous' }}</div>
                      <div class="d-md-none small mt-1"><span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`" style="font-size:0.7rem">{{ m.type_name }}</span></div>
                    </div>
                  </div>
                </td>
                <td class="d-none d-md-table-cell"><span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`">{{ m.type_name }}</span></td>
                <td class="d-none d-sm-table-cell"><span v-if="m.country" :title="`Country: ${m.country}`" class="text-nowrap small text-muted">{{ getCountryFlag(m.country) }} {{ m.country.toUpperCase() }}</span><span v-else class="text-muted small">—</span></td>
                <td class="d-none d-lg-table-cell">
                  <a v-if="m.waypoint" :href="waypointExternalUrl(m.waypoint)" target="_blank" rel="noopener" @click.stop class="text-decoration-none font-monospace small" :title="waypointTooltip(m.waypoint)">{{ displayWaypoint(m.waypoint) }}</a>
                  <span v-else class="text-muted small">—</span>
                </td>
                <td class="text-end fw-bold pe-3"><PointsValue :value="m.points" /></td>
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
            <div class="fw-semibold"><span class="fs-3">{{ getCountryFlag(c.country) }}</span><br/>{{ c.country.toUpperCase() }}</div>
            <div class="text-muted small">{{ c.move_count?.toLocaleString() }} moves</div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'related-users'">
      <RelatedUsersTab :endpoint="`/geokrety/${gkId}/related-users`" title="Users who moved this GeoKret" />
    </div>

    <div v-if="activeTab === 'points'">
      <div class="d-flex flex-wrap gap-2 align-items-center mb-2">
        <AwardingOnlyToggle v-model="pointsAwardingOnly" />
        <MoveTypeFilterDropdown v-model="selectedPointsMoveTypes" :options="moveTypeOptions" id-prefix="gk-points-type" />
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
                <th class="ps-3" style="cursor:pointer" @click="toggleSort(pointsSortCol, pointsSortOrder, 'date')" :class="pointsSortCol==='date' ? 'text-warning' : ''" title="Award date">Date <i class="bi" :class="sortIcon(pointsSortCol, pointsSortOrder, 'date')"></i></th>
                <th style="cursor:pointer" @click="toggleSort(pointsSortCol, pointsSortOrder, 'user', ['user'])" :class="pointsSortCol==='user' ? 'text-warning' : ''" title="User receiving points">User <i class="bi" :class="sortIcon(pointsSortCol, pointsSortOrder, 'user')"></i></th>
                <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort(pointsSortCol, pointsSortOrder, 'label', ['label'])" :class="pointsSortCol==='label' ? 'text-warning' : ''" title="Reward category label">Reward <i class="bi" :class="sortIcon(pointsSortCol, pointsSortOrder, 'label')"></i></th>
                <th class="d-none d-sm-table-cell" style="cursor:pointer" @click="toggleSort(pointsSortCol, pointsSortOrder, 'type')" :class="pointsSortCol==='type' ? 'text-warning' : ''" title="Move type tied to reward">Type <i class="bi" :class="sortIcon(pointsSortCol, pointsSortOrder, 'type')"></i></th>
                <th class="d-none d-sm-table-cell" style="cursor:pointer" @click="toggleSort(pointsSortCol, pointsSortOrder, 'country', ['country'])" :class="pointsSortCol==='country' ? 'text-warning' : ''" title="Country where reward event happened">Country <i class="bi" :class="sortIcon(pointsSortCol, pointsSortOrder, 'country')"></i></th>
                <th class="text-end pe-3" style="cursor:pointer" @click="toggleSort(pointsSortCol, pointsSortOrder, 'points')" :class="pointsSortCol==='points' ? 'text-warning' : ''" title="Awarded points value">Points <i class="bi" :class="sortIcon(pointsSortCol, pointsSortOrder, 'points')"></i></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in pointsLog" :key="p.id" @click="$router.push(`/users/${p.user_id}`)" style="cursor: pointer">
                <td class="small text-muted text-nowrap ps-3">{{ p.awarded_at?.slice(0, 10) }}</td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <img v-if="userAvatarUrl(p.author_avatar)" :src="userAvatarUrl(p.author_avatar)" :alt="`${p.username || p.user_id} avatar`" class="author-avatar" />
                    <div>
                      <div class="fw-bold text-truncate" style="max-width: 150px">{{ p.username || p.user_id }}</div>
                      <div class="d-md-none small mt-1"><span class="badge bg-light text-dark border overflow-hidden text-truncate" style="max-width: 160px">{{ (p.label || '—').replace(/_/g, ' ') }}</span></div>
                    </div>
                  </div>
                </td>
                <td class="d-none d-md-table-cell"><span class="badge bg-light text-dark border" :title="p.reason || ''">{{ (p.label || '—').replace(/_/g, ' ') }}</span></td>
                <td class="d-none d-sm-table-cell"><span :class="`badge ${getMoveTypeBadgeClass(p.type_name)}`">{{ p.type_name }}</span></td>
                <td class="d-none d-sm-table-cell small text-muted"><span v-if="p.country">{{ getCountryFlag(p.country) }} {{ p.country.toUpperCase() }}</span><span v-else>—</span></td>
                <td class="text-end fw-bold pe-3"><PointsValue :value="p.points" :show-plus="true" /></td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="pointsLogMeta.total > 0" :meta="pointsLogMeta" v-model:page="pointsLogPage" class="mt-3" />
    </div>
    <div v-else-if="activeTab === 'chains'">
      <Suspense>
        <template #default>
          <GeokretChainsTable :gk-id="gkId" @meta-updated="handleChainsMeta" />
        </template>
        <template #fallback>
          <div class="text-center text-muted py-5">
            <div class="spinner-border spinner-border-sm me-2"></div>
            Loading chains…
          </div>
        </template>
      </Suspense>
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
