/**
 * Composable for generating links and display info for GeoKrety items.
 */

/**
 * Returns the router path to a GeoKret's detail page.
 * @param {number|string} gkId - The numeric GK ID
 * @returns {string}
 */
export function geokretyPath(gkId) {
  return `/geokrety/${gkId}`
}

/**
 * Returns the router path to a GeoKret's detail page with a specific tab.
 * @param {number|string} gkId
 * @param {string} tab - one of: overview, moves, countries, related-users, points
 * @returns {string}
 */
export function geokretyTabPath(gkId, tab) {
  return `/geokrety/${gkId}#${tab}`
}

/**
 * Format a GK display name: "GKXXXX – Name" or just the hex ID
 * @param {string} gkHexId - e.g. "GK1234"
 * @param {string} gkName  - optional name
 * @returns {string}
 */
export function displayGkName(gkHexId, gkName) {
  if (gkName && gkHexId) return `${gkHexId} – ${gkName}`
  return gkHexId || gkName || 'Unknown GK'
}

/**
 * Returns a tooltip string for a GK with its ID and name.
 * @param {string} gkHexId
 * @param {string} gkName
 * @returns {string}
 */
export function gkTooltip(gkHexId, gkName) {
  if (gkName && gkHexId) return `${gkHexId}: ${gkName}`
  return gkHexId || gkName || ''
}
