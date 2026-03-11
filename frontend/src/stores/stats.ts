import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  useStats,
  useRecentActivity,
  useLeaderboard,
  type GlobalStats,
  type RecentMove,
  type LeaderboardUser,
} from '@/composables/useApi'

export const useStatsStore = defineStore('stats', () => {
  // ── State ────────────────────────────────────────────────────────────────
  const globalStats = ref<GlobalStats | null>(null)
  const recentActivity = ref<RecentMove[]>([])
  const leaderboard = ref<LeaderboardUser[]>([])

  const loadingStats = ref(false)
  const loadingActivity = ref(false)
  const loadingLeaderboard = ref(false)

  const errorStats = ref<string | null>(null)
  const errorActivity = ref<string | null>(null)
  const errorLeaderboard = ref<string | null>(null)

  // ── Actions ──────────────────────────────────────────────────────────────
  async function fetchGlobalStats() {
    loadingStats.value = true
    errorStats.value = null
    try {
      const { fetchStats } = useStats()
      globalStats.value = await fetchStats()
    } catch (e) {
      errorStats.value = 'Failed to load statistics'
      console.error(e)
    } finally {
      loadingStats.value = false
    }
  }

  async function fetchRecentActivity() {
    loadingActivity.value = true
    errorActivity.value = null
    try {
      const { fetchRecentActivity: fetch } = useRecentActivity()
      recentActivity.value = await fetch()
    } catch (e) {
      errorActivity.value = 'Failed to load recent activity'
      console.error(e)
    } finally {
      loadingActivity.value = false
    }
  }

  async function fetchLeaderboard() {
    loadingLeaderboard.value = true
    errorLeaderboard.value = null
    try {
      const { fetchLeaderboard: fetch } = useLeaderboard()
      leaderboard.value = await fetch()
    } catch (e) {
      errorLeaderboard.value = 'Failed to load leaderboard'
      console.error(e)
    } finally {
      loadingLeaderboard.value = false
    }
  }

  async function fetchAll() {
    await Promise.all([fetchGlobalStats(), fetchRecentActivity(), fetchLeaderboard()])
  }

  return {
    // state
    globalStats,
    recentActivity,
    leaderboard,
    loadingStats,
    loadingActivity,
    loadingLeaderboard,
    errorStats,
    errorActivity,
    errorLeaderboard,
    // actions
    fetchGlobalStats,
    fetchRecentActivity,
    fetchLeaderboard,
    fetchAll,
  }
})
