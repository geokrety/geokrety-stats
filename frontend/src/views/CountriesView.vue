<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  Globe,
  ChevronUp,
  ChevronDown,
  ChevronsUpDown,
  Heart,
  Camera,
  Package,
  MapPinOff,
  TrendingUp,
  Users,
  UserCheck,
  Footprints,
  Info,
} from 'lucide-vue-next'
import WorldChoropleth from '@/components/WorldChoropleth.vue'
import ViewToggle from '@/components/ViewToggle.vue'
import MoveTypeBreakdown from '@/components/breakdowns/MoveTypeBreakdown.vue'
import { Input } from '@/components/ui/input'
import { useCountriesStore } from '@/stores/countries'
import type { CountryStats } from '@/composables/useApi'

// ── Store ───────────────────────────────────────────────────────────────────
const store = useCountriesStore()
onMounted(() => store.fetchCountries())

// ── View mode: table | cards ─────────────────────────────────────────────────
const viewMode = ref<'table' | 'cards'>('table')

// ── Sort state ───────────────────────────────────────────────────────────────
type SortKey =
  | keyof CountryStats
  | 'movesByType.dropped'
  | 'movesByType.dipped'
  | 'movesByType.seen'

const sortKey = ref<SortKey>('movesCount')
const sortDir = ref<'asc' | 'desc'>('desc')

/** Map a sortKey to the metric used for the choropleth (must be a top-level numeric key) */
const choroplethMetric = computed<keyof CountryStats>(() => {
  switch (sortKey.value) {
    case 'movesByType.dropped':
    case 'movesByType.dipped':
    case 'movesByType.seen':
      return 'movesCount'
    default:
      return sortKey.value as keyof CountryStats
  }
})

function resolveValue(row: CountryStats, key: SortKey): number {
  switch (key) {
    case 'movesByType.dropped':
      return row.movesByType.dropped
    case 'movesByType.dipped':
      return row.movesByType.dipped
    case 'movesByType.seen':
      return row.movesByType.seen
    default:
      return (row[key as keyof CountryStats] as number) ?? 0
  }
}

function toggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'desc' ? 'asc' : 'desc'
  } else {
    sortKey.value = key
    sortDir.value = 'desc'
  }
}

// ── Sorted & ranked data ─────────────────────────────────────────────────────
const sortedCountries = computed<(CountryStats & { rank: number })[]>(() => {
  const rows = [...store.countries]
  rows.sort((a, b) => {
    const va = resolveValue(a, sortKey.value)
    const vb = resolveValue(b, sortKey.value)
    return sortDir.value === 'desc' ? vb - va : va - vb
  })
  return rows.map((r, i) => ({ ...r, rank: i + 1 }))
})

// ── Column definitions ────────────────────────────────────────────────────────
interface Column {
  key: SortKey
  label: string
  shortLabel?: string
  tooltip?: string
  align?: 'right' | 'left' | 'center'
}

