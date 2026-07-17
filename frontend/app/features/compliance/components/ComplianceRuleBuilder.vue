<script setup lang="ts">
import { computed, reactive, ref } from "vue";

import { useI18n, type AppTranslationKey } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";

import { useComplianceRuleCreate } from "../composables/useComplianceRules";
import { RULE_PARAMETER_SAMPLES } from "../lib/ruleParameterSamples";
import { RULE_CATALOG, lookupRuleCatalog } from "../lib/ruleTypeCatalog";
import type { ComplianceCreateRuleInstanceRequest } from "../services/complianceApi";

const emit = defineEmits<{
  created: [id: string];
}>();

const { t } = useI18n();
const mutation = useComplianceRuleCreate();

const STEPS: readonly { key: "identity" | "scope" | "logic" | "message" | "review" | "submit"; labelKey: AppTranslationKey }[] = [
  { key: "identity", labelKey: "compliance.builder.steps.identity" },
  { key: "scope", labelKey: "compliance.builder.steps.scope" },
  { key: "logic", labelKey: "compliance.builder.steps.logic" },
  { key: "message", labelKey: "compliance.builder.steps.message" },
  { key: "review", labelKey: "compliance.builder.steps.review" },
  { key: "submit", labelKey: "compliance.builder.steps.submit" },
] as const;

type StepKey = (typeof STEPS)[number]["key"];

const stepIndex = ref(0);
const currentStep = computed<StepKey>(() => STEPS[stepIndex.value]!.key);

function go(direction: "next" | "prev") {
  if (direction === "next" && stepIndex.value < STEPS.length - 1) {
    stepIndex.value += 1;
  } else if (direction === "prev" && stepIndex.value > 0) {
    stepIndex.value -= 1;
  }
}

interface FormState {
  ruleTypeId: string;
  name: string;
  description: string;
  parametersJson: string;
  effectiveFrom: string;
  effectiveTo: string;
  isActive: boolean;
  changeReason: string;
}

const today = new Date().toISOString().slice(0, 10);

const form = reactive<FormState>({
  ruleTypeId: "",
  name: "",
  description: "",
  parametersJson: "{}",
  effectiveFrom: today,
  effectiveTo: "",
  isActive: true,
  changeReason: "initial creation",
});

function pickRuleType(typeId: string) {
  form.ruleTypeId = typeId;
  const sample = RULE_PARAMETER_SAMPLES[typeId];
  if (sample) form.parametersJson = sample;
  const entry = lookupRuleCatalog(typeId);
  if (entry && !form.name) form.name = t(entry.labelKey);
}

function onRuleTypeChange(event: Event) {
  if (event.target instanceof HTMLSelectElement) {
    pickRuleType(event.target.value);
  }
}

const jsonError = computed<string | null>(() => {
  try {
    const parsed = JSON.parse(form.parametersJson || "{}");
    if (parsed === null || typeof parsed !== "object" || Array.isArray(parsed)) {
      return t("compliance.builder.logic.invalidJson");
    }
    return null;
  } catch {
    return t("compliance.builder.logic.invalidJson");
  }
});

const identityValid = computed(
  () =>
    form.ruleTypeId.trim().length > 0 &&
    form.name.trim().length >= 3,
);
const scopeValid = computed(
  () =>
    /^\d{4}-\d{2}-\d{2}$/.test(form.effectiveFrom) &&
    (form.effectiveTo === "" || /^\d{4}-\d{2}-\d{2}$/.test(form.effectiveTo)),
);
const logicValid = computed(() => jsonError.value === null);
const reviewValid = computed(
  () => identityValid.value && scopeValid.value && logicValid.value,
);

