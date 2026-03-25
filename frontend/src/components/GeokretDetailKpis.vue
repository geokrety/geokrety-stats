<script setup lang="ts">
import type { GeokretDetails } from '@/services/api/geokrety'
import ActivityKpiCard from '@/components/kpi/ActivityKpiCard.vue'
import GeokretyKpiCard from '@/components/kpi/GeokretyKpiCard.vue'
import LovesKpiCard from '@/components/kpi/LovesKpiCard.vue'
import PictureKpiCard from '@/components/kpi/PictureKpiCard.vue'

defineProps<{
  geokret: GeokretDetails
}>()
</script>

<template>
  <div class="grid grid-cols-2 gap-4 sm:grid-cols-4 mb-8">
    <ActivityKpiCard stat="moves" label="Km traveled" :value="geokret.distanceKm" />
    <GeokretyKpiCard label="Caches visited" :value="geokret.cachesCount" />
    <LovesKpiCard label="Loves" :value="geokret.lovesCount" />
    <PictureKpiCard stat="pictures" label="Pictures" :value="geokret.picturesCount" />
  </div>
</template>
