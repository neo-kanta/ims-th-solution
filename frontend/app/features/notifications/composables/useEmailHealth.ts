import { ref } from "vue";
import { notificationApi, notificationErrorMessage } from "../services/notificationApi";
import type { EmailHealth } from "../types";

export function useEmailHealth() {
  const health = ref<EmailHealth | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function load() {
    loading.value = true;
    error.value = null;
    try {
      health.value = await notificationApi.emailHealth();
    } catch (e) {
      error.value = notificationErrorMessage(e, "Unable to load email health");
    } finally {
      loading.value = false;
    }
  }

  return { health, loading, error, load };
}
