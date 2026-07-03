<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "~/composables/useI18n";
import AppIcon from "~/shared/ui/AppIcon.vue";

const props = defineProps<{
  readiness?: Record<string, { ready?: boolean; reason?: string }>;
}>();

const { t } = useI18n();

const modulesList = computed(() => {
  if (!props.readiness) return [];
  return Object.entries(props.readiness).map(([name, status]) => ({
    name,
    ready: !!status.ready,
    reason: status.reason || "",
  }));
});

// Check if any critical module (excluding approval) is not ready
const hasCriticalBlockers = computed(() => {
  return modulesList.value.some(
    (mod) => !mod.ready && mod.name !== "approval",
  );
});

defineExpose({ hasCriticalBlockers });
</script>

<template>
  <div v-if="modulesList.length > 0" class="workflow-readiness-panel">
    <!-- Module cards grid -->
    <div class="workflow-readiness-grid">
      <div
        v-for="mod in modulesList"
        :key="mod.name"
        class="workflow-readiness-card"
        :class="{ 'is-ready': mod.ready, 'is-not-ready': !mod.ready }"
      >
        <div class="workflow-readiness-card__header">
          <span class="workflow-readiness-card__name">{{ t(`workflow.modules.${mod.name}` as any, mod.name) }}</span>
          <div class="workflow-readiness-card__indicator">
            <AppIcon
              v-if="mod.ready"
              name="check"
              size="sm"
              class="icon-success"
            />
            <AppIcon
              v-else
              name="warning"
              size="sm"
              class="icon-warning"
            />
          </div>
        </div>
        <p v-if="!mod.ready && mod.name === 'approval'" class="workflow-readiness-card__description">
          {{ t("workflow.modules.approvalWarning", "Approval module is not fully active. Workflow is using internal workflow approval settings.") }}
        </p>
        <p v-else-if="!mod.ready && mod.reason" class="workflow-readiness-card__description">
          {{ mod.reason }}
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.workflow-readiness-panel {
  display: grid;
  gap: var(--space-4);
}

.workflow-readiness-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--space-4);
}

.workflow-readiness-card {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--space-4);
  background: var(--bg-card);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.workflow-readiness-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.03);
}

.workflow-readiness-card.is-ready {
  border-left: 4px solid var(--color-success-500);
}

.workflow-readiness-card.is-not-ready {
  border-left: 4px solid var(--color-warning-500);
}

.workflow-readiness-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-2);
}

.workflow-readiness-card__name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  text-transform: capitalize;
}

.workflow-readiness-card__indicator {
  display: flex;
  align-items: center;
}

.workflow-readiness-card__description {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  line-height: 1.4;
}

.icon-success {
  color: var(--color-success-500);
}

.icon-warning {
  color: var(--color-warning-500);
}
</style>
