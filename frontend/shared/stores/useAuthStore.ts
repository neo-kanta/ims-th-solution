import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

import type { ApiResponse } from '../types/api.types'
import { useApi } from '../composables/useApi'

// Auth store manages authentication state.
export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const user = ref<AuthUser | null>(null)
  const permissions = ref<UserPermissions>({
    functions: [],
    contracts: [],
  })

  const isLoading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  function setAuth(authToken: string, authUser: AuthUser, userPermissions: UserPermissions) {
    token.value = authToken
    user.value = authUser
    permissions.value = userPermissions
  }

  function clearAuth() {
    token.value = null
    user.value = null
    permissions.value = { functions: [], contracts: [] }
  }

  async function login(credentials: LoginCredentials) {
    const { apiFetch } = useApi()
    isLoading.value = true
    error.value = null

    try {
      const response = await apiFetch<ApiResponse<LoginResponse>>('/auth/login', {
        method: 'POST',
        body: credentials,
      })

      const data = response.data
      setAuth(data.token, data.user, data.permissions)
      
      // Store token in localStorage for persistence (or use cookies for better security)
      if (import.meta.client) {
        localStorage.setItem('ims_token', data.token)
      }
      
      return true
    } catch (err: any) {
      error.value = err.data?.message || err.message || 'Login failed'
      return false
    } finally {
      isLoading.value = false
    }
  }

  async function logout() {
    clearAuth()
    if (import.meta.client) {
      localStorage.removeItem('ims_token')
    }
  }

  async function fetchMe() {
    if (!token.value) return false

    const { apiFetch } = useApi()
    isLoading.value = true

    try {
      const response = await apiFetch<ApiResponse<AuthUser>>('/auth/me')
      user.value = response.data
      return true
    } catch (err) {
      // Token might be invalid/expired
      logout()
      return false
    } finally {
      isLoading.value = false
    }
  }

  function hasPermission(code: string): boolean {
    return permissions.value.functions.includes(code)
  }

  function hasContract(contractId: string): boolean {
    return permissions.value.contracts.includes(contractId)
  }

  return {
    token,
    user,
    permissions,
    isLoading,
    error,
    isAuthenticated,
    setAuth,
    clearAuth,
    login,
    logout,
    fetchMe,
    hasPermission,
    hasContract,
  }
})

// DTOs for Auth Actions
export interface LoginCredentials {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  expires_at: string
  user: AuthUser
  permissions: UserPermissions
}

export interface AuthUser {
  id: string
  username: string
  displayName: string
  groups: string[]
}

export interface UserPermissions {
  functions: string[]
  contracts: string[]
}
