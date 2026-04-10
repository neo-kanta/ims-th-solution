import { useRuntimeConfig } from "#imports";

export function useApi() {
  const config = useRuntimeConfig();
  const authStore = useAuthStore();
  const baseURL = (process.server
    ? config.apiBaseUrl
    : config.public.apiBaseUrl) as string;

  async function apiFetch<T>(url: string, options: Record<string, unknown> = {}): Promise<T> {
    const headers: Record<string, string> = {
      ...((options.headers as Record<string, string> | undefined) || {}),
    };

    if (authStore.token) {
      headers.Authorization = `Bearer ${authStore.token}`;
    }

    try {
      return await $fetch<T>(url, {
        baseURL,
        ...options,
        headers,
      });
    } catch (error: any) {
      if (error?.status === 401 || error?.statusCode === 401) {
        authStore.clearAuth();
      }
      throw error;
    }
  }

  return { apiFetch };
}
