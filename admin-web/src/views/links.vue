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
import { listLinks, type LinkItem } from '@/api/admin'

const items = ref<LinkItem[]>([])
const error = ref('')

const statusMeta: Record<number, { text: string; cls: string }> = {
  1: { text: 'Active', cls: 'bg-tertiary-container/10 text-tertiary border-tertiary/20' },
  0: { text: 'Expired', cls: 'bg-surface-variant text-secondary border-outline-variant' },
}
const statusFallback = { text: 'Flagged', cls: 'bg-error-container/20 text-error border-error/20' }
function statusOf(s: number) {
  return statusMeta[s] ?? statusFallback
}

const { page, total, loading, totalPages, go } = usePagination(load)

async function load(p: number, size: number) {
  loading.value = true
  error.value = ''
  try {
    const d = await listLinks(p, size)
    items.value = d.items
    total.value = d.total
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <AdminLayout>
    <TopNavBar
      title="Link Management"
      name="Alex Rivera"
      subtitle="Admin Role"
      avatar="https://lh3.googleusercontent.com/aida-public/AB6AXuDH5PVXbxxWOAoplirZBZOzJpheVICmBgEJG_RB8kZPOBsgTD-aBREuYi5liIMBwrBVcmWlNXqNwxGk8ri3-EtLQQIn472QnFT-JwGCBrG2yUrAxUiNlhWcb0kl_oeoCXXzS7n_D6G0CSM1ujGaKdkOuYh-Lqlwi5JRIAOFLsPRp7Sp2tNtGEyAungJwFB9RwgZYuJShSmJ_E5mRcNKj_lmVH3SUavdP-IpowCMd94NZ-1AP-Y1BOvyx_UEedFLeqa9xyT054lnkAE"
    />

    <!-- Content Canvas -->
    <div class="p-gutter flex flex-col flex-1">
      <ErrorBanner :message="error" />

      <PageHeader title="Link Management" subtitle="Monitor, filter, and audit all generated links across the infrastructure." />

      <Card>
        <template #header>
          <h3 class="text-label-bold font-label-bold text-on-surface">All Links</h3>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-surface-container-low border-b border-outline-variant">
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">User</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">Original URL</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">Shortened URL</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">Source</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">Created At</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold text-right">Visits</th>
                <th class="px-6 py-4 text-label-caps text-secondary font-bold">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-outline-variant">
              <tr v-for="row in items" :key="row.code" class="hover:bg-primary/5 transition-colors duration-150">
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <UserAvatar :name="row.user_name" size="md" />
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
                <td class="px-6 py-4">
                  <span
                    class="inline-flex items-center rounded-full border px-2.5 py-0.5 text-[11px] font-medium"
                    :class="row.source === 'web'
                      ? 'bg-primary/10 text-primary border-primary/20'
                      : 'bg-tertiary-container/10 text-tertiary border-tertiary/20'"
                  >{{ row.source === 'web' ? 'Web' : 'API' }}</span>
                </td>
                <td class="px-6 py-4 text-body-sm text-secondary">{{ row.created_at }}</td>
                <td class="px-6 py-4 text-body-sm text-on-surface text-right font-medium">{{ row.clicks }}</td>
                <td class="px-6 py-4">
                  <StatusBadge :text="statusOf(row.status).text" :cls="statusOf(row.status).cls" />
                </td>
              </tr>
              <tr v-if="!loading && !items.length">
                <td class="px-6 py-10 text-center text-secondary text-body-sm" colspan="7">No links found</td>
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
            label="links"
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
