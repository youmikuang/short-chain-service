<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import AdminLayout from '@/components/AdminLayout.vue'
import UserMenu from '@/components/UserMenu.vue'
import AppFooter from '@/components/AppFooter.vue'
import { listLinks, type LinkItem } from '@/api/admin'

const items = ref<LinkItem[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(10)
const loading = ref(false)
const error = ref('')

const statusMeta: Record<number, { text: string; cls: string }> = {
  1: { text: 'Active', cls: 'bg-tertiary-container/10 text-tertiary border-tertiary/20' },
  0: { text: 'Expired', cls: 'bg-surface-variant text-secondary border-outline-variant' },
}
function statusOf(s: number) {
  if (s === 1) return statusMeta[1]
  if (s === 0) return statusMeta[0]
  return { text: 'Flagged', cls: 'bg-error-container/20 text-error border-error/20' }
}

function initials(name: string) {
  return (name || '?')
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / size.value)))
const pageNumbers = computed(() => {
  const t = totalPages.value
  const cur = page.value
  if (t <= 7) return Array.from({ length: t }, (_, i) => i + 1)
  const set = new Set([1, t, cur, cur - 1, cur + 1].filter((n) => n >= 1 && n <= t))
  return Array.from(set).sort((a, b) => a - b)
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const d = await listLinks(page.value, size.value)
    items.value = d.items
    total.value = d.total
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function go(p: number) {
  if (p < 1 || p > totalPages.value || p === page.value) return
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
        <nav aria-label="Breadcrumb" class="flex text-secondary font-label-bold text-label-bold">
          <ol class="flex items-center space-x-2">
            <li><a class="hover:text-primary" href="#">Home</a></li>
            <li class="flex items-center space-x-2">
              <span class="material-symbols-outlined text-sm">chevron_right</span>
              <span class="text-on-surface">Link Management</span>
            </li>
          </ol>
        </nav>
      </div>
      <div class="flex items-center gap-6">
        <div class="relative hidden md:block">
          <span class="absolute inset-y-0 left-0 pl-3 flex items-center text-outline">
            <span class="material-symbols-outlined text-[20px]">search</span>
          </span>
          <input
            class="pl-10 pr-4 py-1.5 bg-surface-container-low border border-outline-variant rounded-full text-body-sm focus:ring-2 focus:ring-primary/15 focus:border-primary outline-none transition-all w-64"
            placeholder="Search system logs..."
            type="text"
          />
        </div>
        <div class="flex items-center gap-3">
          <button
            class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-surface-container-high transition-colors"
          >
            <span class="material-symbols-outlined text-on-surface-variant">notifications</span>
          </button>
          <div class="h-6 w-px bg-outline-variant"></div>
          <UserMenu
            name="Alex Rivera"
            subtitle="Admin Role"
            avatar="https://lh3.googleusercontent.com/aida-public/AB6AXuDH5PVXbxxWOAoplirZBZOzJpheVICmBgEJG_RB8kZPOBsgTD-aBREuYi5liIMBwrBVcmWlNXqNwxGk8ri3-EtLQQIn472QnFT-JwGCBrG2yUrAxUiNlhWcb0kl_oeoCXXzS7n_D6G0CSM1ujGaKdkOuYh-Lqlwi5JRIAOFLsPRp7Sp2tNtGEyAungJwFB9RwgZYuJShSmJ_E5mRcNKj_lmVH3SUavdP-IpowCMd94NZ-1AP-Y1BOvyx_UEedFLeqa9xyT054lnkAE"
          />
        </div>
      </div>
    </header>

    <!-- Content Canvas -->
    <div class="p-gutter flex flex-col flex-1">
      <div v-if="error" class="mb-6 rounded-lg bg-error-container text-on-error-container px-4 py-3 text-body-sm">
        Failed to load links: {{ error }}
      </div>

      <div class="mb-8 flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h2 class="font-headline-lg text-headline-lg text-on-surface mb-1">Link Management</h2>
          <p class="text-body-sm text-secondary">
            Monitor, filter, and audit all generated links across the infrastructure.
          </p>
        </div>
      </div>

      <!-- Data Table Container -->
      <div
        class="bg-surface-container-lowest border border-outline-variant rounded-xl overflow-hidden shadow-sm flex-1 flex flex-col"
      >
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-surface-container-low border-b border-outline-variant">
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">User</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">Original URL</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">Shortened URL</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">Created At</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold text-right">Visits</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-outline-variant">
              <tr v-for="row in items" :key="row.code" class="hover:bg-primary/5 transition-colors duration-150">
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div
                      class="w-8 h-8 rounded-full flex items-center justify-center font-bold text-[10px] bg-secondary-container text-on-secondary-container"
                    >
                      {{ initials(row.user_name) }}
                    </div>
                    <div>
                      <p class="text-body-sm font-bold text-on-surface">{{ row.user_name || '—' }}</p>
                      <p class="text-[11px] text-secondary">{{ row.user_email }}</p>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4">
                  <p class="text-body-sm text-on-surface-variant truncate max-w-[200px]" :title="row.long_url">
                    {{ row.long_url }}
                  </p>
                </td>
                <td class="px-6 py-4">
                  <p class="text-technical-mono text-primary font-medium">{{ row.short_url }}</p>
                </td>
                <td class="px-6 py-4 text-body-sm text-secondary">{{ row.created_at }}</td>
                <td class="px-6 py-4 text-body-sm text-on-surface text-right font-medium">{{ row.clicks }}</td>
                <td class="px-6 py-4">
                  <span
                    class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-bold border uppercase tracking-wider"
                    :class="statusOf(row.status).cls"
                    >{{ statusOf(row.status).text }}</span
                  >
                </td>
              </tr>
              <tr v-if="!loading && !items.length">
                <td class="px-6 py-10 text-center text-secondary text-body-sm" colspan="6">No links found</td>
              </tr>
              <tr v-if="loading">
                <td class="px-6 py-10 text-center text-secondary text-body-sm" colspan="6">Loading…</td>
              </tr>
            </tbody>
          </table>
        </div>
        <!-- Pagination -->
        <div
          class="mt-auto border-t border-outline-variant px-6 py-4 flex items-center justify-between bg-surface-container-low"
        >
          <p class="text-body-sm text-secondary">
            Showing <span class="font-bold text-on-surface">{{ items.length }}</span> of
            <span class="font-bold text-on-surface">{{ total }}</span> links
          </p>
          <div class="flex items-center gap-2">
            <button
              class="p-1 rounded hover:bg-surface-container-high text-outline disabled:opacity-50 transition-colors"
              :disabled="page <= 1"
              @click="go(page - 1)"
            >
              <span class="material-symbols-outlined">chevron_left</span>
            </button>
            <button
              v-for="p in pageNumbers"
              :key="p"
              class="w-8 h-8 rounded-md text-body-sm font-medium transition-colors"
              :class="p === page ? 'bg-primary text-on-primary font-bold' : 'hover:bg-surface-container-high text-on-surface'"
              @click="go(p)"
            >
              {{ p }}
            </button>
            <button
              class="p-1 rounded hover:bg-surface-container-high text-on-surface transition-colors"
              :disabled="page >= totalPages"
              @click="go(page + 1)"
            >
              <span class="material-symbols-outlined">chevron_right</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <AppFooter />
  </AdminLayout>
</template>

<style scoped></style>
