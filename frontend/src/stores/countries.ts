import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getCountriesStats } from '@/services/api/stats'
import { ApiError } from '@/services/api/client'
import type { CountryStats } from '@/types/api'

export const useCountriesStore = defineStore('countries', () => {
  // ── State ────────────────────────────────────────────────────────────────
  const countries = ref<CountryStats[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // ── Actions ──────────────────────────────────────────────────────────────
  async function fetchCountries(): Promise<void> {
    if (countries.value.length > 0) return // already loaded
    loading.value = true
    error.value = null
    try {
      const result = await getCountriesStats()
      countries.value = result.data
    } catch (e) {
      error.value = e instanceof ApiError ? e.userMessage : 'Failed to load country statistics'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  return { countries, loading, error, fetchCountries }
})
