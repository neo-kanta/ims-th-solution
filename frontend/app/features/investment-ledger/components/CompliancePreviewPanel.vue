<script setup lang="ts">
import { computed } from "vue";

import { severityTone, shortenId, verdictLabel, verdictTone } from "../lib/ledgerFormat";
import type { ApiCompliancePreview } from "../services/investmentLedgerApi";

interface Props {
  compliance: ApiCompliancePreview | null | undefined;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const verdict = computed(() => props.compliance?.verdict ?? null);
const tone = computed(() => verdictTone(verdict.value));
const breaches = computed(() => props.compliance?.breaches ?? []);
const blocking = computed(() =>
  breaches.value.filter((b) => (b.verdict ?? "").toUpperCase() === "BLOCK"),
);
const warnings = computed(() =>
  breaches.value.filter((b) => (b.verdict ?? "").toUpperCase() === "WARN"),
);
</script>

<template>
  <div class="compliance-preview">
    <div class="compliance-preview__header">
      <h3 class="compliance-preview__title">Compliance verdict</h3>
      <span
        v-if="verdict"
        class="compliance-preview__verdict"
        :class="`compliance-preview__verdict--${tone}`"
      >
        {{ verdictLabel(verdict) }}
      </span>
    </div>

    <div v-if="loading" class="compliance-preview__placeholder">
      Running pre-trade checks…
    </div>

    <div v-else-if="!compliance" class="compliance-preview__placeholder">
      Run a simulation to evaluate pre-trade rules.
    </div>

    <template v-else>
      <div class="compliance-preview__summary">
        <div class="compliance-preview__stat">
          <span class="compliance-preview__stat-label">Rules evaluated</span>
          <span class="compliance-preview__stat-value">
            {{ compliance.rules_evaluated ?? 0 }}
          </span>
        </div>
        <div class="compliance-preview__stat">
          <span class="compliance-preview__stat-label">Blocking</span>
          <span
            class="compliance-preview__stat-value"
            :class="blocking.length > 0 ? 'compliance-preview__stat-value--negative' : ''"
          >
            {{ blocking.length }}
          </span>
        </div>
        <div class="compliance-preview__stat">
          <span class="compliance-preview__stat-label">Warnings</span>
          <span
            class="compliance-preview__stat-value"
            :class="warnings.length > 0 ? 'compliance-preview__stat-value--warn' : ''"
          >
            {{ warnings.length }}
          </span>
        </div>
        <div class="compliance-preview__stat">
          <span class="compliance-preview__stat-label">Check group</span>
          <span
            class="compliance-preview__stat-value compliance-preview__stat-value--mono"
            :title="compliance.check_group_id ?? ''"
          >
            {{ shortenId(compliance.check_group_id) || "—" }}
          </span>
        </div>
      </div>

      <div v-if="breaches.length > 0" class="compliance-preview__list" role="list">
        <article
          v-for="(breach, index) in breaches"
          :key="breach.breach_id ?? `${breach.rule_type_id ?? 'rule'}-${index}`"
          class="compliance-preview__item"
          :class="`compliance-preview__item--${severityTone(breach.severity)}`"
          role="listitem"
        >
          <header class="compliance-preview__item-head">
            <span class="compliance-preview__item-rule">
              {{ breach.rule_type_id ?? "Unknown rule" }}
            </span>
            <span class="compliance-preview__item-tags">
              <span
                v-if="breach.verdict"
                class="compliance-preview__tag"
                :class="`compliance-preview__tag--${verdictTone(breach.verdict)}`"
              >
                {{ verdictLabel(breach.verdict) }}
              </span>
              <span
                v-if="breach.severity"
                class="compliance-preview__tag"
                :class="`compliance-preview__tag--${severityTone(breach.severity)}`"
              >
                {{ breach.severity }}
              </span>
              <span
                v-if="breach.overridable"
                class="compliance-preview__tag compliance-preview__tag--neutral"
                title="A user with the appropriate override permission may force-post."
              >
                Overridable
              </span>
            </span>
          </header>
          <p class="compliance-preview__item-body">
            {{ breach.message || "Rule reported a breach without a message." }}
          </p>
        </article>
      </div>

      <p v-else class="compliance-preview__clean">
        No breaches reported by pre-trade rules.
      </p>
    </template>
  </div>
</template>

<style scoped>
.compliance-preview {
  display: grid;
  gap: var(--space-3);
}

.compliance-preview__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.compliance-preview__title {
  margin: 0;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
}

.compliance-preview__verdict {
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  border: 1px solid transparent;
}

.compliance-preview__verdict--success {
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
  border-color: var(--alert-success-border);
}

.compliance-preview__verdict--warning {
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  border-color: var(--alert-warning-border);
}

.compliance-preview__verdict--error {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border-color: var(--alert-danger-border);
}

.compliance-preview__verdict--neutral {
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  border-color: var(--border-subtle);
}

.compliance-preview__placeholder {
  padding: var(--space-3);
  text-align: center;
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
  border: 1px dashed var(--border-subtle);
  border-radius: var(--radius-md);
}

.compliance-preview__summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-2);
}

.compliance-preview__stat {
  padding: var(--space-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card-muted);
  display: grid;
  gap: 2px;
}

.compliance-preview__stat-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
}

.compliance-preview__stat-value {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  font-variant-numeric: tabular-nums;
}

.compliance-preview__stat-value--mono {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
}

.compliance-preview__stat-value--negative {
  color: var(--state-danger, #dc2626);
}

.compliance-preview__stat-value--warn {
  color: var(--state-warning, #b45309);
}

.compliance-preview__list {
  display: grid;
  gap: var(--space-2);
  max-height: 220px;
  overflow-y: auto;
}

.compliance-preview__item {
  padding: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  display: grid;
  gap: var(--space-2);
}

.compliance-preview__item--error {
  border-left: 3px solid var(--state-danger, #dc2626);
}

.compliance-preview__item--warning {
  border-left: 3px solid var(--state-warning, #b45309);
}

.compliance-preview__item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.compliance-preview__item-rule {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.compliance-preview__item-tags {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.compliance-preview__tag {
  display: inline-flex;
  align-items: center;
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 8px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  font-weight: var(--font-weight-semibold);
}

.compliance-preview__tag--success {
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
}

.compliance-preview__tag--warning {
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
}

.compliance-preview__tag--error {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
}

.compliance-preview__tag--neutral {
  background: var(--bg-card-muted);
  color: var(--text-secondary);
}

.compliance-preview__item-body {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  line-height: 1.5;
}

.compliance-preview__clean {
  margin: 0;
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
  border: 1px solid var(--alert-success-border);
  border-radius: var(--radius-sm);
}

@media (max-width: 720px) {
  .compliance-preview__summary {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
