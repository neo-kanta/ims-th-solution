# Authentication System — IMS Thailand Frontend

## Overview

The authentication system provides:
- **JWT-based login** with secure token storage
- **Session persistence** across browser refreshes via localStorage
- **Automatic token refresh** before expiration (5-minute buffer)
- **Global permission checking** (function + data scope access)
- **Route protection** via auth middleware

## Architecture

### Components

```
├── pages/auth/login.vue              # Login page (modern UI, full-featured)
├── shared/
│   ├── api/authApi.ts                # Auth API client (login, refresh, getMe)
│   ├── stores/useAuthStore.ts         # Global auth state (Pinia)
│   ├── types/auth.types.ts            # TypeScript interfaces
│   ├── composables/useFetch.ts        # Authenticated fetch wrapper
│   └── middleware/
│       └── auth.ts                    # Route protection middleware
└── plugins/
    └── auth.client.ts                 # Auth initialization on app startup
```

## Development Credentials

**For development/testing only. Remove before production.**

| Field | Value |
|-------|-------|
| **Username** | `guest` |
| **Password** | `guest123` |
| **Permissions** | Admin (full access to all features) |

## Login Flow

```mermaid
User Login
    ↓
[POST /api/v1/auth/login]
    ↓
Backend validates credentials & returns:
  - JWT token
  - Token expiration timestamp
  - User profile (ID, username, display_name, email, groups)
  - Permissions (function codes + contract access)
    ↓
Frontend stores in:
  - Memory (Pinia store)
  - localStorage (session persistence)
    ↓
Schedule automatic token refresh
    ↓
Redirect to dashboard
```

## Session Persistence

Tokens are stored in **localStorage** with keys:
- `ims_auth_token` — JWT token for API requests
- `ims_auth_expires_at` — Token expiration timestamp (ISO 8601)
- `ims_auth_user` — User profile JSON

On app refresh:
1. Nuxt auth plugin calls `authStore.restoreSession()`
2. Tokens are restored from localStorage
3. If token is expired, session is cleared
4. If token is valid, app is ready for authenticated requests

## Using the Auth Store

### In Components (Script Setup)

```typescript
<script setup lang="ts">
const authStore = useAuthStore()

// Check if user is authenticated
if (authStore.isAuthenticated) {
  console.log('User:', authStore.user?.username)
  console.log('Permissions:', authStore.permissions?.functions)
}

// Check specific permission
if (authStore.hasPermission('WORKFLOW_VIEW')) {
  // User can view workflow
}

// Check contract access
if (authStore.hasContractAccess('contract-123')) {
  // User has access to this contract
}

// Logout
authStore.logout()

// Error from previous login attempt
if (authStore.error) {
  console.log('Login error:', authStore.error)
  authStore.clearError()
}
</script>
```

### Login Example

```typescript
// In login page or component
const authStore = useAuthStore()

const success = await authStore.login({
  username: 'guest',
  password: 'guest123',
})

if (success) {
  // Login successful, user is authenticated
  navigateTo('/')
} else {
  // Login failed, check authStore.error for details
  console.log('Error:', authStore.error)
}
```

### Check Permission

```typescript
// In components or composables
const canApprove = authStore.hasPermission('APPROVAL_VIEW')
const canSeeContract = authStore.hasContractAccess('ABCFLEX1')

if (!canApprove) {
  // Show "not authorized" message
}
```

## Route Protection

### Protecting Routes

Routes are automatically protected by the `auth` middleware. Define it in `definePageMeta`:

```typescript
// pages/investment/decision/index.vue
<script setup lang="ts">
definePageMeta({
  layout: 'dashboard',
  middleware: ['auth', 'permission'],  // auth = require login, permission = check role
  meta: {
    permission: 'INVESTMENT_VIEW'      // Required permission code
  }
})
</script>
```

### Middleware Logic

- **`auth` middleware**:
  - Restores session from localStorage on first load
  - Redirects to `/auth/login` if not authenticated
  - Redirects authenticated users away from login page

- **`permission` middleware** (planned):
  - Checks if user has required permission code
  - Shows 403 if permission is denied

## Token Refresh

### Automatic Refresh

Tokens are automatically refreshed **5 minutes before expiration**:

```typescript
// In auth store
function scheduleTokenRefresh() {
  // Calculate time until token expires
  // Schedule refresh 5 minutes before expiration
  // When timer fires, call authApi.refreshToken()
}
```

### Manual Refresh

```typescript
const authStore = useAuthStore()

// Refresh token immediately
const success = await authStore.refreshToken()

if (!success) {
  // Refresh failed — session expired, user needs to login again
  navigateTo('/auth/login')
}
```

## Token Expiration Handling

### Detection

```typescript
const isExpired = authStore.isTokenExpired     // True if expired
const isValid = authStore.isAuthenticated      // False if expired or no token
```

### On API Failure (401)

When an API call returns **401 Unauthorized**:

```typescript
// In authApi.ts, refresh() catches 401 and:
// 1. Calls authStore.logout()
// 2. Clears localStorage
// 3. Sets error message
// 4. (Optional) Redirects to login
```

## Security Considerations

### ✅ What We Do

- **Tokens in localStorage**: Survives page refresh (good UX), vulnerable to XSS
- **HttpOnly cookies**: Not used (SSR complexity, but more secure)
- **HTTPS**: Required in production (enforced by backend)
- **CORS**: Backend restricts to frontend origin
- **Token expiration**: Short-lived tokens (e.g., 1-2 hours)
- **Refresh tokens**: Auto-refresh before expiration

