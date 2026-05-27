<script setup lang="ts">
import { computed } from "vue";

interface Props {
  name: string;
  role?: string;
  timestamp?: string | null;
  delegated?: boolean;
  status?: "approved" | "rejected" | string;
}

const props = withDefaults(defineProps<Props>(), {
  role: "",
  timestamp: null,
  delegated: false,
  status: "approved",
});

const isApproved = computed(() => props.status.toLowerCase() === "approved");

const stampClass = computed(() => {
  return {
    "ims-stamp--approved": isApproved.value,
    "ims-stamp--rejected": !isApproved.value,
  };
});

const formattedTime = computed(() => {
  if (!props.timestamp) return "";
  try {
    const d = new Date(props.timestamp);
    return d.toLocaleDateString(undefined, { year: "numeric", month: "2-digit", day: "2-digit" });
  } catch {
    return props.timestamp;
  }
});
</script>

<template>
  <div class="ims-stamp" :class="stampClass">
    <div class="ims-stamp__border">
      <div class="ims-stamp__role">{{ role || "APPROVER" }}</div>
      <div class="ims-stamp__name">
        {{ name }}
        <span v-if="delegated" class="ims-stamp__delegate-tag" title="Delegated (代)">(代)</span>
      </div>
      <div class="ims-stamp__status">{{ isApproved ? "APPROVED" : "REJECTED" }}</div>
      <div v-if="formattedTime" class="ims-stamp__date">{{ formattedTime }}</div>
    </div>
  </div>
</template>

<style scoped>
.ims-stamp {
  display: inline-block;
  font-family: "Courier New", Courier, monospace, sans-serif;
  text-align: center;
  user-select: none;
}

.ims-stamp__border {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 90px;
  height: 90px;
  border: 3px double currentColor;
  border-radius: 50%;
  padding: 4px;
  transform: rotate(-8deg); /* Slight angle for realistic stamp feel */
}

.ims-stamp--approved {
  color: var(--state-success, #1f883d);
}

.ims-stamp--rejected {
  color: var(--state-danger, #cf222e);
}

.ims-stamp__role {
  font-size: 8px;
  font-weight: bold;
  text-transform: uppercase;
  border-bottom: 1px solid currentColor;
  padding-bottom: 2px;
  width: 80%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ims-stamp__name {
  font-size: 11px;
  font-weight: 900;
  padding: 2px 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 82px;
}

.ims-stamp__delegate-tag {
  font-size: 9px;
  font-weight: bold;
  font-style: normal;
  display: inline-block;
}

.ims-stamp__status {
  font-size: 9px;
  font-weight: bold;
  letter-spacing: 0.05em;
  border-top: 1px solid currentColor;
  border-bottom: 1px solid currentColor;
  padding: 1px 0;
  width: 90%;
}

.ims-stamp__date {
  font-size: 8px;
  margin-top: 2px;
}
</style>
