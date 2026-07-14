import { http } from './client'

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------
export interface LoginResp {
  token: string
}
export function adminLogin(username: string, password: string) {
  return http.post<LoginResp>('/login', { username, password })
}

// ---------------------------------------------------------------------------
// Dashboard
// ---------------------------------------------------------------------------
export interface KpiItem {
  key: string
  label: string
  value: string
  badge: string
}
export interface TrafficPoint {
  date: string
  value: number
}
export interface AdminActionItem {
  title: string
  meta: string
  time: string
}
export interface DashboardResp {
  kpis: KpiItem[]
  traffic: TrafficPoint[]
  actions: AdminActionItem[]
}
export function getDashboard() {
  return http.get<DashboardResp>('/dashboard')
}

// ---------------------------------------------------------------------------
// Links
// ---------------------------------------------------------------------------
export interface LinkItem {
  code: string
  long_url: string
  short_url: string
  clicks: number
  status: number
  user_name: string
  user_email: string
  created_at: string
}
export interface ListLinksResp {
  total: number
  items: LinkItem[]
}
export function listLinks(page = 1, size = 10) {
  return http.get<ListLinksResp>(`/links?page=${page}&size=${size}`)
}

// ---------------------------------------------------------------------------
// Blacklist
// ---------------------------------------------------------------------------
export interface BlacklistItem {
  domain: string
  reason: string
  attempts: number
  created_at: string
}
export interface ListBlacklistResp {
  total: number
  items: BlacklistItem[]
}
export function listBlacklist(page = 1, size = 10) {
  return http.get<ListBlacklistResp>(`/blacklist?page=${page}&size=${size}`)
}
export function addBlacklist(domain: string, reason: string) {
  return http.post<{ ok: boolean }>('/blacklist', { domain, reason })
}

// ---------------------------------------------------------------------------
// Tokens
// ---------------------------------------------------------------------------
export interface TokenItem {
  id: number
  token_id: string
  user_name: string
  user_email: string
  usage_limit: number
  remaining: number
  created_at: string
  status: number
}
export interface ListTokensResp {
  total: number
  items: TokenItem[]
}
export function listTokens(page = 1, size = 10) {
  return http.get<ListTokensResp>(`/tokens?page=${page}&size=${size}`)
}
export function provisionToken(userId: number, name: string, quota: number) {
  return http.post<{ ok: boolean; token_id: string; token: string }>('/tokens', {
    user_id: userId,
    name,
    quota,
  })
}
export function revokeToken(id: number) {
  return http.post<{ ok: boolean }>('/tokens/revoke', { id })
}
