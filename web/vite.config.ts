import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// Go backend targets (override with env vars if they differ).
//   - slink-api  (open API + user system) listens on :8888  → /api, /r
//   - admin-api      (management backend)        listens on :8889 → /admin
const API_TARGET = process.env.VITE_API_TARGET || 'http://localhost:8888'
const ADMIN_TARGET = process.env.VITE_ADMIN_TARGET || 'http://localhost:8889'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), vueDevTools()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      // Dedicated handle for static assets (images, fonts, icons…).
      // Reference them as `@assets/foo.png` instead of fragile relative paths.
      '@assets': fileURLToPath(new URL('./src/assets', import.meta.url)),
    },
  },
  server: {
    proxy: {
      '/api': { target: API_TARGET, changeOrigin: true },
      '/r': { target: API_TARGET, changeOrigin: true },
      '/admin': { target: ADMIN_TARGET, changeOrigin: true },
    },
  },
})
