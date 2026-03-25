import { Activity, BarChart3, Globe, Home, Package, Users } from 'lucide-vue-next'
import type { Component } from 'vue'

export interface NavigationItem {
  title: string
  to: string
  icon: Component
}

export const PRIMARY_NAV_ITEMS: NavigationItem[] = [
  { title: 'Home', to: '/', icon: Home },
  { title: 'Countries', to: '/countries', icon: Globe },
  { title: 'GeoKrety', to: '/geokrety', icon: Package },
  { title: 'Leaderboard', to: '/leaderboard', icon: BarChart3 },
  { title: 'Recent Moves', to: '/recent-moves', icon: Activity },
  { title: 'Users', to: '/users', icon: Users },
]
