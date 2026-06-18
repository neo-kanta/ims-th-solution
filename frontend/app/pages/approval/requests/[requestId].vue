<script setup lang="ts">
import { computed, onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import { useAuthStore } from "~/stores/useAuthStore";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import ApprovalRequestDetail from "~/features/approval/components/ApprovalRequestDetail.vue";
import { useApprovalRequest } from "~/features/approval/composables/useApprovalRequest";
import { useApprovalActions } from "~/features/approval/composables/useApprovalActions";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "APPROVAL_VIEW_REQUEST",
});

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const requestId = computed(() => String(route.params.requestId));
const { detail, loading, error, forbidden, notFound, fetchRequest } = useApprovalRequest();
const actions = useApprovalActions();

const currentUserId = computed(() => authStore.user?.id ?? "");

async function reload() {
  await fetchRequest(requestId.value);
}

async function onApprove(comment: string) {
  const task = detail.value?.viewer_task;
  if (!task?.id) return;
  try {
    await actions.approve(task.id, comment);
    await reload();
  } catch {
    /* error surfaced via actions.error */
  }
}

async function onReject(reason: string) {
  const task = detail.value?.viewer_task;
  if (!task?.id) return;
  try {
    await actions.reject(task.id, reason);
    await reload();
  } catch {
    /* handled */
  }
}

async function onWithdraw() {
  if (!detail.value?.request?.id) return;
  try {
    await actions.withdraw(detail.value.request.id);
    await reload();
  } catch {
    /* handled */
  }
}

async function onRevoke(reason: string) {
  if (!detail.value?.request?.id) return;
  try {
    await actions.revoke(detail.value.request.id, reason);
    await reload();
  } catch {
    /* handled */
  }
}

onMounted(reload);
</script>

<template>
  <section class="approval-request-page">
    <AppPageHeader
      :title="t('approval.request.title', 'Approval Request')"
      :description="t('approval.request.description', 'Review the request, its timeline and sign off.')"
    >
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="router.push('/approval')">
          {{ t('approval.actions.backToInbox', 'Back to inbox') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppLoadingState v-if="loading" :message="t('approval.request.loading', 'Loading request…')" />

    <AppCard v-else-if="forbidden">
      <p class="approval-request-page__state">
        {{ t('approval.errors.noPermission', 'You do not have permission to view this approval request.') }}
      </p>
    </AppCard>

    <AppCard v-else-if="notFound">
      <p class="approval-request-page__state">
        {{ t('approval.errors.notFound', 'Approval request not found.') }}
      </p>
    </AppCard>

    <AppCard v-else-if="error">
      <p class="approval-request-page__state approval-request-page__state--error">{{ error }}</p>
    </AppCard>

    <ApprovalRequestDetail
      v-else-if="detail"
      :detail="detail"
      :submitting="actions.submitting.value"
      :action-error="actions.error.value"
      :current-user-id="currentUserId"
      @approve="onApprove"
      @reject="onReject"
      @withdraw="onWithdraw"
      @revoke="onRevoke"
    />
  </section>
</template>

<style scoped>
.approval-request-page {
  display: grid;
  gap: var(--space-5, 20px);
}
.approval-request-page__state {
  margin: 0;
  color: var(--text-secondary, #57606a);
}
.approval-request-page__state--error {
  color: var(--alert-danger-text, #cf222e);
}
</style>
