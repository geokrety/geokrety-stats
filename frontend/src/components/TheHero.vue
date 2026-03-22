<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useGlobalStats } from '@/composables/useGlobalStats'
import ActivityKpiCard from '@/components/kpi/ActivityKpiCard.vue'
import GeokretyKpiCard from '@/components/kpi/GeokretyKpiCard.vue'
import MoveKpiCard from '@/components/kpi/MoveKpiCard.vue'
import PictureKpiCard from '@/components/kpi/PictureKpiCard.vue'

const visible = ref(false)
onMounted(() => {
  setTimeout(() => (visible.value = true), 100)
  fetch()
})

const { stats, fetch } = useGlobalStats()
const geokretyByType = computed(() => (stats.value?.geokretyByType ? { ...stats.value.geokretyByType } : undefined))
const movesByType = computed(() => (stats.value?.movesByType ? { ...stats.value.movesByType } : undefined))
</script>

<template>
  <section class="relative overflow-hidden bg-background pt-24 pb-16 md:pt-32" aria-label="Hero">
    <div class="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div
        class="flex flex-col items-center text-center transition-all duration-700"
        :class="visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'"
      >
        <!-- Headline -->
        <h1
          class="mb-6 max-w-4xl text-5xl font-extrabold leading-tight tracking-tight text-foreground sm:text-6xl lg:text-7xl"
        >
          GeoKrety Stats
        </h1>

        <!-- Subtitle -->
        <p class="mb-10 max-w-2xl text-lg text-muted-foreground leading-relaxed">
          Discover how physical objects travel across countries, who moves them the most, and where
          they end up.
        </p>
      </div>
    </div>
    <!-- Stats KPI grid -->
    <div class="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 mt-10">
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
        <GeokretyKpiCard
          label="GeoKrety"
          :value="stats?.totalGeokrety ?? '—'"
          :hidden-count="stats?.totalGeokretyHidden ?? 0"
          :geokrety-by-type="geokretyByType"
        />
        <MoveKpiCard
          label="Total Moves"
          :value="stats?.totalMoves ?? '—'"
          :moves-by-type="movesByType"
          :moves-last30-days="stats?.movesLast30Days"
        />
        <ActivityKpiCard
          stat="users"
          label="Users"
          :value="stats?.registeredUsers ?? '—'"
          :secondary-value="stats?.activeUsersLast30d"
          secondary-label="active in last 30 days"
        />
        <ActivityKpiCard
          stat="countries"
          label="Countries"
          :value="stats?.countriesReached ?? '—'"
        />
        <PictureKpiCard
          stat="pictures"
          label="Pictures"
          :value="stats?.picturesUploaded ?? '—'"
          :pictures-by-type="stats?.picturesByType"
          class="col-span-2 sm:col-span-1"
        />
      </div>
    </div>
  </section>
</template>
