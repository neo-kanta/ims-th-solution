<script setup lang="ts">
import { reactive, watch } from "vue";

import type { PersonalWorkPreferences } from "../personal.types";

const props = defineProps<{
  preferences: PersonalWorkPreferences | null;
  loading: boolean;
}>();

const emit = defineEmits<{
  save: [preferences: PersonalWorkPreferences];
}>();

const form = reactive<PersonalWorkPreferences>({
  defaultDashboard: "overview",
  defaultFund: "",
  language: "en",
  tablePageSize: 25,
  source: "local",
});

watch(
  () => props.preferences,
  (next) => {
    if (next) Object.assign(form, next);
  },
  { immediate: true },
);
</script>

<template>
  <section class="settings-panel">
    <header class="settings-panel__header">
      <div>
        <h2 class="settings-panel__title">Work Preferences</h2>
        <p class="settings-panel__subtitle">
          Defaults that make repeated IMS work faster without changing permissions or system policy.
        </p>
      </div>
    </header>

    <form class="personal-work-form" :aria-busy="loading" @submit.prevent="emit('save', { ...form })">
      <label class="form-group">
        <span class="label">Default dashboard</span>
        <select v-model="form.defaultDashboard" class="input">
          <option value="overview">Overview</option>
          <option value="workflow">Workflow</option>
          <option value="investment">Investment research</option>
          <option value="notifications">Notifications</option>
        </select>
      </label>

      <label class="form-group">
        <span class="label">Default fund / contract</span>
        <input v-model="form.defaultFund" class="input" placeholder="Optional contract or fund code" />
      </label>


      <label class="form-group">
        <span class="label">Language</span>
        <select v-model="form.language" class="input">
          <option value="en">English</option>
          <option value="th">Thai</option>
          <option value="zh">Chinese</option>
        </select>
      </label>

      <label class="form-group">
        <span class="label">Table page size</span>
        <select v-model.number="form.tablePageSize" class="input">
          <option :value="12">12</option>
          <option :value="25">25</option>
          <option :value="50">50</option>
          <option :value="100">100</option>
        </select>
      </label>

      <div class="personal-work-form__actions">
        <AppButton variant="primary" size="sm" type="submit" :loading="loading">
          Save preferences
        </AppButton>
      </div>
    </form>
  </section>
</template>

<style scoped>
.personal-work-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4);
  padding: var(--space-5);
}

.personal-work-form__actions {
  grid-column: 1 / -1;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 760px) {
  .personal-work-form {
    grid-template-columns: 1fr;
  }
}
</style>
