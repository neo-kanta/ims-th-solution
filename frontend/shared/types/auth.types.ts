/**
 * Authentication Types — shared across the application
 */

export interface LoginRequest {
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
  display_name: string
  email?: string
  groups?: string[]
  is_on_leave: boolean
}

export interface UserPermissions {
  functions: string[] // Function permission codes (e.g., 'WORKFLOW_VIEW')
  contracts: string[] // Contract IDs the user has data access to
}

export interface AuthState {
  token: string | null
  expiresAt: string | null
  user: AuthUser | null
  permissions: UserPermissions | null
  isLoading: boolean
  error: string | null
  isAuthenticated: boolean
}

export interface TokenPayload {
  user_id: string
  username: string
  exp: number // Unix timestamp
}

export const DEFAULT_AUTH_STATE: AuthState = {
  token: null,
  expiresAt: null,
  user: null,
  permissions: null,
  isLoading: false,
  error: null,
  isAuthenticated: false,
}
