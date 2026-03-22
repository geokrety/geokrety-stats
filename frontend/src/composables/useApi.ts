/**
 * useApi — v3 API composable re-exports.
 *
 * This module re-exports types and composables from the new service layer
 * for backward compatibility. New code should import from @/types/api
 * and @/services/api directly.
 */

// Re-export types for backward compatibility
export type {
  GlobalStats,
  CountryStats,
  RecentMove,
  LeaderboardUser,
  GeokretListItem,
} from '@/types/api'

// Re-export composable functions from the new service layer
export { useGlobalStats as useStats } from '@/composables/useGlobalStats'
export { useRecentMoves as useRecentActivity } from '@/composables/useRecentMoves'
export { useLeaderboard } from '@/composables/useLeaderboard'
export { useCountries } from '@/composables/useCountries'
