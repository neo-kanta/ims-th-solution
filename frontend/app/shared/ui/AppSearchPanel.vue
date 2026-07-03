<script setup lang="ts">
import AppButton from "./AppButton.vue";

interface Props {
  loading?: boolean;
  showClear?: boolean;
  searchLabel?: string;
  clearLabel?: string;
}

withDefaults(defineProps<Props>(), {
  loading: false,
  showClear: true,
  searchLabel: "Apply Filters",
  clearLabel: "Clear",
});

const emit = defineEmits<{
  search: [];
  clear: [];
}>();

function handleSubmit() {
  emit("search");
}
</script>

<template>
  <form class="app-search-panel" @submit.prevent="handleSubmit">
    <!-- Grid of Search Fields -->
    <div class="app-search-panel__fields">
      <slot />
    </div>

    <!-- Actions Row -->
    <div class="app-search-panel__actions">
      <slot name="actions">
        <AppButton
          type="submit"
          variant="secondary"
          size="sm"
          :loading="loading"
        >
          {{ searchLabel }}
        </AppButton>
        <AppButton
          v-if="showClear"
          type="button"
          variant="ghost"
          size="sm"
          :disabled="loading"
          @click="emit('clear')"
        >
          {{ clearLabel }}
        </AppButton>
      </slot>
    </div>
  </form>
</template>

<style scoped>
.app-search-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4, 16px);
  padding: var(--space-5, 20px);
  background: var(--bg-card, #ffffff);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.app-search-panel__fields {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-4, 16px);
  align-items: end;
  width: 100%;
}

.app-search-panel__actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2, 8px);
}

@media (max-width: 640px) {
  .app-search-panel__fields {
    grid-template-columns: 1fr;
  }
}
</style>
