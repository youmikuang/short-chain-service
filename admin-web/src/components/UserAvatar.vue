<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    name?: string
    src?: string
    size?: 'sm' | 'md'
  }>(),
  { name: '?', src: '', size: 'md' },
)

const initials = computed(() =>
  (props.name || '?')
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase(),
)

const box = computed(() =>
  props.size === 'sm' ? 'w-7 h-7 text-[12px]' : 'w-8 h-8 text-[10px]',
)
const ring = computed(() =>
  props.size === 'sm' ? 'bg-primary-container/10 text-primary' : 'bg-secondary-container text-on-secondary-container',
)
</script>

<template>
  <img
    v-if="src"
    class="rounded-full border border-outline-variant object-cover"
    :class="box"
    alt="avatar"
    :src="src"
  />
  <div
    v-else
    class="rounded-full flex items-center justify-center font-bold"
    :class="[box, ring]"
  >
    {{ initials }}
  </div>
</template>

<style scoped></style>
