<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";

interface Props {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  cancelLabel?: string;
  tone?: "danger" | "warning" | "neutral";
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
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-backdrop settings-confirm"
      role="presentation"
      @click.self="handleBackdropClick"
      @keydown="handleKeydown"
    >
      <section
        ref="dialogRef"
        class="modal settings-confirm__dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="settings-confirm-title"
        aria-describedby="settings-confirm-description"
        tabindex="-1"
      >
        <div class="modal-header settings-confirm__header">
          <div
            class="settings-confirm__icon"
            :class="`settings-confirm__icon--${tone}`"
            aria-hidden="true"
          >
            <AppIcon :name="tone === 'danger' ? 'warning' : 'shield'" />
          </div>
          <div>
            <h2 id="settings-confirm-title" class="settings-confirm__title">
              {{ title }}
            </h2>
            <p
              id="settings-confirm-description"
              class="settings-confirm__description"
            >
              {{ description }}
            </p>
          </div>
        </div>

        <div class="modal-footer settings-confirm__footer">
          <AppButton
            variant="secondary"
            size="sm"
            :disabled="loading"
            @click="emit('cancel')"
          >
            {{ cancelLabel }}
          </AppButton>
          <AppButton
            :variant="tone === 'danger' ? 'danger' : 'primary'"
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
.settings-confirm { align-items: center; }
.settings-confirm__dialog { width: min(34rem, 100%); border-radius: var(--radius-md); }
.settings-confirm__dialog:focus { outline: none; }
.settings-confirm__header { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: var(--space-4); align-items: flex-start; }
.settings-confirm__icon { display: inline-flex; align-items: center; justify-content: center; width: 2.5rem; height: 2.5rem; border-radius: var(--radius-md); background: var(--status-executed-bg); color: var(--status-executed-text); }
.settings-confirm__icon--warning { background: var(--status-pending-bg); color: var(--status-pending-text); }
.settings-confirm__icon--danger { background: var(--status-rejected-bg); color: var(--status-rejected-text); }
.settings-confirm__title { margin: 0; color: var(--text-primary); font-size: var(--font-size-lg); font-weight: var(--font-weight-semibold); }
.settings-confirm__description { margin: var(--space-2) 0 0; color: var(--text-secondary); line-height: var(--line-height-relaxed); }
.settings-confirm__footer { background: var(--bg-card-muted); }
@media (max-width: 640px) { .settings-confirm__header { grid-template-columns: 1fr; } }
</style>
