/**
 * Convert a 2-letter ISO country code to its emoji flag.
 *
 * Each letter is offset into the Regional Indicator Symbol block (U+1F1E6..U+1F1FF).
 */
export function countryCodeToFlag(code: string | null | undefined): string {
  if (!code || code.length !== 2) return ''
  const upper = code.toUpperCase()
  const offset = 0x1f1e6 - 65 // 'A' = 65
  return String.fromCodePoint(upper.charCodeAt(0) + offset, upper.charCodeAt(1) + offset)
}
