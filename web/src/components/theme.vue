<script setup lang="ts">
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()

// View Transitions API circular reveal (BewlyCat style), matching
// https://tools.gaoheng.top/ theme-btn behavior.
function onToggle(e: MouseEvent) {
  const isAppearanceTransition =
    typeof document !== 'undefined' &&
    !!(document as any).startViewTransition &&
    !window.matchMedia('(prefers-reduced-motion: reduce)').matches

  if (!isAppearanceTransition) {
    theme.toggle()
    return
  }

  const x = e.clientX
  const y = e.clientY
  const endRadius = Math.hypot(
    Math.max(x, innerWidth - x),
    Math.max(y, innerHeight - y),
  )

  // Disable all CSS transitions during the view transition to avoid double animation.
  const disableTransitions = document.createElement('style')
  disableTransitions.textContent = `*, *::before, *::after { transition: none !important; }`
  document.head.appendChild(disableTransitions)

  // Disable default cross-fade of ::view-transition root.
  const disableDefault = document.createElement('style')
  disableDefault.textContent = `
    ::view-transition-old(root),
    ::view-transition-new(root) {
      animation: none !important;
      mix-blend-mode: normal;
    }
  `
  document.head.appendChild(disableDefault)

  const transition = (document as any).startViewTransition(() => {
    theme.toggle()
  })

  transition.ready.then(() => {
    const isDarkNow = theme.theme === 'dark'

    const zIndex = document.createElement('style')
    zIndex.textContent = `
      ::view-transition-old(root) { z-index: ${isDarkNow ? 1 : 9999}; }
      ::view-transition-new(root) { z-index: ${isDarkNow ? 9999 : 1}; }
    `
    document.head.appendChild(zIndex)

    const clipPath = [
      `circle(0px at ${x}px ${y}px)`,
      `circle(${endRadius}px at ${x}px ${y}px)`,
    ]

    const animation = document.documentElement.animate(
      { clipPath: isDarkNow ? clipPath : [...clipPath].reverse() },
      {
        duration: 300,
        easing: 'ease-in-out',
        pseudoElement: isDarkNow
          ? '::view-transition-new(root)'
          : '::view-transition-old(root)',
      },
    )

    animation.finished.then(() => zIndex.remove())
  })

  transition.finished.then(() => {
    disableTransitions.remove()
    disableDefault.remove()
  })
}
</script>

<template>
  <button
    class="btn-icon theme-btn"
    :title="theme.theme === 'dark' ? '切换到浅色模式' : '切换到深色模式'"
    :aria-label="theme.theme === 'dark' ? '切换到浅色模式' : '切换到深色模式'"
    @click="onToggle"
  >
    <!-- Sun (shown in dark mode -> switch to light) -->
    <svg
      v-if="theme.theme === 'dark'"
      key="sun"
      class="icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <circle cx="12" cy="12" r="4" />
      <path
        d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"
      />
    </svg>
    <!-- Moon (shown in light mode -> switch to dark) -->
    <svg
      v-else
      key="moon"
      class="icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
    </svg>
  </button>
</template>

<style scoped>
.btn-icon.theme-btn {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: transparent;
  color: inherit;
  border: 0px;
  transition: background 0.2s ease, color 0.2s ease, transform 0.15s ease;
  cursor: pointer;
}

.btn-icon.theme-btn .icon {
  width: 20px;
  height: 20px;
}
</style>
