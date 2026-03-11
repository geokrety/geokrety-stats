/**
 * Base statistics type definitions and color tokens.
 * Used for main KPI cards across the dashboard.
 */
import { Package, Footprints, Users, Globe, Image, type LucideIcon } from 'lucide-vue-next'

export type BaseStatName = 'geokrety' | 'moves' | 'users' | 'countries' | 'pictures'

export interface BaseStatMeta {
  icon: LucideIcon
  label: string
  colors: {
    /** Tailwind text-* class for icon */
    icon: string
    /** Tailwind text-* class for value */
    text: string
    /** Tailwind bg-* class for background */
    bg: string
    /** Tailwind ring-* class for border ring */
    ring: string
  }
}

export const BASE_STAT_META: Record<BaseStatName, BaseStatMeta> = {
  geokrety: {
    icon: Package,
    label: 'GeoKrety',
    colors: {
      icon: 'text-emerald-600 dark:text-emerald-400',
      text: 'text-emerald-600 dark:text-emerald-400',
      bg: 'bg-emerald-50 dark:bg-emerald-950/30',
      ring: 'ring-emerald-500/20',
    },
  },
  moves: {
    icon: Footprints,
    label: 'Moves',
    colors: {
      icon: 'text-blue-600 dark:text-blue-400',
      text: 'text-blue-600 dark:text-blue-400',
      bg: 'bg-blue-50 dark:bg-blue-950/30',
      ring: 'ring-blue-500/20',
    },
  },
  users: {
    icon: Users,
    label: 'Users',
    colors: {
      icon: 'text-purple-600 dark:text-purple-400',
      text: 'text-purple-600 dark:text-purple-400',
      bg: 'bg-purple-50 dark:bg-purple-950/30',
      ring: 'ring-purple-500/20',
    },
  },
  countries: {
    icon: Globe,
    label: 'Countries',
    colors: {
      icon: 'text-cyan-600 dark:text-cyan-400',
      text: 'text-cyan-600 dark:text-cyan-400',
      bg: 'bg-cyan-50 dark:bg-cyan-950/30',
      ring: 'ring-cyan-500/20',
    },
  },
  pictures: {
    icon: Image,
    label: 'Pictures',
    colors: {
      icon: 'text-amber-600 dark:text-amber-400',
      text: 'text-amber-600 dark:text-amber-400',
      bg: 'bg-amber-50 dark:bg-amber-950/30',
      ring: 'ring-amber-500/20',
    },
  },
}
