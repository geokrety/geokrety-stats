/**
 * useRecentMoves — composable for fetching recent GeoKret moves with cursor-based pagination.
 */
import { ref, type Ref } from 'vue'
import { getRecentMoves } from '@/services/api/moves'
import { ApiError } from '@/services/api/client'
import type { RecentMove } from '@/types/api'

export function useRecentMoves(): {
  moves: Ref<RecentMove[]>
  loading: Ref<boolean>
  error: Ref<string | null>
  hasMore: Ref<boolean>
  fetch: (limit?: number) => Promise<void>
  loadMore: (limit?: number) => Promise<void>
  reset: () => void
} {
  const moves = ref<RecentMove[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const hasMore = ref(false)
  let nextCursor: string | undefined

  async function fetch(limit = 20): Promise<void> {
    loading.value = true
    error.value = null
    nextCursor = undefined
    try {
      const result = await getRecentMoves(limit)
      moves.value = result.data
      nextCursor = result.nextCursor
      hasMore.value = result.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load recent moves'
      console.error('[useRecentMoves]', e)
    } finally {
      loading.value = false
    }
  }

  async function loadMore(limit = 20): Promise<void> {
    if (loading.value || !hasMore.value) return
    loading.value = true
    error.value = null
    try {
      const result = await getRecentMoves(limit, nextCursor)
      moves.value = [...moves.value, ...result.data]
      nextCursor = result.nextCursor
      hasMore.value = result.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load more moves'
      console.error('[useRecentMoves.loadMore]', e)
    } finally {
      loading.value = false
    }
  }

  function reset(): void {
    moves.value = []
    loading.value = false
    error.value = null
    hasMore.value = false
    nextCursor = undefined
  }

  return { moves, loading, error, hasMore, fetch, loadMore, reset }
}
