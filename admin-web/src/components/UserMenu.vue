<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'

const props = withDefaults(
  defineProps<{
    name?: string
    subtitle?: string
    avatar?: string
  }>(),
  {
    name: 'Admin Role',
    subtitle: 'Super Administrator',
    avatar:
      'https://lh3.googleusercontent.com/aida-public/AB6AXuAVAvI7qKzqpTaE6g10DcEWbt_cFYcs20iFVH9uJyVw3EY6-dS8NzIs_ovNv6l0QzLwaEN8ksyzyKRH2ZdXSdXR1SbKqJGFO5n0xwY_23ox8ur8LnA4zvwvNjvyo2vVnttEUFwGRfcv9284HfEp3DOSeX8cEjt9SL0SNj-AntiuYuMHWVJYA0bTZep7bmseDE2kApVFzsyXsUzrqez7SFTgVFa529tXmyijHUV3AWB4RRWAF-wezKlohJ9Dy_YjdmbQuvVQVsSNuZ4',
  },
)

const emit = defineEmits<{ (e: 'logout'): void }>()

const open = ref(false)
const menuRef = ref<HTMLElement | null>(null)

function toggle() {
  open.value = !open.value
}

function onClickOutside(e: MouseEvent) {
  if (menuRef.value && !menuRef.value.contains(e.target as Node)) {
    open.value = false
  }
}

function onLogout() {
  open.value = false
  emit('logout')
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))
</script>

<template>
  <div ref="menuRef" class="relative">
    <button
      type="button"
      class="flex items-center gap-2 rounded-full hover:bg-surface-container-high transition-colors pr-1"
      @click.stop="toggle"
    >
      <div v-if="name || subtitle" class="text-right hidden sm:block">
        <p class="font-label-bold text-label-bold text-on-surface leading-none">{{ name }}</p>
        <p v-if="subtitle" class="text-[11px] text-secondary font-body-sm">{{ subtitle }}</p>
      </div>
      <img
        class="w-8 h-8 rounded-full border border-outline-variant object-cover"
        alt="Administrator avatar"
        :src="avatar"
      />
      <span class="material-symbols-outlined text-[18px] text-secondary">expand_more</span>
    </button>

    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 scale-95 -translate-y-1"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="open"
        class="absolute right-0 mt-2 w-56 bg-surface-container-lowest border border-outline-variant rounded-xl shadow-lg overflow-hidden z-50"
      >
        <div class="px-4 py-3 border-b border-outline-variant flex items-center gap-3">
          <img
            class="w-9 h-9 rounded-full border border-outline-variant object-cover"
            alt="Administrator avatar"
            :src="avatar"
          />
          <div class="min-w-0">
            <p class="font-label-bold text-label-bold text-on-surface truncate">{{ name }}</p>
            <p v-if="subtitle" class="text-[11px] text-secondary truncate">{{ subtitle }}</p>
          </div>
        </div>
        <div class="py-1">
          <button
            class="w-full flex items-center gap-3 px-4 py-2.5 text-body-sm text-on-surface hover:bg-surface-container-high transition-colors text-left"
          >
            <span class="material-symbols-outlined text-[20px] text-secondary">person</span>
            Profile
          </button>
          <button
            class="w-full flex items-center gap-3 px-4 py-2.5 text-body-sm text-on-surface hover:bg-surface-container-high transition-colors text-left"
          >
            <span class="material-symbols-outlined text-[20px] text-secondary">settings</span>
            Settings
          </button>
          <div class="my-1 border-t border-outline-variant"></div>
          <button
            class="w-full flex items-center gap-3 px-4 py-2.5 text-body-sm text-error hover:bg-error-container/20 transition-colors text-left"
            @click="onLogout"
          >
            <span class="material-symbols-outlined text-[20px]">logout</span>
            Logout
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped></style>
