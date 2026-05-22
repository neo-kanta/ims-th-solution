<script setup lang="ts">
/**
 * AppConfirmDialog
 * ------------------------------------------------------------------
 * Accessible, reusable confirmation dialog primitive.
 *
 * Designed for destructive or state-changing actions where the operator
 * needs to consciously confirm before the system commits (delete, submit,
 * cancel-submission, force-post, etc.).
 *
 * Accessibility notes:
 *   - Uses role="dialog" + aria-modal + aria-labelledby/aria-describedby.
 *   - Focus is trapped within the dialog while open (Tab / Shift+Tab).
 *   - Escape dismisses unless `loading` (so a fired request cannot be
 *     interrupted mid-flight).
 *   - Backdrop click dismisses for non-danger tones only — destructive
 *     actions require explicit Cancel to prevent accidental loss of
 *     intent recovery (matches the SettingsConfirmDialog convention).
 *   - The previously-focused element is restored when the dialog closes.
 *
 * Visual notes:
 *   - The dialog uses the global `.modal`, `.modal-backdrop`,
 *     `.modal-header`, `.modal-footer` design tokens already defined in
 *     `app/assets/css`. No new global styles are introduced.
 */
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";

import AppButton from "./AppButton.vue";
import AppIcon from "./AppIcon.vue";

type Tone = "danger" | "warning" | "neutral";

interface Props {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  cancelLabel?: string;
  tone?: Tone;
  /**
   * When true the confirm button shows a spinner and Escape / backdrop
   * dismiss are suppressed. Callers should keep this in sync with their
   * mutation lifecycle so the operator cannot double-click submit.
   */
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  cancelLabel: "Cancel",
  tone: "neutral",
  loading: false,
});

const emit = defineEmits<{
  cancel: [];
  confirm: [];
}>();

const dialogRef = ref<HTMLElement | null>(null);
let previouslyFocused: HTMLElement | null = null;

const allowBackdropDismiss = computed(() => props.tone !== "danger");

const iconName = computed(() => {
  if (props.tone === "danger") return "warning";
  if (props.tone === "warning") return "warning";
  return "info";
});

function focusableElements(): HTMLElement[] {
  if (!dialogRef.value) return [];
  return Array.from(
    dialogRef.value.querySelectorAll<HTMLElement>(
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  );
}

function focusFirst() {
  const items = focusableElements();
  // Prefer the cancel button when present so a destructive default is not
  // pre-armed for accidental Enter-keypress. Cancel is the first button in
  // the footer, so items[0] already lands there when the dialog opens.
  items[0]?.focus();
}

function handleBackdropClick() {
  if (allowBackdropDismiss.value && !props.loading) {
    emit("cancel");
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    if (!props.loading) {
      event.preventDefault();
      emit("cancel");
    }
    return;
  }

  if (event.key !== "Tab") return;

  const items = focusableElements();
  if (items.length === 0) {
    event.preventDefault();
    dialogRef.value?.focus();
    return;
  }

  const first = items[0];
  const last = items[items.length - 1];
  if (!first || !last) return;

  const active = document.activeElement as HTMLElement | null;

  if (event.shiftKey && (active === first || !dialogRef.value?.contains(active))) {
    event.preventDefault();
    last.focus();
  } else if (!event.shiftKey && active === last) {
    event.preventDefault();
    first.focus();
  }
}

watch(
  () => props.open,
  (open) => {
    if (!import.meta.client) return;
    if (open) {
      previouslyFocused = document.activeElement as HTMLElement | null;
      void nextTick(() => focusFirst());
    } else if (previouslyFocused && document.body.contains(previouslyFocused)) {
      previouslyFocused.focus();
      previouslyFocused = null;
    }
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  previouslyFocused = null;
});

const confirmVariant = computed(() => {
  if (props.tone === "danger") return "danger" as const;
  if (props.tone === "warning") return "warning" as const;
  return "primary" as const;
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-backdrop app-confirm"
      role="presentation"
      @click.self="handleBackdropClick"
      @keydown="handleKeydown"
    >
      <section
        ref="dialogRef"
        class="modal app-confirm__dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="app-confirm-title"
        aria-describedby="app-confirm-description"
        tabindex="-1"
      >
        <div class="modal-header app-confirm__header">
          <div
            class="app-confirm__icon"
            :class="`app-confirm__icon--${tone}`"
            aria-hidden="true"
          >
            <AppIcon :name="iconName" />
          </div>
          <div class="app-confirm__copy">
            <h2 id="app-confirm-title" class="app-confirm__title">
              {{ title }}
            </h2>
            <p id="app-confirm-description" class="app-confirm__description">
              {{ description }}
            </p>
          </div>
        </div>

        <div class="modal-footer app-confirm__footer">
          <AppButton
            variant="secondary"
            size="sm"
            :disabled="loading"
            @click="emit('cancel')"
          >
            {{ cancelLabel }}
          </AppButton>
          <AppButton
            :variant="confirmVariant"
            size="sm"
            :loading="loading"
            @click="emit('confirm')"
          >
            {{ confirmLabel }}
          </AppButton>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.app-confirm {
  align-items: center;
}

.app-confirm__dialog {
  width: min(34rem, 100%);
  border-radius: var(--radius-md);
}

.app-confirm__dialog:focus {
  outline: none;
}

.app-confirm__header {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-4);
  align-items: flex-start;
}

.app-confirm__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: var(--radius-md);
  background: var(--status-executed-bg, rgba(37, 99, 235, 0.12));
  color: var(--status-executed-text, #1d4ed8);
}

.app-confirm__icon--warning {
  background: var(--status-pending-bg, rgba(217, 119, 6, 0.12));
  color: var(--status-pending-text, #b45309);
}

.app-confirm__icon--danger {
  background: var(--status-rejected-bg, rgba(220, 38, 38, 0.12));
  color: var(--status-rejected-text, #b91c1c);
}

.app-confirm__copy {
  min-width: 0;
}

.app-confirm__title {
  margin: 0;
  color: var(--text-primary, #111827);
  font-size: var(--font-size-lg, 1.125rem);
  font-weight: var(--font-weight-semibold, 600);
}

.app-confirm__description {
  margin: var(--space-2) 0 0;
  color: var(--text-secondary, #4b5563);
  line-height: var(--line-height-relaxed, 1.6);
  white-space: pre-line;
}

.app-confirm__footer {
  background: var(--bg-card-muted, #f9fafb);
}

@media (max-width: 640px) {
  .app-confirm__header {
    grid-template-columns: 1fr;
  }
}
</style>
