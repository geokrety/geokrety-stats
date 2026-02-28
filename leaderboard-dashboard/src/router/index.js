import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const routes = [
  { path: '/', component: HomeView },
  { path: '/users/:id', component: () => import('../views/UserView.vue') },
  { path: '/users/:id/awards', component: () => import('../views/PointAwardsView.vue') },
  { path: '/geokrety/:id', component: () => import('../views/GeokretView.vue') },
  { path: '/stats', component: () => import('../views/StatsView.vue') },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
