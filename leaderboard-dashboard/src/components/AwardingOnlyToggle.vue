<script setup>
const props = defineProps({
  modelValue: { type: Boolean, default: false },
  label: { type: String, default: 'Only awarding points' },
})

const emit = defineEmits(['update:modelValue'])

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>

<template>
  <button
    type="button"
    class="btn btn-sm"
    :class="modelValue ? 'btn-primary' : 'btn-outline-secondary'"
    @click="toggle"
  >
    <i class="bi bi-coin me-1"></i>{{ label }}
  </button>
</template>
