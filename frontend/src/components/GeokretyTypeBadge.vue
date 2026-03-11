<script setup lang="ts">
import { computed } from 'vue'
import GeokretyTypeIcon from '@/components/GeokretyTypeIcon.vue'
import {
  GEOKRETY_TYPES,
  GEOKRETY_TYPE_BY_ID,
  type GeokretyTypeName,
} from '@/constants/geokretyTypes'

interface Props {
  /** Integer type id (0–10) or name string */
  type: number | GeokretyTypeName
  /** When true, show only the icon without the label text */
  iconOnly?: boolean
}

const props = withDefaults(defineProps<Props>(), { iconOnly: false })

const meta = computed(() => {
  const name: GeokretyTypeName | undefined =
    typeof props.type === 'number'
      ? GEOKRETY_TYPE_BY_ID[props.type]
      : (props.type as GeokretyTypeName)

  return GEOKRETY_TYPES.find((t) => t.name === name) ?? null
})
</script>

<template>
  <span
    v-if="meta"
    class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium ring-1"
    :class="[meta.colors.bg, meta.colors.text, meta.colors.ring]"
    :title="meta.label"
  >
    <GeokretyTypeIcon :type="meta.name" size="inline" />
    <span v-if="!iconOnly">{{ meta.label }}</span>
  </span>
  <span
    v-else
    class="inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground ring-1 ring-border/30"
  >
    Unknown
  </span>
</template>
