<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import TheNavBar from '@/components/navbar.vue'
import TheFooter from '@/components/footer.vue'
import { fetchLinks, createShortLink, type ShortLink } from '@/api'

const links = ref<ShortLink[]>([])
const loading = ref(false)
const search = ref('')
const page = ref(1)
const pageSize = ref(8)

const filtered = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return links.value
  return links.value.filter(
    (l) =>
      l.longUrl.toLowerCase().includes(q) ||
      l.shortUrl.toLowerCase().includes(q) ||
      (l.note ?? '').toLowerCase().includes(q),
  )
})

const totalPages = computed(() =>
  Math.max(1, Math.ceil(filtered.value.length / pageSize.value)),
)
const pagedLinks = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return filtered.value.slice(start, start + pageSize.value)
})
const rangeStart = computed(() =>
  filtered.value.length === 0 ? 0 : (page.value - 1) * pageSize.value + 1,
)
const rangeEnd = computed(() =>
  Math.min(page.value * pageSize.value, filtered.value.length),
)

const copiedCode = ref<string | null>(null)

async function load() {
  loading.value = true
  try {
    links.value = await fetchLinks()
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  return d.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

function formatClicks(n: number): string {
  return n.toLocaleString('en-US')
}

async function copyLink(link: ShortLink) {
  try {
    await navigator.clipboard.writeText(`https://${link.shortUrl}`)
    copiedCode.value = link.code
    setTimeout(() => {
      if (copiedCode.value === link.code) copiedCode.value = null
    }, 2000)
  } catch {
    /* ignore */
  }
}

// --- Create New URL modal ---
const showCreate = ref(false)
const newUrl = ref('')
const creating = ref(false)
const createError = ref('')

function openCreate() {
  showCreate.value = true
  newUrl.value = ''
  createError.value = ''
}
function closeCreate() {
  showCreate.value = false
}

function validateUrl(raw: string): string | null {
  const trimmed = raw.trim()
  if (!trimmed) return null
  const withProto = /^https?:\/\//i.test(trimmed) ? trimmed : `https://${trimmed}`
  try {
    const u = new URL(withProto)
    if (u.protocol !== 'http:' && u.protocol !== 'https:') return null
    if (!u.hostname.includes('.')) return null
    return u.toString()
  } catch {
    return null
  }
}

async function submitCreate() {
  const normalized = validateUrl(newUrl.value)
  if (!normalized) {
    createError.value = 'Please enter a valid URL, e.g. https://example.com'
    return
  }
  creating.value = true
  createError.value = ''
  try {
    const res = await createShortLink(normalized)
    links.value.unshift(res)
    closeCreate()
  } catch {
    createError.value = 'Failed to create link. Please try again.'
  } finally {
    creating.value = false
  }
}

watch(search, onSearch)
watch(totalPages, (tp) => {
  if (page.value > tp) page.value = tp
})

onMounted(load)
</script>

<template>
  <div class="app-shell">
    <TheNavBar />

    <main class="app-main">
      <div class="urls">
        <div class="urls__inner">
          <!-- Header Section -->
          <div class="urls__head">
            <div>
              <h1 class="urls__title">URL Management</h1>
              <p class="urls__subtitle">
                Monitor and manage all your shortened links from a central dashboard.
              </p>
            </div>
            <button class="urls__create-btn" @click="openCreate">
              <span class="material-symbols-outlined">add</span>
              Create New URL
            </button>
          </div>

          <!-- Management Card -->
          <div class="urls__card">
            <!-- Table Controls -->
            <div class="urls__toolbar">
              <div class="urls__search">
                <span class="material-symbols-outlined">search</span>
                <input
                  v-model="search"
                  type="text"
                  class="field-input urls__search-input"
                  placeholder="Search by original or shortened URL..."
                />
              </div>
              <div class="urls__toolbar-actions">
                <button class="urls__btn-ghost">
                  <span class="material-symbols-outlined">filter_list</span>
                  Filter
                </button>
                <button class="urls__btn-ghost">
                  <span class="material-symbols-outlined">download</span>
                  Export
                </button>
              </div>
            </div>

            <!-- URLs Table -->
            <div class="urls__table-wrap">
              <table class="urls__table">
                <thead>
                  <tr class="urls__thead-row">
                    <th class="urls__th">Original URL</th>
                    <th class="urls__th">Shortened URL</th>
                    <th class="urls__th">Created At</th>
                    <th class="urls__th">Clicks</th>
                    <th class="urls__th urls__th--right">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="link in pagedLinks"
                    :key="link.code"
                    class="urls__row"
                  >
                    <td class="urls__td">
                      <div class="urls__orig">
                        <span class="urls__orig-url" :title="link.longUrl">{{
                          link.longUrl
                        }}</span>
                        <span v-if="link.note" class="urls__orig-note">{{
                          link.note
                        }}</span>
                      </div>
                    </td>
                    <td class="urls__td">
                      <div class="urls__short">
                        <span class="urls__short-url">{{ link.shortUrl }}</span>
                        <button
                          class="urls__copy"
                          :title="copiedCode === link.code ? 'Copied!' : 'Copy'"
                          @click="copyLink(link)"
                        >
                          <span class="material-symbols-outlined">{{
                            copiedCode === link.code ? 'done' : 'content_copy'
                          }}</span>
                        </button>
                      </div>
                    </td>
                    <td class="urls__td urls__td--muted">
                      {{ formatDate(link.createdAt) }}
                    </td>
                    <td class="urls__td">
                      <span class="urls__clicks">{{
                        formatClicks(link.clicks ?? 0)
                      }}</span>
                    </td>
                    <td class="urls__td urls__td--right">
                      <RouterLink
                        to="/logs"
                        class="urls__action"
                        title="View Logs"
                      >
                        <span class="material-symbols-outlined">analytics</span>
                      </RouterLink>
                    </td>
                  </tr>
                  <tr v-if="pagedLinks.length === 0">
                    <td colspan="5" class="urls__empty">
                      {{ loading ? 'Loading…' : 'No URLs found.' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Pagination -->
            <div class="urls__pager">
              <span class="urls__count"
                >Showing {{ rangeStart }} to {{ rangeEnd }} of
                {{ filtered.length }} URLs</span
              >
              <div class="urls__pager-nav">
                <button
                  class="urls__page-btn"
                  :disabled="page === 1"
                  @click="page > 1 && page--"
                >
                  <span class="material-symbols-outlined">chevron_left</span>
                </button>
                <button
                  v-for="p in totalPages"
                  :key="p"
                  class="urls__page-num"
                  :class="{ 'urls__page-num--active': p === page }"
                  @click="page = p"
                >
                  {{ p }}
                </button>
                <button
                  class="urls__page-btn"
                  :disabled="page === totalPages"
                  @click="page < totalPages && page++"
                >
                  <span class="material-symbols-outlined">chevron_right</span>
                </button>
              </div>
            </div>
          </div>

          <div class="urls__spacer"></div>
        </div>
      </div>
    </main>

    <TheFooter />

    <!-- Create New URL Modal -->
    <Teleport to="body">
      <div v-if="showCreate" class="modal" @click.self="closeCreate">
        <div class="modal__box" role="dialog" aria-modal="true">
          <div class="modal__head">
            <h2 class="modal__title">Create New URL</h2>
            <button class="modal__close" title="Close" @click="closeCreate">
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>
          <p class="modal__sub">
            Paste a long URL to generate a new short link.
          </p>
          <input
            v-model="newUrl"
            type="text"
            class="field-input"
            placeholder="https://example.com/very/long/path"
            @keyup.enter="submitCreate"
          />
          <div v-if="createError" class="modal__error">{{ createError }}</div>
          <div class="modal__actions">
            <button class="urls__btn-ghost" @click="closeCreate">Cancel</button>
            <button
              class="urls__create-btn"
              :disabled="creating"
              @click="submitCreate"
            >
              {{ creating ? 'Creating…' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style src="@/styles/urls.css" scoped></style>
