import { ref } from 'vue'
import { defineStore } from 'pinia'
import { login as apiLogin, logout as apiLogout, fetchProfile, type Profile } from '@/api'

const TOKEN_KEY = 'slink_token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<Profile | null>(null)

  // Restore session on app start if a token exists.
  async function init() {
    if (!token.value) return
    try {
      user.value = await fetchProfile()
    } catch {
      token.value = null
      user.value = null
      localStorage.removeItem(TOKEN_KEY)
    }
  }

  async function login(
    provider: 'github' | 'email',
    credentials?: { email?: string; password?: string },
  ) {
    const res = await apiLogin(provider, credentials)
    token.value = res.token
    user.value = res.user
    localStorage.setItem(TOKEN_KEY, res.token)
  }

  async function logout() {
    try {
      await apiLogout()
    } catch {
      /* ignore network errors on logout */
    }
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  return { user, token, init, login, logout }
})
