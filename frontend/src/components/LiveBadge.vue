<script setup lang="ts">
import { Wifi } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { ref, onMounted, onBeforeUnmount } from 'vue'

const props = withDefaults(defineProps<{ showText?: boolean }>(), { showText: true })

const collapsed = ref(false)
let observer: MutationObserver | null = null

function updateCollapsed() {
  const sidebar = document.querySelector('[data-slot="sidebar"]')
  collapsed.value = !!sidebar && sidebar.getAttribute('data-state') === 'collapsed'
}

onMounted(() => {
  updateCollapsed()
  const target = document.querySelector('[data-slot="sidebar"]')
  if (target) {
    observer = new MutationObserver(() => updateCollapsed())
    observer.observe(target, { attributes: true, attributeFilter: ['data-state'] })
  }
})

onBeforeUnmount(() => {
  observer?.disconnect()
  observer = null
})
</script>

<template>
  <Badge variant="secondary" class="inline-flex items-center gap-2">
    <Wifi class="h-3 w-3 flex-shrink-0 animate-pulse" />
    <span v-if="props.showText && !collapsed" class="whitespace-nowrap">Live</span>
  </Badge>
</template>

<style>
/* Hide the textual part of the badge when the sidebar is collapsed (icon-only mode) */
[data-slot='sidebar'][data-state='collapsed'] .whitespace-nowrap {
  display: none !important;
}

/* If the sidebar-wrapper contains a collapsed sidebar, hide badge text inside footer/card areas too. */
[data-slot='sidebar-wrapper']:has([data-slot='sidebar'][data-state='collapsed'])
  .bg-card
  .whitespace-nowrap {
  display: none !important;
}
</style>
