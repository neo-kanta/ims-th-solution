import { ref, onBeforeUnmount } from "vue";

export type ToastTone = "success" | "danger" | "info";

export interface ToastState {
  id: number;
  tone: ToastTone;
  message: string;
}

const AUTO_DISMISS_MS: Record<ToastTone, number | null> = {
  success: 5000,
  info: 5000,
  // Danger persists until manually dismissed.
  danger: null,
};

/**
 * Bounded toast queue for the Settings console. Success/info toasts auto-dismiss
 * after 5s; danger toasts persist until the user closes them. Avoids the
 * single-slot "last write wins" problem where a quick error after success
 * overwrites the success message before the user sees it.
 */
export function useSettingsToasts() {
  const toasts = ref<ToastState[]>([]);
  const timers = new Map<number, ReturnType<typeof setTimeout>>();
  let nextId = 1;

  function dismiss(id: number) {
    toasts.value = toasts.value.filter((toast) => toast.id !== id);
    const timer = timers.get(id);
    if (timer) {
      clearTimeout(timer);
      timers.delete(id);
    }
  }

  function show(message: string, tone: ToastTone = "success") {
    const id = nextId++;
    toasts.value = [...toasts.value, { id, tone, message }];

    const dismissAfter = AUTO_DISMISS_MS[tone];
    if (dismissAfter !== null) {
      timers.set(
        id,
        setTimeout(() => dismiss(id), dismissAfter),
      );
    }
  }

  function clear() {
    for (const timer of timers.values()) clearTimeout(timer);
    timers.clear();
    toasts.value = [];
  }

  onBeforeUnmount(() => clear());

  return { toasts, show, dismiss, clear };
}
