export function usePointsDisplay() {
  function toNumber(value) {
    if (value === null || value === undefined || value === '') return 0
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : 0
  }

  function formatPoints(value, digits = 0, showPlus = false) {
    const amount = toNumber(value)
    const formatted = amount.toLocaleString(undefined, {
      minimumFractionDigits: digits,
      maximumFractionDigits: digits,
    })

    if (showPlus && amount > 0) return `+${formatted}`
    return formatted
  }

  function pointsClass(value) {
    const amount = toNumber(value)
    if (amount > 0) return 'text-success'
    if (amount < 0) return 'text-danger'
    return 'text-muted'
  }

  return {
    formatPoints,
    pointsClass,
  }
}
