import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/home.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/login.vue'),
    },
    {
      path: '/token',
      name: 'token',
      component: () => import('@/views/token.vue'),
    },
    {
      path: '/logs',
      name: 'logs',
      component: () => import('@/views/logs.vue'),
    },
    {
      path: '/urls',
      name: 'urls',
      component: () => import('@/views/urls.vue'),
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/settings.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
  scrollBehavior(to, _from, _savedPosition) {
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return { top: 0 }
  },
})

export default router
