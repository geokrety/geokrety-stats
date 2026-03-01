/**
 * Map move type names to Bootstrap badge colors
 * The color palette mirrors the charts and badges used across the dashboard.
 */
const MOVE_TYPE_COLORS = {
  drop: 'bg-success text-dark',
  drops: 'bg-success text-dark',
  dropped: 'bg-success text-dark',
  grab: 'bg-warning text-dark',
  grabs: 'bg-warning text-dark',
  take: 'bg-warning text-dark',
  catch: 'bg-warning text-dark',
  found: 'bg-warning text-dark',
  dip: 'bg-info text-dark',
  dipped: 'bg-info text-dark',
  move: 'bg-info text-dark',
  seen: 'bg-secondary',
  comment: 'bg-secondary',
  archived: 'bg-dark text-light',
  archive: 'bg-dark text-light',
  recovered: 'bg-success text-dark',
}

export function getMoveTypeBadgeClass(typeName) {
  if (!typeName) return 'bg-secondary'

  const lowerType = typeof typeName === 'string' ? typeName.toLowerCase() : typeName
  return MOVE_TYPE_COLORS[lowerType] || 'bg-secondary'
}

/**
 * GeoKret type ID to name mapping (from backend API)
 */
const GEOKRETY_TYPES = {
  0: 'Traditional',
  1: 'Book/CD/DVD',
  2: 'Human',
  3: 'Coin',
  4: 'KretyPost',
  5: 'Pebble',
  6: 'Car',
  7: 'Playing Card',
  8: 'Dog Tag',
  9: 'Jigsaw',
  10: 'Easter Egg',
}

/**
 * Map GeoKret type names/IDs to Bootstrap badge colors
 */
export function getGkTypeBadgeClass(gkType) {
  if (!gkType) return 'bg-secondary'

  let typeId = typeof gkType === 'number' ? gkType : null
  let typeName = typeof gkType === 'string' ? gkType : GEOKRETY_TYPES[gkType]

  const colorsByType = {
    0: 'bg-primary',
    1: 'bg-info',
    2: 'bg-success',
    3: 'bg-warning text-dark',
    4: 'bg-danger',
    5: 'bg-secondary',
    6: 'bg-primary',
    7: 'bg-info',
    8: 'bg-success',
    9: 'bg-warning text-dark',
    10: 'bg-danger',
  }

  if (typeId !== null && colorsByType[typeId]) {
    return colorsByType[typeId]
  }

  const nameColors = {
    Traditional: 'bg-primary',
    'Book/CD/DVD': 'bg-info',
    Human: 'bg-success',
    Coin: 'bg-warning text-dark',
    KretyPost: 'bg-danger',
    Pebble: 'bg-secondary',
    Car: 'bg-primary',
    'Playing Card': 'bg-info',
    'Dog Tag': 'bg-success',
    Jigsaw: 'bg-warning text-dark',
    'Easter Egg': 'bg-danger',
  }

  return nameColors[typeName] || 'bg-secondary'
}

/**
 * Get tooltip text for move types
 */
export function getMoveTypeTooltip(moveType) {
  if (!moveType) return ''

  const tooltips = {
    drop: 'Left in a cache',
    grab: 'Taken from a cache',
    take: 'Taken from a cache',
    catch: 'Found in a cache',
    dip: 'Visited without taking',
    seen: 'Encountered but not taken',
    move: 'Moved',
    recovered: 'Recovered by original owner',
    dropped: 'Left in a cache',
    found: 'Taken from cache',
    dipped: 'Visited without taking',
    comment: 'Just a comment: no physical move',
    archive: 'Missing for long time',
  }

  const lowerType = typeof moveType === 'string' ? moveType.toLowerCase() : ''
  return tooltips[lowerType] || 'Move action'
}
