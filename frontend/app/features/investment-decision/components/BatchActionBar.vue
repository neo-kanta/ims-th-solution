<script setup lang="ts">
const props = defineProps<{ selectedNos: string[]; loading: boolean }>();
const emit = defineEmits<{
  (e: "approve"): void;
  (e: "reject"): void;
  (e: "clear"): void;
}>();
</script>

<template>
  <div v-if="selectedNos.length > 0" class="batch-bar">
    <span class="batch-count">{{ selectedNos.length }} selected</span>
    <button class="batch-btn batch-btn--approve" :disabled="loading" @click="emit('approve')">
      Batch Approve
    </button>
    <button class="batch-btn batch-btn--reject" :disabled="loading" @click="emit('reject')">
      Batch Reject
    </button>
    <button class="batch-btn" @click="emit('clear')">Clear</button>
  </div>
</template>

<style scoped>
.batch-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: var(--bg-card-muted);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md, 6px);
  margin-bottom: 8px;
}

.batch-count {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
}

.batch-btn {
  padding: 5px 14px;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  color: var(--text-primary);
  cursor: pointer;
}

.batch-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.batch-btn--approve {
  background: var(--state-success, #1a7f37);
  color: #fff;
  border-color: transparent;
}

.batch-btn--reject {
  background: var(--state-danger, #cf222e);
  color: #fff;
  border-color: transparent;
}
</style>
