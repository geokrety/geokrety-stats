/**
 * Users API service — fetches user profile details and user-scoped lists.
 */
import { apiGet } from './client'
import type { GeokretListItem, UserListItem } from '@/types/api'

export interface UserDetails {
  id: number
  username: string
  joinedAt: string
  homeCountry?: string | null
  avatarId?: number | null
  avatarUrl?: string | null
  picturesCount: number
  ownedGeokretyCount: number
  movesCount: number
  distinctGeokretyCount: number
  activeCountriesCount: number
  lastMoveAt?: string | null
  homeCountryFlag: string
}

export async function getUserDetails(userId: number): Promise<UserDetails> {
  const response = await apiGet<UserDetails>(`/api/v3/users/${encodeURIComponent(userId)}`)
  return response.data
}

export interface UserGeokretyPage {
  data: GeokretListItem[]
  nextCursor?: string
  hasMore: boolean
}

export async function getUserOwnedGeokrety(userId: number, limit = 10, cursor?: string): Promise<UserGeokretyPage> {
  const params: Record<string, string | number> = { limit }
  if (cursor) params.cursor = cursor
  const response = await apiGet<GeokretListItem[]>(`/api/v3/users/${encodeURIComponent(userId)}/geokrety-owned`, params)
  return {
    data: response.data,
    nextCursor: response.meta.pagination?.nextCursor,
    hasMore: response.meta.pagination?.hasMore ?? false,
  }
}

export async function getUserFoundGeokrety(userId: number, limit = 10, cursor?: string): Promise<UserGeokretyPage> {
  const params: Record<string, string | number> = { limit }
  if (cursor) params.cursor = cursor
  const response = await apiGet<GeokretListItem[]>(`/api/v3/users/${encodeURIComponent(userId)}/geokrety-found`, params)
  return {
    data: response.data,
    nextCursor: response.meta.pagination?.nextCursor,
    hasMore: response.meta.pagination?.hasMore ?? false,
  }
}

// ── User list & search ──────────────────────────────────────────────────────

export interface UsersPage {
  data: UserListItem[]
  nextCursor?: string
  hasMore: boolean
}

export async function listUsers(limit = 20, cursor?: string): Promise<UsersPage> {
  const params: Record<string, string | number> = { limit }
  if (cursor) params.cursor = cursor
  const response = await apiGet<UserListItem[]>('/api/v3/users/', params)
  return {
    data: response.data,
    nextCursor: response.meta.pagination?.nextCursor,
    hasMore: response.meta.pagination?.hasMore ?? false,
  }
}

export async function searchUsers(query: string, limit = 20, cursor?: string): Promise<UsersPage> {
  const params: Record<string, string | number> = { q: query, limit }
  if (cursor) params.cursor = cursor
  const response = await apiGet<UserListItem[]>('/api/v3/users/search', params)
  return {
    data: response.data,
    nextCursor: response.meta.pagination?.nextCursor,
    hasMore: response.meta.pagination?.hasMore ?? false,
  }
}
