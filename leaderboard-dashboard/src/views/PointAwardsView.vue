<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { fetchOne, fetchList } from '../composables/useApi.js'
import { idToGkId } from '../composables/useGkId.js'
import Pagination from '../components/Pagination.vue'

const route  = useRoute()
const userId = ref(route.params.id)

const user      = ref(null)
const awards    = ref([])
const meta      = ref({})
const page      = ref(Number(route.query.page) || 1)
const perPage   = ref(50)
const sortCol   = ref(route.query.sort || 'date')
const loading   = ref(false)
const error     = ref(null)

// Optional label filter from query string
const labelFilter = ref(route.query.label || '')

async function load() {
  loading.value = true
  error.value   = null
  try {
    const params = { 
      page: page.value, 
      per_page: perPage.value,
      sort: sortCol.value,
    }
    if (labelFilter.value) params.label = labelFilter.value
    const { items, meta: m } = await fetchList(`/users/${userId.value}/points/awards`, params)
    awards.value = items
    meta.value   = m
    // Collect unique labels if not already done
    if (!availableLabels.value.length) {
      const labels = await fetchList(`/users/${userId.value}/points/awards`, { per_page: 200 })
      const seen = new Set()
      for (const a of labels.items || []) { if (a.label) seen.add(a.label) }
      availableLabels.value = [...seen].sort()
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(() => { loadUser(); load() })
watch([page, labelFilter, sortCol], load)
watch(() => route.params.id, id => { userId.value = id; page.value = 1; loadUser(); load() })

function setLabel(l) { labelFilter.value = l; page.value = 1 }

const pointsClass = (pts) => pts > 0 ? 'text-success fw-semibold' : pts < 0 ? 'text-danger fw-semibold' : 'text-muted'
</script>

<template>
  <div>
    <!-- Breadcrumb -->
    <nav aria-label="breadcrumb" class="mb-2">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><RouterLink to="/">Home</RouterLink></li>
        <li class="breadcrumb-item">
          <RouterLink v-if="user" :to="`/users/${userId}`">{{ user.username }}</RouterLink>
          <span v-else>User #{{ userId }}</span>
        </li>
        <li class="breadcrumb-item active" aria-current="page">Point Awards</li>
      </ol>
    </nav>

    <!-- Header -->
    <div class="d-flex align-items-center justify-content-between flex-wrap gap-2 mb-2">
      <h4 class="mb-0">
        <i class="bi bi-list-stars text-warning me-2"></i>
        Point Awards
        <span v-if="user" class="text-muted fs-6 ms-2">— {{ user.username }}</span>
      </h4>
      <RouterLink v-if="user" :to="`/users/${userId}`" class="btn btn-sm btn-outline-secondary">
        <i class="bi bi-arrow-left me-1"></i>Back to user
      </RouterLink>
    </div>

    <!-- Label filter chips -->
    <div v-if="availableLabels.length" class="mb-3 d-flex flex-wrap gap-1">
      <button
        class="btn btn-sm"
        :class="!labelFilter ? 'btn-primary' : 'btn-outline-secondary'"
        @click="setLabel('')"
      >All</button>
      <button
        v-for="lbl in availableLabels" :key="lbl"
        class="btn btn-sm"
        :class="labelFilter === lbl ? 'btn-primary' : 'btn-outline-secondary'"
        @click="setLabel(lbl)"
      >{{ lbl }}</button>
    </div>

    <!-- Error -->
    <div v-if="error" class="alert alert-danger">{{ error }}</div>

    <!-- Table -->
    <div class="card shadow-sm">
      <div class="table-responsive border-0 mb-0">
        <table class="table table-sm table-hover mb-0 align-middle">
          <thead class="table-dark">
            <tr>
              <th style="cursor:pointer" @click="sortCol='date'" :class="sortCol==='date' ? 'text-warning' : ''">
                Date <i class="bi" :class="sortCol==='date' ? 'bi-sort-down-alt' : 'bi-sort-down'"></i>
              </th>
              <th style="cursor:pointer" @click="sortCol='label'" :class="sortCol==='label' ? 'text-warning' : ''">
                Label <i class="bi" :class="sortCol==='label' ? 'bi-sort-down-alt' : 'bi-sort-down'"></i>
              </th>
              <th>Reason / Details</th>
              <th>GeoKret</th>
              <th class="text-end" style="cursor:pointer" @click="sortCol='points'" :class="sortCol==='points' ? 'text-warning' : ''">
                Points <i class="bi" :class="sortCol==='points' ? 'bi-sort-down-alt' : 'bi-sort-down'"></i>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading && !awards.length">
              <td colspan="5" class="text-center py-4">
                <div class="spinner-border spinner-border-sm me-2"></div>Loading…
              </td>
            </tr>
            <tr v-else-if="!awards.length && !loading">
              <td colspan="5" class="text-center text-muted py-4">No awards found.</td>
            </tr>
            <tr v-for="a in awards" :key="a.id">
              <td class="small text-muted text-nowrap">{{ a.awarded_at?.slice(0, 10) }}</td>
              <td>
                <span class="badge bg-secondary">{{ a.label || '—' }}</span>
              </td>
              <td class="small">{{ a.reason || '—' }}</td>
              <td>
                <RouterLink v-if="a.gk_id" :to="`/geokrety/${a.gk_id}`" class="small">
                  {{ idToGkId(a.gk_id) }}
                </RouterLink>
                <span v-else class="text-muted">—</span>
              </td>
              <td class="text-end" :class="pointsClass(a.points)">
                {{ a.points > 0 ? '+' : '' }}{{ a.points?.toFixed(2) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <Pagination v-if="meta.total" :meta="meta" v-model:page="page" class="mt-3" />
  </div>
</template>
