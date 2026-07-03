<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";

import { permissionWorkflowApi } from "../services/permissionWorkflowApi";
import type { PermissionChangeRequest } from "../types";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

const requests = ref<PermissionChangeRequest[]>([]);
const total = ref(0);
const loading = ref(false);
const error = ref<string | null>(null);

const filters = reactive({
  status: "",
  risk: "",
  label: "",
  search: "",
});

const draft = reactive({
  title: "",
  description: "",
  request_type: "PERMISSION_CHANGE",
  risk_level: "LOW" as "LOW" | "MEDIUM" | "HIGH" | "CRITICAL",
});

const statusOptions = ["", "DRAFT", "READY_FOR_REVIEW", "CHANGES_REQUESTED", "APPROVED", "REJECTED", "MERGED"];
const riskOptions = ["", "LOW", "MEDIUM", "HIGH", "CRITICAL"];

const hasRows = computed(() => requests.value.length > 0);

async function loadRequests() {
  loading.value = true;
  error.value = null;
  try {
    const response = await permissionWorkflowApi.listRequests({
      status: filters.status || undefined,
      risk: filters.risk || undefined,
      label: filters.label || undefined,
      search: filters.search || undefined,
      page: 1,
      limit: 50,
    });
    requests.value = response.items;
    total.value = response.total;
  } catch (err: any) {
    error.value = err?.data?.error || err?.message || "Unable to load permission requests";
  } finally {
    loading.value = false;
  }
}

async function createDraft() {
  if (!draft.title.trim()) return;
  loading.value = true;
  error.value = null;
  try {
    const created = await permissionWorkflowApi.createRequest({
      title: draft.title.trim(),
      description: draft.description.trim(),
      request_type: draft.request_type,
      risk_level: draft.risk_level,
    });
    draft.title = "";
    draft.description = "";
    await navigateTo(`/permissions/change-requests/${created.id}`);
  } catch (err: any) {
    error.value = err?.data?.error || err?.message || "Unable to create draft";
  } finally {
    loading.value = false;
  }
}

function badgeClass(value: string) {
  if (value === "APPROVED" || value === "MERGED" || value === "LOW") return "badge-success";
  if (value === "READY_FOR_REVIEW" || value === "MEDIUM") return "badge-info";
  if (value === "CHANGES_REQUESTED" || value === "HIGH") return "badge-warning";
  if (value === "REJECTED" || value === "CRITICAL") return "badge-error";
  return "badge-neutral";
}

function currentStep(request: PermissionChangeRequest) {
  return request.steps?.find((step) => step.status === "PENDING")
    || request.steps?.find((step) => step.status === "NOT_STARTED")
    || request.steps?.[request.steps.length - 1];
}

watch(filters, loadRequests);
onMounted(loadRequests);
</script>

<template>
  <main class="permission-list">
    <header class="permission-list__header">
      <div>
        <p class="permission-list__eyebrow">Permission Approval</p>
        <h1>Change Requests</h1>
        <p>Review, validate, and apply permission changes through approval steps.</p>
      </div>
      <form class="permission-list__draft" @submit.prevent="createDraft">
        <input v-model="draft.title" class="input" placeholder="Draft request title" />
        <select v-model="draft.risk_level" class="select">
          <option v-for="risk in riskOptions.filter(Boolean)" :key="risk" :value="risk">{{ risk }}</option>
        </select>
        <button class="btn btn-primary btn-sm" :disabled="loading || !draft.title.trim()">
          New draft
        </button>
      </form>
    </header>

    <section class="permission-list__filters" aria-label="Permission request filters">
      <input v-model="filters.search" class="input" placeholder="Search request number or title" />
      <select v-model="filters.status" class="select">
        <option v-for="status in statusOptions" :key="status || 'all-status'" :value="status">
          {{ status || "All statuses" }}
        </option>
      </select>
      <select v-model="filters.risk" class="select">
        <option v-for="risk in riskOptions" :key="risk || 'all-risk'" :value="risk">
          {{ risk || "All risks" }}
        </option>
      </select>
      <input v-model="filters.label" class="input" placeholder="Label code" />
    </section>

    <p v-if="error" class="alert alert-danger">{{ error }}</p>

    <section class="table-wrap permission-list__table" aria-live="polite">
      <table class="table">
        <thead>
          <tr>
            <th>Request</th>
            <th>Status</th>
            <th>Risk</th>
            <th>Labels</th>
            <th>Requester</th>
            <th>Current step</th>
            <th>Updated</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading && !hasRows">
            <td colspan="7" class="app-table__state-row">
              <AppLoadingState message="Loading permission requests..." />
            </td>
          </tr>
          <tr v-else-if="!hasRows">
            <td colspan="7" class="is-muted">No permission change requests found.</td>
          </tr>
          <tr v-for="request in requests" :key="request.id">
            <td>
              <NuxtLink class="permission-list__request" :to="`/permissions/change-requests/${request.id}`">
                <span>{{ request.request_no }}</span>
                <strong>{{ request.title }}</strong>
              </NuxtLink>
            </td>
            <td><span class="badge" :class="badgeClass(request.status)">{{ request.status }}</span></td>
            <td><span class="badge" :class="badgeClass(request.risk_level)">{{ request.risk_level }}</span></td>
            <td>
              <div class="permission-list__labels">
                <span v-for="label in request.labels" :key="label.id" class="permission-label">
                  {{ label.label_code }}
                </span>
              </div>
            </td>
            <td>{{ request.created_by_name || request.created_by }}</td>
            <td>{{ currentStep(request)?.step_name || "-" }}</td>
            <td>{{ new Date(request.updated_at).toLocaleString() }}</td>
          </tr>
        </tbody>
      </table>
    </section>

    <footer class="permission-list__footer">
      {{ total }} requests
    </footer>
  </main>
</template>

<style scoped>
.permission-list {
  display: grid;
  gap: var(--space-6);
}

.permission-list__header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(320px, 520px);
  gap: var(--space-6);
  align-items: end;
}

.permission-list__eyebrow {
  margin: 0 0 var(--space-2);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
}

.permission-list__draft,
.permission-list__filters {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 140px auto;
  gap: var(--space-3);
}

.permission-list__filters {
  grid-template-columns: minmax(240px, 1fr) 180px 160px 180px;
  align-items: center;
}

.permission-list__request {
  display: grid;
  gap: var(--space-1);
  color: var(--text-primary);
}

.permission-list__request span {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.permission-list__labels {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.permission-label {
  min-height: 22px;
  display: inline-flex;
  align-items: center;
  padding: 0 var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-pill);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.permission-list__footer {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

@media (max-width: 960px) {
  .permission-list__header,
  .permission-list__draft,
  .permission-list__filters {
    grid-template-columns: 1fr;
  }
}

.app-table__state-row {
  text-align: center;
  background: transparent !important;
}
</style>
