<script setup>
const props = defineProps({
  years: { type: Array, default: () => [] },
  modelValue: { type: String, default: '' },
})

const emit = defineEmits(['update:modelValue'])

function selectYear(year) {
  emit('update:modelValue', String(year))
}
</script>

<template>
  <div v-if="years.length" class="dropdown">
    <button
      class="btn btn-sm w-100"
      :class="modelValue ? 'btn-info' : 'btn-outline-secondary'"
      type="button"
      data-bs-toggle="dropdown"
      aria-expanded="false"
      title="Select a year period"
    >
      {{ modelValue ? `${modelValue} 📅` : 'Select Year...' }}
    </button>
    <ul class="dropdown-menu dropdown-menu-end" style="max-height: 300px; overflow-y: auto;">
      <li v-for="year in years" :key="year">
        <button type="button" class="dropdown-item" :class="modelValue === String(year) ? 'active' : ''" @click="selectYear(year)">
          {{ year }}
        </button>
      </li>
    </ul>
  </div>
</template>
