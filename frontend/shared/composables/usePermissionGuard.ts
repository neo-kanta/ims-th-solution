/**
 * Permission guard composable — provides permission checking helpers.
 * NOTE: These are for frontend UX only. Real enforcement is server-side.
 */
import { useAuthStore } from '../stores/useAuthStore'

export function usePermissionGuard() {
  const authStore = useAuthStore()

  function hasFunction(code: string): boolean {
    return authStore.hasPermission(code)
  }

  function hasContract(contractId: string): boolean {
    return authStore.hasContract(contractId)
  }

  return { hasFunction, hasContract }
}
