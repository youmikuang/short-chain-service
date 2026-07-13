<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import TheNavBar from '@/components/navbar.vue'
import TheFooter from '@/components/footer.vue'
import { fetchLogs, type LogRow } from '@/api'

const rows = ref<LogRow[]>([])
const total = ref(0)
const search = ref('')
const page = ref(1)
const pageSize = ref(10)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const rangeStart = computed(() =>
  total.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1,
)
const rangeEnd = computed(() => Math.min(page.value * pageSize.value, total.value))

async function load() {
  const { items, total: t } = await fetchLogs({
    search: search.value,
    page: page.value,
    pageSize: pageSize.value,
  })
  rows.value = items
  total.value = t
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

onMounted(load)
watch([page, pageSize], load)
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
                  placeholder="Search URL or Endpoint..."
                  class="field-input logs__search-input"
                />
              </div>
              <div class="logs__range">
                <span class="material-symbols-outlined" style="font-size: 18px;">calendar_today</span>
                <span>Past 30 Days</span>
              </div>
            </div>

            <div class="logs__table-wrap">
              <table class="logs__table">
                <thead>
                  <tr class="logs__thead-row">
                    <th class="logs__th">Timestamp</th>
                    <th class="logs__th">Endpoint/URL</th>
                    <th class="logs__th">Status</th>
                    <th class="logs__th">Latency</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, i) in rows" :key="i" class="logs__row">
                    <td class="logs__td logs__td--muted">{{ row.timestamp }}</td>
                    <td class="logs__td">
                      <span class="logs__endpoint">{{ row.endpoint }}</span>
                    </td>
                    <td class="logs__td">
                      <span :class="statusClass(row.status)">{{ statusText(row.status) }}</span>
                    </td>
                    <td class="logs__td">{{ row.latency }}</td>
                  </tr>
                  <tr v-if="rows.length === 0">
                    <td colspan="4" class="logs__empty">No logs found.</td>
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

          <div class="logs__spacer"></div>
        </div>
      </div>
    </main>

    <TheFooter />
  </div>
</template>

<style src="@/styles/logs.css" scoped></style>
