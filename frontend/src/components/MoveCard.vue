<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { RecentMove } from '@/types/api'
import MoveTypeIcon from '@/components/MoveTypeIcon.vue'
import MoveTypeBadge from '@/components/MoveTypeBadge.vue'
import { countryCodeToFlag } from '@/lib/countryFlag'
import { relativeTime } from '@/lib/dates'
import { Card, CardContent } from '@/components/ui/card'

const props = defineProps<{
  move: RecentMove
}>()
</script>

<template>
  <Card class="group rounded-xl border transition-all hover:border-primary/30">
    <CardContent class="flex items-center gap-4 p-4">
      <MoveTypeIcon :type="move.type" class="flex-shrink-0" />
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2 flex-wrap">
          <RouterLink
            v-if="move.geokretGkid"
            :to="{ name: 'geokret-detail', params: { gkid: move.geokretGkid } }"
            class="truncate font-semibold text-foreground hover:text-primary hover:underline"
          >
            {{ move.geokretName }}
          </RouterLink>
          <span v-else class="truncate font-semibold text-foreground">{{ move.geokretName }}</span>
          <MoveTypeBadge :type="move.type" />
        </div>
        <div class="mt-0.5 text-sm text-muted-foreground">
          by
          <RouterLink
            v-if="move.userId"
            :to="{ name: 'user-profile', params: { id: move.userId } }"
            class="font-medium text-foreground hover:text-primary hover:underline"
          >
            {{ move.username }}
          </RouterLink>
          <span v-else class="font-medium text-foreground">{{ move.username }}</span>
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
