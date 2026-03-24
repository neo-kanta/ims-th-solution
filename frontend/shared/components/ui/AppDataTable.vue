<script setup lang="ts">
interface Column {
  key: string
  label: string
  align?: 'left' | 'right' | 'center'
  sortable?: boolean
  width?: string
  mono?: boolean
}

interface Props {
  columns: Column[]
  rows: Record<string, unknown>[]
  selectable?: boolean
  loading?: boolean
  emptyTitle?: string
  emptyDesc?: string
  sortKey?: string
  sortDir?: 'asc' | 'desc'
}

const props = withDefaults(defineProps<Props>(), {
  selectable: false,
  loading: false,
  emptyTitle: 'No results',
  emptyDesc: undefined,
  sortKey: undefined,
  sortDir: 'asc',
})

const emit = defineEmits<{
  sort: [key: string]
  select: [rows: Record<string, unknown>[]]
}>()

defineSlots<{
  row(props: { row: Record<string, unknown>; index: number }): unknown
}>()

const selectedRows = ref<Set<number>>(new Set())
const allSelected = computed(() => {
  return props.rows.length > 0 && selectedRows.value.size === props.rows.length
})

function toggleAll() {
  if (allSelected.value) {
    selectedRows.value = new Set()
  } else {
    selectedRows.value = new Set(props.rows.map((_, i) => i))
  }
  emitSelected()
}

function toggleRow(index: number) {
  if (selectedRows.value.has(index)) {
    selectedRows.value.delete(index)
  } else {
    selectedRows.value.add(index)
  }
  // trigger reactivity
  selectedRows.value = new Set(selectedRows.value)
  emitSelected()
}

function emitSelected() {
  const selected = props.rows.filter((_, i) => selectedRows.value.has(i))
  emit('select', selected)
}

function onSort(col: Column) {
  if (!col.sortable) return
  emit('sort', col.key)
}

function colClass(col: Column): string[] {
  const classes: string[] = []
  if (col.align === 'right') classes.push('col-right')
  if (col.align === 'center') classes.push('col-center')
  if (col.mono) classes.push('col-mono')
  return classes
}
</script>

<template>
  <div class="data-table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <!-- Checkbox column -->
          <th v-if="selectable" class="col-check">
            <input
              type="checkbox"
              :checked="allSelected"
              :indeterminate="selectedRows.size > 0 && !allSelected"
              @change="toggleAll"
              style="cursor: pointer; width: 14px; height: 14px;"
            />
          </th>

          <!-- Column headers -->
          <th
            v-for="col in columns"
            :key="col.key"
            :class="[
              col.sortable ? 'sortable' : undefined,
              sortKey === col.key ? 'sorted' : undefined,
              col.align === 'right' ? 'col-right' : undefined,
              col.align === 'center' ? 'col-center' : undefined,
            ]"
            :style="col.width ? `width: ${col.width};` : undefined"
            @click="onSort(col)"
          >
            <span style="display: inline-flex; align-items: center; gap: 4px;">
              {{ col.label }}
              <template v-if="col.sortable">
                <svg
                  v-if="sortKey === col.key && sortDir === 'asc'"
                  xmlns="http://www.w3.org/2000/svg"
                  width="11"
                  height="11"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <polyline points="18 15 12 9 6 15" />
                </svg>
                <svg
                  v-else-if="sortKey === col.key && sortDir === 'desc'"
                  xmlns="http://www.w3.org/2000/svg"
                  width="11"
                  height="11"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <polyline points="6 9 12 15 18 9" />
                </svg>
                <svg
                  v-else
                  xmlns="http://www.w3.org/2000/svg"
                  width="11"
                  height="11"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="#cbd5e1"
                  stroke-width="2.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <polyline points="18 15 12 9 6 15" />
                  <polyline points="6 9 12 15 18 9" style="transform: translateY(4px);" />
                </svg>
              </template>
            </span>
          </th>
        </tr>
      </thead>

      <tbody>
        <!-- Loading skeleton rows -->
        <template v-if="loading">
          <tr v-for="n in 5" :key="`skel-${n}`">
            <td v-if="selectable" class="col-check">
              <span class="skeleton" style="display:block; width:14px; height:14px; border-radius:3px;" />
            </td>
            <td v-for="col in columns" :key="col.key">
              <span
                class="skeleton"
                :style="`display:block; height:13px; width:${col.align === 'right' ? '60%' : '80%'}; margin-left:${col.align === 'right' ? 'auto' : '0'};`"
              />
            </td>
          </tr>
        </template>

        <!-- Empty state -->
        <template v-else-if="rows.length === 0">
          <tr>
            <td
              :colspan="selectable ? columns.length + 1 : columns.length"
              style="padding: 0; border-bottom: none;"
            >
              <div class="empty-state">
                <div class="empty-icon">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="36"
                    height="36"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  >
                    <rect x="3" y="3" width="18" height="18" rx="2" />
                    <line x1="3" y1="9" x2="21" y2="9" />
                    <line x1="3" y1="15" x2="21" y2="15" />
                    <line x1="9" y1="9" x2="9" y2="21" />
                  </svg>
                </div>
                <div class="empty-title">{{ emptyTitle }}</div>
                <p v-if="emptyDesc" class="empty-desc">{{ emptyDesc }}</p>
              </div>
            </td>
          </tr>
        </template>

        <!-- Data rows -->
        <template v-else>
          <tr
            v-for="(row, index) in rows"
            :key="index"
            :class="{ selected: selectable && selectedRows.has(index) }"
          >
            <td v-if="selectable" class="col-check">
              <input
                type="checkbox"
                :checked="selectedRows.has(index)"
                @change="toggleRow(index)"
                style="cursor: pointer; width: 14px; height: 14px;"
              />
            </td>
            <slot name="row" :row="row" :index="index" />
          </tr>
        </template>
      </tbody>
    </table>
  </div>
</template>
