<script setup lang="ts">
withDefaults(
  defineProps<{
    bodyClass?: string
    flush?: boolean
  }>(),
  { bodyClass: '', flush: false },
)
</script>

<template>
  <div
    class="bg-surface-container-lowest border border-outline-variant rounded-xl overflow-hidden shadow-sm flex flex-col"
    :class="{ 'flex-1': !flush }"
  >
    <div v-if="$slots.header" class="px-6 py-4 border-b border-outline-variant flex items-center justify-between bg-surface-container-low/50">
      <slot name="header" />
    </div>
    <div :class="bodyClass" class="flex-1 min-h-0">
      <slot />
    </div>
    <div v-if="$slots.footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<style scoped></style>
