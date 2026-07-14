<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { RouterLink } from 'vue-router'

const auth = useAuthStore()
const open = ref(false)
const root = ref<HTMLElement | null>(null)

function toggle() {
  open.value = !open.value
}

function onClickOutside(e: MouseEvent) {
  if (root.value && !root.value.contains(e.target as Node)) {
    open.value = false
  }
}

async function onLogout() {
  open.value = false
  await auth.logout()
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))
</script>

<template>
  <div ref="root" class="user-menu">
    <div class="user-menu__avatar" @click="toggle">
      {{ auth.user?.initials ?? 'U' }}
    </div>
    <div v-show="open" class="user-menu__dropdown">
      <RouterLink to="/settings" class="user-menu__item" @click="open = false"
        >Settings</RouterLink
      >
      <a
        href="#"
        class="user-menu__item user-menu__item--danger user-menu__divider"
        @click.prevent="onLogout"
        >Logout</a
      >
    </div>
  </div>
</template>

<style src="@/styles/menus.css" scoped></style>
