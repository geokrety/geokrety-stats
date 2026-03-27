<script setup lang="ts">
import { ref, computed } from 'vue'
import { Avatar } from '@/components/ui/avatar'
import AvatarDisplay from '@/components/AvatarDisplay.vue'

interface AvatarItem {
  id: string | number
  src?: string | null
  thumbnailSrc?: string | null
  alt?: string
  caption?: string
  badge?: { level?: number; iconSrc?: string; color?: string }
}

interface Props {
  /**
   * One or more avatar items.
   * Accepts a single item or an array.
   */
  avatars: AvatarItem | AvatarItem[]
  /** Pixel size of each avatar (default 36) */
  size?: number
  /** How much each avatar overlaps the previous one in pixels (default 10) */
  overlap?: number
  /** Maximum number of avatars to render before showing +N indicator */
  max?: number
  /** Shape forwarded to AvatarDisplay */
  shape?: 'circle' | 'rounded'
}

const props = withDefaults(defineProps<Props>(), {
  size: 36,
  overlap: 10,
  max: 8,
  shape: 'circle',
})

const list = computed<AvatarItem[]>(() =>
  Array.isArray(props.avatars) ? props.avatars : [props.avatars],
)

const visible = computed(() => list.value.slice(0, props.max))
const overflow = computed(() => Math.max(0, list.value.length - props.max))
const avatarSize = computed(() => `${props.size}px`)
const overlapOffset = computed(() => `-${props.overlap}px`)
const overflowFontSize = computed(() => `${Math.max(9, props.size * 0.27)}px`)

// Track which index is being hovered to bring it to the front
const hoveredIndex = ref<number | null>(null)

/** Compute the stacking z-index: hovered item always on top */
function zIndex(i: number): number {
  if (hoveredIndex.value === i) return 20
  // items later in the list sit on top by default so -i gives a descending stack
  return visible.value.length - i
}

const itemStyle = computed(() => (i: number) => ({
  '--avatar-group-z-index': zIndex(i),
}))
</script>

<template>
  <div class="inline-flex items-center">
    <div
      v-for="(avatar, i) in visible"
      :key="avatar.id"
      class="avatar-group__item relative"
      :style="itemStyle(i)"
      @mouseenter="hoveredIndex = i"
      @mouseleave="hoveredIndex = null"
    >
      <AvatarDisplay
        :src="avatar.src"
        :thumbnail-src="avatar.thumbnailSrc"
        :alt="avatar.alt ?? String(avatar.id)"
        :caption="avatar.caption"
        :size="size"
        :shape="shape"
        :badge="avatar.badge"
      />
    </div>

    <!-- Overflow indicator using shadcn Avatar -->
    <Avatar
      v-if="overflow > 0"
      class="avatar-group__overflow relative ring-2 ring-border"
    >
      <div
        class="avatar-group__overflow-label flex h-full w-full select-none items-center justify-center bg-muted font-semibold text-muted-foreground"
        :class="shape === 'rounded' ? 'rounded-lg' : 'rounded-full'"
      >
        +{{ overflow }}
      </div>
    </Avatar>
  </div>
</template>

<style scoped>
.avatar-group__item {
  margin-left: v-bind(overlapOffset);
  transition: z-index 0s;
  z-index: var(--avatar-group-z-index);
}

.avatar-group__item:first-child {
  margin-left: 0;
}

.avatar-group__overflow {
  height: v-bind(avatarSize);
  margin-left: v-bind(overlapOffset);
  width: v-bind(avatarSize);
  z-index: 1;
}

.avatar-group__overflow-label {
  font-size: v-bind(overflowFontSize);
}
</style>
