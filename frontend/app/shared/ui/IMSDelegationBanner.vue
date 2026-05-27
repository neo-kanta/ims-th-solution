<script setup lang="ts">
import AppIcon from "./AppIcon.vue";

interface Props {
  principalName: string;
  scope?: string;
  contractName?: string;
  startDate?: string;
  endDate?: string;
}

withDefaults(defineProps<Props>(), {
  scope: "Full Authority",
  contractName: "",
  startDate: "",
  endDate: "",
});

const emit = defineEmits<{
  exit: [];
}>();
</script>

<template>
  <div class="ims-delegation-banner" role="status">
    <div class="ims-delegation-banner__container">
      <div class="ims-delegation-banner__content">
        <AppIcon name="lock" size="sm" class="ims-delegation-banner__icon" />
        <div class="ims-delegation-banner__message">
          Acting as delegate for <strong>{{ principalName }}</strong>.
          Scope: <span class="ims-delegation-banner__scope">{{ scope }}</span>
          <span v-if="contractName"> on <strong>{{ contractName }}</strong></span>.
          <span v-if="startDate && endDate" class="ims-delegation-banner__dates">
            (Effective: {{ startDate }} to {{ endDate }})
          </span>
        </div>
      </div>
      <button
        type="button"
        class="ims-delegation-banner__exit"
        @click="emit('exit')"
      >
        Exit Session
      </button>
    </div>
  </div>
</template>

<style scoped>
.ims-delegation-banner {
  background: var(--alert-warning-bg, #fff8c5);
  border: 1px solid var(--alert-warning-border, #9a6700);
  border-radius: var(--radius-md, 6px);
  padding: var(--space-3, 12px) var(--space-4, 16px);
  color: var(--alert-warning-text, #9a6700);
  width: 100%;
}

.ims-delegation-banner__container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4, 16px);
  flex-wrap: wrap;
}

.ims-delegation-banner__content {
  display: flex;
  align-items: center;
  gap: var(--space-3, 12px);
  font-size: var(--font-size-sm, 14px);
  line-height: 1.4;
}

.ims-delegation-banner__icon {
  flex-shrink: 0;
  color: currentColor;
}

.ims-delegation-banner__scope {
  font-weight: var(--font-weight-semibold, 600);
}

.ims-delegation-banner__dates {
  font-size: var(--font-size-xs, 12px);
  opacity: 0.85;
  margin-left: 4px;
}

.ims-delegation-banner__exit {
  background: var(--alert-warning-text, #9a6700);
  border: 1px solid transparent;
  color: var(--alert-warning-bg, #fff8c5);
  border-radius: var(--radius-sm, 4px);
  padding: 4px var(--space-3, 12px);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-bold, 700);
  cursor: pointer;
  transition: opacity 0.15s ease;
}

.ims-delegation-banner__exit:hover {
  opacity: 0.9;
}
</style>
