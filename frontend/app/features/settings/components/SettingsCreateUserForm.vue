<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";

import type { CreateAdminUserInput } from "../admin.types";

const props = defineProps<{
  loading: boolean;
  error: string | null;
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

const passwordToggleLabel = computed(() =>
  showPassword.value
    ? t("auth.hidePassword", "Hide password")
    : t("auth.showPassword", "Show password"),
);

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
          <button
            type="button"
            class="settings-password-field__toggle"
            :aria-label="passwordToggleLabel"
            :title="passwordToggleLabel"
            :aria-pressed="showPassword"
            @click="showPassword = !showPassword"
          >
            <svg
              v-if="showPassword"
              xmlns="http://www.w3.org/2000/svg"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <path
                d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
              />
              <line x1="1" y1="1" x2="23" y2="23" />
            </svg>
            <svg
              v-else
              xmlns="http://www.w3.org/2000/svg"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </button>
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

.settings-password-field__toggle {
  position: absolute;
  top: 50%;
  right: var(--space-2);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: var(--radius-sm);
  color: var(--text-tertiary);
  transform: translateY(-50%);
}

.settings-password-field__toggle:hover {
  color: var(--text-primary);
  background: var(--action-ghost-hover);
}
</style>
