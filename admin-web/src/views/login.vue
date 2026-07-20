<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/api/client'
import ThemeToggle from '@/components/ThemeToggle.vue'

const router = useRouter()
const auth = useAuthStore()

const username = ref('admin')
const password = ref('admin123')
const error = ref('')
const loading = ref(false)

async function onSubmit() {
  error.value = ''
  if (!username.value.trim() || !password.value) {
    error.value = 'Please enter both username and password.'
    return
  }
  loading.value = true
  try {
    await auth.login(username.value.trim(), password.value)
    router.replace({ name: 'dashboard' })
  } catch (e) {
    if (e instanceof ApiError) {
      error.value = e.message || 'Login failed.'
    } else {
      error.value = 'Network error, please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="relative min-h-screen w-full bg-background flex items-center justify-center p-4">
    <div class="absolute top-4 right-4">
      <ThemeToggle />
    </div>
    <div class="w-full max-w-[400px]">
      <!-- Brand -->
      <div class="flex items-center justify-center gap-3 mb-8">
        <div
          class="w-10 h-10 bg-primary rounded-xl flex items-center justify-center text-on-primary"
        >
          <span class="material-symbols-outlined text-[24px]" style="font-variation-settings: 'FILL' 1"
            >link</span
          >
        </div>
        <h1 class="font-headline-lg text-headline-lg font-bold text-primary">SLink Admin</h1>
      </div>

      <!-- Card -->
      <div class="bg-surface-container-lowest border border-outline-variant rounded-xl shadow-sm p-8">
        <h2 class="font-headline-md text-headline-md font-semibold text-on-surface mb-1">
          Sign in
        </h2>
        <p class="text-secondary font-body-lg mb-6">Use your administrator credentials.</p>

        <form class="space-y-4" @submit.prevent="onSubmit">
          <div>
            <label class="block font-label-bold text-label-bold text-on-surface mb-1.5">Username</label>
            <input
              v-model="username"
              type="text"
              autocomplete="username"
              class="w-full px-3.5 py-2.5 rounded-lg border border-outline-variant bg-surface text-on-surface font-body-lg focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition"
              placeholder="admin"
            />
          </div>

          <div>
            <label class="block font-label-bold text-label-bold text-on-surface mb-1.5">Password</label>
            <input
              v-model="password"
              type="password"
              autocomplete="current-password"
              class="w-full px-3.5 py-2.5 rounded-lg border border-outline-variant bg-surface text-on-surface font-body-lg focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition"
              placeholder="••••••••"
            />
          </div>

          <p
            v-if="error"
            class="text-error font-body-sm bg-error-container/30 border border-error-container rounded-lg px-3 py-2"
          >
            {{ error }}
          </p>

          <button
            type="submit"
            :disabled="loading"
            class="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-primary text-on-primary rounded-lg font-label-bold text-label-bold hover:bg-primary-container transition-all active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed"
          >
            <span v-if="loading" class="material-symbols-outlined text-[18px] animate-spin">progress_activity</span>
            <span>{{ loading ? 'Signing in…' : 'Sign in' }}</span>
          </button>
        </form>
      </div>

      <p class="text-center text-secondary font-body-sm mt-6">
        SLink Admin Console · Protected area
      </p>
    </div>
  </div>
</template>

<style scoped></style>
