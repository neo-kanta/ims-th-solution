<script setup lang="ts">
import { ref } from "vue";
import AppIcon from "./AppIcon.vue";

interface ListItem {
  label: string;
  value: any;
  copyable?: boolean;
}

interface Props {
  items: ListItem[];
  layout?: "horizontal" | "vertical";
}

withDefaults(defineProps<Props>(), {
  layout: "vertical",
});

const copiedIndex = ref<number | null>(null);

function copyToClipboard(text: string, index: number) {
  if (!import.meta.client) return;
  navigator.clipboard.writeText(text).then(() => {
    copiedIndex.value = index;
    setTimeout(() => {
      if (copiedIndex.value === index) {
        copiedIndex.value = null;
      }
    }, 2000);
  });
}
</script>

<template>
  <dl class="app-desc-list" :class="`app-desc-list--${layout}`">
    <div
      v-for="(item, idx) in items"
      :key="idx"
      class="app-desc-list__item"
    >
      <dt class="app-desc-list__label">{{ item.label }}</dt>
      <dd class="app-desc-list__value-wrapper">
        <span class="app-desc-list__value">
          <slot :name="`value(${item.label})`" :item="item">
            {{ item.value !== undefined && item.value !== null && item.value !== "" ? item.value : "—" }}
          </slot>
        </span>
        <button
          v-if="item.copyable && item.value"
          type="button"
          class="app-desc-list__copy-btn"
          title="Copy value"
          @click="copyToClipboard(String(item.value), idx)"
        >
          <AppIcon
            :name="copiedIndex === idx ? 'check' : 'review'"
            size="xs"
            :class="{ 'copied-success': copiedIndex === idx }"
          />
        </button>
      </dd>
    </div>
  </dl>
</template>

<style scoped>
.app-desc-list {
  margin: 0;
  display: grid;
  gap: var(--space-4, 16px);
}

.app-desc-list--vertical {
  grid-template-columns: 1fr;
}

.app-desc-list--horizontal .app-desc-list__item {
  grid-template-columns: 10rem 1fr;
  align-items: center;
}

.app-desc-list__item {
  display: grid;
  gap: var(--space-1, 4px);
}

.app-desc-list__label {
  font-size: 11px;
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-tertiary, #6e7781);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.app-desc-list__value-wrapper {
  margin: 0;
  display: inline-flex;
  align-items: center;
  gap: var(--space-2, 8px);
  font-size: var(--font-size-sm, 14px);
  color: var(--text-primary, #1f2328);
}

.app-desc-list__value {
  min-width: 0;
  overflow-wrap: anywhere;
}

.app-desc-list__copy-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  padding: var(--space-1, 4px);
  border-radius: var(--radius-sm, 4px);
  color: var(--text-secondary, #57606a);
  cursor: pointer;
  opacity: 0.6;
  transition: opacity 0.15s ease, background-color 0.15s ease;
}

.app-desc-list__copy-btn:hover {
  opacity: 1;
  background: rgba(0, 0, 0, 0.05);
}

.copied-success {
  color: var(--state-success, #1f883d);
}
</style>
