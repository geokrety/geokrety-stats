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

/**
 * Convert a 2-letter ISO 3166-1 alpha-2 country code to its full English name
 * using the browser's built-in Intl.DisplayNames API.
 *
 * Falls back to the raw code if the code is unknown or the API is unavailable.
 */
let _displayNames: Intl.DisplayNames | null = null

export function countryCodeToName(code: string | null | undefined): string {
  if (!code) return ''
  try {
    if (!_displayNames) {
      _displayNames = new Intl.DisplayNames(['en'], { type: 'region' })
    }
    return _displayNames.of(code.toUpperCase()) ?? code
  } catch {
    return code
  }
}

/**
 * Mapping of ISO 3166-1 alpha-2 country code → BCP 47 primary language subtag.
 * Used by the multilingual markdown parser to match country-code headers
 * (e.g. `## DE` or `## 🇩🇪 de`) to a language code for browser-language matching.
 */
export const ALPHA2_TO_LANG: Record<string, string> = {
  GB: 'en', US: 'en', AU: 'en', CA: 'en', NZ: 'en', IE: 'en', IN: 'en', ZA: 'en',
  FR: 'fr', BE: 'fr',
  DE: 'de', AT: 'de', CH: 'de',
  ES: 'es', MX: 'es', AR: 'es', CO: 'es',
  PL: 'pl',
  PT: 'pt', BR: 'pt',
  IT: 'it',
  NL: 'nl',
  RU: 'ru',
  JP: 'ja',
  CN: 'zh', TW: 'zh', HK: 'zh',
  KR: 'ko',
  SE: 'sv', NO: 'nb', DK: 'da', FI: 'fi',
  CZ: 'cs', SK: 'sk', HU: 'hu', RO: 'ro', HR: 'hr',
  UA: 'uk', BY: 'be',
  TR: 'tr', GR: 'el',
  IL: 'he', SA: 'ar', EG: 'ar',
  LT: 'lt', LV: 'lv', EE: 'et',
  SI: 'sl', RS: 'sr', MK: 'mk', BA: 'bs',
  BG: 'bg',
  GE: 'ka', AM: 'hy', AZ: 'az', KZ: 'kk', KG: 'ky', MD: 'ro', UZ: 'uz',
  IS: 'is',
}
