<script setup lang="ts">
import { computed, onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";

import { useComplianceUserDirectory } from "../composables/useComplianceUserDirectory";
import {
  deriveRuleStatus,
  formatEffectiveWindow,
  formatIsoDateTime,
} from "../lib/formatters";
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
const meta = computed(() => props.rule.type_metadata);
const ownerLabel = computed(() => users.labelFor(props.rule.createdBy));
</script>

<template>
  <AppCard class="detail-rail-card">
    <div class="rail-header">
      <h3 class="rail-title">{{ t("compliance.rules.table.category") }} &amp; {{ t("compliance.dashboard.tabs.settings") }}</h3>
    </div>
    <dl class="rail">
      <div class="rail__row">
        <dt>{{ t("compliance.detail.rail.status") }}</dt>
        <dd><ComplianceRuleStatusBadge :status="status" /></dd>
      </div>
      <div class="rail__row">
        <dt>{{ t("compliance.detail.rail.ruleTypeId") }}</dt>
        <dd><code class="rail__code">{{ rule.ruleTypeID }}</code></dd>
      </div>
      <div class="rail__row">
        <dt>{{ t("compliance.detail.rail.owner") }}</dt>
        <dd :title="rule.createdBy" class="rail__owner">{{ ownerLabel }}</dd>
      </div>
      <div class="rail__row">
        <dt>{{ t("compliance.detail.rail.version") }}</dt>
        <dd class="rail__version">v{{ rule.currentVersion }}</dd>
      </div>
      <div v-if="meta?.category" class="rail__row">
        <dt>{{ t("compliance.detail.rail.category") }}</dt>
        <dd class="rail__category">{{ meta.category }}</dd>
      </div>
      <div v-if="meta?.default_severity" class="rail__row">
        <dt>{{ t("compliance.detail.rail.defaultSeverity") }}</dt>
        <dd>
          <ComplianceSeverityBadge :severity="meta.default_severity" />
        </dd>
      </div>
      <div v-if="meta?.supported_timings?.length" class="rail__row rail__row--chips">
        <dt>{{ t("compliance.detail.rail.timings") }}</dt>
        <dd class="rail__chips-container">
          <span
            v-for="ti in meta.supported_timings"
            :key="ti"
            class="rail__chip"
          >
            {{ ti }}
          </span>
        </dd>
      </div>
      <div v-if="meta?.supported_scopes?.length" class="rail__row rail__row--chips">
        <dt>{{ t("compliance.detail.rail.scopes") }}</dt>
        <dd class="rail__chips-container">
          <span
            v-for="sc in meta.supported_scopes"
            :key="sc"
            class="rail__chip"
          >
            {{ sc }}
          </span>
        </dd>
      </div>
      <div v-if="meta" class="rail__row">
        <dt>{{ t("compliance.detail.rail.overridable") }}</dt>
        <dd class="rail__overridable">{{ meta.overridable ? "Yes" : "No" }}</dd>
      </div>
      <div class="rail__row rail__row--stacked">
        <dt>{{ t("compliance.detail.rail.effectiveWindow") }}</dt>
        <dd class="rail__window">{{ formatEffectiveWindow(rule.effectiveWindow) }}</dd>
      </div>
      <div class="rail__row">
        <dt>{{ t("compliance.detail.rail.createdAt") }}</dt>
        <dd class="rail__time">{{ formatIsoDateTime(rule.createdAt) }}</dd>
      </div>
      <div class="rail__row">
        <dt>{{ t("compliance.detail.rail.updatedAt") }}</dt>
        <dd class="rail__time">{{ formatIsoDateTime(rule.updatedAt) }}</dd>
      </div>
    </dl>
  </AppCard>
</template>

<style scoped>
.detail-rail-card {
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-subtle);
  transition: all var(--transition-base);
}

.detail-rail-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--border-strong);
}

.rail-header {
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-3);
  margin-bottom: var(--space-4);
}

.rail-title {
  margin: 0;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
}

.rail {
  display: grid;
  gap: var(--space-1);
  margin: 0;
}

.rail__row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-size-sm);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}

.rail__row:hover {
  background: var(--bg-card-hover);
}

.rail__row dt {
  color: var(--text-tertiary);
  font-weight: var(--font-weight-medium);
  font-size: var(--font-size-xs);
}

.rail__row dd {
  margin: 0;
  text-align: right;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
  font-weight: var(--font-weight-semibold);
}

.rail__row--chips {
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-2);
}

.rail__row--chips dt {
  align-self: flex-start;
}

.rail__row--chips dd {
  align-self: flex-end;
  width: 100%;
}

.rail__row--stacked {
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-1);
}

.rail__row--stacked dt {
  align-self: flex-start;
}

.rail__row--stacked dd {
  width: 100%;
  text-align: left;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  background: var(--bg-card-muted);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
  box-sizing: border-box;
}

.rail__code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  background: var(--bg-card-muted);
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-xs);
  border: 1px solid var(--border-subtle);
}

.rail__owner {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rail__chips-container {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
  justify-content: flex-end;
}

.rail__chip {
  display: inline-flex;
  align-items: center;
  padding: 0 var(--space-2);
  height: 20px;
  border-radius: var(--radius-xs);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-family: var(--font-family-mono);
  font-size: 10px;
  border: 1px solid var(--border-subtle);
  text-transform: uppercase;
}

.rail__time {
  font-family: var(--font-family-mono);
  font-size: 11px;
  color: var(--text-secondary);
}
</style>
