<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";

import CompliancePager from "~/features/compliance/components/CompliancePager.vue";
import ComplianceRuleFilters from "~/features/compliance/components/ComplianceRuleFilters.vue";
import ComplianceRuleTable from "~/features/compliance/components/ComplianceRuleTable.vue";
import ComplianceSectionTabs from "~/features/compliance/components/ComplianceSectionTabs.vue";
import { useComplianceRulesList } from "~/features/compliance/composables/useComplianceRules";
import { useComplianceRuleDirectory } from "~/features/compliance/composables/useComplianceRuleDirectory";
import { useComplianceBreachesList } from "~/features/compliance/composables/useComplianceBreaches";
import type { ComplianceRuleListFilters } from "~/features/compliance/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const route = useRoute();
const { t } = useI18n();
const authStore = useAuthStore();
const canCreate = computed(() => authStore.hasPermission("IRG_EDIT_RULE_INSTANCE"));

const filters = ref<ComplianceRuleListFilters>({});
const rules = useComplianceRulesList();

const ruleDir = useComplianceRuleDirectory();
const openBreaches = useComplianceBreachesList();

const openBreachCount = computed(() => {
  if (openBreaches.loading.value || openBreaches.error.value) return null;
  return openBreaches.total.value;
});

const tabCounts = computed(() => ({
  library: {
    value: ruleDir.loading.value || ruleDir.error.value ? null : ruleDir.total.value,
  },
  approvals: {
    value: null,
    unavailableReason: "Approvals feed is not yet supported by the backend.",
  },
  breaches: { value: openBreachCount.value },
  exceptions: {
    value: null,
    unavailableReason: "Exception inbox is not yet supported by the backend.",
  },
  audit: {
    value: null,
    unavailableReason: "Compliance audit feed is not yet supported.",
  },
}));

// Permits Phase 1 deep-link from the pre-trade result panel — `?highlight=<rule_type_id>`
// scrolls the matching row into view without (yet) opening the detail page.
const highlight = ref<string | null>(
  typeof route.query.highlight === "string" ? route.query.highlight : null,
);

async function refresh() {
  await rules.fetchList(filters.value);
}

function setOffset(next: number) {
  rules.offset.value = next;
  void refresh();
}

function setLimit(next: number) {
  rules.limit.value = next;
  rules.offset.value = 0;
  void refresh();
}

onMounted(() => {
  void refresh();
  void ruleDir.ensureLoaded();
  void openBreaches.fetchList({ limit: 1, status: "OPEN" });
});
</script>

<template>
  <section class="rules-page">
    <div class="rules-page__toolbar">
      <NuxtLink v-if="canCreate" to="/compliance/rules/new">
        <AppButton variant="primary" size="sm" class="rules-page__create-btn">
          {{ t("compliance.nav.newRule") }}
        </AppButton>
      </NuxtLink>
      <span v-else class="rules-page__note">
        IRG_EDIT_RULE_INSTANCE required to create rules.
      </span>
    </div>

    <div class="rules-page__layout">
      <aside class="rules-page__rail">
        <ComplianceRuleFilters
          v-model="filters"
          :loading="rules.loading.value"
          @apply="refresh"
          @reset="refresh"
        />
      </aside>
      <div class="rules-page__main">
        <ComplianceRuleTable
          :items="rules.items.value"
          :loading="rules.loading.value"
          :error="rules.error.value"
          :highlight-id="highlight"
        />
        <CompliancePager
          :total="rules.total.value"
          :offset="rules.offset.value"
          :limit="rules.limit.value"
          :loading="rules.loading.value"
          @update:offset="setOffset"
          @update:limit="setLimit"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.rules-page {
  display: grid;
  gap: var(--space-4);
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}

.rules-page__toolbar {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-bottom: var(--space-2);
}

.rules-page__create-btn {
  box-shadow: 0 2px 8px rgba(9, 105, 218, 0.2);
  transition: all 0.2s ease;
}

.rules-page__create-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(9, 105, 218, 0.3);
}

.rules-page__note {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  max-width: 22rem;
}

.rules-page__layout {
  display: grid;
  grid-template-columns: minmax(260px, 280px) 1fr;
  gap: var(--space-6);
  align-items: start;
  margin-top: var(--space-4);
}

.rules-page__main {
  display: grid;
  gap: var(--space-4);
}

.rules-page__pagination {
  display: flex;
  justify-content: flex-end;
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

@media (max-width: 1024px) {
  .rules-page__layout {
    grid-template-columns: 1fr;
  }
}
</style>
