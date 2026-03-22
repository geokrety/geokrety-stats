<script setup lang="ts">
import { RouterLink } from 'vue-router'
import GeokretyTypeBadge from '@/components/GeokretyTypeBadge.vue'
import type { GeokretListItem } from '@/types/api'

defineProps<{
  gk: GeokretListItem
}>()
</script>

<template>
  <div class="flex items-center gap-3 py-2">
    <GeokretyTypeBadge :type="gk.type" icon-only class="shrink-0" />
    <div class="min-w-0 flex-1">
      <RouterLink
        :to="{ name: 'geokret-detail', params: { gkid: gk.gkid ?? gk.id } }"
        class="text-sm font-medium text-primary hover:underline truncate block"
      >
        {{ gk.gkid ?? `#${gk.id}` }} — {{ gk.name }}
      </RouterLink>
      <div class="text-xs text-muted-foreground capitalize">{{ gk.typeName }}</div>
    </div>
  </div>
</template>
