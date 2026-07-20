<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import UserMenu from '@/components/UserMenu.vue'
import ThemeToggle from '@/components/ThemeToggle.vue'
import { useAuthStore } from '@/stores/auth'

// 顶部栏为统一管理组件：用户信息统一来自登录态，页面只需传入当前页标题。
const props = withDefaults(defineProps<{ title?: string }>(), {
  title: 'Dashboard',
})

const DEFAULT_AVATAR =
  'https://lh3.googleusercontent.com/aida-public/AB6AXuAVAvI7qKzqpTaE6g10DcEWbt_cFYcs20iFVH9uJyVw3EY6-dS8NzIs_ovNv6l0QzLwaEN8ksyzyKRH2ZdXSdXR1SbKqJGFO5n0xwY_23ox8ur8LnA4zvwvNjvyo2vVnttEUFwGRfcv9284HfEp3DOSeX8cEjt9SL0SNj-AntiuYuMHWVJYA0bTZep7bmseDE2kApVFzsyXsUzrqez7SFTgVFa529tXmyijHUV3AWB4RRWAF-wezKlohJ9Dy_YjdmbQuvVQVsSNuZ4'

const router = useRouter()
const auth = useAuthStore()

// 用户名统一取自登录态，未登录时回落到统一默认值
const displayName = computed(() => auth.username || 'Admin')

function onLogout() {
  auth.logout()
  router.replace({ name: 'login' })
}
</script>

<template>
  <header
    class="h-header-height sticky top-0 z-40 w-full bg-surface border-b border-outline-variant flex justify-between items-center px-gutter transition-all duration-200"
  >
    <div class="flex items-center gap-6">
      <nav class="flex items-center gap-2 text-secondary font-label-bold text-label-bold">
        <span class="hover:text-primary cursor-pointer transition-colors">Home</span>
        <span class="material-symbols-outlined text-[16px]">chevron_right</span>
        <span class="text-primary font-bold">{{ title }}</span>
      </nav>
    </div>
    <div class="flex items-center gap-2 pl-2">
      <ThemeToggle />
      <div class="border-outline-variant">
        <UserMenu :name="displayName" :avatar="DEFAULT_AVATAR" @logout="onLogout" />
      </div>
    </div>
  </header>
</template>

<style scoped></style>
