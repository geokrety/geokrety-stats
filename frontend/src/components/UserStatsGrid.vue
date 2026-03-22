<script setup lang="ts">
import type { UserDetails } from '@/services/api/users'
import { relativeTime } from '@/lib/dates'
import ActivityKpiCard from '@/components/kpi/ActivityKpiCard.vue'
import GeokretyKpiCard from '@/components/kpi/GeokretyKpiCard.vue'
import MoveKpiCard from '@/components/kpi/MoveKpiCard.vue'
import PictureKpiCard from '@/components/kpi/PictureKpiCard.vue'

defineProps<{
  user: UserDetails
}>()
</script>

<template>
  <div class="grid grid-cols-2 sm:grid-cols-3 gap-4 mb-8">
    <GeokretyKpiCard label="GeoKrety owned" :value="user.ownedGeokretyCount" />
    <MoveKpiCard label="Moves" :value="user.movesCount" />
    <GeokretyKpiCard label="Distinct GeoKrety" :value="user.distinctGeokretyCount" />
    <ActivityKpiCard stat="countries" label="Countries active" :value="user.activeCountriesCount" />
    <PictureKpiCard stat="pictures" label="Pictures" :value="user.picturesCount" />
    <ActivityKpiCard
      v-if="user.lastMoveAt"
      stat="moves"
      label="Last move"
      :value="relativeTime(user.lastMoveAt)"
    />
  </div>
</template>
