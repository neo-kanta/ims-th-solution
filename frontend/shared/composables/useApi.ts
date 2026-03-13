/**
 * API composable — provides configured API client using $fetch.
 * All API calls should go through this composable for consistent
 * error handling, auth headers, and base URL configuration.
 */
import { useRuntimeConfig } from '#imports'
import { useAuthStore } from '../stores/useAuthStore'

export function useApi() {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()
  const baseURL = config.public.apiBaseUrl as string

  async function apiFetch<T>(url: string, options: Record<string, unknown> = {}): Promise<T> {
    const headers: Record<string, string> = {
      ...(options.headers as Record<string, string> || {}),
    }

    if (authStore.token) {
      headers['Authorization'] = `Bearer ${authStore.token}`
    }

    return $fetch<T>(url, {
      baseURL,
      ...options,
      headers,
    })
  }

  return { apiFetch }
}
