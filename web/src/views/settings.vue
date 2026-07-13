<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import TheNavBar from '@/components/navbar.vue'
import TheFooter from '@/components/footer.vue'
import { fetchProfile, fetchSettings } from '@/api'

const fullName = ref('')
const emailAddress = ref('')
const avatar = ref('')
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

const emailNotif = ref(false)
const securityAlerts = ref(false)
const marketingComm = ref(false)

const root = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null

onMounted(async () => {
  const [p, s] = await Promise.all([fetchProfile(), fetchSettings()])
  fullName.value = p.fullName
  emailAddress.value = p.email
  avatar.value = p.avatar
  emailNotif.value = s.emailNotif
  securityAlerts.value = s.securityAlerts
  marketingComm.value = s.marketingComm

  const el = root.value
  if (!el) return
  const sections = Array.from(el.querySelectorAll('section')) as HTMLElement[]
  const navLinks = Array.from(el.querySelectorAll('aside nav a')) as HTMLAnchorElement[]

  observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return
        const id = entry.target.getAttribute('id')
        navLinks.forEach((link) => {
          const active = link.getAttribute('href') === `#${id}`
          link.classList.toggle('settings__nav-link--active', active)
        })
      })
    },
    { root: null, rootMargin: '-20% 0px -80% 0px', threshold: 0 },
  )

  sections.forEach((section) => observer?.observe(section))
})

onBeforeUnmount(() => observer?.disconnect())
</script>

<template>
  <div ref="root" class="app-shell">
    <TheNavBar />

    <main class="settings__main">
      <div class="settings__page-head">
        <h1 class="settings__page-title">Personal Settings</h1>
        <p class="settings__page-sub">
          Manage your account preferences, profile details, and security.
        </p>
      </div>

      <div class="settings__grid">
        <!-- Side Navigation -->
        <aside class="settings__aside">
          <nav class="settings__nav">
            <a class="settings__nav-link settings__nav-link--active" href="#profile">
              <span class="material-symbols-outlined">person</span> Profile
            </a>
            <a class="settings__nav-link" href="#security">
              <span class="material-symbols-outlined">lock</span> Security
            </a>
            <a class="settings__nav-link" href="#preferences">
              <span class="material-symbols-outlined">tune</span> Preferences
            </a>
          </nav>
        </aside>

        <!-- Settings Panels -->
        <div class="settings__panels">
          <!-- Profile -->
          <section class="settings__section" id="profile">
            <div class="settings__section-head">
              <h2 class="settings__section-title">Profile</h2>
              <p class="settings__section-sub">Update your personal information and avatar.</p>
            </div>
            <div class="settings__profile-row">
              <div class="settings__avatar-col">
                <img class="settings__avatar" :src="avatar" alt="User avatar" />
                <button class="settings__avatar-btn">Change Avatar</button>
              </div>
              <div class="settings__fields">
                <div class="settings__field">
                  <label class="settings__label" for="fullName">Full Name</label>
                  <input class="settings__input" id="fullName" type="text" v-model="fullName" />
                </div>
                <div class="settings__field">
                  <label class="settings__label" for="emailAddress">Email Address</label>
                  <input
                    class="settings__input"
                    id="emailAddress"
                    type="email"
                    v-model="emailAddress"
                  />
                </div>
              </div>
            </div>
            <div class="settings__actions">
              <button class="settings__btn-primary">Save Changes</button>
            </div>
          </section>

          <!-- Security -->
          <section class="settings__section" id="security">
            <div class="settings__section-head">
              <h2 class="settings__section-title">Security</h2>
              <p class="settings__section-sub">Manage your password and security preferences.</p>
            </div>
            <div class="settings__fields" style="max-width: 512px; margin-bottom: 32px">
              <div class="settings__field">
                <label class="settings__label" for="currentPassword">Current Password</label>
                <input
                  class="settings__input"
                  id="currentPassword"
                  placeholder="••••••••"
                  type="password"
                  v-model="currentPassword"
                />
              </div>
              <div class="settings__field">
                <label class="settings__label" for="newPassword">New Password</label>
                <input
                  class="settings__input"
                  id="newPassword"
                  placeholder="••••••••"
                  type="password"
                  v-model="newPassword"
                />
              </div>
              <div class="settings__field">
                <label class="settings__label" for="confirmPassword">Confirm New Password</label>
                <input
                  class="settings__input"
                  id="confirmPassword"
                  placeholder="••••••••"
                  type="password"
                  v-model="confirmPassword"
                />
              </div>
            </div>
            <div class="settings__actions">
              <button class="settings__btn-primary">Update Password</button>
            </div>
          </section>

          <!-- Preferences -->
          <section class="settings__section" id="preferences">
            <div class="settings__section-head">
              <h2 class="settings__section-title">Preferences</h2>
              <p class="settings__section-sub">Control your app experience and notifications.</p>
            </div>
            <div style="display: flex; flex-direction: column; gap: 24px; margin-bottom: 32px">
              <!-- Toggle: Email Notifications -->
              <div class="settings__toggle">
                <div class="settings__toggle-text">
                  <span class="settings__toggle-title">Email Notifications</span>
                  <span class="settings__toggle-desc"
                    >Receive weekly reports on your link performance.</span
                  >
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="emailNotif" />
                  <span class="switch__track"></span>
                </label>
              </div>
              <!-- Toggle: Security Alerts -->
              <div class="settings__toggle">
                <div class="settings__toggle-text">
                  <span class="settings__toggle-title">Security Alerts</span>
                  <span class="settings__toggle-desc"
                    >Get notified about suspicious login attempts.</span
                  >
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="securityAlerts" />
                  <span class="switch__track"></span>
                </label>
              </div>
              <!-- Toggle: Marketing Communications -->
              <div class="settings__toggle">
                <div class="settings__toggle-text">
                  <span class="settings__toggle-title">Marketing Communications</span>
                  <span class="settings__toggle-desc"
                    >Receive updates about new features and promotions.</span
                  >
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="marketingComm" />
                  <span class="switch__track"></span>
                </label>
              </div>
            </div>
            <div class="settings__actions">
              <button class="settings__btn-primary">Save Preferences</button>
            </div>
          </section>
        </div>
      </div>
    </main>

    <TheFooter />
  </div>
</template>

<style src="@/styles/settings.css" scoped></style>
