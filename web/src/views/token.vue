<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import TheNavBar from '@/components/navbar.vue'
import TheFooter from '@/components/footer.vue'
import {
  listApiKeys,
  createApiKey,
  revokeApiKey,
  fetchUsageTrends,
  fetchLogs,
  type ApiKey,
  type UsagePoint,
  type LogRow,
} from '@/api'

/* ---- API Keys ---- */
const keys = ref<ApiKey[]>([])
const newKeyName = ref('')
const creating = ref(false)
const copiedKey = ref(false)

// 真实 key 仅在创建那一刻返回，并写入 localStorage；之后复制都从这里读取
const storedKey = ref(localStorage.getItem('slink_api_key') || '')

async function loadKeys() {
  try {
    const r = await listApiKeys()
    keys.value = r.items
  } catch {
    keys.value = []
  }
}

async function onCreateKey() {
  const name = newKeyName.value.trim()
  if (!name || creating.value) return
  creating.value = true
  try {
    const r = await createApiKey(name)
    // The raw key is only returned once at creation time.
    storedKey.value = r.key
    localStorage.setItem('slink_api_key', r.key)
    newKeyName.value = ''
    await loadKeys()
  } catch {
    /* ignore */
  } finally {
    creating.value = false
  }
}

async function onRevoke(id: number) {
  try {
    await revokeApiKey(id)
    localStorage.removeItem('slink_api_key') // key is gone, drop the stored raw value
    storedKey.value = ''
    await loadKeys()
  } catch {
    /* ignore */
  }
}

async function copyKey() {
  const key = storedKey.value
  if (!key) return
  let ok = false
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(key)
      ok = true
    }
  } catch {
    /* fall through to legacy path */
  }
  if (!ok) {
    // Fallback for non-secure contexts where the Clipboard API is unavailable.
    const ta = document.createElement('textarea')
    ta.value = key
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    ok = document.execCommand('copy')
    document.body.removeChild(ta)
  }
  if (ok) {
    copiedKey.value = true
    setTimeout(() => (copiedKey.value = false), 2000)
  }
}

/* ---- Usage Trends (Last 7 Days) ---- */
const weekly = ref<UsagePoint[]>([])

const WEEKDAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
function weekday(day: string): string {
  const d = new Date(day)
  if (isNaN(d.getTime())) return ''
  return WEEKDAYS[d.getDay()] ?? ''
}

// Compute the max once; reused by barPct/barClass instead of re-scanning per bar.
const maxValue = computed(() =>
  weekly.value.reduce((m, b) => Math.max(m, b.value || 0), 0),
)
function barPct(value: number): number {
  const max = maxValue.value
  if (max <= 0) return 0
  const pct = Math.round((value || 0) / max * 100)
  return (value || 0) > 0 ? Math.max(pct, 8) : 0
}
function barClass(value: number): string {
  const pct = barPct(value)
  if (pct >= 100) return 'token__bar--100'
  if (pct >= 90) return 'token__bar--90'
  if (pct >= 75) return 'token__bar--70'
  if (pct >= 60) return 'token__bar--60'
  if (pct >= 45) return 'token__bar--45'
  if (pct >= 30) return 'token__bar--30'
  return 'token__bar--25'
}

// 月度用量进度（后端未返回配额/用量，这里以近 7 天点击总量近似展示）
const USAGE_QUOTA = 100000
const usageUsed = computed(() =>
  weekly.value.reduce((s, b) => s + (b.value || 0), 0),
)
const usagePct = computed(() =>
  USAGE_QUOTA <= 0 ? 0 : Math.min(100, Math.round((usageUsed.value / USAGE_QUOTA) * 100)),
)

/* ---- Access Logs (preview) ---- */
const logs = ref<LogRow[]>([])

async function loadLogs() {
  try {
    const r = await fetchLogs({ pageSize: 4 })
    logs.value = r.items
  } catch {
    logs.value = []
  }
}

function statusClass(status: number) {
  return status < 400 ? 'badge badge-ok' : 'badge badge-rate'
}

