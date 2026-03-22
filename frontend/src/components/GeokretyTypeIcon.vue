<script setup lang="ts">
import { computed } from 'vue'
import {
  Package,
  BookOpen,
  PersonStanding,
  CircleDollarSign,
  MailOpen,
  Gem,
  Car,
  Spade,
  PawPrint,
  Puzzle,
  Egg,
} from 'lucide-vue-next'
import type { Component } from 'vue'
import {
  GEOKRETY_TYPES,
  GEOKRETY_TYPE_BY_ID,
  type GeokretyTypeName,
} from '@/constants/geokretyTypes'

interface Props {
  /**
   * Accept either the integer type id (0–10) or the type name string.
   * Accepts both so components receiving raw API data can pass the id directly.
   */
  type: number | GeokretyTypeName
  /** `large` = icon in a coloured card  |  `inline` = small inline icon (default) */
  size?: 'large' | 'inline'
}

const props = withDefaults(defineProps<Props>(), { size: 'inline' })

const ICON_MAP: Record<GeokretyTypeName, Component> = {
  traditional: Package,
  book: BookOpen,
  human: PersonStanding,
  coin: CircleDollarSign,
  kretypost: MailOpen,
  pebble: Gem,
  car: Car,
  playingcard: Spade,
  dogtag: PawPrint,
  jigsaw: Puzzle,
  easteregg: Egg,
}

const meta = computed(() => {
  const name: GeokretyTypeName | undefined =
    typeof props.type === 'number'
      ? GEOKRETY_TYPE_BY_ID[props.type]
      : (props.type as GeokretyTypeName)

  const typeInfo = GEOKRETY_TYPES.find((t) => t.name === name)
  const icon = name ? ICON_MAP[name] : null
  return typeInfo && icon ? { ...typeInfo, icon } : null
})
</script>

<template>
  <!-- Large: coloured background card icon -->
  <span
    v-if="size === 'large' && meta"
    class="inline-flex h-10 w-10 items-center justify-center rounded-xl ring-1"
    :class="[meta.colors.bg, meta.colors.ring]"
    :title="meta.label"
  >
    <component :is="meta.icon" class="h-5 w-5" :class="meta.colors.icon" />
  </span>

  <!-- Inline: small icon -->
  <component
    v-else-if="meta"
    :is="meta.icon"
    class="h-3.5 w-3.5"
    :class="meta?.colors.icon"
    :title="meta.label"
  />
</template>
