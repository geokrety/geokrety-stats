import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const routes = [
  { path: '/', component: HomeView, meta: { title: 'Live Activity' } },
  { path: '/users/:id', component: () => import('../views/UserView.vue'), meta: { title: 'User Profile' } },
  {
    path: '/users/:id/awards',
    redirect: (to) => {
      const query = new URLSearchParams(to.query || {}).toString()
      const hash = query ? `#awards?${query}` : '#awards'
      return { path: `/users/${to.params.id}`, hash }
    },
  },
  { path: '/users/:id/chains', component: () => import('../views/UserChainsView.vue'), meta: { title: 'User Interaction Chains' } },
  { path: '/geokrety', component: () => import('../views/GeokretyLeaderboardView.vue'), meta: { title: 'GeoKrety Leaderboard' } },
  { path: '/geokrety/:id', component: () => import('../views/GeokretView.vue'), meta: { title: 'GeoKret Details' } },
  { path: '/geokrety/:id/chains', component: () => import('../views/GeokretChainsView.vue'), meta: { title: 'GeoKret Evolution Chains' } },
  { path: '/moves/:id/chains', component: () => import('../views/MoveChainsView.vue'), meta: { title: 'Log Multi-Chains' } },
  { path: '/chains/:id', component: () => import('../views/ChainDetailView.vue'), meta: { title: 'Chain Discovery' } },
  { path: '/countries', component: () => import('../views/CountryLeaderboardView.vue'), meta: { title: 'Country Leaderboard' } },
  { path: '/country/:country', component: () => import('../views/CountryDetailView.vue'), meta: { title: 'Country Discovery Statistics' } },
  { path: '/stats', component: () => import('../views/StatsView.vue'), meta: { title: 'Global Statistics' } },
  { path: '/map/:waypoint?', component: () => import('../views/MapView.vue'), meta: { title: 'Interactive Map' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.afterEach((to) => {
  const baseTitle = 'GeoKrety Leaderboard'
  const routeTitle = to.meta.title || ''
  document.title = routeTitle ? `${routeTitle} | ${baseTitle}` : baseTitle
})

export default router
