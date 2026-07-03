<script setup lang="ts">
import { computed, onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

import ComplianceCategoryPanel from "~/features/compliance/components/ComplianceCategoryPanel.vue";
import ComplianceHighRiskList from "~/features/compliance/components/ComplianceHighRiskList.vue";
import ComplianceKpiCard from "~/features/compliance/components/ComplianceKpiCard.vue";
import ComplianceNoRulesEmptyState from "~/features/compliance/components/ComplianceNoRulesEmptyState.vue";
import ComplianceRecentChangesPanel from "~/features/compliance/components/ComplianceRecentChangesPanel.vue";
import ComplianceRecentFailuresPanel from "~/features/compliance/components/ComplianceRecentFailuresPanel.vue";
import ComplianceSearchCard from "~/features/compliance/components/ComplianceSearchCard.vue";
import ComplianceSectionTabs from "~/features/compliance/components/ComplianceSectionTabs.vue";
import { useComplianceBreachesList } from "~/features/compliance/composables/useComplianceBreaches";
import { useComplianceRuleDirectory } from "~/features/compliance/composables/useComplianceRuleDirectory";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const { t } = useI18n();
const authStore = useAuthStore();

const ruleDir = useComplianceRuleDirectory();
const recentBreaches = useComplianceBreachesList();
const openBreaches = useComplianceBreachesList();

const canCreate = computed(() =>
  authStore.hasPermission("IRG_EDIT_RULE_INSTANCE"),
);

// All KPIs derive from the shared rule directory. If the call errors or is
// still loading we surface null → "—" rather than a misleading "0".
const errored = computed(() => Boolean(ruleDir.error.value));
const loadingDir = computed(() => ruleDir.loading.value);

function bucketCount(status: "ACTIVE" | "SCHEDULED" | "EXPIRED" | "DISABLED"): number | null {
  if (loadingDir.value || errored.value) return null;
  return ruleDir.byDerivedStatus.value.get(status)?.length ?? 0;
}

const activeCount = computed(() => bucketCount("ACTIVE"));
const expiredCount = computed(() => bucketCount("EXPIRED"));
const disabledCount = computed(() => bucketCount("DISABLED"));

// "Draft" has no backend representation today. Best-honest approximation:
// rules that are currently SCHEDULED (effective_from in the future) AND
// flagged active=false. Backend can't distinguish "drafted but unsubmitted"
// from "intentionally pre-staged".
const draftCount = computed(() => {
  if (loadingDir.value || errored.value) return null;
  const scheduled = ruleDir.byDerivedStatus.value.get("SCHEDULED") ?? [];
  return scheduled.filter((r) => !r.isActive).length;
});

// "Pending review" requires an approval flow. None exists → always 0,
// surfaced with a hint tooltip.
const pendingReviewCount = computed(() =>
  loadingDir.value || errored.value ? null : 0,
);

const highRiskCount = computed(() =>
  loadingDir.value || errored.value
    ? null
    : ruleDir.highRiskRules.value.length,
);

const recentChangesCount = computed(() => null); // No audit feed endpoint yet.

const openBreachCount = computed(() => {
  if (openBreaches.loading.value || openBreaches.error.value) return null;
  return openBreaches.total.value;
});

const breachTabCount = computed(() => openBreachCount.value);

const tabCounts = computed(() => ({
  library: {
    value: loadingDir.value || errored.value ? null : ruleDir.total.value,
  },
  approvals: {
    value: null,
    unavailableReason: "Approvals feed is not yet supported by the backend.",
  },
  breaches: { value: breachTabCount.value },
  exceptions: {
    value: null,
    unavailableReason: "Exception inbox is not yet supported by the backend.",
  },
  audit: {
    value: null,
    unavailableReason: "Compliance audit feed is not yet supported.",
  },
}));

const noRules = computed(
  () =>
    !ruleDir.loading.value &&
    !ruleDir.error.value &&
    ruleDir.total.value === 0,
);

onMounted(() => {
  void ruleDir.ensureLoaded();
  void recentBreaches.fetchList({ limit: 5 });
  void openBreaches.fetchList({ limit: 1, status: "OPEN" });
});
</script>

<template>
  <section class="overview">
    <!-- Header strip: breadcrumb + restricted chip + action buttons -->
    <div class="overview__topbar">

      <div class="overview__topbar-actions">
        <NuxtLink to="/compliance/audit" class="overview__top-link">
          <AppIcon name="clock" size="xs" />
          <span>{{ t("compliance.dashboard.headerActions.recentChanges") }}</span>
          <span
            class="overview__top-badge overview__top-badge--muted"
            :title="t('compliance.dashboard.recentChanges.empty')"
          >—</span>
        </NuxtLink>
        <NuxtLink to="/compliance/post-trade" class="overview__top-link">
          <AppIcon name="warning" size="xs" />
          <span>{{ t("compliance.dashboard.headerActions.openBreaches") }}</span>
          <span
            class="overview__top-badge overview__top-badge--danger"
          >
            {{ openBreachCount === null ? "—" : openBreachCount }}
          </span>
        </NuxtLink>
        <NuxtLink v-if="canCreate" to="/compliance/rules/new">
          <AppButton variant="primary" size="sm" class="overview__new-rule-btn">
            + {{ t("compliance.dashboard.headerActions.newRule") }}
          </AppButton>
        </NuxtLink>
      </div>
    </div>



    <!-- Optional truncation banner if backend has more rules than we pulled -->
    <div
      v-if="ruleDir.truncated.value"
      class="overview__truncated"
      role="status"
    >
      {{
        t("compliance.dashboard.truncatedNotice", {
          shown: ruleDir.items.value.length,
          total: ruleDir.total.value,
        })
      }}
    </div>

    <!-- No-rules state owns the screen below tabs -->
    <ComplianceNoRulesEmptyState v-if="noRules" />

    <!-- KPI strip -->
    <section v-if="!noRules" class="overview__kpis" aria-label="Compliance KPIs">
      <ComplianceKpiCard
        :label="t('compliance.dashboard.kpi.activeRules')"
        :value="activeCount"
        :subtitle="t('compliance.dashboard.kpi.activeRulesSub')"
        tone="success"
        :loading="loadingDir"
      />
      <ComplianceKpiCard
        :label="t('compliance.dashboard.kpi.draft')"
        :value="draftCount"
        :subtitle="t('compliance.dashboard.kpi.draftSub')"
        :hint="t('compliance.dashboard.kpi.draftHint')"
        :loading="loadingDir"
      />
      <ComplianceKpiCard
        :label="t('compliance.dashboard.kpi.pendingReview')"
        :value="pendingReviewCount"
        :subtitle="t('compliance.dashboard.kpi.pendingReviewSub')"
        :hint="t('compliance.dashboard.kpi.pendingReviewHint')"
        tone="warning"
        :loading="loadingDir"
      />
      <ComplianceKpiCard
        :label="t('compliance.dashboard.kpi.disabled')"
        :value="disabledCount"
        :subtitle="t('compliance.dashboard.kpi.disabledSub')"
        tone="muted"
        :loading="loadingDir"
      />
      <ComplianceKpiCard
        :label="t('compliance.dashboard.kpi.expired')"
        :value="expiredCount"
        :subtitle="t('compliance.dashboard.kpi.expiredSub')"
        tone="danger"
        :loading="loadingDir"
      />
      <ComplianceKpiCard
        :label="t('compliance.dashboard.kpi.highRisk')"
        :value="highRiskCount"
        :subtitle="t('compliance.dashboard.kpi.highRiskSub')"
        tone="danger"
        :loading="loadingDir"
      />
    </section>

    <!-- Error alerts (one per failing endpoint) -->
    <div v-if="ruleDir.error.value" class="overview__error" role="alert">
      <strong>Rule list:</strong> {{ ruleDir.error.value }}
    </div>
    <div v-if="recentBreaches.error.value" class="overview__error" role="alert">
      <strong>Breach feed:</strong> {{ recentBreaches.error.value }}
    </div>

    <!-- Two-column body -->
    <div v-if="!noRules" class="overview__body">
      <div class="overview__col-main">
        <ComplianceRecentChangesPanel />
        <ComplianceRecentFailuresPanel
          :items="recentBreaches.items.value"
          :loading="recentBreaches.loading.value"
          :error="recentBreaches.error.value"
        />
      </div>
      <div class="overview__col-rail">
        <ComplianceSearchCard />
        <ComplianceCategoryPanel
          :by-category="ruleDir.byCategory.value"
          :loading="loadingDir"
        />
        <ComplianceHighRiskList
          :items="ruleDir.highRiskRules.value"
          :loading="loadingDir"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.overview {
  display: grid;
  gap: var(--space-5);
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Topbar */
.overview__topbar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-4);
  flex-wrap: wrap;
  padding-bottom: var(--space-2);
}

