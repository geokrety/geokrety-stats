<script setup>
import { ref, onMounted, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { fetchList } from '../composables/useApi.js'
import { idToGkId } from '../composables/useGkId.js'
import GkTypeBadge from '../components/GkTypeBadge.vue'
import Pagination from '../components/Pagination.vue'
import PointsValue from '../components/PointsValue.vue'
import AwardingOnlyToggle from '../components/AwardingOnlyToggle.vue'
import GeokretTypeFilterDropdown from '../components/GeokretTypeFilterDropdown.vue'
import StatusFilterDropdown from '../components/StatusFilterDropdown.vue'
import LoveRangeFilter from '../components/LoveRangeFilter.vue'

const route = useRoute()
const router = useRouter()

const page = ref(Number(route.query.page) || 1)
const perPage = ref(50)
const loading = ref(false)
const error = ref(null)
const geokrety = ref([])
const meta = ref({})

const sortCol = ref(route.query.sort || 'points')
const sortOrder = ref(route.query.order === 'asc' ? 'asc' : 'desc')

const awardingOnly = ref(route.query.awarding_only === 'true')
const multiplierGtOne = ref(route.query.multiplier_gt_one === 'true')
const selectedGkTypes = ref(route.query.gk_types ? String(route.query.gk_types).split(',').map((v) => Number(v)).filter((v) => !Number.isNaN(v)) : [])
const selectedStatuses = ref(route.query.status ? String(route.query.status).split(',').filter(Boolean) : [])
const loveMin = ref(route.query.love_min ? Number(route.query.love_min) : null)
const loveMax = ref(route.query.love_max ? Number(route.query.love_max) : null)

async function loadGeokrety() {
  loading.value = true
  error.value = null
  try {
    const params = {
      page: page.value,
      per_page: perPage.value,
      sort: sortCol.value,
      order: sortOrder.value,
      awarding_only: awardingOnly.value,
      multiplier_gt_one: multiplierGtOne.value,
    }

    if (selectedGkTypes.value.length) params.gk_types = selectedGkTypes.value.join(',')
    if (selectedStatuses.value.length) params.status = selectedStatuses.value.join(',')
    if (loveMin.value !== null) params.love_min = loveMin.value
    if (loveMax.value !== null) params.love_max = loveMax.value

    const { items, meta: m } = await fetchList('/geokrety', params)
    geokrety.value = items
    meta.value = m
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function syncQuery() {
  const query = {
    page: page.value > 1 ? String(page.value) : undefined,
    sort: sortCol.value !== 'points' ? sortCol.value : undefined,
    order: sortOrder.value !== 'desc' ? sortOrder.value : undefined,
    awarding_only: awardingOnly.value ? 'true' : undefined,
    multiplier_gt_one: multiplierGtOne.value ? 'true' : undefined,
    gk_types: selectedGkTypes.value.length ? selectedGkTypes.value.join(',') : undefined,
    status: selectedStatuses.value.length ? selectedStatuses.value.join(',') : undefined,
    love_min: loveMin.value !== null ? String(loveMin.value) : undefined,
    love_max: loveMax.value !== null ? String(loveMax.value) : undefined,
  }
  router.replace({ query })
}

function toggleSort(col) {
  if (sortCol.value === col) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
    return
  }

  sortCol.value = col
  sortOrder.value = ['name', 'type', 'owner', 'status', 'id'].includes(col) ? 'asc' : 'desc'
}

function sortIcon(col) {
  if (sortCol.value !== col) return 'bi-sort-down'
  return sortOrder.value === 'asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt'
}

function tableStatus(gk) {
  if (gk.missing) return 'Missing'
  if (gk.in_cache) return 'In Cache'
  if (gk.holder_username) return `Held by ${gk.holder_username}`
  return 'Unknown'
}

onMounted(loadGeokrety)

watch([page], async () => {
  syncQuery()
  await loadGeokrety()
})

watch([sortCol, sortOrder, awardingOnly, multiplierGtOne, selectedGkTypes, selectedStatuses, loveMin, loveMax], async () => {
  page.value = 1
  syncQuery()
  await loadGeokrety()
}, { deep: true })
</script>

<template>
  <div>
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item active" aria-current="page">GeoKrety</li>
      </ol>
    </nav>

    <div class="card mb-4 shadow-sm">
      <div class="card-body">
        <h2 class="mb-1">🎁 GeoKrety Leaderboard</h2>
        <p class="text-muted mb-0">Browse GeoKrety ranked by points and circulation metrics</p>
      </div>
    </div>

    <div class="d-flex flex-wrap gap-2 align-items-center mb-3">
      <AwardingOnlyToggle v-model="awardingOnly" />
      <button type="button" class="btn btn-sm" :class="multiplierGtOne ? 'btn-primary' : 'btn-outline-secondary'" title="Show only GeoKrety with multiplier above 1" @click="multiplierGtOne = !multiplierGtOne">
        <i class="bi bi-graph-up me-1"></i>Only multiplier >1
      </button>
      <GeokretTypeFilterDropdown v-model="selectedGkTypes" id-prefix="gk-list-type" />
      <StatusFilterDropdown v-model="selectedStatuses" id-prefix="gk-list-status" />
      <LoveRangeFilter v-model:min="loveMin" v-model:max="loveMax" />
      <span class="text-muted ms-auto small">{{ meta.total || 0 }} GeoKrety total</span>
    </div>

    <div v-if="loading && !geokrety.length" class="text-center py-5"><div class="spinner-border"></div></div>
    <div v-else-if="error" class="alert alert-danger">{{ error }}</div>
    <div v-else-if="geokrety.length === 0" class="alert alert-info">No GeoKrety found.</div>

    <template v-else>
      <div class="table-responsive border-0 mb-0">
        <table class="table table-hover table-sm align-middle border">
          <thead class="table-dark">
            <tr>
              <th style="width: 60px" title="Static ranking index in this page">#</th>
              <th style="width: 90px; cursor:pointer" @click="toggleSort('id')" :class="sortCol==='id' ? 'text-warning' : ''" title="Public GeoKret ID">ID <i class="bi" :class="sortIcon('id')"></i></th>
              <th style="cursor:pointer" @click="toggleSort('name')" :class="sortCol==='name' ? 'text-warning' : ''" title="GeoKret name">Name <i class="bi" :class="sortIcon('name')"></i></th>
              <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('type')" :class="sortCol==='type' ? 'text-warning' : ''" title="GeoKret category">Type <i class="bi" :class="sortIcon('type')"></i></th>
              <th class="d-none d-lg-table-cell" style="cursor:pointer" @click="toggleSort('owner')" :class="sortCol==='owner' ? 'text-warning' : ''" title="Owner username">Owner <i class="bi" :class="sortIcon('owner')"></i></th>
              <th class="d-none d-sm-table-cell" style="cursor:pointer" @click="toggleSort('status')" :class="sortCol==='status' ? 'text-warning' : ''" title="Current status">Status <i class="bi" :class="sortIcon('status')"></i></th>
              <th class="text-end d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('moves')" :class="sortCol==='moves' ? 'text-warning' : ''" title="Total recorded moves">Moves <i class="bi" :class="sortIcon('moves')"></i></th>
              <th class="text-end d-none d-lg-table-cell" style="cursor:pointer" @click="toggleSort('users')" :class="sortCol==='users' ? 'text-warning' : ''" title="Distinct users who moved it">Users <i class="bi" :class="sortIcon('users')"></i></th>
              <th class="text-end d-none d-lg-table-cell" style="cursor:pointer" @click="toggleSort('countries')" :class="sortCol==='countries' ? 'text-warning' : ''" title="Distinct countries reached">Countries <i class="bi" :class="sortIcon('countries')"></i></th>
              <th class="text-end" style="cursor:pointer" @click="toggleSort('points')" :class="sortCol==='points' ? 'text-warning' : ''" title="Total points generated">Points <i class="bi" :class="sortIcon('points')"></i></th>
              <th class="text-end d-none d-xl-table-cell" style="cursor:pointer" @click="toggleSort('multiplier')" :class="sortCol==='multiplier' ? 'text-warning' : ''" title="Current multiplier">Multiplier <i class="bi" :class="sortIcon('multiplier')"></i></th>
              <th class="text-end" style="cursor:pointer" @click="toggleSort('loves')" :class="sortCol==='loves' ? 'text-warning' : ''" title="Number of loves">❤️ <i class="bi" :class="sortIcon('loves')"></i></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(gk, idx) in geokrety" :key="gk.gk_id" @click="$router.push(`/geokrety/${gk.gk_id}`)">
              <td class="text-center text-muted fw-semibold small">{{ (page - 1) * perPage + idx + 1 }}</td>
              <td class="small fw-bold">{{ idToGkId(gk.gk_id) }}</td>
              <td>
                <div class="fw-semibold text-truncate" style="max-width: 220px">{{ gk.gk_name || `GK #${gk.gk_id}` }}</div>
                <div class="d-md-none small text-muted mt-1">{{ gk.total_moves?.toLocaleString() }} moves • {{ gk.distinct_users?.toLocaleString() }} users</div>
              </td>
              <td class="d-none d-md-table-cell"><GkTypeBadge :gk-type="gk.gk_type" :type-name="gk.gk_type_name" /></td>
              <td class="d-none d-lg-table-cell"><span class="text-truncate d-inline-block" style="max-width: 120px">{{ gk.owner_username || '—' }}</span></td>
              <td class="d-none d-sm-table-cell"><span class="badge bg-secondary">{{ tableStatus(gk) }}</span></td>
              <td class="text-end fw-bold d-none d-md-table-cell">{{ gk.total_moves?.toLocaleString() }}</td>
              <td class="text-end d-none d-lg-table-cell">{{ gk.distinct_users?.toLocaleString() }}</td>
              <td class="text-end d-none d-lg-table-cell">{{ gk.countries_count?.toLocaleString() }}</td>
              <td class="text-end fw-bold"><PointsValue :value="gk.total_points_generated" /></td>
              <td class="text-end text-muted small d-none d-xl-table-cell">{{ Number(gk.current_multiplier || 1).toFixed(2) }}×</td>
              <td class="text-end text-danger">{{ gk.loves_count?.toLocaleString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <Pagination v-if="meta.total" :meta="meta" v-model:page="page" class="mt-3" />
    </template>
  </div>
</template>

<style scoped>
tbody tr { cursor: pointer; }
tbody tr:hover { background-color: rgba(0, 123, 255, 0.04); }
</style>
