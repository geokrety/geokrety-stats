<script setup lang="ts">
import { computed } from 'vue'
import { UserCheck, Heart, Activity, TrendingUp, MapPinOff } from 'lucide-vue-next'
import type { Component } from 'vue'
import { BASE_STAT_META, type BaseStatName } from '@/constants/statTypes'
import { STAT_KPI_COLORS, type StatKpiName } from '@/constants/moveTypes'

interface StatMeta {
  icon: Component
  color: string
  bgColor: string
  ring: string
}

export type StatKey = BaseStatName | StatKpiName

interface Props {
  stat: StatKey
  /**
   * `large`  – standalone card icon with coloured background ring (default)
   * `inline` – small icon used inline inside KPI/label rows
   */
  size?: 'large' | 'inline'
}

const props = withDefaults(defineProps<Props>(), { size: 'large' })

// Build stat metadata by merging base stats and KPI stats
const buildStatMeta = (): Record<StatKey, StatMeta> => {
  const meta: Record<string, StatMeta> = {}

  // Add base stats (geokrety, moves, users, countries, pictures)
  for (const [key, value] of Object.entries(BASE_STAT_META)) {
    meta[key] = {
      icon: value.icon,
      color: value.colors.icon,
      bgColor: value.colors.bg,
      ring: value.colors.ring,
    }
  }

  // Add KPI stats
  meta.activeUsers = {
    icon: UserCheck,
    color: STAT_KPI_COLORS.activeUsers.icon,
    bgColor: STAT_KPI_COLORS.activeUsers.bg,
    ring: 'ring-border/60',
  }
  meta.usersHome = {
    icon: BASE_STAT_META.users.icon,
    color: STAT_KPI_COLORS.usersHome.icon,
    bgColor: STAT_KPI_COLORS.usersHome.bg,
    ring: 'ring-border/60',
  }
  meta.avgPointsPerMove = {
    icon: Activity,
    color: STAT_KPI_COLORS.avgPointsPerMove.icon,
    bgColor: STAT_KPI_COLORS.avgPointsPerMove.bg,
    ring: 'ring-border/60',
  }
  meta.loves = {
    icon: Heart,
    color: STAT_KPI_COLORS.loves.icon,
    bgColor: STAT_KPI_COLORS.loves.bg,
    ring: 'ring-border/60',
  }
  meta.inCache = {
    icon: BASE_STAT_META.geokrety.icon,
    color: STAT_KPI_COLORS.inCache.icon,
    bgColor: STAT_KPI_COLORS.inCache.bg,
    ring: 'ring-border/60',
  }
  meta.lost = {
    icon: MapPinOff,
    color: STAT_KPI_COLORS.lost.icon,
    bgColor: STAT_KPI_COLORS.lost.bg,
    ring: 'ring-border/60',
  }
  meta.pointsHome = {
    icon: TrendingUp,
    color: STAT_KPI_COLORS.pointsHome.icon,
    bgColor: STAT_KPI_COLORS.pointsHome.bg,
    ring: 'ring-border/60',
  }
  meta.pointsMoves = {
    icon: TrendingUp,
    color: STAT_KPI_COLORS.pointsMoves.icon,
    bgColor: STAT_KPI_COLORS.pointsMoves.bg,
    ring: 'ring-border/60',
  }

  return meta as Record<StatKey, StatMeta>
}

const statMeta = buildStatMeta()
const meta = computed(() => statMeta[props.stat])
</script>

<template>
  <!-- Large: coloured background card icon (used in StatsCounters) -->
  <component
    :is="meta?.icon"
    v-if="size === 'large'"
    class="h-10 w-10 rounded-xl p-2.5 ring-1"
    :class="[meta?.color, meta?.bgColor, meta?.ring]"
  />
  <!-- Inline: small icon for KPI label rows -->
  <component :is="meta?.icon" v-else class="h-3.5 w-3.5" :class="meta?.color" />
</template>
