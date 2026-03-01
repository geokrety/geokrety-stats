<script setup>
import { ref, computed, onMounted, watch, defineAsyncComponent } from 'vue'
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
import PointsValue from '../components/PointsValue.vue'
import AwardingOnlyToggle from '../components/AwardingOnlyToggle.vue'
import MoveTypeFilterDropdown from '../components/MoveTypeFilterDropdown.vue'
import GeokretTypeFilterDropdown from '../components/GeokretTypeFilterDropdown.vue'
import { useAwardLabels } from '../composables/useAwardLabels.js'

const route = useRoute()
const router = useRouter()

const userId = ref(route.params.id)
const user = ref(null)
const timeline = ref([])
const countries = ref([])
const moves = ref([])
const geokrety = ref([])
const breakdown = ref([])
const awards = ref([])
const { labels: availableAwardLabels } = useAwardLabels(userId)

const movePage = ref(1)
const moveMeta = ref({})
const gkPage = ref(1)
const gkMeta = ref({})
const awardsPage = ref(1)
const awardsMeta = ref({})

const moveSortCol = ref('date')
const moveSortOrder = ref('desc')
const gkSortCol = ref('last_interaction')
const gkSortOrder = ref('desc')
const awardsSortCol = ref('date')
const awardsSortOrder = ref('desc')

const moveAwardingOnly = ref(false)
const selectedMoveTypes = ref([])

const gkAwardingOnly = ref(false)
const gkMultiplierOnly = ref(false)
const selectedGkTypes = ref([])

const awardsLabelFilter = ref('')

const chainsMeta = ref({})

