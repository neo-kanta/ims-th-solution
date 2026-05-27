<script setup lang="ts">
import { computed, ref } from "vue";
import AppBadge from "./AppBadge.vue";
import AppEmptyState from "./AppEmptyState.vue";

interface ContractOption {
  contractId: string;
  currentState?: string;
}

interface Props {
  options: readonly ContractOption[];
  selectedContractId: string | null;
  title?: string;
  placeholder?: string;
}

const props = withDefaults(defineProps<Props>(), {
  title: "Select Contract / Fund",
  placeholder: "Search contracts...",
});

const emit = defineEmits<{
  select: [contractId: string];
}>();

const search = ref("");

const filtered = computed<ContractOption[]>(() => {
  const needle = search.value.trim().toLowerCase();
  if (!needle) return [...props.options];
  return props.options.filter((opt) =>
    opt.contractId.toLowerCase().includes(needle)
  );
});
</script>

<template>
  <aside class="ims-contract-selector" data-testid="ims-contract-selector">
    <label class="ims-contract-selector__label" for="ims-contract-search">
      {{ title }}
    </label>
    <input
      id="ims-contract-search"
      v-model="search"
      type="search"
      class="ims-contract-selector__input"
      :placeholder="placeholder"
      autocomplete="off"
    />

    <div v-if="options.length === 0" class="ims-contract-selector__empty">
      <AppEmptyState
        title="No contracts"
        description="There are no active contracts for selection."
        icon="folder"
      />
    </div>

    <ul v-else class="ims-contract-selector__list">
      <li
        v-for="opt in filtered"
        :key="opt.contractId"
        class="ims-contract-selector__item"
        :class="{ 'is-active': opt.contractId === selectedContractId }"
      >
        <button
          type="button"
          class="ims-contract-selector__button"
          @click="emit('select', opt.contractId)"
        >
          <span class="ims-contract-selector__id" :title="opt.contractId">
            {{ opt.contractId }}
          </span>
          <AppBadge
            v-if="opt.currentState"
            :variant="opt.contractId === selectedContractId ? 'info' : 'neutral'"
            size="sm"
          >
            {{ opt.currentState }}
          </AppBadge>
        </button>
      </li>
    </ul>
  </aside>
</template>

<style scoped>
.ims-contract-selector {
  display: grid;
  gap: var(--space-3, 12px);
  align-content: start;
}

.ims-contract-selector__label {
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-medium, 500);
  color: var(--text-secondary, #57606a);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.ims-contract-selector__input {
  width: 100%;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  padding: var(--space-2, 8px) var(--space-3, 12px);
  font-size: var(--font-size-sm, 14px);
  font-family: inherit;
  background: var(--bg-input, #ffffff);
  color: var(--text-primary, #1f2328);
}

.ims-contract-selector__input:focus {
  outline: none;
  border-color: var(--border-focus, #0969da);
  box-shadow: 0 0 0 3px rgba(9, 105, 218, 0.15);
}

.ims-contract-selector__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-1, 4px);
  max-height: 26rem;
  overflow-y: auto;
}

.ims-contract-selector__button {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2, 8px);
  padding: var(--space-2, 8px) var(--space-3, 12px);
  border: 1px solid transparent;
  border-radius: var(--radius-md, 6px);
  background: transparent;
  font-size: var(--font-size-sm, 14px);
  color: var(--text-primary, #1f2328);
  cursor: pointer;
  text-align: left;
}

.ims-contract-selector__button:hover {
  background: var(--bg-card-muted, #f6f8fa);
}

.ims-contract-selector__item.is-active .ims-contract-selector__button {
  background: var(--bg-selected, #ddf4ff);
  border-color: var(--border-focus, #0969da);
}

.ims-contract-selector__id {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
