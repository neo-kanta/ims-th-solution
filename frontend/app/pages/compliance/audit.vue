<script setup lang="ts">
import { onMounted, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import ComplianceAuditTimeline from "~/features/compliance/components/ComplianceAuditTimeline.vue";
import { useComplianceCheckGroup } from "~/features/compliance/composables/useComplianceChecks";
import { isUuid } from "~/features/compliance/lib/formatters";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const group = useComplianceCheckGroup();

const groupId = ref<string>(
  typeof route.query.group === "string" ? route.query.group : "",
);
const formError = ref<string | null>(null);

async function load(id: string) {
  formError.value = null;
  if (!id) return;
  if (!isUuid(id)) {
    formError.value = "Check group id must be a UUID.";
    return;
  }
  try {
    await group.fetchGroup(id);
  } catch {
    /* error in composable */
  }
}

function submit() {
  // Reflect the lookup in the URL so the view is shareable.
  void router.replace({ path: "/compliance/audit", query: { group: groupId.value } });
  void load(groupId.value.trim());
}

watch(
  () => route.query.group,
  (next) => {
    if (typeof next === "string" && next !== groupId.value) {
      groupId.value = next;
      void load(next);
    }
  },
);

onMounted(() => {
  if (groupId.value) void load(groupId.value);
});
</script>

<template>
  <section class="audit-page">
    <AppPageHeader
      :title="t('compliance.audit.title')"
      :description="t('compliance.audit.description')"
    >
      <template #actions>
        <AppButton
          variant="secondary"
          size="sm"
          :disabled="true"
          :title="t('compliance.audit.export.disabled')"
        >
          {{ t("compliance.audit.export.label") }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard :title="t('compliance.audit.lookup.title')">
      <form class="audit-lookup" @submit.prevent="submit">
        <label class="audit-lookup__field">
          <span class="audit-lookup__label">
            {{ t("compliance.audit.lookup.groupIdLabel") }}
          </span>
          <input
            v-model="groupId"
            type="text"
            class="form-control"
            autocomplete="off"
            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
          />
        </label>
        <AppButton
          variant="primary"
          size="sm"
          :loading="group.loading.value"
          @click="submit"
        >
          {{ t("compliance.audit.lookup.cta") }}
        </AppButton>
        <span v-if="formError" class="audit-lookup__error" role="alert">
          {{ formError }}
        </span>
      </form>
    </AppCard>

    <ComplianceAuditTimeline
      :data="group.result.value"
      :loading="group.loading.value"
      :error="group.error.value"
    />
  </section>
</template>

<style scoped>
.audit-page {
  display: grid;
  gap: var(--space-7);
}

.audit-lookup {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: var(--space-3);
  align-items: end;
}

.audit-lookup__field {
  display: grid;
  gap: var(--space-2);
}

.audit-lookup__label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.form-control {
  width: 100%;
  height: var(--size-control-md);
  padding: 0 var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-family: var(--font-family-mono);
}

.form-control:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.audit-lookup__error {
  grid-column: 1 / -1;
  color: var(--state-danger);
  font-size: var(--font-size-xs);
}
</style>
