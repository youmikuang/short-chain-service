// Mock API client.
// Talks to the Vite dev middleware defined in `vite.mock.ts`, which serves
// data from `public/api.json` over real HTTP endpoints (/api/profile,
// /api/logs, /api/auth/login, ...). Swap the base for the Go service later.

export interface Profile {
  fullName: string
  initials: string
  email: string
  githubId: string
  avatar: string
}

export interface TokenKey {
  token: string
  status: 'active' | 'disabled'
  usage: number
  quota: number
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

const BASE = import.meta.env.BASE_URL
const TOKEN_KEY = 'slink_token'

function authHeaders(): Record<string, string> {
  const token = localStorage.getItem(TOKEN_KEY)
  return token ? { Authorization: `Bearer ${token}` } : {}
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...authHeaders(),
    ...((init?.headers as Record<string, string>) ?? {}),
  }
  const res = await fetch(`${BASE}api/${path}`, { ...init, headers })
  if (res.status === 401) {
    localStorage.removeItem(TOKEN_KEY)
    if (window.location.pathname !== '/login') window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) throw new Error(`API ${path} failed: ${res.status}`)
  return res.json() as Promise<T>
}

export function fetchProfile(): Promise<Profile> {
  return request<Profile>('profile')
}

export function fetchTokenKey(): Promise<TokenKey> {
  return request<TokenKey>('tokenKey')
}

export function fetchUsageTrends(): Promise<UsagePoint[]> {
  return request<UsagePoint[]>('usage-trends')
}

export function fetchLogs(params: FetchLogsParams = {}): Promise<FetchLogsResult> {
  const qs = new URLSearchParams()
  if (params.search) qs.set('search', params.search)
  if (params.page) qs.set('page', String(params.page))
  if (params.pageSize) qs.set('pageSize', String(params.pageSize))
  const query = qs.toString()
  return request<FetchLogsResult>(`logs${query ? `?${query}` : ''}`)
}

export function fetchSettings(): Promise<SettingsData> {
  return request<SettingsData>('settings')
}

export interface SaveProfileParams {
  fullName: string
  email: string
  avatar: string
}

export function saveProfile(params: SaveProfileParams): Promise<Profile> {
  return request<Profile>('profile', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
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
  return request<{ ok: boolean }>('password', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })
}

export interface SaveSettingsParams {
  emailNotif: boolean
  securityAlerts: boolean
  marketingComm: boolean
}

export function saveSettings(
  params: SaveSettingsParams,
): Promise<SettingsData> {
  return request<SettingsData>('settings', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })
}

export interface LoginResult {
  user: Profile
  token: string
  provider: string
}

export function login(provider: 'github' | 'email'): Promise<LoginResult> {
  return request<LoginResult>('auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ provider }),
  })
}

export function logout(): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>('auth/logout', { method: 'POST' })
}

export interface ShortLink {
  code: string
  shortUrl: string
  longUrl: string
  createdAt: string
  clicks?: number
  note?: string
}

export function createShortLink(longUrl: string): Promise<ShortLink> {
  return request<ShortLink>('shorten', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url: longUrl }),
  })
}

export function fetchLinks(): Promise<ShortLink[]> {
  return request<ShortLink[]>('links')
}
