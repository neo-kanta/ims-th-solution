<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";

import type { PersonalMfaTotpInput } from "../account.types";

interface Props {
  open: boolean;
  loading: boolean;
  error: string | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  cancel: [];
  confirm: [payload: PersonalMfaTotpInput];
}>();

const dialogRef = ref<HTMLElement | null>(null);
const totpInput = ref("");
let previouslyFocused: HTMLElement | null = null;
const { t } = useI18n();

const totpDigits = computed(() => totpInput.value.replace(/\D/g, "").slice(0, 6));
const canConfirm = computed(() => !props.loading && totpDigits.value.length === 6);

watch(
  () => props.open,
  (open) => {
    if (!import.meta.client) return;
    if (open) {
      previouslyFocused = document.activeElement as HTMLElement | null;
      void nextTick(() => {
        dialogRef.value?.querySelector<HTMLElement>("input")?.focus();
      });
    } else {
      totpInput.value = "";
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

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape" && !props.loading) {
    event.preventDefault();
    emit("cancel");
  }
}

function submit() {
  if (!canConfirm.value) return;
  emit("confirm", { totp_code: totpDigits.value });
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-backdrop settings-mfa-disable"
      role="presentation"
      @keydown="handleKeydown"
    >
      <section
        ref="dialogRef"
        class="modal settings-mfa-disable__dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="settings-mfa-disable-title"
        tabindex="-1"
      >
        <header class="modal-header">
          <h2 id="settings-mfa-disable-title">
            {{ t("settings.console.personal.mfaDisable.title") }}
          </h2>
          <p>{{ t("settings.console.personal.mfaDisable.body") }}</p>
        </header>

        <div class="modal-body">
          <div class="form-group" :class="{ 'field-error': error }">
            <label for="settings-mfa-disable-totp" class="label label-required">
              {{ t("settings.console.personal.mfaDisable.totpLabel") }}
            </label>
            <input
              id="settings-mfa-disable-totp"
              v-model="totpInput"
              class="input settings-mfa-disable__input"
              type="text"
              inputmode="numeric"
              autocomplete="one-time-code"
              pattern="[0-9]{6}"
              maxlength="6"
              :aria-invalid="Boolean(error)"
            />
            <div v-if="error" class="error-text">{{ error }}</div>
          </div>
        </div>

        <footer class="modal-footer">
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="loading"
            @click="emit('cancel')"
          >
            {{ t("settings.console.common.cancel") }}
          </button>
          <button
            type="button"
            class="btn btn-danger btn-sm"
            :disabled="!canConfirm"
            @click="submit"
          >
            {{
              loading
                ? t("settings.console.personal.mfaDisable.disabling")
                : t("settings.console.personal.mfaDisable.confirmCta")
            }}
          </button>
        </footer>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.settings-mfa-disable { align-items: center; }
.settings-mfa-disable__dialog { width: min(28rem, 100%); }
.settings-mfa-disable__dialog:focus { outline: none; }
.settings-mfa-disable__input {
  letter-spacing: 0.4em;
  font-family: var(--font-family-mono);
  text-align: center;
  font-size: var(--font-size-lg);
  max-width: 12rem;
}
</style>
