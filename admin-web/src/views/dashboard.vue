<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import AdminLayout from '@/components/AdminLayout.vue'
import TopNavBar from '@/components/TopNavBar.vue'
import AppFooter from '@/components/AppFooter.vue'
import PageHeader from '@/components/PageHeader.vue'
import Card from '@/components/Card.vue'
import ErrorBanner from '@/components/ErrorBanner.vue'
import { getDashboard, type KpiItem, type TrafficPoint, type AdminActionItem } from '@/api/admin'

const kpis = ref<KpiItem[]>([])
const traffic = ref<TrafficPoint[]>([])
const actions = ref<AdminActionItem[]>([])
const loading = ref(true)
const error = ref('')

const kpiIcon: Record<string, string> = {
  links: 'link',
  visits: 'visibility',
  tokens: 'api',
  blocked: 'block',
}

const maxTraffic = computed(() =>
  traffic.value.reduce((m, p) => Math.max(m, p.value), 0),
)
const days = ['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN']

async function load() {
  loading.value = true
  error.value = ''
  try {
    const d = await getDashboard()
    kpis.value = d.kpis
    traffic.value = d.traffic
    actions.value = d.actions
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
    <TopNavBar />

    <main class="p-gutter overflow-y-auto flex-1">
      <ErrorBanner :message="error" />

      <PageHeader title="System Overview" subtitle="Infrastructure health and traffic analytics for the last 24 hours.">
        <template #actions>
          <button
            class="px-4 py-2 bg-primary text-white rounded-lg font-label-bold text-label-bold flex items-center gap-2 hover:bg-primary-container transition-all active:scale-[0.98]"
            @click="load"
          >
            <span class="material-symbols-outlined text-[18px]">refresh</span>
            Refresh
          </button>
        </template>
      </PageHeader>

      <!-- KPI Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-gutter mb-8">
        <div
          v-for="k in kpis"
          :key="k.key"
          class="bg-white p-card-padding border border-outline-variant rounded-xl shadow-[0_1px_2px_0_rgba(0,0,0,0.05)] hover:border-primary/50 transition-colors"
        >
          <div class="flex justify-between items-start mb-4">
            <div class="p-2 rounded-lg bg-primary/10">
              <span class="material-symbols-outlined text-primary">{{ kpiIcon[k.key] || 'insights' }}</span>
            </div>
            <span
              v-if="k.badge"
              class="text-tertiary font-label-bold text-label-bold bg-tertiary/10 px-2 py-1 rounded"
              >{{ k.badge }}</span
            >
          </div>
          <p class="text-secondary font-label-caps uppercase tracking-wider mb-1">{{ k.label }}</p>
          <p class="font-headline-lg text-headline-lg text-on-surface">{{ k.value }}</p>
        </div>
        <div v-if="loading" class="col-span-full text-secondary text-body-sm">Loading…</div>
      </div>

      <!-- Dashboard Body Grid -->
      <div class="grid grid-cols-12 gap-gutter">
        <!-- System Traffic Chart -->
        <Card body-class="p-card-padding" class="col-span-12 lg:col-span-8">
          <template #header>
            <h3 class="font-headline-md text-headline-md text-on-surface">System Traffic (7 days)</h3>
          </template>
          <div class="h-[320px] w-full flex items-end justify-between gap-3">
            <div
              v-for="(p, i) in traffic"
              :key="p.date"
              class="flex-1 flex flex-col items-center justify-end h-full group"
            >
              <span class="text-[11px] font-technical-mono text-secondary mb-2">{{ p.value }}</span>
              <div
                class="w-full rounded-t bg-primary transition-all group-hover:bg-primary-container"
                :style="{ height: (maxTraffic ? (p.value / maxTraffic) * 100 : 0) + '%' }"
              ></div>
              <span class="text-[11px] font-technical-mono text-secondary mt-2">{{ days[i] || p.date.slice(5) }}</span>
            </div>
            <div v-if="!traffic.length" class="w-full text-center text-secondary text-body-sm py-20">
              No traffic data
            </div>
          </div>
        </Card>

        <!-- Recent Admin Actions Table -->
        <Card class="col-span-12 lg:col-span-4">
          <template #header>
            <h3 class="font-headline-md text-headline-md text-on-surface">Recent Admin Actions</h3>
          </template>
          <div class="overflow-x-auto">
            <table class="w-full text-left">
              <thead class="bg-surface-container-low">
                <tr class="border-b border-outline-variant">
                  <th class="px-card-padding py-3 font-label-caps text-label-caps text-secondary uppercase tracking-wider">
                    Action
                  </th>
                  <th class="px-card-padding py-3 font-label-caps text-label-caps text-secondary uppercase tracking-wider">
                    Time
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-outline-variant">
                <tr v-for="(a, i) in actions" :key="i" class="hover:bg-primary/5 transition-colors group">
                  <td class="px-card-padding py-4">
                    <p class="font-body-lg text-body-lg text-on-surface mb-0.5">{{ a.title }}</p>
                    <p class="text-[11px] text-secondary font-technical-mono uppercase">{{ a.meta }}</p>
                  </td>
                  <td class="px-card-padding py-4 text-right">
                    <span class="text-[11px] text-secondary font-technical-mono">{{ a.time }}</span>
                  </td>
                </tr>
                <tr v-if="!actions.length">
                  <td class="px-card-padding py-4 text-secondary text-body-sm" colspan="2">No recent actions</td>
                </tr>
              </tbody>
            </table>
          </div>
        </Card>
      </div>
    </main>

    <AppFooter />
  </AdminLayout>
</template>

<style scoped></style>
