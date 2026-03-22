/**
 * GeoKrety API service — fetches geokrety lists, search, and details.
 */
import { apiGet } from './client'
import type { GeokretListItem, MoveRecord } from '@/types/api'

export interface GeokretyPage {
  data: GeokretListItem[]
  nextCursor?: string
  hasMore: boolean
}

export async function listGeokrety(limit = 20, cursor?: string): Promise<GeokretyPage> {
  const params: Record<string, string | number> = { limit }
  if (cursor) {
    params.cursor = cursor
  }
  const response = await apiGet<GeokretListItem[]>('/api/v3/geokrety', params)
  return {
    data: response.data,
    nextCursor: response.meta.pagination?.nextCursor,
    hasMore: response.meta.pagination?.hasMore ?? false,
  }
}

export async function searchGeokrety(query: string, limit = 20, cursor?: string): Promise<GeokretyPage> {
  const params: Record<string, string | number> = { q: query, limit }
  if (cursor) {
    params.cursor = cursor
  }
  const response = await apiGet<GeokretListItem[]>('/api/v3/geokrety/search', params)
  return {
    data: response.data,
    nextCursor: response.meta.pagination?.nextCursor,
    hasMore: response.meta.pagination?.hasMore ?? false,
  }
}

export interface GeokretDetails extends GeokretListItem {
  mission?: string | null
  distanceKm: number
  cachesCount: number
  commentsHidden: boolean
}

export async function getGeokretDetails(gkid: string | number): Promise<GeokretDetails> {
  const response = await apiGet<GeokretDetails>(`/api/v3/geokrety/${encodeURIComponent(gkid)}`)
  return response.data
}

export interface GeokretMovesPage {
  data: MoveRecord[]
  nextCursor?: string
  hasMore: boolean
}

export async function getGeokretMoves(gkid: string | number, limit = 20, cursor?: string): Promise<GeokretMovesPage> {
  const params: Record<string, string | number> = { limit }
  if (cursor) params.cursor = cursor
  const response = await apiGet<MoveRecord[]>(`/api/v3/geokrety/${encodeURIComponent(gkid)}/moves`, params)
  return {
    data: response.data,
    nextCursor: response.meta.pagination?.nextCursor,
    hasMore: response.meta.pagination?.hasMore ?? false,
  }
}
