import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import { mockApi } from './vite.mock'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
    mockApi(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      // Dedicated handle for static assets (images, fonts, icons…).
      // Reference them as `@assets/foo.png` instead of fragile relative paths.
      '@assets': fileURLToPath(new URL('./src/assets', import.meta.url)),
    },
  },
})
