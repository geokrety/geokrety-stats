<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { escapeHtml } from '@/lib/html'

// ── Country center lookup (lat, lon, zoom) ────────────────────────────────
const COUNTRY_CENTERS: Record<string, [number, number, number]> = {
  PL: [52.0, 19.0, 6],
  DE: [51.2, 10.5, 6],
  FR: [46.5, 2.5, 6],
  CZ: [49.8, 15.5, 7],
  SK: [48.7, 19.5, 7],
  RU: [61.0, 60.0, 4],
  UA: [49.0, 32.0, 6],
  AT: [47.5, 14.5, 7],
  HU: [47.2, 19.2, 7],
  CH: [46.8, 8.2, 7],
  IT: [41.9, 12.5, 6],
  ES: [40.4, -3.7, 6],
  PT: [39.4, -8.2, 7],
  NL: [52.3, 5.3, 7],
  BE: [50.8, 4.5, 8],
  GB: [54.0, -2.1, 6],
  IE: [53.2, -8.2, 7],
  SE: [62.0, 16.0, 5],
  NO: [64.0, 15.0, 5],
  DK: [56.0, 10.0, 7],
  FI: [64.0, 27.0, 5],
  US: [39.5, -98.4, 4],
  CA: [60.0, -95.0, 4],
  BR: [-15.0, -53.0, 4],
  AU: [-27.0, 133.0, 4],
  JP: [36.0, 138.0, 5],
  CN: [35.9, 104.3, 4],
  IN: [22.0, 79.0, 5],
  ZA: [-30.0, 25.0, 5],
  FJ: [-17.7, 178.1, 7],
  DEFAULT: [20.0, 10.0, 3],
}

export interface GeoKretMarker {
  id: number
  name: string
  lat: number
  lon: number
  status: 'cache' | 'lost'
  lastActivity: Date
  gkid: string
}

interface Props {
  countryCode: string
  inCacheCount: number
  lostCount: number
}

const props = defineProps<Props>()

// ── UI state ─────────────────────────────────────────────────────────────
type Filter = 'all' | 'cache' | 'lost'
const filter = ref<Filter>('all')
const monthsBack = ref(24)

// ── Map refs ──────────────────────────────────────────────────────────────
const mapRef = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let markerLayer: L.LayerGroup | null = null

// ── Pseudo-random seeded generator ────────────────────────────────────────
function seededRand(seed: number) {
  let s = seed
  return () => {
    s = (s * 1664525 + 1013904223) & 0xffffffff
    return Math.abs(s) / 0xffffffff
  }
}

// ── Generate mock GeoKret positions ──────────────────────────────────────
const markers = computed<GeoKretMarker[]>(() => {
  const center = COUNTRY_CENTERS[props.countryCode] ?? COUNTRY_CENTERS['DEFAULT']
  const [clat, clon] = center as [number, number, number]
  const rand = seededRand(props.countryCode.charCodeAt(0) * 31 + props.countryCode.charCodeAt(1))
  const now = Date.now()
  const out: GeoKretMarker[] = []
  const totalShown = Math.min(props.inCacheCount + props.lostCount, 80) // cap for performance
  const cacheCount = Math.round(
    (totalShown * props.inCacheCount) / (props.inCacheCount + props.lostCount || 1),
  )
  const lostCount = totalShown - cacheCount

  const push = (status: 'cache' | 'lost', i: number) => {
    const lat = clat + (rand() - 0.5) * 8
    const lon = clon + (rand() - 0.5) * 10
    const daysBack = Math.floor(rand() * 730) // up to 2 years
    out.push({
      id: i,
      name: `GeoKret #${Math.floor(rand() * 90000 + 10000)}`,
      lat,
      lon,
      status,
      lastActivity: new Date(now - daysBack * 86_400_000),
      gkid: `GK${Math.floor(rand() * 0xffff)
        .toString(16)
        .toUpperCase()
        .padStart(4, '0')}`,
    })
  }

  for (let i = 0; i < cacheCount; i++) push('cache', i)
  for (let i = 0; i < lostCount; i++) push('cache_count' as 'lost', cacheCount + i)
  // Assign lost status
  out.slice(cacheCount).forEach((m) => {
    m.status = 'lost'
  })
  return out
})

// ── Filtered markers (by filter + time slider) ─────────────────────────
const visibleMarkers = computed(() => {
  const cutoff = new Date(Date.now() - monthsBack.value * 30 * 86_400_000)
  return markers.value.filter((m) => {
    if (m.lastActivity < cutoff) return false
    if (filter.value === 'cache') return m.status === 'cache'
    if (filter.value === 'lost') return m.status === 'lost'
    return true
  })
})