const columns: Column[] = [
  {
    key: 'movesCount',
    label: 'Moves',
    tooltip: 'Drop + Dip/Visit + Seen (with location)',
    align: 'right',
  },
  {
    key: 'movesByType.dropped',
    label: 'Dropped',
    tooltip: 'Logtype 0 — GeoKret left in a cache',
    align: 'right',
  },
  {
    key: 'movesByType.dipped',
    label: 'Dipped',
    tooltip: 'Logtype 5 — visiting/dipping (always w/ location)',
    align: 'right',
  },
  {
    key: 'movesByType.seen',
    label: 'Seen',
    tooltip: 'Logtype 3 — seen/met WITH location confirmed',
    align: 'right',
  },
  {
    key: 'usersHome',
    label: 'Home users',
    tooltip: 'Users with home coordinates in this country',
    align: 'right',
  },
  {
    key: 'activeUsers',
    label: 'Active users',
    tooltip: 'Distinct users with qualifying moves here',
    align: 'right',
  },
  {
    key: 'loves',
    label: 'Loves ❤️',
    tooltip: 'Love/heart reactions on GeoKrety actually spotted in this country',
    align: 'right',
  },
  {
    key: 'pictures',
    label: 'Pictures 📸',
    tooltip: 'Photos attached to moves in this country',
    align: 'right',
  },
  {
    key: 'pointsSum',
    label: 'Points Σ (home)',
    tooltip: 'Sum of points earned by users whose home is in this country',
    align: 'right',
  },
  {
    key: 'pointsSumMoves',
    label: 'Points Σ (moves)',
    tooltip: 'Sum of points earned by moves made inside this country',
    align: 'right',
  },
  {
    key: 'geokretyInCache',
    label: 'In cache',
    tooltip: 'GeoKrety currently sitting in a local cache',
    align: 'right',
  },
  {
    key: 'geokretyLost',
    label: 'Lost',
    tooltip: 'GeoKrety marked missing, last position here',
    align: 'right',
  },
  {
    key: 'avgPointsPerMove',
    label: 'Avg pts/move',
    tooltip: 'Average points per move',
    align: 'right',
  },
]

function sortIcon(key: SortKey) {
  if (sortKey.value !== key) return 'none'
  return sortDir.value
}

// ── Search / filter ───────────────────────────────────────────────────────────
const search = ref('')
const filteredCountries = computed(() =>
  search.value.trim()
    ? sortedCountries.value.filter(
        (c) =>
          c.name.toLowerCase().includes(search.value.toLowerCase()) ||
          c.code.toLowerCase().includes(search.value.toLowerCase()),
      )
    : sortedCountries.value,
)

// ── Number formatting ─────────────────────────────────────────────────────────
function fmt(n: number | undefined, decimals = 0): string {
  if (n === undefined || n === null) return '—'
  return n.toLocaleString('en', { maximumFractionDigits: decimals })
}

// ── Map legend ────────────────────────────────────────────────────────────────
const legendMetricLabel = computed(() => {
  return columns.find((c) => c.key === sortKey.value)?.label ?? 'Moves'
})
</script>

