<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue'

const props = withDefaults(
  defineProps<{
    type?: 'ok' | 'error'
    message?: string
    duration?: number
  }>(),
  {
    type: 'ok',
    message: '',
    duration: 3000,
  },
)

const emit = defineEmits<{ close: [] }>()

const visible = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

watch(
  () => props.message,
  (msg) => {
    if (msg) show()
  },
  { immediate: true },
)

function show() {
  visible.value = true
  clearTimeout(timer)
  timer = setTimeout(close, props.duration)
}

function close() {
  clearTimeout(timer)
  visible.value = false
  emit('close')
}

onBeforeUnmount(() => clearTimeout(timer))
</script>

<template>
  <Transition name="toast">
    <div
      v-if="visible"
      class="toast"
      :class="`toast--${type}`"
      role="alert"
    >
      <span class="toast__icon material-symbols-outlined">
        {{ type === 'ok' ? 'check_circle' : 'error' }}
      </span>
      <span class="toast__msg">{{ message }}</span>
      <button class="toast__close" type="button" @click="close" aria-label="Dismiss">
        <span class="material-symbols-outlined">close</span>
      </button>
      <div
        class="toast__progress"
        :style="{ animationDuration: duration + 'ms' }"
      ></div>
    </div>
  </Transition>
</template>

<style scoped>
.toast {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 1000;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 280px;
  max-width: 380px;
  padding: 14px 16px;
  border-radius: 12px;
  color: #fff;
  font-size: 14px;
  line-height: 20px;
  font-weight: 600;
  background: rgb(var(--color-primary) / 0.85);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.25);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.25), inset 0 1px 0 rgba(255, 255, 255, 0.3);
  overflow: hidden;
}
.toast--ok {
  background: rgb(var(--color-primary));
}
.toast--error {
  background: rgb(var(--color-error) / 0.85);
}
.toast__icon {
  font-size: 22px;
  flex-shrink: 0;
}
.toast__msg {
  flex: 1;
  word-break: break-word;
}
.toast__close {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.85);
  cursor: pointer;
  border-radius: 6px;
  transition: background 0.2s ease;
}
.toast__close:hover {
  background: rgba(255, 255, 255, 0.18);
}
.toast__close .material-symbols-outlined {
  font-size: 18px;
}
.toast__progress {
  position: absolute;
  left: 0;
  bottom: 0;
  height: 3px;
  width: 100%;
  background: rgba(255, 255, 255, 0.7);
  transform-origin: left;
  animation-name: toast-progress;
  animation-timing-function: linear;
  animation-fill-mode: forwards;
}
@keyframes toast-progress {
  from {
    transform: scaleX(1);
  }
  to {
    transform: scaleX(0);
  }
}

.toast-enter-active,
.toast-leave-active {
  transition: transform 0.3s ease, opacity 0.3s ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(16px);
}
</style>
