<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import AdminLayout from '@/components/AdminLayout.vue'
import UserMenu from '@/components/UserMenu.vue'
import AppFooter from '@/components/AppFooter.vue'
import { listBlacklist, addBlacklist, type BlacklistItem } from '@/api/admin'

const items = ref<BlacklistItem[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(10)
const loading = ref(false)
const error = ref('')
const adding = ref(false)

const reasonClasses: Record<string, string> = {
  Phishing: 'bg-error-container text-on-error-container',
  Spam: 'bg-secondary-container text-on-secondary-container',
  Malware: 'bg-error text-white',
}
function reasonClass(r: string) {
  return reasonClasses[r] || 'bg-surface-variant text-secondary'
}

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / size.value)))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const d = await listBlacklist(page.value, size.value)
    items.value = d.items
    total.value = d.total
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function onAdd() {
  const domain = window.prompt('Domain to block (e.g. evil-phishing.cc)')
  if (!domain) return
  const reason = window.prompt('Reason (Phishing / Spam / Malware)', 'Phishing') || 'Phishing'
  adding.value = true
  try {
    await addBlacklist(domain.trim(), reason.trim())
    await load()
  } catch (e) {
    window.alert('Failed: ' + (e as Error).message)
  } finally {
    adding.value = false
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
        <nav aria-label="Breadcrumb" class="flex text-secondary font-label-bold text-label-bold">
          <ol class="flex items-center space-x-2">
            <li><a class="hover:text-primary" href="#">Home</a></li>
            <li class="flex items-center space-x-2">
              <span class="material-symbols-outlined text-sm">chevron_right</span>
              <span class="text-on-surface">Domain Blacklist</span>
            </li>
          </ol>
        </nav>
      </div>
      <div class="flex items-center gap-6">
        <div class="h-8 w-px bg-outline-variant mx-1"></div>
        <UserMenu
          name="Admin Role"
          subtitle="ID: SL-9921"
          avatar="https://lh3.googleusercontent.com/aida-public/AB6AXuAcLtIAF8N60yyilIfSZKjF8tJ8fHDIZQDpOqPxqHtjgtzVOUVFbIaZagHEewbws1hWmRsv-bITYBrD9DarHPg-UMKqKvm9euT1Yh-d9j-Xji8RqPDlifAZeSnIm7Oy7uKWNZySSMwpdFWgI0rRpA2rJQearb8UzioU3avUYPeT_rvSdVm7nb8DaDlONpXHBLwXMpCKJGDu0SZ0zF7HkSoOuImascfQK0Zms6Y74VZCRsJZiGgyZK8ZqiXPoPPgE6Ew2l10pah-UZg"
        />
      </div>
    </header>

    <!-- Content Canvas -->
    <div class="p-gutter flex flex-col flex-1">
      <div v-if="error" class="mb-6 rounded-lg bg-error-container text-on-error-container px-4 py-3 text-body-sm">
        Failed to load blacklist: {{ error }}
      </div>

      <!-- Page Header & CTA -->
      <div class="mb-8 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 class="font-headline-lg text-headline-lg text-on-surface mb-1">Domain Blacklist</h2>
          <p class="text-body-sm text-secondary">
            Manage global restrictions and security interceptors for suspicious domains.
          </p>
        </div>
        <button
          class="bg-primary hover:bg-primary-container text-white px-5 py-2.5 rounded-lg flex items-center gap-2 font-label-bold text-label-bold transition-all active:scale-95 shadow-sm disabled:opacity-50"
          :disabled="adding"
          @click="onAdd"
        >
          <span class="material-symbols-outlined text-[20px]">add_circle</span>
          Add New Blocked Domain
        </button>
      </div>

      <!-- List/Table Section -->
      <div
        class="bg-surface-container-lowest border border-outline-variant rounded-xl overflow-hidden shadow-sm flex-1 flex flex-col"
      >
        <div
          class="px-6 py-4 border-b border-outline-variant flex items-center justify-between bg-surface-container-low/50"
        >
          <h3 class="text-label-bold font-label-bold text-on-surface">Blacklisted Domains</h3>
          <div class="flex items-center gap-2">
            <button class="p-1.5 hover:bg-surface-container-high rounded transition-colors" @click="load">
              <span class="material-symbols-outlined text-[20px] text-secondary">refresh</span>
            </button>
          </div>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-surface-container-low/30">
                <th class="px-6 py-3 text-label-caps font-label-caps text-secondary border-b border-outline-variant uppercase">Domain Name</th>
                <th class="px-6 py-3 text-label-caps font-label-caps text-secondary border-b border-outline-variant uppercase">Reason</th>
                <th class="px-6 py-3 text-label-caps font-label-caps text-secondary border-b border-outline-variant uppercase">Date Added</th>
                <th class="px-6 py-3 text-label-caps font-label-caps text-secondary border-b border-outline-variant uppercase text-right">Attempts</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-outline-variant">
              <tr v-for="row in items" :key="row.domain" class="hover:bg-primary/5 transition-colors duration-150">
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div class="w-8 h-8 rounded bg-error/10 flex items-center justify-center text-error">
                      <span class="material-symbols-outlined text-[18px]">language</span>
                    </div>
                    <span class="font-technical-mono text-technical-mono text-on-surface">{{ row.domain }}</span>
                  </div>
                </td>
                <td class="px-6 py-4">
                  <span
                    class="text-[11px] font-bold px-2 py-0.5 rounded-full uppercase tracking-tight"
                    :class="reasonClass(row.reason)"
                    >{{ row.reason }}</span
                  >
                </td>
                <td class="px-6 py-4 text-body-sm text-secondary">{{ row.created_at }}</td>
                <td class="px-6 py-4 text-right font-technical-mono text-on-surface">{{ row.attempts }}</td>
              </tr>
              <tr v-if="!loading && !items.length">
                <td class="px-6 py-10 text-center text-secondary text-body-sm" colspan="4">No blocked domains</td>
              </tr>
              <tr v-if="loading">
                <td class="px-6 py-10 text-center text-secondary text-body-sm" colspan="4">Loading…</td>
              </tr>
            </tbody>
          </table>
        </div>
        <!-- Pagination -->
        <div
          class="mt-auto border-t border-outline-variant px-6 py-4 flex items-center justify-between bg-surface-container-lowest"
        >
          <p class="text-body-sm text-secondary">
            Showing <span class="font-bold text-on-surface">{{ items.length }}</span> of
            <span class="font-bold text-on-surface">{{ total }}</span> entries
          </p>
          <div class="flex items-center gap-1">
            <button
              class="w-8 h-8 flex items-center justify-center rounded border border-outline-variant text-secondary hover:bg-surface-container-high transition-colors disabled:opacity-50"
              :disabled="page <= 1"
              @click="go(page - 1)"
            >
              <span class="material-symbols-outlined text-[18px]">chevron_left</span>
            </button>
            <button
              class="w-8 h-8 flex items-center justify-center rounded font-label-bold text-label-bold transition-colors"
              :class="page === 1 ? 'bg-primary text-white' : 'border border-outline-variant text-secondary hover:bg-surface-container-high'"
            >
              {{ page }}
            </button>
            <button
              class="w-8 h-8 flex items-center justify-center rounded border border-outline-variant text-secondary hover:bg-surface-container-high transition-colors disabled:opacity-50"
              :disabled="page >= totalPages"
              @click="go(page + 1)"
            >
              <span class="material-symbols-outlined text-[18px]">chevron_right</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <AppFooter />
  </AdminLayout>
</template>

<style scoped></style>
