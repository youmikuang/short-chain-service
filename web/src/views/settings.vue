<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import TheNavBar from '@/components/navbar.vue'
import TheFooter from '@/components/footer.vue'
import Toast from '@/components/toast.vue'
import {
  fetchProfile,
  fetchSettings,
  saveProfile,
  updatePassword,
  saveSettings,
} from '@/api'

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

// --- Toast ---
const toast = reactive({ type: 'ok' as 'ok' | 'error', message: '' })
const toastKey = ref(0)
function showToast(type: 'ok' | 'error', message: string) {
  toast.type = type
  toast.message = message
  toastKey.value++
}

// --- Profile ---
const savingProfile = ref(false)
const emailError = ref('')

function clearEmailError() {
  emailError.value = ''
}

async function onSaveProfile() {
  emailError.value = ''
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailAddress.value)) {
    emailError.value = 'Please enter a valid email address.'
    return
  }
  savingProfile.value = true
  try {
    await saveProfile({
      fullName: fullName.value,
      email: emailAddress.value,
      avatar: avatar.value,
    })
    showToast('ok', 'Profile saved.')
  } catch {
    showToast('error', 'Failed to save profile. Please try again.')
  } finally {
    savingProfile.value = false
  }
}

// --- Password ---
const savingPassword = ref(false)
const currentPasswordError = ref('')
const newPasswordError = ref('')
const confirmPasswordError = ref('')

function clearPasswordErrors() {
  currentPasswordError.value = ''
  newPasswordError.value = ''
  confirmPasswordError.value = ''
}

async function onUpdatePassword() {
  currentPasswordError.value = ''
  newPasswordError.value = ''
  confirmPasswordError.value = ''
  if (!currentPassword.value) {
    currentPasswordError.value = 'Please enter your current password.'
    return
  }
  if (newPassword.value.length < 8) {
    newPasswordError.value = 'New password must be at least 8 characters.'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    confirmPasswordError.value = 'New password and confirmation do not match.'
    return
  }
  savingPassword.value = true
  try {
    await updatePassword({
      currentPassword: currentPassword.value,
      newPassword: newPassword.value,
    })
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    showToast('ok', 'Password updated.')
  } catch {
    showToast('error', 'Failed to update password. Please try again.')
  } finally {
    savingPassword.value = false
  }
}

// --- Preferences ---
const savingPrefs = ref(false)

async function onSavePreferences() {
  savingPrefs.value = true
  try {
    await saveSettings({
      emailNotif: emailNotif.value,
      securityAlerts: securityAlerts.value,
      marketingComm: marketingComm.value,
    })
    showToast('ok', 'Preferences saved.')
  } catch {
    showToast('error', 'Failed to save preferences. Please try again.')
  } finally {
    savingPrefs.value = false
  }
}
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
                    :class="{ 'settings__input--error': emailError }"
                    id="emailAddress"
                    type="email"
                    v-model="emailAddress"
                    @input="clearEmailError"
                  />
                  <span v-if="emailError" class="settings__field-error">{{ emailError }}</span>
                </div>
              </div>
            </div>
            <div class="settings__actions">
              <button
                class="settings__btn-primary"
                @click="onSaveProfile"
                :disabled="savingProfile"
              >
                {{ savingProfile ? 'Saving…' : 'Save Changes' }}
              </button>
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
                  :class="{ 'settings__input--error': currentPasswordError }"
                  id="currentPassword"
                  placeholder="••••••••"
                  type="password"
                  v-model="currentPassword"
                  @input="clearPasswordErrors"
                />
                <span v-if="currentPasswordError" class="settings__field-error">{{ currentPasswordError }}</span>
              </div>
              <div class="settings__field">
                <label class="settings__label" for="newPassword">New Password</label>
                <input
                  class="settings__input"
                  :class="{ 'settings__input--error': newPasswordError }"
                  id="newPassword"
                  placeholder="••••••••"
                  type="password"
                  v-model="newPassword"
                  @input="clearPasswordErrors"
                />
                <span v-if="newPasswordError" class="settings__field-error">{{ newPasswordError }}</span>
              </div>
              <div class="settings__field">
                <label class="settings__label" for="confirmPassword">Confirm New Password</label>
                <input
                  class="settings__input"
                  :class="{ 'settings__input--error': confirmPasswordError }"
                  id="confirmPassword"
                  placeholder="••••••••"
                  type="password"
                  v-model="confirmPassword"
                  @input="clearPasswordErrors"
                />
                <span v-if="confirmPasswordError" class="settings__field-error">{{ confirmPasswordError }}</span>
              </div>
            </div>
            <div class="settings__actions">
              <button
                class="settings__btn-primary"
                @click="onUpdatePassword"
                :disabled="savingPassword"
              >
                {{ savingPassword ? 'Updating…' : 'Update Password' }}
              </button>
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
              <button
                class="settings__btn-primary"
                @click="onSavePreferences"
                :disabled="savingPrefs"
              >
                {{ savingPrefs ? 'Saving…' : 'Save Preferences' }}
              </button>
            </div>
          </section>
        </div>
      </div>
    </main>

    <TheFooter />

    <Toast
      :key="toastKey"
      :type="toast.type"
      :message="toast.message"
      :duration="3000"
      @close="toast.message = ''"
    />
  </div>
</template>

<style src="@/styles/settings.css" scoped></style>
