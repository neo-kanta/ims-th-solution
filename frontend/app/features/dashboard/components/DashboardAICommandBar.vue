<script setup lang="ts">
import { computed, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppIcon from "~/shared/ui/AppIcon.vue";

const emit = defineEmits<{
  previewSubmit: [query: string];
}>();

const { t } = useI18n();

const query = ref("");

const actionPills = computed(() => [
  { label: t("dashboard.aiBar.pills.createTask", "Create task"), icon: "plus" },
  { label: t("dashboard.aiBar.pills.summarizeAlerts", "Summarize alerts"), icon: "warning" },
  { label: t("dashboard.aiBar.pills.draftReviewNote", "Draft review note"), icon: "review" },
] as const);

const canSend = computed(() => query.value.trim().length > 0);

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    handleSubmit();
  }
}

function handleSubmit() {
  const trimmed = query.value.trim();
  if (!trimmed) return;

  emit("previewSubmit", trimmed);
  query.value = "";
}
</script>

<template>
  <section class="ai-panel" aria-labelledby="dashboard-ai-title">
    <div class="ai-panel__box">
      <div class="ai-panel__status">
        <h2 id="dashboard-ai-title" class="ai-panel__status-title">
          <AppIcon name="system" size="xs" />
          <span>{{ t("dashboard.aiBar.assistant", "IMS Assistant") }}</span>
        </h2>
        <span class="ai-panel__preview">{{ t("dashboard.aiBar.preview", "Preview") }}</span>
      </div>

      <label class="ai-panel__sr-only" for="dashboard-ai-query">
        {{ t("dashboard.aiBar.srPrompt", "Dashboard assistant prompt") }}
      </label>
      <textarea
        id="dashboard-ai-query"
        v-model="query"
        class="ai-panel__textarea"
        :placeholder="
          t(
            'dashboard.aiBar.placeholder',
            'Ask about portfolios, tasks, alerts, workflow status...',
          )
        "
        rows="3"
        @keydown="handleKeydown"
      />

      <div class="ai-panel__footer">
        <div class="ai-panel__actions" :aria-label="t('dashboard.aiBar.srPresets', 'Assistant command presets')">
          <button class="ai-panel__pill" type="button">
            <AppIcon name="chat" size="xs" />
            <span>{{ t("dashboard.aiBar.ask", "Ask") }}</span>
            <AppIcon name="chevron-down" size="xs" />
          </button>
          <button class="ai-panel__pill" type="button">
            <AppIcon name="portfolio" size="xs" />
            <span>{{ t("dashboard.aiBar.allPortfolios", "All portfolios") }}</span>
            <AppIcon name="chevron-down" size="xs" />
          </button>
          <button
            v-for="pill in actionPills"
            :key="pill.label"
            class="ai-panel__pill"
            type="button"
          >
            <AppIcon :name="pill.icon" size="xs" />
            <span>{{ pill.label }}</span>
          </button>
        </div>

        <div class="ai-panel__controls">
          <div class="ai-panel__model" :aria-label="t('dashboard.aiBar.srModel', 'Assistant model')">
            <span>{{ t("dashboard.aiBar.assistant", "IMS Assistant") }}</span>
            <AppIcon name="chevron-down" size="xs" />
          </div>

          <button
            class="ai-panel__send"
            type="button"
            :disabled="!canSend"
            :aria-label="t('dashboard.aiBar.srSend', 'Send assistant prompt')"
            @click="handleSubmit"
          >
            <AppIcon name="send" size="xs" />
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.ai-panel {
  display: block;
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}

.ai-panel__status,
.ai-panel__status-title,
.ai-panel__model,
.ai-panel__pill,
.ai-panel__send {
  display: inline-flex;
  align-items: center;
}

.ai-panel__status {
  justify-content: space-between;
  gap: var(--space-3);
  width: 100%;
}

.ai-panel__status-title {
  gap: var(--space-2);
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.ai-panel__preview {
  padding: 2px 7px;
  color: var(--status-executed-text);
  background: var(--status-executed-bg);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-pill);
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.ai-panel__model {
  gap: var(--space-2);
  min-height: 30px;
  padding: 0 var(--space-3);
  color: var(--text-secondary);
  background: var(--bg-card-muted);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  white-space: nowrap;
  cursor: pointer;
  transition:
    background var(--transition-fast),
    border-color var(--transition-fast),
    color var(--transition-fast);
}

.ai-panel__model:hover {
  color: var(--text-primary);
  background: var(--bg-row-hover);
  border-color: var(--border-default);
}

.ai-panel__box {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--bg-card);
  border: 0;
  border-radius: var(--radius-lg);
  transition:
    border-color var(--transition-base),
    box-shadow var(--transition-base);
}

.ai-panel__textarea {
  width: 100%;
  min-height: 4rem;
  padding: 0;
  color: var(--text-primary);
  background: transparent;
  border: 0;
  outline: none;
  resize: vertical;
  font: inherit;
  line-height: var(--line-height-relaxed);
}

.ai-panel__textarea::placeholder {
  color: var(--text-placeholder);
}

.ai-panel__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.ai-panel__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.ai-panel__controls {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

.ai-panel__pill {
  gap: var(--space-2);
  min-height: 30px;
  padding: 0 var(--space-3);
  color: var(--text-secondary);
  background: var(--action-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition:
    background var(--transition-fast),
    border-color var(--transition-fast),
    color var(--transition-fast);
}

.ai-panel__pill:hover {
  color: var(--text-primary);
  background: var(--bg-row-hover);
  border-color: var(--border-default);
}

.ai-panel__send {
  justify-content: center;
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  color: var(--text-on-primary);
  background: var(--action-primary);
  border: 1px solid var(--action-primary);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition:
    background var(--transition-fast),
    transform var(--transition-fast),
    opacity var(--transition-fast);
}

.ai-panel__send:hover:not(:disabled) {
  background: var(--action-primary-hover);
  transform: translateY(-1px);
}

.ai-panel__send:disabled {
  color: var(--text-disabled);
  background: var(--bg-disabled);
  border-color: var(--border-subtle);
  cursor: not-allowed;
}

.ai-panel__sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 720px) {
  .ai-panel__footer {
    align-items: stretch;
    flex-direction: column;
  }

  .ai-panel__controls,
  .ai-panel__model {
    justify-content: space-between;
  }

  .ai-panel__controls {
    width: 100%;
  }

  .ai-panel__model {
    flex: 1 1 auto;
  }

  .ai-panel__send {
    width: 38px;
  }
}
</style>
