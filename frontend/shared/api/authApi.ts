import type { LoginRequest, LoginResponse, AuthUser } from '../types/auth.types'

const API_BASE = '/api/v1'

export const authApi = {
  /**
   * Login with username and password.
   * Returns JWT token, user profile, and permissions.
   */
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    try {
      // Login doesn't require auth token
      const response = await $fetch<LoginResponse>(`${API_BASE}/auth/login`, {
        method: 'POST',
        body: credentials,
        headers: {
          'Content-Type': 'application/json',
        },
      })
      return response
    } catch (error: any) {
      // Handle specific error responses from backend
      if (error.data?.error) {
        throw new Error(error.data.error)
      }
      if (error.status === 401) {
        throw new Error('Invalid username or password')
      }
      if (error.status === 403) {
        throw new Error('Your account is currently restricted from login')
      }
      if (error.status === 400) {
        throw new Error(error.data?.details?.[0]?.message || 'Invalid request')
      }
      throw new Error('Login failed. Please try again.')
    }
  },

  /**
   * Get current authenticated user profile.
   * Requires valid JWT token in Authorization header.
   */
  async getMe(token: string): Promise<AuthUser> {
    try {
      const response = await $fetch<AuthUser>(`${API_BASE}/auth/me`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })
      return response
    } catch (error: any) {
      if (error.status === 401) {
        throw new Error('Session expired. Please login again.')
      }
      throw new Error('Failed to fetch user profile')
    }
  },

  /**
   * Refresh JWT token.
   * Returns new token with extended expiration.
   */
  async refreshToken(token: string): Promise<{ token: string; expires_at: string }> {
    try {
      const response = await $fetch<{ token: string; expires_at: string }>(
        `${API_BASE}/auth/refresh`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      )
      return response
    } catch (error: any) {
      if (error.status === 401) {
        throw new Error('Session expired')
      }
      throw new Error('Failed to refresh token')
    }
  },

  /**
   * Logout — clear local session and token.
   * (Backend logout is optional — we primarily manage tokens client-side)
   */
  logout(): void {
    // Clear stored token and auth data
    if (process.client) {
      localStorage.removeItem('ims_auth_token')
      localStorage.removeItem('ims_auth_expires_at')
      localStorage.removeItem('ims_auth_user')
    }
  },
}
