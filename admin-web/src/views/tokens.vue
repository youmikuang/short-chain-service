<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import AdminLayout from '@/components/AdminLayout.vue'
import UserMenu from '@/components/UserMenu.vue'
import AppFooter from '@/components/AppFooter.vue'
import { listTokens, provisionToken, revokeToken, type TokenItem } from '@/api/admin'

const items = ref<TokenItem[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(10)
const loading = ref(false)
const error = ref('')
const busyId = ref<number | null>(null)

const statusMeta: Record<number, { text: string; cls: string }> = {
  1: { text: 'Active', cls: 'bg-tertiary-container/10 text-tertiary-container border-tertiary-container/20' },
  0: { text: 'Revoked', cls: 'bg-error-container/20 text-error border-error/20' },
}
function statusOf(s: number) {
  if (s === 1) return statusMeta[1]
  if (s === 0) return statusMeta[0]
  return { text: 'Expired', cls: 'bg-surface-container-highest text-secondary border-outline-variant' }
}

function usedPct(t: TokenItem) {
  if (t.usage_limit <= 0) return 0
  const used = t.usage_limit - t.remaining
  return Math.max(0, Math.min(100, Math.round((used / t.usage_limit) * 100)))
}
function barClass(t: TokenItem) {
  if (t.remaining <= 0) return 'bg-error'
  return 'bg-primary'
}
function remainingClass(t: TokenItem) {
  if (t.remaining <= 0) return 'text-error'
  return 'text-secondary'
}

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / size.value)))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const d = await listTokens(page.value, size.value)
    items.value = d.items
    total.value = d.total
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function onProvision() {
  const userId = Number(window.prompt('User ID to provision token for', '1')) || 1
  const quota = Number(window.prompt('Monthly quota (requests)', '100000')) || 100000
  const name = window.prompt('Token name', 'admin-provisioned') || 'admin-provisioned'
  try {
    await provisionToken(userId, name, quota)
    await load()
  } catch (e) {
    window.alert('Failed: ' + (e as Error).message)
  }
}

async function onRevoke(t: TokenItem) {
  if (!window.confirm(`Revoke token ${t.token_id}?`)) return
  busyId.value = t.id
  try {
    await revokeToken(t.id)
    await load()
  } catch (e) {
    window.alert('Failed: ' + (e as Error).message)
  } finally {
    busyId.value = null
  }
}

function go(p: number) {
  if (p < 1 || p > totalPages.value) return
  page.value = p
  load()
}

onMounted(load)
</script>

