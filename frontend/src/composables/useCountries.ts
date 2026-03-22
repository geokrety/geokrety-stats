/**
 * useCountries — composable for fetching and managing country statistics.
 */
import { ref, type Ref } from 'vue'
import { getCountriesStats } from '@/services/api/stats'
import { ApiError } from '@/services/api/client'
import type { CountryStats } from '@/types/api'

export function useCountries(): {
  countries: Ref<CountryStats[]>
  loading: Ref<boolean>
  error: Ref<string | null>
  total: Ref<number | undefined>
  fetch: (limit?: number, offset?: number) => Promise<void>
} {
  const countries = ref<CountryStats[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref<number | undefined>(undefined)

  async function fetch(limit = 100, offset = 0): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const result = await getCountriesStats(limit, offset)
      countries.value = result.data
      total.value = result.total
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load country statistics'
      console.error('[useCountries]', e)
    } finally {
      loading.value = false
    }
  }

  return { countries, loading, error, total, fetch }
}
