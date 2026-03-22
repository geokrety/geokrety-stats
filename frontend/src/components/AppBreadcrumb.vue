<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { ChevronRight, Home } from 'lucide-vue-next'

interface BreadcrumbItem {
  label: string
  to?: string | { name: string; params?: Record<string, string | number> }
}

const props = defineProps<{
  items: BreadcrumbItem[]
}>()

const route = useRoute()

const allItems = computed(() => {
  const home: BreadcrumbItem = { label: 'Home', to: '/' }
  return [home, ...props.items]
})
</script>

<template>
  <nav aria-label="Breadcrumb" class="mb-6">
    <ol class="flex items-center gap-1.5 text-sm text-muted-foreground">
      <li v-for="(item, index) in allItems" :key="index" class="flex items-center gap-1.5">
        <ChevronRight v-if="index > 0" class="h-3.5 w-3.5 flex-shrink-0" />
        <template v-if="index === 0 && item.to">
          <RouterLink :to="item.to" class="hover:text-foreground transition-colors">
            <Home class="h-4 w-4" />
          </RouterLink>
        </template>
        <template v-else-if="item.to && index < allItems.length - 1">
          <RouterLink :to="item.to" class="hover:text-foreground transition-colors truncate max-w-[200px]">
            {{ item.label }}
          </RouterLink>
        </template>
        <template v-else>
          <span class="font-medium text-foreground truncate max-w-[200px]">{{ item.label }}</span>
        </template>
      </li>
    </ol>
  </nav>
</template>
