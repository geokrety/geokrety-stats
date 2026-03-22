/**
 * useGeokretDetail — composable for fetching GeoKret details and move history.
 */
import { ref, type Ref } from 'vue'
import { getGeokretDetails, getGeokretMoves, type GeokretDetails } from '@/services/api/geokrety'
import { ApiError } from '@/services/api/client'
import type { MoveRecord } from '@/types/api'

export function useGeokretDetail(): {
  geokret: Ref<GeokretDetails | null>
  loading: Ref<boolean>
  error: Ref<string | null>
  moves: Ref<MoveRecord[]>
  movesLoading: Ref<boolean>
  movesHasMore: Ref<boolean>
  fetchDetails: (gkid: string) => Promise<void>
  fetchMoves: (gkid: string, limit?: number) => Promise<void>
  loadMoreMoves: (gkid: string, limit?: number) => Promise<void>
} {
  const geokret = ref<GeokretDetails | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const moves = ref<MoveRecord[]>([])
  const movesLoading = ref(false)
  const movesHasMore = ref(false)
  let movesCursor: string | undefined

  async function fetchDetails(gkid: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      geokret.value = await getGeokretDetails(gkid)
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load GeoKret details'
      console.error('[useGeokretDetail]', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchMoves(gkid: string, limit = 20): Promise<void> {
    movesLoading.value = true
    movesCursor = undefined
    try {
      const page = await getGeokretMoves(gkid, limit)
      moves.value = page.data
      movesCursor = page.nextCursor
      movesHasMore.value = page.hasMore
    } catch (e) {
      console.error('[useGeokretDetail] moves', e)
    } finally {
      movesLoading.value = false
    }
  }

  async function loadMoreMoves(gkid: string, limit = 20): Promise<void> {
    if (movesLoading.value || !movesHasMore.value) return
    movesLoading.value = true
    try {
      const page = await getGeokretMoves(gkid, limit, movesCursor)
      moves.value = [...moves.value, ...page.data]
      movesCursor = page.nextCursor
      movesHasMore.value = page.hasMore
    } catch (e) {
      console.error('[useGeokretDetail] moves', e)
    } finally {
      movesLoading.value = false
    }
  }

  return { geokret, loading, error, moves, movesLoading, movesHasMore, fetchDetails, fetchMoves, loadMoreMoves }
}
