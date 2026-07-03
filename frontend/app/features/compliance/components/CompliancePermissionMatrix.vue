<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

interface MatrixRow {
  labelKey: string;
  code: string;
  /** True if backend already wires this code; false if it's planned (Phase 2 + Missing API). */
  shipped: boolean;
}

const { t } = useI18n();
const authStore = useAuthStore();

const ROWS: readonly MatrixRow[] = [
  { labelKey: "compliance.permissions.table.viewRules", code: "IRG_VIEW_RULES", shipped: true },
  { labelKey: "compliance.permissions.table.createRule", code: "IRG_EDIT_RULE_INSTANCE", shipped: true },
  { labelKey: "compliance.permissions.table.editBinding", code: "IRG_EDIT_BINDING", shipped: false },
  { labelKey: "compliance.permissions.table.overrideBreach", code: "IRG_OVERRIDE_BREACH", shipped: true },
  { labelKey: "compliance.permissions.table.adminRuleType", code: "IRG_ADMIN_RULE_TYPE", shipped: false },
  { labelKey: "compliance.permissions.table.runPreTrade", code: "WORKFLOW_EXECUTE", shipped: true },
  { labelKey: "compliance.permissions.table.viewExceptions", code: "IRG_VIEW_EXCEPTIONS", shipped: false },
  { labelKey: "compliance.permissions.table.requestException", code: "IRG_REQUEST_EXCEPTION", shipped: false },
  { labelKey: "compliance.permissions.table.approveException", code: "IRG_APPROVE_EXCEPTION", shipped: false },
] as const;

const rows = computed(() =>
  ROWS.map((r) => ({
    ...r,
    held: authStore.hasPermission(r.code),
  })),
);
</script>

<template>
  <AppCard
    :title="t('compliance.permissions.title')"
    :subtitle="t('compliance.permissions.description')"
  >
    <div class="perm-table-wrap">
      <table class="perm-table">
        <thead>
          <tr>
            <th>{{ t("compliance.permissions.action") }}</th>
            <th>{{ t("compliance.permissions.code") }}</th>
            <th class="perm-table__held">
              {{ t("compliance.permissions.held") }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row.code">
            <td>
              <div class="perm-table__label">{{ t(row.labelKey as any) }}</div>
              <div v-if="!row.shipped" class="perm-table__planned">
                Planned · backend not wired yet
              </div>
            </td>
            <td>
              <code class="perm-table__code">{{ row.code }}</code>
            </td>
            <td class="perm-table__held">
              <span
                v-if="row.held"
                class="perm-pill perm-pill--yes"
                :title="t('compliance.permissions.youHave')"
              >
                <AppIcon name="check" size="xs" />
                Yes
              </span>
              <span
                v-else
                class="perm-pill perm-pill--no"
                :title="t('compliance.permissions.youDontHave')"
              >
                No
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </AppCard>
</template>

<style scoped>
.perm-table-wrap {
  overflow-x: auto;
}

.perm-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
}

.perm-table thead th {
  background: var(--bg-table-header);
  color: var(--text-secondary);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-xs);
  text-align: left;
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.perm-table tbody td {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
}

.perm-table__held {
  text-align: right;
  width: 1%;
  white-space: nowrap;
}

.perm-table__label {
  font-weight: var(--font-weight-medium);
}

.perm-table__planned {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  font-style: italic;
}

.perm-table__code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.perm-pill {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  height: 22px;
  padding: 0 var(--space-3);
  border-radius: var(--radius-pill);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  border: 1px solid var(--border-subtle);
}

.perm-pill--yes {
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
  border-color: var(--alert-success-border);
}

.perm-pill--no {
  background: var(--bg-card-muted);
  color: var(--text-tertiary);
}
</style>
