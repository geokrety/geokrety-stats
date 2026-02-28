<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import { idToGkId } from '../composables/useGkId.js'
import { getMoveTypeBadgeClass } from '../composables/useMoveTypeColors.js'
import { getCountryFlag } from '../composables/useCountryFlags.js'
import { waypointExternalUrl, displayWaypoint, waypointTooltip, waypointMapUrl } from '../composables/useWaypoint.js'
import GkTypeBadge from '../components/GkTypeBadge.vue'
import LineChart from '../components/LineChart.vue'
import WorldMap from '../components/WorldMap.vue'
import Pagination from '../components/Pagination.vue'
import RelatedUsersTab from '../components/RelatedUsersTab.vue'
import PointsBreakdownChart from '../components/PointsBreakdownChart.vue'

const route   = useRoute()
const gkId    = ref(route.params.id)
const gk      = ref(null)
const timeline = ref([])
const countries = ref([])
const moves    = ref([])
const movePage = ref(1)
const moveMeta = ref({})
const pointsLog = ref([])
const pointsLogPage = ref(1)
const pointsLogMeta = ref({})
const loading  = ref(false)
const error    = ref(null)
const activeTab = ref('overview')

const today = new Date().toISOString().slice(0, 10)

const chartStartDate = computed(() => {
  if (gk.value?.first_move_at) return gk.value.first_move_at.slice(0, 10)
  if (gk.value?.born_at)       return gk.value.born_at.slice(0, 10)
  if (gk.value?.created_at)    return gk.value.created_at.slice(0, 10)
  return null
})