const buildPayload = computed<ComplianceCreateRuleInstanceRequest | null>(() => {
  if (!reviewValid.value) return null;
  let parameters: Record<string, unknown>;
  try {
    parameters = JSON.parse(form.parametersJson || "{}");
  } catch {
    return null;
  }
  const payload: ComplianceCreateRuleInstanceRequest = {
    rule_type_id: form.ruleTypeId.trim(),
    name: form.name.trim(),
    description: form.description.trim(),
    parameters,
    effective_from: form.effectiveFrom,
    is_active: form.isActive,
    change_reason: form.changeReason.trim() || "initial creation",
  };
  if (form.effectiveTo) payload.effective_to = form.effectiveTo;
  return payload;
});

async function submit() {
  const payload = buildPayload.value;
  if (!payload) return;
  try {
    const res = await mutation.create(payload);
    if (res?.instance?.id) emit("created", res.instance.id);
  } catch {
    /* surfaced via composable */
  }
}
</script>

<template>
  <div class="builder">
    <AppCard>
      <p class="builder__notice">
        {{ t("compliance.builder.lifecycleNotice") }}
      </p>

      <ol class="builder__stepper" role="list">
        <li
          v-for="(s, i) in STEPS"
          :key="s.key"
          class="builder__step"
          :class="{
            'builder__step--active': i === stepIndex,
            'builder__step--done': i < stepIndex,
          }"
          :aria-current="i === stepIndex ? 'step' : undefined"
        >
          <span class="builder__step-num">{{ i + 1 }}</span>
          <span class="builder__step-label">{{ t(s.labelKey) }}</span>
        </li>
      </ol>
    </AppCard>

    <!-- Step 1: Identity -->
    <AppCard v-if="currentStep === 'identity'" :title="t('compliance.builder.steps.identity')">
      <div class="builder__notice builder__notice--info">
        {{ t("compliance.builder.identity.catalogUnavailable") }}
      </div>
      <div class="builder__grid">
        <label class="builder__field">
          <span class="builder__label">
            {{ t("compliance.builder.identity.ruleTypeId") }}
          </span>
          <select class="form-control" :value="form.ruleTypeId" @change="onRuleTypeChange">
            <option value="" disabled>{{ t("compliance.builder.identity.ruleTypeId") }}</option>
            <option
              v-for="entry in RULE_CATALOG"
              :key="entry.typeId"
              :value="entry.typeId"
              :disabled="!entry.selectable"
            >
              {{ t(entry.labelKey) }} — {{ entry.typeId }}
            </option>
          </select>
          <small class="builder__hint">
            {{ t("compliance.builder.identity.ruleTypeIdHelp") }}
          </small>
        </label>

        <label class="builder__field">
          <span class="builder__label">
            {{ t("compliance.builder.identity.name") }}
          </span>
          <input v-model="form.name" type="text" class="form-control" />
          <small v-if="form.name && form.name.trim().length < 3" class="builder__error">
            {{ t("compliance.builder.validation.nameTooShort") }}
          </small>
        </label>

        <label class="builder__field builder__field--wide">
          <span class="builder__label">
            {{ t("compliance.builder.identity.description") }}
          </span>
          <textarea v-model="form.description" rows="3" class="form-control" />
        </label>
      </div>
    </AppCard>

    <!-- Step 2: Scope -->
    <AppCard v-else-if="currentStep === 'scope'" :title="t('compliance.builder.steps.scope')">
      <div class="builder__grid">
        <label class="builder__field">
          <span class="builder__label">
            {{ t("compliance.builder.scope.effectiveFrom") }}
          </span>
          <input v-model="form.effectiveFrom" type="date" class="form-control" />
        </label>
        <label class="builder__field">
          <span class="builder__label">
            {{ t("compliance.builder.scope.effectiveTo") }}
          </span>
          <input v-model="form.effectiveTo" type="date" class="form-control" />
        </label>
        <label class="builder__field builder__field--wide builder__field--inline">
          <input v-model="form.isActive" type="checkbox" />
          <span>
            <strong>{{ t("compliance.builder.scope.isActive") }}</strong>
            <small class="builder__hint">{{ t("compliance.builder.scope.isActiveHelp") }}</small>
          </span>
        </label>
      </div>
    </AppCard>

    <!-- Step 3: Logic -->
    <AppCard v-else-if="currentStep === 'logic'" :title="t('compliance.builder.steps.logic')">
      <label class="builder__field">
        <span class="builder__label">
          {{ t("compliance.builder.logic.parametersTitle") }}
        </span>
        <textarea
          v-model="form.parametersJson"
          rows="10"
          class="form-control form-control--mono"
          spellcheck="false"
        />
        <small class="builder__hint">{{ t("compliance.builder.logic.parametersHelp") }}</small>
        <small v-if="jsonError" class="builder__error">{{ jsonError }}</small>
      </label>
    </AppCard>

    <!-- Step 4: Message (preview / change reason) -->
    <AppCard v-else-if="currentStep === 'message'" :title="t('compliance.builder.message.title')">
      <div class="builder__notice builder__notice--info">
        {{ t("compliance.builder.message.notice") }}
      </div>
      <label class="builder__field">
        <span class="builder__label">
          {{ t("compliance.builder.message.changeReason") }}
        </span>
        <input v-model="form.changeReason" type="text" class="form-control" />
        <small class="builder__hint">{{ t("compliance.builder.message.changeReasonHelp") }}</small>
      </label>
    </AppCard>

    <!-- Step 5: Review -->
    <AppCard v-else-if="currentStep === 'review'" :title="t('compliance.builder.review.title')">
      <h3 class="builder__section-title">{{ t("compliance.builder.review.summary") }}</h3>
      <dl class="builder__summary">
        <div><dt>{{ t("compliance.builder.identity.ruleTypeId") }}</dt><dd><code>{{ form.ruleTypeId || t("compliance.common.none") }}</code></dd></div>
        <div><dt>{{ t("compliance.builder.identity.name") }}</dt><dd>{{ form.name || t("compliance.common.none") }}</dd></div>
        <div><dt>{{ t("compliance.builder.identity.description") }}</dt><dd>{{ form.description || t("compliance.common.none") }}</dd></div>
        <div><dt>{{ t("compliance.builder.scope.effectiveFrom") }}</dt><dd>{{ form.effectiveFrom }}</dd></div>
        <div><dt>{{ t("compliance.builder.scope.effectiveTo") }}</dt><dd>{{ form.effectiveTo || t("compliance.common.none") }}</dd></div>
        <div><dt>{{ t("compliance.builder.scope.isActive") }}</dt><dd>{{ form.isActive ? t("compliance.badges.status.ACTIVE") : t("compliance.badges.status.DISABLED") }}</dd></div>
        <div><dt>{{ t("compliance.builder.message.changeReason") }}</dt><dd>{{ form.changeReason }}</dd></div>
      </dl>
      <h3 class="builder__section-title">{{ t("compliance.builder.review.payloadTitle") }}</h3>
      <pre class="builder__payload">{{ JSON.stringify(buildPayload, null, 2) || t("compliance.preTrade.form.invalid") }}</pre>
    </AppCard>

    <!-- Step 6: Submit -->
    <AppCard v-else-if="currentStep === 'submit'" :title="t('compliance.builder.submit.title')">
      <p class="builder__notice builder__notice--info">
        {{ t("compliance.builder.submit.notice") }}
      </p>

      <div class="builder__lifecycle">
        <AppButton variant="primary" size="sm" :disabled="true" :title="t('compliance.common.notYetAvailable')">
          {{ t("compliance.builder.lifecycleDisabled.submitForApproval") }}
        </AppButton>
        <AppButton variant="secondary" size="sm" :disabled="true" :title="t('compliance.common.notYetAvailable')">
          {{ t("compliance.builder.lifecycleDisabled.approve") }}
        </AppButton>
        <AppButton variant="ghost" size="sm" :disabled="true" :title="t('compliance.common.notYetAvailable')">
          {{ t("compliance.builder.lifecycleDisabled.reject") }}
        </AppButton>
        <AppButton variant="ghost" size="sm" :disabled="true" :title="t('compliance.common.notYetAvailable')">
          {{ t("compliance.builder.lifecycleDisabled.disable") }}
        </AppButton>
        <AppButton variant="ghost" size="sm" :disabled="true" :title="t('compliance.common.notYetAvailable')">
          {{ t("compliance.builder.lifecycleDisabled.archive") }}
        </AppButton>
      </div>

      <div
        v-if="mutation.error.value"
        class="builder__error builder__error--alert"
        role="alert"
      >
        {{ mutation.error.value }}
      </div>

      <div
        v-if="mutation.lastResult.value"
        class="builder__success"
        role="status"
      >
        {{ t("compliance.builder.submit.success") }} ·
        <code>{{ mutation.lastResult.value.instance.id }}</code>
      </div>

      <AppButton
        variant="primary"
        size="md"
        :loading="mutation.submitting.value"
        :disabled="!reviewValid || mutation.submitting.value"
        @click="submit"
      >
        {{
          mutation.submitting.value
            ? t("compliance.builder.submit.submitting")
            : t("compliance.builder.submit.cta")
        }}
      </AppButton>
    </AppCard>

    <div class="builder__nav">
      <AppButton
        variant="ghost"
        size="sm"
        :disabled="stepIndex === 0"
        @click="go('prev')"
      >
        Back
      </AppButton>
      <AppButton
        variant="secondary"
        size="sm"
        :disabled="stepIndex >= STEPS.length - 1"
        @click="go('next')"
      >
        Next
      </AppButton>
    </div>
  </div>
