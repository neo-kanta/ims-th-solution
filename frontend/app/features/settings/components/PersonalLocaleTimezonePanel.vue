<script setup lang="ts">
import { ref, watch, computed } from "vue";
import type { PersonalAppearancePreferences, PersonalWorkPreferences } from "../personal.types";
import AppSettingsActions from "~/shared/ui/AppSettingsActions.vue";

const props = defineProps<{
  preferences: PersonalWorkPreferences | null;
  appearancePreferences: PersonalAppearancePreferences | null;
  loading: boolean;
}>();

const emit = defineEmits<{
  "save-work": [preferences: PersonalWorkPreferences];
  "save-appearance": [preferences: PersonalAppearancePreferences];
}>();

const localWork = ref<PersonalWorkPreferences | null>(null);
const localAppearance = ref<PersonalAppearancePreferences | null>(null);

watch(
  () => props.preferences,
  (val) => {
    if (val) localWork.value = { ...val };
  },
  { immediate: true }
);

watch(
  () => props.appearancePreferences,
  (val) => {
    if (val) localAppearance.value = { ...val };
  },
  { immediate: true }
);

const isDirty = computed(() => {
  const workDirty = props.preferences && localWork.value && 
    props.preferences.language !== localWork.value.language;
  const appearanceDirty = props.appearancePreferences && localAppearance.value && (
    props.appearancePreferences.timezone !== localAppearance.value.timezone ||
    props.appearancePreferences.dateFormat !== localAppearance.value.dateFormat
  );
  return workDirty || appearanceDirty;
});

function handleSave() {
  if (localWork.value && localWork.value.language !== props.preferences?.language) {
    emit("save-work", { ...localWork.value });
  }
  if (localAppearance.value && (
    localAppearance.value.timezone !== props.appearancePreferences?.timezone ||
    localAppearance.value.dateFormat !== props.appearancePreferences?.dateFormat
  )) {
    emit("save-appearance", { ...localAppearance.value });
  }
}

function handleCancel() {
  if (props.preferences) {
    localWork.value = { ...props.preferences };
  }
  if (props.appearancePreferences) {
    localAppearance.value = { ...props.appearancePreferences };
  }
}
</script>

<template>
  <section class="settings-panel">
    <header class="settings-panel__header">
      <h2 class="settings-panel__title">Locale & timezone</h2>
    </header>
    <div class="settings-panel__body" :aria-busy="loading" style="padding: var(--space-5); display: grid; gap: var(--space-4);">
      <div v-if="localWork">
        <label>Language</label>
        <select :value="localWork.language" @change="localWork.language = ($event.target as HTMLSelectElement).value as any">
          <option value="en">English</option>
          <option value="th">Thai</option>
        </select>
      </div>

      <div v-if="localAppearance">
        <label>Timezone</label>
        <select :value="localAppearance.timezone" @change="localAppearance.timezone = ($event.target as HTMLSelectElement).value">
          <option value="Asia/Bangkok">Asia/Bangkok</option>
          <option value="UTC">UTC</option>
        </select>
      </div>
      
      <div v-if="localAppearance">
        <label>Date Format</label>
        <select :value="localAppearance.dateFormat" @change="localAppearance.dateFormat = ($event.target as HTMLSelectElement).value as any">
          <option value="YYYY-MM-DD">YYYY-MM-DD</option>
          <option value="DD/MM/YYYY">DD/MM/YYYY</option>
          <option value="MM/DD/YYYY">MM/DD/YYYY</option>
        </select>
      </div>

      <!-- Configurable Save & Cancel Actions -->
      <AppSettingsActions
        :show-save="true"
        :show-cancel="true"
        :save-loading="loading"
        :save-disabled="!isDirty"
        :cancel-disabled="!isDirty"
        @save="handleSave"
        @cancel="handleCancel"
      />
    </div>
  </section>
</template>
