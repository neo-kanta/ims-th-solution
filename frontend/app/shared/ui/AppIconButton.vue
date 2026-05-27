<script setup lang="ts">
import { computed } from "vue";
import AppButton from "./AppButton.vue";
import AppIcon from "./AppIcon.vue";

interface Props {
  icon: string;
  ariaLabel: string;
  tooltip?: string;
  variant?: "primary" | "secondary" | "danger" | "warning" | "success" | "ghost";
  size?: "xs" | "sm" | "md" | "lg";
  disabled?: boolean;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  tooltip: "",
  variant: "secondary",
  size: "md",
  disabled: false,
  loading: false,
});

const emit = defineEmits<{
  click: [event: MouseEvent];
}>();

const iconSize = computed(() => {
  if (props.size === "xs") return "xs" as const;
  if (props.size === "sm") return "xs" as const;
  if (props.size === "md") return "sm" as const;
  return "md" as const;
});
</script>

<template>
  <AppButton
    :variant="variant"
    :size="size"
    :disabled="disabled"
    :loading="loading"
    icon
    :aria-label="ariaLabel"
    :title="tooltip || ariaLabel"
    @click="emit('click', $event)"
  >
    <AppIcon v-if="!loading" :name="icon" :size="iconSize" />
  </AppButton>
</template>
