import { ref } from "vue";

import type {
  AdminSession,
  AdminUser,
  AdminUserStatusAction,
} from "../admin.types";
import type { PersonalAccountSession } from "../account.types";

export type ConfirmAction =
  | {
      type: "status";
      action: AdminUserStatusAction;
      user: AdminUser;
    }
  | {
      type: "reset-password";
      user: AdminUser;
    }
  | {
      type: "revoke-session";
      user: AdminUser;
      session: AdminSession;
    }
  | {
      type: "revoke-own-session";
      session: PersonalAccountSession;
    }
  | {
      type: "export-audit";
    };

export interface ConfirmHandlers {
  onStatus: (user: AdminUser, action: AdminUserStatusAction) => Promise<void>;
  onResetPassword: (user: AdminUser, newPassword: string) => Promise<void>;
  onRevokeSession: (session: AdminSession) => Promise<void>;
  onRevokeOwnSession: (session: PersonalAccountSession) => Promise<void>;
  onExportAudit: () => Promise<void>;
}

/**
 * Owns the confirm-dialog state machine for the Settings console:
 * - which destructive action is pending
 * - whether it is currently executing
 * - the cleartext password for reset-password (held outside reactive state
 *   so it never enters Vue devtools / heap snapshots)
 *
 * The parent supplies execution handlers via `confirm(handlers)` so the
 * composable stays decoupled from API-call wiring.
 */
export function useSettingsConfirm() {
  const confirmAction = ref<ConfirmAction | null>(null);
  const confirmLoading = ref(false);

  // Held outside reactive state so cleartext never enters Vue reactivity.
  let pendingResetPassword: string | null = null;

  function openStatusChange(action: AdminUserStatusAction, user: AdminUser) {
    confirmAction.value = { type: "status", action, user };
  }

  function openPasswordReset(user: AdminUser, newPassword: string) {
    pendingResetPassword = newPassword;
    confirmAction.value = { type: "reset-password", user };
  }

  function openSessionRevoke(user: AdminUser, session: AdminSession) {
    confirmAction.value = { type: "revoke-session", user, session };
  }

  function openOwnSessionRevoke(session: PersonalAccountSession) {
    confirmAction.value = { type: "revoke-own-session", session };
  }

  function openAuditExport() {
    confirmAction.value = { type: "export-audit" };
  }

  function cancelConfirm() {
    if (confirmLoading.value) return;
    pendingResetPassword = null;
    confirmAction.value = null;
  }

  async function confirm(handlers: ConfirmHandlers): Promise<void> {
    const action = confirmAction.value;
    if (!action || confirmLoading.value) return;

    confirmLoading.value = true;

    try {
      switch (action.type) {
        case "status":
          await handlers.onStatus(action.user, action.action);
          break;
        case "reset-password": {
          const password = pendingResetPassword;
          if (password) {
            await handlers.onResetPassword(action.user, password);
          }
          break;
        }
        case "revoke-session":
          await handlers.onRevokeSession(action.session);
          break;
        case "revoke-own-session":
          await handlers.onRevokeOwnSession(action.session);
          break;
        case "export-audit":
          await handlers.onExportAudit();
          break;
      }
      confirmAction.value = null;
    } finally {
      pendingResetPassword = null;
      confirmLoading.value = false;
    }
  }

  return {
    confirmAction,
    confirmLoading,
    openStatusChange,
    openPasswordReset,
    openSessionRevoke,
    openOwnSessionRevoke,
    openAuditExport,
    cancelConfirm,
    confirm,
  };
}
