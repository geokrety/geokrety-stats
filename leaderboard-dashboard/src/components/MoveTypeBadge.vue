<script setup>
import { getMoveTypeBadgeClass } from '../composables/useMoveTypeColors.js'

defineProps({
  type: {
    type: [Number, String],
    required: false
  },
  text: {
    type: String,
    required: true
  }
})
</script>

<template>
  <span class="badge" :class="getMoveTypeBadgeClass(text)">
    {{ text }}
  </span>
</template>
