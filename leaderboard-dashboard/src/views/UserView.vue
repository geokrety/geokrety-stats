<script setup>
import { ref, computed, onMounted, watch, defineAsyncComponent } from 'vue'
import { useRoute, RouterLink, useRouter } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import { userAvatarUrl } from '../composables/useAvatarUrl.js'
import RelatedUsersTab from '../components/RelatedUsersTab.vue'
import PointsValue from '../components/PointsValue.vue'
import UserOverviewTab from '../components/user/UserOverviewTab.vue'
import UserMovesTab from '../components/user/UserMovesTab.vue'
import UserGkTab from '../components/user/UserGkTab.vue'
import UserCountriesTab from '../components/user/UserCountriesTab.vue'
import UserAwardsTab from '../components/user/UserAwardsTab.vue'
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

    <!-- Hero Header -->
    <div class="row mb-4 g-3 align-items-stretch">
      <div class="col-12 col-lg-8">
        <div class="card h-100 shadow-sm border-0 bg-dark text-white overflow-hidden hero-card">
          <div class="card-body p-4 position-relative d-flex flex-column justify-content-center">
            <div class="d-flex align-items-center gap-4 flex-wrap flex-md-nowrap">
              <div class="flex-shrink-0 hero-avatar-container">
                <img v-if="userAvatarUrl(user.avatar)" :src="userAvatarUrl(user.avatar)"
                     class="rounded hero-avatar"
                     style="width: 100px; height: 100px; object-fit: cover; border: 3px solid rgba(255,255,255,0.2)"
                     @error="e => e.target.src = '/user-default.png'" />
                <div v-else class="hero-avatar d-flex align-items-center justify-content-center bg-secondary rounded" style="width: 100px; height: 100px; font-size: 3rem;">👤</div>
              </div>
              <div class="flex-grow-1">
                <div class="d-flex align-items-center gap-2 mb-2 flex-wrap">
                  <h2 class="mb-0 fw-bold">{{ userMedal }} {{ user.username }}</h2>
                  <span class="badge bg-light text-dark opacity-75" style="font-size: 0.8rem">User #{{ user.user_id }}</span>
                </div>
                <div class="fs-5 mb-3 opacity-75 d-flex align-items-center gap-2 flex-wrap" v-if="joinedYear">
                  <i class="bi bi-calendar-check-fill text-info small"></i> Joined in {{ joinedYear }}
                </div>
                <div class="d-flex gap-4 flex-wrap opacity-75 small">
                  <div class="d-flex align-items-center gap-2">
                    <i class="bi bi-trophy-fill text-warning"></i> Rank: <b>#{{ user.rank_all_time?.toLocaleString() || '—' }}</b>
                  </div>
                  <div class="d-flex align-items-center gap-2">
                    <i class="bi bi-geo-fill text-success"></i> Moves: <b>{{ user.total_moves?.toLocaleString() }}</b>
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
            <i class="bi bi-person-badge" style="font-size: 5rem; transform: rotate(-15deg); display: block;"></i>
          </div>
          <div class="position-relative">
            <div class="display-5 fw-bold mb-0">
              <PointsValue :value="displayTotalPoints" />
            </div>
            <div class="text-uppercase small opacity-75 ls-1 fw-semibold mb-3">Total Points Earned</div>
            <div class="row g-2 border-top border-white border-opacity-25 pt-3">
              <div class="col-4 px-1">
                <div class="h5 mb-0 fw-bold">{{ user.distinct_gks?.toLocaleString() || 0 }}</div>
                <div class="x-small text-uppercase opacity-75">GeoKrety</div>
              </div>
              <div class="col-4 px-1">
                <div class="h5 mb-0 fw-bold">{{ user.countries_count?.toLocaleString() || 0 }}</div>
                <div class="x-small text-uppercase opacity-75">Countries</div>
              </div>
              <div class="col-4 px-1">
                <div class="h5 mb-0 fw-bold">
                  <PointsValue :value="avgPointsPerMove" :digits="2" />
                </div>
                <div class="x-small text-uppercase opacity-75">Avg/Move</div>
              </div>
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
      <UserOverviewTab
        :user-id="userId"
        :user="user"
        :timeline="timeline"
        :breakdown="breakdown"
        @open-awards="openAwards"
      />
    </div>

    <div v-if="activeTab === 'moves'">
      <UserMovesTab
        :moves="moves"
        :meta="moveMeta"
        :loading="loading"
        :sort-col="moveSortCol"
        :sort-order="moveSortOrder"
        :awarding-only="moveAwardingOnly"
        :selected-types="selectedMoveTypes"
        @update:page="p => movePage = p"
        @toggle-sort="(col, defaults) => toggleSort(moveSortCol, moveSortOrder, col, defaults)"
        @update:awarding-only="v => moveAwardingOnly = v"
        @update:selected-types="v => selectedMoveTypes = v"
      />
    </div>

    <div v-if="activeTab === 'geokrety'">
      <UserGkTab
        :geokrety="geokrety"
        :meta="gkMeta"
        :loading="loading"
        :sort-col="gkSortCol"
        :sort-order="gkSortOrder"
        :awarding-only="gkAwardingOnly"
        :multiplier-only="gkMultiplierOnly"
        :selected-types="selectedGkTypes"
        @update:page="p => gkPage = p"
        @toggle-sort="(col, defaults) => toggleSort(gkSortCol, gkSortOrder, col, defaults)"
        @update:awarding-only="v => gkAwardingOnly = v"
        @update:multiplier-only="v => gkMultiplierOnly = v"
        @update:selected-types="v => selectedGkTypes = v"
      />
    </div>

    <div v-if="activeTab === 'countries'">
      <UserCountriesTab :countries="countries" />
    </div>

    <div v-if="activeTab === 'awards'">
      <UserAwardsTab
        :awards="awards"
        :meta="awardsMeta"
        :loading="loading"
        :sort-col="awardsSortCol"
        :sort-order="awardsSortOrder"
        :label-filter="awardsLabelFilter"
        :available-labels="availableAwardLabels"
        @update:page="p => awardsPage = p"
        @toggle-sort="(col, defaults) => toggleSort(awardsSortCol, awardsSortOrder, col, defaults)"
        @set-label="setAwardLabel"
      />
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
</style>
