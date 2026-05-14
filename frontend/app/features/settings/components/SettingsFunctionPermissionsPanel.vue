<script setup lang="ts">
import type {
  FunctionPermissionRow,
  PermissionActionColumn,
} from "../ui.types";

defineProps<{
  actions: PermissionActionColumn[];
  rows: FunctionPermissionRow[];
  sessionPermissionCount: number;
}>();

const { t } = useI18n();
</script>

<template>
  <section
    class="settings-panel settings-permissions"
    aria-labelledby="settings-permissions-title"
  >
    <header class="settings-panel__header">
      <div>
        <h2 id="settings-permissions-title" class="settings-panel__title">
          {{ t("settings.console.functionPermissions.title") }}
        </h2>
        <p class="settings-panel__subtitle">
          {{ t("settings.console.functionPermissions.subtitle") }}
        </p>
      </div>
      <span class="badge badge-info">
        {{ t("settings.console.functionPermissions.sessionGrants", { count: sessionPermissionCount }) }}
      </span>
    </header>

    <div class="alert alert-warning settings-panel__alert" role="note">
      <strong>{{ t("settings.console.common.sampleDataTitle") }}</strong>
      <span>{{ t("settings.console.functionPermissions.sampleNotice") }}</span>
    </div>

    <div class="settings-permissions__body">
      <div class="table-wrap settings-permissions__matrix">
        <table class="table">
          <thead>
            <tr>
              <th class="settings-permissions__sticky-col">
                {{ t("settings.console.functionPermissions.menuFunction") }}
              </th>
              <th>{{ t("settings.console.functionPermissions.owner") }}</th>
              <th
                v-for="action in actions"
                :key="action.key"
                class="settings-permissions__action-col"
              >
                {{ action.label }}
              </th>
              <th>{{ t("settings.console.functionPermissions.source") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td class="settings-permissions__sticky-col">
                <div class="settings-record-primary">{{ row.menu }}</div>
                <div class="settings-record-secondary">{{ row.functionName }}</div>
              </td>
              <td>{{ row.owner }}</td>
              <td
                v-for="action in actions"
                :key="`${row.id}-${action.key}`"
                class="settings-permissions__cell"
              >
                <span
                  class="settings-permissions__state"
                  :class="{ 'is-granted': row.permissions[action.key] }"
                >
                  <AppIcon
                    v-if="row.permissions[action.key]"
                    name="check"
                    size="xs"
                  />
                  <span v-else aria-hidden="true">-</span>
                  <span class="settings-permissions__sr">
                    {{
                      row.permissions[action.key]
                        ? t("settings.console.functionPermissions.granted")
                        : t("settings.console.functionPermissions.notGranted")
                    }}
                  </span>
                </span>
              </td>
              <td>
                <span
                  class="badge"
                  :class="row.status === 'live-session' ? 'badge-success' : 'badge-warning'"
                >
                  {{
                    row.status === "live-session"
                      ? t("settings.console.common.session")
                      : t("settings.console.common.demo")
                  }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>

<style scoped>
.settings-permissions__body {
  padding: var(--space-5);
}

.settings-permissions__matrix {
  max-height: 34rem;
  border-radius: var(--radius-md);
}

.settings-permissions__matrix .table {
  min-width: 1180px;
}

.settings-permissions__matrix thead th {
  position: sticky;
  top: 0;
  z-index: 2;
}

.settings-permissions__sticky-col {
  position: sticky;
  left: 0;
  z-index: 1;
  min-width: 15rem;
  background: var(--bg-card);
}

.settings-permissions__matrix thead .settings-permissions__sticky-col {
  z-index: 3;
  background: var(--bg-table-header);
}

.settings-permissions__action-col,
.settings-permissions__cell {
  text-align: center;
}

.settings-permissions__state {
  width: 1.75rem;
  height: 1.75rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
  color: var(--text-tertiary);
}

.settings-permissions__state.is-granted {
  border-color: var(--border-success);
  background: var(--status-approved-bg);
  color: var(--status-approved-text);
}

.settings-record-primary {
  display: block;
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}

.settings-record-secondary {
  display: block;
  margin-top: var(--space-1);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-relaxed);
}

.settings-permissions__sr {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
}
</style>
