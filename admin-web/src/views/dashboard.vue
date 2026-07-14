<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import AdminLayout from '@/components/AdminLayout.vue'
import UserMenu from '@/components/UserMenu.vue'
import AppFooter from '@/components/AppFooter.vue'
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
    <!-- TopNavBar -->
    <header
      class="h-header-height sticky top-0 z-40 w-full bg-surface border-b border-outline-variant flex justify-between items-center px-gutter transition-all duration-200"
    >
      <div class="flex items-center gap-6">
        <nav class="flex items-center gap-2 text-secondary font-label-bold text-label-bold">
          <span class="hover:text-primary cursor-pointer transition-colors">Home</span>
          <span class="material-symbols-outlined text-[16px]">chevron_right</span>
          <span class="text-primary font-bold">Dashboard</span>
        </nav>
      </div>
      <div class="flex items-center gap-6">
        <div class="relative group">
          <span class="absolute -top-1 -right-1 w-2 h-2 bg-error rounded-full"></span>
        </div>
        <div class="pl-4 border-l border-outline-variant">
          <UserMenu
            name="Admin Role"
            subtitle="Super Administrator"
            avatar="https://lh3.googleusercontent.com/aida-public/AB6AXuAVAvI7qKzqpTaE6g10DcEWbt_cFYcs20iFVH9uJyVw3EY6-dS8NzIs_ovNv6l0QzLwaEN8ksyzyKRH2ZdXSdXR1SbKqJGFO5n0xwY_23ox8ur8LnA4zvwvNjvyo2vVnttEUFwGRfcv9284HfEp3DOSeX8cEjt9SL0SNj-AntiuYuMHWVJYA0bTZep7bmseDE2kApVFzsyXsUzrqez7SFTgVFa529tXmyijHUV3AWB4RRWAF-wezKlohJ9Dy_YjdmbQuvVQVsSNuZ4"
          />
        </div>
      </div>
    </header>

    <main class="p-gutter overflow-y-auto flex-1">
      <div v-if="error" class="mb-6 rounded-lg bg-error-container text-on-error-container px-4 py-3 text-body-sm">
        Failed to load dashboard: {{ error }}
      </div>

      <!-- Header Section -->
      <div class="flex justify-between items-end mb-8">
        <div>
          <h2 class="font-headline-lg text-headline-lg text-on-surface">System Overview</h2>
          <p class="text-secondary font-body-lg">Infrastructure health and traffic analytics for the last 24 hours.</p>
        </div>
        <button
          class="px-4 py-2 bg-primary text-white rounded-lg font-label-bold text-label-bold flex items-center gap-2 hover:bg-primary-container transition-all active:scale-[0.98]"
          @click="load"
        >
          <span class="material-symbols-outlined text-[18px]">refresh</span>
          Refresh
        </button>
      </div>

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
        <div
          class="col-span-12 lg:col-span-8 bg-white border border-outline-variant rounded-xl shadow-[0_1px_2px_0_rgba(0,0,0,0.05)]"
        >
          <div class="px-card-padding py-4 border-b border-outline-variant flex justify-between items-center">
            <h3 class="font-headline-md text-headline-md text-on-surface">System Traffic (7 days)</h3>
          </div>
          <div class="p-card-padding">
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
          </div>
        </div>

        <!-- Recent Admin Actions Table -->
        <div
          class="col-span-12 lg:col-span-4 bg-white border border-outline-variant rounded-xl shadow-[0_1px_2px_0_rgba(0,0,0,0.05)] overflow-hidden"
        >
          <div
            class="px-card-padding py-4 border-b border-outline-variant bg-surface-container-low/50 flex justify-between items-center"
          >
            <h3 class="font-headline-md text-headline-md text-on-surface">Recent Admin Actions</h3>
          </div>
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
        </div>
      </div>
    </main>

    <AppFooter />
  </AdminLayout>
</template>

<style scoped></style>
