<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import {
  resolveAuthRedirectReasonKey,
  resolveLoginErrorTranslationKey,
  sanitizeAuthRedirectTarget,
} from "../../features/auth/lib/auth";

definePageMeta({
  layout: "auth",
  auth: false,
});

const { t } = useI18n();
const route = useRoute();
const authStore = useAuthStore();

const username = ref("");
const password = ref("");
const showPassword = ref(false);
const isSubmitting = ref(false);
const fieldErrors = ref<Record<string, string>>({});
const failedAttempts = ref(0);
const isCapsLockOn = ref(false);
const usernameInputRef = ref<HTMLInputElement | null>(null);
const passwordInputRef = ref<HTMLInputElement | null>(null);
const submitAlertRef = ref<HTMLElement | null>(null);

const authReasonMessage = computed(() => {
  const translationKey = resolveAuthRedirectReasonKey(route.query.reason);
  return translationKey ? t(translationKey) : null;
});

const passwordToggleLabel = computed(() =>
  showPassword.value
    ? t("auth.hidePassword", "Hide password")
    : t("auth.showPassword", "Show password"),
);

const failedAttemptsMessage = computed(() =>
  t(
    "auth.failedAttempts",
    { count: failedAttempts.value },
    `Failed sign-in attempts: ${failedAttempts.value}`,
  ),
);

const accountLockWarning = computed(() =>
  t(
    "auth.accountLockWarning",
    "Your account may lock after 5 failed attempts.",
  ),
);

const capsLockWarning = computed(() =>
  t("auth.capsLockWarning", "Warning: Caps Lock is on."),
);

const genericLoginFailure = computed(() =>
  t("auth.loginFailed", "We couldn't sign you in. Please try again."),
);

const passwordDescribedBy = computed(() => {
  const ids: string[] = [];

  if (fieldErrors.value.password) {
    ids.push("password-error");
  }

  if (isCapsLockOn.value) {
    ids.push("caps-lock-warning");
  }

  return ids.length > 0 ? ids.join(" ") : undefined;
});

const submitDisabled = computed(
  () =>
    authStore.isLoading ||
    isSubmitting.value ||
    !username.value.trim() ||
    !password.value,
);

watch(username, () => {
  delete fieldErrors.value.username;
  delete fieldErrors.value.submit;
});

watch(password, () => {
  delete fieldErrors.value.password;
  delete fieldErrors.value.submit;
});

function checkCapsLock(event: KeyboardEvent) {
  if (event.getModifierState) {
    isCapsLockOn.value = event.getModifierState("CapsLock");
  }
}

async function focusElement(element: HTMLElement | null) {
  await nextTick();
  element?.focus();
}

function resolveSubmitErrorMessage() {
  const translationKey = resolveLoginErrorTranslationKey(authStore.error);

  if (translationKey) {
    return t(translationKey);
  }

  return authStore.error || genericLoginFailure.value;
}

