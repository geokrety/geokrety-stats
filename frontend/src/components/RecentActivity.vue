<script setup lang="ts">
import { computed } from 'vue'
import { Activity } from 'lucide-vue-next'
import { useStatsStore } from '@/stores/stats'
import MoveTypeIcon from '@/components/MoveTypeIcon.vue'
import MoveTypeBadge from '@/components/MoveTypeBadge.vue'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

const store = useStatsStore()

function relativeTime(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

const items = computed(() => store.recentActivity)
</script>

<template>
  <div class="py-12 px-4 sm:px-6 lg:px-8">
    <div class="mb-10 flex items-center justify-between">
      <div>
        <div class="flex items-center gap-2 mb-1">
          <Activity class="h-5 w-5 text-emerald-400" />
          <h2 class="text-3xl font-bold text-foreground">Recent Activity</h2>
        </div>
        <p class="text-muted-foreground">Latest GeoKret moves from around the globe</p>
      </div>
    </div>

    <!-- Skeleton -->
    <div v-if="store.loadingActivity" class="space-y-3">
      <Card v-for="i in 5" :key="i" class="rounded-xl border">
        <CardContent class="flex items-center gap-4 p-4">
          <Skeleton class="h-9 w-9 flex-shrink-0 rounded-lg" />
          <div class="flex-1 space-y-2">
            <Skeleton class="h-4 w-40" />
            <Skeleton class="h-3 w-56" />
          </div>
          <Skeleton class="h-3 w-12" />
        </CardContent>
      </Card>
    </div>

    <!-- Activity feed -->
    <div v-else class="space-y-3">
      <TransitionGroup
        tag="div"
        class="space-y-3"
        enter-active-class="transition-all duration-300 ease-out"
        leave-active-class="transition-all duration-300 ease-in"
        enter-from-class="-translate-y-2 opacity-0"
        leave-to-class="translate-y-2 opacity-0"
      >
        <Card
          v-for="move in items"
          :key="move.id"
          class="group rounded-xl border transition-all hover:border-emerald-500/30"
        >
          <CardContent class="flex items-center gap-4 p-4">
            <!-- Type icon -->
            <MoveTypeIcon :type="move.type" class="flex-shrink-0" />

            <!-- Info -->
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="truncate font-semibold text-foreground">{{ move.geokretName }}</span>
                <MoveTypeBadge :type="move.type" />
              </div>
              <div class="mt-0.5 text-sm text-muted-foreground">
                by
                <span class="font-medium text-foreground">{{ move.username }}</span>
                &nbsp;·&nbsp;
                {{ move.countryFlag }} {{ move.country }}
              </div>
            </div>

            <!-- Time -->
            <div class="flex-shrink-0 text-xs text-muted-foreground">
              {{ relativeTime(move.timestamp) }}
            </div>
          </CardContent>
        </Card>
      </TransitionGroup>
    </div>
  </div>
</template>