// ── Marker icons ──────────────────────────────────────────────────────────
const ICON_CACHE = L.divIcon({
  html: '<span class="leaflet-marker-dot leaflet-marker-dot--cache"></span>',
  iconSize: [10, 10],
  iconAnchor: [5, 5],
  className: '',
})
const ICON_LOST = L.divIcon({
  html: '<span class="leaflet-marker-dot leaflet-marker-dot--lost"></span>',
  iconSize: [10, 10],
  iconAnchor: [5, 5],
  className: '',
})

function relativeDate(d: Date): string {
  const diff = Math.floor((Date.now() - d.getTime()) / 86_400_000)
  if (diff < 30) return `${diff}d ago`
  if (diff < 365) return `${Math.floor(diff / 30)}mo ago`
  return `${Math.floor(diff / 365)}y ago`
}

function renderMarkers() {
  if (!map) return
  markerLayer?.clearLayers()
  visibleMarkers.value.forEach((m) => {
    const icon = m.status === 'cache' ? ICON_CACHE : ICON_LOST
    const marker = L.marker([m.lat, m.lon], { icon })
    marker.bindTooltip(
      `<div class="leaflet-tooltip-panel">
        <strong>${escapeHtml(m.gkid)}</strong> — ${escapeHtml(m.name)}<br/>
        <span class="leaflet-tooltip-meta">${m.status === 'lost' ? '🔴 Missing' : '🟢 In cache'} · Last: ${escapeHtml(relativeDate(m.lastActivity))}</span>
       </div>`,
      { className: 'leaflet-dark-tooltip', sticky: true },
    )
    markerLayer?.addLayer(marker)
  })
}

function initMap() {
  if (!mapRef.value || map) return
  const center = COUNTRY_CENTERS[props.countryCode] ?? COUNTRY_CENTERS['DEFAULT']
  const [clat, clon, zoom] = center as [number, number, number]

  map = L.map(mapRef.value, {
    center: [clat, clon],
    zoom,
    minZoom: 2,
    maxZoom: 10,
    zoomControl: true,
    attributionControl: true,
  })

  L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_nolabels/{z}/{x}/{y}{r}.png', {
    attribution:
      '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>',
    subdomains: 'abcd',
    maxZoom: 20,
  }).addTo(map)

  markerLayer = L.layerGroup().addTo(map)
  renderMarkers()
}

onMounted(() => setTimeout(initMap, 50))
onUnmounted(() => {
  map?.remove()
  map = null
})

watch(visibleMarkers, () => renderMarkers())

const visibleCount = computed(() => visibleMarkers.value.length)
const cachedCount = computed(() => visibleMarkers.value.filter((m) => m.status === 'cache').length)
const lostCount = computed(() => visibleMarkers.value.filter((m) => m.status === 'lost').length)
</script>

<template>
  <section class="rounded-xl bg-card/70 border border-border p-5">
    <h2 class="text-xs font-semibold uppercase tracking-widest text-muted-foreground mb-4">
      GeoKrety map
    </h2>

    <!-- Controls -->
    <div class="flex flex-wrap items-center gap-3 mb-4">
      <!-- Status filter -->
      <div class="flex rounded-lg border border-border overflow-hidden text-xs">
        <button
          v-for="opt in [
            { key: 'all', label: 'All' },
            { key: 'cache', label: '🟢 In cache' },
            { key: 'lost', label: '🔴 Missing' },
          ] as const"
          :key="opt.key"
          :class="[
            'px-3 py-1.5 font-medium transition-colors border-r border-border last:border-r-0',
            filter === opt.key
              ? 'bg-accent text-accent-foreground'
              : 'text-muted-foreground hover:text-foreground hover:bg-card/5',
          ]"
          @click="filter = opt.key"
        >
          {{ opt.label }}
        </button>
      </div>

      <!-- Time range slider -->
      <div class="flex items-center gap-2 flex-1 min-w-[180px]">
        <span class="text-xs text-muted-foreground whitespace-nowrap">Last activity:</span>
        <input
          v-model.number="monthsBack"
          type="range"
          min="1"
          max="24"
          step="1"
          class="flex-1 h-1.5 rounded-full appearance-none bg-muted accent-foreground cursor-pointer"
        />
        <span class="text-xs text-foreground whitespace-nowrap min-w-[3rem] text-right">
          {{ monthsBack >= 24 ? 'All time' : `≤ ${monthsBack}mo` }}
        </span>
      </div>
    </div>

    <!-- Stats row -->
    <div class="flex gap-4 text-xs text-muted-foreground mb-3">
      <span
        >Showing <strong class="text-foreground">{{ visibleCount }}</strong> GeoKrety</span
      >
      <span
        >🟢 <strong class="text-foreground">{{ cachedCount }}</strong> in cache</span
      >
      <span
        >🔴 <strong class="text-foreground">{{ lostCount }}</strong> missing</span
      >
    </div>

    <!-- Map -->
    <div
      ref="mapRef"
      class="h-72 w-full rounded-lg overflow-hidden border border-border"
      aria-label="Country GeoKrety map"
    />
  </section>
</template>
