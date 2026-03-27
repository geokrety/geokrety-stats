<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import L from 'leaflet'
import LeafletMap from '@/components/LeafletMap.vue'
import { escapeHtml } from '@/lib/html'
import type { UserGeokretyPage } from '@/services/api/users'
import type { GeokretListItem } from '@/types/api'

const props = defineProps<{
  fetchFn: (limit: number, cursor?: string) => Promise<UserGeokretyPage>
}>()

const items = ref<GeokretListItem[]>([])
const loading = ref(false)
let map: L.Map | null = null
let markerLayer: L.LayerGroup | null = null

const geocodedItems = computed(() => items.value.filter((item) => item.lat != null && item.lon != null))
const center = computed<[number, number]>(() => {
  const first = geocodedItems.value[0]
  return first ? [first.lat!, first.lon!] : [30, 0]
})

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

  if (geocodedItems.value.length === 0) return

  const latLngs: L.LatLngExpression[] = []
  geocodedItems.value.forEach((item) => {
    const latLng: L.LatLngExpression = [item.lat!, item.lon!]
    latLngs.push(latLng)
    const marker = L.circleMarker(latLng, {
      radius: 6,
      fillColor: '#10b981',
      color: '#ffffff',
      weight: 2,
      fillOpacity: 0.9,
    })
    marker.bindTooltip(
      `<strong>${escapeHtml(item.gkid)}</strong> ${escapeHtml(item.name)}<br/>${escapeHtml(item.waypoint)} ${escapeHtml(item.country)}`,
      { className: 'leaflet-dark-tooltip', sticky: true },
    )
    marker.addTo(markerLayer!)
  })

  instance.fitBounds(L.latLngBounds(latLngs), { padding: [30, 30], maxZoom: 8 })
}

watch(geocodedItems, () => {
  if (map) render(map)
}, { deep: true })

onMounted(() => {
  loadItems()
})
</script>

<template>
  <section class="space-y-3">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold">Owned GeoKrety map</h2>
      <p class="text-xs text-muted-foreground">
        {{ geocodedItems.length }} geocoded of {{ items.length }} owned
      </p>
    </div>
    <div v-if="loading" class="flex justify-center py-6">
      <div class="h-6 w-6 animate-spin rounded-full border-2 border-border/20 border-t-primary" />
    </div>
    <p v-else-if="geocodedItems.length === 0" class="text-sm text-muted-foreground">
      No owned GeoKrety with known coordinates.
    </p>
    <LeafletMap
      v-else
      :center="center"
      :zoom="4"
      :min-zoom="2"
      :max-zoom="12"
      :require-two-finger-pan-on-mobile="true"
      @ready="render"
    />
  </section>
</template>
