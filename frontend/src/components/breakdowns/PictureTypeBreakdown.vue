<script setup lang="ts">
import { HelpCircle } from 'lucide-vue-next'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import BaseBreakdown, { type BreakdownItem } from '@/components/breakdowns/BaseBreakdown.vue'
import { PICTURE_TYPES, type PictureTypeName } from '@/constants/pictureTypes'

interface PicturesByType {
  geokretAvatars: number
  geokretMoves: number
  userAvatars: number
}

interface Props {
  picturesByType: PicturesByType
  picturesCount: number
  /** compact = omit the section wrapper / heading (for use inside cards) */
  compact?: boolean
}

const props = withDefaults(defineProps<Props>(), { compact: false })

// Map API field names to our constant names
const mappedData = {
  geokretAvatar: props.picturesByType.geokretAvatars,
  geokretMove: props.picturesByType.geokretMoves,
  userAvatar: props.picturesByType.userAvatars,
}

// Convert PICTURE_TYPES to BreakdownItem format
const breakdownItems: BreakdownItem<PictureTypeName>[] = PICTURE_TYPES.map((type) => ({
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
      :total="picturesCount"
      :items="breakdownItems"
      :compact="compact"
    />
  </TooltipProvider>
</template>
