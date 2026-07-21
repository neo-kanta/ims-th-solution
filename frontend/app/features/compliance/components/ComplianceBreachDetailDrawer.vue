<script setup lang="ts">
/**
 * Breach detail drawer — the "why" behind a row: rule, portfolio, message,
 * risk, status, dates, structured evidence, and the check-group records that
 * produced the breach (`GET /compliance/checks/{groupID}`). Opened by
 * clicking a row's Breach cell or "View details".
 */
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import { useBangkokFormatter } from "~/shared/composables/useBangkokFormatter";
import AppButton from "~/shared/ui/AppButton.vue";
import AppErrorState from "~/shared/ui/AppErrorState.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";

import { useComplianceCheckGroup } from "../composables/useComplianceCheckGroup";
import { useCompliancePortfolioDirectory } from "../composables/useCompliancePortfolioDirectory";
import { canOverrideBreach } from "../lib/breachActions";
import { formatIsoDate } from "../lib/formatters";
import { ruleLabel } from "../lib/ruleTypeCatalog";
import type { ComplianceBreach, ComplianceBreachStatus } from "../types";
import ComplianceSeverityBadge from "./ComplianceSeverityBadge.vue";
import ComplianceVerdictBadge from "./ComplianceVerdictBadge.vue";

const props = defineProps<{
  breach: ComplianceBreach | null;
  canOverride: boolean;
}>();

const emit = defineEmits<{
  close: [];
  override: [breach: ComplianceBreach];
}>();

const { t } = useI18n();
const { formatDateTime } = useBangkokFormatter();
const portfolios = useCompliancePortfolioDirectory();
const checkGroup = useComplianceCheckGroup();

const closeButtonRef = ref<HTMLElement | null>(null);
let previouslyFocused: HTMLElement | null = null;

const isOpen = computed(() => props.breach !== null);

watch(
  () => props.breach,
  (breach) => {
    if (!breach) {
      checkGroup.clear();
      return;
    }
    void checkGroup.fetch(breach.checkGroupID);
  },
);

watch(isOpen, (open) => {
  if (!import.meta.client) return;
  if (open) {
    previouslyFocused = document.activeElement as HTMLElement | null;
    void nextTick(() => closeButtonRef.value?.focus());
  } else if (previouslyFocused && document.body.contains(previouslyFocused)) {
    previouslyFocused.focus();
    previouslyFocused = null;
  }
});

onBeforeUnmount(() => {
  previouslyFocused = null;
});

function portfolioLabel(id: string): string {
  const hit = portfolios.byId.value.get(id);
  if (!hit) return t("common.notAvailable");
  return hit.name ? `${hit.code} — ${hit.name}` : hit.code;
}

function statusTone(status: ComplianceBreachStatus): "active" | "warning" | "success" {
  switch (status) {
    case "OPEN":
      return "active";
    case "OVERRIDDEN":
      return "warning";
    case "RESOLVED":
      return "success";
    default:
      return "active";
  }
}

function statusLabel(status: ComplianceBreachStatus): string {
  switch (status) {
    case "OPEN":
      return t("compliance.postTrade.toolbar.statusOpen");
    case "OVERRIDDEN":
      return t("compliance.postTrade.toolbar.statusOverridden");
    case "RESOLVED":
      return t("compliance.postTrade.toolbar.statusResolved");
    default:
      return status;
  }
}

const canOverrideThisBreach = computed(
  () => props.breach !== null && canOverrideBreach(props.breach, props.canOverride),
);

const thresholdLine = computed(() => {
  const threshold = props.breach?.evidence?.threshold_breached;
  if (!threshold) return null;
  const parts: string[] = [];
  if (threshold.metric_name) parts.push(threshold.metric_name);
  if (threshold.actual) parts.push(`${t("compliance.postTrade.drawer.thresholdActual")}: ${threshold.actual}${threshold.unit ?? ""}`);
  if (threshold.operator) parts.push(threshold.operator);
  if (threshold.limit) parts.push(`${t("compliance.postTrade.drawer.thresholdLimit")}: ${threshold.limit}${threshold.unit ?? ""}`);
  return parts.length > 0 ? parts.join(" ") : null;
});

