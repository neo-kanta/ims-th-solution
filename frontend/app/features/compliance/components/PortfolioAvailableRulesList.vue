<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";
import AppSection from "~/shared/ui/AppSection.vue";

import type { ApiPortfolioRuleCatalogEntry } from "../../portfolio-workspace/services/portfolioComplianceApi";
import { ruleExplanation, ruleLabel } from "../lib/ruleTypeCatalog";

interface Props {
  entries: ApiPortfolioRuleCatalogEntry[];
  loading: boolean;
  canManage: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  bind: [entry: ApiPortfolioRuleCatalogEntry];
}>();

const { t } = useI18n();
</script>

<template>
  <AppSection
    :title="t('portfolio.compliance.availableTitle')"
    :description="t('portfolio.compliance.availableSubtitle')"
  >
    <AppEmptyState
      v-if="!loading && entries.length === 0"
      icon="folder"
      :title="t('portfolio.compliance.noneAvailable')"
    />

    <ul v-else class="available-list" role="list">
      <li v-for="entry in entries" :key="entry.rule_instance_id" class="available-item">
        <div class="available-item__copy">
          <div class="available-item__label">{{ ruleLabel(entry.rule_type_id ?? "", t) }}</div>
          <div class="available-item__explanation">
            {{ ruleExplanation(entry.rule_type_id ?? "", entry.description ?? "", t) }}
          </div>
          <code class="available-item__type-id">{{ entry.rule_type_id }}</code>
        </div>
        <AppButton
          v-if="canManage"
          variant="secondary"
          size="sm"
          @click="emit('bind', entry)"
        >
          {{ t("portfolio.compliance.bind") }}
        </AppButton>
        <span v-else class="available-item__readonly">{{ t("portfolio.compliance.readOnly") }}</span>
      </li>
    </ul>
  </AppSection>
</template>

<style scoped>
.available-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-3);
}

.available-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.available-item__copy {
  min-width: 0;
  display: grid;
  gap: 2px;
}

.available-item__label {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.available-item__explanation {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  max-width: 34rem;
}

.available-item__type-id {
  margin-top: 2px;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-2xs);
  color: var(--text-tertiary);
}

.available-item__readonly {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  font-style: italic;
  white-space: nowrap;
}
</style>
