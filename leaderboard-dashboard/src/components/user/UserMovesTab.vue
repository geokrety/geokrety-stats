<script setup>
import { idToGkId } from '../../composables/useGkId.js'
import { getMoveTypeBadgeClass } from '../../composables/useMoveTypeColors.js'
import { getCountryFlag } from '../../composables/useCountryFlags.js'
import { gkAvatarUrl } from '../../composables/useAvatarUrl.js'
import Pagination from '../Pagination.vue'
import PointsValue from '../PointsValue.vue'
import AwardingOnlyToggle from '../AwardingOnlyToggle.vue'
import MoveTypeFilterDropdown from '../MoveTypeFilterDropdown.vue'

const props = defineProps({
  moves: { type: Array, default: () => [] },
  meta: { type: Object, default: () => ({}) },
  loading: { type: Boolean, default: false },
  sortCol: { type: String, default: 'date' },
  sortOrder: { type: String, default: 'desc' },
  awardingOnly: { type: Boolean, default: false },
  selectedTypes: { type: Array, default: () => [] }
})

const emit = defineEmits(['update:page', 'toggle-sort', 'update:awarding-only', 'update:selected-types'])

const moveTypeOptions = [
  { value: 0, label: 'Drop' },
  { value: 1, label: 'Grab' },
  { value: 2, label: 'Comment' },
  { value: 3, label: 'Seen' },
  { value: 4, label: 'Archived' },
  { value: 5, label: 'Dip' },
]

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
    <div class="d-flex flex-wrap gap-2 align-items-center mb-2">
      <AwardingOnlyToggle :model-value="awardingOnly" @update:model-value="v => emit('update:awarding-only', v)" />
      <MoveTypeFilterDropdown :model-value="selectedTypes" @update:model-value="v => emit('update:selected-types', v)" :options="moveTypeOptions" id-prefix="user-move-type" />
    </div>

    <div class="card shadow-sm border-0">
      <div class="table-responsive border-0 mb-0">
        <table class="table table-hover table-sm mb-0 align-middle border">
          <thead class="table-dark">
            <tr>
              <th class="ps-3" style="cursor:pointer" @click="toggleSort('date')" :class="sortCol==='date' ? 'text-warning' : ''" title="Date the user logged the move">Date <i class="bi" :class="sortIcon('date')"></i></th>
              <th style="cursor:pointer" @click="toggleSort('gk', ['gk'])" :class="sortCol==='gk' ? 'text-warning' : ''" title="GeoKret that was moved">GeoKret <i class="bi" :class="sortIcon('gk')"></i></th>
              <th class="d-none d-md-table-cell" style="cursor:pointer" @click="toggleSort('type')" :class="sortCol==='type' ? 'text-warning' : ''" title="Type of activity">Type <i class="bi" :class="sortIcon('type')"></i></th>
              <th class="d-none d-sm-table-cell pe-3" style="cursor:pointer" @click="toggleSort('country', ['country'])" :class="sortCol==='country' ? 'text-warning' : ''" title="Country where activity occurred">Country <i class="bi" :class="sortIcon('country')"></i></th>
              <th class="text-end" style="cursor:pointer" @click="toggleSort('points')" :class="sortCol==='points' ? 'text-warning' : ''" title="Total points earned for this move">Points <i class="bi" :class="sortIcon('points')"></i></th>
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
    <Pagination v-if="meta.total" :meta="meta" :page="meta.page" @update:page="p => emit('update:page', p)" class="mt-3" />
  </div>
</template>

<style scoped>
.gk-thumb {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--bs-border-color);
  flex-shrink: 0;
}
</style>
