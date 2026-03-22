/**
 * Stats API service — fetches global stats, country stats, and leaderboard.
 */
import { apiGet } from './client'
import type { GlobalStats, CountryStats, LeaderboardUser } from '@/types/api'

export async function getGlobalStats(): Promise<GlobalStats> {
  const response = await apiGet<GlobalStats>('/api/v3/stats/kpis')
  return response.data
}

export async function getCountriesStats(limit = 100, offset = 0): Promise<{ data: CountryStats[]; total?: number }> {
  const response = await apiGet<CountryStats[]>('/api/v3/stats/countries', { limit, offset })
  return {
    data: response.data,
    total: response.meta.pagination?.totalItems,
  }
}

export interface LeaderboardPage {
  data: LeaderboardUser[]
  nextCursor?: string
  hasMore: boolean
}

export async function getLeaderboard(limit = 10, cursor?: string): Promise<LeaderboardPage> {
  const params: Record<string, string | number> = { limit }
  if (cursor) {
    params.cursor = cursor
  }
  const response = await apiGet<LeaderboardUser[]>('/api/v3/stats/leaderboard', params)
  return {
    data: response.data,
    nextCursor: response.meta.pagination?.nextCursor,
    hasMore: response.meta.pagination?.hasMore ?? false,
  }
}
