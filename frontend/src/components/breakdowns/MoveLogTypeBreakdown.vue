<script setup lang="ts">
import { HelpCircle } from 'lucide-vue-next'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import BaseBreakdown, { type BreakdownItem } from '@/components/breakdowns/BaseBreakdown.vue'
import { MOVE_TYPES, type MoveTypeName } from '@/constants/moveTypes'

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
  movesByType: MovesByType
  movesCount: number
  compact?: boolean
}

const props = withDefaults(defineProps<Props>(), { compact: false })

// Convert MOVE_TYPES to BreakdownItem format
const breakdownItems: BreakdownItem<MoveTypeName>[] = MOVE_TYPES.map((type) => ({
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
      :data="movesByType"
      :total="movesCount"
      :items="breakdownItems"
      :compact="compact"
    />
  </TooltipProvider>
</template>
