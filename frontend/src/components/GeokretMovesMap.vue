<script setup lang="ts">
import { watch, computed } from 'vue'
import L from 'leaflet'
import LeafletMap from '@/components/LeafletMap.vue'
import { escapeHtml } from '@/lib/html'
import { MOVE_TYPES } from '@/constants/moveTypes'
import type { MoveRecord } from '@/types/api'

interface Props {
  moves: MoveRecord[]
}

const props = defineProps<Props>()

let map: L.Map | null = null
let markerLayer: L.LayerGroup | null = null
let segmentLayer: L.LayerGroup | null = null

/**
 * Move types actually present in geo-located moves, used to filter the legend.
 */
const presentMoveTypes = computed(() => {
  const ids = new Set(
    props.moves.filter((m) => m.lat != null && m.lon != null).map((m) => m.moveType),
  )
  return MOVE_TYPES.filter((mt) => ids.has(mt.id))
})

function moveTimestamp(value: string | Date | null | undefined): number {
  if (!value) return 0
  return new Date(value).getTime()
}

/**
 * Gradient palette: indigo → sky → emerald → amber (oldest → newest).
 * Viridis-inspired: cool hues for old moves, warm for recent.
 */
function gradientColor(t: number): string {
  const stops: [number, [number, number, number]][] = [
    [0,    [99,  102, 241]],  // indigo  (#6366f1)
    [0.33, [6,   182, 212]],  // cyan    (#06b6d4)
    [0.67, [34,  197, 94]],   // emerald (#22c55e)
    [1,    [251, 191, 36]],   // amber   (#fbbf24)
  ]
  let r = 0, g = 0, b = 0
  for (let i = 0; i < stops.length - 1; i++) {
    const [t0, c0] = stops[i]!
    const [t1, c1] = stops[i + 1]!
    if (t >= t0 && t <= t1) {
      const f = (t - t0) / (t1 - t0)
      r = Math.round(c0[0] + (c1[0] - c0[0]) * f)
      g = Math.round(c0[1] + (c1[1] - c0[1]) * f)
      b = Math.round(c0[2] + (c1[2] - c0[2]) * f)
      break
    }
  }
  return `rgb(${r},${g},${b})`
}

/** Return the leaflet hex color for a given move type id */
function moveTypeHex(moveType: number): string {
  return MOVE_TYPES.find((mt) => mt.id === moveType)?.colors.hex ?? '#94a3b8'
}

function handleMapReady(instance: L.Map): void {
  map = instance
  markerLayer = L.layerGroup().addTo(map)
  segmentLayer = L.layerGroup().addTo(map)
  renderMoves()
}

function renderMoves(): void {
  if (!map || !markerLayer || !segmentLayer) return
  markerLayer.clearLayers()
  segmentLayer.clearLayers()

  const geoMoves = props.moves.filter((m) => m.lat != null && m.lon != null)
  if (geoMoves.length === 0) return

  // Sort chronologically (oldest first)
  const sortedMoves = [...geoMoves].sort((a, b) => {
    const movedOnDiff = moveTimestamp(a.movedOn) - moveTimestamp(b.movedOn)
    if (movedOnDiff !== 0) return movedOnDiff
    const createdOnDiff = moveTimestamp(a.createdOn) - moveTimestamp(b.createdOn)
    if (createdOnDiff !== 0) return createdOnDiff
    return a.id - b.id
  })

  const n = sortedMoves.length

  // Draw gradient trajectory: each segment gets a color from the gradient
  for (let i = 0; i < n - 1; i++) {
    const a = sortedMoves[i]!
    const b = sortedMoves[i + 1]!
    // Use midpoint t for segment color
    const t = n > 1 ? (i + 0.5) / (n - 1) : 0
    L.polyline([[a.lat!, a.lon!], [b.lat!, b.lon!]], {
      color: gradientColor(t),
      weight: 3,
      opacity: 0.75,
    }).addTo(segmentLayer!)
  }

  // Add markers colored by move type
  for (let i = 0; i < n; i++) {
    const m = sortedMoves[i]!
    const isFirst = i === 0
    const isLast = i === n - 1

    const dotColor = moveTypeHex(m.moveType)
    const radius = isLast ? 9 : isFirst ? 8 : 6
    const strokeColor = '#000'

    const marker = L.circleMarker([m.lat!, m.lon!], {
      radius,
      fillColor: dotColor,
      color: strokeColor,
      weight: isFirst || isLast ? 2.5 : 1.5,
      fillOpacity: m.moveType === 5 ? 0.75 : 0.95,
    }).addTo(markerLayer!)

    const moveLabel = isFirst ? '🏠 Born' : isLast ? '📍 Current' : (m.moveTypeName || '')
    marker.bindTooltip(
      `<strong>${escapeHtml(moveLabel)}</strong><br/>${escapeHtml(m.waypoint)} ${escapeHtml(m.country)}<br/>${new Date(m.movedOn).toLocaleDateString()}${m.username ? `<br/>by ${escapeHtml(m.username)}` : ''}`,
      { className: 'leaflet-dark-tooltip', sticky: true },
    )
  }

  // Fit bounds
  const latlngs: L.LatLngExpression[] = sortedMoves.map((m) => [m.lat!, m.lon!])
  map.fitBounds(L.latLngBounds(latlngs), { padding: [30, 30], maxZoom: 12 })
}

watch(() => props.moves, renderMoves, { deep: true })
</script>

<template>
  <div class="relative">
    <LeafletMap
      :require-two-finger-pan-on-mobile="true"
      height="400px"
      @ready="handleMapReady"
    />
    <!-- Path gradient legend -->
    <div
      v-if="moves.filter(m => m.lat != null && m.lon != null).length > 1"
      class="absolute bottom-3 left-3 z-[400] flex items-center gap-2 rounded-lg border border-border bg-card/90 px-3 py-2 text-xs text-muted-foreground pointer-events-none"
    >
      <span>Oldest</span>
      <div class="h-3 w-20 rounded-sm" style="background: linear-gradient(to right, #6366f1, #06b6d4, #22c55e, #fbbf24)" />
      <span>Newest</span>
    </div>
    <!-- Move type legend (only types present in this geokrety's history) -->
    <div
      v-if="presentMoveTypes.length > 0"
      class="absolute bottom-3 right-3 z-[400] rounded-lg border border-border bg-card/90 px-3 py-2 text-xs text-muted-foreground pointer-events-none"
    >
      <p class="mb-1 font-medium text-foreground">Move types</p>
      <div v-for="mt in presentMoveTypes" :key="mt.id" class="flex items-center gap-1.5 leading-5">
        <span class="inline-block h-3 w-3 rounded-full border border-white/60" :style="{ background: mt.colors.hex }" />
        {{ mt.label }}
      </div>
    </div>
  </div>
</template>
