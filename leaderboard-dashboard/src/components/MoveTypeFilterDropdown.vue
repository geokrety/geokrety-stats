<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  options: { type: Array, required: true },
  idPrefix: { type: String, default: 'move-type' },
})

const emit = defineEmits(['update:modelValue'])

const buttonLabel = computed(() => {
  if (!props.modelValue.length) return 'All move types'
  return props.options
    .filter((opt) => props.modelValue.includes(opt.value))
    .map((opt) => opt.label)
    .join(', ')
})

function toggleType(type) {
  if (props.modelValue.includes(type)) {
    emit('update:modelValue', props.modelValue.filter((item) => item !== type))
    return
  }

  emit('update:modelValue', [...props.modelValue, type].sort((a, b) => a - b))
}

function resetAll() {
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
      :title="'Filter rows by move types'"
    >
      <i class="bi bi-funnel me-1"></i>{{ buttonLabel }}
    </button>
    <div class="dropdown-menu p-2" style="min-width: 220px;">
      <div class="d-flex justify-content-between align-items-center mb-2">
        <strong class="small">Move types</strong>
        <button type="button" class="btn btn-link btn-sm p-0 text-decoration-none" @click="resetAll">All</button>
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
