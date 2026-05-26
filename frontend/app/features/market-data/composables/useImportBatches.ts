/**
 * Import batch composable — drives the chunk-based provider sync workflow.
 *
 * The backend exposes:
 *   POST /market-data/import-batches            (create)
 *   POST /market-data/import-batches/:id/run    (run synchronously)
 *   GET  /market-data/import-batches/:id        (status + chunk summary)
 *   GET  /market-data/import-batches/:id/errors (per-symbol error rows)
 *
 * This composable holds the in-flight batch's status, errors, and a simple
 * activity log the UI can render. It intentionally keeps the public surface
 * small — components shouldn't need to know the batch endpoint shapes.
 */
import {
  marketDataApi,
  type ApiBatchStatusResponse,
  type ApiCreateBatchRequest,
  type ApiCreateBatchResponse,
  type ApiImportChunkItem,
  type ApiImportChunkSummary,
  type ApiRunBatchResponse,
} from "../services/marketDataApi";

export interface UseImportBatchesOptions {
  /** Polling interval in ms when watching a running batch. Default: 2000 */
  pollIntervalMs?: number;
}

export function useImportBatches(options: UseImportBatchesOptions = {}) {
  const pollMs = options.pollIntervalMs ?? 2000;

  const activeBatchId = ref<string | null>(null);
  const status = ref<ApiBatchStatusResponse | null>(null);
  const lastRunResult = ref<ApiRunBatchResponse | null>(null);
  const errors = ref<ApiImportChunkItem[]>([]);
  const isRunning = ref(false);
  const isCreating = ref(false);
  const errorMessage = ref<string | null>(null);

  let pollHandle: ReturnType<typeof setTimeout> | null = null;

  function stopPolling() {
    if (pollHandle) {
      clearTimeout(pollHandle);
      pollHandle = null;
    }
  }

  async function create(req: ApiCreateBatchRequest): Promise<ApiCreateBatchResponse> {
    isCreating.value = true;
    errorMessage.value = null;
    try {
      const resp = await marketDataApi.createImportBatch(req);
      if (resp.batch_id) {
        activeBatchId.value = resp.batch_id;
        await refreshStatus(resp.batch_id);
      }
      return resp;
    } catch (err) {
      errorMessage.value = err instanceof Error ? err.message : "Failed to create batch";
      throw err;
    } finally {
      isCreating.value = false;
    }
  }

  async function run(batchId: string): Promise<ApiRunBatchResponse> {
    isRunning.value = true;
    errorMessage.value = null;
    try {
      const resp = await marketDataApi.runImportBatch(batchId);
      lastRunResult.value = resp;
      await refreshStatus(batchId);
      await refreshErrors(batchId);
      return resp;
    } catch (err) {
      errorMessage.value = err instanceof Error ? err.message : "Failed to run batch";
      throw err;
    } finally {
      isRunning.value = false;
    }
  }

  async function createAndRun(req: ApiCreateBatchRequest): Promise<ApiRunBatchResponse | null> {
    const created = await create(req);
    if (!created.batch_id) return null;
    return run(created.batch_id);
  }

  async function refreshStatus(batchId?: string) {
    const id = batchId ?? activeBatchId.value;
    if (!id) return;
    try {
      status.value = await marketDataApi.getImportBatch(id);
    } catch (err) {
      errorMessage.value = err instanceof Error ? err.message : "Failed to load batch status";
    }
  }

  async function refreshErrors(batchId?: string) {
    const id = batchId ?? activeBatchId.value;
    if (!id) return;
    try {
      errors.value = await marketDataApi.getImportBatchErrors(id);
    } catch {
      /* errors endpoint failure is non-fatal */
    }
  }

  /**
   * Poll the batch until status leaves PENDING/RUNNING. Useful if the backend
   * gains background-worker execution; today the /run endpoint is synchronous.
   */
  async function watch(batchId: string) {
    activeBatchId.value = batchId;
    stopPolling();
    await refreshStatus(batchId);
    const s = status.value?.batch?.status;
    if (s === "PENDING" || s === "RUNNING") {
      pollHandle = setTimeout(() => { void watch(batchId); }, pollMs);
    } else {
      await refreshErrors(batchId);
    }
  }

  function reset() {
    stopPolling();
    activeBatchId.value = null;
    status.value = null;
    lastRunResult.value = null;
    errors.value = [];
    errorMessage.value = null;
  }

  onBeforeUnmount(() => stopPolling());

  return {
    // state
    activeBatchId,
    status,
    lastRunResult,
    errors,
    isRunning,
    isCreating,
    errorMessage,

    // actions
    create,
    run,
    createAndRun,
    refreshStatus,
    refreshErrors,
    watch,
    reset,
  };
}

export type { ApiBatchStatusResponse, ApiImportChunkSummary, ApiImportChunkItem };
