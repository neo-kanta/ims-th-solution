<script setup lang="ts">
import type { ChangePersonalPasswordInput } from "../account.types";
import AppSettingsActions from "~/shared/ui/AppSettingsActions.vue";

defineProps<{
  passwordLoading: boolean;
  passwordError: string | null;
  passwordSuccessNonce: number;
}>();

const emit = defineEmits<{
  "change-password": [payload: ChangePersonalPasswordInput];
}>();
</script>

<template>
  <section class="settings-panel">
    <header class="settings-panel__header">
      <h2 class="settings-panel__title">Password & authentication</h2>
    </header>
    <div class="settings-panel__body" style="padding: var(--space-5);">
      <p class="text-secondary" style="margin-bottom: var(--space-4);">
        Update your password. Backend API pending.
      </p>
      
      <div style="display: grid; gap: var(--space-3); max-width: 400px;">
        <div>
          <label style="display: block; font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); margin-bottom: var(--space-1);">Current password</label>
          <input type="password" class="personal-input" disabled />
        </div>
        <div>
          <label style="display: block; font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); margin-bottom: var(--space-1);">New password</label>
          <input type="password" class="personal-input" disabled />
        </div>
        <div>
          <label style="display: block; font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); margin-bottom: var(--space-1);">Confirm new password</label>
          <input type="password" class="personal-input" disabled />
        </div>
        
        <!-- Configurable Save & Cancel Actions (Only Save) -->
        <AppSettingsActions
          :show-save="true"
          :show-cancel="false"
          save-label="Save password"
          :save-disabled="true"
          style="margin-top: var(--space-2); border-top: none; padding-top: 0;"
        />
      </div>
    </div>
  </section>
</template>

<style scoped>
.personal-input {
  width: 100%;
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}
.personal-input:disabled {
  background: var(--bg-card-muted);
  color: var(--text-tertiary);
  cursor: not-allowed;
}
</style>
