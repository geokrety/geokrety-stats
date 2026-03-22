<script setup lang="ts">
import { onMounted, watch, nextTick } from 'vue'
import { useRecentMoves } from '@/composables/useRecentMoves'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'
import MoveCard from '@/components/MoveCard.vue'
import { Button } from '@/components/ui/button'
import { Activity } from 'lucide-vue-next'

const pageSize = 20
const { moves, loading, error, hasMore, fetch, loadMore } = useRecentMoves()
const { sentinel, observe } = useInfiniteScroll(() => {
  if (!loading.value && hasMore.value) loadMore(pageSize)
})

onMounted(() => fetch(pageSize))

watch([moves, hasMore], () => {
  nextTick(() => observe())
})
</script>

<template>
  <main class="min-h-screen bg-background text-foreground pb-16">
    <div class="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 pt-10">
      <div class="flex items-center gap-3 mb-6">
        <Activity class="h-7 w-7 text-primary" />
        <h1 class="text-3xl font-bold tracking-tight">Recent Moves</h1>
      </div>

      <!-- Error state -->
      <div v-if="error" class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 mb-6">
        <p class="text-sm text-destructive">{{ error }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="fetch(pageSize)">
          Retry
        </Button>
      </div>

      <!-- Loading initial -->
      <div v-if="loading && moves.length === 0" class="flex justify-center py-16">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
      </div>

      <!-- Moves list -->
      <div v-else-if="moves.length > 0" class="space-y-3">
        <MoveCard v-for="move in moves" :key="move.id" :move="move" />

        <!-- Infinite scroll sentinel -->
        <div ref="sentinel" class="flex justify-center py-4">
          <div v-if="loading" class="h-6 w-6 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
          <p v-else-if="!hasMore" class="text-sm text-muted-foreground">All moves loaded.</p>
        </div>
      </div>

      <!-- Empty state -->
      <div v-else class="text-center py-16">
        <Activity class="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
        <p class="text-muted-foreground">No recent moves available.</p>
      </div>
    </div>
  </main>
</template>
