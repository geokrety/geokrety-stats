/**
 * Map move type names to Bootstrap badge colors
 * Handles both API format (drop, grab, take, etc.) and display format
 */
export function getMoveTypeBadgeClass(typeName) {
  if (!typeName) return 'bg-secondary'

  const lowerType = typeof typeName === 'string' ? typeName.toLowerCase() : typeName

  // API format move types (from backend: drop, grab, take, etc.)
  const typeColors = {
    'drop': 'bg-danger',
    'grab': 'bg-warning text-dark',
    'take': 'bg-info',
    'catch': 'bg-success',
    'dip': 'bg-primary',
    'seen': 'bg-secondary',
    'move': 'bg-primary',
    'recovered': 'bg-success',
    'dropped': 'bg-danger',
    'found': 'bg-success',
    'dipped': 'bg-info',
  }

  return typeColors[lowerType] || 'bg-secondary'
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

  // Support both numeric IDs and string names
  let typeId = typeof gkType === 'number' ? gkType : null
  let typeName = typeof gkType === 'string' ? gkType : GEOKRETY_TYPES[gkType]

  // Map type IDs to Bootstrap colors
  const colorsByType = {
    0: 'bg-primary',      // Traditional
    1: 'bg-info',         // Book/CD/DVD
    2: 'bg-success',      // Human
    3: 'bg-warning text-dark', // Coin
    4: 'bg-danger',       // KretyPost
    5: 'bg-secondary',    // Pebble
    6: 'bg-primary',      // Car
    7: 'bg-info',         // Playing Card
    8: 'bg-success',      // Dog Tag
    9: 'bg-warning text-dark', // Jigsaw
    10: 'bg-danger',      // Easter Egg
  }

  // Try numeric ID first, fallback to name-based mapping
  if (typeId !== null && colorsByType[typeId]) {
    return colorsByType[typeId]
  }

  // Fallback name-based mapping for backward compatibility
  const nameColors = {
    'Traditional': 'bg-primary',
    'Book/CD/DVD': 'bg-info',
    'Human': 'bg-success',
    'Coin': 'bg-warning text-dark',
    'KretyPost': 'bg-danger',
    'Pebble': 'bg-secondary',
    'Car': 'bg-primary',
    'Playing Card': 'bg-info',
    'Dog Tag': 'bg-success',
    'Jigsaw': 'bg-warning text-dark',
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
    'drop': 'Dropped at a location',
    'grab': 'Picked up from a location',
    'take': 'Picked up from a location',
    'catch': 'Caught/found',
    'dip': 'Visited without taking',
    'seen': 'Seen in photo',
    'move': 'Moved',
    'recovered': 'Recovered by original owner',
    'dropped': 'Dropped at a location',
    'found': 'Found/picked up',
    'dipped': 'Visited without taking',
  }

  const lowerType = typeof moveType === 'string' ? moveType.toLowerCase() : ''
  return tooltips[lowerType] || 'Move action'
}
