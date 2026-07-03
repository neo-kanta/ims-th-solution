import { ref } from "vue";

import { approvalApi, approvalErrorMessage, approvalErrorStatus } from "../services/approvalApi";
import type {
  ApprovalGroup,
  ApprovalProcessConfig,
  ApprovalTeam,
} from "../types";

/** Loads approval groups. */
export function useApprovalGroups() {
  const groups = ref<ApprovalGroup[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const forbidden = ref(false);

  async function fetchGroups() {
    loading.value = true;
    error.value = null;
    forbidden.value = false;
    try {
      groups.value = await approvalApi.listGroups();
    } catch (err) {
      if (approvalErrorStatus(err) === 403) forbidden.value = true;
      error.value = approvalErrorMessage(err, "Failed to load approval groups.");
      groups.value = [];
    } finally {
      loading.value = false;
    }
  }

  return { groups, loading, error, forbidden, fetchGroups };
}

/** Loads approval teams. */
export function useApprovalTeams() {
  const teams = ref<ApprovalTeam[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const forbidden = ref(false);

  async function fetchTeams() {
    loading.value = true;
    error.value = null;
    forbidden.value = false;
    try {
      teams.value = await approvalApi.listTeams();
    } catch (err) {
      if (approvalErrorStatus(err) === 403) forbidden.value = true;
      error.value = approvalErrorMessage(err, "Failed to load approval teams.");
      teams.value = [];
    } finally {
      loading.value = false;
    }
  }

  return { teams, loading, error, forbidden, fetchTeams };
}

/** Loads approval process configurations. */
export function useApprovalProcesses() {
  const processes = ref<ApprovalProcessConfig[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const forbidden = ref(false);

  async function fetchProcesses() {
    loading.value = true;
    error.value = null;
    forbidden.value = false;
    try {
      processes.value = await approvalApi.listProcesses();
    } catch (err) {
      if (approvalErrorStatus(err) === 403) forbidden.value = true;
      error.value = approvalErrorMessage(err, "Failed to load approval processes.");
      processes.value = [];
    } finally {
      loading.value = false;
    }
  }

  return { processes, loading, error, forbidden, fetchProcesses };
}
