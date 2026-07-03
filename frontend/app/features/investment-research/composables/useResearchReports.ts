import { ref } from "vue";

import { useI18n } from "~/composables/useI18n";

import { researchReportApi } from "../services/researchReportApi";
import type {
  CreateResearchReportInput,
  ResearchReport,
  ResearchReportFilters,
  ResearchReportListPayload,
  UpdateResearchReportInput,
} from "../types";

function extractErrorMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const data = (err as { data?: { error?: unknown } }).data;
  if (data && typeof data.error === "string" && data.error.trim()) {
    return data.error;
  }
  const message = (err as { message?: unknown }).message;
  if (typeof message === "string" && message.trim()) {
    return message;
  }
  return fallback;
}

export function useResearchReportsList() {
  const { t } = useI18n();
  const items = ref<ResearchReport[]>([]);
  const total = ref(0);
  const page = ref(1);
  const limit = ref(20);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchList(filters: ResearchReportFilters = {}) {
    loading.value = true;
    error.value = null;
    try {
      const payload: ResearchReportListPayload = await researchReportApi.list({
        page: page.value,
        limit: limit.value,
        ...filters,
      });
      items.value = payload.items ?? [];
      total.value = payload.total ?? 0;
      page.value = payload.page ?? page.value;
      limit.value = payload.limit ?? limit.value;
    } catch (err) {
      error.value = extractErrorMessage(err, t("investmentResearch.errors.loadList"));
      items.value = [];
      total.value = 0;
    } finally {
      loading.value = false;
    }
  }

  return {
    items,
    total,
    page,
    limit,
    loading,
    error,
    fetchList,
  };
}

export function useResearchReportDetail() {
  const { t } = useI18n();
  const report = ref<ResearchReport | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetch(id: string) {
    loading.value = true;
    error.value = null;
    try {
      report.value = await researchReportApi.get(id);
    } catch (err) {
      report.value = null;
      error.value = extractErrorMessage(err, t("investmentResearch.errors.loadDetail"));
    } finally {
      loading.value = false;
    }
  }

  return { report, loading, error, fetch };
}

export function useResearchReportMutation() {
  const { t } = useI18n();
  const saving = ref(false);
  const error = ref<string | null>(null);

  async function create(input: CreateResearchReportInput) {
    saving.value = true;
    error.value = null;
    try {
      return await researchReportApi.create(input);
    } catch (err) {
      error.value = extractErrorMessage(err, t("investmentResearch.errors.create"));
      throw err;
    } finally {
      saving.value = false;
    }
  }

  async function update(id: string, input: UpdateResearchReportInput) {
    saving.value = true;
    error.value = null;
    try {
      return await researchReportApi.update(id, input);
    } catch (err) {
      error.value = extractErrorMessage(err, t("investmentResearch.errors.update"));
      throw err;
    } finally {
      saving.value = false;
    }
  }

  async function remove(id: string) {
    saving.value = true;
    error.value = null;
    try {
      await researchReportApi.remove(id);
    } catch (err) {
      error.value = extractErrorMessage(err, t("investmentResearch.errors.delete"));
      throw err;
    } finally {
      saving.value = false;
    }
  }

  async function submit(id: string) {
    saving.value = true;
    error.value = null;
    try {
      return await researchReportApi.submit(id);
    } catch (err) {
      error.value = extractErrorMessage(err, t("investmentResearch.errors.submit"));
      throw err;
    } finally {
      saving.value = false;
    }
  }

  async function cancelSubmit(id: string) {
    saving.value = true;
    error.value = null;
    try {
      return await researchReportApi.cancelSubmit(id);
    } catch (err) {
      error.value = extractErrorMessage(
        err,
        t("investmentResearch.errors.cancelSubmit"),
      );
      throw err;
    } finally {
      saving.value = false;
    }
  }

  return {
    saving,
    error,
    create,
    update,
    remove,
    submit,
    cancelSubmit,
  };
}