async function handleLogin() {
  if (isSubmitting.value || authStore.isLoading) {
    return;
  }

  fieldErrors.value = {};

  if (!username.value.trim()) {
    fieldErrors.value.username = t("common.required", "Required");
    await focusElement(usernameInputRef.value);
    return;
  }

  if (!password.value) {
    fieldErrors.value.password = t("common.required", "Required");
    await focusElement(passwordInputRef.value);
    return;
  }

  isSubmitting.value = true;

  try {
    const success = await authStore.login({
      username: username.value.trim(),
      password: password.value,
      totp_code: "",
      recovery_code: "",
    });

    if (success) {
      failedAttempts.value = 0;
      const redirectTarget = sanitizeAuthRedirectTarget(route.query.redirect);

      await navigateTo(redirectTarget);
      return;
    }

    failedAttempts.value += 1;
    fieldErrors.value.submit = resolveSubmitErrorMessage();
    await focusElement(submitAlertRef.value);
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div class="login-form">
    <h2 class="login-title">{{ t("auth.login") }}</h2>
    <p class="login-subtitle">{{ t("app.tagline") }}</p>

    <form
      class="login-fields"
      :aria-busy="authStore.isLoading || isSubmitting"
      @submit.prevent="handleLogin"
    >
      <div
        v-if="authReasonMessage && !fieldErrors.submit"
        class="alert alert-info"
        role="status"
        aria-live="polite"
      >
        {{ authReasonMessage }}
      </div>

      <div
        v-if="fieldErrors.submit"
        ref="submitAlertRef"
        class="alert alert-danger submit-alert"
        role="alert"
        aria-live="assertive"
        tabindex="-1"
      >
        <div class="alert-content">
          <div>{{ fieldErrors.submit }}</div>
          <div
            v-if="failedAttempts >= 1 && failedAttempts < 5"
            class="failed-attempts"
            aria-live="polite"
          >
            {{ failedAttemptsMessage }}
          </div>
          <div
            v-if="failedAttempts >= 3"
            class="lock-warning"
            aria-live="polite"
          >
            {{ accountLockWarning }}
          </div>
        </div>
      </div>

      <div class="form-group">
        <label for="login-username" class="label label-required">
          {{ t("auth.username") }}
        </label>
        <input
          id="login-username"
          ref="usernameInputRef"
          v-model="username"
          type="text"
          :placeholder="t('auth.enterUsername')"
          :disabled="authStore.isLoading"
          :aria-describedby="
            fieldErrors.username ? 'username-error' : undefined
          "
          :aria-invalid="Boolean(fieldErrors.username)"
          required
          autofocus
          autocomplete="username"
          class="input"
          :class="{ 'is-error': fieldErrors.username }"
          @keydown.enter="handleLogin"
        />
        <div v-if="fieldErrors.username" id="username-error" class="error-text">
          {{ fieldErrors.username }}
        </div>
      </div>

      <div class="form-group">
        <div class="password-label-row">
          <label for="login-password" class="label label-required">
            {{ t("auth.password") }}
          </label>
        </div>
        <div class="password-input-wrapper">
          <input
            id="login-password"
            ref="passwordInputRef"
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            :placeholder="t('auth.enterPassword')"
            :disabled="authStore.isLoading"
            :aria-describedby="passwordDescribedBy"
            :aria-invalid="Boolean(fieldErrors.password)"
            required
            autocomplete="current-password"
            class="input"
            :class="{ 'is-error': fieldErrors.password }"
            @keyup="checkCapsLock"
            @blur="isCapsLockOn = false"
            @keydown.enter="handleLogin"
          />

          <button
            type="button"
            class="password-toggle"
            :aria-label="passwordToggleLabel"
            :aria-pressed="showPassword"
            :title="passwordToggleLabel"
            @click="showPassword = !showPassword"
          >
            <svg
              v-if="showPassword"
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path
                d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
              />
              <line x1="1" y1="1" x2="23" y2="23" />
            </svg>
            <svg
              v-else
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </button>
        </div>

        <div
          v-if="isCapsLockOn"
          id="caps-lock-warning"
          class="caps-warning"
          aria-live="polite"
        >
          {{ capsLockWarning }}
        </div>
        <div v-if="fieldErrors.password" id="password-error" class="error-text">
          {{ fieldErrors.password }}
        </div>
      </div>

      <button
        type="submit"
        class="btn btn-primary btn-lg"
        :aria-busy="authStore.isLoading || isSubmitting"
        :disabled="submitDisabled"
      >
        <svg
          v-if="authStore.isLoading || isSubmitting"
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          style="animation: spin 0.75s linear infinite"
        >
          <path d="M21 12a9 9 0 1 1-6.219-8.56" />
        </svg>
        {{
          authStore.isLoading || isSubmitting
            ? t("auth.signingIn")
            : t("auth.login")
        }}
      </button>
    </form>

    <div class="login-footer">
      {{
        t(
          "auth.contactAdmin",
          "Need an account? Contact your system administrator.",
        )
      }}
    </div>
  </div>
</template>

<style scoped>
.login-form {
  margin: 10px 50px 20px;
  display: grid;
  gap: var(--space-6);
}

.login-title {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-neutral-900);
  text-align: center;
  letter-spacing: -0.02em;
}

.login-subtitle {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-neutral-500);
  text-align: center;
}

.login-fields {
  display: grid;
  gap: var(--space-5);
}

.form-group {
  display: grid;
  gap: var(--space-3);
}

.label {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-neutral-700);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.label-required::after {
  content: "*";
  color: var(--color-danger-500);
  font-weight: var(--font-weight-bold);
}

.input {
  width: 100%;
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-md);
  font-family: var(--font-family-sans);
  border: 1px solid var(--color-neutral-300);
  border-radius: var(--radius-md);
  background-color: var(--color-neutral-0);
  color: var(--color-neutral-900);
  caret-color: var(--color-neutral-900);
  color-scheme: light;
  transition:
    border-color 0.15s ease-in-out,
    box-shadow 0.15s ease-in-out;
  outline: none;
}

