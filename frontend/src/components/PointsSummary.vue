<script setup lang="ts">
import { computed } from 'vue'
import { TrendingUp } from 'lucide-vue-next'
import { STAT_KPI_COLORS } from '@/constants/moveTypes'
import { Card, CardContent } from '@/components/ui/card'
import { formatNumber } from '@/lib/format'

interface Props {
  countryName: string
  pointsSum: number
  pointsSumMoves: number
}

const props = defineProps<Props>()

function pct(part: number, total: number): string {
  if (!total) return '0%'
  return ((part / total) * 100).toFixed(1) + '%'
}

const homePointsWidth = computed(() => pct(props.pointsSum, props.pointsSum + props.pointsSumMoves))
</script>

<template>
  <Card class="rounded-xl border bg-card/70">
    <CardContent class="p-5">
      <h2 class="text-xs font-semibold uppercase tracking-widest text-muted-foreground mb-4">
        Points summary
      </h2>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
        <div>
          <p
            class="text-sm text-muted-foreground mb-1 flex items-center gap-1.5"
            :class="STAT_KPI_COLORS.pointsHome.icon"
          >
            <TrendingUp class="h-4 w-4" :class="STAT_KPI_COLORS.pointsHome.icon" />
            Points Σ — home users
          </p>
          <p class="text-3xl font-bold tabular-nums" :class="STAT_KPI_COLORS.pointsHome.text">
            {{ formatNumber(pointsSum) }}
          </p>
          <p class="text-xs text-muted-foreground mt-1">
            Sum of all points earned by users whose home coordinates are in {{ countryName }}.
          </p>
        </div>
        <div>
          <p class="text-sm text-muted-foreground mb-1 flex items-center gap-1.5">
            <TrendingUp class="h-4 w-4" :class="STAT_KPI_COLORS.pointsMoves.icon" />
            Points Σ — moves
          </p>
          <p class="text-3xl font-bold tabular-nums" :class="STAT_KPI_COLORS.pointsMoves.text">
            {{ formatNumber(pointsSumMoves) }}
          </p>
          <p class="text-xs text-muted-foreground mt-1">
            Sum of all points earned by moves physically made inside {{ countryName }}.
          </p>
        </div>
      </div>

      <!-- Comparison bar -->
      <div class="mt-4 pt-4 border-t border-border">
        <p class="text-xs text-muted-foreground mb-2">Home vs moves points ratio</p>
        <div class="flex gap-1 h-2 rounded-full overflow-hidden bg-muted">
          <div
            class="points-summary__fill bg-primary rounded-full transition-all duration-500"
            title="Home users pts"
          />
        </div>
        <div class="flex justify-between text-xs text-muted-foreground mt-1">
          <span>Home: {{ pct(pointsSum, pointsSum + pointsSumMoves) }}</span>
          <span>Moves: {{ pct(pointsSumMoves, pointsSum + pointsSumMoves) }}</span>
        </div>
      </div>
    </CardContent>
  </Card>
</template>

<style scoped>
.points-summary__fill {
  width: v-bind(homePointsWidth);
}
</style>
