<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import L from 'leaflet'
import LeafletMap from '@/components/LeafletMap.vue'
import { escapeHtml } from '@/lib/html'
import { MOVE_TYPES } from '@/constants/moveTypes'
import type { UserGeokretyPage } from '@/services/api/users'
import type { GeokretListItem } from '@/types/api'

type ColorMode = 'lastMove' | 'activity' | 'moveType'

const props = defineProps<{
  fetchFn: (limit: number, cursor?: string) => Promise<UserGeokretyPage>
}>()

const items = ref<GeokretListItem[]>([])
const loading = ref(false)
const colorMode = ref<ColorMode>('lastMove')

let map: L.Map | null = null
let markerLayer: L.LayerGroup | null = null

const geocodedItems = computed(() => items.value.filter((item) => item.lat != null && item.lon != null))

/** Items sorted oldest-first so newest are drawn on top by Leaflet's z-order */
const sortedItems = computed(() =>
  [...geocodedItems.value].sort((a, b) => {
    const ta = a.lastMoveAt ? new Date(a.lastMoveAt).getTime() : 0
    const tb = b.lastMoveAt ? new Date(b.lastMoveAt).getTime() : 0
    return ta - tb
  }),
)

const center = computed<[number, number]>(() => {
  const first = geocodedItems.value[0]
  return first ? [first.lat!, first.lon!] : [30, 0]
})

/** Move type entries for the "move type" legend — only types present in map items */
const presentMoveTypes = computed(() => {
  const ids = new Set(geocodedItems.value.map((it) => it.lastMoveType).filter((t): t is number => t != null))
  return MOVE_TYPES.filter((mt) => ids.has(mt.id))
})

function lerp(c0: [number, number, number], c1: [number, number, number], t: number): string {
  const r = Math.round(c0[0] + (c1[0] - c0[0]) * t)
  const g = Math.round(c0[1] + (c1[1] - c0[1]) * t)
  const b = Math.round(c0[2] + (c1[2] - c0[2]) * t)
  return `rgb(${r},${g},${b})`
}

/**
 * Last move date mode — indigo (old) → amber (recent), matching the geokrety map palette.
 */
function colorByLastMove(item: GeokretListItem): string {
  if (!item.lastMoveAt) return '#94a3b8'
  const ageDays = (Date.now() - new Date(item.lastMoveAt).getTime()) / 86_400_000
  const t = Math.min(ageDays / (365 * 3), 1) // 3 years = fully "old"
  // amber (recent) → indigo (old) — invert t so "small age = warm"
  const inv = 1 - t
  if (inv >= 0.67) return lerp([251, 191, 36], [34, 197, 94], (inv - 0.67) / 0.33)     // amber→emerald
  if (inv >= 0.33) return lerp([6, 182, 212], [34, 197, 94], (inv - 0.33) / 0.34)      // cyan→emerald
  return lerp([99, 102, 241], [6, 182, 212], inv / 0.33)                                // indigo→cyan
}

/**
 * Activity mode (by cachesCount) — indigo (0) → amber (many).
 */
function colorByActivity(item: GeokretListItem, maxCaches: number): string {
  if (maxCaches === 0) return '#94a3b8'
  const t = Math.min((item.cachesCount ?? 0) / maxCaches, 1)
  if (t >= 0.67) return lerp([34, 197, 94], [251, 191, 36], (t - 0.67) / 0.33)
  if (t >= 0.33) return lerp([6, 182, 212], [34, 197, 94], (t - 0.33) / 0.34)
  return lerp([99, 102, 241], [6, 182, 212], t / 0.33)
}

/** Last move type mode — use MOVE_TYPES hex colors */
function colorByMoveType(item: GeokretListItem): string {
  if (item.lastMoveType == null) return '#94a3b8'
  return MOVE_TYPES.find((mt) => mt.id === item.lastMoveType)?.colors.hex ?? '#94a3b8'
}

async function loadItems(): Promise<void> {
  loading.value = true
  try {
    const loaded: GeokretListItem[] = []
    let cursor: string | undefined
    let hasMore = true
    while (hasMore) {
      const page = await props.fetchFn(100, cursor)
      loaded.push(...page.data)
      cursor = page.nextCursor
      hasMore = page.hasMore && Boolean(cursor)
    }
    items.value = loaded
  } catch (e) {
    console.error('[UserOwnedGeokretyMap]', e)
  } finally {
    loading.value = false
  }
}

