import { defineStore } from 'pinia'
import { adminLogin } from '@/api/admin'

// 开发/演示用默认凭据，生产应接入真实账号体系
const DEFAULT_USER = 'admin'
const DEFAULT_PASS = 'admin123'

interface AuthState {
  token: string
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem('admin_token') || '',
  }),
  getters: {
    isAuthenticated: (s) => !!s.token,
  },
  actions: {
    async login(username = DEFAULT_USER, password = DEFAULT_PASS) {
      const { token } = await adminLogin(username, password)
      this.token = token
      localStorage.setItem('admin_token', token)
    },
    logout() {
      this.token = ''
      localStorage.removeItem('admin_token')
    },
    async ensureLogin() {
      if (!this.token) {
        await this.login()
      }
    },
  },
})
