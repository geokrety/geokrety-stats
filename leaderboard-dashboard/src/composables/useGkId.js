/**
 * GeoKrety ID (GKID) utility functions
 * GKID format: "GK" + hexadecimal ID padded to 4 digits
 * Example: 4660 becomes "GK1234"
 */

/**
 * Convert integer GK ID to GKID format
 * @param {number} id - The numeric GK ID
 * @returns {string} - GKXXXX format
 */
export function idToGkId(id) {
  if (!id || id <= 0) return 'UNKNOWN'
  const hex = id.toString(16).toUpperCase().padStart(4, '0')
  return `GK${hex}`
}

/**
 * Convert GKID format to integer
 * @param {string} gkid - GKXXXX format
 * @returns {number} - Numeric ID
 */
export function gkIdToNumber(gkid) {
  if (!gkid || typeof gkid !== 'string') return 0
  const hex = gkid.replace(/^GK/, '')
  return parseInt(hex, 16)
}

/**
 * Format a GK ID for display (both numeric and GKID)
 * @param {number} id - The numeric GK ID
 * @returns {string} - "GKXXXX (#12345)" format
 */
export function formatGkId(id) {
  if (!id || id <= 0) return 'Unknown'
  return `${idToGkId(id)} (#${id})`
}
