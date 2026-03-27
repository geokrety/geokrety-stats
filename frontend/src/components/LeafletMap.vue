<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

const props = withDefaults(defineProps<{
  center?: [number, number]
  zoom?: number
  minZoom?: number
  maxZoom?: number
  scrollWheelZoom?: boolean
  requireTwoFingerPanOnMobile?: boolean
  height?: string
}>(), {
  center: () => [30, 0],
  zoom: 2,
  minZoom: 2,
  maxZoom: 18,
  scrollWheelZoom: true,
  requireTwoFingerPanOnMobile: false,
  height: '400px',
})

const emit = defineEmits<{
  ready: [map: L.Map]
}>()

const mapRef = ref<HTMLElement | null>(null)
const showTouchHint = ref(false)
const isTouchDevice = computed(() => typeof window !== 'undefined' && window.matchMedia('(pointer: coarse)').matches)
let map: L.Map | null = null
let touchHintTimer: number | null = null
let resizeObserver: ResizeObserver | null = null

function disableSingleFingerPan(): void {
  if (!map || !props.requireTwoFingerPanOnMobile || !isTouchDevice.value) return
  map.dragging.disable()
  map.touchZoom.disable()
}

function enableTouchInteraction(): void {
  if (!map) return
  map.dragging.enable()
  map.touchZoom.enable()
}

function showTwoFingerHint(): void {
  showTouchHint.value = true
  if (touchHintTimer) window.clearTimeout(touchHintTimer)
  touchHintTimer = window.setTimeout(() => {
    showTouchHint.value = false
  }, 1600)
}

function syncTouchInteraction(touchCount: number): void {
  if (!props.requireTwoFingerPanOnMobile || !isTouchDevice.value) return
  if (touchCount >= 2) {
    showTouchHint.value = false
    enableTouchInteraction()
    return
  }
  disableSingleFingerPan()
  showTwoFingerHint()
}

function handleTouchStart(event: TouchEvent): void {
  syncTouchInteraction(event.touches.length)
}

function handleTouchMove(event: TouchEvent): void {
  syncTouchInteraction(event.touches.length)
}

function handleTouchEnd(event: TouchEvent): void {
  if (!props.requireTwoFingerPanOnMobile || !isTouchDevice.value) return
  if (event.touches.length >= 2) {
    enableTouchInteraction()
    return
  }
  disableSingleFingerPan()
}

onMounted(() => {
  if (!mapRef.value) return
  map = L.map(mapRef.value, {
    center: props.center,
    zoom: props.zoom,
    minZoom: props.minZoom,
    maxZoom: props.maxZoom,
    scrollWheelZoom: props.scrollWheelZoom,
  })

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    maxZoom: 19,
  }).addTo(map)

  if (props.requireTwoFingerPanOnMobile && isTouchDevice.value) {
    disableSingleFingerPan()
    mapRef.value.addEventListener('touchstart', handleTouchStart, { passive: true })
    mapRef.value.addEventListener('touchmove', handleTouchMove, { passive: true })
    mapRef.value.addEventListener('touchend', handleTouchEnd, { passive: true })
    mapRef.value.addEventListener('touchcancel', handleTouchEnd, { passive: true })
  }

  map.whenReady(() => {
    emit('ready', map as L.Map)
    window.setTimeout(() => map?.invalidateSize(), 0)
  })

  if (typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => map?.invalidateSize())
    resizeObserver.observe(mapRef.value)
  }
})

onUnmounted(() => {
  if (mapRef.value && props.requireTwoFingerPanOnMobile && isTouchDevice.value) {
    mapRef.value.removeEventListener('touchstart', handleTouchStart)
    mapRef.value.removeEventListener('touchmove', handleTouchMove)
    mapRef.value.removeEventListener('touchend', handleTouchEnd)
    mapRef.value.removeEventListener('touchcancel', handleTouchEnd)
  }
  if (touchHintTimer) window.clearTimeout(touchHintTimer)
  resizeObserver?.disconnect()
  resizeObserver = null
  map?.remove()
  map = null
})
</script>

<template>
  <div class="leaflet-map-shell relative overflow-hidden rounded-lg border border-border">
    <div ref="mapRef" class="leaflet-map-canvas w-full" />
    <div
      v-if="showTouchHint"
      class="pointer-events-none absolute inset-x-4 top-4 rounded-md bg-background/90 px-3 py-2 text-center text-xs text-muted-foreground shadow-sm backdrop-blur"
    >
      Use two fingers to move the map
    </div>
  </div>
</template>

<style scoped>
.leaflet-map-shell {
  min-height: v-bind(height);
}

.leaflet-map-canvas {
  height: v-bind(height);
  min-height: v-bind(height);
}

:deep(.leaflet-container) {
  height: 100%;
  width: 100%;
}
</style>
