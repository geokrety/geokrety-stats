<script setup lang="ts">
import { computed } from 'vue'
import StatIcon, { type StatKey } from '@/components/StatIcon.vue'
import { Card, CardContent } from '@/components/ui/card'

interface Props {
  stat: StatKey
  label: string
  value: string | number
  secondaryValue?: number
  secondaryLabel?: string
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

const formattedSecondary = computed(() => {
  if (!props.secondaryValue) return ''
  return props.secondaryValue.toLocaleString('en-US', { maximumFractionDigits: 0 })
})
</script>

<template>
  <Card class="rounded-xl border bg-card text-card-foreground p-0">
    <CardContent class="p-4">
      <div class="mb-1 flex items-center gap-2 text-muted-foreground">
        <StatIcon :stat="stat" size="inline" />
        {{ label }}
      </div>
      <p class="text-2xl font-bold tabular-nums text-foreground">
        {{ formattedValue }}
      </p>
      <p v-if="secondaryValue" class="mt-1 text-xs text-muted-foreground">
        {{ formattedSecondary }} {{ secondaryLabel || 'in last 30 days' }}
      </p>
    </CardContent>
  </Card>
</template>
