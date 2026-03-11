<script setup lang="ts">
import { Trophy } from 'lucide-vue-next'
import { useStatsStore } from '@/stores/stats'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

const store = useStatsStore()

function medalClasses(rank: number): string {
  if (rank === 1) return 'text-yellow-400'
  if (rank === 2) return 'text-muted-foreground'
  if (rank === 3) return 'text-amber-600'
  return 'text-muted-foreground'
}

function formatNumber(n: number): string {
  return n.toLocaleString('en-US')
}
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
    <div v-if="store.loadingLeaderboard" class="space-y-2">
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

    <!-- Table -->
    <div v-else class="overflow-hidden rounded-2xl border border-border bg-card">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border text-left text-muted-foreground">
            <th class="px-4 py-3 font-medium">#</th>
            <th class="px-4 py-3 font-medium">User</th>
            <th class="px-4 py-3 font-medium text-right">Moves</th>
            <th class="px-4 py-3 font-medium text-right">Points</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="user in store.leaderboard"
            :key="user.rank"
            class="group border-b border-border last:border-b-0 transition-colors hover:bg-card/5"
          >
            <!-- Rank -->
            <td class="px-4 py-3 w-10">
              <Trophy v-if="user.rank <= 3" class="h-5 w-5" :class="medalClasses(user.rank)" />
              <span v-else class="font-medium text-muted-foreground">{{ user.rank }}</span>
            </td>

            <!-- Avatar + username -->
            <td class="px-4 py-3">
              <div class="flex items-center gap-3">
                <Avatar class="h-9 w-9 flex-shrink-0">
                  <AvatarFallback
                    class="text-xs font-bold text-foreground"
                    :class="user.avatarColor"
                  >
                    {{ user.initials }}
                  </AvatarFallback>
                </Avatar>
                <span
                  class="font-medium text-foreground group-hover:text-emerald-400 transition-colors"
                >
                  {{ user.username }}
                </span>
              </div>
            </td>

            <!-- Moves -->
            <td class="px-4 py-3 text-right text-muted-foreground">
              {{ formatNumber(user.movesCount) }}
            </td>

            <!-- Points -->
            <td class="px-4 py-3 text-right">
              <span class="font-semibold text-emerald-400">
                {{ formatNumber(user.points) }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
