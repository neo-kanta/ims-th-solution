<script setup lang="ts">
import { computed } from "vue";

import AppBadge from "~/shared/ui/AppBadge.vue";
import { useI18n } from "~/composables/useI18n";

import { formatMoneyCompact, formatPercent, relativeTime } from "../../lib/format";
import type { MyFundCard } from "../../types";

import ComplianceBadge from "../ComplianceBadge.vue";
import FundWorkflowBar from "../FundWorkflowBar.vue";

interface Props {
  card: MyFundCard;
  activeTab: string;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), { loading: false });
const emit = defineEmits<{
  (e: "back"): void;
  (e: "select-tab", tab: string): void;
}>();
const { t } = useI18n();

interface TabDef {
  key: string;
  label: string;
  badge?: number;
  disabled?: boolean;
  hint?: string;
}

const tabs = computed<TabDef[]>(() => [
  { key: "holdings", label: t("myFunds.detail.tabs.holdings", "Holdings") },
  {
    key: "stages",
    label: t("myFunds.detail.tabs.stages", "Workflow"),
    badge: undefined,
  },
  {
    key: "compliance",
    label: t("myFunds.detail.tabs.compliance", "Compliance"),
    badge: props.card.compliance.open_count || undefined,
  },
  {
    key: "decisions",
    label: t("myFunds.detail.tabs.decisions", "Decisions"),
    disabled: true,
    hint: t("myFunds.detail.tabs.notImplemented", "Backend pending"),
  },
  {
    key: "audit",
    label: t("myFunds.detail.tabs.audit", "Activity"),
  },
  {
    key: "reviewers",
    label: t("myFunds.detail.tabs.reviewers", "Reviewers"),
    disabled: true,
    hint: t("myFunds.detail.tabs.notImplemented", "Backend pending"),
  },
  {
    key: "settings",
    label: t("myFunds.detail.tabs.settings", "Settings"),
  },
]);

const statusVariant = computed(() => {
  switch (props.card.status) {
    case "ACTIVE": return "success";
    case "LOCKED": return "locked";
    case "CLOSED": return "neutral";
    case "BREACH": return "error";
    default: return "neutral";
  }
});

const statusLabel = computed(() => {
  switch (props.card.status) {
    case "ACTIVE": return t("myFunds.status.active", "Active");
    case "LOCKED": return t("myFunds.status.locked", "Locked");
    case "CLOSED": return t("myFunds.status.closed", "Closed");
    case "BREACH": return t("myFunds.status.breach", "Breach");
    default: return props.card.fund_status_raw || "—";
  }
});

const navLabel = computed(() =>
  formatMoneyCompact(props.card.valuation.nav_numeric, props.card.valuation.valuation_ccy),
);

const aumLabel = computed(() =>
  formatMoneyCompact(props.card.valuation.aum_numeric, props.card.valuation.valuation_ccy),
);

const unrealisedLabel = computed(() =>
  formatMoneyCompact(
    props.card.valuation.unrealised_pnl_numeric,
    props.card.valuation.valuation_ccy,
    true,
  ),
);

const cashBufferLabel = computed(() => formatPercent(props.card.valuation.cash_buffer_pct, 2));

const trend = computed<"up" | "down" | "flat">(() => {
  const v = props.card.valuation.unrealised_pnl_numeric;
  if (v > 0) return "up";
  if (v < 0) return "down";
  return "flat";
});

const updatedLabel = computed(() => relativeTime(props.card.updated_at));

function selectTab(t: TabDef) {
  if (t.disabled) return;
  emit("select-tab", t.key);
}
</script>

