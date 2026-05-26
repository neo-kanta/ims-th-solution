<script setup lang="ts">
import { useImportBatches } from "../composables/useImportBatches";
import type { ApiCreateBatchRequest } from "../services/marketDataApi";

const emit = defineEmits<{
  (e: "completed"): void;
}>();

const {
  status,
  lastRunResult,
  errors,
  isCreating,
  isRunning,
  errorMessage,
  createAndRun,
  reset,
} = useImportBatches();

const symbolsText = ref("");
const importType = ref<"QUOTE_SYNC" | "HISTORY_SYNC" | "QUOTE_AND_HISTORY_SYNC">("QUOTE_SYNC");
const chunkSize = ref(25);
const provider = ref<string>("default");

const parsedSymbols = computed<string[]>(() => {
  return symbolsText.value
    .split(/[\s,;]+/)
    .map((s) => s.trim().toUpperCase())
    .filter(Boolean);
});

const totalSymbols = computed(() => parsedSymbols.value.length);

async function onRun() {
  if (parsedSymbols.value.length === 0) return;
  const req: ApiCreateBatchRequest = {
    provider: provider.value || "default",
    import_type: importType.value,
    symbols: parsedSymbols.value,
    chunk_size: chunkSize.value,
    include_quote: importType.value !== "HISTORY_SYNC",
    include_history: importType.value !== "QUOTE_SYNC",
  };
  try {
    await createAndRun(req);
    emit("completed");
  } catch {
    /* error surfaced via errorMessage. */
  }
}

function onReset() {
  reset();
  symbolsText.value = "";
}

function statusVariant(s?: string): string {
  switch (s) {
    case "COMPLETED": return "badge-success";
    case "COMPLETED_WITH_WARNINGS": return "badge-warning";
    case "PARTIAL_FAILED": return "badge-warning";
    case "FAILED": return "badge-danger";
    case "RUNNING": return "badge-neutral";
    case "PENDING": return "badge-neutral";
    default: return "badge-neutral";
  }
}
</script>

