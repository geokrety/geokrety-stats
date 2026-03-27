<script setup lang="ts">
import { watch } from 'vue'
import L from 'leaflet'
import LeafletMap from '@/components/LeafletMap.vue'
import { escapeHtml } from '@/lib/html'
import type { MoveRecord } from '@/types/api'

interface Props {
  moves: MoveRecord[]
}

const props = defineProps<Props>()

let map: L.Map | null = null
let markerLayer: L.LayerGroup | null = null
let polylineLayer: L.Polyline | null = null

function handleMapReady(instance: L.Map): void {
  map = instance
  markerLayer = L.layerGroup().addTo(map)
  renderMoves()
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
      `<strong>${escapeHtml(label)}</strong><br/>${escapeHtml(m.waypoint)} ${escapeHtml(m.country)}<br/>${new Date(m.movedOn).toLocaleDateString()}${m.username ? `<br/>by ${escapeHtml(m.username)}` : ''}`,
      { className: 'leaflet-dark-tooltip', sticky: true },
    )
  }

  // Fit bounds to show all points
  const bounds = L.latLngBounds(latlngs)
  map.fitBounds(bounds, { padding: [30, 30], maxZoom: 12 })
}

watch(() => props.moves, renderMoves, { deep: true })
</script>

<template>
  <LeafletMap
    :require-two-finger-pan-on-mobile="true"
    height="400px"
    @ready="handleMapReady"
  />
</template>
