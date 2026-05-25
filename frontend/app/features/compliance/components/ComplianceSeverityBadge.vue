<script setup lang="ts">
import { computed } from "vue";

import AppBadge from "~/shared/ui/AppBadge.vue";

import { severityLabel, severityTone } from "../lib/formatters";
import type { ComplianceBackendSeverity } from "../types";

interface Props {
  severity: ComplianceBackendSeverity;
  size?: "sm" | "md";
}

const props = withDefaults(defineProps<Props>(), {
  size: "sm",
});

const variant = computed(() => {
  const tone = severityTone(props.severity);
  if (tone === "success") return "success" as const;
  if (tone === "warning") return "warning" as const;
  if (tone === "error") return "error" as const;
  if (tone === "info") return "info" as const;
  return "neutral" as const;
});

const label = computed(() => severityLabel(props.severity));
</script>

<template>
  <AppBadge :variant="variant" :size="size" dot>
    <span aria-hidden="true">{{ label }}</span>
    <span class="sr-only">Severity: {{ label }}</span>
  </AppBadge>
</template>

<style scoped>
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
