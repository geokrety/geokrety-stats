<script setup lang="ts">
import TheHero from '@/components/TheHero.vue'
import RecentActivity from '@/components/RecentActivity.vue'
import UserLeaderboard from '@/components/UserLeaderboard.vue'
</script>

<template>
  <main class="w-full">
    <TheHero />
    <!-- Two-column layout on large screens -->
    <div class="w-full bg-background">
      <div class="mx-auto max-w-7xl">
        <div class="grid lg:grid-cols-2 lg:divide-x lg:divide-border/50">
          <RecentActivity />
          <UserLeaderboard />
        </div>
      </div>
    </div>
  </main>
</template>
