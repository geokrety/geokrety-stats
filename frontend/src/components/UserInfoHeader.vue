<script setup lang="ts">
import type { UserDetails } from '@/services/api/users'
import AvatarDisplay from '@/components/AvatarDisplay.vue'
import { countryCodeToFlag } from '@/lib/countryFlag'
import { formatDateTime } from '@/lib/dates'
import { Calendar } from 'lucide-vue-next'

defineProps<{
  user: UserDetails
}>()
</script>

<template>
  <div class="flex items-center gap-4 mb-8">
    <AvatarDisplay :src="user.avatarUrl || undefined" :alt="user.username" size="xl" :hover-delay="0" />
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{{ user.username }}</h1>
      <div class="flex items-center gap-2 text-sm text-muted-foreground mt-1">
        <Calendar class="h-4 w-4" />
        <span>Joined {{ formatDateTime(user.joinedAt) }}</span>
        <span v-if="user.homeCountry"> · {{ countryCodeToFlag(user.homeCountry) }} {{ user.homeCountry }}</span>
      </div>
    </div>
  </div>
</template>
