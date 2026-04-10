import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { useCookie } from "#imports";

import type {
  AuthUser,
  AuthUserPayload,
  UserPermissions,
} from "../features/auth/types";
import type { ApiResponse } from "../types/api.types";
import { useApi } from "../composables/useApi";
import {
  createClearedAuthState,
  createEmptyPermissions,
  getTokenExpiry,
  mapAuthUser,
} from "../features/auth/lib/auth";

export const useAuthStore = defineStore("auth", () => {
  const tokenCookie = useCookie<string | null>("auth_token", {
    default: () => null,
    maxAge: 60 * 60 * 24 * 7,
    sameSite: "lax",
    path: "/",
  });

  const expiresAtCookie = useCookie<string | null>("auth_token_expires_at", {
    default: () => null,
    sameSite: "lax",
    path: "/",
  });

  const user = ref<AuthUser | null>(null);
  const permissions = ref<UserPermissions>(createEmptyPermissions());
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  let restorePromise: Promise<boolean> | null = null;

  const isTokenExpired = computed(() => {
    if (!tokenCookie.value) {
      return true;
    }

    const effectiveExpiry =
      expiresAtCookie.value || getTokenExpiry(tokenCookie.value);
    if (!effectiveExpiry) {
      return false;
    }

    return new Date(effectiveExpiry).getTime() <= Date.now();
  });

  const isAuthenticated = computed(
    () => Boolean(tokenCookie.value) && !isTokenExpired.value,
  );
  const token = computed(() => tokenCookie.value);

  function setAuth(
    authToken: string,
    expiresAt: string,
    authUser: AuthUser,
    userPermissions: UserPermissions,
  ) {
    const effectiveExpiry = expiresAt || getTokenExpiry(authToken);
    const expiresAtMs = effectiveExpiry ? new Date(effectiveExpiry).getTime() : NaN;

    tokenCookie.value = authToken;

    if (Number.isFinite(expiresAtMs)) {
      expiresAtCookie.value = effectiveExpiry;
    } else {
      expiresAtCookie.value = effectiveExpiry || null;
    }

    user.value = authUser;
    permissions.value = userPermissions;
    error.value = null;
  }

  function clearAuth(nextError: string | null = null) {
    const clearedState = createClearedAuthState(nextError);

    tokenCookie.value = clearedState.token;
    expiresAtCookie.value = clearedState.expiresAt;
    user.value = clearedState.user;
    permissions.value = clearedState.permissions;
    error.value = clearedState.error;
  }

  async function login(credentials: LoginCredentials) {
    const { apiFetch } = useApi();
    isLoading.value = true;
    error.value = null;

    try {
      const response = await apiFetch<ApiResponse<LoginResponse>>("/auth/login", {
        method: "POST",
        body: credentials,
      });

      const data = response.data;
      setAuth(
        data.access_token,
        data.access_token_expires_at,
        mapAuthUser(data.user),
        data.permissions,
      );

      return true;
    } catch (err: any) {
      clearAuth(err?.data?.message || err?.message || "Login failed");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  async function logout() {
    clearAuth();
  }

  async function fetchMe() {
    if (!tokenCookie.value || isTokenExpired.value) {
      clearAuth();
      return false;
    }

    const { apiFetch } = useApi();
    isLoading.value = true;
    error.value = null;

    try {
      const response = await apiFetch<ApiResponse<{ user: LoginResponseUser; permissions: UserPermissions }>>("/auth/me");

      user.value = mapAuthUser(response.data.user);
      permissions.value = response.data.permissions;
      return true;
    } catch (err: any) {
      clearAuth(err?.data?.message || err?.message || "Failed to restore session");
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  function hasPermission(code: string): boolean {
    return permissions.value.functions.includes(code);
  }

  function hasContract(contractId: string): boolean {
    return permissions.value.contracts.includes(contractId);
  }

  function filterByPermission<T extends { requiresPermission?: string }>(items: T[]): T[] {
    return items.filter((item) => (
      !item.requiresPermission || hasPermission(item.requiresPermission)
    ));
  }

  async function restoreSession() {
    if (restorePromise) {
      return restorePromise;
    }

    if (!tokenCookie.value || isTokenExpired.value) {
      clearAuth();
      return false;
    }

    if (!user.value) {
      restorePromise = (async () => {
        try {
          return await fetchMe();
        } finally {
          restorePromise = null;
        }
      })();

      return restorePromise;
    }

    return true;
  }

  return {
    token,
    user,
    permissions,
    isLoading,
    error,
    isAuthenticated,
    isTokenExpired,
    setAuth,
    clearAuth,
    login,
    logout,
    fetchMe,
    hasPermission,
    hasContract,
    filterByPermission,
    restoreSession,
  };
});

export interface LoginCredentials {
  username: string;
  password: string;
  totp_code: string;
  recovery_code: string;
}

interface LoginResponse {
  access_token: string;
  access_token_expires_at: string;
  user: LoginResponseUser;
  permissions: UserPermissions;
}

interface LoginResponseUser extends AuthUserPayload {
  groups: string[];
}
