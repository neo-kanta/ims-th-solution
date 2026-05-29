<script setup lang="ts">
/**
 * Confirmation dialog wrapper for workflow operations.
 *
 * Maps the wire action tone (`primary` / `warning` / `danger`) to the
 * AppConfirmDialog tone vocabulary (`neutral` / `warning` / `danger`).
 * While `executing=true` the dialog suppresses cancel + Escape (delegated
 * to AppConfirmDialog) so the operator cannot abort a fired transition.
 */
import { computed } from "vue";

import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";

import type { WorkflowDialogTone } from "../types";

type AppTone = "danger" | "warning" | "neutral";

const props = defineProps<{
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  cancelLabel?: string;
  executing: boolean;
  tone: WorkflowDialogTone;
}>();

const emit = defineEmits<{
  cancel: [];
  confirm: [];
}>();

const mappedTone = computed<AppTone>(() => {
  if (props.tone === "danger") return "danger";
  if (props.tone === "warning") return "warning";
  return "neutral";
});

defineExpose({ mappedTone });
</script>

<template>
  <AppConfirmDialog
    :open="open"
    :title="title"
    :description="description"
    :confirm-label="confirmLabel"
    :cancel-label="cancelLabel ?? 'Cancel'"
    :tone="mappedTone"
    :loading="executing"
    @cancel="emit('cancel')"
    @confirm="emit('confirm')"
  />
</template>
