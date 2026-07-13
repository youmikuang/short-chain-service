import { createApp } from 'vue'
import { createPinia } from 'pinia'

// @ts-ignore
import App from '@/App.vue'
import router from '@/router'
import '@/style.css'
import '@/styles/ui.css'

// Apply persisted/system theme before mount to avoid a flash.
const saved = localStorage.getItem('theme')
const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
if (saved === 'dark' || (!saved && prefersDark)) {
  document.documentElement.classList.add('dark')
}

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
