/**
 * Composable for waypoint links and display.
 * Generates links to the GeoKrety.org cache map and external geocaching sites.
 */

/**
 * Returns a link to view waypoint on map (Vue route).
 * @param {string} waypoint - Waypoint code (e.g. "GC1A2B3")
 * @returns {string}
 */
export function waypointMapUrl(waypoint) {
  if (!waypoint) return '#'
  return `/map/${encodeURIComponent(waypoint)}`
}

/**
 * Returns GeoKrety.org go2geo link for any waypoint.
 * @param {string} waypoint
 * @returns {string|null}
 */
export function waypointExternalUrl(waypoint) {
  if (!waypoint) return null
  return `https://geokrety.org/go2geo/${encodeURIComponent(waypoint)}`
}

/**
 * Formats a waypoint for display. Returns the code with an icon if known.
 * @param {string} waypoint
 * @returns {string}
 */
export function displayWaypoint(waypoint) {
  if (!waypoint) return '—'
  return waypoint.toUpperCase()
}

/**
 * Returns a tooltip string for a waypoint.
 * @param {string} waypoint
 * @returns {string}
 */
export function waypointTooltip(waypoint) {
  if (!waypoint) return ''
  const upper = waypoint.toUpperCase()
  if (upper.startsWith('GC')) return `Geocaching.com waypoint: ${upper}`
  if (upper.startsWith('OP') || upper.startsWith('OK') || upper.startsWith('OZ') || upper.startsWith('OX')) {
    return `OpenCaching waypoint: ${upper}`
  }
  return `Waypoint: ${upper}`
}
