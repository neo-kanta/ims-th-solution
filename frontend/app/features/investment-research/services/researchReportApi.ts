import { useApi } from "~/composables/useApi";

import type {
  CreateResearchReportInput,
  ResearchReport,
  ResearchReportFilters,
  ResearchReportListPayload,
  UpdateResearchReportInput,
} from "../types";

interface SuccessEnvelope<T> {
  data: T;
  message?: string;
}

function unwrap<T>(envelope: SuccessEnvelope<T>): T {
  return envelope.data;
}

function buildQuery(filters: ResearchReportFilters = {}): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value === undefined || value === null) continue;
    const str = typeof value === "string" ? value.trim() : String(value);
    if (str === "") continue;
    params.set(key, str);
  }
  const query = params.toString();
  return query ? `?${query}` : "";
}

export const researchReportApi = {
  async list(filters: ResearchReportFilters = {}): Promise<ResearchReportListPayload> {
    const { apiFetch } = useApi();
    const response = await apiFetch<SuccessEnvelope<ResearchReportListPayload>>(
      `/investment/research-reports${buildQuery(filters)}`,
    );
    return unwrap(response);
  },

  async get(id: string): Promise<ResearchReport> {
    const { apiFetch } = useApi();
    const response = await apiFetch<SuccessEnvelope<ResearchReport>>(
      `/investment/research-reports/${id}`,
    );
    return unwrap(response);
  },

  async create(payload: CreateResearchReportInput): Promise<ResearchReport> {
    const { apiFetch } = useApi();
    const response = await apiFetch<SuccessEnvelope<ResearchReport>>(
      `/investment/research-reports`,
      {
        method: "POST",
        body: payload,
      },
    );
    return unwrap(response);
  },

  async update(id: string, payload: UpdateResearchReportInput): Promise<ResearchReport> {
    const { apiFetch } = useApi();
    const response = await apiFetch<SuccessEnvelope<ResearchReport>>(
      `/investment/research-reports/${id}`,
      {
        method: "PUT",
        body: payload,
      },
    );
    return unwrap(response);
  },

  async remove(id: string): Promise<void> {
    const { apiFetch } = useApi();
    await apiFetch<void>(`/investment/research-reports/${id}`, {
      method: "DELETE",
    });
  },

  async submit(id: string): Promise<ResearchReport> {
    const { apiFetch } = useApi();
    const response = await apiFetch<SuccessEnvelope<ResearchReport>>(
      `/investment/research-reports/${id}/submit`,
      {
        method: "POST",
      },
    );
    return unwrap(response);
  },

  async cancelSubmit(id: string): Promise<ResearchReport> {
    const { apiFetch } = useApi();
    const response = await apiFetch<SuccessEnvelope<ResearchReport>>(
      `/investment/research-reports/${id}/cancel-submit`,
      {
        method: "POST",
      },
    );
    return unwrap(response);
  },
};
