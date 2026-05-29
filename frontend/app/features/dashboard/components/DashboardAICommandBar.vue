<script setup lang="ts">
import { ref } from "vue";

// Phase 1: AI command bar is always in disabled state.
// This component renders the input shell so the design is in place,
// but submitting does nothing. Real AI routing is wired in a later phase
// once backend AI gateway and explicit config are in place.

const { t } = useI18n();
const query = ref("");

function handleSubmit() {
  // Intentionally no-op in Phase 1 — AI provider not configured.
  query.value = "";
}
</script>

<template>
  <div class="ai-bar" aria-label="AI command bar (coming soon)">
    <div class="ai-bar__icon-wrap">
      <AppIcon name="system" size="sm" class="ai-bar__icon" />
    </div>
    <input
      v-model="query"
      class="ai-bar__input"
      type="text"
      :placeholder="
        t('dashboard.aiBar.placeholder', 'Ask about portfolios, tasks, alerts… (coming soon)')
      "
      :title="t('dashboard.aiBar.disabledTitle', 'AI assistant is not configured yet')"
      disabled
      @keydown.enter.prevent="handleSubmit"
    />
    <span class="ai-bar__label">
      {{ t("dashboard.aiBar.soon", "Soon") }}
    </span>
  </div>
</template>

<style scoped>
.ai-bar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-4);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  opacity: 0.6;
  cursor: not-allowed;
}

.ai-bar__icon {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.ai-bar__input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  cursor: not-allowed;
}

.ai-bar__input::placeholder {
  color: var(--text-tertiary);
}

.ai-bar__label {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  background: var(--border-subtle);
  padding: 2px var(--space-2);
  border-radius: var(--radius-pill);
  flex-shrink: 0;
}
</style>
