import { ref } from "vue";
import { notificationApi, notificationErrorMessage } from "../services/notificationApi";
import type { TestEmailBody, TestEmailResult } from "../types";

export function useTestEmail() {
  const sending = ref(false);
  const result = ref<TestEmailResult | null>(null);
  const error = ref<string | null>(null);

  // Note: allow_raw_email is absent from the generated healthResponse schema.
  // We always use to_username mode and never expose raw email input.
  const form = ref<TestEmailBody>({
    to_username: "",
    subject: "IMS Test Email",
    body: "This is a test email from the IMS notification system.",
  });

  function reset() {
    result.value = null;
    error.value = null;
    form.value = {
      to_username: "",
      subject: "IMS Test Email",
      body: "This is a test email from the IMS notification system.",
    };
  }

  async function send() {
    sending.value = true;
    error.value = null;
    result.value = null;
    try {
      result.value = await notificationApi.sendTestEmail(form.value);
    } catch (e) {
      error.value = notificationErrorMessage(e, "Unable to send test email");
    } finally {
      sending.value = false;
    }
  }

  return { form, sending, result, error, send, reset };
}
