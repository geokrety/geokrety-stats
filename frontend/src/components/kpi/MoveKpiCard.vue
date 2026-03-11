<script setup lang="ts">
import BaseKpiCard from './BaseKpiCard.vue'
import MoveLogTypeBreakdown from '@/components/breakdowns/MoveLogTypeBreakdown.vue'
import { computed } from 'vue'

interface MovesByType {
  dropped: number
  grabbed: number
  commented: number
  seen: number
  archived: number
  dipped: number
  [key: string]: number
}

interface Props {
  label: string
  value: string | number
  movesByType?: MovesByType
  movesLast30Days?: number
  decimals?: number
}

const props = withDefaults(defineProps<Props>(), { decimals: 0 })

const formattedValue = computed(() => {
  if (typeof props.value === 'string') return props.value
  return props.value.toLocaleString('en-US', {
    minimumFractionDigits: props.decimals,
    maximumFractionDigits: props.decimals,
  })
})

const formattedLast30Days = computed(() => {
  if (!props.movesLast30Days) return ''
  return props.movesLast30Days.toLocaleString('en-US', { maximumFractionDigits: 0 })
})
</script>

<template>
  <BaseKpiCard
    stat="moves"
    :label="label"
    :value="formattedValue"
    :decimals="props.decimals"
    :secondaryValue="props.movesLast30Days"
    secondaryLabel="in last 30 days"
  >
    <template #breakdown>
      <div v-if="movesByType" class="mt-4">
        <MoveLogTypeBreakdown
          :moves-by-type="movesByType"
          :moves-count="typeof value === 'number' ? value : 0"
          compact
        />
      </div>
    </template>
  </BaseKpiCard>
</template>
