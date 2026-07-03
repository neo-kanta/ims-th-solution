<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";
import type { WorkflowDialogTone } from "../types";

const props = withDefaults(
  defineProps<{
    open: boolean;
    title: string;
    description: string;
    confirmLabel: string;
    cancelLabel?: string;
    executing: boolean;
    tone: WorkflowDialogTone;
    disabledConfirm?: boolean;
  }>(),
  {
    cancelLabel: "Cancel",
    disabledConfirm: false,
  },
);

const emit = defineEmits<{
  cancel: [];
  confirm: [];
}>();

const { t } = useI18n();
const dialogRef = ref<HTMLElement | null>(null);

const mappedTone = computed(() => {
  if (props.tone === "danger") return "danger";
  if (props.tone === "warning") return "warning";
  return "neutral";
});

const iconName = computed(() => {
  if (props.tone === "danger" || props.tone === "warning") return "warning";
  return "info";
});

const confirmVariant = computed(() => {
  if (props.tone === "danger") return "danger" as const;
  if (props.tone === "warning") return "warning" as const;
  return "primary" as const;
});

function handleBackdropClick() {
  if (props.tone !== "danger" && !props.executing) {
    emit("cancel");
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape" && !props.executing) {
    emit("cancel");
  }
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      void nextTick(() => {
        dialogRef.value?.focus();
      });
    }
  },
);
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
        aria-labelledby="wf-confirm-title"
        aria-describedby="wf-confirm-description"
        tabindex="-1"
      >
        <!-- Header -->
        <div class="modal-header app-confirm__header">
          <div
            class="app-confirm__icon"
            :class="`app-confirm__icon--${mappedTone}`"
            aria-hidden="true"
          >
            <AppIcon :name="iconName" />
          </div>
          <div class="app-confirm__copy">
            <h2 id="wf-confirm-title" class="app-confirm__title">
              {{ title }}
            </h2>
            <p id="wf-confirm-description" class="app-confirm__description">
              {{ description }}
            </p>
          </div>
        </div>

        <!-- Body / Slot for Inputs -->
        <div class="app-confirm__body">
          <slot name="inputs" />
        </div>

        <!-- Footer -->
        <div class="modal-footer app-confirm__footer">
          <AppButton
            variant="secondary"
            size="sm"
            :disabled="executing"
            @click="emit('cancel')"
          >
            {{ cancelLabel }}
          </AppButton>
          <AppButton
            :variant="confirmVariant"
            size="sm"
            :loading="executing"
            :disabled="disabledConfirm || executing"
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
  display: flex;
  align-items: center;
  justify-content: center;
  position: fixed;
  inset: 0;
  z-index: 100;
}

.app-confirm__dialog {
  width: min(34rem, 100%);
  border-radius: var(--radius-md);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.app-confirm__dialog:focus {
  outline: none;
}

.app-confirm__header {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-4);
  align-items: flex-start;
  padding: var(--space-5) var(--space-5) var(--space-2) var(--space-5);
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
  color: var(--text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.app-confirm__description {
  margin: var(--space-2) 0 0;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.5;
}

.app-confirm__body {
  padding: 0 var(--space-5) var(--space-5) var(--space-5);
}

.app-confirm__footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  background: var(--bg-card-muted, #f9fafb);
  border-top: 1px solid var(--border-subtle);
}

@media (max-width: 640px) {
  .app-confirm__header {
    grid-template-columns: 1fr;
  }
}
</style>
