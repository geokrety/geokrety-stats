/**
 * Moves API service — fetches recent moves with cursor-based pagination.
 */
import { apiGet } from './client'
import type { RecentMove } from '@/types/api'

export interface RecentMovesPage {
  data: RecentMove[]
  nextCursor?: string
  hasMore: boolean
}

export async function getRecentMoves(limit = 20, cursor?: string): Promise<RecentMovesPage> {
  const params: Record<string, string | number> = { limit }
  if (cursor) {
    params.cursor = cursor
  }
  const response = await apiGet<RecentMove[]>('/api/v3/geokrety/recent-moves', params)
  return {
    data: response.data,
    nextCursor: response.meta.pagination?.nextCursor,
    hasMore: response.meta.pagination?.hasMore ?? false,
  }
}