async function load() {
  loading.value = true
  error.value   = null
  try {
    gk.value        = await fetchOne(`/geokrety/${gkId.value}`)
    const tl        = await fetchList(`/geokrety/${gkId.value}/points/timeline`, { per_page: 3650 })
    timeline.value  = tl.items
    const co        = await fetchList(`/geokrety/${gkId.value}/countries`, { per_page: 300 })
    countries.value = co.items
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadMoves() {
  const { items, meta } = await fetchList(`/geokrety/${gkId.value}/moves`, {
    page: movePage.value, per_page: 25,
  })
  moves.value    = items
  moveMeta.value = meta
}

async function loadPointsLog() {
  const { items, meta } = await fetchList(`/geokrety/${gkId.value}/points/log`, {
    page: pointsLogPage.value, per_page: 25,
  })
  pointsLog.value    = items
  pointsLogMeta.value = meta
}

onMounted(() => {
  // Read tab from URL hash
  const hash = window.location.hash.slice(1)
  if (hash && ['overview', 'moves', 'countries', 'related-users', 'points'].includes(hash)) {
    activeTab.value = hash
  }
  load()
  loadMoves()
  loadPointsLog()
})
watch(movePage, loadMoves)
watch(pointsLogPage, loadPointsLog)
watch(() => route.params.id, (id) => { gkId.value = id; load(); loadMoves(); loadPointsLog() })
watch(activeTab, (tab) => {
  // Update URL hash when tab changes
  window.location.hash = tab
})
</script>

<template>
  <div v-if="loading && !gk" class="text-center py-5">
    <div class="spinner-border"></div>
  </div>
  <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
  <div v-else-if="gk">
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-3">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Leaderboard</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">{{ gk.gk_name }}</li>
      </ol>
    </nav>

    <!-- GK Header -->
    <div class="card mb-4 shadow-sm">
      <div class="card-body d-flex align-items-center gap-4 flex-wrap">
        <div class="fs-1">🐢</div>
        <div class="flex-grow-1">
          <div class="d-flex align-items-center gap-2 flex-wrap">
            <h3 class="mb-0">{{ gk.gk_name }}</h3>
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
            &ensp;|&ensp;
            <span v-if="gk.in_cache" class="fw-semibold">
              <span class="badge bg-success">🏦 In Cache</span>
              <span v-if="gk.cache_country" class="ms-1" :title="`Cache location: ${gk.cache_country.toUpperCase()}`">
                {{ getCountryFlag(gk.cache_country) }}
              </span>
            </span>
            <span v-else>
              Holder:
              <RouterLink v-if="gk.holder_id" :to="`/users/${gk.holder_id}`">{{ gk.holder_username }}</RouterLink>
              <span v-else>—</span>
              <span v-if="gk.holder_home_country" class="ms-1" :title="`Home country: ${gk.holder_home_country.toUpperCase()}`">
                {{ getCountryFlag(gk.holder_home_country) }}
              </span>
            </span>
          </p>
        </div>
        <div class="row g-3 text-center">
          <div class="col">
            <div class="fw-bold text-success fs-4">{{ gk.total_points_generated?.toLocaleString() }}</div>
            <div class="text-muted small" title="Total gamification points generated by moves with this GeoKret">Points Generated</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.total_moves?.toLocaleString() }}</div>
            <div class="text-muted small" title="Total number of recorded moves (drops, grabs, dips, seen, comments)">Total Moves</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.distance_km?.toLocaleString() }} km</div>
            <div class="text-muted small" title="Total distance traveled in kilometers">Distance</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.countries_count?.toLocaleString() }}</div>
            <div class="text-muted small" title="Number of distinct countries this GeoKret has visited">Countries</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.distinct_caches?.toLocaleString() }}</div>
            <div class="text-muted small" title="Number of unique places (waypoints) visited where the position was recorded">Places Visited</div>
          </div>
          <div class="col">
            <div class="fw-bold fs-5">{{ gk.current_multiplier?.toFixed(2) }}×</div>
            <div class="text-muted small" title="Points multiplier applied to moves with this GeoKret">Multiplier</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <ul class="nav nav-tabs mb-3">
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'overview' }" @click="activeTab = 'overview'">
          <i class="bi bi-bar-chart-line me-1"></i>Overview
        </button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'moves' }" @click="activeTab = 'moves'">
          <i class="bi bi-list-ul me-1"></i>Moves
          <span v-if="moveMeta.total" class="badge bg-secondary ms-1">{{ moveMeta.total.toLocaleString() }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'countries' }" @click="activeTab = 'countries'">
          <i class="bi bi-globe me-1"></i>Countries
          <span v-if="countries.length" class="badge bg-secondary ms-1">{{ countries.length }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'related-users' }" @click="activeTab = 'related-users'">
          <i class="bi bi-people me-1"></i>Movers
          <span v-if="gk?.distinct_users" class="badge bg-secondary ms-1">{{ gk.distinct_users.toLocaleString() }}</span>
        </button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'points' }" @click="activeTab = 'points'">
          <i class="bi bi-coin me-1"></i>Points Log
          <span v-if="pointsLogMeta.total" class="badge bg-success ms-1">{{ pointsLogMeta.total.toLocaleString() }}</span>
        </button>
      </li>
    </ul>

    <!-- Overview -->
    <div v-if="activeTab === 'overview'">
      <div class="card mb-4 shadow-sm">
        <div class="card-header d-flex justify-content-between align-items-center">
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
      <!-- GK stats mini summary -->
      <div class="row g-3">
        <div class="col-md-4">
          <div class="card shadow-sm h-100">
            <div class="card-body">
              <h6 class="card-title text-muted">Move breakdown</h6>
              <ul class="list-unstyled mb-0 small">
                <li><span class="fw-semibold">{{ gk.total_drops?.toLocaleString() }}</span> drops</li>
                <li><span class="fw-semibold">{{ gk.total_grabs?.toLocaleString() }}</span> grabs</li>
                <li><span class="fw-semibold">{{ gk.total_seen?.toLocaleString() }}</span> seen</li>
                <li><span class="fw-semibold">{{ gk.total_dips?.toLocaleString() }}</span> dips</li>
              </ul>
            </div>
          </div>
        </div>
        <div class="col-md-4">
          <div class="card shadow-sm h-100">
            <div class="card-body">
              <h6 class="card-title text-muted">Reach</h6>
              <ul class="list-unstyled mb-0 small">
                <li><span class="fw-semibold">{{ gk.distinct_users?.toLocaleString() }}</span> distinct users</li>
                <li><span class="fw-semibold">{{ gk.distinct_caches?.toLocaleString() }}</span> distinct waypoints</li>
                <li><span class="fw-semibold">{{ gk.users_awarded?.toLocaleString() }}</span> users awarded</li>
              </ul>
            </div>
          </div>
        </div>
        <div class="col-md-4">
          <div class="card shadow-sm h-100">
            <div class="card-body">
              <h6 class="card-title text-muted">Dates</h6>
              <ul class="list-unstyled mb-0 small">
                <li v-if="gk.born_at">Born: {{ gk.born_at?.slice(0, 10) }}</li>
                <li v-if="gk.first_move_at">First move: {{ gk.first_move_at?.slice(0, 10) }}</li>
                <li v-if="gk.last_move_at">Last move: {{ gk.last_move_at?.slice(0, 10) }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Moves -->
    <div v-if="activeTab === 'moves'">
      <div class="card shadow-sm border-0">
        <div class="table-responsive">
          <table class="table table-hover table-sm mb-0 align-middle border">
            <thead class="table-dark">
              <tr>
                <th class="ps-3">Date</th>
                <th>Author</th>
                <th class="d-none d-md-table-cell">Type</th>
                <th class="text-end">Points</th>
                <th class="d-none d-sm-table-cell">Waypoint</th>
                <th class="d-none d-lg-table-cell pe-3">Country</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in moves" :key="m.move_id" @click="$router.push(`/users/${m.author_id}`)" style="cursor: pointer">
                <td class="small text-muted ps-3">{{ m.moved_on?.slice(0, 10) }}</td>
                <td>
                  <div class="fw-bold text-truncate" style="max-width: 140px">{{ m.author_username }}</div>
                  <div class="d-md-none small">
                    <span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`" style="font-size:0.7rem">{{ m.type_name }}</span>
                    <span v-if="m.waypoint" class="text-muted ms-1 small font-monospace">@{{ displayWaypoint(m.waypoint) }}</span>
                  </div>
                </td>
                <td class="d-none d-md-table-cell">
                  <span :class="`badge ${getMoveTypeBadgeClass(m.type_name)}`">{{ m.type_name }}</span>
                </td>
                <td class="text-end fw-bold text-success">{{ m.points !== null && m.points !== undefined ? m.points.toLocaleString() : '—' }}</td>
                <td class="d-none d-sm-table-cell">
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
                <td class="d-none d-lg-table-cell pe-3">
                  <span v-if="m.country" :title="`Country: ${m.country}`" class="text-nowrap small text-muted">
                    {{ getCountryFlag(m.country) }} {{ m.country.toUpperCase() }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <Pagination v-if="moveMeta.total" :meta="moveMeta" v-model:page="movePage" class="mt-3" />
    </div>

    <!-- Countries -->
    <div v-if="activeTab === 'countries'">
      <div class="card shadow-sm mb-3">
        <div class="card-header"><b>Countries visited</b></div>
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

    <!-- Related Users tab -->
    <div v-if="activeTab === 'related-users'">
      <RelatedUsersTab
        :endpoint="`/geokrety/${gkId}/related-users`"
        title="Users who moved this GeoKret"
      />
    </div>

    <!-- Points Log tab -->
    <div v-if="activeTab === 'points'">
      <div v-if="pointsLog.length === 0" class="text-center text-muted py-5">
        <i class="bi bi-inbox fs-1 d-block mb-2"></i>
        No points recorded for this GeoKret yet.
      </div>
      <div v-else class="card shadow-sm border-0">
        <div class="table-responsive">
          <table class="table table-hover table-sm mb-0 align-middle border">
            <thead class="table-dark">
              <tr>
                <th class="ps-3">Date</th>
                <th>User</th>
                <th class="d-none d-md-table-cell">Reward</th>
                <th class="text-end pe-3">Points</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in pointsLog" :key="p.id" @click="$router.push(`/users/${p.user_id}`)" style="cursor: pointer">
                <td class="small text-muted text-nowrap ps-3">{{ p.awarded_at?.slice(0, 10) }}</td>
                <td>
                  <div class="fw-bold">{{ p.username || p.user_id }}</div>
                  <div class="d-md-none small mt-1">
                    <span class="badge bg-light text-dark border overflow-hidden text-truncate" style="max-width: 150px">
                      {{ (p.label || p.module_source || '—').replace(/_/g, ' ') }}
                    </span>
                    <span v-if="p.is_owner_reward" class="badge bg-warning text-dark ms-1">Owner</span>
                  </div>
                </td>
                <td class="d-none d-md-table-cell">
                  <span class="badge bg-light text-dark border" :title="p.reason || ''">
                    {{ (p.label || p.module_source || '—').replace(/_/g, ' ') }}
                  </span>
                  <span v-if="p.is_owner_reward" class="badge bg-warning text-dark ms-1" title="Owner reward">Owner</span>
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
