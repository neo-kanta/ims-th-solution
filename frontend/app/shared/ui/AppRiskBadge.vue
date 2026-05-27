<script setup lang="ts">
import { computed } from "vue";
import AppBadge from "./AppBadge.vue";

type RiskType = "pass" | "fail" | "warning" | "blocked" | "stale" | "unknown";

interface Props {
  risk: RiskType | string;
  label?: string;
  dot?: boolean;
  size?: "sm" | "md";
}

const props = withDefaults(defineProps<Props>(), {
  label: "",
  dot: true,
  size: "md",
});

const riskMap: Record<RiskType, { badgeVariant: string; label: string }> = {
  pass: { badgeVariant: "success", label: "Pass" },
  fail: { badgeVariant: "error", label: "Fail" },
  warning: { badgeVariant: "warning", label: "Warning" },
  blocked: { badgeVariant: "error", label: "Blocked" },
  stale: { badgeVariant: "locked", label: "Stale" },
  unknown: { badgeVariant: "neutral", label: "Unknown" },
};

const resolvedRisk = computed(() => {
  const norm = String(props.risk).toLowerCase() as RiskType;
  return riskMap[norm] || { badgeVariant: "neutral", label: String(props.risk) };
});

const displayLabel = computed(() => {
  return props.label || resolvedRisk.value.label;
});
</script>

<template>
  <AppBadge
    :variant="resolvedRisk.badgeVariant as any"
    :dot="dot"
    :size="size"
  >
    {{ displayLabel }}
  </AppBadge>
</template>
