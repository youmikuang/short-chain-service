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
      // 让 GitHub 的 redirect_uri 指向后端回调地址（与 GitHub App 中登记的
      // Authorization callback URL 一致）。后端完成 OAuth 交换后会 302 跳转回
      // /login?token=...，由 finishGithubLogin 读取并落库。
      const { url } = await githubAuthUrl(
        window.location.origin + '/api/auth/github/callback',
      )
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
    // 注册响应带回的 API Key 存入本地（web 创建短链时随请求带上，但网关按 JWT 鉴权，key 不参与校验）
    if (res.apiKey) {
      localStorage.setItem('slink_api_key', res.apiKey)
    }
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

  // 后端 GitHub 回调 302 跳回 /login?token=...&user_id=...&nickname=... 后，
  // 由 /login 页面读取并落库，再回源拉取完整 profile。
  async function finishGithubLogin(tokenStr: string, nickname: string) {
    applySession(tokenStr, deriveProfile({ nickname }))
    try {
      user.value = await fetchProfile()
    } catch {
      /* 保留由 nickname 推导出的基础 profile */
    }
  }

  return { user, token, init, login, logout, finishGithubLogin }
})
