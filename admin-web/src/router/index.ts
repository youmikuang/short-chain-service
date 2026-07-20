import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/login.vue'),
    },
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

// 登录态守卫：未登录跳转登录页，已登录访问登录页则跳回后台首页
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!auth.isAuthenticated && to.name !== 'login') {
    return { name: 'login' }
  }
  if (auth.isAuthenticated && to.name === 'login') {
    return { name: 'dashboard' }
  }
})

export default router
