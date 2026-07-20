import { defineStore } from 'pinia'
import { adminLogin } from '@/api/admin'

interface AuthState {
  token: string
  username: string
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem('admin_token') || '',
    username: localStorage.getItem('admin_username') || '',
  }),
  getters: {
    isAuthenticated: (s) => !!s.token,
  },
  actions: {
    // 校验后台凭据并保存登录态（token + 用户名）
    async login(username: string, password: string) {
      const { token } = await adminLogin(username, password)
      this.token = token
      this.username = username
      localStorage.setItem('admin_token', token)
      localStorage.setItem('admin_username', username)
    },
    // 清除登录态
    logout() {
      this.token = ''
      this.username = ''
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_username')
    },
  },
})
