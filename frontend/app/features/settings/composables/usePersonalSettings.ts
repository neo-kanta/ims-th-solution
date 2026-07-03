import { ref } from "vue";
import type { PersonalAppearancePreferences } from "../personal.types";
import { personalSettingsApi } from "../services/personalSettingsApi";

export function usePersonalSettings() {
  const appearancePreferences = ref<PersonalAppearancePreferences | null>(null);
  const preferenceLoading = ref(false);
  const preferenceError = ref<string | null>(null);
  
  const { setTheme } = useTheme();

  async function loadAppearancePreferences() {
    preferenceLoading.value = true;
    preferenceError.value = null;
    try {
      const data = await personalSettingsApi.getAppearance();
      appearancePreferences.value = {
        theme: data.theme || "system",
        density: data.density || "comfortable",
        increaseContrast: data.increaseContrast ?? false,
        reduceMotion: data.reduceMotion ?? false,
        showKeyboardHints: data.showKeyboardHints ?? false,
        dateFormat: data.dateFormat || "YYYY-MM-DD",
        timezone: data.timezone || "Asia/Bangkok",
        numberFormat: data.numberFormat || "1,234.56",
        currencyDisplay: data.currencyDisplay || "THB 1,234.56",
        sidebarBehavior: data.sidebarBehavior || "expanded",
        source: data.source || "api",
      };
      
      applyAppearanceToDocument(appearancePreferences.value);
    } catch (error) {
      preferenceError.value = error instanceof Error ? error.message : "Unable to load preferences.";
    } finally {
      preferenceLoading.value = false;
    }
  }

  async function updateAppearancePreferences(patch: Partial<PersonalAppearancePreferences>) {
    if (!appearancePreferences.value) return;
    
    preferenceLoading.value = true;
    try {
      const next = { ...appearancePreferences.value, ...patch };
      // Attempt API save
      appearancePreferences.value = await personalSettingsApi.updateAppearance(next);
      
      applyAppearanceToDocument(appearancePreferences.value);
    } catch (error) {
      console.warn("Failed to save to API, falling back to local state", error);
      // TODO: Implement backend persistence once API is available.
      appearancePreferences.value = { ...appearancePreferences.value, ...patch, source: "local" };
      applyAppearanceToDocument(appearancePreferences.value);
    } finally {
      preferenceLoading.value = false;
    }
  }

  function applyAppearanceToDocument(prefs: PersonalAppearancePreferences) {
    // Map existing themes
    let actualTheme = prefs.theme;
    if (actualTheme === "system") {
      actualTheme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
    }
    
    const baseTheme = actualTheme.includes("dark") ? "dark" : "light";
    
    // We update local setting immediately as requested, via useTheme or directly to document
    setTheme(baseTheme);
    
    if (import.meta.client) {
      const root = document.documentElement;
      
      // We apply standard data attributes. The UI relies on them.
      root.dataset.theme = prefs.theme; // Original selected theme (system, light-high-contrast, etc.)
      root.dataset.density = prefs.density;
      root.dataset.contrast = prefs.increaseContrast ? "high" : "normal";
      root.dataset.reduceMotion = prefs.reduceMotion ? "true" : "false";
      root.dataset.keyboardHints = prefs.showKeyboardHints ? "true" : "false";
      
      // Optional css variables if global css needs them
      if (prefs.increaseContrast) {
        root.style.setProperty("--contrast-boost", "1");
      } else {
        root.style.removeProperty("--contrast-boost");
      }
      
      window.localStorage.setItem("app_sidebar_collapsed", String(prefs.sidebarBehavior === "collapsed"));
    }
  }

  return {
    appearancePreferences,
    preferenceLoading,
    preferenceError,
    loadAppearancePreferences,
    updateAppearancePreferences,
  };
}
