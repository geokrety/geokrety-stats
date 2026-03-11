<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useStatsStore } from '@/stores/stats'
import ActivityKpiCard from '@/components/kpi/ActivityKpiCard.vue'
import GeokretyKpiCard from '@/components/kpi/GeokretyKpiCard.vue'
import MoveKpiCard from '@/components/kpi/MoveKpiCard.vue'
import PictureKpiCard from '@/components/kpi/PictureKpiCard.vue'

const visible = ref(false)
onMounted(() => {
  setTimeout(() => (visible.value = true), 100)
})

const store = useStatsStore()
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
          :value="store.globalStats?.totalGeokrety ?? '—'"
          :hidden-count="store.globalStats?.totalGeokretyHidden ?? 0"
          :geokrety-by-type="store.globalStats?.geokretyByType"
        />
        <MoveKpiCard
          label="Total Moves"
          :value="store.globalStats?.totalMoves ?? '—'"
          :moves-by-type="store.globalStats?.movesByType"
          :moves-last30-days="store.globalStats?.movesLast30Days"
        />
        <ActivityKpiCard
          stat="users"
          label="Users"
          :value="store.globalStats?.registeredUsers ?? '—'"
          :secondary-value="store.globalStats?.activeUsersLast30d"
          secondary-label="active in last 30 days"
        />
        <ActivityKpiCard
          stat="countries"
          label="Countries"
          :value="store.globalStats?.countriesReached ?? '—'"
        />
        <PictureKpiCard
          stat="pictures"
          label="Pictures"
          :value="store.globalStats?.picturesUploaded ?? '—'"
          :pictures-by-type="store.globalStats?.picturesByType"
          class="col-span-2 sm:col-span-1"
        />
      </div>
    </div>
  </section>
</template>
