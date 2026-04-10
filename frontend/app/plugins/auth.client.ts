import { isPublicRouteMeta } from "../shared/routing/routeAccess";

export default defineNuxtPlugin(async () => {
  const authStore = useAuthStore();
  const route = useRoute();

  if (!authStore.isAuthenticated || authStore.user) {
    return;
  }

  const restored = await authStore.restoreSession();

  if (restored) {
    return;
  }

  const publicRoutes = [
    "/auth/login",
    "/auth/register",
    "/auth/forgot-password",
    "/auth/mfa-setup",
  ];

  const isPublicRoute =
    isPublicRouteMeta(route.meta)
    || publicRoutes.some((candidate) => route.path.startsWith(candidate));

  if (!isPublicRoute) {
    await navigateTo(
      `/auth/login?reason=session_restore_failed&redirect=${encodeURIComponent(route.fullPath)}`,
    );
  }
});
