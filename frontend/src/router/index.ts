import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    return { top: 0, behavior: 'smooth' }
  },
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/countries',
      name: 'countries',
      component: () => import('../views/CountriesView.vue'),
    },
    {
      path: '/countries/:code',
      name: 'country-detail',
      component: () => import('../views/CountryDetailView.vue'),
    },
    {
      path: '/geokrety',
      name: 'geokrety',
      component: () => import('../views/GeokretyView.vue'),
    },
    {
      path: '/geokrety/:gkid',
      name: 'geokret-detail',
      component: () => import('../views/GeokretDetailView.vue'),
    },
    {
      path: '/leaderboard',
      name: 'leaderboard',
      component: () => import('../views/LeaderboardView.vue'),
    },
    {
      path: '/recent-moves',
      name: 'recent-moves',
      component: () => import('../views/RecentMovesView.vue'),
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('../views/UsersView.vue'),
    },
    {
      path: '/users/:id',
      name: 'user-profile',
      component: () => import('../views/UserProfileView.vue'),
    },
    {
      path: '/stats',
      redirect: '/',
    },
    {
      path: '/about',
      redirect: '/',
    },
  ],
})

export default router
