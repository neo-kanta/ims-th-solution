import { computed } from "vue";
import { useState } from "#imports";

export type ToastTone = "success" | "danger" | "warning" | "info";

export interface Toast {
  id: number;
  message: string;
  tone: ToastTone;
  duration?: number;
}

const AUTO_DISMISS_MS: Record<ToastTone, number | null> = {
  success: 5000,
  info: 5000,
  warning: 6000,
  // Danger persists until manually dismissed by default
  danger: null,
};

export function useAppToast() {
  const toasts = useState<Toast[]>("app-toasts-state", () => []);
  const activeTimers = useState<Record<number, boolean>>("app-toasts-timers", () => ({}));
  const nextId = useState<number>("app-toasts-next-id", () => 1);

  function dismiss(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id);
    delete activeTimers.value[id];
  }

  function show(message: string, tone: ToastTone = "success", duration?: number) {
    const id = nextId.value++;
    const resolvedDuration = duration !== undefined ? duration : AUTO_DISMISS_MS[tone];

    toasts.value.push({
      id,
      message,
      tone,
      duration: resolvedDuration ?? undefined,
    });

    if (resolvedDuration && resolvedDuration > 0 && import.meta.client) {
      activeTimers.value[id] = true;
      setTimeout(() => {
        if (activeTimers.value[id]) {
          dismiss(id);
        }
      }, resolvedDuration);
    }
    
    return id;
  }

  function clearAll() {
    toasts.value = [];
    activeTimers.value = {};
  }

  return {
    toasts: computed(() => toasts.value),
    show,
    showSuccess: (msg: string, dur?: number) => show(msg, "success", dur),
    showError: (msg: string, dur?: number) => show(msg, "danger", dur),
    showWarning: (msg: string, dur?: number) => show(msg, "warning", dur),
    showInfo: (msg: string, dur?: number) => show(msg, "info", dur),
    dismiss,
    clearAll,
  };
}
