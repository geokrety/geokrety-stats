/**
 * Composable for waypoint links and display.
 * Generates links to the GeoKrety.org cache map and external geocaching sites.
 */

/**
 * Returns a link to the GeoKrety.org map centered on a waypoint.
 * @param {string} waypoint - Waypoint code (e.g. "GC1A2B3")
 * @returns {string}
 */
export function waypointMapUrl(waypoint) {
  if (!waypoint) return '#'
  return `https://geokrety.org/mapa.php?wpt=${encodeURIComponent(waypoint)}`
}

/**
 * Returns a link to OpenCachingMap or Geocaching.com depending on prefix.
 * GC → geocaching.com, OP/OK/OZ → opencaching, OX → opencaching.de
 * @param {string} waypoint
 * @returns {string|null}
 */
export function waypointExternalUrl(waypoint) {
  if (!waypoint) return null
  const upper = waypoint.toUpperCase()
  if (upper.startsWith('GC')) {
    return `https://www.geocaching.com/geocache/${waypoint}`
  }
  if (upper.startsWith('OP') || upper.startsWith('OK') || upper.startsWith('OZ') || upper.startsWith('OX')) {
    return `https://opencaching.pl/viewcache.php?wp=${waypoint}`
  }
  return null
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
