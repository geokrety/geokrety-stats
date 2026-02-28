/**
 * Composable for generating links and display info for GeoKrety users.
 */

/**
 * Returns the router path to a user's profile page.
 * @param {number|string} userId
 * @returns {string}
 */
export function userPath(userId) {
  return `/users/${userId}`
}

/**
 * Returns the router path to a user's point awards page.
 * @param {number|string} userId
 * @returns {string}
 */
export function userAwardsPath(userId) {
  return `/users/${userId}/awards`
}

/**
 * Format a username with a fallback
 * @param {string} username
 * @param {number|string} userId
 * @returns {string}
 */
export function displayUsername(username, userId) {
  return username || `User #${userId}`
}
