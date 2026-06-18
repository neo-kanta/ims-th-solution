<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppCard from "~/shared/ui/AppCard.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "APPROVAL_CONFIG_VIEW",
});

const { t } = useI18n();

const links = [
  {
    to: "/approval/config/processes",
    title: t("approval.config.processes.title", "Approval processes"),
    desc: t("approval.config.processes.description", "Configure approval processes and stages by type and contract."),
  },
  {
    to: "/approval/config/groups",
    title: t("approval.config.groups.title", "Approval groups"),
    desc: t("approval.config.groups.description", "Manage reusable approver groups and members."),
  },
  {
    to: "/approval/config/teams",
    title: t("approval.config.teams.title", "Approval teams"),
    desc: t("approval.config.teams.description", "Configure per contract/fund approval teams."),
  },
];
</script>

<template>
  <section class="approval-config">
    <AppPageHeader
      :title="t('approval.config.title', 'Approval Configuration')"
      :description="t('approval.config.description', 'Configure approval processes, groups and teams.')"
    />
    <div class="approval-config__grid">
      <NuxtLink v-for="link in links" :key="link.to" :to="link.to" class="approval-config__link">
        <AppCard :title="link.title">
          <p class="approval-config__desc">{{ link.desc }}</p>
        </AppCard>
      </NuxtLink>
    </div>
  </section>
</template>

<style scoped>
.approval-config {
  display: grid;
  gap: var(--space-5, 20px);
}
.approval-config__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--space-4, 16px);
}
.approval-config__link {
  text-decoration: none;
  color: inherit;
}
.approval-config__desc {
  margin: 0;
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-sm, 0.875rem);
}
</style>
