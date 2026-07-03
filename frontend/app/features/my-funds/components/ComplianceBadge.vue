<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";

import type { FundComplianceSummary } from "../types";

interface Props {
  compliance: FundComplianceSummary;
}

const props = defineProps<Props>();
const { t } = useI18n();

const tone = computed<"danger" | "warn" | "info" | "muted">(() => {
  if (!props.compliance.open_count) return "muted";
  if (props.compliance.worst_severity === "BLOCK") return "danger";
  if (props.compliance.worst_severity === "WARN") return "warn";
  return "info";
});

const label = computed(() => {
  if (!props.compliance.open_count) {
    return t("myFunds.compliance.clear", "No open breaches");
  }
  return t(
    "myFunds.compliance.openCount",
    { count: props.compliance.open_count },
    `${props.compliance.open_count} open breach${props.compliance.open_count === 1 ? "" : "es"}`,
  );
});

const detail = computed(() => props.compliance.latest_message ?? "");
</script>

<template>
  <div class="compliance-badge" :data-tone="tone" :title="detail">
    <span class="compliance-badge__dot" />
    <span class="compliance-badge__label">{{ label }}</span>
    <span v-if="detail" class="compliance-badge__detail">— {{ detail }}</span>
  </div>
</template>

<style scoped>
.compliance-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary, #6e7781);
  max-width: 100%;
}

.compliance-badge__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-tertiary, #6e7781);
  flex-shrink: 0;
}

.compliance-badge[data-tone="danger"] .compliance-badge__dot {
  background: var(--state-danger, #cf222e);
}
.compliance-badge[data-tone="danger"] .compliance-badge__label {
  color: var(--state-danger, #cf222e);
  font-weight: 600;
}

.compliance-badge[data-tone="warn"] .compliance-badge__dot {
  background: var(--state-warning, #9a6700);
}
.compliance-badge[data-tone="warn"] .compliance-badge__label {
  color: var(--state-warning, #9a6700);
  font-weight: 600;
}

.compliance-badge[data-tone="info"] .compliance-badge__dot {
  background: var(--state-info, #1f6feb);
}
.compliance-badge[data-tone="info"] .compliance-badge__label {
  color: var(--state-info, #1f6feb);
}

.compliance-badge__detail {
  color: var(--text-tertiary, #6e7781);
  font-weight: 400;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}
</style>
