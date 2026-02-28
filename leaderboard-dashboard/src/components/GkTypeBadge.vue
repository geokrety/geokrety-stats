<script setup>
/**
 * GkTypeBadge - displays a GeoKret type as a colored badge with emoji icon
 * Props:
 *   gkType (Number): numeric GeoKret type ID (0-10)
 *   typeName (String): type name string (optional, used as fallback)
 *   showLabel (Boolean): whether to show the type name text (default true)
 */
defineProps({
  gkType: { type: Number, default: null },
  typeName: { type: String, default: null },
  showLabel: { type: Boolean, default: true },
})

const TYPE_MAP = {
  0:  { name: 'Traditional', emoji: '🏷️', color: 'bg-success' },
  1:  { name: 'Book/CD/DVD', emoji: '📚', color: 'bg-info text-dark' },
  2:  { name: 'Human',       emoji: '👤', color: 'bg-primary' },
  3:  { name: 'Coin',        emoji: '🪙', color: 'bg-warning text-dark' },
  4:  { name: 'KretyPost',   emoji: '✉️',  color: 'bg-secondary' },
  5:  { name: 'Pebble',      emoji: '🪨', color: 'bg-dark' },
  6:  { name: 'Car',         emoji: '🚗', color: 'bg-danger' },
  7:  { name: 'Playing Card',emoji: '🃏', color: 'bg-info text-dark' },
  8:  { name: 'Dog Tag',     emoji: '🐶', color: 'bg-secondary' },
  9:  { name: 'Jigsaw',      emoji: '🧩', color: 'bg-warning text-dark' },
  10: { name: 'Easter Egg',  emoji: '🥚', color: 'bg-danger' },
}

function resolveType(gkType, typeName) {
  if (gkType !== null && gkType !== undefined && TYPE_MAP[gkType]) {
    return TYPE_MAP[gkType]
  }
  // fallback: find by name
  if (typeName) {
    const entry = Object.values(TYPE_MAP).find(t => t.name === typeName)
    if (entry) return entry
  }
  return { name: typeName || 'Unknown', emoji: '❓', color: 'bg-secondary' }
}
</script>

<template>
  <span :class="['badge', resolveType(gkType, typeName).color]"
        :title="resolveType(gkType, typeName).name">
    {{ resolveType(gkType, typeName).emoji }}
    <span v-if="showLabel"> {{ resolveType(gkType, typeName).name }}</span>
  </span>
</template>
