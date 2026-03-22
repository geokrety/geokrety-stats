<script setup lang="ts">
import { onMounted, watch, nextTick } from 'vue'
import { useLeaderboard } from '@/composables/useLeaderboard'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'
import LeaderboardTable from '@/components/LeaderboardTable.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Trophy } from 'lucide-vue-next'

const pageSize = 20
const { users, loading, error, hasMore, fetch, loadMore } = useLeaderboard()
const { sentinel, observe } = useInfiniteScroll(() => {
  if (!loading.value && hasMore.value) loadMore(pageSize)
})

onMounted(() => fetch(pageSize))

watch([users, hasMore], () => {
  nextTick(() => observe())
})
</script>

<template>
  <main class="min-h-screen bg-background text-foreground pb-16">
    <div class="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 pt-10">
      <div class="flex items-center gap-3 mb-6">
        <Trophy class="h-7 w-7 text-primary" />
        <h1 class="text-3xl font-bold tracking-tight">Leaderboard</h1>
      </div>

      <!-- Error state -->
      <div v-if="error" class="rounded-lg border border-destructive/50 bg-destructive/10 p-4 mb-6">
        <p class="text-sm text-destructive">{{ error }}</p>
        <Button variant="outline" size="sm" class="mt-2" @click="fetch(pageSize)">Retry</Button>
      </div>

      <!-- Loading initial -->
      <div v-if="loading && users.length === 0" class="flex justify-center py-16">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
      </div>

      <!-- Leaderboard table -->
      <Card v-else-if="users.length > 0">
        <CardHeader>
          <CardTitle class="text-lg">Top Users</CardTitle>
        </CardHeader>
        <CardContent>
          <LeaderboardTable :users="users" />

          <!-- Infinite scroll sentinel -->
          <div ref="sentinel" class="flex justify-center pt-4 mt-4 border-t border-border">
            <div v-if="loading" class="h-6 w-6 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
            <p v-else-if="!hasMore" class="text-sm text-muted-foreground">All users loaded.</p>
          </div>
        </CardContent>
      </Card>

      <!-- Empty state -->
      <div v-else class="text-center py-16">
        <Trophy class="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
        <p class="text-muted-foreground">No leaderboard data available.</p>
      </div>
    </div>
  </main>
</template>
