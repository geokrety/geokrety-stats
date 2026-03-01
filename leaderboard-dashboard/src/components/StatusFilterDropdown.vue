<script setup>
import { computed, ref, watch } from 'vue'

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

const allValues = computed(() => STATUS_OPTIONS.map((opt) => opt.value))
const initialized = ref(false)

watch(
  () => props.modelValue,
  (value) => {
    if (!initialized.value && value.length === 0 && allValues.value.length) {
      emit('update:modelValue', [...allValues.value])
    }
    initialized.value = true
  },
  { immediate: true }
)

const buttonLabel = computed(() => {
  if (!props.modelValue.length) return 'None selected'
  if (props.modelValue.length === allValues.value.length) return 'All statuses'
  return STATUS_OPTIONS
    .filter((opt) => props.modelValue.includes(opt.value))
    .map((opt) => opt.label)
    .join(', ')
})

const isDefaultSelection = computed(() => props.modelValue.length === allValues.value.length)

function toggleStatus(status) {
  if (props.modelValue.includes(status)) {
    emit('update:modelValue', props.modelValue.filter((item) => item !== status))
    return
  }
  emit('update:modelValue', [...props.modelValue, status])
}

function selectAll() {
  emit('update:modelValue', [...allValues.value])
}

function selectNone() {
  emit('update:modelValue', [])
}
</script>

<template>
  <div class="dropdown">
    <button
      class="btn btn-sm dropdown-toggle"
      :class="isDefaultSelection ? 'btn-outline-secondary' : 'btn-primary'"
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
        <div class="btn-group btn-group-sm" role="group">
          <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="selectAll">All</button>
          <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="selectNone">None</button>
        </div>
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
