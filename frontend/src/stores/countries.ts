import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useCountries, type CountryStats } from '@/composables/useApi'

export const useCountriesStore = defineStore('countries', () => {
  // ── State ────────────────────────────────────────────────────────────────
  const countries = ref<CountryStats[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // ── Actions ──────────────────────────────────────────────────────────────
  async function fetchCountries() {
    if (countries.value.length > 0) return // already loaded
    loading.value = true
    error.value = null
    try {
      const { fetchCountries: fetch } = useCountries()
      countries.value = await fetch()
    } catch (e) {
      error.value = 'Failed to load country statistics'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  return { countries, loading, error, fetchCountries }
})
