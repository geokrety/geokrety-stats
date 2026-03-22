<script setup lang="ts">
import type { RecentMove } from '@/types/api'
import MoveTypeIcon from '@/components/MoveTypeIcon.vue'
import MoveTypeBadge from '@/components/MoveTypeBadge.vue'
import { countryCodeToFlag } from '@/lib/countryFlag'
import { relativeTime } from '@/lib/dates'
import { Card, CardContent } from '@/components/ui/card'

defineProps<{
  move: RecentMove
}>()
</script>

<template>
  <Card class="group rounded-xl border transition-all hover:border-primary/30">
    <CardContent class="flex items-center gap-4 p-4">
      <MoveTypeIcon :type="move.type" class="flex-shrink-0" />
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2 flex-wrap">
          <span class="truncate font-semibold text-foreground">{{ move.geokretName }}</span>
          <MoveTypeBadge :type="move.type" />
        </div>
        <div class="mt-0.5 text-sm text-muted-foreground">
          by <span class="font-medium text-foreground">{{ move.username }}</span>
          &nbsp;·&nbsp;
          {{ countryCodeToFlag(move.country) }} {{ move.country }}
        </div>
      </div>
      <div class="flex-shrink-0 text-xs text-muted-foreground">
        {{ relativeTime(move.timestamp) }}
      </div>
    </CardContent>
  </Card>
</template>
