/**
 * useGkid — convert between integer GeoKret IDs and the public GKID format.
 *
 * Mapping rule (from geokrety.org):
 *   GKID = "GK" + integer.toString(16).toUpperCase().padStart(4, "0")
 *
 * Examples:
 *   intToGkid(1)      → "GK0001"
 *   intToGkid(255)    → "GK00FF"
 *   intToGkid(65535)  → "GKFFFF"
 *   gkidToInt("GK00FF") → 255
 */

export function useGkid() {
  /**
   * Convert an integer GeoKret ID to the GK-prefixed hex format.
   * The hex part is at least 4 digits (zero-padded); longer if needed.
   */
  function intToGkid(id: number): string {
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
  function gkidToInt(gkid: string): number | null {
    const upper = gkid.trim().toUpperCase()
    if (!/^GK[0-9A-F]{1,}$/.test(upper)) return null
    const hex = upper.slice(2) // strip "GK"
    const value = parseInt(hex, 16)
    return isNaN(value) ? null : value
  }

  /**
   * Returns true if `value` looks like a valid GKID string.
   */
  function isGkid(value: string): boolean {
    return /^[Gg][Kk][0-9A-Fa-f]{1,}$/.test(value.trim())
  }

  return { intToGkid, gkidToInt, isGkid }
}
