<script setup lang="ts">
import { RouterLink } from 'vue-router'

const navItems = [
  { to: '/', label: 'Dashboard', icon: 'dashboard', fill: false },
  { to: '/links', label: 'Links', icon: 'link', fill: true },
  { to: '/blacklist', label: 'Blacklist', icon: 'block', fill: false },
  { to: '/tokens', label: 'Tokens', icon: 'vpn_key', fill: false },
]
</script>

<template>
  <div class="bg-background text-on-background font-body-lg min-h-screen flex">
    <!-- SideNavBar -->
    <aside
      class="w-sidebar-width h-screen sticky top-0 left-0 bg-surface-container-lowest border-r border-outline-variant flex flex-col py-baseline px-2 z-50"
    >
      <div class="px-4 py-6 flex items-center gap-3">
        <div class="w-8 h-8 bg-primary rounded-lg flex items-center justify-center text-on-primary">
          <span class="material-symbols-outlined text-[20px]" style="font-variation-settings: 'FILL' 1">link</span>
        </div>
        <div>
          <h1 class="font-headline-md text-headline-md font-bold text-primary leading-tight">SLink Admin</h1>
        </div>
      </div>
      <nav class="flex-1 mt-4 space-y-1">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="group flex items-center px-4 py-2.5 text-secondary hover:bg-surface-container-high transition-colors duration-200 rounded-lg font-body-lg"
          active-class="active-pill !text-primary !font-bold !bg-secondary-container/30 !border-primary"
        >
          <span
            class="material-symbols-outlined mr-3 text-on-surface-variant group-hover:text-primary"
            :style="item.fill ? 'font-variation-settings: \'FILL\' 1' : ''"
            >{{ item.icon }}</span
          >
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
    </aside>

    <!-- Main Content Area -->
    <main class="flex-1 flex flex-col min-w-0">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.active-pill {
  position: relative;
}
.active-pill::before {
  content: '';
  position: absolute;
  left: 0;
  top: 15%;
  height: 70%;
  width: 2px;
  background-color: #004ac6;
}
</style>
