<script setup lang="ts">
import { computed, onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";

import { useComplianceUserDirectory } from "../composables/useComplianceUserDirectory";
import { deriveRuleStatus, formatIsoDateTime, ruleStatusTone } from "../lib/formatters";
import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceRule } from "../types";
import ComplianceRuleStatusBadge from "./ComplianceRuleStatusBadge.vue";
import ComplianceSeverityBadge from "./ComplianceSeverityBadge.vue";

interface Props {
  rule: ComplianceRule;
}

const props = defineProps<Props>();

const { t } = useI18n();
const users = useComplianceUserDirectory();

onMounted(() => {
  void users.ensureLoaded();
});

const status = computed(() => deriveRuleStatus(props.rule));
const displayName = computed(
  () => props.rule.name || ruleLabel(props.rule.ruleTypeID, t),
);
const ownerLabel = computed(() => users.labelFor(props.rule.createdBy));

const accentClass = computed(() => {
  const tone = ruleStatusTone(status.value);
  return `rule-header--accent-${tone}`;
});
</script>

<template>
  <header class="rule-header-container">
    <div class="rule-header__back">
      <NuxtLink to="/compliance/rules" class="rule-header__back-link">
        <span class="rule-header__back-arrow">←</span> {{ t("compliance.detail.backToLibrary") }}
      </NuxtLink>
    </div>

    <div :class="['rule-header', accentClass]">
      <div class="rule-header__main">
        <div class="rule-header__copy">
          <div class="rule-header__top-row">
            <code class="rule-header__type">{{ rule.ruleTypeID }}</code>
            <span class="rule-header__version">v{{ rule.currentVersion }}</span>
          </div>
          <h1 class="rule-header__title">{{ displayName }}</h1>
          <div class="rule-header__badges">
            <ComplianceRuleStatusBadge :status="status" size="md" />
            <ComplianceSeverityBadge
              v-slot:default
              v-if="rule.type_metadata?.default_severity"
              :severity="rule.type_metadata.default_severity"
              size="md"
            />
          </div>
          <div class="rule-header__meta">
            <div class="rule-header__meta-item">
              <span class="rule-header__meta-label">Updated:</span>
              <span class="rule-header__meta-value">{{ formatIsoDateTime(rule.updatedAt) }}</span>
            </div>
            <div class="rule-header__meta-item">
              <span class="rule-header__meta-label">Owner:</span>
              <span class="rule-header__meta-value" :title="rule.createdBy">{{ ownerLabel }}</span>
            </div>
          </div>
        </div>

        <div class="rule-header__actions">
          <AppButton
            variant="secondary"
            size="sm"
            :disabled="true"
            class="rule-header__btn"
            :title="t('compliance.detail.header.editUnavailable')"
          >
            Edit
          </AppButton>
          <AppButton
            variant="ghost"
            size="sm"
            :disabled="true"
            class="rule-header__btn"
            :title="t('compliance.detail.header.disableUnavailable')"
          >
            Disable
          </AppButton>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
.rule-header-container {
  display: grid;
  gap: var(--space-3);
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}

.rule-header__back-link {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  text-decoration: none;
  font-weight: var(--font-weight-medium);
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  transition: all var(--transition-fast);
}

.rule-header__back-arrow {
  transition: transform var(--transition-fast);
}

.rule-header__back-link:hover {
  color: var(--text-primary);
  text-decoration: none;
}

.rule-header__back-link:hover .rule-header__back-arrow {
  transform: translateX(-4px);
}

.rule-header {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--space-5) var(--space-6);
  box-shadow: var(--shadow-sm);
  position: relative;
  overflow: hidden;
  transition: all var(--transition-base);
}

.rule-header:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--border-strong);
}

/* Status Accents */
.rule-header::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: var(--text-tertiary);
  transition: background var(--transition-base);
}

.rule-header--accent-success::before {
  background: linear-gradient(90deg, var(--state-success), var(--color-success-400));
}

.rule-header--accent-info::before {
  background: linear-gradient(90deg, var(--state-info), var(--color-info-400));
}

.rule-header--accent-warning::before {
  background: linear-gradient(90deg, var(--state-warning), var(--color-warning-400));
}

.rule-header--accent-error::before {
  background: linear-gradient(90deg, var(--state-danger), var(--color-danger-400));
}

.rule-header--accent-neutral::before {
  background: var(--border-default);
}

.rule-header__main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-5);
  flex-wrap: wrap;
}

.rule-header__copy {
  display: grid;
  gap: var(--space-3);
  min-width: 0;
  flex: 1;
}

.rule-header__top-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.rule-header__type {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-2xs);
  color: var(--text-tertiary);
  background: var(--bg-card-muted);
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-xs);
  border: 1px solid var(--border-subtle);
}

.rule-header__version {
  display: inline-flex;
  align-items: center;
  height: 18px;
  padding: 0 var(--space-2);
  border-radius: var(--radius-pill);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-bold);
  border: 1px solid var(--border-subtle);
}

.rule-header__title {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  letter-spacing: -0.02em;
  line-height: var(--line-height-tight);
}

.rule-header__badges {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.rule-header__meta {
  display: flex;
  gap: var(--space-5);
  flex-wrap: wrap;
  margin-top: var(--space-1);
}

.rule-header__meta-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-xs);
}

.rule-header__meta-label {
  color: var(--text-tertiary);
  font-weight: var(--font-weight-medium);
}

.rule-header__meta-value {
  color: var(--text-secondary);
}

.rule-header__actions {
  display: inline-flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  align-self: center;
}

.rule-header__btn {
  opacity: 0.8;
  transition: all var(--transition-fast);
}

.rule-header__btn:hover {
  opacity: 1;
}
</style>

