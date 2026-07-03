<script setup lang="ts">
import { computed } from "vue";
import AppBadge from "./AppBadge.vue";

type StatusType =
  | "active"
  | "inactive"
  | "pending"
  | "approved"
  | "rejected"
  | "expired"
  | "locked"
  | "warning"
  | "danger"
  | "success"
  | "neutral";

interface Props {
  status: StatusType | string;
  label?: string;
  dot?: boolean;
  size?: "sm" | "md";
}

const props = withDefaults(defineProps<Props>(), {
  label: "",
  dot: true,
  size: "md",
});

const statusMap: Record<StatusType, { badgeVariant: string; label: string }> = {
  active: { badgeVariant: "success", label: "Active" },
  inactive: { badgeVariant: "neutral", label: "Inactive" },
  pending: { badgeVariant: "warning", label: "Pending" },
  approved: { badgeVariant: "success", label: "Approved" },
  rejected: { badgeVariant: "error", label: "Rejected" },
  expired: { badgeVariant: "locked", label: "Expired" },
  locked: { badgeVariant: "locked", label: "Locked" },
  warning: { badgeVariant: "warning", label: "Warning" },
  danger: { badgeVariant: "error", label: "Danger" },
  success: { badgeVariant: "success", label: "Success" },
  neutral: { badgeVariant: "neutral", label: "Neutral" },
};

const resolvedStatus = computed(() => {
  const norm = String(props.status).toLowerCase() as StatusType;
  return statusMap[norm] || { badgeVariant: "neutral", label: String(props.status) };
});

const displayLabel = computed(() => {
  return props.label || resolvedStatus.value.label;
});
</script>

<template>
  <AppBadge
    :variant="resolvedStatus.badgeVariant as any"
    :dot="dot"
    :size="size"
  >
    {{ displayLabel }}
  </AppBadge>
</template>
