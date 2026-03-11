<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { ArrowLeft, Globe } from 'lucide-vue-next'
import { useCountriesStore } from '@/stores/countries'
import { useStatsStore } from '@/stores/stats'
import ActivityKpiCard from '@/components/kpi/ActivityKpiCard.vue'
import MoveTypeBreakdown from '@/components/breakdowns/MoveTypeBreakdown.vue'
import PointsSummary from '@/components/PointsSummary.vue'
import CountryGeokretyMap from '@/components/CountryGeokretyMap.vue'
import UserLeaderboard from '@/components/UserLeaderboard.vue'
import RecentActivity from '@/components/RecentActivity.vue'

const route = useRoute()
const store = useCountriesStore()
const statsStore = useStatsStore()

onMounted(() => {
  store.fetchCountries()
  statsStore.fetchAll()
})

const code = computed(() => String(route.params.code).toUpperCase())

const country = computed(() => store.countries.find((c) => c.code === code.value) ?? null)
</script>

<template>
  <div class="min-h-screen bg-background text-foreground pb-16">
    <!-- ── Back navigation ─────────────────────────────────────── -->
    <div class="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 pt-6">
      <RouterLink
        to="/countries"
        class="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
      >
        <ArrowLeft class="h-4 w-4" />
        All countries
      </RouterLink>
    </div>

    <!-- ── Loading state ───────────────────────────────────────── -->
    <div
      v-if="store.loading && !country"
      class="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 flex justify-center py-24"
    >
      <div class="h-8 w-8 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
    </div>

    <!-- ── Not found ───────────────────────────────────────────── -->
    <div
      v-else-if="!store.loading && !country"
      class="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 py-24 text-center"
    >
      <Globe class="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
      <h1 class="text-xl font-semibold text-muted-foreground">Country not found</h1>
      <p class="text-muted-foreground mt-1">No data for country code "{{ code }}".</p>
      <RouterLink
        to="/countries"
        class="mt-6 inline-flex items-center gap-2 text-sm text-foreground hover:underline"
      >
        <ArrowLeft class="h-4 w-4" /> Back to rankings
      </RouterLink>
    </div>

    <!-- ── Country page ────────────────────────────────────────── -->
    <template v-else-if="country">
      <!-- Header -->
      <header class="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 py-8">
        <div class="flex items-center gap-4">
          <span class="text-6xl leading-none" role="img" :aria-label="country.name">
            {{ country.flag }}
          </span>
          <div>
            <h1 class="text-3xl font-bold tracking-tight">{{ country.name }}</h1>
            <p class="mt-0.5 font-mono text-sm text-muted-foreground">{{ country.code }}</p>
          </div>
        </div>
      </header>

      <div class="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 space-y-8">
        <!-- ── Top KPIs ─────────────────────────────────────────── -->
        <section>
          <h2 class="mb-3 text-xs font-semibold uppercase tracking-widest text-muted-foreground">
            Activity overview
          </h2>
          <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
            <ActivityKpiCard stat="moves" label="Total moves" :value="country.movesCount" />
            <ActivityKpiCard stat="activeUsers" label="Active users" :value="country.activeUsers" />
            <ActivityKpiCard stat="usersHome" label="Home users" :value="country.usersHome" />
            <ActivityKpiCard
              stat="avgPointsPerMove"
              label="Avg pts / move"
              :value="country.avgPointsPerMove"
              :decimals="2"
            />
            <ActivityKpiCard stat="loves" label="Loves" :value="country.loves" />
            <ActivityKpiCard stat="pictures" label="Pictures" :value="country.pictures" />
            <ActivityKpiCard stat="inCache" label="In cache" :value="country.geokretyInCache" />
            <ActivityKpiCard stat="lost" label="Lost" :value="country.geokretyLost" />
          </div>
        </section>

        <!-- ── Move type breakdown ───────────────────────────────── -->
        <MoveTypeBreakdown :moves-by-type="country.movesByType" :moves-count="country.movesCount" />

        <!-- ── Points summary ────────────────────────────────────── -->
        <PointsSummary
          :country-name="country.name"
          :points-sum="country.pointsSum"
          :points-sum-moves="country.pointsSumMoves"
        />

        <!-- ── GeoKrety map ──────────────────────────────────────── -->
        <CountryGeokretyMap
          :country-code="country.code"
          :in-cache-count="country.geokretyInCache"
          :lost-count="country.geokretyLost"
        />

        <!-- ── Leaderboards ──────────────────────────────────────── -->
        <section>
          <h2 class="mb-4 text-xs font-semibold uppercase tracking-widest text-muted-foreground">
            Leaderboard — {{ country.name }}
          </h2>
          <p class="-mt-2 mb-4 text-xs text-muted-foreground">
            Top users and GeoKrety activity in this country (global data shown until
            country-specific API is available).
          </p>
          <UserLeaderboard />
        </section>

        <!-- ── Recent activity ──────────────────────────────────── -->
        <section class="overflow-hidden rounded-xl border border-border bg-card/70">
          <div class="px-5 pt-5 pb-1">
            <h2 class="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
              Recent activity — {{ country.name }}
            </h2>
            <p class="mt-1 text-xs text-muted-foreground">
              Latest GeoKret moves in this country (global feed shown until country-specific API is
              available).
            </p>
          </div>
          <RecentActivity />
        </section>
      </div>
    </template>
  </div>
</template>
