<script setup lang="ts">
import { onMounted } from 'vue'
import { Activity } from 'lucide-vue-next'
import { useRecentMoves } from '@/composables/useRecentMoves'
import MoveCard from '@/components/MoveCard.vue'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { RouterLink } from 'vue-router'

const { moves, loading, fetch } = useRecentMoves()

onMounted(() => fetch(10))
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
    <div v-if="loading && moves.length === 0" class="space-y-3">
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
        <MoveCard v-for="move in moves" :key="move.id" :move="move" />
      </TransitionGroup>

      <div class="text-center pt-2">
        <RouterLink
          :to="{ name: 'recent-moves' }"
          class="text-sm text-primary hover:underline"
        >
          View all recent moves →
        </RouterLink>
      </div>
    </div>
  </div>
</template>
