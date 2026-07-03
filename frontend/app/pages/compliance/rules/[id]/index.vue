<script setup lang="ts">
import { onMounted, ref } from "vue";

import { useI18n } from "~/composables/useI18n";

import ComplianceRuleDetailHeader from "~/features/compliance/components/ComplianceRuleDetailHeader.vue";
import ComplianceRuleDetailRail from "~/features/compliance/components/ComplianceRuleDetailRail.vue";
import ComplianceRuleTabs, {
  type TabKey,
} from "~/features/compliance/components/ComplianceRuleTabs.vue";
import { useComplianceRuleDetail } from "~/features/compliance/composables/useComplianceRules";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const detail = useComplianceRuleDetail();

const activeTab = ref<TabKey>("overview");

function handleOpenSimulator() {
  // Pre-fill simulator with the rule_type_id so the operator can see at a
  // glance which rule they're trying to exercise. The simulator itself uses
  // the query as a hint, not as form state — actual order context still
  // requires the user to pick portfolio / contract.
  const rule = detail.rule.value;
  void router.push({
    path: "/compliance/pre-trade",
    query: rule?.ruleTypeID ? { rule_type_id: rule.ruleTypeID } : {},
  });
}

onMounted(() => {
  const id = String(route.params.id ?? "");
  if (id) {
    void detail.fetchById(id);
  }
});
</script>

<template>
  <section class="rule-detail">
    <div v-if="detail.loading.value" class="rule-detail__state">
      {{ t("compliance.common.loading") }}
    </div>

    <div
      v-else-if="detail.error.value || !detail.rule.value"
      class="rule-detail__state rule-detail__state--error"
      role="alert"
    >
      {{ detail.error.value || "Rule not found." }}
    </div>

    <template v-else-if="detail.rule.value">
      <ComplianceRuleDetailHeader :rule="detail.rule.value" />
      <div class="rule-detail__layout">
        <div class="rule-detail__main">
          <ComplianceRuleTabs
            :rule="detail.rule.value"
            :active-tab="activeTab"
            @update:active-tab="(t: TabKey) => (activeTab = t)"
            @open-simulator="handleOpenSimulator"
          />
        </div>
        <aside class="rule-detail__rail">
          <ComplianceRuleDetailRail :rule="detail.rule.value" />
        </aside>
      </div>
    </template>
  </section>
</template>

<style scoped>
.rule-detail {
  display: grid;
  gap: var(--space-7);
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.rule-detail__state {
  padding: var(--space-7);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  color: var(--text-secondary);
}

.rule-detail__state--error {
  color: var(--state-danger);
}

.rule-detail__layout {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(260px, 1fr);
  gap: var(--space-6);
  align-items: start;
}

@media (max-width: 1024px) {
  .rule-detail__layout {
    grid-template-columns: 1fr;
  }
}
</style>
