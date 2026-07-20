<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import ThemeToggle from '@/components/theme.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const email = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)

// GitHub OAuth 回调：后端 302 跳回 /login?token=...&user_id=...&nickname=...
// 此处读取并落库，然后进入首页。
onMounted(async () => {
  const token = route.query.token as string | undefined
  if (token) {
    await auth.finishGithubLogin(
      token,
      (route.query.nickname as string) || '',
    )
    router.replace('/')
  }
})

async function doLogin(provider: 'github' | 'email') {
  if (loading.value) return
  loading.value = true
  try {
    await auth.login(provider, {
      email: email.value.trim(),
      password: password.value,
    })
    router.push('/')
  } catch {
    /* ignore for demo */
  } finally {
    loading.value = false
  }
}

function loginWithGitHub() {
  doLogin('github')
}

function continueWithEmail() {
  if (!email.value.trim()) return
  doLogin('email')
}
</script>

<template>
  <div class="login">
    <!-- Top Navigation -->
    <nav class="login__nav">
      <div class="login__nav-inner">
        <a href="/" class="login__brand">
          SLink
        </a>
        <ThemeToggle />
      </div>
    </nav>

    <!-- Main Content -->
    <main class="login__main">
      <!-- Login Card -->
      <div class="login__card">
        <!-- Branding -->
        <div class="login__brand-block">
          <div class="login__logo">
            <span class="material-symbols-outlined" style="font-size: 28px; font-variation-settings: 'FILL' 1;">link</span>
          </div>
          <h1 class="login__title">Welcome to SLink</h1>
          <p class="login__subtitle">Log in to manage your shortened URLs.</p>
        </div>

        <!-- Primary Auth Action -->
        <button class="login__github" :disabled="loading" @click="loginWithGitHub">
          <svg aria-hidden="true" viewBox="0 0 24 24">
            <path
              d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
            ></path>
          </svg>
          <span>Continue with GitHub</span>
        </button>

        <!-- Divider -->
        <div class="login__divider">
          <div class="login__divider-line"></div>
          <span class="login__divider-text">or</span>
          <div class="login__divider-line"></div>
        </div>

        <!-- Email Fallback -->
        <form class="login__form" @submit.prevent="continueWithEmail">
          <div>
            <label class="sr-only" for="email">Email address</label>
            <input
              id="email"
              v-model="email"
              name="email"
              type="email"
              placeholder="Email address"
              class="field-input"
            />
          </div>
          <div>
            <label class="sr-only" for="password">Password</label>
            <div class="login__password">
              <input
                id="password"
                v-model="password"
                name="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="Password"
                class="field-input"
              />
              <button
                type="button"
                class="login__password-toggle"
                :aria-label="showPassword ? 'Hide password' : 'Show password'"
                @click="showPassword = !showPassword"
              >
                <span class="material-symbols-outlined">{{ showPassword ? 'visibility_off' : 'visibility' }}</span>
              </button>
            </div>
          </div>
          <p class="login__hint">
            No account yet? We'll create one for you automatically and sign you in.
          </p>
          <button type="submit" class="btn btn-primary btn-block" :disabled="loading">Continue with Email</button>
          <p class="login__terms">
            By continuing, you agree to our
            <a href="#">Terms of Service</a> and
            <a href="#">Privacy Policy</a>.
          </p>
        </form>
      </div>
    </main>
  </div>
</template>

<style src="@/styles/login.css" scoped></style>
