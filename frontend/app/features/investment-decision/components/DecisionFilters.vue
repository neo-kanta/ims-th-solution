<script setup lang="ts">
import { DECISION_STATUSES, type DecisionListFilters } from "../services/decisionApi";

const props = defineProps<{ modelValue: DecisionListFilters }>();
const emit = defineEmits<{
  (e: "update:modelValue", value: DecisionListFilters): void;
  (e: "change"): void;
}>();

function update(field: keyof DecisionListFilters, value: string) {
  emit("update:modelValue", { ...props.modelValue, [field]: value || undefined });
}

function onSearchInput(e: Event) {
  update("search", (e.target as HTMLInputElement).value);
}

function onDateChange(e: Event) {
  update("business_date", (e.target as HTMLInputElement).value);
  emit("change");
}

function onStatusChange(e: Event) {
  update("status", (e.target as HTMLSelectElement).value);
  emit("change");
}

function onSearchKeydown(e: KeyboardEvent) {
  if (e.key === "Enter") emit("change");
}
</script>

<template>
  <div class="filters">
    <input
      type="date"
      class="filter-input"
      :value="modelValue.business_date ?? ''"
      @change="onDateChange"
    />
    <select
      class="filter-input"
      :value="modelValue.status ?? ''"
      @change="onStatusChange"
    >
      <option value="">All Statuses</option>
      <option v-for="s in DECISION_STATUSES" :key="s" :value="s">
        {{ s.replace(/_/g, " ") }}
      </option>
    </select>
    <input
      type="text"
      class="filter-input filter-input--search"
      placeholder="Search decision no, instrument…"
      :value="modelValue.search ?? ''"
      @input="onSearchInput"
      @keydown="onSearchKeydown"
    />
  </div>
</template>

<style scoped>
.filters {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.filter-input {
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  color: var(--text-primary);
  height: 32px;
}

.filter-input:focus {
  outline: none;
  border-color: var(--border-focus);
}

.filter-input--search {
  min-width: 220px;
  flex: 1;
}
</style>
