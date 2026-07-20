<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import type { TrafficPoint } from '@/api/admin'

const props = defineProps<{ data: TrafficPoint[] }>()

const H = 320
const PAD = { top: 20, right: 18, bottom: 30, left: 40 }

const wrap = ref<HTMLElement | null>(null)
const width = ref(800)
let ro: ResizeObserver | null = null

onMounted(() => {
  if (wrap.value) {
    width.value = wrap.value.clientWidth || 800
    ro = new ResizeObserver((entries) => {
      for (const e of entries) width.value = e.contentRect.width || 800
    })
    ro.observe(wrap.value)
  }
})
onBeforeUnmount(() => ro?.disconnect())

const plotW = computed(() => Math.max(1, width.value - PAD.left - PAD.right))
const plotH = H - PAD.top - PAD.bottom

const WEEKDAYS = ['SUN', 'MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT']

const labels = computed(() =>
  props.data.map((d) => {
    const t = new Date(d.date)
    return isNaN(t.getTime()) ? d.date.slice(5) : WEEKDAYS[t.getDay()]
  }),
)

const maxValue = computed(() => {
  let m = 0
  for (const d of props.data) {
    if (d.actions > m) m = d.actions
    if (d.rpc > m) m = d.rpc
  }
  // 向上取整到「漂亮」的刻度，至少留 4 格
  if (m <= 0) return 4
  const step = Math.pow(10, Math.floor(Math.log10(m)))
  return Math.ceil(m / step) * step
})

function xAt(i: number): number {
  const n = props.data.length
  if (n <= 1) return PAD.left + plotW.value / 2
  return PAD.left + (i * plotW.value) / (n - 1)
}
function yAt(v: number): number {
  return PAD.top + plotH - (v / maxValue.value) * plotH
}

// Catmull-Rom → 三次贝塞尔，生成带曲率的平滑路径
function smoothPath(values: number[]): string {
  const pts = values.map((v, i) => ({ x: xAt(i), y: yAt(v) }))
  if (pts.length === 0) return ''
  if (pts.length === 1) return `M ${pts[0]!.x},${pts[0]!.y}`
  let d = `M ${pts[0]!.x},${pts[0]!.y}`
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = pts[i - 1] ?? pts[i]!
    const p1 = pts[i]!
    const p2 = pts[i + 1]!
    const p3 = pts[i + 2] ?? p2!
    const c1x = p1.x + (p2.x - p0.x) / 6
    const c1y = p1.y + (p2.y - p0.y) / 6
    const c2x = p2.x - (p3.x - p1.x) / 6
    const c2y = p2.y - (p3.y - p1.y) / 6
    d += ` C ${c1x.toFixed(2)},${c1y.toFixed(2)} ${c2x.toFixed(2)},${c2y.toFixed(2)} ${p2.x.toFixed(2)},${p2.y.toFixed(2)}`
  }
  return d
}

function areaPath(values: number[]): string {
  const line = smoothPath(values)
  if (!line) return ''
  const lastX = xAt(values.length - 1)
  const firstX = xAt(0)
  return `${line} L ${lastX.toFixed(2)},${(PAD.top + plotH).toFixed(2)} L ${firstX.toFixed(2)},${(PAD.top + plotH).toFixed(2)} Z`
}

const actionsLine = computed(() => smoothPath(props.data.map((d) => d.actions)))
const rpcLine = computed(() => smoothPath(props.data.map((d) => d.rpc)))
const actionsArea = computed(() => areaPath(props.data.map((d) => d.actions)))
const rpcArea = computed(() => areaPath(props.data.map((d) => d.rpc)))

const gridLines = computed(() => {
  const n = 4
  return Array.from({ length: n + 1 }, (_, i) => {
    const v = (maxValue.value / n) * i
    return { y: yAt(v), label: Math.round(v).toString() }
  })
})

// 悬停交互
const hover = ref<number | null>(null)
function onMove(e: MouseEvent) {
  const rect = (e.currentTarget as SVGElement).getBoundingClientRect()
  const relX = ((e.clientX - rect.left) / rect.width) * width.value
  const n = props.data.length
  if (n <= 1) {
    hover.value = n - 1
    return
  }
  const idx = Math.round(((relX - PAD.left) / plotW.value) * (n - 1))
  hover.value = Math.min(n - 1, Math.max(0, idx))
}
function onLeave() {
  hover.value = null
}

const hoverX = computed(() => (hover.value != null ? xAt(hover.value) : 0))
</script>

