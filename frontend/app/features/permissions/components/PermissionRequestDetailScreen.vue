<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import { permissionWorkflowApi } from "../services/permissionWorkflowApi";
import type { PermissionChangeRequest } from "../types";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

const props = defineProps<{
  requestId: string;
}>();

const authStore = useAuthStore();
const request = ref<PermissionChangeRequest | null>(null);
const activeTab = ref("conversation");
const loading = ref(false);
const actionError = ref<string | null>(null);
const comment = ref("");
const reviewComment = ref("");

const tabs = [
  { key: "conversation", label: "Conversation" },
  { key: "changes", label: "Changes" },
  { key: "checks", label: "Checks" },
  { key: "steps", label: "Approval Steps" },
  { key: "reviewers", label: "Reviewers / Approvers" },
  { key: "audit", label: "Audit Trail" },
  { key: "settings", label: "Settings" },
];

const currentUserId = computed(() => authStore.user?.id ?? "");
const pendingStep = computed(() => request.value?.steps?.find((step) => step.status === "PENDING"));
const failedChecks = computed(() => request.value?.checks?.filter((check) => check.status === "FAILED") ?? []);
const hasBlockingCheck = computed(() => failedChecks.value.some((check) => check.severity === "BLOCKER"));
const isCreator = computed(() => request.value?.created_by === currentUserId.value);

const canSubmit = computed(() => request.value?.status === "DRAFT" || request.value?.status === "CHANGES_REQUESTED");
const canReview = computed(() => request.value?.status === "READY_FOR_REVIEW" && !isCreator.value && Boolean(pendingStep.value));
const canMerge = computed(() => request.value?.status === "APPROVED" && !hasBlockingCheck.value);
const rejectReasonValid = computed(() => reviewComment.value.trim().length > 0);

async function loadRequest() {
  loading.value = true;
  actionError.value = null;
  try {
    request.value = await permissionWorkflowApi.getRequest(props.requestId);
  } catch (err: any) {
    actionError.value = err?.data?.error || err?.message || "Unable to load permission request";
  } finally {
    loading.value = false;
  }
}

async function runAction(action: () => Promise<PermissionChangeRequest | unknown>) {
  loading.value = true;
  actionError.value = null;
  try {
    await action();
    await loadRequest();
  } catch (err: any) {
    actionError.value = err?.data?.error || err?.message || "Action failed";
  } finally {
    loading.value = false;
  }
}

function submit() {
  return runAction(() => permissionWorkflowApi.submit(props.requestId));
}

function approve() {
  return runAction(() => permissionWorkflowApi.approve(props.requestId, reviewComment.value, pendingStep.value?.id));
}

function requestChanges() {
  return runAction(() => permissionWorkflowApi.requestChanges(props.requestId, reviewComment.value, pendingStep.value?.id));
}

function reject() {
  return runAction(() => permissionWorkflowApi.reject(props.requestId, reviewComment.value, pendingStep.value?.id));
}

function merge() {
  return runAction(() => permissionWorkflowApi.merge(props.requestId));
}

function rerunChecks() {
  return runAction(() => permissionWorkflowApi.rerunChecks(props.requestId));
}

async function addComment() {
  if (!comment.value.trim()) return;
  await runAction(() => permissionWorkflowApi.addComment(props.requestId, comment.value.trim()));
  comment.value = "";
}

function badgeClass(value?: string) {
  if (value === "APPROVED" || value === "MERGED" || value === "PASSED" || value === "LOW") return "badge-success";
  if (value === "READY_FOR_REVIEW" || value === "PENDING" || value === "WARNING" || value === "MEDIUM") return "badge-info";
  if (value === "CHANGES_REQUESTED" || value === "HIGH") return "badge-warning";
  if (value === "REJECTED" || value === "FAILED" || value === "CRITICAL") return "badge-error";
  return "badge-neutral";
}

function pretty(value: unknown) {
  return JSON.stringify(value || {}, null, 2);
}

onMounted(loadRequest);
</script>

