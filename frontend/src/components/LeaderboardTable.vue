<script setup lang="ts">
import type { LeaderboardUser } from '@/types/api'
import AvatarDisplay from '@/components/AvatarDisplay.vue'
import RankBadge from '@/components/RankBadge.vue'
import { RouterLink } from 'vue-router'

defineProps<{
  users: LeaderboardUser[]
}>()

function fmt(n: number): string {
  return n.toLocaleString('en-US')
}
</script>

<template>
  <table class="w-full text-sm">
    <thead>
      <tr class="border-b border-border text-left text-xs text-muted-foreground uppercase tracking-wider">
        <th class="py-2 px-3 w-12">#</th>
        <th class="py-2 px-3">User</th>
        <th class="py-2 px-3 text-right">Moves</th>
        <th class="py-2 px-3 text-right">Points</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="user in users"
        :key="user.userId"
        class="border-b border-border/50 last:border-0 hover:bg-muted/50 transition-colors"
      >
        <td class="py-3 px-3 font-medium">
          <RankBadge :rank="user.rank" />
        </td>
        <td class="py-3 px-3">
          <RouterLink
            :to="{ name: 'user-profile', params: { id: user.userId } }"
            class="flex items-center gap-3 hover:opacity-80 transition-opacity"
          >
            <AvatarDisplay
              :alt="user.username"
              size="sm"
              :hover-delay="0"
            />
            <span class="font-medium text-primary hover:underline">{{ user.username }}</span>
          </RouterLink>
        </td>
        <td class="py-3 px-3 text-right tabular-nums text-muted-foreground">
          {{ fmt(user.movesCount) }}
        </td>
        <td class="py-3 px-3 text-right tabular-nums font-semibold">
          {{ fmt(user.points) }}
        </td>
      </tr>
    </tbody>
  </table>
</template>
