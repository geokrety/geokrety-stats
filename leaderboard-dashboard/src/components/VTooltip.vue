<script setup>
import { onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({
  text: {
    type: String,
    default: ''
  },
  placement: {
    type: String,
    default: 'top'
  }
})

const activator = ref(null)
let tooltipInstance = null

const initTooltip = () => {
  if (activator.value && window.bootstrap?.Tooltip) {
    if (tooltipInstance) {
      tooltipInstance.dispose()
    }
    // The first child of the activator slot is our target
    const target = activator.value.firstElementChild || activator.value
    tooltipInstance = new window.bootstrap.Tooltip(target, {
      title: props.text,
      placement: props.placement,
      trigger: 'hover focus',
      html: true // Allow HTML in tooltips as used in App.vue
    })
  }
}

watch(() => props.text, (newText) => {
  if (tooltipInstance) {
    tooltipInstance.setContent({ '.tooltip-inner': newText })
  }
})

onMounted(() => {
  // Small delay to ensure slot content is rendered
  setTimeout(initTooltip, 0)
})

onUnmounted(() => {
  if (tooltipInstance) {
    tooltipInstance.dispose()
  }
})
</script>

<template>
  <span ref="activator" class="d-inline-block">
    <slot name="activator" :props="{}"></slot>
    <slot></slot>
  </span>
</template>
