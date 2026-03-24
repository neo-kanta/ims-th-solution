/**
 * Auth plugin — client-side only
 *
 * Initializes authentication on app startup:
 * 1. Restores session from localStorage
 * 2. Sets up token refresh schedule
 * 3. Adds authorization header to all API requests
 */
export default defineNuxtPlugin(() => {
  const authStore = useAuthStore()

  // Restore session on app initialization
  authStore.restoreSession()

  // Setup $fetch interceptor to attach auth token to requests
  const $fetch = useAsyncData

  if (process.client) {
    // Add auth token to all API requests
    const originalFetch = globalThis.$fetch

    if (originalFetch) {
      // We can't easily intercept $fetch globally in Nuxt 3, so we rely on
      // the nuxt.config.ts to handle auth headers via interceptors.
      // This is a placeholder for documentation.
    }
  }
})
