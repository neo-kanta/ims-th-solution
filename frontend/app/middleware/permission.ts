/**
 * Permission middleware — checks if the user has the required function permission.
 * Uses the `permission` meta field from definePageMeta().
 *
 * NOTE: This is frontend UX only. Real permission enforcement is server-side.
 */
export default defineNuxtRouteMiddleware((to) => {
  const requiredPermission = to.meta.permission as string | undefined

  if (!requiredPermission) {
    return // No permission required for this page
  }

  const authStore = useAuthStore()
  if (!authStore.hasPermission(requiredPermission)) {
    return navigateTo('/403') // Redirect to forbidden page
  }
})