<template>
  <main class="min-h-screen w-full overflow-x-hidden bg-background pb-16">
    <!-- ── Page header ──────────────────────────────────────────────────────── -->
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-6 pb-4">
      <div class="flex flex-col md:flex-row md:items-end md:justify-between gap-3">
        <div>
          <div class="flex items-center gap-2 mb-1">
            <Globe class="h-6 w-6 text-muted-foreground" />
            <h1 class="text-2xl font-bold text-foreground">Country Rankings</h1>
          </div>
          <p class="text-sm text-muted-foreground">
            Explore GeoKrety activity across the world. Click a column to sort — the map updates
            accordingly.
          </p>
        </div>

        <!-- Controls row -->
        <div class="flex items-center gap-3 flex-shrink-0">
          <!-- Search -->
          <Input v-model="search" type="search" placeholder="Search country…" />
          <!-- View toggle -->
          <ViewToggle v-model="viewMode" />
        </div>
      </div>
    </div>

    <!-- ── World Map ─────────────────────────────────────────────────────────── -->
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 mb-6">
      <div class="relative h-[380px] overflow-hidden rounded-xl border border-border bg-card">
        <!-- Loading overlay -->
        <div
          v-if="store.loading"
          class="absolute inset-0 z-[500] flex items-center justify-center bg-card/80"
        >
          <div
            class="h-8 w-8 animate-spin rounded-full border-2 border-border/20 border-t-primary"
          />
        </div>

        <!-- Map (only render when data is ready) -->
        <WorldChoropleth
          v-if="store.countries.length > 0"
          :countries="store.countries"
          :metric="choroplethMetric"
          class="h-full w-full"
        />

        <!-- Map legend -->
        <div
          class="absolute bottom-3 left-3 z-[400] flex items-center gap-2 rounded-lg border border-border bg-card/90 px-3 py-2 text-xs text-muted-foreground"
        >
          <span class="text-muted-foreground">Low</span>
          <div class="h-3 w-28 rounded-sm bg-gradient-to-r from-muted to-foreground" />
          <span class="text-muted-foreground">High</span>
          <span class="ml-2 font-medium text-foreground">{{ legendMetricLabel }}</span>
          <span
            class="ml-1 text-muted-foreground"
            title="Countries with no data are shown in dark grey"
          >
            <Info class="h-3 w-3" />
          </span>
        </div>
      </div>
    </div>

    <!-- ── Loading / Error states ────────────────────────────────────────────── -->
    <div
      v-if="store.loading && store.countries.length === 0"
      class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 flex justify-center py-16"
    >
      <div class="h-8 w-8 animate-spin rounded-full border-2 border-border/30 border-t-primary" />
    </div>

    <div v-else-if="store.error" class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div
        class="rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
      >
        {{ store.error }}
      </div>
    </div>

    <!-- ── TABLE view ────────────────────────────────────────────────────────── -->
    <div v-else-if="viewMode === 'table'" class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div
        class="max-w-full overflow-x-auto overscroll-x-contain rounded-xl border border-border bg-card"
      >
        <table class="min-w-full text-sm text-left">
          <thead>
            <tr class="border-b border-border">
              <!-- Rank -->
              <th
                class="w-10 whitespace-nowrap px-3 py-3 text-center text-xs font-semibold text-muted-foreground"
              >
                #
              </th>
              <!-- Country -->
              <th
                class="px-3 py-3 text-xs font-semibold text-muted-foreground whitespace-nowrap sticky left-0 bg-card z-10"
              >
                Country
              </th>
              <!-- Sortable columns -->
              <th
                v-for="col in columns"
                :key="col.key"
                class="px-3 py-3 text-xs font-semibold whitespace-nowrap cursor-pointer select-none transition-colors group"
                :class="[
                  'text-right',
                  sortKey === col.key
                    ? 'text-foreground'
                    : 'text-muted-foreground hover:text-foreground',
                ]"
                :title="col.tooltip"
                @click="toggleSort(col.key)"
              >
                <span class="inline-flex items-center justify-end gap-1">
                  {{ col.label }}
                  <ChevronUp
                    v-if="sortIcon(col.key) === 'asc'"
                    class="h-3.5 w-3.5 text-foreground"
                  />
                  <ChevronDown
                    v-else-if="sortIcon(col.key) === 'desc'"
                    class="h-3.5 w-3.5 text-foreground"
                  />
                  <ChevronsUpDown v-else class="h-3.5 w-3.5 opacity-0 group-hover:opacity-50" />
                </span>
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border/50">
            <tr
              v-for="row in filteredCountries"
              :key="row.code"
              class="transition-colors hover:bg-accent/30"
            >
              <!-- Rank -->
              <td class="px-3 py-2.5 text-center text-xs text-muted-foreground">{{ row.rank }}</td>

              <!-- Country name + flag + link -->
              <td class="px-3 py-2.5 sticky left-0 bg-card z-10">
                <RouterLink
                  :to="`/countries/${row.code.toLowerCase()}`"
                  class="inline-flex items-center gap-2 font-medium text-foreground hover:text-foreground/80 transition-colors"
                >
                  <span class="text-lg leading-none">{{ row.flag }}</span>
                  <span>{{ row.name }}</span>
                  <span class="text-xs text-muted-foreground">({{ row.code }})</span>
                </RouterLink>
              </td>

              <!-- Metric columns -->
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'movesCount' ? 'font-medium text-foreground' : 'text-foreground'
                "
              >
                {{ fmt(row.movesCount) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'movesByType.dropped'
                    ? 'text-foreground font-medium'
                    : 'text-muted-foreground'
                "
              >
                {{ fmt(row.movesByType.dropped) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'movesByType.dipped'
                    ? 'text-foreground font-medium'
                    : 'text-muted-foreground'
                "
              >
                {{ fmt(row.movesByType.dipped) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'movesByType.seen'
                    ? 'font-medium text-foreground'
                    : 'text-muted-foreground'
                "
              >
                {{ fmt(row.movesByType.seen) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="sortKey === 'usersHome' ? 'font-medium text-foreground' : 'text-foreground'"
              >
                {{ fmt(row.usersHome) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'activeUsers'
                    ? 'font-medium text-foreground'
                    : 'text-muted-foreground'
                "
              >
                {{ fmt(row.activeUsers) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="sortKey === 'loves' ? 'font-medium text-foreground' : 'text-foreground'"
              >
                {{ fmt(row.loves) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'pictures' ? 'font-medium text-foreground' : 'text-muted-foreground'
                "
              >
                {{ fmt(row.pictures) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="sortKey === 'pointsSum' ? 'font-medium text-foreground' : 'text-foreground'"
              >
                {{ fmt(row.pointsSum) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'pointsSumMoves'
                    ? 'font-medium text-foreground'
                    : 'text-muted-foreground'
                "
              >
                {{ fmt(row.pointsSumMoves) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'geokretyInCache' ? 'font-medium text-foreground' : 'text-foreground'
                "
              >
                {{ fmt(row.geokretyInCache) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'geokretyLost'
                    ? 'font-medium text-foreground'
                    : 'text-muted-foreground'
                "
              >
                {{ fmt(row.geokretyLost) }}
              </td>
              <td
                class="px-3 py-2.5 text-right tabular-nums"
                :class="
                  sortKey === 'avgPointsPerMove'
                    ? 'font-medium text-foreground'
                    : 'text-muted-foreground'
                "
              >
                {{ fmt(row.avgPointsPerMove, 2) }}
              </td>
            </tr>

            <!-- Empty state -->
            <tr v-if="filteredCountries.length === 0">
              <td
                :colspan="columns.length + 2"
                class="px-3 py-8 text-center text-sm text-muted-foreground"
              >
                No countries match your search.
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Row count -->
      <p class="mt-2 text-right text-xs text-muted-foreground">
        {{ filteredCountries.length }} of {{ store.countries.length }} countries
      </p>
    </div>

    <!-- ── CARDS view ─────────────────────────────────────────────────────────── -->
    <div v-else class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <!-- Sort picker for cards view -->
      <div class="mb-4 flex flex-wrap gap-2 items-center">
        <span class="text-xs text-muted-foreground">Sort by:</span>
        <button
          v-for="col in columns"
          :key="col.key"
          :class="[
            'rounded-full px-3 py-1 text-xs font-medium transition-colors border',
            sortKey === col.key
              ? 'border-border bg-accent text-accent-foreground'
              : 'border-border bg-card text-muted-foreground hover:text-foreground hover:border-border/20',
          ]"
          @click="toggleSort(col.key)"
        >
          {{ col.label }}
          <ChevronUp v-if="sortIcon(col.key) === 'asc'" class="inline h-3 w-3" />
          <ChevronDown v-else-if="sortIcon(col.key) === 'desc'" class="inline h-3 w-3" />
        </button>
      </div>

      <div class="grid gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
        <RouterLink
          v-for="row in filteredCountries"
          :key="row.code"
          :to="`/countries/${row.code.toLowerCase()}`"
          class="group relative flex flex-col rounded-xl border border-border bg-card p-4 hover:border-border/80 hover:bg-card/95 transition-all"
        >
          <!-- Rank badge -->
          <span class="absolute right-3 top-3 text-xs font-mono text-muted-foreground"
            >#{{ row.rank }}</span
          >

          <!-- Flag + Name -->
          <div class="flex items-center gap-2.5 mb-3">
            <span class="text-3xl leading-none">{{ row.flag }}</span>
            <div>
              <p
                class="font-semibold text-foreground group-hover:text-foreground/80 transition-colors leading-tight"
              >
                {{ row.name }}
              </p>
              <p class="text-xs text-muted-foreground">{{ row.code }}</p>
            </div>
          </div>

          <!-- Key stats grid -->
          <div class="grid grid-cols-2 gap-x-3 gap-y-2 text-xs">
            <div>
              <p class="flex items-center gap-1 text-muted-foreground">
                <Footprints class="h-3 w-3" /> Moves
              </p>
              <p class="tabular-nums font-semibold text-foreground">{{ fmt(row.movesCount) }}</p>
            </div>
            <div>
              <p class="flex items-center gap-1 text-muted-foreground">
                <Users class="h-3 w-3" /> Home users
              </p>
              <p class="tabular-nums font-semibold text-foreground">{{ fmt(row.usersHome) }}</p>
            </div>
            <div>
              <p class="flex items-center gap-1 text-muted-foreground">
                <Heart class="h-3 w-3" /> Loves
              </p>
              <p class="tabular-nums font-semibold text-foreground">{{ fmt(row.loves) }}</p>
            </div>
            <div>
              <p class="flex items-center gap-1 text-muted-foreground">
                <Camera class="h-3 w-3" /> Pics
              </p>
              <p class="tabular-nums font-semibold text-foreground">{{ fmt(row.pictures) }}</p>
            </div>
            <div>
              <p class="flex items-center gap-1 text-muted-foreground">
                <Package class="h-3 w-3" /> In cache
              </p>
              <p class="tabular-nums font-semibold text-foreground">
                {{ fmt(row.geokretyInCache) }}
              </p>
            </div>
            <div>
              <p class="flex items-center gap-1 text-muted-foreground">
                <MapPinOff class="h-3 w-3" /> Lost
              </p>
              <p class="tabular-nums font-semibold text-foreground">{{ fmt(row.geokretyLost) }}</p>
            </div>
            <div>
              <p class="flex items-center gap-1 text-muted-foreground">
                <TrendingUp class="h-3 w-3" /> Pts Σ home
              </p>
              <p class="tabular-nums font-semibold text-foreground">{{ fmt(row.pointsSum) }}</p>
            </div>
            <div>
              <p class="flex items-center gap-1 text-muted-foreground">
                <TrendingUp class="h-3 w-3" /> Pts Σ moves
              </p>
              <p class="tabular-nums font-semibold text-foreground">
                {{ fmt(row.pointsSumMoves) }}
              </p>
            </div>
            <div>
              <p class="flex items-center gap-1 text-muted-foreground">
                <UserCheck class="h-3 w-3" /> Active
              </p>
              <p class="tabular-nums font-semibold text-foreground">{{ fmt(row.activeUsers) }}</p>
            </div>
          </div>

          <!-- Move breakdown bar -->
          <div class="mt-3 border-t border-border/50 pt-3">
            <MoveTypeBreakdown
              :moves-by-type="row.movesByType"
              :moves-count="row.movesCount"
              compact
            />
          </div>

          <!-- Avg pts / move pill -->
          <div class="mt-3 self-end">
            <span
              class="rounded-full border border-border bg-muted px-2 py-0.5 text-[11px] font-medium text-foreground"
            >
              {{ fmt(row.avgPointsPerMove, 2) }} pts/move
            </span>
          </div>
        </RouterLink>

        <!-- Empty state -->
        <div
          v-if="filteredCountries.length === 0"
          class="col-span-full py-12 text-center text-sm text-muted-foreground"
        >
          No countries match your search.
        </div>
      </div>

      <p class="mt-2 text-right text-xs text-muted-foreground">
        {{ filteredCountries.length }} of {{ store.countries.length }} countries
      </p>
    </div>
  </main>
</template>