<template>
  <div class="card md-batch-panel">
    <div class="card-header md-batch-panel__head">
      <div>
        <span class="card-title">Bulk import</span>
        <div class="card-subtitle">
          Run a chunk-based provider sync. Failures are isolated per chunk and
          unmapped symbols land in the <strong>Unmapped</strong> tab for review.
        </div>
      </div>
      <button
        v-if="status"
        type="button"
        class="btn btn-secondary btn-sm"
        @click="onReset"
      >Reset</button>
    </div>

    <div class="card-body md-batch-panel__body">
      <div class="md-batch-panel__form">
        <label class="md-batch-panel__field">
          <span class="md-batch-panel__label">Symbols</span>
          <textarea
            v-model="symbolsText"
            class="form-input md-batch-panel__textarea"
            rows="4"
            placeholder="KBANK.BK, PTT.BK, AAPL, ..."
          />
          <span class="md-batch-panel__hint">
            {{ totalSymbols }} symbol{{ totalSymbols === 1 ? "" : "s" }} detected
          </span>
        </label>

        <div class="md-batch-panel__controls">
          <label class="md-batch-panel__field">
            <span class="md-batch-panel__label">Type</span>
            <select v-model="importType" class="form-input">
              <option value="QUOTE_SYNC">Quote only</option>
              <option value="HISTORY_SYNC">History only</option>
              <option value="QUOTE_AND_HISTORY_SYNC">Quote + history</option>
            </select>
          </label>
          <label class="md-batch-panel__field">
            <span class="md-batch-panel__label">Chunk size</span>
            <input
              v-model.number="chunkSize"
              type="number"
              min="1"
              max="200"
              class="form-input"
            />
          </label>
          <label class="md-batch-panel__field">
            <span class="md-batch-panel__label">Provider</span>
            <select v-model="provider" class="form-input">
              <option value="default">Default chain</option>
              <option value="alpha_vantage">Alpha Vantage</option>
              <option value="yahoo">Yahoo Finance</option>
            </select>
          </label>
        </div>

        <div class="md-batch-panel__actions">
          <button
            class="btn btn-primary"
            type="button"
            :disabled="isCreating || isRunning || totalSymbols === 0"
            @click="onRun"
          >
            {{ isRunning ? "Running…" : isCreating ? "Creating…" : "Run batch" }}
          </button>
          <span v-if="errorMessage" class="md-batch-panel__error">{{ errorMessage }}</span>
        </div>
      </div>

      <div v-if="status?.batch" class="md-batch-panel__summary">
        <div class="md-batch-panel__summary-head">
          <span class="md-batch-panel__summary-id">batch {{ status.batch.batch_id?.slice(0, 8) }}…</span>
          <span class="badge" :class="statusVariant(status.batch.status)">
            {{ status.batch.status }}
          </span>
        </div>
        <div class="md-batch-panel__counts">
          <div><span class="md-batch-panel__count">{{ status.batch.total_symbols ?? 0 }}</span> symbols</div>
          <div><span class="md-batch-panel__count">{{ status.batch.accepted_records ?? 0 }}</span> accepted</div>
          <div><span class="md-batch-panel__count">{{ status.batch.rejected_records ?? 0 }}</span> rejected</div>
          <div><span class="md-batch-panel__count">{{ status.batch.warning_records ?? 0 }}</span> warnings</div>
          <div><span class="md-batch-panel__count">{{ status.batch.total_chunks ?? 0 }}</span> chunks</div>
        </div>

        <div v-if="status.chunks?.length" class="md-batch-panel__chunks">
          <div
            v-for="c in status.chunks"
            :key="c.chunk_id"
            class="md-batch-panel__chunk"
          >
            <div class="md-batch-panel__chunk-head">
              <span>Chunk #{{ (c.chunk_index ?? 0) + 1 }}</span>
              <span class="badge" :class="statusVariant(c.status)">{{ c.status }}</span>
            </div>
            <div class="md-batch-panel__chunk-meta">
              {{ c.accepted_records ?? 0 }} ok ·
              {{ c.rejected_records ?? 0 }} rejected ·
              {{ c.warning_records ?? 0 }} warn ·
              {{ c.total_records ?? 0 }} total
              <span v-if="c.error_message" class="md-batch-panel__chunk-error">— {{ c.error_message }}</span>
            </div>
          </div>
        </div>

        <div v-if="errors.length" class="md-batch-panel__errors">
          <div class="md-batch-panel__errors-head">Errors ({{ errors.length }})</div>
          <ul class="md-batch-panel__errors-list">
            <li v-for="e in errors" :key="e.item_id">
              <code>{{ e.symbol }}</code>
              <span class="md-batch-panel__error-code">{{ e.status }}</span>
              <span v-if="e.error_code" class="md-batch-panel__error-code">{{ e.error_code }}</span>
              <span class="md-batch-panel__error-message">{{ e.error_message }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.md-batch-panel__head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.5rem;
}
.md-batch-panel__body {
  display: grid;
  gap: var(--space-4, 1rem);
}
.md-batch-panel__form {
  display: grid;
  gap: var(--space-3, 0.75rem);
}
.md-batch-panel__field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.md-batch-panel__label {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted, #6b7280);
}
.md-batch-panel__textarea {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.82rem;
}
.md-batch-panel__hint {
  font-size: 0.78rem;
  color: var(--color-text-muted, #6b7280);
}
.md-batch-panel__controls {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: var(--space-3, 0.75rem);
}
.md-batch-panel__actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.md-batch-panel__error {
  color: var(--color-danger, #dc2626);
  font-size: 0.85rem;
}
.md-batch-panel__summary {
  border-top: 1px solid var(--color-border, #e5e7eb);
  padding-top: var(--space-3, 0.75rem);
  display: grid;
  gap: var(--space-3, 0.75rem);
}
.md-batch-panel__summary-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.md-batch-panel__summary-id {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.78rem;
  color: var(--color-text-muted, #6b7280);
}
.md-batch-panel__counts {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  font-size: 0.85rem;
}
.md-batch-panel__count {
  font-weight: 600;
  font-size: 1rem;
}
.md-batch-panel__chunks {
  display: grid;
  gap: 0.5rem;
}
.md-batch-panel__chunk {
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 0.4rem;
  padding: 0.5rem 0.75rem;
}
.md-batch-panel__chunk-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
}
.md-batch-panel__chunk-meta {
  font-size: 0.78rem;
  color: var(--color-text-muted, #6b7280);
  margin-top: 0.2rem;
}
.md-batch-panel__chunk-error {
  color: var(--color-danger, #dc2626);
}
.md-batch-panel__errors-head {
  font-weight: 600;
  font-size: 0.85rem;
  margin-bottom: 0.4rem;
}
.md-batch-panel__errors-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.3rem;
}
.md-batch-panel__errors-list li {
  display: flex;
  gap: 0.5rem;
  font-size: 0.78rem;
  align-items: center;
  flex-wrap: wrap;
}
.md-batch-panel__error-code {
  display: inline-block;
  padding: 0.05rem 0.4rem;
  border-radius: 0.25rem;
  background: var(--color-bg-muted, #f1f5f9);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.md-batch-panel__error-message {
  color: var(--color-text-muted, #6b7280);
  flex: 1;
  min-width: 200px;
}
@media (max-width: 768px) {
  .md-batch-panel__controls {
    grid-template-columns: 1fr;
  }
}
</style>