function statusText(status: number) {
  if (status === 200) return '200 OK'
  if (status === 429) return '429 Rate Limit'
  return `${status} Error`
}

onMounted(async () => {
  await loadKeys()
  try {
    weekly.value = await fetchUsageTrends()
  } catch {
    weekly.value = []
  }
  await loadLogs()
})
</script>

<template>
  <div class="app-shell">
    <TheNavBar />

    <main class="app-main">
      <div class="token">
        <div class="token__grid">
          <!-- API Token Card -->
          <div class="token__token-card card">
            <div class="token__token-head">
              <div>
                <h2 class="token__card-title">Production Key</h2>
              </div>
              <span class="material-symbols-outlined token__token-icon">key</span>
            </div>

            <div class="token__token-field">
              <label class="token__label">Token</label>
              <div class="token__token-input-row">
                <input
                  type="password"
                  :value="storedKey"
                  readonly
                  class="token__token-input"
                  placeholder="No token yet"
                />
                <button
                  class="token__copy"
                  type="button"
                  :disabled="!storedKey"
                  :title="copiedKey ? 'Copied!' : 'Copy Token'"
                  @click="copyKey"
                >
                  <span class="material-symbols-outlined">{{ copiedKey ? 'done' : 'content_copy' }}</span>
                </button>
              </div>
            </div>

            <div class="token__usage">
              <div class="token__usage-row">
                <span class="text-secondary">Usage (This Month)</span>
                <span class="text-primary">
                  {{ usageUsed.toLocaleString() }} / {{ USAGE_QUOTA.toLocaleString() }}
                </span>
              </div>
              <div class="token__progress">
                <div class="token__progress-bar" :style="{ width: usagePct + '%' }"></div>
              </div>
            </div>
          </div>

          <!-- Usage Trends -->
          <div class="token__trends card">
            <h2 class="token__card-title">Usage Trends (Last 7 Days)</h2>
            <div v-if="weekly.length" class="token__chart">
              <div
                v-for="(bar, i) in weekly"
                :key="i"
                class="token__bar"
                :class="barClass(bar.value)"
                :style="{ height: barPct(bar.value) + '%' }"
              >
                <span class="token__bar-tip">{{ bar.value.toLocaleString() }}</span>
              </div>
            </div>
            <p v-else class="token__empty-note">No usage data available yet.</p>
            <div v-if="weekly.length" class="token__chart-labels">
              <span v-for="(bar, i) in weekly" :key="i">{{ weekday(bar.day) }}</span>
            </div>
          </div>
        </div>

        <!-- Logs Section -->
        <div id="logs" class="token__logs">
          <div class="token__logs-card card">
            <div class="token__logs-bar">
              <h3 class="token__logs-bar-title">Recent Logs</h3>
              <RouterLink to="/logs" class="token__view-all">
                View All
                <span class="material-symbols-outlined" style="font-size: 18px; margin-left: 4px;">arrow_forward</span>
              </RouterLink>
            </div>
            <div class="token__table-wrap">
              <table class="token__table">
                <thead>
                  <tr class="token__thead-row">
                    <th class="token__th">Timestamp</th>
                    <th class="token__th">Shortened URL</th>
                    <th class="token__th">Status</th>
                    <th class="token__th">Latency</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, i) in logs" :key="i" class="token__row">
                    <td class="token__td token__td--muted">{{ row.timestamp }}</td>
                    <td class="token__td">
                      <span class="token__endpoint">/r/{{ row.code }}</span>
                    </td>
                    <td class="token__td">
                      <span :class="statusClass(row.status)">{{ statusText(row.status) }}</span>
                    </td>
                    <td class="token__td">{{ row.latency_ms }}ms</td>
                  </tr>
                  <tr v-if="logs.length === 0">
                    <td colspan="4" class="token__empty">No logs found.</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </main>

    <TheFooter />
  </div>
</template>

<style src="@/styles/token.css" scoped></style>
