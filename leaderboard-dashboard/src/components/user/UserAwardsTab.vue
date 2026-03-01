<script setup>
import { idToGkId } from '../../composables/useGkId.js'
import Pagination from '../Pagination.vue'
import PointsValue from '../PointsValue.vue'

const props = defineProps({
  awards: { type: Array, default: () => [] },
  meta: { type: Object, default: () => ({}) },
  loading: { type: Boolean, default: false },
  sortCol: { type: String, default: 'date' },
  sortOrder: { type: String, default: 'desc' },
  labelFilter: { type: String, default: '' },
  availableLabels: { type: Array, default: () => [] }
})

const emit = defineEmits(['update:page', 'toggle-sort', 'set-label'])

function toggleSort(col, ascDefaults = []) {
  emit('toggle-sort', col, ascDefaults)
}

function sortIcon(col) {
  if (props.sortCol !== col) return 'bi-sort-down'
  return props.sortOrder === 'asc' ? 'bi-sort-up-alt' : 'bi-sort-down-alt'
}
</script>

<template>
  <div>
    <div class="d-flex flex-wrap gap-1 mb-3" v-if="availableLabels.length">
      <button type="button" class="btn btn-sm" :class="!labelFilter ? 'btn-primary' : 'btn-outline-secondary'" @click="emit('set-label', '')">All</button>
      <button v-for="lbl in availableLabels" :key="lbl" type="button" class="btn btn-sm" :class="labelFilter === lbl ? 'btn-primary' : 'btn-outline-secondary'" @click="emit('set-label', lbl)">{{ lbl }}</button>
    </div>

    <div class="card shadow-sm">
      <div class="table-responsive border-0 mb-0">
        <table class="table table-sm table-hover mb-0 align-middle">
          <thead class="table-dark">
            <tr>
              <th style="cursor:pointer" @click="toggleSort('date')" :class="sortCol==='date' ? 'text-warning' : ''" title="Award date">Date <i class="bi" :class="sortIcon('date')"></i></th>
              <th style="cursor:pointer" @click="toggleSort('label', ['label'])" :class="sortCol==='label' ? 'text-warning' : ''" title="Award label">Label <i class="bi" :class="sortIcon('label')"></i></th>
              <th title="Award details">Reason / Details</th>
              <th title="Related GeoKret">GeoKret</th>
              <th class="text-end" style="cursor:pointer" @click="toggleSort('points')" :class="sortCol==='points' ? 'text-warning' : ''" title="Awarded points">Points <i class="bi" :class="sortIcon('points')"></i></th>
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
    <Pagination v-if="meta.total" :meta="meta" :page="meta.page" @update:page="p => emit('update:page', p)" class="mt-3" />
  </div>
</template>
