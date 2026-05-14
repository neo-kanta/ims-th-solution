<script setup lang="ts">
import { computed } from "vue";

interface Props {
  total: number;
  offset: number;
  limit: number;
  pageCount: number;
  loading?: boolean;
  pageSizeOptions?: readonly number[];
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  pageSizeOptions: () => [12, 25, 50, 100],
});

const emit = defineEmits<{
  page: [offset: number];
  pageSize: [limit: number];
}>();

const { t } = useI18n();

const rangeLabel = computed(() => {
  if (props.total === 0) {
    return t("common.pagination.empty");
  }

  const start = props.offset + 1;
  const end = Math.min(props.offset + props.pageCount, props.total);
  return t("common.pagination.range", { start, end, total: props.total });
});

const canPageBack = computed(() => props.offset > 0 && !props.loading);
const canPageForward = computed(
  () => props.offset + props.limit < props.total && !props.loading,
);

function previous() {
  emit("page", Math.max(0, props.offset - props.limit));
}

function next() {
  emit("page", props.offset + props.limit);
}

function onPageSizeChange(event: Event) {
  const target = event.target as HTMLSelectElement;
  const next = Number(target.value);
  if (!Number.isFinite(next) || next <= 0) return;
  emit("pageSize", next);
}
</script>

<template>
  <footer class="settings-panel__footer settings-pagination-footer">
    <span class="settings-pagination-footer__range">{{ rangeLabel }}</span>
    <div class="settings-pagination-footer__controls">
      <label class="settings-pagination-footer__size">
        <span class="settings-pagination-footer__size-label">
          {{ t("common.pagination.pageSize") }}
        </span>
        <select
          class="select"
          :value="limit"
          :disabled="loading"
          @change="onPageSizeChange"
        >
          <option v-for="size in pageSizeOptions" :key="size" :value="size">
            {{ size }}
          </option>
        </select>
      </label>
      <div class="settings-pagination">
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="!canPageBack"
          @click="previous"
        >
          {{ t("common.previous") }}
        </AppButton>
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="!canPageForward"
          @click="next"
        >
          {{ t("common.next") }}
        </AppButton>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.settings-pagination-footer {
  flex-wrap: wrap;
  gap: var(--space-3);
}

.settings-pagination-footer__controls {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.settings-pagination-footer__size {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

.settings-pagination-footer__size-label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.settings-pagination-footer__size .select {
  min-height: 2rem;
  padding: 0 var(--space-3);
  width: auto;
}
</style>
