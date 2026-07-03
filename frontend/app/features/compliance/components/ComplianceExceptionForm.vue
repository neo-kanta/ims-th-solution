<script setup lang="ts">
import { reactive } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";

const { t } = useI18n();

interface FormState {
  failedRule: string;
  orderRef: string;
  portfolioId: string;
  contractId: string;
  reason: string;
  justification: string;
  expiry: string;
  approver: string;
  ack: boolean;
}

const form = reactive<FormState>({
  failedRule: "",
  orderRef: "",
  portfolioId: "",
  contractId: "",
  reason: "",
  justification: "",
  expiry: "",
  approver: "",
  ack: false,
});
</script>

<template>
  <AppCard
    :title="t('compliance.exceptions.form.title')"
  >
    <div class="exception-form__banner" role="status">
      <strong>API not available.</strong>
      {{ t("compliance.exceptions.form.submitDisabled") }}
    </div>

    <form class="exception-form" @submit.prevent>
      <div class="exception-form__grid">
        <label class="exception-form__field">
          <span class="exception-form__label">
            {{ t("compliance.exceptions.form.failedRule") }}
          </span>
          <input
            v-model="form.failedRule"
            type="text"
            class="form-control"
            placeholder="e.g. concentration.single_issuer"
          />
        </label>

        <label class="exception-form__field">
          <span class="exception-form__label">
            {{ t("compliance.exceptions.form.orderRef") }}
          </span>
          <input v-model="form.orderRef" type="text" class="form-control" />
        </label>

        <label class="exception-form__field">
          <span class="exception-form__label">
            {{ t("compliance.exceptions.form.portfolioId") }}
          </span>
          <input v-model="form.portfolioId" type="text" class="form-control" />
        </label>

        <label class="exception-form__field">
          <span class="exception-form__label">
            {{ t("compliance.exceptions.form.contractId") }}
          </span>
          <input v-model="form.contractId" type="text" class="form-control" />
        </label>

        <label class="exception-form__field exception-form__field--wide">
          <span class="exception-form__label">
            {{ t("compliance.exceptions.form.reason") }}
          </span>
          <input v-model="form.reason" type="text" class="form-control" />
        </label>

        <label class="exception-form__field exception-form__field--wide">
          <span class="exception-form__label">
            {{ t("compliance.exceptions.form.justification") }}
          </span>
          <textarea
            v-model="form.justification"
            rows="3"
            class="form-control"
          />
        </label>

        <label class="exception-form__field">
          <span class="exception-form__label">
            {{ t("compliance.exceptions.form.expiry") }}
          </span>
          <input v-model="form.expiry" type="date" class="form-control" />
        </label>

        <label class="exception-form__field">
          <span class="exception-form__label">
            {{ t("compliance.exceptions.form.approver") }}
          </span>
          <input v-model="form.approver" type="text" class="form-control" />
        </label>
      </div>

      <label class="exception-form__check">
        <input v-model="form.ack" type="checkbox" />
        {{ t("compliance.exceptions.form.ack") }}
      </label>

      <p class="exception-form__attach">
        {{ t("compliance.exceptions.form.attachmentNote") }}
      </p>

      <div class="exception-form__footer">
        <AppButton
          variant="primary"
          size="sm"
          :disabled="true"
          :title="t('compliance.exceptions.form.submitDisabled')"
        >
          Submit exception (disabled)
        </AppButton>
      </div>
    </form>
  </AppCard>
</template>

<style scoped>
.exception-form__banner {
  margin-bottom: var(--space-5);
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--alert-warning-border);
  border-radius: var(--radius-md);
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  font-size: var(--font-size-sm);
}

.exception-form {
  display: grid;
  gap: var(--space-5);
}

.exception-form__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--space-4);
}

.exception-form__field {
  display: grid;
  gap: var(--space-2);
  margin: 0;
}

.exception-form__field--wide {
  grid-column: 1 / -1;
}

.exception-form__label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.form-control {
  width: 100%;
  min-height: var(--size-control-md);
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-family: inherit;
}

.form-control:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.exception-form__check {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.exception-form__attach {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  font-style: italic;
}

.exception-form__footer {
  display: flex;
  justify-content: flex-end;
}
</style>
