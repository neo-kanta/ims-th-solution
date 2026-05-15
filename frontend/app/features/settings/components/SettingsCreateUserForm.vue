<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";

import type { CreateAdminUserInput } from "../admin.types";
import SettingsPasswordToggle from "./SettingsPasswordToggle.vue";

interface ServerFieldError {
  field: string;
  message: string;
}

const props = defineProps<{
  loading: boolean;
  error: string | null;
  serverFieldErrors?: ServerFieldError[];
  successNonce: number;
}>();

const emit = defineEmits<{
  submit: [payload: CreateAdminUserInput];
}>();

const { t } = useI18n();

const form = reactive<CreateAdminUserInput>({
  username: "",
  display_name: "",
  email: "",
  password: "",
});
const fieldErrors = ref<Partial<Record<keyof CreateAdminUserInput, string>>>({});
const showPassword = ref(false);

const MIN_PASSWORD_LENGTH = 12;

const isSubmitDisabled = computed(
  () =>
    props.loading ||
    !form.username.trim() ||
    !form.display_name.trim() ||
    !form.email.trim() ||
    form.password.length < MIN_PASSWORD_LENGTH,
);

watch(
  () => props.successNonce,
  () => {
    form.username = "";
    form.display_name = "";
    form.email = "";
    form.password = "";
    fieldErrors.value = {};
    showPassword.value = false;
  },
);

watch(
  () => props.serverFieldErrors,
  (errors) => {
    if (!errors || errors.length === 0) return;
    const next: Partial<Record<keyof CreateAdminUserInput, string>> = {
      ...fieldErrors.value,
    };
    for (const err of errors) {
      if (
        err.field === "username" ||
        err.field === "display_name" ||
        err.field === "email" ||
        err.field === "password"
      ) {
        next[err.field] = err.message;
      }
    }
    fieldErrors.value = next;
  },
  { deep: true },
);

function validateForm(): boolean {
  const nextErrors: Partial<Record<keyof CreateAdminUserInput, string>> = {};

  if (form.username.trim().length < 3) {
    nextErrors.username = t("settings.console.otherAccounts.validationUsername");
  }

  if (!form.display_name.trim()) {
    nextErrors.display_name = t("common.required", "Required");
  }

  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) {
    nextErrors.email = t("settings.console.otherAccounts.validationEmail");
  }

  if (form.password.length < MIN_PASSWORD_LENGTH) {
    nextErrors.password = t("settings.console.otherAccounts.validationPassword");
  }

  fieldErrors.value = nextErrors;
  return Object.keys(nextErrors).length === 0;
}

function handleSubmit() {
  if (props.loading || !validateForm()) {
    return;
  }

  emit("submit", {
    username: form.username.trim(),
    display_name: form.display_name.trim(),
    email: form.email.trim(),
    password: form.password,
  });
}
</script>

<template>
  <section class="settings-panel settings-create-user">
    <header class="settings-panel__header">
      <div>
        <h2 class="settings-panel__title">
          {{ t("settings.console.otherAccounts.createTitle") }}
        </h2>
        <p class="settings-panel__subtitle">
          {{ t("settings.console.otherAccounts.createSubtitle") }}
        </p>
      </div>
    </header>

    <form
      class="settings-create-user__form"
      :aria-busy="loading"
      @submit.prevent="handleSubmit"
    >
      <div class="form-group" :class="{ 'field-error': fieldErrors.username }">
        <label for="settings-create-username" class="label label-required">
          {{ t("settings.usernamePlaceholder") }}
        </label>
        <input
          id="settings-create-username"
          v-model="form.username"
          class="input"
          type="text"
          autocomplete="off"
          :aria-invalid="Boolean(fieldErrors.username)"
          :aria-describedby="
            fieldErrors.username ? 'settings-create-username-error' : undefined
          "
        />
        <div
          v-if="fieldErrors.username"
          id="settings-create-username-error"
          class="error-text"
        >
          {{ fieldErrors.username }}
        </div>
      </div>

      <div
        class="form-group"
        :class="{ 'field-error': fieldErrors.display_name }"
      >
        <label for="settings-create-display-name" class="label label-required">
          {{ t("settings.displayNamePlaceholder") }}
        </label>
        <input
          id="settings-create-display-name"
          v-model="form.display_name"
          class="input"
          type="text"
          autocomplete="name"
          :aria-invalid="Boolean(fieldErrors.display_name)"
          :aria-describedby="
            fieldErrors.display_name
              ? 'settings-create-display-name-error'
              : undefined
          "
        />
        <div
          v-if="fieldErrors.display_name"
          id="settings-create-display-name-error"
          class="error-text"
        >
          {{ fieldErrors.display_name }}
        </div>
      </div>

      <div class="form-group" :class="{ 'field-error': fieldErrors.email }">
        <label for="settings-create-email" class="label label-required">
          {{ t("settings.emailPlaceholder") }}
        </label>
        <input
          id="settings-create-email"
          v-model="form.email"
          class="input"
          type="email"
          autocomplete="email"
          :aria-invalid="Boolean(fieldErrors.email)"
          :aria-describedby="
            fieldErrors.email ? 'settings-create-email-error' : undefined
          "
        />
        <div
          v-if="fieldErrors.email"
          id="settings-create-email-error"
          class="error-text"
        >
          {{ fieldErrors.email }}
        </div>
      </div>

      <div class="form-group" :class="{ 'field-error': fieldErrors.password }">
        <label for="settings-create-password" class="label label-required">
          {{ t("settings.temporaryPasswordPlaceholder") }}
        </label>
        <div class="settings-password-field">
          <input
            id="settings-create-password"
            v-model="form.password"
            class="input"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="new-password"
            :aria-invalid="Boolean(fieldErrors.password)"
            :aria-describedby="
              fieldErrors.password ? 'settings-create-password-error' : undefined
            "
          />
          <SettingsPasswordToggle
            :shown="showPassword"
            @toggle="showPassword = !showPassword"
          />
        </div>
        <div
          v-if="fieldErrors.password"
          id="settings-create-password-error"
          class="error-text"
        >
          {{ fieldErrors.password }}
        </div>
      </div>

      <div v-if="error" class="alert alert-danger" role="alert">
        {{ error }}
      </div>

      <AppButton
        type="submit"
        variant="primary"
        :loading="loading"
        :disabled="isSubmitDisabled"
      >
        {{ loading ? t("settings.creating") : t("settings.createAccount") }}
      </AppButton>
    </form>
  </section>
</template>

<style scoped>
.settings-create-user__form {
  display: grid;
  gap: var(--space-5);
  padding: var(--space-5);
}

.settings-create-user__form .form-group {
  margin-bottom: 0;
}

.settings-password-field {
  position: relative;
}

.settings-password-field .input {
  padding-right: var(--space-10);
}

</style>
