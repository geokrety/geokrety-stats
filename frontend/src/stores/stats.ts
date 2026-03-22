import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getGlobalStats } from '@/services/api/stats'
import { ApiError } from '@/services/api/client'
import type { GlobalStats } from '@/types/api'

export const useStatsStore = defineStore('stats', () => {
  // ── State ────────────────────────────────────────────────────────────────
  const globalStats = ref<GlobalStats | null>(null)

  const loadingStats = ref(false)

  const errorStats = ref<string | null>(null)

  // ── Actions ──────────────────────────────────────────────────────────────
  async function fetchGlobalStats(): Promise<void> {
    loadingStats.value = true
    errorStats.value = null
    try {
      globalStats.value = await getGlobalStats()
    } catch (e) {
      errorStats.value = e instanceof ApiError ? e.userMessage : 'Failed to load statistics'
      console.error(e)
    } finally {
      loadingStats.value = false
    }
  }

  async function fetchAll(): Promise<void> {
    await fetchGlobalStats()
  }

  return {
    // state
    globalStats,
    loadingStats,
    errorStats,
    // actions
    fetchGlobalStats,
    fetchAll,
  }
})
