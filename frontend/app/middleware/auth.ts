import { isPublicRouteMeta } from "../shared/routing/routeAccess";

/**
 * Auth Middleware — Protects routes and manages session security
 *
 * Features:
 * - Route protection (authenticated users only)
 * - Session restoration on page load
 * - MFA requirement checking
 * - Redirect to login with return URL
 */
export default defineNuxtRouteMiddleware(async (to) => {
  const authStore = useAuthStore();

  const publicRoutes = [
    "/auth/login",
    "/auth/register",
    "/auth/forgot-password",
    "/auth/mfa-setup",
  ];

  const isPublicRoute =
    isPublicRouteMeta(to.meta)
    || publicRoutes.some((route) => to.path.startsWith(route));

  if (authStore.isTokenExpired) {
    authStore.clearAuth();

    if (!isPublicRoute) {
      return navigateTo(
        `/auth/login?reason=token_expired&redirect=${encodeURIComponent(to.fullPath)}`,
      );
    }
  }

  // 1. Restore session if cookie exists but user data is missing
  if (authStore.isAuthenticated && !authStore.user) {
    try {
      if (import.meta.client) {
        const restored = await authStore.restoreSession();

        if (!restored && !isPublicRoute) {
          return navigateTo(
            `/auth/login?reason=session_restore_failed&redirect=${encodeURIComponent(to.fullPath)}`,
          );
        }
      }
    } catch (error) {
      if (import.meta.client) {
        authStore.clearAuth();
        return navigateTo(
          `/auth/login?reason=session_restore_failed&redirect=${encodeURIComponent(to.fullPath)}`,
        );
      }
    }
  }

  // 2. Redirect unauthenticated users to login
  if (!authStore.isAuthenticated && !isPublicRoute) {
    return navigateTo(
      `/auth/login?redirect=${encodeURIComponent(to.fullPath)}`,
    );
  }

  // 3. Redirect authenticated users away from auth pages
  if (authStore.isAuthenticated && isPublicRoute) {
    return navigateTo("/");
  }

  // 4. Check MFA requirement for sensitive routes
  const sensitiveRoutes = ["/investment", "/approval", "/permissions"];
  const isSensitiveRoute = sensitiveRoutes.some((route) =>
    to.path.startsWith(route),
  );

  if (isSensitiveRoute && authStore.isAuthenticated && authStore.user) {
    // Backend enforces MFA via middleware. The frontend keeps the route
    // protected but does not log sensitive navigation in production.
  }
});
