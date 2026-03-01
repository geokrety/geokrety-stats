<script setup>
import { onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({
  text: {
    type: String,
    default: ''
  },
  html: {
    type: String,
    default: ''
  },
  placement: {
    type: String,
    default: 'top'
  }
})

const activator = ref(null)
const contentRef = ref(null)
let tooltipInstance = null

const getTooltipTitle = () => {
  if (props.text) return props.text
  if (props.html) return props.html
  if (contentRef.value && contentRef.value.innerHTML.trim()) return contentRef.value.innerHTML
  return ''
}

const initTooltip = () => {
  if (activator.value && window.bootstrap?.Tooltip) {
    const target = activator.value.firstElementChild || activator.value

    tooltipInstance = new window.bootstrap.Tooltip(target, {
      title: getTooltipTitle,
      placement: props.placement,
      trigger: 'hover focus',
      html: true,
      container: 'body'
    })
  }
}

watch([() => props.text, () => props.html], () => {
  if (tooltipInstance) {
    tooltipInstance.setContent({ '.tooltip-inner': getTooltipTitle() })
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
  <span ref="activator" style="display: contents;">
    <slot name="activator" :props="{}"></slot>
  </span>
  <span ref="contentRef" style="display: none;">
    <slot></slot>
  </span>
</template>
