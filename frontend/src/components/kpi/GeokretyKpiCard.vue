<script setup lang="ts">
import BaseKpiCard from './BaseKpiCard.vue'
import GeokretyTypeBreakdown from '@/components/breakdowns/GeokretyTypeBreakdown.vue'
import { computed } from 'vue'

interface GeokretyByType {
  traditional: number
  book: number
  human: number
  coin: number
  kretypost: number
  pebble: number
  car: number
  playingcard: number
  dogtag: number
  jigsaw: number
  easteregg: number
  [key: string]: number
}

interface Props {
  label: string
  value: string | number
  hiddenCount?: number
  geokretyByType?: GeokretyByType
  decimals?: number
}

const props = withDefaults(defineProps<Props>(), { decimals: 0, hiddenCount: 0 })

const formattedValue = computed(() => {
  if (typeof props.value === 'string') return props.value
  return props.value.toLocaleString('en-US', {
    minimumFractionDigits: props.decimals,
    maximumFractionDigits: props.decimals,
  })
})
</script>

<template>
  <BaseKpiCard
    stat="geokrety"
    :label="label"
    :value="formattedValue"
    :decimals="props.decimals"
    :secondaryValue="undefined"
  >
    <template #breakdown>
      <p class="mt-1 text-xs text-muted-foreground">
        Hidden in caches: {{ hiddenCount?.toLocaleString('en-US') ?? '0' }}
      </p>
      <div v-if="geokretyByType" class="mt-4">
        <GeokretyTypeBreakdown
          :geokrety-by-type="geokretyByType"
          :geokrety-count="typeof value === 'number' ? value : 0"
          compact
        />
      </div>
    </template>
  </BaseKpiCard>
</template>
