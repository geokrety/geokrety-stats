/**
 * Parse an ISO 8601 timestamp string to a Date object.
 */
export function parseTimestamp(iso: string): Date {
  return new Date(iso)
}

/**
 * Format a Date or ISO string as a locale-friendly date/time.
 */
export function formatDateTime(value: string | Date, locale = 'en'): string {
  const d = typeof value === 'string' ? new Date(value) : value
  return d.toLocaleString(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/**
 * Format a relative time (e.g. "2h ago").
 */
export function relativeTime(value: string | Date): string {
  const d = typeof value === 'string' ? new Date(value) : value
  const diff = Date.now() - d.getTime()
  const seconds = Math.floor(diff / 1000)

  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  if (months < 12) return `${months}mo ago`
  return `${Math.floor(months / 12)}y ago`
}
