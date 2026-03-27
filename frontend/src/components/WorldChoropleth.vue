<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import L from 'leaflet'
import * as topojson from 'topojson-client'
import type { Topology, GeometryCollection } from 'topojson-specification'
import { countryCodeToFlag } from '@/lib/countryFlag'
// Using ?url avoids bundling the large JSON as an ES module
import worldAtlasUrl from 'world-atlas/countries-110m.json?url'
import { scaleSequentialLog } from 'd3'
import { interpolateGreens } from 'd3'
import { useRouter } from 'vue-router'
import type { CountryStats } from '@/types/api'
import { numericToAlpha2 } from '@/data/iso3166'

// ── Props ────────────────────────────────────────────────────────────────────
interface Props {
  countries: CountryStats[]
  /**
   * Which numeric field to use for the choropleth gradient.
   * Non-numeric or nested fields are ignored (map stays grey).
   */
  metric: keyof CountryStats
}
const props = defineProps<Props>()

// ── Internals ────────────────────────────────────────────────────────────────
const mapRef = ref<HTMLElement | null>(null)
const router = useRouter()

/** World TopoJSON data — loaded at runtime to avoid Vite static-import issues */
let worldDataCache: unknown = null

async function loadWorldData(): Promise<unknown> {
  if (worldDataCache) return worldDataCache
  const res = await fetch(worldAtlasUrl)
  if (!res.ok) throw new Error(`Failed to load world-atlas: ${res.status}`)
  worldDataCache = await res.json()
  return worldDataCache
}

let map: L.Map | null = null
let geoLayer: L.GeoJSON | null = null

type GeoFeature = GeoJSON.Feature<GeoJSON.Geometry, Record<string, unknown>>

/** Build a code → stats lookup */
function buildLookup(countries: CountryStats[]) {
  return new Map<string, CountryStats>(countries.map((c) => [c.code, c]))
}

/** Return the numeric value of `metric` for a country (or −1 if unavailable) */
function metricValue(stats: CountryStats | undefined, metric: keyof CountryStats): number {
  if (!stats) return -1
  const raw = stats[metric]
  if (typeof raw === 'number') return raw
  // nested (movesByType) — not directly colourable; return -1
  return -1
}

/** Build a D3 scaleSequential for the current dataset + metric */
function buildScale(countries: CountryStats[], metric: keyof CountryStats) {
  const values = countries.map((c) => metricValue(c, metric)).filter((v) => v > 0)
  if (values.length === 0) return null
  const max = Math.max(...values)
  // Log scale so small values are still distinguishable from zero
  return scaleSequentialLog<string>().domain([1, max]).interpolator(interpolateGreens).clamp(true)
}

function cssHsl(tokenName: string, fallback: string): string {
  const token = getComputedStyle(document.documentElement).getPropertyValue(tokenName).trim()
  return token ? `hsl(${token})` : fallback
}

function mapPalette() {
  return {
    fillZero: cssHsl('--muted', '#374151'),
    fillNone: cssHsl('--card', '#1f2937'),
    fillHover: cssHsl('--primary', '#6ee7b7'),
    stroke: cssHsl('--border', '#4b5563'),
    fillDefault: cssHsl('--primary', '#10b981'),
  }
}

function getStyle(
  feature: GeoFeature | undefined,
  lookup: Map<string, CountryStats>,
  scale: ReturnType<typeof buildScale>,
  metric: keyof CountryStats,
  palette: ReturnType<typeof mapPalette>,
): L.PathOptions {
  const id = String(feature?.id ?? '')
  const alpha2 = numericToAlpha2(id)
  const stats = alpha2 ? lookup.get(alpha2) : undefined
  const value = metricValue(stats, metric)

  let fillColor: string
  if (value < 0) fillColor = palette.fillNone
  else if (value === 0) fillColor = palette.fillZero
  else fillColor = scale ? scale(value) : palette.fillDefault

  return {
    fillColor,
    weight: 0.5,
    opacity: 1,
    color: palette.stroke,
    fillOpacity: 0.9,
  }
}

/** Convert numeric ISO id → alpha2, then look up full name */
function countryLabel(feature: GeoFeature, lookup: Map<string, CountryStats>): string {
  const alpha2 = numericToAlpha2(String(feature.id ?? ''))
  const stats = alpha2 ? lookup.get(alpha2) : undefined
  return stats ? `${countryCodeToFlag(stats.code)} ${stats.name}` : (alpha2 ?? 'Unknown')
}

/**
 * Normalize all GeoJSON polygon coordinates to the [-180, 180] longitude
 * range **in-place**.  This prevents Leaflet from drawing a horizontal line
 * across the map for countries that straddle the antimeridian (Russia, Fiji,
 * Kiribati, etc.) because topojson-client sometimes produces coordinates
 * outside that range when converting from the source TopoJSON topology.
 */
