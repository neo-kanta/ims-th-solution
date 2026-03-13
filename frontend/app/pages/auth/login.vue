<script setup lang="ts">
definePageMeta({
  layout: 'auth',
})

const authStore = useAuthStore()
const router = useRouter()

const username = ref('')
const password = ref('')

async function handleLogin() {
  if (!username.value || !password.value) return

  const success = await authStore.login({
    username: username.value,
    password: password.value,
  })

  if (success) {
    router.push('/')
  }
}
</script>

<template>
  <div>
    <h2 class="mb-6 text-xl font-semibold text-gray-800 text-center">Sign in to your account</h2>

    <form @submit.prevent="handleLogin" class="space-y-4">
      <div v-if="authStore.error" class="p-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg">
        {{ authStore.error }}
      </div>

      <div>
        <label for="username" class="label">Username</label>
        <input
          id="username"
          v-model="username"
          type="text"
          required
          autofocus
          class="input-field"
          placeholder="Enter your username"
          :disabled="authStore.isLoading"
        />
      </div>

      <div>
        <label for="password" class="label">Password</label>
        <input
          id="password"
          v-model="password"
          type="password"
          required
          class="input-field"
          placeholder="••••••••"
          :disabled="authStore.isLoading"
        />
      </div>

      <div class="pt-2">
        <button
          type="submit"
          class="btn-primary w-full flex items-center justify-center transition-all duration-200"
          :disabled="authStore.isLoading || !username || !password"
          :class="{ 'opacity-75 cursor-not-allowed': authStore.isLoading }"
        >
          <span v-if="authStore.isLoading" class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
          {{ authStore.isLoading ? 'Signing in...' : 'Sign in' }}
        </button>
      </div>
    </form>
    
    <div class="mt-6 text-center text-sm text-gray-500">
      Need an account? Contact your system administrator.
    </div>
  </div>
</template>
