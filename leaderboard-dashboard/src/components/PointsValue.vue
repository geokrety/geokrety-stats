<script setup>
import { computed } from 'vue'
import { usePointsDisplay } from '../composables/usePointsDisplay.js'

const props = defineProps({
  value: { type: [Number, String, null], default: 0 },
  digits: { type: Number, default: 3 },
  showPlus: { type: Boolean, default: false },
  bold: { type: Boolean, default: false },
})

const { formatPoints, pointsClass } = usePointsDisplay()

const textClass = computed(() => {
  const classes = [pointsClass(props.value)]
  if (props.bold) classes.push('fw-semibold')
  return classes
})
</script>

<template>
  <span :class="textClass">{{ formatPoints(value, digits, showPlus) }}</span>
</template>
