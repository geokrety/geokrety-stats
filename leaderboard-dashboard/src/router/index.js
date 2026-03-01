import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const routes = [
  { path: '/', component: HomeView },
  { path: '/users/:id', component: () => import('../views/UserView.vue') },
  {
    path: '/users/:id/awards',
    redirect: (to) => {
      const query = new URLSearchParams(to.query || {}).toString()
      const hash = query ? `#awards?${query}` : '#awards'
      return { path: `/users/${to.params.id}`, hash }
    },
  },
  { path: '/users/:id/chains', component: () => import('../views/UserChainsView.vue') },
  { path: '/geokrety', component: () => import('../views/GeokretyLeaderboardView.vue') },
  { path: '/geokrety/:id', component: () => import('../views/GeokretView.vue') },
  { path: '/geokrety/:id/chains', component: () => import('../views/GeokretChainsView.vue') },
  { path: '/moves/:id/chains', component: () => import('../views/MoveChainsView.vue') },
  { path: '/chains/:id', component: () => import('../views/ChainDetailView.vue') },
  { path: '/countries', component: () => import('../views/CountryLeaderboardView.vue') },
  { path: '/country/:country', component: () => import('../views/CountryDetailView.vue') },
  { path: '/stats', component: () => import('../views/StatsView.vue') },
  { path: '/map/:waypoint?', component: () => import('../views/MapView.vue') },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
