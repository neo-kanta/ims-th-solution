<script setup lang="ts">
import { computed } from "vue";
import AppIcon from "./AppIcon.vue";

export interface TabItem {
  key: string;
  label: string;
  to?: string;
  icon?: string;
  count?: number | string | null;
  disabled?: boolean;
}

interface Props {
  items: TabItem[];
  modelValue?: string;
  ariaLabel?: string;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: "",
  ariaLabel: "Tabs",
});

const emit = defineEmits<{
  "update:modelValue": [value: string];
  change: [value: string];
}>();

function onSelect(tab: TabItem) {
  if (tab.disabled) return;
  emit("update:modelValue", tab.key);
  emit("change", tab.key);
}
</script>

<template>
  <nav class="app-tabs" role="tablist" :aria-label="ariaLabel">
    <template v-for="item in items" :key="item.key">
      <!-- Route link tab -->
      <NuxtLink
        v-if="item.to"
        :to="item.to"
        role="tab"
        :aria-selected="modelValue === item.key || undefined"
        class="app-tabs__tab"
        :class="{
          'app-tabs__tab--active': modelValue === item.key,
          'app-tabs__tab--disabled': item.disabled
        }"
      >
        <AppIcon v-if="item.icon" :name="item.icon" size="xs" class="app-tabs__icon" />
        <span class="app-tabs__label">{{ item.label }}</span>
        <span
          v-if="item.count !== undefined && item.count !== null"
          class="app-tabs__badge"
          :class="{ 'app-tabs__badge--muted': item.count === '—' || item.count === '' }"
        >
          {{ item.count }}
        </span>
      </NuxtLink>

      <!-- Local state change button tab -->
      <button
        v-else
        type="button"
        role="tab"
        :aria-selected="modelValue === item.key"
        :disabled="item.disabled"
        class="app-tabs__tab"
        :class="{
          'app-tabs__tab--active': modelValue === item.key,
          'app-tabs__tab--disabled': item.disabled
        }"
        @click="onSelect(item)"
      >
        <AppIcon v-if="item.icon" :name="item.icon" size="xs" class="app-tabs__icon" />
        <span class="app-tabs__label">{{ item.label }}</span>
        <span
          v-if="item.count !== undefined && item.count !== null"
          class="app-tabs__badge"
          :class="{ 'app-tabs__badge--muted': item.count === '—' || item.count === '' }"
        >
          {{ item.count }}
        </span>
      </button>
    </template>
  </nav>
</template>

<style scoped>
.app-tabs {
  display: flex;
  gap: var(--space-4); /* GitHub-like horizontal spacing between tabs */
  border-bottom: 1px solid var(--border-subtle);
  overflow-x: auto;
  scrollbar-width: none; /* Hide scrollbar for clean dashboard view */
}

.app-tabs::-webkit-scrollbar {
  display: none;
}

.app-tabs__tab {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-1) var(--space-3) var(--space-1); /* Spacing like GitHub */
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  font-family: inherit;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
  text-decoration: none;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
  outline: none;
  border-radius: 0;
}

.app-tabs__tab:hover:not(.app-tabs__tab--disabled) {
  color: var(--text-primary);
  opacity: 0.85;
}

.app-tabs__tab--active {
  color: var(--text-primary);
  background: transparent !important;
  border-bottom: 2px solid var(--accent-orange, #f78166); /* Orange indicator line like GitHub */
  margin-bottom: -1px;
  font-weight: var(--font-weight-semibold);
  box-shadow: none !important;
  border-radius: 0 !important;
}

/* Specific styling for button tab selectors when active to look like a modern dashboard */
button.app-tabs__tab--active {
  border-bottom-color: var(--accent-orange, #f78166);
  background: transparent !important;
  box-shadow: none !important;
  margin-bottom: -1px;
  border-radius: 0 !important;
  border-left-color: transparent;
  border-right-color: transparent;
  border-top-color: transparent;
}

.app-tabs__tab--disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.app-tabs__icon {
  flex-shrink: 0;
  opacity: 0.7;
  transition: opacity 0.2s ease;
}

.app-tabs__tab--active .app-tabs__icon {
  opacity: 1;
  color: var(--accent-orange, #f78166);
}

.app-tabs__label {
  line-height: 1;
}

.app-tabs__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 18px;
  min-width: 18px;
  padding: 0 var(--space-2);
  border-radius: var(--radius-pill);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
  transition: all 0.2s ease;
}

.app-tabs__tab--active .app-tabs__badge {
  background: var(--bg-selected);
  color: var(--text-primary);
}

.app-tabs__badge--muted {
  opacity: 0.5;
}

/* Theme variations for dark mode orange indicator */
:root[data-theme="dark"] {
  --accent-orange: #fd8c73;
}
</style>
