import { createRouter, createWebHistory } from 'vue-router'

// @ts-ignore
// @ts-ignore
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: () => import('@/views/dashboard.vue'),
    },
    {
      path: '/links',
      name: 'links',
      component: () => import('@/views/links.vue'),
    },
    {
      path: '/blacklist',
      name: 'blacklist',
      component: () => import('@/views/blacklist.vue'),
    },
    {
      path: '/tokens',
      name: 'tokens',
      component: () => import('@/views/tokens.vue'),
    },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

export default router
