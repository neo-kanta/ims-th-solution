<script setup lang="ts">
import { computed } from "vue";
import AppIcon from "./AppIcon.vue";
import AppLoadingState from "./AppLoadingState.vue";
import AppEmptyState from "./AppEmptyState.vue";

export interface TableColumn {
  key: string;
  label: string;
  sortable?: boolean;
  align?: "left" | "center" | "right";
  width?: string;
}

interface Props {
  columns: TableColumn[];
  items: any[];
  loading?: boolean;
  emptyText?: string;
  error?: string | null;
  sortBy?: string;
  sortOrder?: "asc" | "desc";
  rowClickable?: boolean;
  selectable?: boolean;
  selectedIds?: any[];
  density?: "compact" | "comfortable";
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  emptyText: "No records found.",
  error: null,
  sortBy: "",
  sortOrder: "asc",
  rowClickable: false,
  selectable: false,
  selectedIds: () => [],
  density: "comfortable",
});

const emit = defineEmits<{
  "row-click": [item: any];
  sort: [field: string, order: "asc" | "desc"];
  "select-change": [selectedIds: any[]];
}>();

const isAllSelected = computed(() => {
  if (props.items.length === 0) return false;
  return props.items.every((item) => props.selectedIds.includes(item.id));
});

function toggleSelectAll() {
  if (isAllSelected.value) {
    emit("select-change", []);
  } else {
    emit(
      "select-change",
      props.items.map((item) => item.id).filter((id) => id !== undefined)
    );
  }
}

function toggleSelectRow(id: any) {
  const next = [...props.selectedIds];
  const idx = next.indexOf(id);
  if (idx > -1) {
    next.splice(idx, 1);
  } else {
    next.push(id);
  }
  emit("select-change", next);
}

function handleSort(col: TableColumn) {
  if (!col.sortable) return;
  const isCurrent = props.sortBy === col.key;
  const order = isCurrent && props.sortOrder === "asc" ? "desc" : "asc";
  emit("sort", col.key, order);
}

function handleRowClick(item: any) {
  if (props.rowClickable) {
    emit("row-click", item);
  }
}
</script>

<template>
  <div class="app-table-shell" :class="`app-table-shell--${density}`">
    <!-- Error State Overlay -->
    <div v-if="error" class="alert alert-danger app-table-shell__error" role="alert">
      {{ error }}
    </div>

    <!-- Table Wraps -->
    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <!-- Multi-select checkbox column -->
            <th v-if="selectable" class="app-table__checkbox-col">
              <input
                type="checkbox"
                class="checkbox"
                :checked="isAllSelected"
                aria-label="Select all rows"
                @change="toggleSelectAll"
              />
            </th>
            
            <th
              v-for="col in columns"
              :key="col.key"
              :class="[
                `text-${col.align || 'left'}`,
                { 'is-sortable': col.sortable },
                col.key === sortBy ? 'is-active-sort' : ''
              ]"
              :style="{ width: col.width }"
              @click="handleSort(col)"
            >
              <div class="app-table__header-cell" :class="`justify-${col.align || 'left'}`">
                <span>{{ col.label }}</span>
                <span v-if="col.sortable" class="app-table__sort-icon">
                  <template v-if="sortBy === col.key">
                    <AppIcon
                      :name="sortOrder === 'asc' ? 'trend-up' : 'trend-down'"
                      size="xs"
                    />
                  </template>
                  <template v-else>
                    <span class="app-table__sort-placeholder">↕</span>
                  </template>
                </span>
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          <!-- Loading Row -->
          <tr v-if="loading">
            <td :colspan="columns.length + (selectable ? 1 : 0)" class="app-table__state-row">
              <AppLoadingState message="Fetching data..." />
            </td>
          </tr>

          <!-- Empty Row -->
          <tr v-else-if="items.length === 0">
            <td :colspan="columns.length + (selectable ? 1 : 0)" class="app-table__state-row">
              <AppEmptyState :title="emptyText" icon="table" />
            </td>
          </tr>

          <!-- Table Content Rows -->
          <template v-else>
            <tr
              v-for="(item, idx) in items"
              :key="item.id || idx"
              :class="{
                'is-clickable': rowClickable,
                'is-selected': selectedIds.includes(item.id)
              }"
              @click="handleRowClick(item)"
            >
              <td v-if="selectable" class="app-table__checkbox-col" @click.stop>
                <input
                  type="checkbox"
                  class="checkbox"
                  :checked="selectedIds.includes(item.id)"
                  :aria-label="`Select row ${idx + 1}`"
                  @change="toggleSelectRow(item.id)"
                />
              </td>
              
              <td
                v-for="col in columns"
                :key="col.key"
                :class="`text-${col.align || 'left'}`"
              >
                <!-- Scoped Slot for custom rendering -->
                <slot :name="`cell(${col.key})`" :item="item" :value="item[col.key]">
                  {{ item[col.key] !== undefined && item[col.key] !== null ? item[col.key] : "—" }}
                </slot>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <!-- Pagination slot -->
    <div v-if="$slots.pagination" class="app-table-shell__pagination">
      <slot name="pagination" />
    </div>
  </div>
</template>

<style scoped>
.app-table-shell {
  display: grid;
  gap: var(--space-4, 16px);
  width: 100%;
}

.app-table-shell__error {
  margin: 0;
}

.app-table__checkbox-col {
  width: 40px;
  text-align: center;
}

.app-table__header-cell {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1, 4px);
}

.justify-left {
  justify-content: flex-start;
}
.justify-center {
  justify-content: center;
}
.justify-right {
  justify-content: flex-end;
}

.is-sortable {
  cursor: pointer;
  user-select: none;
}

.is-sortable:hover {
  background: var(--bg-card-hover, #f6f8fa);
}

.app-table__sort-icon {
  display: inline-flex;
  color: var(--text-tertiary, #6e7781);
}

.is-active-sort .app-table__sort-icon {
  color: var(--action-primary, #0969da);
}

.app-table__sort-placeholder {
  opacity: 0.3;
  font-size: 10px;
}

.app-table__state-row {
  text-align: center;
  background: transparent !important;
}

.is-clickable {
  cursor: pointer;
}

.is-clickable:hover {
  background: var(--bg-card-hover, #f6f8fa);
}

.is-selected {
  background: var(--bg-selected, #ddf4ff) !important;
}

.app-table-shell__pagination {
  border-top: 1px solid var(--border-subtle, #d0d7de);
  padding-top: var(--space-4, 16px);
}
</style>