### ⚠️ Known Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| **XSS attack reads token** | Use CSP headers, sanitize user inputs, audit dependencies |
| **CSRF** | Backend uses SameSite cookies + CSRF tokens in form submissions |
| **Token in URL** | Never log or pass token in URLs (only Authorization header) |
| **Intercepted token (HTTP)** | HTTPS required in production (redirect HTTP → HTTPS) |

### For Production

1. **Move token to HttpOnly cookie** (if SSR complexity is acceptable)
2. **Add CSRF token** to form submissions
3. **Implement Content Security Policy** (CSP) headers
4. **Add rate limiting** on login endpoint
5. **Monitor for suspicious patterns** (multiple failed logins, unusual IPs)
6. **Require HTTPS** everywhere
7. **Consider OAuth/OIDC** if integrating with external identity providers

## Login Page Features

The modern login page at `/auth/login` includes:

- ✅ **Form validation** (username ≥3 chars, password ≥6 chars)
- ✅ **Error display** (dismissible alert)
- ✅ **Loading state** (spinner during login)
- ✅ **Show/hide password** toggle
- ✅ **Keyboard shortcuts** (Enter to submit)
- ✅ **Auto-focus** on mount
- ✅ **Responsive design** (mobile-friendly)
- ✅ **Gradient background** (modern UI)
- ✅ **Development credentials display** (for testing)

## Troubleshooting

### "Invalid username or password" Error

1. Check credentials are correct
2. Verify user exists in database: `SELECT * FROM iam_users WHERE username = 'guest';`
3. Verify password hash: `SELECT * FROM iam_users WHERE username = 'guest' LIMIT 1;`

### "Session expired" on Protected Route

1. Token may have expired
2. Check token expiration: `authStore.expiresAt`
3. Manual refresh: `await authStore.refreshToken()`
4. Logout and re-login

### Token Not Stored in localStorage

1. Check browser localStorage: `DevTools > Application > Local Storage`
2. Verify keys: `ims_auth_token`, `ims_auth_expires_at`, `ims_auth_user`
3. Check browser privacy settings (may block localStorage)

### "Cannot find name 'ref'" TypeScript Errors

These are Nuxt auto-import false positives:
1. They resolve after `nuxi prepare` regenerates types
2. No code changes needed — the app works despite IDE errors
3. Run: `cd frontend && npx nuxi prepare`

## API Reference

### `POST /api/v1/auth/login`

```json
Request:
{
  "username": "guest",
  "password": "guest123"
}

Response (200 OK):
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-03-13T19:30:00Z",
  "user": {
    "id": "a0000000-0000-0000-0000-000000000002",
    "username": "guest",
    "display_name": "Guest User (Dev Only)",
    "email": "guest@ims.local",
    "groups": ["Admin"],
    "is_on_leave": false
  },
  "permissions": {
    "functions": ["WORKFLOW_VIEW", "INVESTMENT_VIEW", "APPROVAL_VIEW", ...],
    "contracts": ["ABCFLEX1", "ABCFLEX2", ...]
  }
}

Error (401 Unauthorized):
{
  "error": "invalid username or password"
}
```

### `GET /api/v1/auth/me`

```
Headers:
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

Response (200 OK):
{
  "id": "a0000000-0000-0000-0000-000000000002",
  "username": "guest",
  "display_name": "Guest User (Dev Only)",
  "email": "guest@ims.local",
  "groups": ["Admin"],
  "is_on_leave": false
}

Error (401 Unauthorized):
{
  "error": "not authenticated"
}
```

### `POST /api/v1/auth/refresh`

```
Headers:
Authorization: Bearer <old_token>

Response (200 OK):
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-03-13T20:30:00Z"
}

Error (401 Unauthorized):
{
  "error": "session expired"
}
```

## Testing

### E2E Test Example

```typescript
// tests/e2e/auth.spec.ts
import { test, expect } from '@playwright/test'

test('user can login and access dashboard', async ({ page }) => {
  // Navigate to login
  await page.goto('http://localhost:3000/auth/login')

  // Fill credentials
  await page.fill('#username', 'guest')
  await page.fill('#password', 'guest123')

  // Submit
  await page.click('button[type="submit"]')

  // Should redirect to dashboard
  await page.waitForURL('http://localhost:3000/')

  // Verify authenticated
  expect(page.url()).toBe('http://localhost:3000/')
})
```

### Unit Test Example

```typescript
// shared/stores/__tests__/useAuthStore.spec.ts
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../useAuthStore'

describe('useAuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('should restore session from localStorage', () => {
    const store = useAuthStore()

    localStorage.setItem('ims_auth_token', 'test-token')
    localStorage.setItem('ims_auth_expires_at', '2099-12-31T00:00:00Z')
    localStorage.setItem('ims_auth_user', JSON.stringify({
      id: '123',
      username: 'guest',
      display_name: 'Guest',
      is_on_leave: false
    }))

    store.restoreSession()

    expect(store.token).toBe('test-token')
    expect(store.isAuthenticated).toBe(true)
  })
})
```

## See Also

- [Backend Authentication](../backend/internal/iam/README.md)
- [Login Page Source](./app/pages/auth/login.vue)
- [Auth Store Source](./shared/stores/useAuthStore.ts)