<template>
  <AdminLayout>
    <!-- TopNavBar -->
    <header
      class="h-header-height sticky top-0 z-40 w-full bg-surface border-b border-outline-variant flex justify-between items-center px-gutter"
    >
      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2 text-secondary font-label-bold text-label-bold">
          <span class="material-symbols-outlined text-[18px]">home</span>
          <span>Home</span>
          <span class="material-symbols-outlined text-[14px]">chevron_right</span>
          <span class="text-primary">Token Management</span>
        </div>
      </div>
      <div class="flex items-center gap-6">
        <div class="relative hidden lg:block">
          <span class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-secondary text-[20px]">search</span>
          <input
            class="bg-surface-container-low border border-outline-variant rounded-full pl-10 pr-4 py-1 text-body-sm w-64 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/15 transition-all"
            placeholder="Global system search..."
            type="text"
          />
        </div>
        <div class="flex items-center gap-4">
          <button class="text-secondary hover:text-primary transition-colors relative">
            <span class="material-symbols-outlined">notifications</span>
            <span class="absolute top-0 right-0 w-2 h-2 bg-error rounded-full border-2 border-surface"></span>
          </button>
          <div class="h-6 w-[1px] bg-outline-variant"></div>
          <UserMenu
            name="Admin Role"
            avatar="https://lh3.googleusercontent.com/aida-public/AB6AXuCSP4HSMxpb_zzfzPJMC4JEN47ebIsfmfEREiPfh4H46AELYrnXMd_57gOBi6Kz46L_i5GRDW9bpIBujUyXsIthHA6NsYmKbuKOvrKZZDhtGPU1O_BsDiwO-FzFU38njrWY-JgHsKHJqcjx2bWVOvSBPUAUDqw9qEXePUng8fIRlfTM2M690Cyg5UhmVxFCsOtLkjirbUphIw3Jl5IwRuNOwgZlrarnY9Ul_5hrkT_XE-W5fQT7kFFAY-oUvpfRCpuAvKLkEOXJsI4"
          />
        </div>
      </div>
    </header>

    <!-- Content Canvas -->
    <div class="flex-1 overflow-y-auto p-gutter scrollbar-hide">
      <div v-if="error" class="mb-6 rounded-lg bg-error-container text-on-error-container px-4 py-3 text-body-sm">
        Failed to load tokens: {{ error }}
      </div>

      <!-- Page Header & Controls -->
      <div class="flex flex-col md:flex-row justify-between items-start md:items-center mb-8 gap-4">
        <div>
          <h2 class="font-headline-lg text-headline-lg text-on-surface">Token Infrastructure</h2>
          <p class="font-body-lg text-body-lg text-secondary">
            Monitor and control API access tokens across the system.
          </p>
        </div>
        <div class="flex items-center gap-3">
          <button
            class="bg-surface-container-lowest border border-outline-variant px-4 py-2 rounded-lg font-label-bold text-label-bold text-secondary flex items-center gap-2 hover:bg-surface-container-high transition-all active:scale-95"
            @click="load"
          >
            <span class="material-symbols-outlined text-[18px]">refresh</span>
            Refresh
          </button>
          <button
            class="bg-primary text-on-primary px-4 py-2 rounded-lg font-label-bold text-label-bold flex items-center gap-2 shadow-sm hover:brightness-110 transition-all active:scale-95"
            @click="onProvision"
          >
            <span class="material-symbols-outlined text-[18px]">add_moderator</span>
            Provision Token
          </button>
        </div>
      </div>

      <!-- Data Table Card -->
      <div class="bg-surface-container-lowest border border-outline-variant rounded-xl overflow-hidden shadow-sm">
        <div class="p-4 border-b border-outline-variant flex justify-between items-center bg-surface-container-low">
          <h3 class="font-label-bold text-label-bold text-on-surface">Active System Tokens</h3>
          <div class="flex items-center gap-2">
            <span class="font-body-sm text-body-sm text-secondary">Showing {{ items.length }} of {{ total }} tokens</span>
          </div>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-surface-container-low/50 border-b border-outline-variant">
                <th class="px-6 py-3 font-label-caps text-label-caps text-secondary">Token ID</th>
                <th class="px-6 py-3 font-label-caps text-label-caps text-secondary">User Account</th>
                <th class="px-6 py-3 font-label-caps text-label-caps text-secondary">Usage Limit</th>
                <th class="px-6 py-3 font-label-caps text-label-caps text-secondary">Remaining Quota</th>
                <th class="px-6 py-3 font-label-caps text-label-caps text-secondary">Created Date</th>
                <th class="px-6 py-3 font-label-caps text-label-caps text-secondary text-center">Status</th>
                <th class="px-6 py-3 font-label-caps text-label-caps text-secondary text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-outline-variant">
              <tr v-for="row in items" :key="row.id" class="hover:bg-secondary-container/10 transition-colors group">
                <td class="px-6 py-4 font-technical-mono text-technical-mono text-primary select-all">{{ row.token_id }}</td>
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div class="w-7 h-7 rounded-full bg-primary-container/10 flex items-center justify-center text-primary text-[12px] font-bold">
                      {{ (row.user_name || '?').slice(0, 2).toUpperCase() }}
                    </div>
                    <div>
                      <div class="font-label-bold text-label-bold text-on-surface">{{ row.user_name || '—' }}</div>
                      <div class="text-[11px] text-secondary">{{ row.user_email }}</div>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4 font-body-sm text-body-sm text-on-surface">{{ row.usage_limit.toLocaleString() }} req/mo</td>
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div class="w-24 bg-surface-container-high h-1.5 rounded-full overflow-hidden">
                      <div class="h-full" :class="barClass(row)" :style="{ width: usedPct(row) + '%' }"></div>
                    </div>
                    <span class="font-technical-mono text-technical-mono" :class="remainingClass(row)">{{ row.remaining.toLocaleString() }}</span>
                  </div>
                </td>
                <td class="px-6 py-4 font-body-sm text-body-sm text-secondary">{{ row.created_at }}</td>
                <td class="px-6 py-4 text-center">
                  <span
                    class="px-2 py-1 rounded text-[11px] font-bold uppercase tracking-tight border"
                    :class="statusOf(row.status).cls"
                    >{{ statusOf(row.status).text }}</span
                  >
                </td>
                <td class="px-6 py-4 text-right">
                  <div class="flex items-center justify-end gap-2">
                    <button
                      v-if="row.status === 1"
                      class="p-1.5 hover:bg-error-container/20 rounded-lg text-error transition-colors disabled:opacity-50"
                      :disabled="busyId === row.id"
                      title="Revoke"
                      @click="onRevoke(row)"
                    >
                      <span class="material-symbols-outlined text-[20px]">block</span>
                    </button>
                    <span v-else class="text-primary font-label-bold text-[12px]">—</span>
                  </div>
                </td>
              </tr>
              <tr v-if="!loading && !items.length">
                <td class="px-6 py-10 text-center text-secondary text-body-sm" colspan="7">No tokens found</td>
              </tr>
              <tr v-if="loading">
                <td class="px-6 py-10 text-center text-secondary text-body-sm" colspan="7">Loading…</td>
              </tr>
            </tbody>
          </table>
        </div>
        <!-- Pagination Footer -->
        <div class="p-4 border-t border-outline-variant flex justify-between items-center bg-surface-container-low/30">
          <button
            class="text-secondary font-label-bold text-label-bold flex items-center gap-1 hover:text-primary transition-colors disabled:opacity-50"
            :disabled="page <= 1"
            @click="go(page - 1)"
          >
            <span class="material-symbols-outlined text-[18px]">chevron_left</span>
            Previous
          </button>
          <div class="flex items-center gap-1">
            <button
              v-for="p in (page <= totalPages ? [page] : [])"
              :key="p"
              class="w-8 h-8 rounded font-label-bold text-label-bold transition-colors"
              :class="p === page ? 'bg-primary text-on-primary' : 'hover:bg-surface-container-high'"
            >
              {{ p }}
            </button>
          </div>
          <button
            class="text-secondary font-label-bold text-label-bold flex items-center gap-1 hover:text-primary transition-colors disabled:opacity-50"
            :disabled="page >= totalPages"
            @click="go(page + 1)"
          >
            Next
            <span class="material-symbols-outlined text-[18px]">chevron_right</span>
          </button>
        </div>
      </div>
    </div>

    <AppFooter />
  </AdminLayout>
</template>

<style scoped>
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
</style>
