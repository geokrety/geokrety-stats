<script setup lang="ts">
import { computed } from 'vue'
import BaseKpiCard from './BaseKpiCard.vue'
import PictureTypeBreakdown from '@/components/breakdowns/PictureTypeBreakdown.vue'
import type { StatKey } from '@/components/StatIcon.vue'

interface PicturesByType {
  geokretAvatars: number
  geokretMoves: number
  userAvatars: number
}

interface Props {
  stat: StatKey
  label: string
  value: string | number
  decimals?: number
  picturesByType?: PicturesByType
}

const props = withDefaults(defineProps<Props>(), { decimals: 0 })

const formattedValue = computed(() => {
  if (typeof props.value === 'string') return props.value
  return props.value.toLocaleString('en-US', {
    minimumFractionDigits: props.decimals,
    maximumFractionDigits: props.decimals,
  })
})
</script>

<template>
  <BaseKpiCard :stat="stat" :label="label" :value="formattedValue" :decimals="props.decimals">
    <template #breakdown>
      <div v-if="picturesByType" class="mt-4">
        <PictureTypeBreakdown
          :pictures-by-type="picturesByType"
          :pictures-count="typeof value === 'number' ? value : 0"
          compact
        />
      </div>
    </template>
  </BaseKpiCard>
</template>
