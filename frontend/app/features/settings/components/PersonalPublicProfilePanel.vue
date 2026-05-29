<script setup lang="ts">
import { ref, watch, computed } from "vue";
import type { PersonalAccountPayload } from "../account.types";
import AppSettingsActions from "~/shared/ui/AppSettingsActions.vue";

const props = defineProps<{
  account: PersonalAccountPayload | null;
  loading: boolean;
}>();

const localName = ref("");
const localEmail = ref("");
const localBio = ref("Tell us a little bit about yourself");

watch(
  () => props.account,
  (acc) => {
    if (acc) {
      localName.value = acc.user.display_name || "guest";
      localEmail.value = acc.user.email || "";
    }
  },
  { immediate: true },
);

const isDirty = computed(() => {
  return (
    localName.value !== (props.account?.user.display_name || "neo") ||
    localEmail.value !== (props.account?.user.email || "") ||
    localBio.value !== "Tell us a little bit about yourself"
  );
});

function handleSave() {
  alert(`Saved: ${localName.value}, ${localEmail.value}, ${localBio.value}`);
}

function handleCancel() {
  localName.value = props.account?.user.display_name || "neo";
  localEmail.value = props.account?.user.email || "";
  localBio.value = "Tell us a little bit about yourself";
}
</script>

<template>
  <section class="settings-panel">
    <header class="settings-panel__header">
      <h2 class="settings-panel__title">Public profile</h2>
    </header>
    <div
      class="settings-panel__body"
      :aria-busy="loading"
      style="padding: var(--space-5); display: grid; gap: var(--space-4)"
    >
      <div style="display: grid; gap: var(--space-3); max-width: 500px">
        <div>
          <label
            style="
              display: block;
              font-size: var(--font-size-sm);
              font-weight: var(--font-weight-medium);
              margin-bottom: var(--space-1);
            "
            >Name</label
          >
          <input type="text" class="personal-input" v-model="localName" />
          <p class="help-text">
            Your name may appear around GitHub where you contribute or are
            mentioned.
          </p>
        </div>

        <div>
          <label
            style="
              display: block;
              font-size: var(--font-size-sm);
              font-weight: var(--font-weight-medium);
              margin-bottom: var(--space-1);
            "
            >Public email</label
          >
          <select class="personal-select" v-model="localEmail">
            <option value="">Select a verified email to display</option>
            <option v-if="account?.user.email" :value="account.user.email">
              {{ account.user.email }}
            </option>
            <option value="admin@ims.local">admin@ims.local</option>
          </select>
          <p class="help-text">You have set your email address to private.</p>
        </div>

        <div>
          <label
            style="
              display: block;
              font-size: var(--font-size-sm);
              font-weight: var(--font-weight-medium);
              margin-bottom: var(--space-1);
            "
            >Bio</label
          >
          <textarea
            class="personal-textarea"
            v-model="localBio"
            rows="3"
          ></textarea>
        </div>

        <!-- Configurable Save & Cancel Actions -->
        <AppSettingsActions
          :show-save="true"
          :show-cancel="true"
          :save-loading="loading"
          :save-disabled="!isDirty"
          :cancel-disabled="!isDirty"
          style="margin-top: var(--space-2)"
          @save="handleSave"
          @cancel="handleCancel"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.personal-input,
.personal-select,
.personal-textarea {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.personal-textarea {
  resize: vertical;
}

.help-text {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-2xs);
  color: var(--text-secondary);
}
</style>
