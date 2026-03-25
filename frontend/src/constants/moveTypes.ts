/**
 * Canonical color tokens for each GeoKrety move-type.
 *
 * Logtype mapping (from AGENTS.md):
 *   0 = dropped   1 = grabbed   2 = commented
 *   3 = seen      4 = archived  5 = dipped/visiting
 */

export type MoveTypeName = 'dropped' | 'grabbed' | 'dipped' | 'seen' | 'commented' | 'archived'

export interface MoveTypeColors {
  /** Tailwind bg-* class for the bar / swatch */
  bg: string
  /** Tailwind text-* class for the value */
  text: string
  /** Tailwind text-* class for the icon */
  icon: string
  /** Tailwind ring-* / border-* class */
  ring: string
  /** Hex value used where raw CSS colour is required (e.g. Leaflet) */
  hex: string
}

// TODO: create custom colors in shadcn theme for each move type and use them here instead of default muted ones
//  see Adding new colors https://www.shadcn-vue.com/docs/theming
export const MOVE_TYPE_COLORS: Record<MoveTypeName, MoveTypeColors> = {
  dropped: {
    bg: 'bg-move-dropped',
    text: 'text-move-dropped-foreground',
    icon: 'text-move-dropped-foreground',
    ring: 'ring-move-dropped/40',
    hex: '#5cb85c',
  },
  grabbed: {
    bg: 'bg-move-grabbed',
    text: 'text-move-grabbed-foreground',
    icon: 'text-move-grabbed-foreground',
    ring: 'ring-move-grabbed/40',
    hex: '#0891b2',
  },
  dipped: {
    bg: 'bg-move-dipped',
    text: 'text-move-dipped-foreground',
    icon: 'text-move-dipped-foreground',
    ring: 'ring-move-dipped/40',
    hex: '#f43f5e',
  },
  seen: {
    bg: 'bg-move-seen',
    text: 'text-move-seen-foreground',
    icon: 'text-move-seen-foreground',
    ring: 'ring-move-seen/40',
    hex: '#8b5cf6',
  },
  commented: {
    bg: 'bg-move-commented',
    text: 'text-move-commented-foreground',
    icon: 'text-move-commented-foreground',
    ring: 'ring-move-commented/40',
    hex: '#3b82f6',
  },
  archived: {
    bg: 'bg-move-archived',
    text: 'text-move-archived-foreground',
    icon: 'text-move-archived-foreground',
    ring: 'ring-move-archived/40',
    hex: '#f59e0b',
  },
}

export interface MoveTypeInfo {
  id: number
  name: MoveTypeName
  label: string
  colors: MoveTypeColors
}

export const MOVE_TYPES: MoveTypeInfo[] = [
  { id: 0, name: 'dropped', label: 'Dropped', colors: MOVE_TYPE_COLORS.dropped },
  { id: 1, name: 'grabbed', label: 'Grabbed', colors: MOVE_TYPE_COLORS.grabbed },
  { id: 2, name: 'commented', label: 'Commented', colors: MOVE_TYPE_COLORS.commented },
  { id: 3, name: 'seen', label: 'Seen', colors: MOVE_TYPE_COLORS.seen },
  { id: 4, name: 'archived', label: 'Archived', colors: MOVE_TYPE_COLORS.archived },
  { id: 5, name: 'dipped', label: 'Dipped', colors: MOVE_TYPE_COLORS.dipped },
]

/**
 * Canonical color tokens for stat KPI cards used in ActivityOverview.
 */
export interface StatKpiColors {
  icon: string
  text: string
  bg: string
}

export type StatKpiName =
  | 'activeUsers'
  | 'usersHome'
  | 'avgPointsPerMove'
  | 'loves'
  | 'inCache'
  | 'lost'
  | 'pointsHome'
  | 'pointsMoves'

export const STAT_KPI_COLORS: Record<StatKpiName, StatKpiColors> = {
  activeUsers: {
    icon: 'text-muted-foreground',
    text: 'text-foreground',
    bg: 'bg-muted',
  },
  usersHome: {
    icon: 'text-muted-foreground',
    text: 'text-foreground',
    bg: 'bg-muted',
  },
  avgPointsPerMove: {
    icon: 'text-muted-foreground',
    text: 'text-foreground',
    bg: 'bg-muted',
  },
  loves: {
    icon: 'text-rose-500',
    text: 'text-rose-700 dark:text-rose-200',
    bg: 'bg-rose-50 dark:bg-rose-950/20',
  },
  inCache: {
    icon: 'text-muted-foreground',
    text: 'text-foreground',
    bg: 'bg-muted',
  },
  lost: {
    icon: 'text-muted-foreground',
    text: 'text-foreground',
    bg: 'bg-muted',
  },
  pointsHome: {
    icon: 'text-muted-foreground',
    text: 'text-foreground',
    bg: 'bg-muted',
  },
  pointsMoves: {
    icon: 'text-muted-foreground',
    text: 'text-foreground',
    bg: 'bg-muted',
  },
}
