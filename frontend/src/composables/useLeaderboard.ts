/**
 * useLeaderboard — composable for fetching leaderboard data with cursor-based pagination.
 */
import { ref, type Ref } from 'vue'
import { getLeaderboard } from '@/services/api/stats'
import { ApiError } from '@/services/api/client'
import type { LeaderboardUser } from '@/types/api'

export function useLeaderboard(): {
  users: Ref<LeaderboardUser[]>
  loading: Ref<boolean>
  error: Ref<string | null>
  hasMore: Ref<boolean>
  fetch: (limit?: number) => Promise<void>
  loadMore: (limit?: number) => Promise<void>
  reset: () => void
} {
  const users = ref<LeaderboardUser[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const hasMore = ref(false)
  let nextCursor: string | undefined

  async function fetch(limit = 20): Promise<void> {
    loading.value = true
    error.value = null
    nextCursor = undefined
    try {
      const result = await getLeaderboard(limit)
      users.value = result.data
      nextCursor = result.nextCursor
      hasMore.value = result.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load leaderboard'
      console.error('[useLeaderboard]', e)
    } finally {
      loading.value = false
    }
  }

  async function loadMore(limit = 20): Promise<void> {
    if (loading.value || !hasMore.value) return
    loading.value = true
    error.value = null
    try {
      const result = await getLeaderboard(limit, nextCursor)
      users.value = [...users.value, ...result.data]
      nextCursor = result.nextCursor
      hasMore.value = result.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load more leaderboard data'
      console.error('[useLeaderboard.loadMore]', e)
    } finally {
      loading.value = false
    }
  }

  function reset(): void {
    users.value = []
    loading.value = false
    error.value = null
    hasMore.value = false
    nextCursor = undefined
  }

  return { users, loading, error, hasMore, fetch, loadMore, reset }
}
