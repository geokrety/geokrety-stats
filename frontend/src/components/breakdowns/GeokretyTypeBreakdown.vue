<script setup lang="ts">
import { HelpCircle } from 'lucide-vue-next'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import BaseBreakdown, { type BreakdownItem } from '@/components/breakdowns/BaseBreakdown.vue'
import { GEOKRETY_TYPES, type GeokretyTypeName } from '@/constants/geokretyTypes'

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
  geokretyByType: GeokretyByType
  geokretyCount: number
  compact?: boolean
}

const props = withDefaults(defineProps<Props>(), { compact: false })

// Convert GEOKRETY_TYPES to BreakdownItem format
const breakdownItems: BreakdownItem<GeokretyTypeName>[] = GEOKRETY_TYPES.map((type) => ({
  id: type.id,
  name: type.name,
  label: type.label,
  colors: {
    bg: type.colors.bg,
    text: type.colors.text,
  },
}))
</script>

<template>
  <TooltipProvider>
    <BaseBreakdown
      :data="geokretyByType"
      :total="geokretyCount"
      :items="breakdownItems"
      :compact="compact"
    />
  </TooltipProvider>
</template>
