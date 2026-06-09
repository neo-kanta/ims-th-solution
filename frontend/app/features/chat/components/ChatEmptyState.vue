<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "~/composables/useI18n";

const emit = defineEmits<{ suggest: [text: string] }>();
const { t } = useI18n();

// Suggested starter prompts. Labels are localized; the sent text is the label.
const suggestions = computed(() => [
  { key: "funds", icon: "📂", text: t("chat.suggestions.funds") },
  { key: "portfolios", icon: "📊", text: t("chat.suggestions.portfolios") },
  { key: "nav", icon: "💹", text: t("chat.suggestions.nav") },
  { key: "valuation", icon: "🧮", text: t("chat.suggestions.valuation") },
]);
</script>

<template>
  <div class="empty">
    <div class="empty__badge" aria-hidden="true">✦</div>
    <h2 class="empty__title">{{ t("chat.emptyTitle") }}</h2>
    <p class="empty__hint">{{ t("chat.emptyHint") }}</p>

    <div class="empty__suggestions">
      <button
        v-for="s in suggestions"
        :key="s.key"
        type="button"
        class="empty__chip"
        @click="emit('suggest', s.text)"
      >
        <span class="empty__chip-icon" aria-hidden="true">{{ s.icon }}</span>
        <span>{{ s.text }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.empty {
  margin: auto;
  max-width: 640px;
  padding: 40px 24px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.empty__badge {
  width: 52px;
  height: 52px;
  display: grid;
  place-items: center;
  border-radius: 16px;
  font-size: 24px;
  color: var(--text-accent, #3730a3);
  background: var(--bg-accent-subtle, #eef2ff);
  border: 1px solid var(--border-accent, #c7d2fe);
  margin-bottom: 16px;
}
.empty__title {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary, #0f172a);
}
.empty__hint {
  margin: 8px 0 24px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-secondary, #475569);
}
.empty__suggestions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  width: 100%;
}
.empty__chip {
  display: flex;
  align-items: center;
  gap: 10px;
  text-align: left;
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  background: var(--bg-elevated, #ffffff);
  color: var(--text-primary, #0f172a);
  font-size: 13px;
  line-height: 1.4;
  cursor: pointer;
  transition: border-color 0.15s ease, transform 0.1s ease, box-shadow 0.15s ease;
}
.empty__chip:hover {
  border-color: var(--border-accent, #c7d2fe);
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.06);
  transform: translateY(-1px);
}
.empty__chip-icon {
  font-size: 16px;
}
@media (max-width: 520px) {
  .empty__suggestions {
    grid-template-columns: 1fr;
  }
}
</style>
