<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import L from 'leaflet'
import type { MoveRecord } from '@/types/api'

interface Props {
  moves: MoveRecord[]
}

const props = defineProps<Props>()

const mapRef = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let markerLayer: L.LayerGroup | null = null
let polylineLayer: L.Polyline | null = null

function initMap(): void {
  if (!mapRef.value || map) return
  map = L.map(mapRef.value, {
    center: [30, 0],
    zoom: 2,
    scrollWheelZoom: true,
  })
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    maxZoom: 18,
  }).addTo(map)
  markerLayer = L.layerGroup().addTo(map)
}

function renderMoves(): void {
  if (!map || !markerLayer) return
  markerLayer.clearLayers()
  if (polylineLayer) {
    map.removeLayer(polylineLayer)
    polylineLayer = null
  }

  const geoMoves = props.moves.filter((m) => m.lat != null && m.lon != null)
  if (geoMoves.length === 0) return

  // Draw polyline connecting the path (chronological order — oldest first)
  const sortedMoves = [...geoMoves].sort((a, b) => new Date(a.movedOn).getTime() - new Date(b.movedOn).getTime())
  const latlngs: L.LatLngExpression[] = sortedMoves.map((m) => [m.lat!, m.lon!])
  polylineLayer = L.polyline(latlngs, {
    color: 'hsl(210, 80%, 55%)',
    weight: 2,
    opacity: 0.6,
    dashArray: '6 4',
  }).addTo(map)

  // Add markers
  for (let i = 0; i < sortedMoves.length; i++) {
    const m = sortedMoves[i]!
    const isFirst = i === 0
    const isLast = i === sortedMoves.length - 1

    const radius = isLast ? 8 : isFirst ? 7 : 5
    const color = isLast ? 'hsl(140, 70%, 45%)' : isFirst ? 'hsl(30, 90%, 55%)' : 'hsl(210, 80%, 55%)'

    const marker = L.circleMarker([m.lat!, m.lon!], {
      radius,
      fillColor: color,
      color: '#fff',
      weight: 2,
      fillOpacity: 0.9,
    }).addTo(markerLayer)

    const label = isFirst ? '🏠 Born' : isLast ? '📍 Current' : m.moveTypeName
    marker.bindTooltip(
      `<strong>${label}</strong><br/>${m.waypoint ?? ''} ${m.country ?? ''}<br/>${new Date(m.movedOn).toLocaleDateString()}${m.username ? `<br/>by ${m.username}` : ''}`,
      { className: 'leaflet-dark-tooltip', sticky: true },
    )
  }

  // Fit bounds to show all points
  const bounds = L.latLngBounds(latlngs)
  map.fitBounds(bounds, { padding: [30, 30], maxZoom: 12 })
}

watch(() => props.moves, renderMoves, { deep: true })

onMounted(() => {
  initMap()
  if (props.moves.length > 0) renderMoves()
})

onUnmounted(() => {
  map?.remove()
  map = null
})
</script>

<template>
  <div ref="mapRef" class="w-full h-[400px] rounded-lg border border-border overflow-hidden" />
</template>

<style scoped>
.leaflet-dark-tooltip {
  background: hsl(var(--popover));
  color: hsl(var(--popover-foreground));
  border: 1px solid hsl(var(--border));
  border-radius: 6px;
  padding: 6px 10px;
  font-size: 0.75rem;
  box-shadow: 0 2px 8px rgb(0 0 0 / 15%);
}
</style>
