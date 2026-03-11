/**
 * Canonical color tokens for picture types.
 *
 * Picture type mapping (from AGENTS.md):
 *   0 = PICTURE_GEOKRET_AVATAR  (GeoKrety item avatar)
 *   1 = PICTURE_GEOKRET_MOVE    (Photo attached to move)
 *   2 = PICTURE_USER_AVATAR     (User profile picture)
 */

export type PictureTypeName = 'geokretAvatar' | 'geokretMove' | 'userAvatar'

export interface PictureTypeColors {
  /** Tailwind bg-* class for the bar / swatch */
  bg: string
  /** Tailwind text-* class for the value */
  text: string
  /** Tailwind ring-* / border-* class */
  ring: string
  /** Hex value used where raw CSS colour is required */
  hex: string
}

export const PICTURE_TYPE_COLORS: Record<PictureTypeName, PictureTypeColors> = {
  geokretAvatar: {
    bg: 'bg-cyan-500',
    text: 'text-cyan-600 dark:text-cyan-400',
    ring: 'ring-cyan-500/40',
    hex: '#06b6d4',
  },
  geokretMove: {
    bg: 'bg-amber-500',
    text: 'text-amber-600 dark:text-amber-400',
    ring: 'ring-amber-500/40',
    hex: '#f59e0b',
  },
  userAvatar: {
    bg: 'bg-rose-500',
    text: 'text-rose-600 dark:text-rose-400',
    ring: 'ring-rose-500/40',
    hex: '#f43f5e',
  },
}

export interface PictureTypeInfo {
  id: number
  name: PictureTypeName
  label: string
  colors: PictureTypeColors
}

export const PICTURE_TYPES: PictureTypeInfo[] = [
  {
    id: 0,
    name: 'geokretAvatar',
    label: 'GeoKrety Avatars',
    colors: PICTURE_TYPE_COLORS.geokretAvatar,
  },
  {
    id: 1,
    name: 'geokretMove',
    label: 'Move Photos',
    colors: PICTURE_TYPE_COLORS.geokretMove,
  },
  {
    id: 2,
    name: 'userAvatar',
    label: 'User Avatars',
    colors: PICTURE_TYPE_COLORS.userAvatar,
  },
]
