<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import QRCode from "qrcode";

import type {
  PersonalMfaEnrollResult,
  PersonalMfaTotpInput,
} from "../account.types";

interface Props {
  open: boolean;
  enrollment: PersonalMfaEnrollResult | null;
  enrolling: boolean;
  verifying: boolean;
  enrollError: string | null;
  verifyError: string | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  cancel: [];
  startEnroll: [];
  verify: [payload: PersonalMfaTotpInput];
}>();

const dialogRef = ref<HTMLElement | null>(null);
const totpInput = ref("");
const codesAcknowledged = ref(false);
const qrDataUrl = ref<string | null>(null);
const qrError = ref<string | null>(null);
let previouslyFocused: HTMLElement | null = null;
const { t } = useI18n();

const step = computed<"intro" | "verify">(() =>
  props.enrollment ? "verify" : "intro",
);

const totpDigits = computed(() => totpInput.value.replace(/\D/g, "").slice(0, 6));
const canVerify = computed(
  () =>
    !props.verifying &&
    codesAcknowledged.value &&
    totpDigits.value.length === 6,
);

watch(
  () => props.enrollment?.provisioning_uri,
  async (uri) => {
    qrError.value = null;
    if (!uri) {
      qrDataUrl.value = null;
      return;
    }
    try {
      qrDataUrl.value = await QRCode.toDataURL(uri, {
        margin: 1,
        width: 220,
        errorCorrectionLevel: "M",
      });
    } catch (err) {
      qrDataUrl.value = null;
      qrError.value = String((err as Error)?.message ?? err);
    }
  },
  { immediate: true },
);

watch(
  () => props.open,
  (open) => {
    if (!import.meta.client) return;
    if (open) {
      previouslyFocused = document.activeElement as HTMLElement | null;
      void nextTick(() => focusFirst());
    } else {
      totpInput.value = "";
      codesAcknowledged.value = false;
      if (previouslyFocused && document.body.contains(previouslyFocused)) {
        previouslyFocused.focus();
        previouslyFocused = null;
      }
    }
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  previouslyFocused = null;
});

function focusableElements(): HTMLElement[] {
  if (!dialogRef.value) return [];
  return Array.from(
    dialogRef.value.querySelectorAll<HTMLElement>(
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  );
}

function focusFirst() {
  focusableElements()[0]?.focus();
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    if (!props.enrolling && !props.verifying) {
      event.preventDefault();
      emit("cancel");
    }
    return;
  }
  if (event.key !== "Tab") return;

  const items = focusableElements();
  if (items.length === 0) {
    event.preventDefault();
    dialogRef.value?.focus();
    return;
  }
  const first = items[0];
  const last = items[items.length - 1];
  if (!first || !last) return;
  const active = document.activeElement as HTMLElement | null;
  if (event.shiftKey && (active === first || !dialogRef.value?.contains(active))) {
    event.preventDefault();
    last.focus();
  } else if (!event.shiftKey && active === last) {
    event.preventDefault();
    first.focus();
  }
}

function copyRecoveryCodes() {
  if (!props.enrollment) return;
  if (!import.meta.client || !navigator.clipboard) return;
  void navigator.clipboard.writeText(props.enrollment.recovery_codes.join("\n"));
}

function startEnroll() {
  emit("startEnroll");
}