<template>
  <main class="permission-detail">
    <NuxtLink class="permission-detail__back" to="/permissions/change-requests">
      Back to change requests
    </NuxtLink>

    <p v-if="actionError" class="alert alert-danger">{{ actionError }}</p>
    <AppLoadingState v-if="loading && !request" message="Loading permission request..." />

    <template v-if="request">
      <header class="permission-detail__header">
        <div>
          <div class="permission-detail__meta">
            <span>{{ request.request_no }}</span>
            <span class="badge" :class="badgeClass(request.status)">{{ request.status }}</span>
            <span class="badge" :class="badgeClass(request.risk_level)">{{ request.risk_level }}</span>
          </div>
          <h1>{{ request.title }}</h1>
          <p>{{ request.description || "No description supplied." }}</p>
          <div class="permission-detail__labels">
            <span v-for="label in request.labels" :key="label.id" class="permission-label">
              {{ label.label_code }}
            </span>
          </div>
        </div>
        <div class="permission-detail__actions">
          <button class="btn btn-secondary btn-sm" :disabled="!canSubmit || loading" @click="submit">
            Ready for Review
          </button>
          <button class="btn btn-success btn-sm" :disabled="!canReview || loading" @click="approve">
            Approve
          </button>
          <button class="btn btn-warning btn-sm" :disabled="!canReview || loading" @click="requestChanges">
            Request Changes
          </button>
          <button class="btn btn-danger btn-sm" :disabled="!canReview || loading || !rejectReasonValid" @click="reject">
            Reject
          </button>
          <button class="btn btn-primary btn-sm" :disabled="!canMerge || loading" @click="merge">
            Merge / Apply
          </button>
        </div>
      </header>

      <section class="permission-detail__review">
        <label class="label" for="review-comment">Reviewer / Approver comment</label>
        <textarea id="review-comment" v-model="reviewComment" class="textarea" rows="2" />
      </section>

      <nav class="tabs permission-detail__tabs" aria-label="Permission request detail tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="tab"
          :class="{ 'is-active': activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </nav>

      <section v-if="activeTab === 'conversation'" class="permission-detail__panel">
        <div class="permission-detail__summary">
          <span>Creator: {{ request.created_by_name || "Unknown User" }}</span>
          <span>Target: {{ request.target_entity_type || "-" }}</span>
          <span>Updated: {{ new Date(request.updated_at).toLocaleString() }}</span>
        </div>
        <form class="permission-detail__comment-form" @submit.prevent="addComment">
          <textarea v-model="comment" class="textarea" placeholder="Add approval discussion" />
          <button class="btn btn-secondary btn-sm" :disabled="loading || !comment.trim()">Comment</button>
        </form>
        <ol class="permission-timeline">
          <li v-for="event in request.events" :key="event.id">
            <strong>{{ event.event_type }}</strong>
            <span>{{ event.actor_name || "System" }} | {{ new Date(event.created_at).toLocaleString() }}</span>
            <p v-if="event.comment">{{ event.comment }}</p>
          </li>
          <li v-for="item in request.comments" :key="item.id">
            <strong>{{ item.user_name }}</strong>
            <span>{{ new Date(item.created_at).toLocaleString() }}</span>
            <p>{{ item.comment }}</p>
          </li>
        </ol>
      </section>

      <section v-else-if="activeTab === 'changes'" class="permission-detail__panel">
        <div v-for="item in request.items" :key="item.id" class="permission-diff">
          <header>
            <strong>{{ item.action_type }}</strong>
            <span>{{ item.target_table }}</span>
          </header>
          <div class="permission-diff__grid">
            <pre>{{ pretty(item.before_json) }}</pre>
            <pre>{{ pretty(item.after_json) }}</pre>
          </div>
        </div>
      </section>

      <section v-else-if="activeTab === 'checks'" class="permission-detail__panel">
        <button class="btn btn-secondary btn-sm" :disabled="loading" @click="rerunChecks">Rerun checks</button>
        <div class="permission-checks">
          <article v-for="check in request.checks" :key="check.id" class="permission-check">
            <span class="badge" :class="badgeClass(check.status)">{{ check.status }}</span>
            <div>
              <strong>{{ check.check_name }}</strong>
              <p>{{ check.message }}</p>
            </div>
            <span>{{ check.severity }}</span>
          </article>
        </div>
      </section>

      <section v-else-if="activeTab === 'steps'" class="permission-detail__panel">
        <ol class="permission-steps">
          <li v-for="step in request.steps" :key="step.id" :class="`is-${step.status.toLowerCase()}`">
            <span>{{ step.step_no }}</span>
            <div>
              <strong>{{ step.step_name }}</strong>
              <p>{{ step.approvals_received }} / {{ step.min_approvals_required }} approvals | {{ step.approval_mode }}</p>
            </div>
            <span class="badge" :class="badgeClass(step.status)">{{ step.status }}</span>
          </li>
        </ol>
      </section>

      <section v-else-if="activeTab === 'reviewers'" class="permission-detail__panel">
        <div v-for="step in request.steps" :key="step.id" class="permission-reviewers">
          <h3>{{ step.step_name }}</h3>
          <article v-for="approver in step.approvers" :key="approver.id">
            <strong>{{ approver.approver_display_name || approver.approver_role_code }}</strong>
            <span class="badge" :class="badgeClass(approver.approval_status)">{{ approver.approval_status }}</span>
            <p>{{ approver.approval_comment }}</p>
          </article>
        </div>
      </section>

      <section v-else-if="activeTab === 'audit'" class="permission-detail__panel">
        <ol class="permission-timeline">
          <li v-for="event in request.events" :key="event.id">
            <strong>{{ event.event_type }}</strong>
            <span>{{ new Date(event.created_at).toLocaleString() }}</span>
          </li>
        </ol>
      </section>

      <section v-else class="permission-detail__panel">
        <dl class="permission-settings">
          <div><dt>Request type</dt><dd>{{ request.request_type }}</dd></div>
          <div><dt>Risk level</dt><dd>{{ request.risk_level }}</dd></div>
          <div><dt>Creator self approval</dt><dd>Blocked by default</dd></div>
          <div><dt>Failed checks</dt><dd>Block submit and merge when severity is blocker</dd></div>
        </dl>
      </section>
    </template>
  </main>
