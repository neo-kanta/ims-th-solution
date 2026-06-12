<script setup lang="ts">
import { computed } from "vue";
import { useI18n, type AppTranslationKey } from "~/composables/useI18n";

import type { BindingView, SourceView } from "../composables/chatStreamReducer";

const props = defineProps<{
  sources?: SourceView[];
  bindings?: BindingView[];
}>();

const { t } = useI18n();

const sources = computed(() => props.sources ?? []);
const bindings = computed(() => props.bindings ?? []);
const show = computed(() => sources.value.length > 0);

// Explicit key map keeps t() type-safe.
const kindKey: Record<string, AppTranslationKey> = {
  tool_raw: "chat.sources.kind.tool_raw",
  calculation: "chat.sources.kind.calculation",
  none: "chat.sources.kind.none",
};

function sourceLabel(s: string): string {
  return t(kindKey[s] ?? "chat.sources.kind.none");
}
</script>

<template>
  <details v-if="show" class="sources-panel">
    <summary class="sources-panel__summary">
      {{ t("chat.sources.toggle") }} ({{ sources.length }})
    </summary>

    <div class="sources-panel__body">
      <!-- Turn-level: which tools produced the data. -->
      <div class="sources-panel__group">
        <div class="sources-panel__group-title">{{ t("chat.sources.toolsTitle") }}</div>
        <ul class="sources-panel__list">
          <li v-for="(s, i) in sources" :key="`src-${i}`" class="source-row">
            <span class="source-row__tool">{{ s.toolName }}</span>
            <span :class="['source-row__state', `source-row__state--${s.state}`]">{{ s.state }}</span>
            <span v-if="s.server" class="source-row__meta">{{ s.server }}</span>
            <span v-if="s.asOf" class="source-row__meta">{{ t("chat.sources.asOf") }} {{ s.asOf }}</span>
          </li>
        </ul>
      </div>

      <!-- Figure-level bindings, when available. -->
      <div v-if="bindings.length" class="sources-panel__group">
        <div class="sources-panel__group-title">{{ t("chat.sources.figuresTitle") }}</div>
        <ul class="sources-panel__list">
          <li v-for="(b, i) in bindings" :key="`bind-${i}`" class="binding-row">
            <span class="binding-row__figure">{{ b.figure }}</span>
            <span class="binding-row__arrow" aria-hidden="true">←</span>
            <span class="binding-row__tool">{{ b.toolName || sourceLabel(b.source) }}</span>
            <span class="binding-row__kind">{{ sourceLabel(b.source) }}</span>
          </li>
        </ul>
      </div>

      <p class="sources-panel__note">{{ t("chat.sources.note") }}</p>
    </div>
  </details>
</template>

<style scoped>
.sources-panel {
  margin-top: 10px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  border-radius: 8px;
  background: var(--bg-canvas, #f8fafc);
  font-size: 12px;
}
.sources-panel__summary {
  cursor: pointer;
  padding: 8px 12px;
  font-weight: 600;
  color: var(--text-secondary, #475569);
  user-select: none;
}
.sources-panel__body {
  padding: 4px 12px 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.sources-panel__group-title {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #64748b);
  margin-bottom: 4px;
}
.sources-panel__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.source-row,
.binding-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.source-row__tool,
.binding-row__tool {
  font-family: var(--font-mono, ui-monospace, monospace);
  color: var(--text-primary, #0f172a);
}
.source-row__state {
  font-size: 11px;
  padding: 0 6px;
  border-radius: 999px;
  border: 1px solid var(--border-subtle, #e2e8f0);
}
.source-row__state--succeeded {
  color: var(--ok-text, #15803d);
}
.source-row__state--failed {
  color: var(--text-danger, #b91c1c);
}
.source-row__state--denied {
  color: var(--warn-text, #b45309);
}
.source-row__meta {
  color: var(--text-tertiary, #64748b);
}
.binding-row__figure {
  font-weight: 600;
  color: var(--text-primary, #0f172a);
}
.binding-row__kind {
  color: var(--text-tertiary, #64748b);
}
.sources-panel__note {
  margin: 0;
  color: var(--text-tertiary, #64748b);
  font-size: 11px;
  line-height: 1.5;
}
</style>
