<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import TheNavBar from '@/components/navbar.vue'
import TheFooter from '@/components/footer.vue'
import {
  fetchTokenKey,
  fetchUsageTrends,
  fetchLogs,
  type UsagePoint,
  type LogRow,
} from '@/api'

/* ---- token Token ---- */
const token = ref('')
const copied = ref(false)
const usage = ref(0)
const quota = ref(0)
const usagePct = computed(() => (quota.value ? Math.round((usage.value / quota.value) * 100) : 0))

async function copyToken() {
  try {
    await navigator.clipboard.writeText(token.value)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
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
  const { items } = await fetchLogs({ search: search.value, pageSize: 5 })
  logs.value = items
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
  const [key, trends] = await Promise.all([fetchTokenKey(), fetchUsageTrends()])
  token.value = key.token
  usage.value = key.usage
  quota.value = key.quota
  weekly.value = trends
  await loadLogs()
})
</script>

<template>
  <div class="app-shell">
    <TheNavBar />

    <main class="app-main">
      <div class="token">
        <div class="token__grid">
          <!-- token Token Card -->
          <div class="token__token-card card">
            <div class="token__token-head">
              <div>
                <h2 class="token__card-title">Production Key</h2>
                <span class="token-pill">
                  <span class="token-dot"></span> Active
                </span>
              </div>
              <span class="material-symbols-outlined token__token-icon">key</span>
            </div>

            <div class="token__token-field">
              <label class="token__label">Token (Masked)</label>
              <div class="token__token-input-row">
                <input type="password" :value="token" readonly class="token__token-input" />
                <button
                  class="token__copy"
                  :title="copied ? 'Copied!' : 'Copy Token'"
                  @click="copyToken"
                >
                  <span class="material-symbols-outlined">{{ copied ? 'done' : 'content_copy' }}</span>
                </button>
              </div>
            </div>

            <div class="token__usage">
              <div class="token__usage-row">
                <span class="text-secondary">Usage (This Month)</span>
                <span class="text-primary"
                  >{{ usage.toLocaleString() }} / {{ quota.toLocaleString() }}</span
                >
              </div>
              <div class="token__progress">
                <div class="token__progress-bar" :style="{ width: usagePct + '%' }"></div>
              </div>
            </div>
          </div>

          <!-- Usage Trends -->
          <div class="token__trends card">
            <h2 class="token__card-title">Usage Trends (Last 7 Days)</h2>
            <div class="token__chart">
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
            <div class="token__chart-labels">
              <span v-for="(bar, i) in weekly" :key="i">{{ bar.day }}</span>
            </div>
          </div>
        </div>

        <!-- Logs Section -->
        <div id="logs" class="token__logs">
          <div class="token__logs-card card">
            <div class="token__logs-bar">
              <h3 class="token__logs-bar-title">Recent token Activity</h3>
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
