<script setup lang="ts">
import AppButton from "./AppButton.vue";

interface Props {
  showSave?: boolean;
  showCancel?: boolean;
  saveLabel?: string;
  cancelLabel?: string;
  saveLoading?: boolean;
  saveDisabled?: boolean;
  cancelDisabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showSave: true,
  showCancel: false,
  saveLabel: "Save changes",
  cancelLabel: "Cancel",
  saveLoading: false,
  saveDisabled: false,
  cancelDisabled: false,
});

const emit = defineEmits<{
  save: [];
  cancel: [];
}>();
</script>

<template>
  <div class="settings-actions" v-if="showSave || showCancel">
    <AppButton
      v-if="showCancel"
      type="button"
      variant="secondary"
      size="sm"
      :disabled="cancelDisabled || saveLoading"
      @click="emit('cancel')"
    >
      {{ cancelLabel }}
    </AppButton>
    <AppButton
      v-if="showSave"
      type="submit"
      variant="primary"
      size="sm"
      :loading="saveLoading"
      :disabled="saveDisabled"
      @click="emit('save')"
    >
      {{ saveLabel }}
    </AppButton>
  </div>
</template>

<style scoped>
.settings-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-5);
  padding-top: var(--space-4);
  border-top: 1px solid var(--border-subtle);
  width: 100%;
}
</style>
