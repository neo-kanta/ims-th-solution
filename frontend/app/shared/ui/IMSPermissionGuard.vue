<script setup lang="ts">
import { computed } from "vue";
import { useAuthStore } from "~/stores/useAuthStore";

interface Props {
  permission: string;
  mode?: "hide" | "disable";
}

const props = withDefaults(defineProps<Props>(), {
  mode: "hide",
});

const auth = useAuthStore();

const isAllowed = computed(() => {
  if (!props.permission) return true;
  return auth.hasPermission(props.permission);
});
</script>

<template>
  <template v-if="isAllowed">
    <slot />
  </template>
  <template v-else-if="mode === 'disable'">
    <div class="permission-disabled" aria-disabled="true" title="You do not have permission to perform this action.">
      <slot />
    </div>
  </template>
  <template v-else>
    <slot name="fallback" />
  </template>
</template>

<style scoped>
.permission-disabled {
  pointer-events: none;
  opacity: 0.55;
  cursor: not-allowed;
}

.permission-disabled :deep(*) {
  pointer-events: none !important;
  cursor: not-allowed !important;
}
</style>
