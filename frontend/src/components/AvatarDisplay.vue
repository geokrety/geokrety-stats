<script setup lang="ts">
import { ref, computed } from 'vue'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'
import { Dialog, DialogContent, DialogClose } from '@/components/ui/dialog'
import { HoverCard, HoverCardContent, HoverCardTrigger } from '@/components/ui/hover-card'
import { X } from 'lucide-vue-next'

const DEFAULT_AVATAR = 'https://cdn.geokrety.org/images/the-mole.svg'

interface LevelBadge {
  /** Numeric level shown as text inside the badge */
  level?: number
  /** Custom icon src instead of numeric level */
  iconSrc?: string
  /** Tailwind colour class for badge background (default: bg-emerald-500) */
  color?: string
}

interface Props {
  /** Full-size image URL (shown on click / hover expansion) */
  src?: string | null
  /** Thumbnail URL – used for the default display; falls back to `src` */
  thumbnailSrc?: string | null
  /** Alt text / accessible label */
  alt?: string
  /** Caption shown in the fullscreen lightbox */
  caption?: string
  /**
   * Preset sizes or an explicit pixel value.
   * xs = 24 px, sm = 32 px, md = 40 px (default), lg = 56 px, xl = 80 px
   */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | number
  /** `circle` (default) or `rounded` (rounded-lg) */
  shape?: 'circle' | 'rounded'
  /** Optional level/rank badge shown in the bottom-right corner */
  badge?: LevelBadge
  /** Fallback image when `src` / `thumbnailSrc` cannot be loaded */
  fallback?: string
  /**
   * Delay (ms) before the hover full-size preview appears.
   * Set to 0 to disable the hover preview.
   */
  hoverDelay?: number
}

const props = withDefaults(defineProps<Props>(), {
  alt: 'Avatar',
  size: 'md',
  shape: 'circle',
  hoverDelay: 400,
  fallback: DEFAULT_AVATAR,
})

// ── Size resolution ──────────────────────────────────────────────────────────
const SIZE_MAP: Record<string, number> = { xs: 24, sm: 32, md: 40, lg: 56, xl: 80 }

const px = computed(() =>
  typeof props.size === 'number' ? props.size : (SIZE_MAP[props.size] ?? 40),
)
const avatarDimension = computed(() => `${px.value}px`)

// ── Derived image URLs ────────────────────────────────────────────────────────
const thumbSrc = computed(() => props.thumbnailSrc ?? props.src ?? props.fallback)
const fullSrc = computed(() => props.src ?? props.fallback)

const shapeClass = computed(() => (props.shape === 'rounded' ? 'rounded-lg' : 'rounded-full'))

// ── Image error handling ─────────────────────────────────────────────────────
const thumbError = ref(false)
const fullError = ref(false)
const displayThumb = computed(() => (thumbError.value ? props.fallback : thumbSrc.value))
const displayFull = computed(() => (fullError.value ? props.fallback : fullSrc.value))

// ── Lightbox ──────────────────────────────────────────────────────────────────
const lightboxOpen = ref(false)

function openLightbox() {
  lightboxOpen.value = true
}

function closeLightbox() {
  lightboxOpen.value = false
}
</script>

<template>
  <!-- ── Avatar with hover preview and lightbox ──────────────────────────── -->
  <HoverCard v-if="hoverDelay > 0" :open-delay="hoverDelay">
    <HoverCardTrigger as-child>
      <Avatar
        :class="[
          'avatar-display',
          shapeClass,
          'cursor-pointer ring-2 ring-border/10 transition-opacity hover:opacity-90',
        ]"
        @click="openLightbox"
      >
        <AvatarImage :src="displayThumb ?? undefined" :alt="alt" @error="thumbError = true" />
        <AvatarFallback>
          <img :src="fallback" :alt="alt" class="h-full w-full object-cover" />
        </AvatarFallback>

        <!-- Level / badge (bottom-right corner) -->
        <span
          v-if="badge"
          class="avatar-display__badge pointer-events-none absolute bottom-0 right-0 flex items-center justify-center rounded-full ring-2 ring-border text-foreground text-[9px] font-bold leading-none"
          :class="badge.color ?? 'bg-emerald-500'"
        >
          <img
            v-if="badge.iconSrc"
            :src="badge.iconSrc"
            class="w-full h-full object-contain rounded-full"
          />
          <span v-else-if="badge.level !== undefined">{{ badge.level }}</span>
        </span>
      </Avatar>
    </HoverCardTrigger>

    <HoverCardContent side="right" :side-offset="8" class="w-auto p-1">
      <img
        :src="displayFull ?? undefined"
        :alt="alt"
        class="max-w-[200px] max-h-[200px] object-cover rounded-lg"
        draggable="false"
        @error="fullError = true"
      />
      <p
        v-if="caption"
        class="mt-1 max-w-[200px] truncate text-center text-xs text-muted-foreground"
      >
        {{ caption }}
      </p>
    </HoverCardContent>
  </HoverCard>

  <!-- Avatar without hover preview -->
  <Avatar
    v-else
    :class="[
      'avatar-display',
      shapeClass,
      'cursor-pointer ring-2 ring-border/10 transition-opacity hover:opacity-90',
    ]"
    @click="openLightbox"
  >
    <AvatarImage :src="displayThumb ?? undefined" :alt="alt" @error="thumbError = true" />
    <AvatarFallback>
      <img :src="fallback" :alt="alt" class="h-full w-full object-cover" />
    </AvatarFallback>

    <!-- Level / badge (bottom-right corner) -->
    <span
      v-if="badge"
      class="avatar-display__badge pointer-events-none absolute bottom-0 right-0 flex items-center justify-center rounded-full ring-2 ring-border text-foreground text-[9px] font-bold leading-none"
      :class="badge.color ?? 'bg-emerald-500'"
    >
      <img
        v-if="badge.iconSrc"
        :src="badge.iconSrc"
        class="w-full h-full object-contain rounded-full"
      />
      <span v-else-if="badge.level !== undefined">{{ badge.level }}</span>
    </span>
  </Avatar>

  <!-- ── Fullscreen lightbox using shadcn Dialog ────────────────────────── -->
  <Dialog v-model:open="lightboxOpen">
    <DialogContent :show-close-button="false" class="max-w-[95vw] border-0 bg-card/85 p-0 backdrop-blur-sm">
      <div class="relative flex flex-col items-center justify-center p-8">
        <DialogClose
          class="absolute right-4 top-4 rounded-full bg-card/10 p-2 text-foreground transition hover:bg-card/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <X class="h-5 w-5" />
          <span class="sr-only">Close</span>
        </DialogClose>

        <!-- Full-size image -->
        <img
          :src="displayFull ?? undefined"
          :alt="alt"
          class="max-h-[85vh] max-w-full rounded-xl object-contain shadow-2xl"
          draggable="false"
          @error="fullError = true"
        />

        <!-- Caption -->
        <p v-if="caption" class="mt-4 max-w-[90vw] text-center text-sm text-muted-foreground">
          {{ caption }}
        </p>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.avatar-display {
  height: v-bind(avatarDimension);
  width: v-bind(avatarDimension);
}

.avatar-display__badge {
  height: 40%;
  min-height: 14px;
  min-width: 14px;
  width: 40%;
}
</style>
