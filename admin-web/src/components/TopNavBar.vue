<script setup lang="ts">
import UserMenu from '@/components/UserMenu.vue'

withDefaults(
  defineProps<{
    title?: string
    name?: string
    subtitle?: string
    avatar?: string
  }>(),
  {
    title: 'Dashboard',
    name: 'Admin Role',
    subtitle: 'Super Administrator',
    avatar:
      'https://lh3.googleusercontent.com/aida-public/AB6AXuAVAvI7qKzqpTaE6g10DcEWbt_cFYcs20iFVH9uJyVw3EY6-dS8NzIs_ovNv6l0QzLwaEN8ksyzyKRH2ZdXSdXR1SbKqJGFO5n0xwY_23ox8ur8LnA4zvwvNjvyo2vVnttEUFwGRfcv9284HfEp3DOSeX8cEjt9SL0SNj-AntiuYuMHWVJYA0bTZep7bmseDE2kApVFzsyXsUzrqez7SFTgVFa529tXmyijHUV3AWB4RRWAF-wezKlohJ9Dy_YjdmbQuvVQVsSNuZ4',
  },
)
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
    <div class="flex items-center gap-6">
      <div class="border-outline-variant">
        <UserMenu :name="name" :subtitle="subtitle" :avatar="avatar" />
      </div>
    </div>
  </header>
</template>

<style scoped></style>
