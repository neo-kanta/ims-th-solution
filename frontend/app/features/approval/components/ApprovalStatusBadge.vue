<script setup lang="ts">
import { computed } from "vue";

import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import { requestStatusBadge, requestStatusLabel } from "../lib/approvalStatus";

const props = defineProps<{ status: string }>();

const badgeStatus = computed(() => requestStatusBadge(props.status));
const label = computed(() => requestStatusLabel(props.status));
</script>

<template>
  <AppStatusBadge :status="badgeStatus" :label="label" />
</template>
