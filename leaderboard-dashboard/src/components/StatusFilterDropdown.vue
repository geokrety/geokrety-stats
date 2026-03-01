<script setup>
import { computed } from 'vue'

const STATUS_OPTIONS = [
  { value: 'cache', label: 'In cache' },
  { value: 'held', label: 'Held' },
  { value: 'missing', label: 'Missing' },
  { value: 'parked', label: 'Parked' },
  { value: 'sealed', label: 'Sealed' },
]

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  idPrefix: { type: String, default: 'status-filter' },
})

const emit = defineEmits(['update:modelValue'])

const buttonLabel = computed(() => {
  if (!props.modelValue.length) return 'All statuses'
  return STATUS_OPTIONS
    .filter((opt) => props.modelValue.includes(opt.value))
    .map((opt) => opt.label)
    .join(', ')
})

function toggleStatus(status) {
  if (props.modelValue.includes(status)) {
    emit('update:modelValue', props.modelValue.filter((item) => item !== status))
    return
  }
  emit('update:modelValue', [...props.modelValue, status])
}

function clearAll() {
  emit('update:modelValue', [])
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
      title="Filter rows by GeoKret status"
    >
      <i class="bi bi-bookmark-check me-1"></i>{{ buttonLabel }}
    </button>
    <div class="dropdown-menu p-2" style="min-width: 220px;">
      <div class="d-flex justify-content-between align-items-center mb-2">
        <strong class="small">Statuses</strong>
        <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="clearAll">All</button>
      </div>
      <div v-for="opt in STATUS_OPTIONS" :key="opt.value" class="form-check">
        <input
          class="form-check-input"
          type="checkbox"
          :id="`${idPrefix}-${opt.value}`"
          :checked="modelValue.includes(opt.value)"
          @change="toggleStatus(opt.value)"
        />
        <label class="form-check-label small" :for="`${idPrefix}-${opt.value}`">{{ opt.label }}</label>
      </div>
    </div>
  </div>
</template>
