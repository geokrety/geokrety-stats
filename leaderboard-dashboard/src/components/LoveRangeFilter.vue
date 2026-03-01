<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  min: { type: [Number, null], default: null },
  max: { type: [Number, null], default: null },
})

const emit = defineEmits(['update:min', 'update:max'])

const localMin = ref(props.min)
const localMax = ref(props.max)

watch(() => props.min, (value) => { localMin.value = value })
watch(() => props.max, (value) => { localMax.value = value })

function apply() {
  emit('update:min', localMin.value === '' || localMin.value === null ? null : Number(localMin.value))
  emit('update:max', localMax.value === '' || localMax.value === null ? null : Number(localMax.value))
}

function clear() {
  localMin.value = null
  localMax.value = null
  emit('update:min', null)
  emit('update:max', null)
}
</script>

<template>
  <div class="dropdown">
    <button
      class="btn btn-sm btn-outline-secondary dropdown-toggle"
      type="button"
      data-bs-toggle="dropdown"
      data-bs-auto-close="outside"
      aria-expanded="false"
      title="Filter by love count threshold"
    >
      <i class="bi bi-heart me-1"></i>Loves
    </button>
    <div class="dropdown-menu p-2" style="min-width: 220px;">
      <label class="form-label small mb-1">Min</label>
      <input v-model="localMin" type="number" min="0" class="form-control form-control-sm mb-2" />
      <label class="form-label small mb-1">Max</label>
      <input v-model="localMax" type="number" min="0" class="form-control form-control-sm mb-2" />
      <div class="d-flex justify-content-between">
        <button type="button" class="btn btn-sm btn-link text-decoration-none p-0" @click="clear">Clear</button>
        <button type="button" class="btn btn-sm btn-primary" @click="apply">Apply</button>
      </div>
    </div>
  </div>
</template>
