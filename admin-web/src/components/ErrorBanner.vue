<script setup lang="ts">
defineProps<{
  message: string
}>()
</script>

<template>
  <div
    v-if="message"
    class="mb-6 rounded-lg bg-error-container text-on-error-container px-4 py-3 text-body-sm"
  >
    {{ message }}
  </div>
</template>

<style scoped></style>
