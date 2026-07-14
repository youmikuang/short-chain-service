import { ref } from 'vue'
import { defineStore } from 'pinia'
import {
  login as apiLogin,
  register as apiRegister,
  githubAuthUrl,
  fetchProfile,
  logout as apiLogout,
  deriveProfile,
  type Profile,
} from '@/api'

const TOKEN_KEY = 'slink_token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<Profile | null>(null)

  function applySession(tok: string, prof: Profile) {
    token.value = tok
    user.value = prof
    localStorage.setItem(TOKEN_KEY, tok)
  }

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
    if (provider === 'github') {
      const { url } = await githubAuthUrl(window.location.origin + '/login')
      window.location.href = url
      return
    }
    const email = (credentials?.email ?? '').trim()
    const password = credentials?.password ?? ''
    // The web auto-creates an account on first use, so fall back to register
    // when login fails (e.g. the user does not exist yet).
    let res
    try {
      res = await apiLogin(email, password)
    } catch {
      res = await apiRegister(email, password)
    }
    applySession(res.token, deriveProfile(res, email))
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
