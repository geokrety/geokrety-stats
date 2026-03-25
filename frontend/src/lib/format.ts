export function formatNumber(value: number | null | undefined, maximumFractionDigits = 0): string {
  if (value === null || value === undefined) return '—'
  return value.toLocaleString('en-US', { maximumFractionDigits })
}
