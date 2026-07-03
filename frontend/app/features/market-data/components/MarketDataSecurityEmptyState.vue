<script setup lang="ts">
import type { DetailMappingStatus } from "../market-data.types";

defineProps<{
  mappingStatus: DetailMappingStatus;
  isRunning?: boolean;
}>();

const emit = defineEmits<{
  (e: "sync-quote"): void;
  (e: "import-history"): void;
  (e: "open-mappings"): void;
}>();
</script>

<template>
  <section
    class="md-sec-empty"
    role="status"
    aria-live="polite"
  >
    <div class="md-sec-empty__main">
      <div class="md-sec-empty__icon" aria-hidden="true">
        <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M8 1v14M1 8h14" />
        </svg>
      </div>
      <div class="md-sec-empty__copy">
        <h2 class="md-sec-empty__title">No market data imported yet</h2>
        <p class="md-sec-empty__text">
          This security is mapped in IMS, but no quote or price history has
          been imported. Sync now to load the latest quote and start building
          the price history.
        </p>
      </div>
    </div>
    <div class="md-sec-empty__actions">
      <button
        v-if="mappingStatus === 'MAPPED'"
        class="btn btn-primary btn-sm"
        type="button"
        :disabled="isRunning"
        @click="emit('sync-quote')"
      >Sync latest quote</button>
      <button
        v-if="mappingStatus === 'MAPPED'"
        class="btn btn-secondary btn-sm"
        type="button"
        :disabled="isRunning"
        @click="emit('import-history')"
      >Import quote + history</button>
      <button
        v-if="mappingStatus !== 'MAPPED'"
        class="btn btn-primary btn-sm"
        type="button"
        @click="emit('open-mappings')"
      >Configure provider mapping</button>
    </div>
  </section>
</template>

<style scoped>
.md-sec-empty {
  border: 1px solid var(--md-border);
  background: var(--md-info-bg, var(--md-surface-muted));
  color: var(--md-text);
  border-radius: 0.75rem;
  padding: 1rem 1.25rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}
.md-sec-empty__main {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}
.md-sec-empty__icon {
  color: var(--md-info, var(--md-accent));
  flex-shrink: 0;
  margin-top: 0.15rem;
}
.md-sec-empty__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}
.md-sec-empty__text {
  margin: 0.2rem 0 0;
  color: var(--md-text-muted);
  font-size: 0.85rem;
  max-width: 60ch;
}
.md-sec-empty__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}
</style>
