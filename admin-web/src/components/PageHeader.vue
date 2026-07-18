<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <div class="flex flex-col md:flex-row md:items-end justify-between gap-4 mb-8">
    <div>
      <h2 class="font-headline-lg text-headline-lg text-on-surface mb-1">{{ title }}</h2>
      <p v-if="subtitle" class="text-body-sm text-secondary">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="flex items-center gap-3">
      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped></style>
