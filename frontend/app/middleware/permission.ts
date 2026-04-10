import { resolveRequiredPermission } from "../shared/routing/routeAccess";

/**
 * Permission middleware — checks if the user has the required function permission.
 * Uses the `permission` meta field from definePageMeta().
 *
 * NOTE: This is frontend UX only. Real permission enforcement is server-side.
 */
export default defineNuxtRouteMiddleware((to) => {
  const requiredPermission = resolveRequiredPermission(to.meta);

  if (!requiredPermission) {
    return;
  }

  const authStore = useAuthStore();
  if (!authStore.hasPermission(requiredPermission)) {
    return navigateTo("/403");
  }
});
