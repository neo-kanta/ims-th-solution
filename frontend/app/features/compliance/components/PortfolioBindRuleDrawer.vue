<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppDateField from "~/shared/ui/AppDateField.vue";
import AppFormField from "~/shared/ui/AppFormField.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";

import type { BindDraft } from "../../portfolio-workspace/lib/complianceBindingState";
import type { ApiPortfolioRuleCatalogEntry } from "../../portfolio-workspace/services/portfolioComplianceApi";
import { validateBindDraft } from "../lib/bindDraftValidation";
import { ruleExplanation, ruleLabel } from "../lib/ruleTypeCatalog";

interface Props {
  open: boolean;
  entry: ApiPortfolioRuleCatalogEntry | null;
  draft: BindDraft;
  submitting: boolean;
  error: string | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  "update:draft": [draft: BindDraft];
  submit: [];
}>();

const { t } = useI18n();

const SEVERITY_OPTIONS = ["BLOCK", "WARN", "REQUIRE_APPROVAL", "MONITOR"] as const;

const severityOptions = computed(() =>
  SEVERITY_OPTIONS.map((value) => ({
    value,
    label: t(`compliance.badges.severity.${value}`),
  })),
);

const severityGuide = computed(() => {
  switch (props.draft.severity) {
    case "BLOCK":
      return t("portfolio.compliance.severityGuide.BLOCK");
    case "WARN":
      return t("portfolio.compliance.severityGuide.WARN");
    case "REQUIRE_APPROVAL":
      return t("portfolio.compliance.severityGuide.REQUIRE_APPROVAL");
    case "MONITOR":
      return t("portfolio.compliance.severityGuide.MONITOR");
    default:
      return "";
  }
});

const validation = computed(() => validateBindDraft(props.draft));

function fieldErrorMessage(code: "REQUIRED" | "INVALID_DATE" | "BEFORE_FROM" | null): string {
  switch (code) {
    case "REQUIRED":
      return t("portfolio.compliance.validation.required");
    case "INVALID_DATE":
      return t("portfolio.compliance.validation.invalidDate");
    case "BEFORE_FROM":
      return t("portfolio.compliance.validation.beforeFrom");
    default:
      return "";
  }
}

function patch(field: keyof BindDraft, value: string) {
  emit("update:draft", { ...props.draft, [field]: value });
}

function handleSubmit() {
  if (!validation.value.isValid || props.submitting) return;
  emit("submit");
}

const dialogRef = ref<HTMLElement | null>(null);
let previouslyFocused: HTMLElement | null = null;

function focusableElements(): HTMLElement[] {
  if (!dialogRef.value) return [];
  return Array.from(
    dialogRef.value.querySelectorAll<HTMLElement>(
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  );
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    if (!props.submitting) {
      event.preventDefault();
      emit("close");
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
  () => props.open,
  (open) => {
    if (!import.meta.client) return;
    if (open) {
      previouslyFocused = document.activeElement as HTMLElement | null;
      void nextTick(() => focusableElements()[0]?.focus());
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
      v-if="open && entry"
      class="drawer-backdrop"
      role="presentation"
      @keydown="handleKeydown"
    >
      <section
        ref="dialogRef"
        class="drawer"
        role="dialog"
        aria-modal="true"
        aria-labelledby="bind-drawer-title"
        tabindex="-1"
      >
        <header class="drawer__header">
          <h2 id="bind-drawer-title" class="drawer__title">
            {{ t("portfolio.compliance.bindDrawer.title") }}
          </h2>
          <button
            type="button"
            class="drawer__close"
            :disabled="submitting"
            :aria-label="t('portfolio.compliance.bindDrawer.close')"
            @click="emit('close')"
          >
            ×
          </button>
        </header>

        <div class="drawer__body">
          <div class="drawer__rule">
            <div class="drawer__rule-label">{{ ruleLabel(entry.rule_type_id ?? "", t) }}</div>
            <p class="drawer__rule-explanation">
              {{ ruleExplanation(entry.rule_type_id ?? "", entry.description ?? "", t) }}
            </p>
            <code class="drawer__rule-type-id">{{ entry.rule_type_id }}</code>
          </div>

          <form class="drawer__form" @submit.prevent="handleSubmit">
            <AppFormField
              id="bind-drawer-severity"
              :label="t('portfolio.compliance.columns.severity')"
            >
              <AppSelect
                id="bind-drawer-severity"
                :model-value="draft.severity"
                :options="severityOptions"
                @update:model-value="(v) => patch('severity', String(v))"
              />
            </AppFormField>
            <p class="drawer__severity-guide">{{ severityGuide }}</p>

            <AppFormField
              id="bind-drawer-from"
              :label="t('portfolio.compliance.effectiveFrom')"
              :error="fieldErrorMessage(validation.effectiveFromError)"
              required
            >
              <AppDateField
                id="bind-drawer-from"
                :model-value="draft.effectiveFrom"
                :error="!!validation.effectiveFromError"
                @update:model-value="(v) => patch('effectiveFrom', v)"
              />
            </AppFormField>

            <AppFormField
              id="bind-drawer-to"
              :label="t('portfolio.compliance.effectiveTo')"
              :error="fieldErrorMessage(validation.effectiveToError)"
            >
              <AppDateField
                id="bind-drawer-to"
                :model-value="draft.effectiveTo"
                :error="!!validation.effectiveToError"
                @update:model-value="(v) => patch('effectiveTo', v)"
              />
            </AppFormField>

            <p v-if="error" class="drawer__error" role="alert">{{ error }}</p>
          </form>
        </div>

        <footer class="drawer__footer">
          <AppButton variant="secondary" size="sm" :disabled="submitting" @click="emit('close')">
            {{ t("portfolio.compliance.cancel") }}
          </AppButton>
          <AppButton
            variant="primary"
            size="sm"
            :loading="submitting"
            :disabled="!validation.isValid || submitting"
            @click="handleSubmit"
          >
            {{ t("portfolio.compliance.confirmBind") }}
          </AppButton>
        </footer>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.drawer-backdrop {
  position: fixed;
  inset: 0;
  background: var(--bg-overlay);
  display: flex;
  justify-content: flex-end;
  z-index: 1040;
}

.drawer {
  width: min(28rem, 100%);
  height: 100%;
  background: var(--bg-card);
  border-left: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
}

.drawer:focus {
  outline: none;
}

.drawer__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
}

.drawer__title {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.drawer__close {
  background: transparent;
  border: none;
  font-size: 1.5rem;
  line-height: 1;
  color: var(--text-secondary);
  cursor: pointer;
  padding: var(--space-1) var(--space-2);
}

.drawer__body {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-5);
  display: grid;
  gap: var(--space-5);
}

.drawer__rule-label {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.drawer__rule-explanation {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.drawer__rule-type-id {
  display: inline-block;
  margin-top: var(--space-2);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-2xs);
  color: var(--text-tertiary);
}

.drawer__form {
  display: grid;
  gap: var(--space-4);
}

.drawer__severity-guide {
  margin: calc(var(--space-2) * -1) 0 0;
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  background: var(--bg-card-muted);
  border-radius: var(--radius-md);
  padding: var(--space-3);
}

.drawer__error {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--state-danger);
}

.drawer__footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
  padding: var(--space-4) var(--space-5);
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-card-muted);
}
</style>
