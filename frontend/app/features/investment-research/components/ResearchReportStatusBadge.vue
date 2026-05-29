<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppBadge from "~/shared/ui/AppBadge.vue";
import type {
  ResearchRecommendation,
  ResearchReportStatus,
  ResearchReviewStatus,
} from "../types";
import {
  recommendationBadgeVariant,
  recommendationLabel,
  reportStatusBadgeVariant,
  reportStatusLabel,
  reviewStatusBadgeVariant,
  reviewStatusLabel,
} from "../lib/researchReportFormat";

interface Props {
  kind: "report" | "review" | "recommendation";
  value: ResearchReportStatus | ResearchReviewStatus | ResearchRecommendation;
  size?: "sm" | "md";
}

const props = withDefaults(defineProps<Props>(), { size: "sm" });
const { t } = useI18n();

const variant = computed(() => {
  if (props.kind === "report") {
    return reportStatusBadgeVariant(props.value as ResearchReportStatus);
  }
  if (props.kind === "review") {
    return reviewStatusBadgeVariant(props.value as ResearchReviewStatus);
  }
  return recommendationBadgeVariant(props.value as ResearchRecommendation);
});

const label = computed(() => {
  if (props.kind === "report") {
    return reportStatusLabel(props.value as ResearchReportStatus, t);
  }
  if (props.kind === "review") {
    return reviewStatusLabel(props.value as ResearchReviewStatus, t);
  }
  return recommendationLabel(props.value as ResearchRecommendation, t);
});

</script>

<template>
  <AppBadge :variant="variant" :size="size" dot>{{ label }}</AppBadge>
</template>