</template>

<style scoped>
.builder {
  display: grid;
  gap: var(--space-5);
}

.builder__notice {
  margin: 0 0 var(--space-4);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  border: 1px solid var(--alert-warning-border);
  font-size: var(--font-size-sm);
}

.builder__notice--info {
  background: var(--alert-info-bg);
  color: var(--alert-info-text);
  border-color: var(--alert-info-border);
}

.builder__stepper {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: var(--space-3);
}

.builder__step {
  display: grid;
  gap: var(--space-1);
  padding: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  text-align: center;
}

.builder__step--active {
  border-color: var(--action-primary);
  background: var(--bg-selected);
}

.builder__step--done {
  border-color: var(--state-success);
}

.builder__step-num {
  font-weight: var(--font-weight-bold);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.builder__step--active .builder__step-num {
  color: var(--action-primary);
}

.builder__step-label {
  font-size: var(--font-size-xs);
  color: var(--text-primary);
}

.builder__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-4);
  align-items: start;
}

@media (min-width: 640px) {
  .builder__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.builder__field {
  display: grid;
  gap: var(--space-2);
  margin: 0;
}

.builder__field--wide {
  grid-column: 1 / -1;
}

.builder__field--inline {
  display: inline-flex;
  align-items: flex-start;
  gap: var(--space-3);
}

.builder__label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.builder__hint {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.builder__error {
  font-size: var(--font-size-xs);
  color: var(--state-danger);
}

.builder__error--alert {
  margin-top: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  color: var(--alert-danger-text);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
}

.builder__success {
  margin-top: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
  border: 1px solid var(--alert-success-border);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
}

.builder__lifecycle {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-bottom: var(--space-5);
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

.form-control--mono {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
}

.form-control:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.builder__section-title {
  margin: 0 0 var(--space-2);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.builder__summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--space-3);
  margin: 0 0 var(--space-5);
}

.builder__summary dt {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.builder__summary dd {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.builder__summary code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
}

.builder__payload {
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-primary);
  overflow-x: auto;
  margin: 0;
}

.builder__nav {
  display: flex;
  justify-content: space-between;
}

@media (max-width: 720px) {
  .builder__stepper {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
