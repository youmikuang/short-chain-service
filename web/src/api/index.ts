// Real API client for the Go backend (go-zero).
//
// In development, requests to /api, /r and /admin are proxied to the Go
// services by Vite (see vite.config.ts). In production these prefixes are
// served by Nginx. No mock data is used anywhere.

const API_BASE = (import.meta.env as Record<string, string>).VITE_API_BASE ?? ''
const TOKEN_KEY = 'slink_token'
const API_KEY_KEY = 'slink_api_key'

export interface Profile {
  fullName: string
  initials: string
  email: string
  githubId: string
  avatar: string
}

// Build a display Profile from the backend's { nickname, email } shape.
export function deriveProfile(
  p: { nickname?: string; email?: string },
  emailFallback = '',
): Profile {
  const name = (p.nickname || emailFallback || 'User').trim()
  const initials =
    name
      .split(/\s+/)
      .map((w) => w[0])
      .join('')
      .slice(0, 2)
      .toUpperCase() || name.slice(0, 1).toUpperCase() || 'U'
  return {
    fullName: name,
    initials,
    email: p.email || emailFallback,
    githubId: '',
    avatar: '',
  }
}

export interface ApiKey {
  id: number
  name: string
  status: number
  createdAt: string
}

export interface ShortLink {
  code: string
  shortUrl: string
  longUrl: string
  createdAt: string
  clicks?: number
}

export interface UsagePoint {
  day: string
  value: number
}

export interface LogRow {
  timestamp: string
  endpoint: string
  status: number
  latency: string
}

export interface SettingsData {
  emailNotif: boolean
  securityAlerts: boolean
  marketingComm: boolean
}

export interface FetchLogsParams {
  search?: string
  page?: number
  pageSize?: number
}

export interface FetchLogsResult {
  items: LogRow[]
  total: number
}

function authHeaders(): Record<string, string> {
  const token = localStorage.getItem(TOKEN_KEY)
  return token ? { Authorization: `Bearer ${token}` } : {}
}

function apiKeyHeader(): Record<string, string> {
  const key = localStorage.getItem(API_KEY_KEY)
  return key ? { 'X-API-Key': key } : {}
}

async function rawRequest(
  path: string,
  init?: RequestInit,
  extraHeaders: Record<string, string> = {},
): Promise<Response> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...authHeaders(),
    ...extraHeaders,
    ...((init?.headers as Record<string, string>) ?? {}),
  }
  return fetch(`${API_BASE}${path}`, { ...init, headers })
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await rawRequest(path, init)
  if (res.status === 401) {
    localStorage.removeItem(TOKEN_KEY)
    if (window.location.pathname !== '/login') window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) throw new Error(`API ${path} failed: ${res.status}`)
  return res.json() as Promise<T>
}

// --- Auth ---------------------------------------------------------------

export interface LoginResult {
  token: string
  userId: number
  nickname: string
}

export function login(
  email: string,
  password: string,
): Promise<LoginResult> {
  return request<LoginResult>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export function register(
  email: string,
  password: string,
): Promise<LoginResult> {
  return request<LoginResult>('/api/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export function githubAuthUrl(redirect: string): Promise<{ url: string }> {
  const qs = new URLSearchParams({ redirect })
  return request<{ url: string }>(`/api/auth/github?${qs.toString()}`)
}

export function logout(): Promise<{ ok: boolean }> {
  // The backend has no logout endpoint; clearing the local session is enough.
  return Promise.resolve({ ok: true })
}

// --- Profile ------------------------------------------------------------

export function fetchProfile(): Promise<Profile> {
  return request<{ user_id: number; email: string; nickname: string }>(
    '/api/profile',
  ).then((p) => deriveProfile(p, p.email))
}

export interface SaveProfileParams {
  nickname: string
  email: string
}

// NOTE: the backend does not yet implement profile updates; kept so the UI
// can call the (future) real endpoint instead of using mock data.
export function saveProfile(params: SaveProfileParams): Promise<Profile> {
  return request<Profile>('/api/profile', {
    method: 'POST',
    body: JSON.stringify(params),
  })
}

// --- API Keys -----------------------------------------------------------

export function listApiKeys(): Promise<{ items: ApiKey[] }> {
  return request<{ items: ApiKey[] }>('/api/keys')
}

export function createApiKey(
  name: string,
): Promise<{ key: string; name: string; id: number }> {
  return request('/api/keys', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

export function revokeApiKey(id: number): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>(`/api/keys/${id}`, { method: 'DELETE' })
}

// --- Short links --------------------------------------------------------

export function createShortLink(longUrl: string): Promise<ShortLink> {
  // Short-link creation requires an API key (X-API-Key); an anonymous call
  // returns 401, which must NOT trigger the global login redirect.
  return rawRequest(
    '/api/short-links',
    {
      method: 'POST',
      body: JSON.stringify({ long_url: longUrl }),
    },
    apiKeyHeader(),
  ).then((res) => {
    if (!res.ok) throw new Error(`shorten failed: ${res.status}`)
    return res.json().then((r: { code: string; long_url: string }) => ({
      code: r.code,
      shortUrl: r.code,
      longUrl: r.long_url,
      createdAt: '-',
    }))
  })
}

export function fetchLinks(): Promise<ShortLink[]> {
  const qs = new URLSearchParams({ page: '1', size: '1000' })
  return request<{
    total: number
    items: Array<{ code: string; long_url: string; clicks: number; status: number }>
  }>(`/admin/api/links?${qs.toString()}`).then((r) =>
    r.items.map((it) => ({
      code: it.code,
      shortUrl: it.code,
      longUrl: it.long_url,
      createdAt: '-',
      clicks: it.clicks,
    })),
  )
}

// --- Settings / password / usage / logs ---------------------------------
// All wired to real backend endpoints (no mock data).

export function fetchSettings(): Promise<SettingsData> {
  return request<SettingsData>('/api/settings')
}

export interface SaveSettingsParams {
  emailNotif: boolean
  securityAlerts: boolean
  marketingComm: boolean
}

export function saveSettings(
  params: SaveSettingsParams,
): Promise<SettingsData> {
  return request<SettingsData>('/api/settings', {
    method: 'PUT',
    body: JSON.stringify(params),
  })
}

export interface UpdatePasswordParams {
  currentPassword: string
  newPassword: string
}

export function updatePassword(
  params: UpdatePasswordParams,
): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>('/api/profile/password', {
    method: 'POST',
    body: JSON.stringify(params),
  })
}

export function fetchUsageTrends(): Promise<UsagePoint[]> {
  return request<UsagePoint[]>('/api/usage-trends')
}

export function fetchLogs(params: FetchLogsParams = {}): Promise<FetchLogsResult> {
  const qs = new URLSearchParams()
  if (params.search) qs.set('search', params.search)
  if (params.page) qs.set('page', String(params.page))
  if (params.pageSize) qs.set('pageSize', String(params.pageSize))
  const query = qs.toString()
  return request<FetchLogsResult>(`/api/logs${query ? `?${query}` : ''}`)
}
