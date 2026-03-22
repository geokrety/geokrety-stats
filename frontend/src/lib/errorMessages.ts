/**
 * Map API error codes to user-friendly messages.
 */
const ERROR_MESSAGES: Record<string, string> = {
  INVALID_GKID: 'The GeoKret ID format is invalid. Please use a format like GK0001.',
  NOT_FOUND: 'The requested resource was not found.',
  VALIDATION_ERROR: 'Some input data is invalid. Please check and try again.',
  RATE_LIMITED: 'Too many requests. Please wait a moment and try again.',
  INTERNAL_ERROR: 'An unexpected error occurred. Please try again later.',
  NETWORK_ERROR: 'Unable to connect to the server. Please check your internet connection.',
  TIMEOUT: 'The request timed out. Please try again.',
}

/**
 * Get a user-friendly error message for a given API error code.
 */
export function getErrorMessage(code: string): string {
  return ERROR_MESSAGES[code] ?? 'An unexpected error occurred. Please try again later.'
}