function normalizeAntimeridian(fc: GeoJSON.FeatureCollection) {
  const normalizeRing = (ring: number[][]) => {
    ring.forEach((coord) => {
      // coord[0] is longitude — reindex into [-180, 180]
      if (coord[0] == null) return
      while (coord[0] > 180) coord[0] -= 360
      while (coord[0] < -180) coord[0] += 360
    })
  }
  fc.features.forEach((f) => {
    if (!f.geometry) return
    if (f.geometry.type === 'Polygon') {
      f.geometry.coordinates.forEach(normalizeRing)
    } else if (f.geometry.type === 'MultiPolygon') {
      f.geometry.coordinates.forEach((poly) => poly.forEach(normalizeRing))
    }
  })
}

function initMap() {
  if (!mapRef.value || map) return

  map = L.map(mapRef.value, {
    center: [20, 10],
    zoom: 2,
    minZoom: 1,
    maxZoom: 6,
    zoomControl: true,
    attributionControl: true,
    // Wrap the world longitude
    worldCopyJump: true,
  })

  // Dark tile layer (CartoDB Dark Matter, no labels for clean look)
  L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_nolabels/{z}/{x}/{y}{r}.png', {
    attribution:
      '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
    subdomains: 'abcd',
    maxZoom: 20,
  }).addTo(map)

  renderLayer()
}

function renderLayer() {
  if (!map) return
  // Load world data asynchronously then draw
  loadWorldData()
    .then((rawData) => {
      _renderLayerWithData(rawData)
    })
    .catch(console.error)
}

function _renderLayerWithData(rawData: unknown) {
  if (!map) return
  if (geoLayer) {
    geoLayer.remove()
    geoLayer = null
  }

  const lookup = buildLookup(props.countries)
  const scale = buildScale(props.countries, props.metric)
  const palette = mapPalette()

  // Convert topojson → GeoJSON
  const topology = rawData as unknown as Topology<{
    countries: GeometryCollection
  }>
  const worldGeo = topojson.feature(topology, topology.objects.countries)

  // ── Antimeridian fix ──────────────────────────────────────────────────
  // Normalize all polygon coordinates to [-180, 180] so Leaflet doesn't
  // draw horizontal lines across Russia, Fiji, and other countries that
  // span the 180° meridian in the source topology.
  normalizeAntimeridian(worldGeo as GeoJSON.FeatureCollection)

  geoLayer = L.geoJSON(worldGeo as GeoJSON.GeoJsonObject, {
    style: (feature) => getStyle(feature as GeoFeature, lookup, scale, props.metric, palette),
    onEachFeature: (feature, layer) => {
      const gf = feature as GeoFeature
      const alpha2 = numericToAlpha2(String(gf.id ?? ''))
      const stats = alpha2 ? lookup.get(alpha2) : undefined
      const label = countryLabel(gf, lookup)
      const value = metricValue(stats, props.metric)
      const valueStr = value >= 0 ? value.toLocaleString() : 'No data'

      layer.bindTooltip(
        `<div class="map-tooltip">
          <span class="font-semibold">${label}</span><br/>
          <span class="text-xs opacity-75">${valueStr}</span>
         </div>`,
        { className: 'leaflet-dark-tooltip', sticky: true },
      )
      ;(layer as L.Path).on({
        mouseover(e) {
          const l = e.target as L.Path
          l.setStyle({ weight: 1.5, color: palette.fillHover, fillOpacity: 1 })
          l.bringToFront()
        },
        mouseout() {
          geoLayer?.resetStyle(layer as L.Path)
        },
        click() {
          if (alpha2) {
            router.push(`/countries/${alpha2.toLowerCase()}`)
          }
        },
      })
    },
  }).addTo(map)
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────
onMounted(() => {
  // Small delay to ensure the container is rendered and has a size
  setTimeout(initMap, 50)
})

onUnmounted(() => {
  map?.remove()
  map = null
  geoLayer = null
})

// Re-render when countries data or metric changes
watch(
  () => [props.countries, props.metric] as const,
  () => renderLayer(),
  { deep: false },
)
</script>

<template>
  <div
    ref="mapRef"
    class="world-choropleth h-full w-full rounded-xl overflow-hidden"
    aria-label="World choropleth map — countries coloured by selected metric"
  />
</template>

<style>
.world-choropleth .leaflet-container {
  background-color: hsl(var(--card));
}

.world-choropleth .leaflet-control-attribution {
  background: color-mix(in oklab, hsl(var(--card)) 70%, transparent) !important;
  color: hsl(var(--muted-foreground));
  font-size: 0.6rem;
}

.world-choropleth .leaflet-control-attribution a {
  color: hsl(var(--foreground));
}

.world-choropleth .leaflet-bar a {
  background: hsl(var(--card));
  color: hsl(var(--foreground));
  border-color: hsl(var(--border));
}

.world-choropleth .leaflet-bar a:hover {
  background: hsl(var(--accent));
  color: hsl(var(--accent-foreground));
}
</style>
