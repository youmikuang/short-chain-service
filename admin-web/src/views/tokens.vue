<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AdminLayout from '@/components/AdminLayout.vue'
import TopNavBar from '@/components/TopNavBar.vue'
import AppFooter from '@/components/AppFooter.vue'
import PageHeader from '@/components/PageHeader.vue'
import Card from '@/components/Card.vue'
import Pagination from '@/components/Pagination.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import ErrorBanner from '@/components/ErrorBanner.vue'
import { usePagination } from '@/composables/usePagination'
import { listTokens, provisionToken, revokeToken, type TokenItem } from '@/api/admin'

const items = ref<TokenItem[]>([])
const error = ref('')
const busyId = ref<number | null>(null)

const statusMeta: Record<number, { text: string; cls: string }> = {
  1: { text: 'Active', cls: 'bg-tertiary-container/10 text-tertiary-container border-tertiary-container/20' },
  0: { text: 'Revoked', cls: 'bg-error-container/20 text-error border-error/20' },
}
const statusFallback = { text: 'Expired', cls: 'bg-surface-container-highest text-secondary border-outline-variant' }
function statusOf(s: number) {
  return statusMeta[s] ?? statusFallback
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

const { page, total, loading, totalPages, go } = usePagination(load)

async function load(p: number, size: number) {
  loading.value = true
  error.value = ''
  try {
    const d = await listTokens(p, size)
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
    await load(page.value, 10)
  } catch (e) {
    window.alert('Failed: ' + (e as Error).message)
  }
}

async function onRevoke(t: TokenItem) {
  if (!window.confirm(`Revoke token ${t.token_id}?`)) return
  busyId.value = t.id
  try {
    await revokeToken(t.id)
    await load(page.value, 10)
  } catch (e) {
    window.alert('Failed: ' + (e as Error).message)
  } finally {
    busyId.value = null
  }
}

onMounted(load)
</script>

<template>
  <AdminLayout>
    <TopNavBar
      title="Token Management"
      name="Admin Role"
      avatar="https://lh3.googleusercontent.com/aida-public/AB6AXuCSP4HSMxpb_zzfzPJMC4JEN47ebIsfmfEREiPfh4H46AELYrnXMd_57gOBi6Kz46L_i5GRDW9bpIBujUyXsIthHA6NsYmKbuKOvrKZZDhtGPU1O_BsDiwO-FzFU38njrWY-JgHsKHJqcjx2bWVOvSBPUAUDqw9qEXePUng8fIRlfTM2M690Cyg5UhmVxFCsOtLkjirbUphIw3Jl5IwRuNOwgZlrarnY9Ul_5hrkT_XE-W5fQT7kFFAY-oUvpfRCpuAvKLkEOXJsI4"
    />

    <!-- Content Canvas -->
    <div class="flex-1 overflow-y-auto p-gutter scrollbar-hide">
      <ErrorBanner :message="error" />

      <PageHeader title="Token Infrastructure" subtitle="Monitor and control API access tokens across the system.">
        <template #actions>
          <button
            class="bg-surface-container-lowest border border-outline-variant px-4 py-2 rounded-lg font-label-bold text-label-bold text-secondary flex items-center gap-2 hover:bg-surface-container-high transition-all active:scale-95"
            @click="load(page, 10)"
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
        </template>
      </PageHeader>

      <Card>
        <template #header>
          <h3 class="font-label-bold text-label-bold text-on-surface">Active System Tokens</h3>
          <span class="font-body-sm text-body-sm text-secondary">Showing {{ items.length }} of {{ total }} tokens</span>
        </template>
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
                    <UserAvatar :name="row.user_name" size="sm" />
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
                  <StatusBadge :text="statusOf(row.status).text" :cls="statusOf(row.status).cls" />
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
        <template #footer>
          <Pagination
            :page="page"
            :total-pages="totalPages"
            :shown="items.length"
            :total="total"
            label="tokens"
            @prev="go(page - 1)"
            @next="go(page + 1)"
            @goto="go"
          />
        </template>
      </Card>
    </div>

    <AppFooter />
  </AdminLayout>
</template>

<style scoped>
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
</style>
