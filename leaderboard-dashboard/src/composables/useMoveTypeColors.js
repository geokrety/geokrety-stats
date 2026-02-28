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
 * Map GeoKret type names to Bootstrap badge colors
 */
export function getGkTypeBadgeClass(gkTypeName) {
  if (!gkTypeName) return 'bg-secondary'

  const gkTypeColors = {
    'Traditional': 'bg-primary',
    'Traditional LPC': 'bg-primary',
    'Micro': 'bg-info',
    'Mystery/Travel Bug': 'bg-warning text-dark',
    'Letterbox': 'bg-success',
    'Event': 'bg-danger',
    'Webcam': 'bg-purple',
    'Virtual': 'bg-secondary',
  }

  return gkTypeColors[gkTypeName] || 'bg-secondary'
}
