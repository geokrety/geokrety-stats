<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  options: { type: Array, required: true },
  idPrefix: { type: String, default: 'move-type' },
})

const emit = defineEmits(['update:modelValue'])

const allValues = computed(() => props.options.map((opt) => opt.value))
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
  if (props.modelValue.length === allValues.value.length) return 'All move types'
  return props.options
    .filter((opt) => props.modelValue.includes(opt.value))
    .map((opt) => opt.label)
    .join(', ')
})

const isDefaultSelection = computed(() => props.modelValue.length === allValues.value.length)

function toggleType(type) {
  if (props.modelValue.includes(type)) {
    emit('update:modelValue', props.modelValue.filter((item) => item !== type))
    return
  }

  emit('update:modelValue', [...props.modelValue, type].sort((a, b) => a - b))
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
      data-bs-auto-close="true"
      aria-expanded="false"
      :title="'Filter rows by move types'"
    >
      <i class="bi bi-funnel me-1"></i>{{ buttonLabel }}
    </button>
    <div class="dropdown-menu p-2" style="min-width: 220px;">
      <div class="d-flex justify-content-between align-items-center mb-2">
        <strong class="small">Move types</strong>
        <div class="btn-group btn-group-sm" role="group">
          <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="selectAll">All</button>
          <span class="mx-1 text-muted small">/</span>
          <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="selectNone">None</button>
        </div>
      </div>
      <div v-for="opt in options" :key="opt.value" class="form-check">
        <input
          class="form-check-input"
          type="checkbox"
          :id="`${idPrefix}-${opt.value}`"
          :checked="modelValue.includes(opt.value)"
          @change="toggleType(opt.value)"
        />
        <label class="form-check-label small" :for="`${idPrefix}-${opt.value}`">{{ opt.label }}</label>
      </div>
    </div>
  </div>
</template>
