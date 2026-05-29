<script setup lang="ts">
import { computed } from "vue";

import IMSAuditTimeline from "~/shared/ui/IMSAuditTimeline.vue";
import { toTimelineEvents } from "../lib/approvalMappers";
import type { ApprovalEvent } from "../types";

const props = defineProps<{ events: ApprovalEvent[] }>();

const vmEvents = computed(() => toTimelineEvents(props.events ?? []));
</script>

<template>
  <IMSAuditTimeline :events="vmEvents" />
</template>