function render(instance: L.Map): void {
  map = instance
  if (!markerLayer) {
    markerLayer = L.layerGroup().addTo(instance)
  }
  markerLayer.clearLayers()
  if (sortedItems.value.length === 0) return

  const maxCaches = sortedItems.value.reduce((max, it) => Math.max(max, it.cachesCount ?? 0), 0)

  // sortedItems is oldest-first; last drawn = topmost in Leaflet
  sortedItems.value.forEach((item) => {
    const latLng: L.LatLngExpression = [item.lat!, item.lon!]
    let dotColor: string
    if (colorMode.value === 'activity') {
      dotColor = colorByActivity(item, maxCaches)
    } else if (colorMode.value === 'moveType') {
      dotColor = colorByMoveType(item)
    } else {
      dotColor = colorByLastMove(item)
    }

    const marker = L.circleMarker(latLng, {
      radius: 6,
      fillColor: dotColor,
      color: '#ffffff',
      weight: 1.5,
      fillOpacity: 0.9,
    })

    const lastMoveLabel = item.lastMoveAt
      ? new Date(item.lastMoveAt).toLocaleDateString()
      : 'never'
    const moveTypeName = item.lastMoveType != null
      ? (MOVE_TYPES.find((mt) => mt.id === item.lastMoveType)?.label ?? String(item.lastMoveType))
      : '—'
    marker.bindTooltip(
      `<strong>${escapeHtml(item.gkid ?? '')} ${escapeHtml(item.name)}</strong><br/>${escapeHtml(item.waypoint ?? '')} ${escapeHtml(item.country ?? '')}<br/>Last move: ${lastMoveLabel}<br/>Last type: ${escapeHtml(moveTypeName)}<br/>Caches: ${item.cachesCount ?? 0}`,
      { className: 'leaflet-dark-tooltip', sticky: true },
    )
    marker.addTo(markerLayer!)
  })

  instance.fitBounds(
    L.latLngBounds(sortedItems.value.map((it) => [it.lat!, it.lon!])),
    { padding: [30, 30], maxZoom: 8 },
  )
}

function rerenderMarkers(): void {
  if (map) render(map)
}

watch(sortedItems, () => { if (map) render(map) }, { deep: true })
watch(colorMode, rerenderMarkers)

onMounted(() => { loadItems() })
</script>

<template>
  <section class="space-y-3">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h2 class="text-lg font-semibold">Owned GeoKrety map</h2>
      <div class="flex items-center gap-1.5">
        <p class="text-xs text-muted-foreground mr-2">
          {{ geocodedItems.length }} geocoded of {{ items.length }}
        </p>
        <!-- Mode selector -->
        <button
          v-for="mode in (['lastMove', 'activity', 'moveType'] as const)"
          :key="mode"
          :class="['rounded-full border px-3 py-1 text-xs font-medium transition-colors',
            colorMode === mode
              ? 'border-border bg-accent text-accent-foreground'
              : 'border-border bg-card text-muted-foreground hover:text-foreground']"
          @click="colorMode = mode"
        >
          {{ mode === 'lastMove' ? 'Last move date' : mode === 'activity' ? 'Activity' : 'Move type' }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="flex justify-center py-6">
      <div class="h-6 w-6 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
    </div>
    <p v-else-if="geocodedItems.length === 0" class="text-sm text-muted-foreground">
      No owned GeoKrety with known coordinates.
    </p>

    <div v-else class="relative">
      <LeafletMap
        :center="center"
        :zoom="4"
        :min-zoom="2"
        :max-zoom="12"
        :require-two-finger-pan-on-mobile="true"
        @ready="render"
      />

      <!-- Legend overlay -->
      <div
        class="absolute bottom-9 left-3 z-[400] rounded-lg border border-border bg-card/90 px-3 py-2 text-xs text-muted-foreground pointer-events-none"
      >
        <template v-if="colorMode === 'lastMove'">
          <p class="mb-1 font-medium text-foreground">Last move date</p>
          <div class="flex items-center gap-2">
            <span>Recent</span>
            <div class="h-3 w-20 rounded-sm" style="background: linear-gradient(to right, #fbbf24, #22c55e, #06b6d4, #6366f1)" />
            <span>Old / none</span>
          </div>
        </template>
        <template v-else-if="colorMode === 'activity'">
          <p class="mb-1 font-medium text-foreground">Activity (caches visited)</p>
          <div class="flex items-center gap-2">
            <span>None</span>
            <div class="h-3 w-20 rounded-sm" style="background: linear-gradient(to right, #6366f1, #06b6d4, #22c55e, #fbbf24)" />
            <span>Many</span>
          </div>
        </template>
        <template v-else>
          <p class="mb-1 font-medium text-foreground">Last move type</p>
          <div v-for="mt in presentMoveTypes" :key="mt.id" class="flex items-center gap-1.5 leading-5">
            <span class="inline-block h-3 w-3 rounded-full border border-white/60" :style="{ background: mt.colors.hex }" />
            {{ mt.label }}
          </div>
          <span v-if="presentMoveTypes.length === 0" class="italic">No data</span>
        </template>
      </div>
    </div>
  </section>
</template>

