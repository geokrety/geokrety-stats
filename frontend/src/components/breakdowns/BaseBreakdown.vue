<script setup lang="ts">
import { ref } from 'vue'
import { ChevronDown } from 'lucide-vue-next'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

export interface BreakdownItem<T extends string = string> {
  id: number | string
  name: T
  label: string
  colors: {
    bg: string
    text: string
  }
}

interface Props {
  /** Breakdown item data (e.g., { dropped: 123, grabbed: 456, ... }) */
  data: Record<string, number>
  /** Total count for percentage calculation */
  total: number
  /** Breakdown type configuration */
  items: BreakdownItem[]
  /** Compact mode (smaller bar, simplified layout) */
  compact?: boolean
  /** Section title (omitted in compact mode) */
  title?: string
}

const props = withDefaults(defineProps<Props>(), {
  compact: false,
  title: 'Breakdown',
})

const showCompactLegend = ref(false)

function pct(part: number, total: number): string {
  if (!total) return '0%'
  return ((part / total) * 100).toFixed(1) + '%'
}

function fmt(n: number): string {
  return n.toLocaleString('en-US', { maximumFractionDigits: 0 })
}

function countFor(name: string): number {
  return props.data[name] ?? 0
}
</script>

<template>
  <div v-if="!compact">
    <!-- Visual bar with tooltips -->
    <TooltipProvider :delay-duration="100">
      <div class="flex gap-0.5 h-3 rounded-full overflow-hidden mb-4">
        <Tooltip v-for="item in items" :key="item.id">
          <TooltipTrigger as-child>
            <div
              :class="item.colors.bg"
              :style="{ width: pct(countFor(item.name), total) }"
              class="cursor-help transition-opacity hover:opacity-80"
            />
          </TooltipTrigger>
          <TooltipContent>
            <div class="flex items-center gap-2">
              <span class="h-3 w-3 rounded inline-block flex-shrink-0" :class="item.colors.bg" />
              <span class="font-medium">{{ item.label }}:</span>
              <span class="tabular-nums">{{ fmt(countFor(item.name)) }}</span>
              <span class="text-muted-foreground text-xs"
                >({{ pct(countFor(item.name), total) }})</span
              >
            </div>
          </TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>

    <!-- Grid view with all items -->
    <div class="grid grid-cols-2 lg:grid-cols-3 gap-2">
      <div
        v-for="item in items"
        :key="item.id"
        class="rounded-md border border-border/50 px-2 py-1.5"
      >
        <p class="text-[11px] text-muted-foreground leading-tight">{{ item.label }}</p>
        <p class="text-sm font-semibold tabular-nums" :class="item.colors.text">
          {{ fmt(countFor(item.name)) }}
        </p>
      </div>
    </div>
  </div>

  <!-- Compact variant -->
  <div v-else>
    <div class="flex items-center justify-between gap-2 mb-2">
      <p class="text-xs text-muted-foreground">{{ title }}</p>
      <button
        @click="showCompactLegend = !showCompactLegend"
        class="p-1 rounded hover:bg-accent transition-colors"
        :aria-label="showCompactLegend ? 'Hide legend' : 'Show legend'"
      >
        <ChevronDown
          :class="showCompactLegend ? 'rotate-180' : ''"
          class="h-4 w-4 text-muted-foreground transition-transform"
        />
      </button>
    </div>
    <TooltipProvider :delay-duration="100">
      <div class="flex gap-0.5 h-2 rounded-full overflow-hidden">
        <Tooltip v-for="item in items" :key="`compact-${item.id}`">
          <TooltipTrigger as-child>
            <div
              :class="item.colors.bg"
              :style="{ width: pct(countFor(item.name), total) }"
              class="cursor-help transition-opacity hover:opacity-80"
            />
          </TooltipTrigger>
          <TooltipContent>
            <div class="flex items-center gap-2">
              <span class="h-3 w-3 rounded inline-block flex-shrink-0" :class="item.colors.bg" />
              <span class="font-medium">{{ item.label }}:</span>
              <span class="tabular-nums">{{ fmt(countFor(item.name)) }}</span>
              <span class="text-muted-foreground text-xs"
                >({{ pct(countFor(item.name), total) }})</span
              >
            </div>
          </TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>

    <!-- Compact legend (toggleable, hidden by default) -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 max-h-0"
      enter-to-class="opacity-100 max-h-32"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 max-h-32"
      leave-to-class="opacity-0 max-h-0"
    >
      <div
        v-if="showCompactLegend"
        class="mt-2 grid grid-cols-1 gap-y-1 text-[10px] text-muted-foreground overflow-hidden"
      >
        <div v-for="item in items" :key="`legend-${item.id}`" class="flex items-center justify-between truncate">
          <div class="flex items-center gap-2 min-w-0">
            <span class="h-2 w-2 rounded inline-block flex-shrink-0" :class="item.colors.bg" />
            <span class="truncate text-[10px] text-muted-foreground">{{ item.label }}</span>
          </div>
          <span class="tabular-nums text-[10px] text-right text-muted-foreground">{{ fmt(countFor(item.name)) }}</span>
        </div>
      </div>
    </Transition>
  </div>
</template>
