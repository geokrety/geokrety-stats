<script setup>
import { getCountryFlag } from '../../composables/useCountryFlags.js'
import WorldMap from '../WorldMap.vue'

const props = defineProps({
  countries: { type: Array, default: () => [] }
})
</script>

<template>
  <div>
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
</template>
