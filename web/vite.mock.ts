import { readFileSync } from 'node:fs'
import { dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import type { Plugin, ViteDevServer, PreviewServer } from 'vite'

const __dirname = dirname(fileURLToPath(import.meta.url))

let _data: any = null
function readData(): any {
  if (!_data) {
    const jsonPath = fileURLToPath(new URL('./public/api.json', import.meta.url))
    _data = JSON.parse(readFileSync(jsonPath, 'utf-8'))
  }
  return _data
}

function send(res: any, body: any, status = 200) {
  res.statusCode = status
  res.setHeader('Content-Type', 'application/json')
  res.end(JSON.stringify(body))
}

function readBody(req: any): Promise<any> {
  return new Promise((resolveBody) => {
    let data = ''
    req.on('data', (chunk: string) => (data += chunk))
    req.on('end', () => {
      try {
        resolveBody(data ? JSON.parse(data) : {})
      } catch {
        resolveBody({})
      }
    })
  })
}

// Mounted at /api — req.url is the part after the prefix (e.g. "/profile").
function handler(req: any, res: any) {
  const data = readData()
  const url = new URL(req.url ?? '/', 'http://localhost')
  const path = url.pathname.replace(/^\/+/, '')
  const method = (req.method ?? 'GET').toUpperCase()

  // Token gate: every route except auth/login & auth/logout requires a
  // valid Bearer token (mock format: "mock-token-<timestamp>").
  const publicPaths = new Set(['auth/login', 'auth/logout', 'shorten'])
  if (!publicPaths.has(path)) {
    const auth = req.headers['authorization'] ?? ''
    if (!/^Bearer\s+mock-token-/.test(auth)) {
      return send(res, { error: 'Unauthorized' }, 401)
    }
  }

  if (path === 'profile' && method === 'POST') {
    return readBody(req).then((body) => {
      data.profile = {
        ...data.profile,
        fullName:
          typeof body.fullName === 'string' ? body.fullName : data.profile.fullName,
        email: typeof body.email === 'string' ? body.email : data.profile.email,
        avatar: typeof body.avatar === 'string' ? body.avatar : data.profile.avatar,
      }
      return send(res, data.profile)
    })
  }

  if (path === 'profile') return send(res, data.profile)

  if (path === 'shorten' && method === 'POST') {
    return readBody(req).then((body) => {
      const longUrl = typeof body.url === 'string' ? body.url : ''
      const code = Math.random().toString(36).slice(2, 8)
      const record = {
        code,
        shortUrl: code,
        longUrl,
        createdAt: new Date().toISOString(),
        clicks: Math.floor(Math.random() * 2000),
      }
      if (!Array.isArray(data.links)) data.links = []
      data.links.unshift(record)
      return send(res, record)
    })
  }

  if (path === 'links') {
    return send(res, Array.isArray(data.links) ? data.links : [])
  }

  if (path === 'tokenKey' || path === 'key') return send(res, data.tokenKey)
  if (path === 'usage-trends') return send(res, data.usageTrends)
  if (path === 'password' && method === 'POST') {
    return readBody(req).then(() => send(res, { ok: true }))
  }

  if (path === 'settings' && method === 'POST') {
    return readBody(req).then((body) => {
      data.settings = {
        emailNotif: !!body.emailNotif,
        securityAlerts: !!body.securityAlerts,
        marketingComm: !!body.marketingComm,
      }
      return send(res, data.settings)
    })
  }

  if (path === 'settings') return send(res, data.settings)

  if (path === 'logs') {
    const q = (url.searchParams.get('search') ?? '').trim().toLowerCase()
    let rows = data.logs as any[]
    if (q) {
      rows = rows.filter(
        (r) =>
          r.endpoint.toLowerCase().includes(q) || r.timestamp.toLowerCase().includes(q),
      )
    }
    const total = rows.length
    const pageSize = Number(url.searchParams.get('pageSize') ?? 10)
    const page = Number(url.searchParams.get('page') ?? 1)
    const items = rows.slice((page - 1) * pageSize, page * pageSize)
    return send(res, { items, total })
  }

  if (path === 'auth/login' && method === 'POST') {
    return readBody(req).then((body) => {
      let user = data.profile
      const email = typeof body.email === 'string' ? body.email.trim() : ''
      if (email) {
        const local = email.split('@')[0] || email
        const name = local
          .replace(/[._-]+/g, ' ')
          .replace(/\b\w/g, (c) => c.toUpperCase())
          .trim()
        const initials =
          name
            .split(/\s+/)
            .map((w: string) => w[0])
            .join('')
            .slice(0, 2)
            .toUpperCase() || local.slice(0, 1).toUpperCase()
        user = {
          ...data.profile,
          fullName: name || local,
          initials,
          email,
          githubId: local,
        }
        data.profile = user
      }
      return send(res, {
        user,
        token: `mock-token-${Date.now()}`,
        provider: body.provider ?? 'github',
      })
    })
  }

  if (path === 'auth/logout' && method === 'POST') {
    return send(res, { ok: true })
  }

  return send(res, { error: 'Not found' }, 404)
}

export function mockApi(): Plugin {
  const apply = (server: any) => server.middlewares.use('/api', handler)
  return {
    name: 'mock-api',
    configureServer(server: ViteDevServer) {
      apply(server)
    },
    configurePreviewServer(server: PreviewServer) {
      apply(server)
    },
  }
}