function submitVerify() {
  if (!canVerify.value) return;
  emit("verify", { totp_code: totpDigits.value });
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-backdrop settings-mfa-enroll"
      role="presentation"
      @keydown="handleKeydown"
    >
      <section
        ref="dialogRef"
        class="modal settings-mfa-enroll__dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="settings-mfa-enroll-title"
        tabindex="-1"
      >
        <header class="modal-header settings-mfa-enroll__header">
          <h2 id="settings-mfa-enroll-title">
            {{ t("settings.console.personal.mfaEnroll.title") }}
          </h2>
          <p>{{ t("settings.console.personal.mfaEnroll.subtitle") }}</p>
        </header>

        <div class="modal-body settings-mfa-enroll__body">
          <template v-if="step === 'intro'">
            <p>{{ t("settings.console.personal.mfaEnroll.introBody") }}</p>
            <ul class="settings-mfa-enroll__steps">
              <li>{{ t("settings.console.personal.mfaEnroll.step1") }}</li>
              <li>{{ t("settings.console.personal.mfaEnroll.step2") }}</li>
              <li>{{ t("settings.console.personal.mfaEnroll.step3") }}</li>
            </ul>
            <div v-if="enrollError" class="alert alert-danger" role="alert">
              {{ enrollError }}
            </div>
          </template>

          <template v-else-if="enrollment">
            <ol class="settings-mfa-enroll__verify-steps">
              <li>
                <strong>{{ t("settings.console.personal.mfaEnroll.scanTitle") }}</strong>
                <p>{{ t("settings.console.personal.mfaEnroll.scanBody") }}</p>
                <div class="settings-mfa-enroll__qr">
                  <img
                    v-if="qrDataUrl"
                    :src="qrDataUrl"
                    :alt="t('settings.console.personal.mfaEnroll.qrAlt')"
                    width="220"
                    height="220"
                  />
                  <p v-else-if="qrError" class="error-text">{{ qrError }}</p>
                </div>
                <details class="settings-mfa-enroll__manual">
                  <summary>{{ t("settings.console.personal.mfaEnroll.manualToggle") }}</summary>
                  <code>{{ enrollment.provisioning_uri }}</code>
                </details>
              </li>

              <li>
                <strong>{{ t("settings.console.personal.mfaEnroll.recoveryTitle") }}</strong>
                <p>{{ t("settings.console.personal.mfaEnroll.recoveryBody") }}</p>
                <ul class="settings-mfa-enroll__codes" aria-label="Recovery codes">
                  <li v-for="code in enrollment.recovery_codes" :key="code">
                    {{ code }}
                  </li>
                </ul>
                <button
                  type="button"
                  class="btn btn-ghost btn-sm settings-mfa-enroll__copy"
                  @click="copyRecoveryCodes"
                >
                  {{ t("settings.console.personal.mfaEnroll.copyCodes") }}
                </button>
                <label class="settings-mfa-enroll__ack">
                  <input
                    v-model="codesAcknowledged"
                    type="checkbox"
                  />
                  {{ t("settings.console.personal.mfaEnroll.ackLabel") }}
                </label>
              </li>

              <li>
                <strong>{{ t("settings.console.personal.mfaEnroll.verifyTitle") }}</strong>
                <p>{{ t("settings.console.personal.mfaEnroll.verifyBody") }}</p>
                <div class="form-group" :class="{ 'field-error': verifyError }">
                  <label for="settings-mfa-totp" class="label label-required">
                    {{ t("settings.console.personal.mfaEnroll.totpLabel") }}
                  </label>
                  <input
                    id="settings-mfa-totp"
                    v-model="totpInput"
                    class="input settings-mfa-enroll__totp-input"
                    type="text"
                    inputmode="numeric"
                    autocomplete="one-time-code"
                    pattern="[0-9]{6}"
                    maxlength="6"
                    :aria-invalid="Boolean(verifyError)"
                  />
                  <div v-if="verifyError" class="error-text">{{ verifyError }}</div>
                </div>
              </li>
            </ol>
          </template>
        </div>

        <footer class="modal-footer settings-mfa-enroll__footer">
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="enrolling || verifying"
            @click="emit('cancel')"
          >
            {{ t("settings.console.common.cancel") }}
          </button>
          <button
            v-if="step === 'intro'"
            type="button"
            class="btn btn-primary btn-sm"
            :disabled="enrolling"
            @click="startEnroll"
          >
            {{
              enrolling
                ? t("settings.console.personal.mfaEnroll.starting")
                : t("settings.console.personal.mfaEnroll.startCta")
            }}
          </button>
          <button
            v-else
            type="button"
            class="btn btn-primary btn-sm"
            :disabled="!canVerify"
            @click="submitVerify"
          >
            {{
              verifying
                ? t("settings.console.personal.mfaEnroll.verifying")
                : t("settings.console.personal.mfaEnroll.verifyCta")
            }}
          </button>
        </footer>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.settings-mfa-enroll {
  align-items: center;
}
.settings-mfa-enroll__dialog {
  width: min(38rem, 100%);
  max-height: calc(100vh - var(--space-7) * 2);
  display: flex;
  flex-direction: column;
}
.settings-mfa-enroll__dialog:focus { outline: none; }
.settings-mfa-enroll__header h2 {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}
.settings-mfa-enroll__header p {
  margin: var(--space-2) 0 0;
  color: var(--text-secondary);
}
.settings-mfa-enroll__body {
  overflow-y: auto;
  display: grid;
  gap: var(--space-4);
}
.settings-mfa-enroll__steps {
  margin: 0;
  padding-left: var(--space-5);
  color: var(--text-secondary);
}
.settings-mfa-enroll__verify-steps {
  margin: 0;
  padding-left: var(--space-5);
  display: grid;
  gap: var(--space-5);
}
.settings-mfa-enroll__verify-steps li > strong {
  display: block;
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
}
.settings-mfa-enroll__verify-steps li > p {
  margin: var(--space-1) 0 var(--space-3);
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}
.settings-mfa-enroll__qr {
  display: flex;
  justify-content: center;
  padding: var(--space-4);
  background: var(--bg-card-muted);
  border-radius: var(--radius-md);
}
.settings-mfa-enroll__manual {
  margin-top: var(--space-3);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}
.settings-mfa-enroll__manual code {
  display: block;
  margin-top: var(--space-2);
  padding: var(--space-3);
  background: var(--bg-card-muted);
  border-radius: var(--radius-sm);
  word-break: break-all;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
}
.settings-mfa-enroll__codes {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-2);
  margin: 0;
  padding: var(--space-3);
  background: var(--bg-card-muted);
  border-radius: var(--radius-md);
  list-style: none;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
}
.settings-mfa-enroll__codes li {
  padding: var(--space-2);
  background: var(--bg-card);
  border-radius: var(--radius-sm);
  text-align: center;
  user-select: all;
}
.settings-mfa-enroll__copy {
  margin-top: var(--space-2);
}
.settings-mfa-enroll__ack {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-3);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  cursor: pointer;
}
.settings-mfa-enroll__totp-input {
  letter-spacing: 0.4em;
  font-family: var(--font-family-mono);
  text-align: center;
  font-size: var(--font-size-lg);
  max-width: 12rem;
}
@media (max-width: 640px) {
  .settings-mfa-enroll__codes { grid-template-columns: 1fr; }
}
</style>
