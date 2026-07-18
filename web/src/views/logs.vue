<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import TheNavBar from '@/components/navbar.vue'
import TheFooter from '@/components/footer.vue'
import { fetchLogs, type LogRow } from '@/api'

const route = useRoute()
const rows = ref<LogRow[]>([])
const search = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const loading = ref(false)

// 服务端分页：避免一次性拉取全量数据（ClickHouse 远程传输慢）。
// 搜索也交给后端 search 参数，前端只持有当前页数据。
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const rangeStart = computed(() =>
  total.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1,
)
const rangeEnd = computed(() => Math.min(page.value * pageSize.value, total.value))

async function load() {
  loading.value = true
  try {
    const q = search.value.trim()
    const { items, total: t } = await fetchLogs({
      page: page.value,
      pageSize: pageSize.value,
      search: q || undefined,
    })
    rows.value = items
    total.value = t
  } catch {
    // /api/logs 查询失败（如 ClickHouse 不可用）时显示空表。
    rows.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  load()
}

function statusClass(status: number) {
  return status < 400 ? 'badge badge-ok' : 'badge badge-rate'
}

function statusText(status: number) {
  if (status === 200) return '200 OK'
  if (status === 429) return '429 Rate Limit'
  return `${status} Error`
}

function downloadCsv(filename: string, data: string[][]) {
  const escape = (v: string) => `"${String(v).replace(/"/g, '""')}"`
  const csv = data.map((r) => r.map(escape).join(',')).join('\r\n')
  const blob = new Blob(['﻿' + csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

function exportCsv() {
  const header = ['Timestamp', 'Shortened URL', 'Long URL', 'IP', 'Status', 'Latency(ms)']
  // 导出当前页数据（已是服务端返回的结果，避免再次拉取全量）。
  const body = rows.value.map((r) => [
    r.timestamp,
    '/r/' + r.code,
    r.longUrl,
    r.ip,
    String(r.status),
    String(r.latency_ms ?? ''),
  ])
  downloadCsv('logs.csv', [header, ...body])
}

onMounted(() => {
  const q = route.query.q
  if (typeof q === 'string' && q.trim()) {
    search.value = q
  }
  load()
})
watch(search, onSearch)
watch(page, () => load())
watch(pageSize, () => {
  page.value = 1
  load()
})
watch(totalPages, (tp) => {
  if (page.value > tp) page.value = tp
})
</script>

<template>
  <div class="app-shell">
    <TheNavBar />

    <main class="app-main">
      <div class="logs">
        <div class="logs__inner">
          <!-- Log Table Card -->
          <div class="logs__card">
            <div class="logs__toolbar">
              <div class="logs__search">
                <span class="material-symbols-outlined">search</span>
                <input
                  v-model="search"
                  @input="onSearch"
                  type="text"
                  placeholder="Search Code"
                  class="field-input logs__search-input"
                />
              </div>
              <div class="logs__toolbar-actions">
                <div class="logs__range">
                  <span class="material-symbols-outlined" style="font-size: 18px;">calendar_today</span>
                  <span>Past 15 Days</span>
                </div>
                <button class="logs__btn-ghost" @click="exportCsv">
                  <span class="material-symbols-outlined">download</span>
                  Export
                </button>
              </div>
            </div>

            <div class="logs__table-wrap">
                <table class="logs__table">
                <thead>
                  <tr class="logs__thead-row">
                    <th class="logs__th">Timestamp</th>
                    <th class="logs__th">Shortened URL</th>
                    <th class="logs__th">Long URL</th>
                    <th class="logs__th">IP</th>
                    <th class="logs__th">Status</th>
                    <th class="logs__th">Latency</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, i) in rows" :key="i" class="logs__row">
                    <td class="logs__td logs__td--muted">{{ row.timestamp }}</td>
                    <td class="logs__td">
                      <span class="logs__endpoint">/r/{{ row.code }}</span>
                    </td>
                    <td class="logs__td logs__td--muted logs__td--break">{{ row.longUrl }}</td>
                    <td class="logs__td logs__td--muted">{{ row.ip }}</td>
                    <td class="logs__td">
                      <span :class="statusClass(row.status)">{{ statusText(row.status) }}</span>
                    </td>
                    <td class="logs__td logs__td--muted">{{ row.latency_ms }} ms</td>
                  </tr>
                  <tr v-if="rows.length === 0">
                    <td colspan="6" class="logs__empty">No logs found.</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Pagination -->
            <div class="logs__pager">
              <div class="logs__pager-left">
                <div class="logs__pager-size">
                  <span>Rows per page:</span>
                  <select v-model.number="pageSize" class="logs__select">
                    <option :value="10">10</option>
                    <option :value="25">25</option>
                    <option :value="50">50</option>
                  </select>
                </div>
                <span class="logs__count"
                  >Showing {{ rangeStart }}-{{ rangeEnd }} of {{ total }} results</span
                >
              </div>
              <div class="logs__pager-nav">
                <button
                  class="logs__page-btn"
                  :disabled="page === 1"
                  @click="page > 1 && page--"
                >
                  <span class="material-symbols-outlined">chevron_left</span>
                </button>
                <button
                  v-for="p in totalPages"
                  :key="p"
                  class="logs__page-num"
                  :class="{ 'logs__page-num--active': p === page }"
                  @click="page = p"
                >
                  {{ p }}
                </button>
                <button
                  class="logs__page-btn"
                  :disabled="page === totalPages"
                  @click="page < totalPages && page++"
                >
                  <span class="material-symbols-outlined">chevron_right</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <TheFooter />
  </div>
</template>

<style src="@/styles/logs.css" scoped></style>
