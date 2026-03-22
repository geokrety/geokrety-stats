<script setup lang="ts">
import {
  ArrowDownToLine,
  ArrowUpFromLine,
  MapPinned,
  Eye,
  MessageCircle,
  Archive,
} from 'lucide-vue-next'

interface MoveTypeInfo {
  label: string
  classes: string
  icon: unknown
}

const props = defineProps<{
  type: string
}>()

// TODO: create custom colors in shadcn theme for each move type and use them here instead of default muted ones
//  see Adding new colors https://www.shadcn-vue.com/docs/theming
// also consider moving this mapping to a separate constants file if it will be used in multiple places across the app
// e.g. src/constants/moveTypes.ts src/constants/geokretyTypes.ts
const moveTypeMeta: Record<string, MoveTypeInfo> = {
  dropped: {
    label: 'Dropped',
    icon: ArrowDownToLine,
    classes: 'bg-move-dropped text-move-dropped-foreground ring-move-dropped/40',
  },
  grabbed: {
    label: 'Grabbed',
    icon: ArrowUpFromLine,
    classes: 'bg-move-grabbed text-move-grabbed-foreground ring-move-grabbed/40',
  },
  dipped: {
    label: 'Dipped',
    icon: MapPinned,
    classes: 'bg-move-dipped text-move-dipped-foreground ring-move-dipped/40',
  },
  seen: {
    label: 'Seen',
    icon: Eye,
    classes: 'bg-move-seen text-move-seen-foreground ring-move-seen/40',
  },
  commented: {
    label: 'Commented',
    icon: MessageCircle,
    classes: 'bg-move-commented text-move-commented-foreground ring-move-commented/40',
  },
  archived: {
    label: 'Archived',
    icon: Archive,
    classes: 'bg-move-archived text-move-archived-foreground ring-move-archived/40',
  },
}

const meta = moveTypeMeta[props.type.toLowerCase()]
</script>

<template>
  <span
    class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ring-1"
    :class="meta?.classes ?? 'bg-muted text-muted-foreground'"
  >
    <component :is="meta?.icon" v-if="meta?.icon" class="h-3 w-3" />
    {{ meta?.label ?? 'Unknown' }}
  </span>
</template>
