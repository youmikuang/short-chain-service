import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import './style.css'

// Apply persisted/system theme before mount to avoid a flash.
const saved = localStorage.getItem('theme')
const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
if (saved === 'dark' || (!saved && prefersDark)) {
  document.documentElement.classList.add('dark')
}

const app = createApp(App)

app.use(createPinia())
app.use(router)

// 启动前确保拿到 admin JWT（开发环境用默认凭据自动登录）
const auth = useAuthStore()
auth
  .ensureLogin()
  .catch((e) => console.error('[admin] auto login failed:', e))
  .finally(() => app.mount('#app'))
