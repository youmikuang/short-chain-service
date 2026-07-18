<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AdminLayout from '@/components/AdminLayout.vue'
import TopNavBar from '@/components/TopNavBar.vue'
import AppFooter from '@/components/AppFooter.vue'
import PageHeader from '@/components/PageHeader.vue'
import Card from '@/components/Card.vue'
import Pagination from '@/components/Pagination.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import ErrorBanner from '@/components/ErrorBanner.vue'
import { usePagination } from '@/composables/usePagination'
import { listBlacklist, addBlacklist, type BlacklistItem } from '@/api/admin'

const items = ref<BlacklistItem[]>([])
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

const { page, total, loading, totalPages, go } = usePagination(load)

async function load(p: number, size: number) {
  loading.value = true
  error.value = ''
  try {
    const d = await listBlacklist(p, size)
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
    await load(page.value, 10)
  } catch (e) {
    window.alert('Failed: ' + (e as Error).message)
  } finally {
    adding.value = false
  }
}

onMounted(load)
</script>

<template>
  <AdminLayout>
    <TopNavBar
      title="Domain Blacklist"
      name="Admin Role"
      subtitle="ID: SL-9921"
      avatar="https://lh3.googleusercontent.com/aida-public/AB6AXuAcLtIAF8N60yyilIfSZKjF8tJ8fHDIZQDpOqPxqHtjgtzVOUVFbIaZagHEewbws1hWmRsv-bITYBrD9DarHPg-UMKqKvm9euT1Yh-d9j-Xji8RqPDlifAZeSnIm7Oy7uKWNZySSMwpdFWgI0rRpA2rJQearb8UzioU3avUYPeT_rvSdVm7nb8DaDlONpXHBLwXMpCKJGDu0SZ0zF7HkSoOuImascfQK0Zms6Y74VZCRsJZiGgyZK8ZqiXPoPPgE6Ew2l10pah-UZg"
    />

    <!-- Content Canvas -->
    <div class="p-gutter flex flex-col flex-1">
      <ErrorBanner :message="error" />

      <PageHeader title="Domain Blacklist" subtitle="Manage global restrictions and security interceptors for suspicious domains.">
        <template #actions>
          <button
            class="bg-primary hover:bg-primary-container text-white px-5 py-2.5 rounded-lg flex items-center gap-2 font-label-bold text-label-bold transition-all active:scale-95 shadow-sm disabled:opacity-50"
            :disabled="adding"
            @click="onAdd"
          >
            <span class="material-symbols-outlined text-[20px]">add_circle</span>
            Add New Blocked Domain
          </button>
        </template>
      </PageHeader>

      <Card>
        <template #header>
          <h3 class="text-label-bold font-label-bold text-on-surface">Blacklisted Domains</h3>
          <button class="p-1.5 hover:bg-surface-container-high rounded transition-colors" @click="load(page, 10)">
            <span class="material-symbols-outlined text-[20px] text-secondary">refresh</span>
          </button>
        </template>
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
                  <StatusBadge :text="row.reason" :cls="reasonClass(row.reason)" />
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
        <template #footer>
          <Pagination
            :page="page"
            :total-pages="totalPages"
            :shown="items.length"
            :total="total"
            label="entries"
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

<style scoped></style>