<template>
  <header class="fund-detail-header" :class="{ 'is-loading': loading }">
    <div class="fund-detail-header__top">
      <button
        type="button"
        class="fund-detail-header__back"
        @click="emit('back')"
      >
        <span aria-hidden="true">←</span>
        {{ t("myFunds.detail.back", "Back to my funds") }}
      </button>
      <div class="fund-detail-header__breadcrumb">
        <span class="fund-detail-header__crumb">{{ t("myFunds.page.title", "My funds") }}</span>
        <span class="fund-detail-header__crumb-sep">/</span>
        <span class="fund-detail-header__crumb fund-detail-header__crumb--current">{{ card.code }}</span>
      </div>
    </div>

    <div class="fund-detail-header__primary">
      <div class="fund-detail-header__id-block">
        <h1 class="fund-detail-header__title">
          {{ card.name || card.short_name || card.code }}
        </h1>
        <div class="fund-detail-header__meta">
          <span class="fund-detail-header__code">{{ card.code }}</span>
          <span class="fund-detail-header__sep">·</span>
          <span class="fund-detail-header__slug">{{ card.short_name }}</span>
          <span class="fund-detail-header__sep">·</span>
          <span class="fund-detail-header__base-currency">{{ card.base_currency }}</span>
        </div>
        <div class="fund-detail-header__badges">
          <AppBadge size="sm" :variant="statusVariant" :dot="card.status === 'BREACH' || card.status === 'LOCKED'">
            {{ statusLabel }}
          </AppBadge>
          <AppBadge size="sm" :variant="card.role === 'MANAGER' ? 'info' : 'neutral'">
            {{ card.role === "MANAGER"
              ? t("myFunds.roles.manager", "Manager")
              : t("myFunds.roles.member", "Member") }}
          </AppBadge>
          <ComplianceBadge :compliance="card.compliance" />
        </div>
      </div>

      <dl class="fund-detail-header__metrics">
        <div>
          <dt>{{ t("myFunds.card.nav", "Net Asset Value") }}</dt>
          <dd class="fund-detail-header__metric-value">{{ navLabel }}</dd>
        </div>
        <div>
          <dt>{{ t("myFunds.card.aum", "AUM") }}</dt>
          <dd class="fund-detail-header__metric-value">{{ aumLabel }}</dd>
        </div>
        <div>
          <dt>{{ t("myFunds.card.unrealised", "unrealised P&L") }}</dt>
          <dd class="fund-detail-header__metric-value" :data-trend="trend">{{ unrealisedLabel }}</dd>
        </div>
        <div>
          <dt>{{ t("myFunds.card.cashBuffer", "Cash buffer") }}</dt>
          <dd class="fund-detail-header__metric-value">{{ cashBufferLabel }}</dd>
        </div>
      </dl>
    </div>

    <FundWorkflowBar
      :current-state="card.workflow.current_state"
      :available="card.workflow.available"
      :business-date="card.workflow.business_date"
    />

    <div class="fund-detail-header__updated">
      {{
        t(
          "myFunds.card.updated",
          { value: updatedLabel },
          `Updated ${updatedLabel}`,
        )
      }}
    </div>

    <nav class="fund-detail-header__tabs" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        role="tab"
        :aria-selected="activeTab === tab.key"
        :aria-disabled="tab.disabled"
        :title="tab.hint"
        :class="[
          'fund-detail-header__tab',
          { 'is-active': activeTab === tab.key, 'is-disabled': tab.disabled },
        ]"
        @click="selectTab(tab)"
      >
        <span>{{ tab.label }}</span>
        <span v-if="tab.badge" class="fund-detail-header__tab-badge">{{ tab.badge }}</span>
      </button>
    </nav>
  </header>
</template>

<style scoped>
.fund-detail-header {
  display: grid;
  gap: var(--space-3);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.fund-detail-header__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  font-size: 12px;
}

.fund-detail-header__back {
  font-family: inherit;
  font-size: 12px;
  font-weight: 500;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--state-info, #1f6feb);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.fund-detail-header__breadcrumb {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--text-tertiary, #6e7781);
  font-size: 11px;
}

.fund-detail-header__crumb--current {
  color: var(--text-primary, #1f2328);
  font-weight: 600;
}

.fund-detail-header__primary {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.fund-detail-header__id-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.fund-detail-header__title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary, #1f2328);
  line-height: 1.15;
}

.fund-detail-header__meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary, #57606a);
}

.fund-detail-header__code {
  font-weight: 600;
  color: var(--state-info, #1f6feb);
}

.fund-detail-header__slug,
.fund-detail-header__base-currency {
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, monospace);
}

.fund-detail-header__sep {
  color: var(--text-tertiary, #6e7781);
}

.fund-detail-header__badges {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.fund-detail-header__metrics {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: var(--space-4);
  margin: 0;
  align-self: flex-end;
}

.fund-detail-header__metrics > div {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.fund-detail-header__metrics dt {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  font-weight: 600;
  margin: 0;
}

.fund-detail-header__metrics dd {
  font-size: 18px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary, #1f2328);
  margin: 0;
  line-height: 1.1;
}

.fund-detail-header__metric-value[data-trend="up"] {
  color: var(--state-success, #1f883d);
}

.fund-detail-header__metric-value[data-trend="down"] {
  color: var(--state-danger, #cf222e);
}

.fund-detail-header__updated {
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
}

.fund-detail-header__tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  margin: 0;
  flex-wrap: wrap;
}

.fund-detail-header__tab {
  font-family: inherit;
  font-size: 13px;
  font-weight: 500;
  padding: 8px 14px;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-secondary, #57606a);
  border-bottom: 2px solid transparent;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-bottom: -1px;
}

.fund-detail-header__tab:hover:not(.is-disabled) {
  color: var(--text-primary, #1f2328);
  background: var(--bg-card-muted, #f6f8fa);
}

.fund-detail-header__tab.is-active {
  color: var(--text-primary, #1f2328);
  font-weight: 600;
  border-bottom-color: var(--state-info, #1f6feb);
}

.fund-detail-header__tab.is-disabled {
  color: var(--text-tertiary, #6e7781);
  cursor: not-allowed;
  opacity: 0.65;
}

.fund-detail-header__tab-badge {
  font-size: 10px;
  font-weight: 700;
  background: var(--state-danger, #cf222e);
  color: #ffffff;
  border-radius: 999px;
  padding: 1px 6px;
  min-width: 18px;
  text-align: center;
}
</style>
