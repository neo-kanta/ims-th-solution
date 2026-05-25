<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppTabs, { type TabItem } from "~/shared/ui/AppTabs.vue";

import { lookupRuleCatalog } from "../lib/ruleTypeCatalog";
import type { ComplianceRule } from "../types";

interface Props {
  rule: ComplianceRule;
  activeTab: TabKey;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  "update:activeTab": [key: TabKey];
  "open-simulator": [];
}>();

export type TabKey =
  | "overview"
  | "logic"
  | "scope"
  | "test"
  | "approval"
  | "audit"
  | "settings";

const TABS: { key: TabKey; labelKey: string }[] = [
  { key: "overview", labelKey: "compliance.detail.tabs.overview" },
  { key: "logic", labelKey: "compliance.detail.tabs.logic" },
  { key: "scope", labelKey: "compliance.detail.tabs.scope" },
  { key: "test", labelKey: "compliance.detail.tabs.test" },
  { key: "approval", labelKey: "compliance.detail.tabs.approval" },
  { key: "audit", labelKey: "compliance.detail.tabs.audit" },
  { key: "settings", labelKey: "compliance.detail.tabs.settings" },
];

const { t } = useI18n();
const catalogEntry = computed(() => lookupRuleCatalog(props.rule.ruleTypeID));

const tabItems = computed<TabItem[]>(() =>
  TABS.map((tab) => ({
    key: tab.key,
    label: t(tab.labelKey as any),
  })),
);

function activate(key: string) {
  emit("update:activeTab", key as TabKey);
}
</script>

<template>
  <div class="rule-tabs">
    <AppTabs :items="tabItems" :model-value="activeTab" @change="activate" />

    <div class="rule-tabs__panel" role="tabpanel">
      <!-- Overview -->
      <AppCard v-if="activeTab === 'overview'">
        <h3 class="rule-tabs__section-title">
          {{ t("compliance.detail.overview.descriptionTitle") }}
        </h3>
        <p v-if="rule.description" class="rule-tabs__body">
          {{ rule.description }}
        </p>
        <p v-else class="rule-tabs__muted">
          {{ t("compliance.detail.overview.descriptionEmpty") }}
        </p>

        <template v-if="rule.type_metadata?.description">
          <h3 class="rule-tabs__section-title">
            {{ t("compliance.detail.overview.metadataDescription") }}
          </h3>
          <p class="rule-tabs__body">{{ rule.type_metadata.description }}</p>
        </template>

        <template v-if="catalogEntry">
          <h3 class="rule-tabs__section-title">Why it matters</h3>
          <p class="rule-tabs__body">{{ catalogEntry.explanation }}</p>
          <h3 class="rule-tabs__section-title">Typical correction</h3>
          <p class="rule-tabs__body">{{ catalogEntry.suggestedCorrection }}</p>
        </template>
      </AppCard>

      <!-- Logic -->
      <AppCard v-else-if="activeTab === 'logic'" :title="t('compliance.detail.logic.title')" :subtitle="t('compliance.detail.logic.description')">
        <h3 class="rule-tabs__section-title">
          {{ t("compliance.detail.logic.parametersTitle") }}
        </h3>
        <div class="rule-tabs__muted rule-tabs__notice">
          {{ t("compliance.detail.logic.parametersUnavailable") }}
        </div>
      </AppCard>

      <!-- Scope -->
      <AppCard v-else-if="activeTab === 'scope'" :title="t('compliance.detail.scope.title')">
        <div v-if="rule.type_metadata?.supported_scopes?.length">
          <h3 class="rule-tabs__section-title">
            {{ t("compliance.detail.scope.scopesLabel") }}
          </h3>
          <ul class="rule-tabs__chips">
            <li v-for="s in rule.type_metadata.supported_scopes" :key="s">
              {{ s }}
            </li>
          </ul>
        </div>
        <h3 class="rule-tabs__section-title">
          {{ t("compliance.detail.scope.bindingsTitle") }}
        </h3>
        <div class="rule-tabs__muted rule-tabs__notice">
          {{ t("compliance.detail.scope.bindingsUnavailable") }}
        </div>
      </AppCard>

      <!-- Test -->
      <AppCard v-else-if="activeTab === 'test'" :title="t('compliance.detail.test.title')" :subtitle="t('compliance.detail.test.description')">
        <AppButton variant="primary" size="sm" @click="emit('open-simulator')">
          {{ t("compliance.detail.test.cta") }}
        </AppButton>
      </AppCard>

      <!-- Approval -->
      <AppCard v-else-if="activeTab === 'approval'" :title="t('compliance.detail.approval.title')">
        <div class="rule-tabs__muted rule-tabs__notice">
          {{ t("compliance.detail.approval.unavailable") }}
        </div>
      </AppCard>

      <!-- Audit -->
      <AppCard v-else-if="activeTab === 'audit'" :title="t('compliance.detail.audit.title')">
        <div class="rule-tabs__muted rule-tabs__notice">
          {{ t("compliance.detail.audit.unavailable") }}
        </div>
      </AppCard>

      <!-- Settings -->
      <AppCard v-else :title="t('compliance.detail.settings.title')">
        <div class="rule-tabs__muted rule-tabs__notice">
          {{ t("compliance.detail.settings.unavailable") }}
        </div>
      </AppCard>
    </div>
  </div>
</template>

<style scoped>
.rule-tabs {
  display: grid;
  gap: var(--space-4);
}

.rule-tabs__bar {
  display: flex;
  gap: var(--space-2);
  border-bottom: 1px solid var(--border-subtle);
  overflow-x: auto;
}

.rule-tabs__tab {
  border: 0;
  background: transparent;
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  font-family: inherit;
}

.rule-tabs__tab:hover {
  color: var(--text-primary);
}

.rule-tabs__tab--active {
  color: var(--text-primary);
  border-bottom-color: var(--action-primary);
  font-weight: var(--font-weight-semibold);
}

.rule-tabs__section-title {
  margin: 0 0 var(--space-3);
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.rule-tabs__section-title + .rule-tabs__body,
.rule-tabs__section-title + .rule-tabs__muted {
  margin-bottom: var(--space-5);
}

.rule-tabs__body {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-relaxed);
}

.rule-tabs__muted {
  color: var(--text-tertiary);
  font-style: italic;
  font-size: var(--font-size-sm);
}

.rule-tabs__notice {
  padding: var(--space-3) var(--space-4);
  border: 1px dashed var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
}

.rule-tabs__chips {
  display: flex;
  gap: var(--space-2);
  list-style: none;
  padding: 0;
  margin: 0 0 var(--space-5);
  flex-wrap: wrap;
}

.rule-tabs__chips li {
  padding: 0 var(--space-3);
  height: 24px;
  display: inline-flex;
  align-items: center;
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  border-radius: var(--radius-pill);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
}
</style>
