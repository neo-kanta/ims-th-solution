<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRouter } from "#imports";

import { useI18n } from "~/composables/useI18n";
import AppFormField from "~/shared/ui/AppFormField.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppSection from "~/shared/ui/AppSection.vue";

import { useCompliancePortfolioDirectory } from "../composables/useCompliancePortfolioDirectory";
import { buildPortfolioComplianceHref, filterPortfolioOptions } from "../lib/portfolioFinder";

const { t } = useI18n();
const router = useRouter();
const directory = useCompliancePortfolioDirectory();

const query = ref("");
const activeIndex = ref(0);

const results = computed(() =>
  filterPortfolioOptions(directory.items.value, query.value),
);

watch(query, () => {
  activeIndex.value = 0;
});

function open(code: string) {
  if (!code) return;
  query.value = "";
  void router.push(buildPortfolioComplianceHref(code));
}

function onFocus() {
  void directory.ensureLoaded();
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    query.value = "";
    return;
  }
  if (results.value.length === 0) return;
  if (event.key === "ArrowDown") {
    event.preventDefault();
    activeIndex.value = (activeIndex.value + 1) % results.value.length;
  } else if (event.key === "ArrowUp") {
    event.preventDefault();
    activeIndex.value =
      (activeIndex.value - 1 + results.value.length) % results.value.length;
  } else if (event.key === "Enter") {
    const hit = results.value[activeIndex.value];
    if (hit?.code) {
      event.preventDefault();
      open(hit.code);
    }
  }
}
</script>

<template>
  <AppSection
    :title="t('compliance.dashboard.finder.title')"
    :description="t('compliance.dashboard.finder.description')"
  >
    <div class="finder">
      <AppFormField id="compliance-portfolio-finder-input" :label="t('compliance.dashboard.finder.inputLabel')">
        <AppInput
          id="compliance-portfolio-finder-input"
          v-model="query"
          type="search"
          :placeholder="t('compliance.dashboard.finder.placeholder')"
          role="combobox"
          :aria-expanded="results.length > 0"
          :aria-activedescendant="
            results.length > 0
              ? `compliance-portfolio-option-${activeIndex}`
              : undefined
          "
          aria-autocomplete="list"
          aria-controls="compliance-portfolio-finder-results"
          @focus="onFocus"
          @keydown="onKeydown"
        />
      </AppFormField>

      <p v-if="directory.error.value" class="finder__error" role="alert">
        {{ directory.error.value }}
      </p>

      <p v-else-if="directory.loading.value" class="finder__loading" role="status">
        {{ t("compliance.common.loading") }}
      </p>

      <ul
        v-else-if="query.trim() && results.length > 0"
        id="compliance-portfolio-finder-results"
        class="finder__results"
        role="listbox"
      >
        <li
          v-for="(option, idx) in results"
          :key="option.id"
          :id="`compliance-portfolio-option-${idx}`"
          role="option"
          :aria-selected="idx === activeIndex"
        >
          <button
            type="button"
            class="finder__option"
            :class="{ 'finder__option--active': idx === activeIndex }"
            @click="open(option.code)"
          >
            <span class="finder__option-code">{{ option.code }}</span>
            <span class="finder__option-name">{{ option.name }}</span>
          </button>
        </li>
      </ul>

      <p v-else-if="query.trim() && results.length === 0" class="finder__empty">
        {{ t("compliance.dashboard.finder.noMatches") }}
      </p>
    </div>
  </AppSection>
</template>

<style scoped>
.finder {
  display: grid;
  gap: var(--space-2);
}

.finder__error {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--state-danger);
}

.finder__loading {
  margin: 0;
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.finder__results {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-1);
  max-height: 16rem;
  overflow-y: auto;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.finder__option {
  width: 100%;
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: transparent;
  border: none;
  text-align: left;
  cursor: pointer;
  font-family: inherit;
  color: var(--text-primary);
}

.finder__option:hover,
.finder__option--active {
  background: var(--bg-card-hover);
}

.finder__option-code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  color: var(--text-secondary);
}

.finder__option-name {
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.finder__empty {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}
</style>