</template>

<style scoped>
.permission-detail {
  display: grid;
  gap: var(--space-6);
}

.permission-detail__back {
  width: fit-content;
  color: var(--text-secondary);
}

.permission-detail__header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-6);
  align-items: start;
}

.permission-detail__meta,
.permission-detail__labels,
.permission-detail__actions,
.permission-detail__summary {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  align-items: center;
}

.permission-detail__meta {
  margin-bottom: var(--space-3);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.permission-detail__actions {
  justify-content: flex-end;
  max-width: 420px;
}

.btn-success {
  background: var(--state-success);
  border-color: var(--state-success);
  color: var(--text-inverse);
}

.btn-warning {
  background: var(--state-warning);
  border-color: var(--state-warning);
  color: var(--text-inverse);
}

.permission-detail__review {
  display: grid;
  gap: var(--space-3);
}

.permission-detail__tabs {
  overflow-x: auto;
}

.permission-detail__panel {
  display: grid;
  gap: var(--space-5);
}

.permission-detail__comment-form {
  display: grid;
  gap: var(--space-3);
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

.permission-timeline,
.permission-steps {
  display: grid;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.permission-timeline li,
.permission-steps li,
.permission-check,
.permission-reviewers article {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
}

.permission-timeline span,
.permission-steps p,
.permission-check p,
.permission-reviewers p {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.permission-steps li {
  grid-template-columns: 32px minmax(0, 1fr) auto;
  align-items: center;
}

.permission-steps li > span:first-child {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border-radius: 999px;
  background: var(--bg-selected);
  color: var(--action-primary);
  font-weight: var(--font-weight-semibold);
}

.permission-check {
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: start;
}

.permission-diff {
  display: grid;
  gap: var(--space-3);
}

.permission-diff header {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
}

.permission-diff__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.permission-diff pre {
  min-height: 180px;
  max-height: 360px;
  overflow: auto;
  margin: 0;
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card-muted);
  color: var(--text-primary);
  font-size: var(--font-size-xs);
}

.permission-settings {
  display: grid;
  gap: var(--space-3);
}

.permission-settings div {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: var(--space-4);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--border-subtle);
}

.permission-settings dt {
  color: var(--text-secondary);
}

.permission-settings dd {
  margin: 0;
}

@media (max-width: 900px) {
  .permission-detail__header,
  .permission-diff__grid,
  .permission-settings div {
    grid-template-columns: 1fr;
  }

  .permission-detail__actions {
    justify-content: stretch;
    max-width: none;
  }
}
</style>
