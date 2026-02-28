/**
 * Map ISO alpha-2 country codes to flag emojis
 * Uses Unicode Regional Indicator Symbols
 */
export function getCountryFlag(countryCode) {
  if (!countryCode || countryCode.length !== 2) return ''
  
  // Convert country code to flag emoji
  // Each letter becomes its regional indicator symbol
  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map(char => 127397 + char.charCodeAt(0))
  
  return String.fromCodePoint(...codePoints)
}

/**
 * Get flag and country name formatted for display
 */
export function formatCountryWithFlag(countryCode, countryName) {
  const flag = getCountryFlag(countryCode)
  return flag ? `${flag} ${countryName || countryCode}` : countryName || countryCode
}
