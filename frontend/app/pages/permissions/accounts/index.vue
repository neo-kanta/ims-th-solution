<script setup lang="ts">
import { onMounted, ref } from "vue";
import { permissionWorkflowApi } from "~/features/permissions/services/permissionWorkflowApi";
import type { PermissionUserSummary } from "~/features/permissions/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.users.view",
});

const users = ref<PermissionUserSummary[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const columns = [
  { key: "user", label: "User" },
  { key: "email", label: "Email" },
  { key: "status", label: "Status" },
  { key: "groups", label: "Groups" },
  { key: "roles", label: "Roles" },
  { key: "effective", label: "Effective" },
];

async function loadUsers() {
  loading.value = true;
  error.value = null;
  try {
    const response = await permissionWorkflowApi.listUsers({ page: 1, limit: 100 });
    users.value = response.items;
  } catch (err: any) {
    error.value = err?.data?.error || err?.message || "Unable to load accounts";
  } finally {
    loading.value = false;
  }
}

onMounted(loadUsers);
</script>

<template>
  <AppPage
    title="Accounts"
    subtitle="Read-only account directory for permission review and workflow targeting."
    :loading="loading"
    :error="error"
    @retry="loadUsers"
  >
    <template #eyebrow>
      <span class="permission-directory__eyebrow">Permission Management</span>
    </template>

    <AppDataTable
      :columns="columns"
      :items="users"
    >
      <!-- Custom User Column -->
      <template #cell(user)="{ item }">
        <div class="user-cell">
          <strong>{{ item.display_name || item.username }}</strong>
          <span class="user-cell__username">{{ item.username }}</span>
        </div>
      </template>

      <!-- Custom Status Column -->
      <template #cell(status)="{ item }">
        <AppStatusBadge :status="item.is_active && !item.is_locked ? 'active' : 'locked'" />
      </template>

      <!-- Custom Groups Column -->
      <template #cell(groups)="{ item }">
        {{ item.groups.join(", ") || "—" }}
      </template>

      <!-- Custom Roles Column -->
      <template #cell(roles)="{ item }">
        {{ item.roles.join(", ") || "—" }}
      </template>

      <!-- Custom Effective Link Column -->
      <template #cell(effective)="{ item }">
        <NuxtLink :to="`/permissions/effective/${item.id}`" class="effective-link">
          View Permissions
        </NuxtLink>
      </template>
    </AppDataTable>
  </AppPage>
</template>

<style scoped>
.permission-directory__eyebrow {
  color: var(--text-tertiary, #6e7781);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
  text-transform: uppercase;
}

.user-cell {
  display: flex;
  flex-direction: column;
  line-height: 1.3;
}

.user-cell__username {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-secondary, #57606a);
}

.effective-link {
  color: var(--text-link, #0969da);
  font-weight: var(--font-weight-medium, 500);
  text-decoration: none;
}

.effective-link:hover {
  text-decoration: underline;
}
</style>
