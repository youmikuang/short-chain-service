<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
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
const createdKey = ref<string | null>(null)
const creating = ref(false)
const copiedKey = ref(false)

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
    createdKey.value = r.key
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
    await loadKeys()
  } catch {
    /* ignore */
  }
}

async function copyKey() {
  if (!createdKey.value) return
  try {
    await navigator.clipboard.writeText(createdKey.value)
    copiedKey.value = true
    setTimeout(() => (copiedKey.value = false), 2000)
  } catch {
    /* ignore */
  }
}

/* ---- Usage Trends (Last 7 Days) ---- */
const weekly = ref<UsagePoint[]>([])

function barClass(value: number): string {
  const supported = [25, 30, 45, 60, 70, 90, 100]
  return supported.includes(value) ? `token__bar--${value}` : 'token__bar--60'
}

/* ---- Access Logs (preview) ---- */
const logs = ref<LogRow[]>([])
const search = ref('')

async function loadLogs() {
  try {
    const r = await fetchLogs({ search: search.value, pageSize: 5 })
    logs.value = r.items
  } catch {
    logs.value = []
  }
}

watch(search, loadLogs)

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
          <!-- API Keys Card -->
          <div class="token__token-card card">
            <div class="token__token-head">
              <div>
                <h2 class="token__card-title">API Keys</h2>
                <span class="token-pill">
                  <span class="token-dot"></span> {{ keys.length ? 'Active' : 'None' }}
                </span>
              </div>
              <span class="material-symbols-outlined token__token-icon">key</span>
            </div>

            <div v-if="createdKey" class="token__token-field">
              <label class="token__label">New Key (copy now, shown once)</label>
              <div class="token__token-input-row">
                <input type="text" :value="createdKey" readonly class="token__token-input" />
                <button
                  class="token__copy"
                  :title="copiedKey ? 'Copied!' : 'Copy Key'"
                  @click="copyKey"
                >
                  <span class="material-symbols-outlined">{{ copiedKey ? 'done' : 'content_copy' }}</span>
                </button>
              </div>
            </div>

            <div class="token__token-field">
              <label class="token__label">Create a new key</label>
              <div class="token__token-input-row">
                <input
                  v-model="newKeyName"
                  type="text"
                  class="token__token-input"
                  placeholder="e.g. Production"
                  @keyup.enter="onCreateKey"
                />
                <button class="token__create-btn" :disabled="creating" @click="onCreateKey">
                  {{ creating ? 'Creating…' : 'Create' }}
                </button>
              </div>
            </div>

            <div v-if="keys.length" class="token__key">
              <div class="token__key-info">
                <span class="token__key-name">{{ keys[0].name }}</span>
                <span class="token__key-meta">
                  {{ statusText(keys[0].status) }} · {{ keys[0].createdAt || '—' }}
                </span>
              </div>
              <button class="token__revoke" @click="onRevoke(keys[0].id)">Revoke</button>
            </div>
            <p v-else class="token__empty-note">No API key yet. Create one above.</p>
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
                :style="{ height: bar.value + '%' }"
              >
                <span class="token__bar-tip">{{ bar.value }}k</span>
              </div>
            </div>
            <p v-else class="token__empty-note">No usage data available yet.</p>
            <div v-if="weekly.length" class="token__chart-labels">
              <span v-for="(bar, i) in weekly" :key="i">{{ bar.day }}</span>
            </div>
          </div>
        </div>

        <!-- Logs Section -->
        <div id="logs" class="token__logs">
          <div class="token__logs-card card">
            <div class="token__logs-bar">
              <h3 class="token__logs-bar-title">Recent Activity</h3>
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
                    <th class="token__th">Endpoint</th>
                    <th class="token__th">Status</th>
                    <th class="token__th">Latency</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, i) in logs" :key="i" class="token__row">
                    <td class="token__td token__td--muted">{{ row.timestamp }}</td>
                    <td class="token__td">
                      <span class="token__endpoint">{{ row.endpoint }}</span>
                    </td>
                    <td class="token__td">
                      <span :class="statusClass(row.status)">{{ statusText(row.status) }}</span>
                    </td>
                    <td class="token__td">{{ row.latency }}</td>
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
