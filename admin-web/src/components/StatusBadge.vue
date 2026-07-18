<script setup lang="ts">
withDefaults(
  defineProps<{
    text: string
    cls?: string
  }>(),
  { cls: 'bg-surface-variant text-secondary border-outline-variant' },
)
</script>

<template>
  <span
    class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-bold border uppercase tracking-wider"
    :class="cls"
    >{{ text }}</span
  >
</template>

<style scoped></style>
