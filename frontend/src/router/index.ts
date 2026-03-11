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
      path: '/stats',
      redirect: '/',
    },
    {
      path: '/leaderboard',
      redirect: '/',
    },
    {
      path: '/about',
      redirect: '/',
    },
  ],
})

export default router
