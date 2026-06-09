<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { useI18n } from "~/composables/useI18n";

const props = defineProps<{
  disabled?: boolean;
  streaming?: boolean;
}>();
const emit = defineEmits<{ submit: [text: string]; stop: [] }>();
const { t } = useI18n();

const MAX = 16384;
const draft = ref("");
const textareaRef = ref<HTMLTextAreaElement | null>(null);

const remaining = computed(() => MAX - draft.value.length);
const nearLimit = computed(() => remaining.value <= 500);
const canSend = computed(
  () => draft.value.trim().length > 0 && !props.disabled,
);

function autoGrow() {
  const el = textareaRef.value;
  if (!el) {
    return;
  }
  el.style.height = "auto";
  el.style.height = `${Math.min(el.scrollHeight, 200)}px`;
}

watch(draft, () => {
  void nextTick(autoGrow);
});

function send() {
  const text = draft.value.trim();
  if (!text || props.disabled) {
    return;
  }
  emit("submit", text);
  draft.value = "";
  void nextTick(autoGrow);
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    send();
  }
}

defineExpose({
  focus: () => textareaRef.value?.focus(),
});
</script>

<template>
  <div class="composer">
    <form class="composer__field" @submit.prevent="send">
      <textarea
        ref="textareaRef"
        v-model="draft"
        :maxlength="MAX"
        :placeholder="t('chat.inputPlaceholder')"
        class="composer__textarea"
        rows="1"
        :aria-label="t('chat.inputPlaceholder')"
        @keydown="onKeydown"
      ></textarea>

      <button
        v-if="streaming"
        type="button"
        class="composer__btn composer__btn--stop"
        :title="t('chat.stop')"
        @click="emit('stop')"
      >
        <span class="composer__stop-icon" aria-hidden="true" />
        {{ t("chat.stop") }}
      </button>
      <button
        v-else
        type="submit"
        class="composer__btn composer__btn--send"
        :disabled="!canSend"
        :title="t('chat.send')"
      >
        {{ t("chat.send") }}
      </button>
    </form>

    <div class="composer__footer">
      <span class="composer__hint">{{ t("chat.inputHint") }}</span>
      <span
        v-if="nearLimit"
        class="composer__count"
        :class="{ 'composer__count--over': remaining < 0 }"
      >
        {{ remaining }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.composer {
  padding: 12px 16px 14px;
  border-top: 1px solid var(--border-subtle, #e2e8f0);
  background: var(--bg-elevated, #ffffff);
}
.composer__field {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  padding: 8px 8px 8px 14px;
  border: 1px solid var(--border-input, #cbd5e1);
  border-radius: 14px;
  background: var(--bg-input, #ffffff);
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}
.composer__field:hover,
.composer__field:focus-within {
  border-color: var(--border-strong, #8c959f);
}
.composer__textarea {
  flex: 1;
  resize: none;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text-primary, #0f172a);
  font: inherit;
  line-height: 1.5;
  max-height: 200px;
  padding: 6px 0;
}
.composer__btn {
  flex-shrink: 0;
  min-width: 84px;
  height: 38px;
  padding: 0 16px;
  border-radius: 10px;
  font-weight: 600;
  font-size: 13px;
  cursor: pointer;
  border: 1px solid transparent;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition:
    background 0.15s ease,
    opacity 0.15s ease;
}
.composer__btn--send {
  background: var(--bg-accent, #2563eb);
  color: var(--text-on-primary, #ffffff);
}
.composer__btn--send:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.composer__btn--stop {
  background: var(--bg-elevated, #ffffff);
  border-color: var(--border-input, #cbd5e1);
  color: var(--text-primary, #0f172a);
}
.composer__stop-icon {
  width: 9px;
  height: 9px;
  border-radius: 2px;
  background: var(--text-danger, #b91c1c);
}
.composer__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
  padding: 0 4px;
}
.composer__hint {
  font-size: 11px;
  color: var(--text-tertiary, #64748b);
}
.composer__count {
  font-size: 11px;
  color: var(--text-tertiary, #64748b);
  font-variant-numeric: tabular-nums;
}
.composer__count--over {
  color: var(--text-danger, #b91c1c);
  font-weight: 600;
}
</style>
