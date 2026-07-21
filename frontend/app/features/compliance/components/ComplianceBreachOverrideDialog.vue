<script setup lang="ts">
/**
 * Explicit, audited override action for one OPEN breach. Submits only
 * `{ reason }` — the actor is derived from the JWT on the backend and is
 * never sent by the client. No invented minimum reason length: the backend
 * contract (`OverrideRequest`) only requires a non-empty reason.
 */
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";

import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceBreach } from "../types";
import ComplianceSeverityBadge from "./ComplianceSeverityBadge.vue";
import ComplianceVerdictBadge from "./ComplianceVerdictBadge.vue";

interface Props {
  breach: ComplianceBreach | null;
  submitting: boolean;
  error: string | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  cancel: [];
  submit: [payload: { reason: string }];
}>();

const { t } = useI18n();

const form = reactive({ reason: "" });
const touched = ref(false);
const dialogRef = ref<HTMLElement | null>(null);
const reasonRef = ref<HTMLTextAreaElement | null>(null);
let previouslyFocused: HTMLElement | null = null;

watch(
  () => props.breach,
  (next) => {
    if (next) {
      form.reason = "";
      touched.value = false;
    }
  },
);

const reasonError = computed(() =>
  form.reason.trim() ? null : t("compliance.postTrade.override.reasonRequired"),
);

const canSubmit = computed(() => !reasonError.value && !props.submitting);

function submit() {
  touched.value = true;
  if (!canSubmit.value) return;
  emit("submit", { reason: form.reason.trim() });
}

function focusableElements(): HTMLElement[] {
  if (!dialogRef.value) return [];
  return Array.from(
    dialogRef.value.querySelectorAll<HTMLElement>(
      'button:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  );
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    if (!props.submitting) {
      event.preventDefault();
      emit("cancel");
    }
    return;
  }

  if (event.key !== "Tab") return;

  const items = focusableElements();
  if (items.length === 0) return;

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
  () => props.breach,
  (breach) => {
    if (!import.meta.client) return;
    if (breach) {
      previouslyFocused = document.activeElement as HTMLElement | null;
      void nextTick(() => reasonRef.value?.focus());
    } else if (previouslyFocused && document.body.contains(previouslyFocused)) {
      previouslyFocused.focus();
      previouslyFocused = null;
    }
  },
);

onBeforeUnmount(() => {
  previouslyFocused = null;
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="breach"
      class="override-modal__backdrop"
      role="presentation"
      @keydown="handleKeydown"
    >
      <div
        ref="dialogRef"
        class="override-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="override-modal-title"
        tabindex="-1"
      >
        <header class="override-modal__head">
          <h2 id="override-modal-title" class="override-modal__title">
            {{ t("compliance.postTrade.override.title") }}
          </h2>
          <p class="override-modal__description">
            {{ t("compliance.postTrade.override.description") }}
          </p>
        </header>

        <section class="override-modal__breach">
          <span class="override-modal__breach-label">{{ t("compliance.postTrade.override.breachContext") }}</span>
          <div class="override-modal__rule">{{ ruleLabel(breach.ruleTypeID, t) }}</div>
          <div class="override-modal__badges">
            <ComplianceVerdictBadge :verdict="breach.verdict" />
            <ComplianceSeverityBadge :severity="breach.severity" />
          </div>
          <p class="override-modal__message">{{ breach.message || "—" }}</p>
        </section>

        <form class="override-modal__form" @submit.prevent="submit">
          <label class="override-modal__field">
            <span class="override-modal__label">
              {{ t("compliance.postTrade.override.reason") }} *
            </span>
            <textarea
              ref="reasonRef"
              v-model="form.reason"
              class="override-modal__textarea"
              rows="4"
              required
              :aria-invalid="touched && !!reasonError"
              aria-describedby="override-reason-help override-reason-error"
              :disabled="submitting"
              @blur="touched = true"
            />
            <span id="override-reason-help" class="override-modal__hint">
              {{ t("compliance.postTrade.override.reasonHelp") }}
            </span>
            <span
              v-if="touched && reasonError"
              id="override-reason-error"
              class="override-modal__error"
              role="alert"
            >
              {{ reasonError }}
            </span>
          </label>

          <div
            v-if="error"
            class="override-modal__error override-modal__error--alert"
            role="alert"
          >
            {{ error }}
          </div>

          <footer class="override-modal__footer">
            <AppButton variant="ghost" size="sm" :disabled="submitting" @click="emit('cancel')">
              {{ t("compliance.postTrade.override.cancel") }}
            </AppButton>
            <AppButton
              variant="primary"
              size="sm"
              type="submit"
              :loading="submitting"
              :disabled="!canSubmit"
            >
              {{
                submitting
                  ? t("compliance.postTrade.override.submitting")
                  : t("compliance.postTrade.override.submit")
              }}
            </AppButton>
          </footer>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.override-modal__backdrop {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay);
  display: grid;
  place-items: center;
  z-index: 950;
  padding: var(--space-5);
}

.override-modal {
  width: min(560px, 100%);
  max-height: 90vh;
  overflow: auto;
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: var(--space-7);
  display: grid;
  gap: var(--space-5);
}

.override-modal:focus {
  outline: none;
}

.override-modal__title {
  margin: 0 0 var(--space-2);
  font-size: var(--font-size-xl);
}

.override-modal__description {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.override-modal__breach {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.override-modal__breach-label {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.override-modal__rule {
  font-weight: var(--font-weight-semibold);
}

.override-modal__badges {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-xs);
}

.override-modal__message {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.override-modal__form {
  display: grid;
  gap: var(--space-4);
}

.override-modal__field {
  display: grid;
  gap: var(--space-2);
}

.override-modal__label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.override-modal__textarea {
  width: 100%;
  padding: var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  font-family: inherit;
  font-size: var(--font-size-sm);
}

.override-modal__textarea:focus-visible {
  outline: 2px solid var(--border-focus);
  outline-offset: 1px;
}

.override-modal__textarea:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.override-modal__hint {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.override-modal__error {
  font-size: var(--font-size-xs);
  color: var(--state-danger);
}

.override-modal__error--alert {
  padding: var(--space-3) var(--space-4);
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
}

.override-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
}
</style>