.input:-webkit-autofill,
.input:-webkit-autofill:hover,
.input:-webkit-autofill:focus,
.input:-webkit-autofill:active {
  -webkit-box-shadow: 0 0 0 30px var(--color-neutral-0) inset !important;
  -webkit-text-fill-color: var(--color-neutral-900) !important;
  transition: background-color 5000s ease-in-out 0s;
}

.input:hover:not(:disabled) {
  border-color: var(--color-neutral-400);
}

.input:focus {
  border-color: var(--color-primary-500);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.input.is-error {
  border-color: var(--color-danger-500);
  background-color: var(--color-danger-50);
}

.input.is-error:focus {
  border-color: var(--color-danger-500);
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.15);
}

.input:disabled {
  background-color: var(--color-neutral-50);
  color: var(--color-neutral-400);
  cursor: not-allowed;
  opacity: 0.65;
}

.password-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.password-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-input-wrapper .input {
  padding-right: var(--space-9);
}

.password-toggle {
  position: absolute;
  right: var(--space-2);
  background: none;
  border: none;
  color: var(--color-neutral-400);
  cursor: pointer;
  font-size: var(--font-size-md);
  padding: var(--space-2);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  transition:
    color 0.15s ease-in-out,
    box-shadow 0.15s ease-in-out;
}

.password-toggle:hover {
  color: var(--color-neutral-600);
}

.password-toggle:focus-visible {
  color: var(--color-neutral-700);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.password-toggle:active {
  transform: scale(0.95);
}

.error-text {
  font-size: var(--font-size-xs);
  color: var(--color-danger-600);
  font-weight: var(--font-weight-medium);
  margin-top: calc(var(--space-2) * -1);
}

.caps-warning {
  font-size: var(--font-size-xs);
  color: var(--color-warning-600);
  font-weight: var(--font-weight-medium);
  margin-top: calc(var(--space-2) * -1);
}

.alert {
  padding: var(--space-4);
  border-radius: var(--radius-md);
  border: 1px solid;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.alert-danger {
  background-color: var(--color-danger-50);
  border-color: var(--color-danger-300);
  color: var(--color-danger-800);
}

.alert-info {
  background-color: var(--color-primary-50);
  border-color: var(--color-primary-200);
  color: var(--color-primary-800);
}

.submit-alert {
  border-left: 3px solid var(--color-primary-500);
}

.alert-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.failed-attempts {
  font-size: var(--font-size-xs);
  margin-top: var(--space-1);
  opacity: 0.9;
}

.lock-warning {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  color: var(--color-danger-700);
}

.btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  font-family: var(--font-family-sans);
  font-weight: var(--font-weight-semibold);
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  outline: none;
  transition: all 0.15s ease-in-out;
}

.btn:focus {
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.btn-primary {
  background-color: var(--color-primary-600);
  color: var(--color-neutral-0);
}

.btn-primary:hover:not(:disabled) {
  background-color: var(--color-primary-700);
  box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.2);
}

.btn-primary:active:not(:disabled) {
  background-color: var(--color-primary-800);
  transform: translateY(1px);
}

.btn-primary:disabled {
  background-color: var(--color-primary-600);
  color: var(--color-neutral-0);
  cursor: not-allowed;
  opacity: 0.5;
}

.btn-lg {
  width: 100%;
  padding: var(--space-3) var(--space-5);
  font-size: var(--font-size-md);
  min-height: 44px;
}

.login-footer {
  text-align: center;
  font-size: var(--font-size-xs);
  color: var(--color-neutral-400);
  line-height: var(--line-height-relaxed);
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-neutral-100);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

svg {
  animation: spin 0.75s linear infinite;
}

.password-toggle svg {
  animation: none;
}

@media (max-width: 480px) {
  .login-form {
    margin: var(--space-4) var(--space-4) var(--space-5);
    gap: var(--space-5);
  }

  .login-title {
    font-size: var(--font-size-xl);
  }

  .login-subtitle {
    font-size: var(--font-size-xs);
  }

  .login-fields {
    gap: var(--space-4);
  }

  .btn-lg {
    padding: var(--space-3) var(--space-4);
    font-size: var(--font-size-sm);
    min-height: 40px;
  }
}

@media (prefers-contrast: more) {
  .input {
    border-width: 2px;
  }

  .btn {
    border-width: 2px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .input,
  .btn,
  .password-toggle,
  .alert,
  .label {
    transition: none;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  @keyframes slideIn {
    from {
      opacity: 0;
    }

    to {
      opacity: 1;
    }
  }
}
</style>
