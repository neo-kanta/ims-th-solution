<script setup lang="ts">
import { computed, ref } from "vue";

import AppBadge from "~/shared/ui/AppBadge.vue";
import { useI18n } from "~/composables/useI18n";
import type { FundHeader } from "../types";

const props = defineProps<{
  header: FundHeader;
}>();

const { t } = useI18n();

// Watch/Star/Subscribe currently emit no-op events. They will be wired
// to a notification module endpoint in a future milestone.
const localSubscribed = ref(props.header.subscribed);
const localWatching = ref(false);
const localStarred = ref(false);

const contractBadge = computed(() =>
  t("holdings.badges.contract", { code: props.header.contract_code }, `Contract ${props.header.contract_code}`),
);

function toggleWatch() {
  localWatching.value = !localWatching.value;
}
function toggleStar() {
  localStarred.value = !localStarred.value;
}
function toggleSubscribe() {
  localSubscribed.value = !localSubscribed.value;
}
</script>

<template>
  <header class="fw-header">
    <div class="fw-header__breadcrumb">
      <span class="fw-header__crumb fw-header__crumb--root">{{ header.code }}</span>
      <span class="fw-header__sep">/</span>
      <span class="fw-header__crumb fw-header__crumb--active">{{ header.short_name }}</span>

      <AppBadge variant="neutral" size="sm" class="fw-header__badge">
        {{ t("holdings.badges.private", "Private") }}
      </AppBadge>
      <AppBadge variant="info" size="sm" class="fw-header__badge">
        {{ contractBadge }}
      </AppBadge>
    </div>

    <div class="fw-header__actions">
      <button
        type="button"
        class="fw-header__action"
        :class="{ 'is-active': localWatching }"
        @click="toggleWatch"
      >
        <span class="fw-header__action-label">{{ t("holdings.headerActions.watch", "Watch") }}</span>
        <span class="fw-header__action-count">{{ header.watch_count }}</span>
      </button>

      <button
        type="button"
        class="fw-header__action"
        :class="{ 'is-active': localStarred }"
        @click="toggleStar"
      >
        <span class="fw-header__action-label">{{ t("holdings.headerActions.star", "Star") }}</span>
        <span class="fw-header__action-count">{{ header.star_count }}</span>
      </button>

      <button
        type="button"
        class="fw-header__action fw-header__action--primary"
        :class="{ 'is-active': localSubscribed }"
        @click="toggleSubscribe"
      >
        {{
          localSubscribed
            ? t("holdings.headerActions.subscribed", "Subscribed")
            : t("holdings.headerActions.subscribe", "Subscribe")
        }}
      </button>
    </div>
  </header>
</template>

<style scoped>
.fw-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  flex-wrap: wrap;
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--border-subtle);
}

.fw-header__breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
  flex-wrap: wrap;
}

.fw-header__crumb {
  font-size: 1.05rem;
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-primary-600, #0969da);
}

.fw-header__crumb--root {
  font-weight: 500;
  color: var(--color-primary-600, #0969da);
}

.fw-header__sep {
  color: var(--text-placeholder);
  font-weight: 400;
}

.fw-header__badge {
  margin-left: var(--space-1);
}

.fw-header__actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.fw-header__action {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  padding: 0 10px;
  background: var(--action-secondary);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary);
  cursor: pointer;
  transition: background-color 0.15s ease, border-color 0.15s ease;
}

.fw-header__action:hover {
  background: var(--action-secondary-hover);
  border-color: var(--border-strong);
}

.fw-header__action.is-active {
  background: var(--status-executed-bg);
  border-color: var(--border-focus);
  color: var(--status-executed-text);
}

.fw-header__action--primary.is-active {
  background: var(--status-approved-bg);
  border-color: var(--border-success);
  color: var(--status-approved-text);
}

.fw-header__action-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 6px;
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-primary);
}

@media (max-width: 640px) {
  .fw-header {
    align-items: stretch;
  }
  .fw-header__actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
