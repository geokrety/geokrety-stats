import { computed } from 'vue'
import { getCountryFlag } from './useCountryFlags.js'

export function useUserCard(user) {
  const username = computed(() => user.username || user.user_username || 'Unknown')
  const userId = computed(() => user.user_id || user.id)
  const homeCountry = computed(() => user.home_country || user.user_home_country)
  const stats = computed(() => ({
    points: user.total_points || user.points || 0,
    moves: user.total_moves || user.moves || 0,
    geokrety: user.shared_geokrety_count || user.total_gks || 0,
    rank: user.rank || user.user_rank || 0,
    achievements: user.achievements_count || 0
  }))

  const flag = computed(() => homeCountry.value ? getCountryFlag(homeCountry.value) : '')
  const profileUrl = computed(() => `https://geokrety.org/mypage.php?userid=${userId.value}`)
  const leaderboardUrl = computed(() => `/users/${userId.value}`)

  return {
    username,
    userId,
    homeCountry,
    stats,
    flag,
    profileUrl,
    leaderboardUrl
  }
}
