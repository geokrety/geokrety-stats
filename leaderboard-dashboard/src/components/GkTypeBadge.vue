<script setup>
/**
 * GkTypeBadge - Reusable badge for GeoKrety type display.
 *
 * GeoKrety type mapping (from geokrety PHP constants):
 *   0  Traditional   🏷️
 *   1  Book/CD/DVD   📚
 *   2  Human         👤
 *   3  Coin          🪙
 *   4  KretyPost     ✉️
 *   5  Pebble        🪨
 *   6  Car           🚗
 *   7  Playing Card  🃏
 *   8  Dog Tag       🐶
 *   9  Jigsaw        🧩
 *  10  Easter Egg    🥚
 *  --  Unknown       ❓
 */

const props = defineProps({
  gkType: { type: Number, default: null },
  typeName: { type: String, default: null },
})

const TYPE_MAP = {
  0:  { icon: '🏷️', label: 'Traditional',  cls: 'bg-warning text-dark' },
  1:  { icon: '📚', label: 'Book/CD/DVD',  cls: 'bg-info text-dark' },
  2:  { icon: '👤', label: 'Human',         cls: 'bg-primary' },
  3:  { icon: '🪙', label: 'Coin',          cls: 'bg-warning text-dark' },
  4:  { icon: '✉️', label: 'KretyPost',     cls: 'bg-danger' },
  5:  { icon: '🪨', label: 'Pebble',        cls: 'bg-secondary' },
  6:  { icon: '🚗', label: 'Car',           cls: 'bg-success' },
  7:  { icon: '🃏', label: 'Playing Card',  cls: 'bg-info text-dark' },
  8:  { icon: '🐶', label: 'Dog Tag',       cls: 'bg-dark' },
  9:  { icon: '🧩', label: 'Jigsaw',        cls: 'bg-primary' },
  10: { icon: '🥚', label: 'Easter Egg',    cls: 'bg-danger' },
}

const UNKNOWN = { icon: '❓', label: 'Unknown', cls: 'bg-secondary' }

function getInfo() {
  if (props.gkType !== null && props.gkType !== undefined) {
    return TYPE_MAP[props.gkType] ?? UNKNOWN
  }
  // Fall back to string name if type id not available
  const name = (props.typeName || '').toLowerCase()
  for (const [k, v] of Object.entries(TYPE_MAP)) {
    if (v.label.toLowerCase() === name) return v
  }
  return UNKNOWN
}

const info = getInfo()
</script>

<template>
  <span
    :class="['badge', info.cls]"
    :title="info.label"
  >
    {{ info.icon }} {{ info.label }}
  </span>
</template>
