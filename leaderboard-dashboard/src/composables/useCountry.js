/**
 * Composable for displaying country flags and names.
 * Uses Unicode Regional Indicator Symbols for flag emojis.
 */
import { getCountryFlag } from './useCountryFlags.js'

// Mapping of ISO alpha-2 codes to English country names (subset)
const COUNTRY_NAMES = {
  PL: 'Poland', DE: 'Germany', FR: 'France', CZ: 'Czech Republic',
  GB: 'United Kingdom', AT: 'Austria', SK: 'Slovakia', HU: 'Hungary',
  NL: 'Netherlands', BE: 'Belgium', IT: 'Italy', ES: 'Spain',
  CH: 'Switzerland', SE: 'Sweden', NO: 'Norway', DK: 'Denmark',
  FI: 'Finland', RU: 'Russia', US: 'United States', CA: 'Canada',
  AU: 'Australia', JP: 'Japan', CN: 'China', BR: 'Brazil',
  IN: 'India', ZA: 'South Africa', AR: 'Argentina', MX: 'Mexico',
  UA: 'Ukraine', RO: 'Romania', BG: 'Bulgaria', HR: 'Croatia',
  SI: 'Slovenia', LT: 'Lithuania', LV: 'Latvia', EE: 'Estonia',
  PT: 'Portugal', GR: 'Greece', TR: 'Turkey', IL: 'Israel',
  LU: 'Luxembourg', IE: 'Ireland', IS: 'Iceland', BY: 'Belarus',
  MD: 'Moldova', RS: 'Serbia', BA: 'Bosnia and Herzegovina',
  MK: 'North Macedonia', AL: 'Albania', ME: 'Montenegro',
  NZ: 'New Zealand', SG: 'Singapore', TH: 'Thailand', MY: 'Malaysia',
}

/**
 * Get country flag emoji for a country code
 * @param {string} code - ISO alpha-2 country code
 * @returns {string} flag emoji
 */
export function countryFlag(code) {
  return getCountryFlag(code)
}

/**
 * Get the English name of a country by its code
 * Falls back to displaying the code in uppercase if not found.
 * @param {string} code - ISO alpha-2 country code
 * @returns {string}
 */
export function countryName(code) {
  if (!code) return ''
  return COUNTRY_NAMES[code.toUpperCase()] || code.toUpperCase()
}

/**
 * Display a country as "🇵🇱 Poland" (flag + name)
 * @param {string} code - ISO alpha-2 country code
 * @returns {string}
 */
export function displayCountry(code) {
  if (!code) return ''
  const flag = countryFlag(code)
  const name = countryName(code)
  return flag ? `${flag} ${name}` : name
}

/**
 * Display a country as "🇵🇱 PL" (flag + short code)
 * @param {string} code - ISO alpha-2 country code
 * @returns {string}
 */
export function displayCountryShort(code) {
  if (!code) return ''
  const flag = countryFlag(code)
  return flag ? `${flag} ${code.toUpperCase()}` : code.toUpperCase()
}
