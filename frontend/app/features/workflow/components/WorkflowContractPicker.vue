<script setup lang="ts">
/**
 * Contract picker for the workflow console.
 * Renders a searchable list of contractId + currentState pairs sourced from
 * the integration dashboard snapshot. previousDay status (NICE-TO-HAVE) is
 * surfaced when the backend reports it for the currently selected contract.
 */
import { computed, ref } from "vue";

import AppBadge from "~/shared/ui/AppBadge.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";

interface ContractOption {
  contractId: string;
  currentState: string;
}

const props = defineProps<{
  options: readonly ContractOption[];
  selectedContractId: string | null;
  labels: {
    title: string;
    placeholder: string;
    emptyTitle: string;
    emptyDescription: string;
  };
}>();

const emit = defineEmits<{
  select: [contractId: string];
}>();

const search = ref("");

const filtered = computed<ContractOption[]>(() => {
  const needle = search.value.trim().toLowerCase();
  if (!needle) return [...props.options];
  return props.options.filter((opt) =>
    opt.contractId.toLowerCase().includes(needle),
  );
});
</script>

<template>
  <aside class="workflow-picker" data-testid="workflow-contract-picker">
    <label class="workflow-picker__label" for="workflow-contract-search">
      {{ labels.title }}
    </label>
    <input
      id="workflow-contract-search"
      v-model="search"
      type="search"
      class="workflow-picker__input"
      :placeholder="labels.placeholder"
    />

    <div v-if="options.length === 0" class="workflow-picker__empty">
      <AppEmptyState
        :title="labels.emptyTitle"
        :description="labels.emptyDescription"
      />
    </div>

    <ul v-else class="workflow-picker__list">
      <li
        v-for="opt in filtered"
        :key="opt.contractId"
        class="workflow-picker__item"
        :class="{ 'is-active': opt.contractId === selectedContractId }"
      >
        <button
          type="button"
          class="workflow-picker__button"
          @click="emit('select', opt.contractId)"
        >
          <span class="workflow-picker__id" :title="opt.contractId">
            {{ opt.contractId }}
          </span>
          <AppBadge
            :variant="opt.contractId === selectedContractId ? 'info' : 'neutral'"
            size="sm"
          >
            {{ opt.currentState || "NOT_STARTED" }}
          </AppBadge>
        </button>
      </li>
    </ul>
  </aside>
</template>

<style scoped>
.workflow-picker {
  display: grid;
  gap: var(--space-3);
  align-content: start;
}

.workflow-picker__label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.workflow-picker__input {
  width: 100%;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  font-family: inherit;
  background: var(--bg-input, white);
  color: var(--text-primary);
}

.workflow-picker__input:focus {
  outline: 2px solid var(--action-primary);
  outline-offset: -1px;
  border-color: var(--action-primary);
}

.workflow-picker__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-1);
  max-height: 26rem;
  overflow-y: auto;
}

.workflow-picker__button {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  cursor: pointer;
  text-align: left;
}

.workflow-picker__button:hover {
  background: var(--bg-card-muted, #f9fafb);
}

.workflow-picker__item.is-active .workflow-picker__button {
  background: var(--bg-card-muted, #f1f5f9);
  border-color: var(--action-primary);
}

.workflow-picker__id {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 14rem;
}
</style>
