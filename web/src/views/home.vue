<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import TheNavBar from '@/components/navbar.vue'
import TheFooter from '@/components/footer.vue'
import { createShortLink } from '@/api'

const url = ref('')
const shortUrl = ref('')
const copied = ref(false)
const loading = ref(false)
const error = ref('')
const showLoginPrompt = ref(false)

// --- Validate that the input is a legal http/https URL ---
function validateUrl(raw: string): string | null {
  const trimmed = raw.trim()
  if (!trimmed) return null
  const withProto = /^https?:\/\//i.test(trimmed) ? trimmed : `https://${trimmed}`
  try {
    const u = new URL(withProto)
    if (u.protocol !== 'http:' && u.protocol !== 'https:') return null
    if (!u.hostname.includes('.') && u.hostname !== 'localhost') return null
    return u.toString()
  } catch {
    return null
  }
}

// --- Frontend click rate limit: max 10 shorten clicks per minute ---
const CLICK_KEY = 'slink_shorten_clicks'
const CLICK_WINDOW = 60_000
const CLICK_LIMIT = 10

function loadClicks(): number[] {
  try {
    const raw = localStorage.getItem(CLICK_KEY)
    if (!raw) return []
    const arr = JSON.parse(raw) as number[]
    const now = Date.now()
    return arr.filter((t) => now - t < CLICK_WINDOW)
  } catch {
    return []
  }
}
function saveClicks(arr: number[]) {
  localStorage.setItem(CLICK_KEY, JSON.stringify(arr))
}

const clickTimes = ref<number[]>(loadClicks())

async function shorten() {
  if (loading.value) return

  const normalized = validateUrl(url.value)
  if (!normalized) {
    error.value = '请输入合法的 URL，例如 https://example.com'
    return
  }

  // Record the click and enforce the 1-minute limit.
  const now = Date.now()
  clickTimes.value = clickTimes.value.filter((t) => now - t < CLICK_WINDOW)
  clickTimes.value.push(now)
  saveClicks(clickTimes.value)
  if (clickTimes.value.length > CLICK_LIMIT) {
    showLoginPrompt.value = true
  }

  loading.value = true
  error.value = ''
  try {
    const res = await createShortLink(normalized)
    shortUrl.value = res.shortUrl
    copied.value = false
  } catch {
    error.value = '短链生成失败，请稍后重试。'
  } finally {
    loading.value = false
  }
}

async function copy() {
  if (!shortUrl.value) return
  try {
    await navigator.clipboard.writeText(`https://${shortUrl.value}`)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    /* ignore */
  }
}
</script>

<template>
  <div class="app-shell">
    <TheNavBar />

    <main class="app-main">
      <!-- Hero Section -->
      <section class="hero">
        <div class="hero__inner">
          <!-- Input Area -->
          <div class="hero__input-wrap">
            <div class="hero__input-box">
              <div class="hero__input-icon">
                <span class="material-symbols-outlined">link</span>
              </div>
              <input
                v-model="url"
                type="text"
                placeholder="Paste your long URL here"
                class="hero__input"
                @keyup.enter="shorten"
              />
              <button class="hero__btn" :disabled="loading" @click="shorten">
                {{ loading ? 'Shortening…' : 'Shorten Now' }}
              </button>
            </div>

            <div v-if="shortUrl" class="hero__result">
              <span class="material-symbols-outlined hero__result-icon">check_circle</span>
              <a :href="`https://${shortUrl}`" class="hero__result-link">{{ shortUrl }}</a>
              <button
                class="hero__copy"
                :title="copied ? 'Copied!' : 'Copy'"
                @click="copy"
              >
                <span class="material-symbols-outlined">{{ copied ? 'done' : 'content_copy' }}</span>
              </button>
            </div>

            <div v-if="error" class="hero__error">{{ error }}</div>

            <div class="hero__features">
              <span class="hero__feature"
                ><span class="material-symbols-outlined">verified</span> No credit card
                required</span
              >
              <span class="hero__feature"
                ><span class="material-symbols-outlined">lock</span> SSL encrypted</span
              >
            </div>
          </div>
        </div>
      </section>
    </main>

    <TheFooter />
  </div>
</template>

<style src="@/styles/home.css" scoped></style>
