<script setup lang="ts">
import { onMounted } from 'vue'
import { Trophy } from 'lucide-vue-next'
import { useLeaderboard } from '@/composables/useLeaderboard'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import LeaderboardTable from '@/components/LeaderboardTable.vue'
import { RouterLink } from 'vue-router'

const { users, loading, error, fetch } = useLeaderboard()

onMounted(() => fetch(10))
</script>

<template>
  <div class="py-12 px-4 sm:px-6 lg:px-8">
    <div class="mb-10 flex items-center gap-3">
      <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-500/10">
        <Trophy class="h-5 w-5 text-amber-400" />
      </div>
      <div>
        <h2 class="text-3xl font-bold text-foreground">Leaderboard</h2>
        <p class="text-sm text-muted-foreground">Top 10 most active GeoKrety users</p>
      </div>
    </div>

    <!-- Skeleton -->
    <div v-if="loading && users.length === 0" class="space-y-2">
      <Card v-for="i in 10" :key="i" class="rounded-xl border">
        <CardContent class="flex items-center gap-4 px-4 py-3">
          <Skeleton class="h-5 w-5" />
          <Skeleton class="h-9 w-9 rounded-full" />
          <Skeleton class="h-4 w-32 flex-1" />
          <Skeleton class="h-4 w-16" />
          <Skeleton class="hidden h-4 w-20 sm:block" />
        </CardContent>
      </Card>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
      <p class="text-sm text-destructive">{{ error }}</p>
    </div>

    <!-- Table -->
    <div v-else-if="users.length > 0" class="overflow-hidden rounded-2xl border border-border bg-card">
      <LeaderboardTable :users="users" />
      <div class="border-t border-border px-4 py-3 text-center">
        <RouterLink
          :to="{ name: 'leaderboard' }"
          class="text-sm text-primary hover:underline"
        >
          View full leaderboard →
        </RouterLink>
      </div>
    </div>
  </div>
</template>
