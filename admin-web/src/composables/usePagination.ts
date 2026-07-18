import { ref, computed } from 'vue'

export function usePagination(load: (page: number, size: number) => Promise<void>, size = 10) {
  const page = ref(1)
  const total = ref(0)
  const loading = ref(false)

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / size)))

  const pageNumbers = computed(() => {
    const t = totalPages.value
    const cur = page.value
    if (t <= 7) return Array.from({ length: t }, (_, i) => i + 1)
    const set = new Set([1, t, cur, cur - 1, cur + 1].filter((n) => n >= 1 && n <= t))
    return Array.from(set).sort((a, b) => a - b)
  })

  async function go(p: number) {
    if (p < 1 || p > totalPages.value || p === page.value) return
    page.value = p
    await load(page.value, size)
  }

  return { page, total, loading, totalPages, pageNumbers, go }
}
