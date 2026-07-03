<script setup lang="ts">
import type { PersonalAppearancePreferences, PersonalWorkPreferences } from "../personal.types";

const props = defineProps<{
  preferences: PersonalWorkPreferences | null;
  appearancePreferences: PersonalAppearancePreferences | null;
  loading: boolean;
}>();

const emit = defineEmits<{
  "save-work": [preferences: PersonalWorkPreferences];
  "save-appearance": [preferences: PersonalAppearancePreferences];
}>();

function updateWork(patch: Partial<PersonalWorkPreferences>) {
  if (props.preferences) {
    emit("save-work", { ...props.preferences, ...patch });
  }
}
</script>

<template>
  <section class="settings-panel">
    <header class="settings-panel__header">
      <h2 class="settings-panel__title">Workstation</h2>
    </header>
    <div class="settings-panel__body" :aria-busy="loading" style="padding: var(--space-5); display: grid; gap: var(--space-4);">
      <div v-if="preferences">
        <label>Default Dashboard</label>
        <select :value="preferences.defaultDashboard" @change="updateWork({ defaultDashboard: ($event.target as HTMLSelectElement).value as any })">
          <option value="overview">Overview</option>
          <option value="workflow">Workflow</option>
          <option value="investment">Investment</option>
        </select>
      </div>
    </div>
  </section>
</template>
