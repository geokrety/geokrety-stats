<script setup lang="ts">
import { HelpCircle } from 'lucide-vue-next'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import BaseBreakdown, { type BreakdownItem } from '@/components/breakdowns/BaseBreakdown.vue'
import { MOVE_TYPES, type MoveTypeName } from '@/constants/moveTypes'

interface MovesByType {
  dropped: number
  dipped: number
  seen: number
  [key: string]: number
}

interface Props {
  movesByType: MovesByType
  movesCount: number
  /** compact = omit the section wrapper / heading (for use inside cards) */
  compact?: boolean
}

const props = withDefaults(defineProps<Props>(), { compact: false })

// Map API/local field names to the keys used by MOVE_TYPES
const mappedData = {
  dropped: props.movesByType.dropped,
  seen: props.movesByType.seen,
  dipped: props.movesByType.dipped,
}

// Convert MOVE_TYPES to BreakdownItem format (only the 3 types for this component)
const breakdownItems: BreakdownItem<MoveTypeName>[] = MOVE_TYPES.filter((type) =>
  ['dropped', 'seen', 'dipped'].includes(type.name),
).map((type) => ({
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
      :data="mappedData"
      :total="movesCount"
      :items="breakdownItems"
      :compact="compact"
    />
  </TooltipProvider>
</template>
