/**
 * Convert an integer GeoKret ID to the GK-prefixed hex format.
 * The hex part is at least 4 digits (zero-padded); longer if needed.
 */
export function intToGkid(id: number): string {
  if (!Number.isInteger(id) || id < 0) {
    throw new RangeError(`GeoKret ID must be a non-negative integer, got: ${id}`)
  }
  const hex = id.toString(16).toUpperCase().padStart(4, '0')
  return `GK${hex}`
}

/**
 * Convert a GKID string (e.g. "GK00FF" or "gk00ff") to its integer ID.
 * Returns `null` if the string is not a valid GKID.
 */
export function gkidToInt(gkid: string): number | null {
  const upper = gkid.trim().toUpperCase()
  if (!/^GK[0-9A-F]+$/.test(upper)) return null
  const value = parseInt(upper.slice(2), 16)
  return isNaN(value) ? null : value
}

/**
 * Returns true if `value` looks like a valid GKID string.
 */
export function validateGkid(value: string): boolean {
  return /^[Gg][Kk][0-9A-Fa-f]+$/.test(value.trim())
}
