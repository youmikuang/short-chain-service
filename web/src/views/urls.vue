<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import TheNavBar from '@/components/navbar.vue'
import TheFooter from '@/components/footer.vue'
import { fetchLinks, CreateSlink, type slink } from '@/api'

const links = ref<slink[]>([])
const loading = ref(false)
const search = ref('')
const page = ref(1)
const pageSize = ref(10)

// Time sort: null = original order, 'desc' = newest first, 'asc' = oldest first
const sortDir = ref<'asc' | 'desc' | null>(null)

function toggleTimeSort() {
  sortDir.value =
    sortDir.value === null ? 'desc' : sortDir.value === 'desc' ? 'asc' : null
  page.value = 1
  load()
}

const filtered = computed(() => links.value)

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
    links.value = await fetchLinks({
      search: search.value.trim(),
      sort: sortDir.value ?? '',
    })
  } finally {
    loading.value = false
  }
}

let searchTimer: ReturnType<typeof setTimeout> | null = null
function onSearch() {
  page.value = 1
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => load(), 300)
}

function formatClicks(n: number): string {
  return n.toLocaleString('en-US')
}

function sourceText(s?: string): string {
  if (s === 'rpc') return 'RPC'
  if (s === 'web') return 'Web'
  return s || '-'
}

function sourceClass(s?: string): string {
  if (s === 'rpc') return 'badge badge-api'
  if (s === 'web') return 'badge badge-web'
  return 'badge'
}

function downloadCsv(filename: string, data: string[][]) {
  const escape = (v: string) => `"${String(v).replace(/"/g, '""')}"`
  const csv = data.map((r) => r.map(escape).join(',')).join('\r\n')
  const blob = new Blob(['﻿' + csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

function exportCsv() {
  const header = ['Original URL', 'Shortened URL', 'Created At', 'Clicks']
  const body = links.value.map((l) => [
    l.longUrl,
    l.sUrl,
    l.createdAt,
    String(l.clicks ?? 0),
  ])
  downloadCsv('urls.csv', [header, ...body])
}

async function copyLink(link: slink) {
  try {
    await navigator.clipboard.writeText(link.sUrl)
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
    const res = await CreateSlink(normalized)
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
                  placeholder="Search Original URL"
                />
              </div>
              <div class="urls__toolbar-actions">
                <button
                  class="urls__btn-ghost"
                  :class="{
                    'urls__btn-ghost--active': sortDir !== null,
                    asc: sortDir === 'asc',
                  }"
                  @click="toggleTimeSort"
                >
                  <span class="material-symbols-outlined urls__sort-icon">Sort</span>
                  Time Sort
                </button>
                <button class="urls__btn-ghost" @click="exportCsv">
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
                    <th class="urls__th">Source</th>
                    <th class="urls__th">Created At</th>
                    <th class="urls__th">Clicks</th>
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
                      </div>
                    </td>
                    <td class="urls__td">
                      <div class="urls__short">
                        <a
                          :href="link.sUrl"
                          class="urls__short-url"
                          target="_blank"
                          rel="noopener noreferrer"
                          :title="`Open ${link.sUrl} in a new tab`"
                          >{{ link.sUrl }}</a
                        >
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
                    <td class="urls__td">
                      <span :class="sourceClass(link.source)">{{ sourceText(link.source) }}</span>
                    </td>
                    <td class="urls__td urls__td--muted">
                      {{ link.createdAt }}
                    </td>
                    <td class="urls__td">
                      <span class="urls__clicks">{{
                        formatClicks(link.clicks ?? 0)
                      }}</span>
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
              <div class="urls__pager-left">
                <div class="urls__pager-size">
                  <span>Rows per page:</span>
                  <select v-model.number="pageSize" class="urls__select">
                    <option :value="10">10</option>
                    <option :value="25">25</option>
                    <option :value="50">50</option>
                  </select>
                </div>
                <span class="urls__count"
                  >Showing {{ rangeStart }} to {{ rangeEnd }} of
                  {{ filtered.length }} URLs</span
                >
              </div>
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