function metricEntries(record?: Record<string, string>): [string, string][] {
  return record ? Object.entries(record) : [];
}

function retryCheckGroup() {
  if (props.breach) void checkGroup.fetch(props.breach.checkGroupID);
}
</script>

<template>
  <Teleport to="body">
    <Transition name="breach-drawer">
      <div v-if="isOpen && breach" class="breach-drawer-overlay" @click.self="emit('close')">
        <aside
          class="breach-drawer"
          role="dialog"
          aria-modal="true"
          :aria-label="t('compliance.postTrade.drawer.ariaLabel')"
          @keydown.esc="emit('close')"
        >
          <header class="breach-drawer__header">
            <div class="breach-drawer__title-row">
              <div>
                <p class="breach-drawer__eyebrow">{{ t("compliance.postTrade.drawer.title") }}</p>
                <h2 class="breach-drawer__title">{{ ruleLabel(breach.ruleTypeID, t) }}</h2>
              </div>
              <AppStatusBadge :status="statusTone(breach.status)" :label="statusLabel(breach.status)" size="sm" />
            </div>
            <button
              ref="closeButtonRef"
              type="button"
              class="breach-drawer__close"
              :aria-label="t('compliance.postTrade.drawer.close')"
              @click="emit('close')"
            >
              ×
            </button>
          </header>

          <div class="breach-drawer__body">
            <section class="breach-drawer__section">
              <h3 class="breach-drawer__section-title">{{ t("compliance.postTrade.drawer.sectionSummary") }}</h3>
              <dl class="breach-drawer__facts">
                <div class="breach-drawer__fact">
                  <dt>{{ t("compliance.postTrade.drawer.portfolio") }}</dt>
                  <dd>{{ portfolioLabel(breach.portfolioID) }}</dd>
                </div>
                <div class="breach-drawer__fact">
                  <dt>{{ t("compliance.postTrade.drawer.message") }}</dt>
                  <dd>{{ breach.message || t("compliance.postTrade.drawer.messageEmpty") }}</dd>
                </div>
                <div class="breach-drawer__fact">
                  <dt>{{ t("compliance.postTrade.drawer.ruleTypeId") }}</dt>
                  <dd><code>{{ breach.ruleTypeID }}</code></dd>
                </div>
              </dl>
            </section>

            <section class="breach-drawer__section">
              <h3 class="breach-drawer__section-title">{{ t("compliance.postTrade.drawer.sectionRisk") }}</h3>
              <div class="breach-drawer__risk">
                <div>
                  <span class="breach-drawer__risk-label">{{ t("compliance.postTrade.drawer.severity") }}</span>
                  <ComplianceSeverityBadge :severity="breach.severity" />
                </div>
                <div>
                  <span class="breach-drawer__risk-label">{{ t("compliance.postTrade.drawer.verdict") }}</span>
                  <ComplianceVerdictBadge :verdict="breach.verdict" />
                </div>
              </div>
            </section>

            <section class="breach-drawer__section">
              <h3 class="breach-drawer__section-title">{{ t("compliance.postTrade.drawer.sectionDates") }}</h3>
              <dl class="breach-drawer__facts">
                <div class="breach-drawer__fact">
                  <dt>{{ t("compliance.postTrade.drawer.businessDate") }}</dt>
                  <dd>{{ formatIsoDate(breach.businessDate) }}</dd>
                </div>
                <div class="breach-drawer__fact">
                  <dt>{{ t("compliance.postTrade.drawer.createdAt") }}</dt>
                  <dd>{{ formatDateTime(breach.createdAt) }}</dd>
                </div>
                <div v-if="breach.resolvedAt" class="breach-drawer__fact">
                  <dt>{{ t("compliance.postTrade.drawer.resolvedAt") }}</dt>
                  <dd>{{ formatDateTime(breach.resolvedAt) }}</dd>
                </div>
                <div v-if="breach.resolvedBy" class="breach-drawer__fact">
                  <dt>&nbsp;</dt>
                  <dd class="breach-drawer__muted">{{ t("compliance.postTrade.drawer.resolvedByHidden") }}</dd>
                </div>
              </dl>
            </section>

            <section class="breach-drawer__section">
              <h3 class="breach-drawer__section-title">{{ t("compliance.postTrade.drawer.sectionEvidence") }}</h3>
              <p
                v-if="!breach.evidence || (!breach.evidence.metrics && !breach.evidence.references && !thresholdLine)"
                class="breach-drawer__muted"
              >
                {{ t("compliance.postTrade.drawer.noEvidence") }}
              </p>
              <div v-else class="breach-drawer__evidence">
                <p v-if="thresholdLine" class="breach-drawer__threshold">
                  <strong>{{ t("compliance.postTrade.drawer.evidenceThreshold") }}:</strong>
                  {{ thresholdLine }}
                </p>
                <div v-if="breach.evidence?.metrics" class="breach-drawer__evidence-group">
                  <h4>{{ t("compliance.postTrade.drawer.evidenceMetrics") }}</h4>
                  <dl class="breach-drawer__facts">
                    <div
                      v-for="[key, value] in metricEntries(breach.evidence.metrics)"
                      :key="key"
                      class="breach-drawer__fact"
                    >
                      <dt>{{ key }}</dt>
                      <dd>{{ value }}</dd>
                    </div>
                  </dl>
                </div>
                <div v-if="breach.evidence?.references" class="breach-drawer__evidence-group">
                  <h4>{{ t("compliance.postTrade.drawer.evidenceReferences") }}</h4>
                  <dl class="breach-drawer__facts">
                    <div
                      v-for="[key, value] in metricEntries(breach.evidence.references)"
                      :key="key"
                      class="breach-drawer__fact"
                    >
                      <dt>{{ key }}</dt>
                      <dd>{{ value }}</dd>
                    </div>
                  </dl>
                </div>
              </div>
            </section>

            <section class="breach-drawer__section">
              <h3 class="breach-drawer__section-title">
                {{ t("compliance.postTrade.drawer.sectionRecords", { count: checkGroup.result.value?.records.length ?? 0 }) }}
              </h3>
              <AppLoadingState v-if="checkGroup.loading.value" :message="t('compliance.postTrade.drawer.loading')" />
              <AppErrorState
                v-else-if="checkGroup.error.value"
                :title="t('compliance.postTrade.drawer.errorTitle')"
                :message="checkGroup.error.value"
                retry
                @retry="retryCheckGroup"
              />
              <p
                v-else-if="!checkGroup.result.value || checkGroup.result.value.records.length === 0"
                class="breach-drawer__muted"
              >
                {{ t("compliance.postTrade.drawer.noRecords") }}
              </p>
              <div v-else class="breach-drawer__records">
                <table class="breach-drawer__records-table">
                  <thead>
                    <tr>
                      <th>{{ t("compliance.postTrade.drawer.rule") }}</th>
                      <th>{{ t("compliance.postTrade.drawer.recordVerdict") }}</th>
                      <th>{{ t("compliance.postTrade.drawer.recordSeverity") }}</th>
                      <th>{{ t("compliance.postTrade.drawer.message") }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="record in checkGroup.result.value.records" :key="record.id">
                      <td>{{ ruleLabel(record.ruleTypeID, t) }}</td>
                      <td><ComplianceVerdictBadge :verdict="record.finalVerdict" size="sm" /></td>
                      <td><ComplianceSeverityBadge :severity="record.effectiveSeverity" size="sm" /></td>
                      <td class="breach-drawer__records-message">{{ record.message || "—" }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>
          </div>

          <footer class="breach-drawer__footer">
            <AppButton
              v-if="canOverrideThisBreach"
              variant="primary"
              size="sm"
              @click="emit('override', breach)"
            >
              {{ t("compliance.postTrade.drawer.override") }}
            </AppButton>
            <AppButton variant="ghost" size="sm" @click="emit('close')">
              {{ t("compliance.postTrade.drawer.close") }}
            </AppButton>
          </footer>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.breach-drawer-overlay {
  position: fixed;
  inset: 0;
  z-index: 900;
  display: flex;
  justify-content: flex-end;
  background: rgba(15, 23, 42, 0.42);
}

.breach-drawer {
  display: flex;
  flex-direction: column;
  width: min(560px, 92vw);
  height: 100%;
  overflow: hidden;
  border-left: 1px solid var(--border-subtle);
  background: var(--bg-card);
  box-shadow: -18px 0 48px rgba(15, 23, 42, 0.16);
}

.breach-drawer__header {
  display: flex;
  flex-shrink: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
}

.breach-drawer__title-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.breach-drawer__eyebrow {
  margin: 0 0 2px;
  color: var(--text-tertiary);
  font-size: 10px;
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.breach-drawer__title {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-lg);
  overflow-wrap: anywhere;
}

.breach-drawer__close {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
}

.breach-drawer__close:focus-visible {
  outline: 2px solid var(--border-focus);
  outline-offset: 2px;
}

.breach-drawer__body {
  display: grid;
  flex: 1;
  gap: var(--space-6);
  overflow-y: auto;
  padding: var(--space-5);
  align-content: start;
}

.breach-drawer__section-title {
  margin: 0 0 var(--space-3);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.breach-drawer__facts {
  margin: 0;
  display: grid;
  gap: var(--space-2);
}

.breach-drawer__fact {
  display: grid;
  grid-template-columns: 9rem 1fr;
  gap: var(--space-2);
  font-size: var(--font-size-sm);
}

.breach-drawer__fact dt {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.breach-drawer__fact dd {
  margin: 0;
  color: var(--text-primary);
  overflow-wrap: anywhere;
}

.breach-drawer__risk {
  display: flex;
  gap: var(--space-6);
}

.breach-drawer__risk-label {
  display: block;
  margin-bottom: var(--space-1);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.breach-drawer__muted {
  margin: 0;
  color: var(--text-tertiary);
  font-size: var(--font-size-sm);
}

.breach-drawer__evidence {
  display: grid;
  gap: var(--space-4);
}

.breach-drawer__threshold {
  margin: 0;
  padding: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
  font-size: var(--font-size-sm);
}

.breach-drawer__evidence-group h4 {
  margin: 0 0 var(--space-2);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.breach-drawer__records {
  overflow-x: auto;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.breach-drawer__records-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-xs);
}

.breach-drawer__records-table th,
.breach-drawer__records-table td {
  padding: var(--space-2) var(--space-3);
  border-bottom: 1px solid var(--border-subtle);
  text-align: left;
  vertical-align: top;
}

.breach-drawer__records-table th {
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-weight: var(--font-weight-semibold);
  white-space: nowrap;
}

.breach-drawer__records-table tbody tr:last-child td {
  border-bottom: 0;
}

.breach-drawer__records-message {
  max-width: 16rem;
}

.breach-drawer__footer {
  display: flex;
  flex-shrink: 0;
  justify-content: flex-end;
  gap: var(--space-2);
  padding: var(--space-4) var(--space-5);
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-card-muted);
}

.breach-drawer-enter-active,
.breach-drawer-leave-active {
  transition: opacity 0.18s ease;
}

.breach-drawer-enter-active .breach-drawer,
.breach-drawer-leave-active .breach-drawer {
  transition: transform 0.22s ease;
}

.breach-drawer-enter-from,
.breach-drawer-leave-to {
  opacity: 0;
}

.breach-drawer-enter-from .breach-drawer,
.breach-drawer-leave-to .breach-drawer {
  transform: translateX(100%);
}

@media (prefers-reduced-motion: reduce) {
  .breach-drawer-enter-active,
  .breach-drawer-leave-active,
  .breach-drawer-enter-active .breach-drawer,
  .breach-drawer-leave-active .breach-drawer {
    transition: none;
  }
}

@media (max-width: 640px) {
  .breach-drawer {
    width: 100vw;
  }

  .breach-drawer__header,
  .breach-drawer__body,
  .breach-drawer__footer {
    padding-right: var(--space-4);
    padding-left: var(--space-4);
  }

  .breach-drawer__fact {
    grid-template-columns: 1fr;
  }
}
</style>
