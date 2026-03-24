/**
 * useFetch composable — wrapper around $fetch with automatic auth token injection
 *
 * Usage:
 * const { fetch } = useFetch()
 * const data = await fetch('/api/v1/workflow', { method: 'GET' })
 */
export function useFetch() {
  const authStore = useAuthStore()
  const config = useRuntimeConfig()

  /**
   * Authenticated fetch — automatically injects Authorization header
   */
  async function fetch<T = any>(url: string, options?: any): Promise<T> {
    const baseURL = config.public.apiBaseUrl || 'http://localhost:8080/api/v1'

    // Ensure URL is absolute
    const fullUrl = url.startsWith('http') ? url : `${baseURL}${url.startsWith('/api') ? url : `/api/v1${url}`}`

    // Build request options
    const requestOptions = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      } as Record<string, string>,
    }

    // Inject auth token if available
    if (authStore.token) {
      requestOptions.headers['Authorization'] = `Bearer ${authStore.token}`
    }

    try {
      const response = await $fetch<T>(fullUrl, requestOptions)
      return response
    } catch (error: any) {
      // If 401 (unauthorized), logout and redirect
      if (error.status === 401) {
        authStore.logout()
        if (process.client) {
          await navigateTo('/auth/login')
        }
      }
      throw error
    }
  }

  return { fetch }
}
