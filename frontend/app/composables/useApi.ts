import { useRuntimeConfig } from "#imports";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function getResponseStatus(error: unknown): number | null {
  if (!isRecord(error)) {
    return null;
  }

  const status = error.status;
  if (typeof status === "number") {
    return status;
  }

  const statusCode = error.statusCode;
  return typeof statusCode === "number" ? statusCode : null;
}

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
    } catch (error) {
      if (getResponseStatus(error) === 401) {
        authStore.clearAuth();
      }
      throw error;
    }
  }

  return { apiFetch };
}
