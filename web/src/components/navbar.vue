<script setup lang="ts">
import { RouterLink } from 'vue-router'
import UserMenu from '@/components/menus.vue'
import ThemeToggle from '@/components/theme.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>

<template>
  <header class="navbar">
    <nav class="navbar__inner">
      <div class="navbar__left">
        <RouterLink to="/" class="navbar__brand">SLink</RouterLink>
        <div class="navbar__links">
          <RouterLink
            to="/"
            class="navbar__link"
            active-class="navbar__link--active"
            :exact-active-class="''"
            ><span class="material-symbols-outlined navbar__link-icon">home</span
            ><span>Home</span></RouterLink
          >
          <RouterLink
            to="/token"
            class="navbar__link"
            active-class="navbar__link--active"
            ><span class="material-symbols-outlined navbar__link-icon">key</span
            ><span>Token</span></RouterLink
          >
          <RouterLink
            to="/logs"
            class="navbar__link"
            active-class="navbar__link--active"
            ><span class="material-symbols-outlined navbar__link-icon">description</span
            ><span>Logs</span></RouterLink
          >
          <RouterLink
            to="/urls"
            class="navbar__link"
            active-class="navbar__link--active"
            ><span class="material-symbols-outlined navbar__link-icon">link</span
            ><span>Urls</span></RouterLink
          >
        </div>
      </div>

      <div class="navbar__right">
        <ThemeToggle />
        <template v-if="auth.user">
          <UserMenu />
        </template>
        <template v-else>
          <RouterLink to="/login" class="navbar__login">Login</RouterLink>
        </template>
        <button class="navbar__menu-btn">
          <span class="material-symbols-outlined">menu</span>
        </button>
      </div>
    </nav>
  </header>
</template>

<style src="@/styles/navbar.css" scoped></style>
