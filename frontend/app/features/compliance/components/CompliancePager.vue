<script setup lang="ts">
import { computed } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";

interface Props {
  total: number;
  offset: number;
  limit: number;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  "update:offset": [next: number];
  "update:limit": [next: number];
}>();

const PAGE_SIZES = [25, 50, 100, 200] as const;

const start = computed(() =>
  props.total === 0 ? 0 : props.offset + 1,
);
const end = computed(() =>
  Math.min(props.offset + props.limit, props.total),
);
const hasPrev = computed(() => props.offset > 0 && !props.loading);
const hasNext = computed(
  () => props.offset + props.limit < props.total && !props.loading,
);

function prev() {
  if (!hasPrev.value) return;
  emit("update:offset", Math.max(0, props.offset - props.limit));
}

function next() {
  if (!hasNext.value) return;
  emit("update:offset", props.offset + props.limit);
}

function onLimitChange(e: Event) {
  const v = Number.parseInt((e.target as HTMLSelectElement).value, 10);
  if (!Number.isFinite(v) || v <= 0) return;
  emit("update:limit", v);
  emit("update:offset", 0);
}
</script>

<template>
  <div class="pager" role="navigation" aria-label="Pagination">
    <span class="pager__range">
      <strong>{{ start.toLocaleString("en-US") }}</strong>–<strong>{{
        end.toLocaleString("en-US")
      }}</strong>
      of <strong>{{ total.toLocaleString("en-US") }}</strong>
    </span>

    <label class="pager__limit">
      <span>Per page</span>
      <select :value="limit" class="pager__select" @change="onLimitChange">
        <option v-for="size in PAGE_SIZES" :key="size" :value="size">
          {{ size }}
        </option>
      </select>
    </label>

    <div class="pager__nav">
      <AppButton
        variant="ghost"
        size="sm"
        :disabled="!hasPrev"
        @click="prev"
      >
        ← Prev
      </AppButton>
      <AppButton
        variant="ghost"
        size="sm"
        :disabled="!hasNext"
        @click="next"
      >
        Next →
      </AppButton>
    </div>
  </div>
</template>

<style scoped>
.pager {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  justify-content: flex-end;
  flex-wrap: wrap;
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.pager__range {
  font-variant-numeric: tabular-nums;
}

.pager__limit {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

.pager__select {
  height: 28px;
  padding: 0 var(--space-2);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: var(--font-size-xs);
}

.pager__nav {
  display: inline-flex;
  gap: var(--space-2);
}
</style>