.overview__topbar-actions {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.overview__top-link {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: 0 var(--space-4);
  height: 32px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  text-decoration: none;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.02);
  transition: all 0.2s ease;
}

.overview__top-link:hover {
  background: var(--bg-card-hover);
  border-color: var(--border-strong);
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
}

.overview__top-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 18px;
  min-width: 18px;
  padding: 0 var(--space-2);
  border-radius: var(--radius-pill);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.overview__top-badge--danger {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border: 1px solid var(--alert-danger-border);
}

.overview__top-badge--muted {
  opacity: 0.6;
}

.overview__new-rule-btn {
  box-shadow: 0 2px 8px rgba(9, 105, 218, 0.2);
  transition: all 0.2s ease;
}

.overview__new-rule-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(9, 105, 218, 0.3);
}

.overview__truncated {
  padding: var(--space-3) var(--space-4);
  background: var(--alert-info-bg);
  color: var(--alert-info-text);
  border: 1px solid var(--alert-info-border);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.02);
}

/* KPI strip */
.overview__kpis {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: var(--space-4);
}

@media (max-width: 1280px) {
  .overview__kpis {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .overview__kpis {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.overview__error {
  padding: var(--space-3) var(--space-4);
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border: 1px solid var(--alert-danger-border);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  box-shadow: 0 2px 4px rgba(240, 68, 56, 0.05);
}

/* Two-column body */
.overview__body {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr);
  gap: var(--space-5);
  align-items: start;
}

.overview__col-main,
.overview__col-rail {
  display: grid;
  gap: var(--space-5);
}

@media (max-width: 1024px) {
  .overview__body {
    grid-template-columns: 1fr;
  }
}
</style>
