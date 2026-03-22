/**
 * useGeokrety — composable for fetching GeoKret lists with cursor-based pagination.
 */
import { ref, type Ref } from 'vue'
import { listGeokrety, searchGeokrety } from '@/services/api/geokrety'
import { ApiError } from '@/services/api/client'
import type { GeokretListItem } from '@/types/api'

export function useGeokrety(): {
  geokrety: Ref<GeokretListItem[]>
  loading: Ref<boolean>
  error: Ref<string | null>
  hasMore: Ref<boolean>
  fetch: (limit?: number) => Promise<void>
  loadMore: (limit?: number) => Promise<void>
  search: (query: string, limit?: number) => Promise<void>
  searchMore: (query: string, limit?: number) => Promise<void>
  reset: () => void
} {
  const geokrety = ref<GeokretListItem[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const hasMore = ref(false)
  let nextCursor: string | undefined

  async function fetch(limit = 20): Promise<void> {
    loading.value = true
    error.value = null
    nextCursor = undefined
    try {
      const result = await listGeokrety(limit)
      geokrety.value = result.data
      nextCursor = result.nextCursor
      hasMore.value = result.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load GeoKrety'
      console.error('[useGeokrety]', e)
    } finally {
      loading.value = false
    }
  }

  async function loadMore(limit = 20): Promise<void> {
    if (loading.value || !hasMore.value) return
    loading.value = true
    error.value = null
    try {
      const result = await listGeokrety(limit, nextCursor)
      geokrety.value = [...geokrety.value, ...result.data]
      nextCursor = result.nextCursor
      hasMore.value = result.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load more GeoKrety'
      console.error('[useGeokrety.loadMore]', e)
    } finally {
      loading.value = false
    }
  }

  async function search(query: string, limit = 20): Promise<void> {
    loading.value = true
    error.value = null
    nextCursor = undefined
    try {
      const result = await searchGeokrety(query, limit)
      geokrety.value = result.data
      nextCursor = result.nextCursor
      hasMore.value = result.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to search GeoKrety'
      console.error('[useGeokrety.search]', e)
    } finally {
      loading.value = false
    }
  }

  async function searchMore(query: string, limit = 20): Promise<void> {
    if (loading.value || !hasMore.value) return
    loading.value = true
    error.value = null
    try {
      const result = await searchGeokrety(query, limit, nextCursor)
      geokrety.value = [...geokrety.value, ...result.data]
      nextCursor = result.nextCursor
      hasMore.value = result.hasMore
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to search more GeoKrety'
      console.error('[useGeokrety.searchMore]', e)
    } finally {
      loading.value = false
    }
  }

  function reset(): void {
    geokrety.value = []
    loading.value = false
    error.value = null
    hasMore.value = false
    nextCursor = undefined
  }

  return { geokrety, loading, error, hasMore, fetch, loadMore, search, searchMore, reset }
}