const loading = ref(false)
const error = ref(null)
const activeTab = ref('overview')
const loadedTabs = ref({
  overview: false,
  moves: false,
  geokrety: false,
  countries: false,
  awards: false,
  chains: false,
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
const validTabs = ['overview', 'moves', 'geokrety', 'countries', 'awards', 'chains', 'related-users']
const UserChainsTab = defineAsyncComponent(() => import('../components/chains/UserChainsTable.vue'))

const joinedYear = computed(() => {
  if (!user.value?.joined_at) return null
  return user.value.joined_at.slice(0, 4)
})

const userMedal = computed(() => {
  const rank = Number(user.value?.rank_all_time || 0)
  if (rank === 1) return '🥇'
  if (rank === 2) return '🥈'
  if (rank === 3) return '🥉'
  return ''
})

const displayTotalPoints = computed(() => {
  const points = Number(user.value?.total_points || 0)
  if (points > 0) return points
  return Number(user.value?.pts_base || 0)
    + Number(user.value?.pts_relay || 0)
    + Number(user.value?.pts_rescuer || 0)
    + Number(user.value?.pts_chain || 0)
    + Number(user.value?.pts_country || 0)
    + Number(user.value?.pts_diversity || 0)
    + Number(user.value?.pts_handover || 0)
    + Number(user.value?.pts_reach || 0)
})

const avgPointsPerMove = computed(() => {
  const movesCount = Number(user.value?.total_moves || 0)
  if (!movesCount) return 0
  return displayTotalPoints.value / movesCount
})

const tabCounts = computed(() => ({
  overview: loadedTabs.value.overview ? String(breakdown.value.length || 0) : (user.value ? '1' : '…'),
  moves: loadedTabs.value.moves
    ? String(moveMeta.value.total ?? 0)
    : (user.value?.total_moves ? user.value.total_moves.toLocaleString() : '…'),
  countries: String(user.value?.countries_count ?? 0),
  geokrety: loadedTabs.value.geokrety
    ? String(gkMeta.value.total ?? 0)
    : (user.value?.distinct_gks ? user.value.distinct_gks.toLocaleString() : '…'),
  awards: loadedTabs.value.awards ? String(awardsMeta.value.total ?? 0) : '…',
  chains: loadedTabs.value.chains ? String(chainsMeta.value.total ?? 0) : '…',
  'related-users': loadedTabs.value['related-users'] ? '0' : '…',
}))

function parseHashState(hashValue) {
  const raw = (hashValue || '').replace(/^#/, '')
  if (!raw) return { tab: 'overview', params: new URLSearchParams() }

  const [tabPart, queryPart = ''] = raw.split('?')
  const tab = validTabs.includes(tabPart) ? tabPart : 'overview'
  const params = new URLSearchParams(queryPart)
  return { tab, params }
}

function buildTabHash(tab) {
  if (tab !== 'awards') return `#${tab}`

  const params = new URLSearchParams()
  if (awardsLabelFilter.value) params.set('label', awardsLabelFilter.value)
  if (awardsSortCol.value !== 'date') params.set('sort', awardsSortCol.value)
  if (awardsSortOrder.value !== 'desc') params.set('order', awardsSortOrder.value)
  if (awardsPage.value > 1) params.set('page', String(awardsPage.value))

  const queryString = params.toString()
  return queryString ? `#awards?${queryString}` : '#awards'
}

function applyHashState(hashValue) {
  const { tab, params } = parseHashState(hashValue)

  if (tab === 'awards') {
    awardsLabelFilter.value = params.get('label') || ''
    awardsSortCol.value = params.get('sort') || 'date'
    awardsSortOrder.value = params.get('order') === 'asc' ? 'asc' : 'desc'
    awardsPage.value = Number(params.get('page') || 1)
  }

  if (activeTab.value !== tab) {
    activeTab.value = tab
    return
  }

  loadActiveTabData(tab)
}

async function loadUser() {
  loading.value = true
  error.value = null
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
  moves.value = items
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
  const params = {
    page: gkPage.value,
    per_page: 25,
    sort: gkSortCol.value,
    order: gkSortOrder.value,
    awarding_only: gkAwardingOnly.value,
    multiplier_gt_one: gkMultiplierOnly.value,
  }
  if (selectedGkTypes.value.length) {
    params.gk_types = selectedGkTypes.value.join(',')
  }

  const { items, meta } = await fetchList(`/users/${userId.value}/geokrety`, params)
  geokrety.value = items
  gkMeta.value = meta
  loadedTabs.value.geokrety = true
}

async function loadAwards() {
  const params = {
    page: awardsPage.value,
    per_page: 25,
    sort: awardsSortCol.value,
    order: awardsSortOrder.value,
  }
  if (awardsLabelFilter.value) {
    params.label = awardsLabelFilter.value
  }

  const { items, meta } = await fetchList(`/users/${userId.value}/points/awards`, params)
  awards.value = items
  awardsMeta.value = meta
  loadedTabs.value.awards = true

}

async function loadActiveTabData(tab) {
  try {
    if (tab === 'overview') await loadOverview()
    if (tab === 'moves') await loadMoves()
    if (tab === 'geokrety') await loadGeokrety()
    if (tab === 'countries') await loadCountries()
    if (tab === 'awards') await loadAwards()
    if (tab === 'chains') loadedTabs.value.chains = true
    if (tab === 'related-users') loadedTabs.value['related-users'] = true
  } catch (e) {
    error.value = e.message
  }
}

function toggleSort(currentColRef, currentOrderRef, col, ascDefaults = []) {
  if (currentColRef.value === col) {
    currentOrderRef.value = currentOrderRef.value === 'asc' ? 'desc' : 'asc'
    return
  }
  currentColRef.value = col
  currentOrderRef.value = ascDefaults.includes(col) ? 'asc' : 'desc'
}

function sortIcon(activeCol, activeOrder, col) {
  if (activeCol !== col) return 'bi-sort-down'
  return activeOrder === 'asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt'
}

function openAwards(label = '') {
  awardsPage.value = 1
  awardsLabelFilter.value = label
  activeTab.value = 'awards'
}

function setAwardLabel(label) {
  awardsLabelFilter.value = label
  awardsPage.value = 1
}

onMounted(async () => {
  applyHashState(route.hash || window.location.hash)
  await loadUser()
  await loadActiveTabData(activeTab.value)
})

watch(movePage, () => {
  if (activeTab.value === 'moves') loadMoves()
})
watch(gkPage, () => {
  if (activeTab.value === 'geokrety') loadGeokrety()
})
watch(awardsPage, () => {
  if (activeTab.value === 'awards') loadAwards()
})

watch([moveSortCol, moveSortOrder, moveAwardingOnly, selectedMoveTypes], () => {
  movePage.value = 1
  if (activeTab.value === 'moves') loadMoves()
}, { deep: true })

watch([gkSortCol, gkSortOrder, gkAwardingOnly, gkMultiplierOnly, selectedGkTypes], () => {
  gkPage.value = 1
  if (activeTab.value === 'geokrety') loadGeokrety()
}, { deep: true })

watch([awardsSortCol, awardsSortOrder, awardsLabelFilter], () => {
  awardsPage.value = 1
  if (activeTab.value === 'awards') loadAwards()
})

watch(() => route.hash, (hash) => {
  applyHashState(hash)
})

watch(activeTab, (tab) => {
  const hash = buildTabHash(tab)
  if (route.hash !== hash) {
    router.push({ path: `/users/${userId.value}`, hash })
  }
  loadActiveTabData(tab)
})

watch(() => route.params.id, async (id) => {
  userId.value = id
  user.value = null
  activeTab.value = 'overview'

  timeline.value = []
  countries.value = []
  moves.value = []
  geokrety.value = []
  breakdown.value = []
  awards.value = []

  moveMeta.value = {}
  gkMeta.value = {}
  awardsMeta.value = {}

  movePage.value = 1
  gkPage.value = 1
  awardsPage.value = 1

  moveSortCol.value = 'date'
  moveSortOrder.value = 'desc'
  gkSortCol.value = 'last_interaction'
  gkSortOrder.value = 'desc'
  awardsSortCol.value = 'date'
  awardsSortOrder.value = 'desc'

  moveAwardingOnly.value = false
  selectedMoveTypes.value = []
  gkAwardingOnly.value = false
  gkMultiplierOnly.value = false
  selectedGkTypes.value = []
  awardsLabelFilter.value = ''
  chainsMeta.value = {}

  loadedTabs.value = {
    overview: false,
    moves: false,
    geokrety: false,
    countries: false,
    awards: false,
    chains: false,
    'related-users': false,
  }

  applyHashState(route.hash)
  await loadUser()
  await loadActiveTabData(activeTab.value)
})
</script>

<template>
  <div v-if="loading && !user" class="text-center py-5">
    <div class="spinner-border"></div>
  </div>
  <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
  <div v-else-if="user">
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">{{ user.username }}</li>
      </ol>
    </nav>

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
            <h3 class="mb-1 text-break d-flex align-items-center gap-2">
              <span v-if="userMedal" class="display-6 lh-1">{{ userMedal }}</span>
              <span>{{ user.username }}</span>
            </h3>
            <p class="text-muted mb-0 small text-nowrap">
              User #{{ user.user_id }}
              <span v-if="joinedYear" class="d-none d-sm-inline"> &mdash; joined {{ joinedYear }}</span>
            </p>
          </div>
          <div class="col-12 col-xl-7 mt-xl-0 mt-3 border-top pt-3 pt-xl-0 border-xl-top-0">
            <div class="row row-cols-2 row-cols-md-3 row-cols-xl-6 g-2 text-center justify-content-center">
              <div class="col">
                <div class="fw-bold text-primary fs-5" title="Total points including bonus rewards"><PointsValue :value="displayTotalPoints" :digits="3" /></div>
                <div class="text-muted small">Points</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5" title="All-time rank on the leaderboard">{{ user.rank_all_time?.toLocaleString() || '—' }}</div>
                <div class="text-muted small">Rank</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5" title="Moves this user logged">{{ user.total_moves?.toLocaleString() }}</div>
                <div class="text-muted small">Moves</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5" title="Distinct GeoKrety this user has interacted with">{{ user.distinct_gks?.toLocaleString() }}</div>
                <div class="text-muted small">GeoKrety</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5" title="Countries visited by this user">{{ user.countries_count?.toLocaleString() }}</div>
                <div class="text-muted small">Countries</div>
              </div>
              <div class="col">
                <div class="fw-bold fs-5" title="Average points per move across the user history"><PointsValue :value="avgPointsPerMove" :digits="3" /></div>
                <div class="text-muted small">Avg/Move</div>
              </div>
            </div>
            <div class="text-center mt-2">
              <RouterLink :to="`/users/${userId}/chains`" class="btn btn-sm btn-outline-secondary">
                <i class="bi bi-link-45deg me-1"></i>Chains
              </RouterLink>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ul class="nav nav-tabs mb-2">
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'overview' }" title="User points timeline and bonus breakdown" @click="activeTab = 'overview'">
          <i class="bi bi-bar-chart-line me-1"></i>Overview <span class="badge bg-secondary ms-1">{{ tabCounts.overview }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'moves' }" title="Chronological list of user moves" @click="activeTab = 'moves'">
          <i class="bi bi-list-ul me-1"></i>Moves <span class="badge bg-secondary ms-1">{{ tabCounts.moves }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'countries' }" title="Countries visited by this user" @click="activeTab = 'countries'">
          <i class="bi bi-globe me-1"></i>Countries <span class="badge bg-secondary ms-1">{{ tabCounts.countries }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'geokrety' }" title="GeoKrety moved by this user" @click="activeTab = 'geokrety'">
          <i class="bi bi-box-seam me-1"></i>GeoKrety <span class="badge bg-secondary ms-1">{{ tabCounts.geokrety }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'awards' }" title="Detailed point award entries and reasons" @click="activeTab = 'awards'">
          <i class="bi bi-award me-1"></i>Points Log <span class="badge bg-secondary ms-1">{{ tabCounts.awards }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'chains' }" title="Chains this user participated in" @click="activeTab = 'chains'">
          <i class="bi bi-link-45deg me-1"></i>Chains <span class="badge bg-secondary ms-1">{{ tabCounts.chains }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'related-users' }" title="Users interacting with the same GeoKrety" @click="activeTab = 'related-users'">
          <i class="bi bi-people me-1"></i>Related Users <span class="badge bg-secondary ms-1">{{ tabCounts['related-users'] }}</span>
        </button>
      </li>
    </ul>

    <div v-if="activeTab === 'overview'">
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

      <div class="card mb-4 shadow-sm" v-if="breakdown.length">
        <div class="card-header"><b>Points by Bonus Type</b></div>
        <div class="card-body">
          <PointsBreakdownChart :data="breakdown" :height="300" />
        </div>
      </div>

      <div class="card shadow-sm">
        <div class="card-header"><b>Points Breakdown</b></div>
        <div class="card-body pb-0">
          <p class="text-muted small mb-2">This panel summarizes where this user earns points, grouped by reward source. Use the eye action or row click to open matching entries in Points Log.</p>
        </div>
        <div class="table-responsive border-0 mb-0">
          <table class="table table-sm table-hover mb-0 align-middle">
            <thead class="table-light">
              <tr>
                <th title="Activity or bonus type that awarded points to this user">Source</th>
                <th class="text-end" title="Total points earned from this source">Points</th>
                <th class="text-end d-none d-sm-table-cell" title="Number of times this reward was earned">Count</th>
                <th class="text-end" style="width: 50px" title="Open this source in Points Log"></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="b in breakdown" :key="b.source" @click="openAwards(b.source)" style="cursor: pointer">
                <td class="fw-medium">{{ b.source }}</td>
                <td class="text-end fw-bold"><PointsValue :value="b.points" /></td>
                <td class="text-end d-none d-sm-table-cell text-muted">{{ b.count?.toLocaleString() }}</td>
                <td class="text-end">
                  <button type="button" class="btn btn-sm btn-outline-secondary py-0 px-1" style="font-size:0.75rem" title="View matching point awards" @click.stop="openAwards(b.source)"><i class="bi bi-eye"></i></button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="card-footer text-end">
          <button type="button" class="btn btn-sm btn-outline-primary shadow-sm" @click="openAwards()">
            <i class="bi bi-list-stars me-1"></i>View all point awards
          </button>
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'moves'">
      <div class="d-flex flex-wrap gap-2 align-items-center mb-2">
        <AwardingOnlyToggle v-model="moveAwardingOnly" />
        <MoveTypeFilterDropdown v-model="selectedMoveTypes" :options="moveTypeOptions" id-prefix="user-move-type" />
      </div>

      <div class="card shadow-sm border-0">
        <div class="table-responsive border-0 mb-0">
          <table class="table table-hover table-sm mb-0 align-middle border">
            <thead class="table-dark">
              <tr>
                <th class="ps-3" style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'date')" :class="moveSortCol==='date' ? 'text-warning' : ''" title="Date the user logged the move">Date <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'date')"></i></th>
                <th style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'gk', ['gk'])" :class="moveSortCol==='gk' ? 'text-warning' : ''" title="GeoKret that was moved">GeoKret <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'gk')"></i></th>
                <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'type')" :class="moveSortCol==='type' ? 'text-warning' : ''" title="Type of activity">Type <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'type')"></i></th>
                <th class="d-none d-sm-table-cell pe-3" style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'country', ['country'])" :class="moveSortCol==='country' ? 'text-warning' : ''" title="Country where activity occurred">Country <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'country')"></i></th>
                <th class="text-end" style="cursor:pointer" @click="toggleSort(moveSortCol, moveSortOrder, 'points')" :class="moveSortCol==='points' ? 'text-warning' : ''" title="Total points earned for this move">Points <i class="bi" :class="sortIcon(moveSortCol, moveSortOrder, 'points')"></i></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in moves" :key="m.move_id" @click="$router.push(`/geokrety/${m.gk_id}`)" style="cursor: pointer">
                <td class="small text-muted ps-3">{{ m.moved_on?.slice(0, 10) }}</td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <img v-if="gkAvatarUrl(m.gk_avatar)" :src="gkAvatarUrl(m.gk_avatar)" :alt="`${m.gk_name || idToGkId(m.gk_id)} avatar`" class="gk-thumb" />
                    <div class="fw-bold text-truncate" style="max-width: 150px">{{ m.gk_name || idToGkId(m.gk_id) }}</div>
                  </div>
                </td>
                <td class="d-none d-md-table-cell"><span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`">{{ m.type_name }}</span></td>
                <td class="d-none d-sm-table-cell pe-3"><span v-if="m.country" :title="`Country: ${m.country}`" class="text-nowrap small text-muted">{{ getCountryFlag(m.country) }} {{ m.country.toUpperCase() }}</span></td>
                <td class="text-end fw-bold"><PointsValue :value="m.points" /></td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="moveMeta.total" :meta="moveMeta" v-model:page="movePage" class="mt-3" />
    </div>

    <div v-if="activeTab === 'geokrety'">
      <div class="d-flex flex-wrap gap-2 align-items-center mb-2">
        <AwardingOnlyToggle v-model="gkAwardingOnly" />
        <button type="button" class="btn btn-sm" :class="gkMultiplierOnly ? 'btn-primary' : 'btn-outline-secondary'" title="Show only GeoKrety with multiplier above 1" @click="gkMultiplierOnly = !gkMultiplierOnly">
          <i class="bi bi-graph-up me-1"></i>Only multiplier >1
        </button>
        <GeokretTypeFilterDropdown v-model="selectedGkTypes" id-prefix="user-gk-type" />
      </div>

      <div class="card shadow-sm border-0">
        <div class="table-responsive border-0 mb-0">
          <table class="table table-hover table-sm mb-0 align-middle border">
            <thead class="table-dark">
              <tr>
                <th class="ps-3" style="cursor:pointer" @click="toggleSort(gkSortCol, gkSortOrder, 'gk', ['gk'])" :class="gkSortCol==='gk' ? 'text-warning' : ''" title="GeoKret name">GeoKret <i class="bi" :class="sortIcon(gkSortCol, gkSortOrder, 'gk')"></i></th>
                <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort(gkSortCol, gkSortOrder, 'type', ['type'])" :class="gkSortCol==='type' ? 'text-warning' : ''" title="GeoKret type">Type <i class="bi" :class="sortIcon(gkSortCol, gkSortOrder, 'type')"></i></th>
                <th class="d-none d-sm-table-cell" style="cursor:pointer" @click="toggleSort(gkSortCol, gkSortOrder, 'last_interaction')" :class="gkSortCol==='last_interaction' ? 'text-warning' : ''" title="Latest interaction date by this user">Last interaction <i class="bi" :class="sortIcon(gkSortCol, gkSortOrder, 'last_interaction')"></i></th>
                <th class="text-end" style="cursor:pointer" @click="toggleSort(gkSortCol, gkSortOrder, 'points')" :class="gkSortCol==='points' ? 'text-warning' : ''" title="Total points generated by this GeoKret">Points Generated <i class="bi" :class="sortIcon(gkSortCol, gkSortOrder, 'points')"></i></th>
                <th class="text-end pe-3" style="cursor:pointer" @click="toggleSort(gkSortCol, gkSortOrder, 'multiplier')" :class="gkSortCol==='multiplier' ? 'text-warning' : ''" title="Current GeoKret multiplier">Multiplier <i class="bi" :class="sortIcon(gkSortCol, gkSortOrder, 'multiplier')"></i></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="g in geokrety" :key="g.gk_id" @click="$router.push(`/geokrety/${g.gk_id}`)" style="cursor: pointer">
                <td class="ps-3">
                  <div class="d-flex align-items-center gap-2">
                    <img v-if="gkAvatarUrl(g.avatar)" :src="gkAvatarUrl(g.avatar)" :alt="`${g.gk_name || idToGkId(g.gk_id)} avatar`" class="gk-thumb" />
                    <div>
                      <RouterLink :to="`/geokrety/${g.gk_id}`" class="fw-bold text-decoration-none" @click.stop>{{ g.gk_name || idToGkId(g.gk_id) }}</RouterLink>
                      <div class="small text-muted">#{{ g.gk_id }}</div>
                    </div>
                  </div>
                </td>
                <td class="d-none d-md-table-cell"><GkTypeBadge :gk-type="g.gk_type" /></td>
                <td class="d-none d-sm-table-cell small text-muted">{{ g.last_interaction ? String(g.last_interaction).slice(0, 10) : '—' }}</td>
                <td class="text-end fw-semibold"><PointsValue :value="g.total_points_generated" /></td>
                <td class="text-end pe-3">{{ Number(g.current_multiplier || 1).toFixed(2) }}×</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="gkMeta.total" :meta="gkMeta" v-model:page="gkPage" class="mt-3" />
    </div>

    <div v-if="activeTab === 'countries'">
      <div class="card shadow-sm mb-2">
        <div class="card-header"><b>Countries visited</b></div>
        <div class="card-body p-2">
          <WorldMap v-if="countries.length" :countries="countries" :height="380" />
          <p v-else class="text-muted text-center py-3">No countries data.</p>
        </div>
      </div>
      <div class="row row-cols-2 row-cols-md-4 row-cols-lg-6 g-2">
        <div v-for="c in countries" :key="c.country" class="col">
          <div class="card text-center p-2 shadow-sm h-100">
            <div class="fw-semibold"><span class="fs-3">{{ getCountryFlag(c.country) }}</span><br/>{{ c.country.toUpperCase() }}</div>
            <div class="text-muted small">{{ (c.move_count || c.moves || 0).toLocaleString() }} moves</div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'awards'">
      <div class="d-flex flex-wrap gap-1 mb-3" v-if="availableAwardLabels.length">
        <button type="button" class="btn btn-sm" :class="!awardsLabelFilter ? 'btn-primary' : 'btn-outline-secondary'" @click="setAwardLabel('')">All</button>
        <button v-for="lbl in availableAwardLabels" :key="lbl" type="button" class="btn btn-sm" :class="awardsLabelFilter === lbl ? 'btn-primary' : 'btn-outline-secondary'" @click="setAwardLabel(lbl)">{{ lbl }}</button>
      </div>

      <div class="card shadow-sm">
        <div class="table-responsive border-0 mb-0">
          <table class="table table-sm table-hover mb-0 align-middle">
            <thead class="table-dark">
              <tr>
                <th style="cursor:pointer" @click="toggleSort(awardsSortCol, awardsSortOrder, 'date')" :class="awardsSortCol==='date' ? 'text-warning' : ''" title="Award date">Date <i class="bi" :class="sortIcon(awardsSortCol, awardsSortOrder, 'date')"></i></th>
                <th style="cursor:pointer" @click="toggleSort(awardsSortCol, awardsSortOrder, 'label', ['label'])" :class="awardsSortCol==='label' ? 'text-warning' : ''" title="Award label">Label <i class="bi" :class="sortIcon(awardsSortCol, awardsSortOrder, 'label')"></i></th>
                <th title="Award details">Reason / Details</th>
                <th title="Related GeoKret">GeoKret</th>
                <th class="text-end" style="cursor:pointer" @click="toggleSort(awardsSortCol, awardsSortOrder, 'points')" :class="awardsSortCol==='points' ? 'text-warning' : ''" title="Awarded points">Points <i class="bi" :class="sortIcon(awardsSortCol, awardsSortOrder, 'points')"></i></th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!awards.length">
                <td colspan="5" class="text-center text-muted py-4">No awards found.</td>
              </tr>
              <tr v-for="a in awards" :key="a.id">
                <td class="small text-muted text-nowrap">{{ a.awarded_at?.slice(0, 10) }}</td>
                <td><span class="badge bg-secondary">{{ a.label || '—' }}</span></td>
                <td class="small">{{ a.reason || '—' }}</td>
                <td>
                  <RouterLink v-if="a.gk_id" :to="`/geokrety/${a.gk_id}`" class="small">{{ idToGkId(a.gk_id) }}</RouterLink>
                  <span v-else class="text-muted">—</span>
                </td>
                <td class="text-end fw-semibold"><PointsValue :value="a.points" /></td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="awardsMeta.total" :meta="awardsMeta" v-model:page="awardsPage" class="mt-3" />
    </div>

    <div v-if="activeTab === 'chains'">
      <Suspense>
        <template #default>
          <UserChainsTab :user-id="userId" @meta-updated="(meta) => (chainsMeta.value = meta)" />
        </template>
        <template #fallback>
          <div class="text-center py-4">
            <div class="spinner-border spinner-border-sm me-2"></div>Loading chains…
          </div>
        </template>
      </Suspense>
    </div>

    <div v-if="activeTab === 'related-users'">
      <RelatedUsersTab :endpoint="`/users/${userId}/related-users`" title="Users who moved same GeoKrety" />
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