<template>
  <div ref="wrap" class="w-full">
    <!-- Legend -->
    <div class="flex items-center gap-5 mb-3">
      <span class="flex items-center gap-2 text-body-sm text-secondary">
        <span class="w-3 h-3 rounded-full bg-primary"></span> Action Logs
      </span>
      <span class="flex items-center gap-2 text-body-sm text-secondary">
        <span class="w-3 h-3 rounded-full bg-tertiary"></span> RPC Logs
      </span>
    </div>

    <svg
      :viewBox="`0 0 ${width} ${H}`"
      class="w-full"
      :style="{ height: H + 'px' }"
      @mousemove="onMove"
      @mouseleave="onLeave"
    >
      <defs>
        <linearGradient id="gradActions" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" :style="{ stopColor: 'rgb(var(--color-primary))' }" stop-opacity="0.22" />
          <stop offset="100%" :style="{ stopColor: 'rgb(var(--color-primary))' }" stop-opacity="0" />
        </linearGradient>
        <linearGradient id="gradRpc" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" :style="{ stopColor: 'rgb(var(--color-tertiary))' }" stop-opacity="0.18" />
          <stop offset="100%" :style="{ stopColor: 'rgb(var(--color-tertiary))' }" stop-opacity="0" />
        </linearGradient>
      </defs>

      <!-- Grid + Y labels -->
      <g>
        <template v-for="(g, i) in gridLines" :key="i">
          <line
            :x1="PAD.left"
            :x2="width - PAD.right"
            :y1="g.y"
            :y2="g.y"
            :style="{ stroke: 'rgb(var(--color-surface-variant))' }"
            stroke-width="1"
            vector-effect="non-scaling-stroke"
          />
          <text
            :x="PAD.left - 8"
            :y="g.y + 3"
            text-anchor="end"
            class="fill-secondary"
            style="font-size: 10px"
          >
            {{ g.label }}
          </text>
        </template>
      </g>

      <!-- Areas -->
      <path :d="actionsArea" fill="url(#gradActions)" />
      <path :d="rpcArea" fill="url(#gradRpc)" />

      <!-- Smooth lines -->
      <path
        :d="actionsLine"
        fill="none"
        :style="{ stroke: 'rgb(var(--color-primary))' }"
        stroke-width="2.5"
        stroke-linecap="round"
        vector-effect="non-scaling-stroke"
      />
      <path
        :d="rpcLine"
        fill="none"
        :style="{ stroke: 'rgb(var(--color-tertiary))' }"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-dasharray="6 4"
        vector-effect="non-scaling-stroke"
      />

      <!-- X labels -->
      <text
        v-for="(lb, i) in labels"
        :key="'x' + i"
        :x="xAt(i)"
        :y="H - 10"
        text-anchor="middle"
        class="fill-secondary"
        style="font-size: 10px"
      >
        {{ lb }}
      </text>

      <!-- Hover guide + points -->
      <template v-if="hover != null">
        <line
          :x1="hoverX"
          :x2="hoverX"
          :y1="PAD.top"
          :y2="PAD.top + plotH"
          :style="{ stroke: 'rgb(var(--color-outline))' }"
          stroke-width="1"
          stroke-dasharray="3 3"
          vector-effect="non-scaling-stroke"
        />
        <circle :cx="xAt(hover)" :cy="yAt(data[hover]?.actions ?? 0)" r="4" :style="{ fill: 'rgb(var(--color-primary))' }" />
        <circle :cx="xAt(hover)" :cy="yAt(data[hover]?.rpc ?? 0)" r="4" :style="{ fill: 'rgb(var(--color-tertiary))' }" />
      </template>
    </svg>

    <!-- Tooltip -->
    <div
      v-if="hover != null && data[hover]"
      class="mt-2 flex items-center justify-center gap-6 text-body-sm"
    >
      <span class="text-secondary">{{ labels[hover] }} · {{ data[hover]?.date }}</span>
      <span class="flex items-center gap-1.5 text-on-surface">
        <span class="w-2.5 h-2.5 rounded-full bg-primary"></span>
        Actions: <b>{{ data[hover]?.actions }}</b>
      </span>
      <span class="flex items-center gap-1.5 text-on-surface">
        <span class="w-2.5 h-2.5 rounded-full bg-tertiary"></span>
        RPC: <b>{{ data[hover]?.rpc }}</b>
      </span>
    </div>
  </div>
</template>

<style scoped></style>
