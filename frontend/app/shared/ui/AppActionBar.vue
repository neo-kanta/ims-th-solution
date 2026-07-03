<script setup lang="ts">
import { computed } from "vue";
import AppButton from "./AppButton.vue";

type ActionType = "save" | "cancel" | "submit" | "review" | "cancel-review" | "delete";

interface Props {
  actions?: ActionType[];
  saveLabel?: string;
  cancelLabel?: string;
  submitLabel?: string;
  reviewLabel?: string;
  cancelReviewLabel?: string;
  deleteLabel?: string;
  loadingAction?: ActionType | null;
  disabled?: boolean;
  sticky?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  actions: () => ["save", "cancel"],
  saveLabel: "Save Changes",
  cancelLabel: "Cancel",
  submitLabel: "Submit",
  reviewLabel: "Request Review",
  cancelReviewLabel: "Cancel Review",
  deleteLabel: "Delete",
  loadingAction: null,
  disabled: false,
  sticky: false,
});

const emit = defineEmits<{
  save: [];
  cancel: [];
  submit: [];
  review: [];
  "cancel-review": [];
  delete: [];
}>();

const hasAction = (type: ActionType) => props.actions.includes(type);

const primaryActions = computed(() => {
  const list: { type: ActionType; label: string; variant: "primary" | "secondary" | "danger" | "success" }[] = [];
  
  if (hasAction("save")) {
    list.push({ type: "save", label: props.saveLabel, variant: "primary" });
  }
  if (hasAction("submit")) {
    list.push({ type: "submit", label: props.submitLabel, variant: "primary" });
  }
  if (hasAction("review")) {
    list.push({ type: "review", label: props.reviewLabel, variant: "primary" });
  }
  if (hasAction("cancel-review")) {
    list.push({ type: "cancel-review", label: props.cancelReviewLabel, variant: "secondary" });
  }
  
  return list;
});

const secondaryActions = computed(() => {
  const list: { type: ActionType; label: string; variant: "secondary" | "danger" }[] = [];
  
  if (hasAction("cancel")) {
    list.push({ type: "cancel", label: props.cancelLabel, variant: "secondary" });
  }
  if (hasAction("delete")) {
    list.push({ type: "delete", label: props.deleteLabel, variant: "danger" });
  }
  
  return list;
});
</script>

<template>
  <div class="app-action-bar" :class="{ 'app-action-bar--sticky': sticky }">
    <div class="app-action-bar__container">
      <!-- Left aligned secondary actions -->
      <div class="app-action-bar__left">
        <AppButton
          v-for="act in secondaryActions"
          :key="act.type"
          :variant="act.variant"
          size="sm"
          :disabled="disabled || (loadingAction !== null && loadingAction !== act.type)"
          :loading="loadingAction === act.type"
          @click="emit(act.type as any)"
        >
          {{ act.label }}
        </AppButton>
      </div>

      <!-- Right aligned primary actions -->
      <div class="app-action-bar__right">
        <AppButton
          v-for="act in primaryActions"
          :key="act.type"
          :variant="act.variant"
          size="sm"
          :disabled="disabled || (loadingAction !== null && loadingAction !== act.type)"
          :loading="loadingAction === act.type"
          @click="emit(act.type as any)"
        >
          {{ act.label }}
        </AppButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-action-bar {
  width: 100%;
  border-top: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card-muted, #f6f8fa);
  padding: var(--space-4, 16px) var(--space-5, 20px);
}

.app-action-bar--sticky {
  position: sticky;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.05);
}

.app-action-bar__container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4, 16px);
  max-width: 1280px;
  margin: 0 auto;
  width: 100%;
}

.app-action-bar__left,
.app-action-bar__right {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  flex-wrap: wrap;
}
</style>
