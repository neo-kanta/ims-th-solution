<script setup lang="ts">
import AppIcon from "./AppIcon.vue";

interface NavItem {
  id: string;
  label: string;
  description?: string;
  icon: string;
}

interface NavGroup {
  id: string;
  label: string;
  items: NavItem[];
}

defineProps<{
  modelValue: string;
  groups: NavGroup[];
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
  select: [id: string];
}>();

function onItemClick(id: string) {
  emit("update:modelValue", id);
  emit("select", id);
}
</script>

<template>
  <aside class="setting-nav" aria-label="Settings navigation">
    <nav class="setting-nav__menu">
      <div v-for="(group, index) in groups" :key="group.id" class="setting-nav__group-container">
        <hr v-if="index > 0" class="setting-nav__menu-divider" />
        <h2 class="setting-nav__group">{{ group.label }}</h2>
        <button
          v-for="item in group.items"
          :key="item.id"
          class="setting-nav__item"
          :class="{ 'is-active': modelValue === item.id, 'is-danger': group.id === 'danger' }"
          type="button"
          :aria-current="modelValue === item.id ? 'page' : undefined"
          :title="item.description"
          @click="onItemClick(item.id)"
        >
          <AppIcon :name="item.icon" size="sm" :class="{ 'icon-danger': group.id === 'danger' }" />
          <span>{{ item.label }}</span>
        </button>
      </div>
    </nav>
  </aside>
</template>

<style scoped>
.setting-nav {
  position: sticky;
  top: calc(var(--header-height) + var(--space-5));
  display: grid;
  gap: var(--space-4);
  align-self: start;
  max-height: calc(100vh - var(--header-height) - var(--space-8));
  overflow-y: auto;
  padding-right: var(--space-2);
}

.setting-nav__item span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.setting-nav__menu {
  display: grid;
  gap: 2px;
}

.setting-nav__menu-divider {
  border: 0;
  border-top: 1px solid var(--border-subtle);
  margin: var(--space-3) 0 var(--space-4) 0;
}

.setting-nav__group-container {
  display: grid;
  gap: 2px;
  margin-bottom: var(--space-4);
}

.setting-nav__group {
  margin: 0 0 var(--space-2);
  padding: 0 var(--space-3);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.setting-nav__item {
  position: relative;
  min-height: 2.25rem;
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: 0 var(--space-3);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  text-align: left;
}

.setting-nav__item:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
}

.setting-nav__item.is-active {
  background: var(--bg-selected);
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.setting-nav__item.is-active::before {
  content: "";
  position: absolute;
  left: 0;
  top: 10%;
  bottom: 10%;
  width: 4px;
  background: var(--action-primary);
  border-radius: 0 4px 4px 0;
}

.setting-nav__item.is-danger {
  color: var(--alert-danger-text, #d73a49);
}

.setting-nav__item.is-danger:hover {
  background: var(--alert-danger-bg, #ffeef0);
}

.icon-danger {
  color: var(--alert-danger-text, #d73a49);
}

@media (max-width: 1024px) {
  .setting-nav {
    position: static;
    max-height: none;
    overflow: visible;
    padding-right: 0;
  }

  .setting-nav__menu {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--space-2);
  }
}

@media (max-width: 640px) {
  .setting-nav__menu {
    grid-template-columns: 1fr;
  }
}
</style>
