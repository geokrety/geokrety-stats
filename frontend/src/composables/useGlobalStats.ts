/**
 * useGlobalStats — composable for fetching and managing global statistics.
 */
import { ref, computed, type ComputedRef, type Ref } from 'vue'
import { getGlobalStats } from '@/services/api/stats'
import { ApiError } from '@/services/api/client'
import type { GlobalStats } from '@/types/api'

export function useGlobalStats(): {
  stats: Ref<GlobalStats | null>
  loading: Ref<boolean>
  error: Ref<string | null>
  totalGeokrety: ComputedRef<number>
  totalMoves: ComputedRef<number>
  fetch: () => Promise<void>
} {
  const stats = ref<GlobalStats | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const totalGeokrety = computed(() => stats.value?.totalGeokrety ?? 0)
  const totalMoves = computed(() => stats.value?.totalMoves ?? 0)

  async function fetch(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      stats.value = await getGlobalStats()
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load statistics'
      console.error('[useGlobalStats]', e)
    } finally {
      loading.value = false
    }
  }

  return { stats, loading, error, totalGeokrety, totalMoves, fetch }
}
