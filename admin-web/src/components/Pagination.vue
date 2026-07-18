<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  totalPages: number
  shown: number
  total: number
  label?: string
}>()

const emit = defineEmits<{ (e: 'prev'): void; (e: 'next'): void; (e: 'goto', p: number): void }>()

const numbers = computed(() => {
  const t = props.totalPages
  const cur = props.page
  if (t <= 7) return Array.from({ length: t }, (_, i) => i + 1)
  const set = new Set([1, t, cur, cur - 1, cur + 1].filter((n) => n >= 1 && n <= t))
  return Array.from(set).sort((a, b) => a - b)
})
</script>

<template>
  <div
    class="border-t border-outline-variant px-6 py-4 flex items-center justify-between bg-surface-container-low"
  >
    <p class="text-body-sm text-secondary">
      Showing <span class="font-bold text-on-surface">{{ shown }}</span> of
      <span class="font-bold text-on-surface">{{ total }}</span> {{ label || 'entries' }}
    </p>
    <div class="flex items-center gap-2">
      <button
        class="p-1 rounded hover:bg-surface-container-high text-outline disabled:opacity-50 transition-colors"
        :disabled="page <= 1"
        @click="emit('prev')"
      >
        <span class="material-symbols-outlined">chevron_left</span>
      </button>
      <button
        v-for="p in numbers"
        :key="p"
        class="w-8 h-8 rounded-md text-body-sm font-medium transition-colors"
        :class="p === page ? 'bg-primary text-on-primary font-bold' : 'hover:bg-surface-container-high text-on-surface'"
        @click="emit('goto', p)"
      >
        {{ p }}
      </button>
      <button
        class="p-1 rounded hover:bg-surface-container-high text-on-surface transition-colors"
        :disabled="page >= totalPages"
        @click="emit('next')"
      >
        <span class="material-symbols-outlined">chevron_right</span>
      </button>
    </div>
  </div>
</template>

<style scoped></style>
