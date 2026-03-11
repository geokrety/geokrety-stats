/**
 * GeoKrety type constants.
 *
 * Type mapping (from geokrety.org):
 *   0 = Traditional   1 = A book      2 = A human
 *   3 = A coin        4 = KretyPost   5 = Pebble
 *   6 = Car           7 = Playing card  8 = Dog tag / pet
 *   9 = Jigsaw part   10 = Easter egg
 */

export type GeokretyTypeName =
  | 'traditional'
  | 'book'
  | 'human'
  | 'coin'
  | 'kretypost'
  | 'pebble'
  | 'car'
  | 'playingcard'
  | 'dogtag'
  | 'jigsaw'
  | 'easteregg'

export const GEOKRETY_TYPE_ID: Record<GeokretyTypeName, number> = {
  traditional: 0,
  book: 1,
  human: 2,
  coin: 3,
  kretypost: 4,
  pebble: 5,
  car: 6,
  playingcard: 7,
  dogtag: 8,
  jigsaw: 9,
  easteregg: 10,
}

export const GEOKRETY_TYPE_BY_ID: Record<number, GeokretyTypeName> = Object.fromEntries(
  Object.entries(GEOKRETY_TYPE_ID).map(([name, id]) => [id, name as GeokretyTypeName]),
)

export interface GeokretyTypeColors {
  bg: string
  text: string
  icon: string
  ring: string
}

export const GEOKRETY_TYPE_COLORS: Record<GeokretyTypeName, GeokretyTypeColors> = {
  traditional: {
    bg: 'bg-gk-traditional',
    text: 'text-gk-traditional-foreground',
    icon: 'text-gk-traditional-foreground',
    ring: 'ring-gk-traditional/40',
  },
  book: {
    bg: 'bg-gk-book',
    text: 'text-gk-book-foreground',
    icon: 'text-gk-book-foreground',
    ring: 'ring-gk-book/40',
  },
  human: {
    bg: 'bg-gk-human',
    text: 'text-gk-human-foreground',
    icon: 'text-gk-human-foreground',
    ring: 'ring-gk-human/40',
  },
  coin: {
    bg: 'bg-gk-coin',
    text: 'text-gk-coin-foreground',
    icon: 'text-gk-coin-foreground',
    ring: 'ring-gk-coin/40',
  },
  kretypost: {
    bg: 'bg-gk-kretypost',
    text: 'text-gk-kretypost-foreground',
    icon: 'text-gk-kretypost-foreground',
    ring: 'ring-gk-kretypost/40',
  },
  pebble: {
    bg: 'bg-gk-pebble',
    text: 'text-gk-pebble-foreground',
    icon: 'text-gk-pebble-foreground',
    ring: 'ring-gk-pebble/40',
  },
  car: {
    bg: 'bg-gk-car',
    text: 'text-gk-car-foreground',
    icon: 'text-gk-car-foreground',
    ring: 'ring-gk-car/40',
  },
  playingcard: {
    bg: 'bg-gk-playingcard',
    text: 'text-gk-playingcard-foreground',
    icon: 'text-gk-playingcard-foreground',
    ring: 'ring-gk-playingcard/40',
  },
  dogtag: {
    bg: 'bg-gk-dogtag',
    text: 'text-gk-dogtag-foreground',
    icon: 'text-gk-dogtag-foreground',
    ring: 'ring-gk-dogtag/40',
  },
  jigsaw: {
    bg: 'bg-gk-jigsaw',
    text: 'text-gk-jigsaw-foreground',
    icon: 'text-gk-jigsaw-foreground',
    ring: 'ring-gk-jigsaw/40',
  },
  easteregg: {
    bg: 'bg-gk-easteregg',
    text: 'text-gk-easteregg-foreground',
    icon: 'text-gk-easteregg-foreground',
    ring: 'ring-gk-easteregg/40',
  },
}

export interface GeokretyTypeInfo {
  id: number
  name: GeokretyTypeName
  label: string
  colors: GeokretyTypeColors
}

export const GEOKRETY_TYPES: GeokretyTypeInfo[] = [
  { id: 0, name: 'traditional', label: 'Traditional', colors: GEOKRETY_TYPE_COLORS.traditional },
  { id: 1, name: 'book', label: 'A Book', colors: GEOKRETY_TYPE_COLORS.book },
  { id: 2, name: 'human', label: 'A Human', colors: GEOKRETY_TYPE_COLORS.human },
  { id: 3, name: 'coin', label: 'A Coin', colors: GEOKRETY_TYPE_COLORS.coin },
  { id: 4, name: 'kretypost', label: 'KretyPost', colors: GEOKRETY_TYPE_COLORS.kretypost },
  { id: 5, name: 'pebble', label: 'Pebble', colors: GEOKRETY_TYPE_COLORS.pebble },
  { id: 6, name: 'car', label: 'Car', colors: GEOKRETY_TYPE_COLORS.car },
  { id: 7, name: 'playingcard', label: 'Playing Card', colors: GEOKRETY_TYPE_COLORS.playingcard },
  { id: 8, name: 'dogtag', label: 'Dog Tag / Pet', colors: GEOKRETY_TYPE_COLORS.dogtag },
  { id: 9, name: 'jigsaw', label: 'Jigsaw Part', colors: GEOKRETY_TYPE_COLORS.jigsaw },
  { id: 10, name: 'easteregg', label: 'Easter Egg', colors: GEOKRETY_TYPE_COLORS.easteregg },
]
