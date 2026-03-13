/**
 * Auth middleware — protects routes that require authentication.
 * Redirects to /login if no auth token is present.
 */
export default defineNuxtRouteMiddleware(async (to) => {
  const authStore = useAuthStore()

  // During SSR or first client load, restore token if possible
  if (import.meta.client && !authStore.token) {
    const token = localStorage.getItem('ims_token')
    if (token) {
      authStore.token = token
      // Optimistically assume still valid, fetch profile in background
      authStore.fetchMe()
    }
  }

  // Define public routes
  const isPublicRoute = to.path.startsWith('/auth/login')

  if (!authStore.isAuthenticated && !isPublicRoute) {
    return navigateTo('/auth/login')
  }

  if (authStore.isAuthenticated && isPublicRoute) {
    return navigateTo('/')
  }
})
