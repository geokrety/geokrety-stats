/**
 * Map move type names to Bootstrap badge colors
 */
export function getMoveTypeBadgeClass(typeName) {
  if (!typeName) return 'bg-secondary'

  const typeColors = {
    'Move': 'bg-primary',
    'Seen': 'bg-info',
    'Recovered': 'bg-success',
    'Grabbed': 'bg-warning text-dark',
    'Dropped': 'bg-danger',
    'Archived': 'bg-secondary',
    'Found': 'bg-success',
    'Dipped': 'bg-info',
  }

  return typeColors[typeName] || 'bg-secondary'
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
