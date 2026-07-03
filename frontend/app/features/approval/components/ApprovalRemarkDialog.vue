<script setup lang="ts">
import { ref, watch, nextTick } from "vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppTextarea from "~/shared/ui/AppTextarea.vue";
import AppFormField from "~/shared/ui/AppFormField.vue";

const props = withDefaults(
  defineProps<{
    open: boolean;
    title: string;
    description: string;
    placeholder?: string;
    confirmLabel?: string;
    loading?: boolean;
  }>(),
  {
    placeholder: "Add an optional comment…",
    confirmLabel: "Confirm",
    loading: false,
  }
);

const emit = defineEmits<{
  cancel: [];
  confirm: [comment: string];
}>();

const remark = ref("");
const textareaRef = ref<any>(null);

watch(
  () => props.open,
  (open) => {
    if (open) {
      remark.value = "";
      void nextTick(() => {
        const textarea = textareaRef.value?.$el?.querySelector("textarea") || textareaRef.value?.$el || textareaRef.value;
        textarea?.focus();
      });
    }
  }
);

function onConfirm() {
  emit("confirm", remark.value.trim());
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape" && !props.loading) {
    emit("cancel");
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-backdrop app-remark-modal"
      role="presentation"
      @click.self="!loading && emit('cancel')"
      @keydown="handleKeydown"
    >
      <section
        class="modal app-remark-modal__dialog animate-fade-in"
        role="dialog"
        aria-modal="true"
        tabindex="-1"
      >
        <div class="modal-header app-remark-modal__header">
          <h2 class="app-remark-modal__title">{{ title }}</h2>
          <p class="app-remark-modal__desc">{{ description }}</p>
        </div>

        <div class="app-remark-modal__body">
          <AppFormField label="Comment (optional)">
            <AppTextarea
              ref="textareaRef"
              v-model="remark"
              :rows="3"
              :placeholder="placeholder"
              :disabled="loading"
              class="custom-textarea"
            />
          </AppFormField>
        </div>

        <div class="modal-footer app-remark-modal__footer">
          <AppButton
            variant="secondary"
            size="sm"
            :disabled="loading"
            @click="emit('cancel')"
          >
            Cancel
          </AppButton>
          <AppButton
            variant="primary"
            size="sm"
            :loading="loading"
            @click="onConfirm"
          >
            {{ confirmLabel }}
          </AppButton>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.app-remark-modal {
  align-items: center;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(8px);
  z-index: 1100;
}

.app-remark-modal__dialog {
  width: min(34rem, 90vw);
  border-radius: var(--radius-lg, 8px);
  border: 1px solid var(--border-default);
  background: var(--bg-card);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.15), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.app-remark-modal__header {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: var(--space-5, 20px) var(--space-5, 20px) var(--space-3, 12px);
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-card);
}

.app-remark-modal__title {
  margin: 0;
  font-size: var(--font-size-lg, 1.125rem);
  font-weight: 700;
  color: var(--text-primary);
}

.app-remark-modal__desc {
  margin: 0;
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--text-secondary);
}

.app-remark-modal__body {
  padding: var(--space-5, 20px);
  background: var(--bg-card);
}

.custom-textarea {
  width: 100%;
  border: 1px solid var(--border-subtle);
  background: var(--bg-input);
  color: var(--text-primary);
  border-radius: var(--radius-md, 6px);
  transition: all 0.15s ease;
}

.custom-textarea:focus {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px var(--focus-ring);
}

.app-remark-modal__footer {
  background: var(--bg-card-muted, #f6f8fa);
  border-top: 1px solid var(--border-subtle);
  padding: var(--space-4, 16px) var(--space-5, 20px);
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3, 12px);
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.animate-fade-in {
  animation: fadeIn 0.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
