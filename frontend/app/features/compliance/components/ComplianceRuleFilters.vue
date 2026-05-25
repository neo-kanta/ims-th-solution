<script setup lang="ts">
import { reactive, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";

import type { ComplianceRuleListFilters } from "../types";

interface Props {
  modelValue: ComplianceRuleListFilters;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  "update:modelValue": [value: ComplianceRuleListFilters];
  apply: [];
  reset: [];
}>();

const { t } = useI18n();

type ActiveChoice = "any" | "true" | "false";

interface FormState {
  ruleTypeId: string;
  isActive: ActiveChoice;
}

function initialFormFrom(filters: ComplianceRuleListFilters): FormState {
  let isActive: ActiveChoice = "any";
  if (filters.is_active === true) isActive = "true";
  else if (filters.is_active === false) isActive = "false";

  return {
    ruleTypeId: filters.rule_type_id ?? "",
    isActive,
  };
}

const form = reactive<FormState>(initialFormFrom(props.modelValue));

watch(
  () => props.modelValue,
  (next) => {
    const fresh = initialFormFrom(next);
    form.ruleTypeId = fresh.ruleTypeId;
    form.isActive = fresh.isActive;
  },
  { deep: true },
);

function buildFilters(): ComplianceRuleListFilters {
  const next: ComplianceRuleListFilters = {};
  const ruleTypeId = form.ruleTypeId.trim();
  if (ruleTypeId) next.rule_type_id = ruleTypeId;
  if (form.isActive === "true") next.is_active = true;
  if (form.isActive === "false") next.is_active = false;
  return next;
}

function apply() {
  emit("update:modelValue", buildFilters());
  emit("apply");
}

function reset() {
  form.ruleTypeId = "";
  form.isActive = "any";
  emit("update:modelValue", {});
  emit("reset");
}
</script>

<template>
  <AppCard :title="t('compliance.rules.filters.title')">
    <form class="rules-filters" @submit.prevent="apply">
      <label class="rules-filters__field">
        <span class="rules-filters__label">
          {{ t("compliance.rules.filters.ruleTypeId") }}
        </span>
        <input
          v-model="form.ruleTypeId"
          class="form-control"
          type="text"
          :placeholder="t('compliance.rules.filters.ruleTypeIdHint')"
        />
      </label>

      <fieldset class="rules-filters__field">
        <legend class="rules-filters__label">
          {{ t("compliance.rules.filters.isActive") }}
        </legend>
        <div class="rules-filters__radios">
          <label>
            <input v-model="form.isActive" type="radio" value="any" />
            {{ t("compliance.rules.filters.any") }}
          </label>
          <label>
            <input v-model="form.isActive" type="radio" value="true" />
            {{ t("compliance.rules.filters.activeOnly") }}
          </label>
          <label>
            <input v-model="form.isActive" type="radio" value="false" />
            {{ t("compliance.rules.filters.inactiveOnly") }}
          </label>
        </div>
      </fieldset>

      <div class="rules-filters__actions">
        <AppButton
          variant="primary"
          size="sm"
          :loading="loading"
          @click="apply"
        >
          {{ t("compliance.rules.filters.apply") }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="reset">
          {{ t("compliance.rules.filters.reset") }}
        </AppButton>
      </div>
    </form>
  </AppCard>
</template>

<style scoped>
.rules-filters {
  display: grid;
  gap: var(--space-5);
}

.rules-filters__field {
  display: grid;
  gap: var(--space-2);
  margin: 0;
  padding: 0;
  border: 0;
}

.rules-filters__label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.form-control {
  height: var(--size-control-md);
  padding: 0 var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.form-control:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.rules-filters__radios {
  display: grid;
  gap: var(--space-2);
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.rules-filters__radios label {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  cursor: pointer;
}

.rules-filters__actions {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}
</style>
