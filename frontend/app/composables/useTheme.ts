import { computed, ref } from "vue";
import { useCookie, useHead, useState } from "#imports";

type Theme = "light" | "dark" | "system";

const isDarkModePreferred = ref(false);
let mediaQueryBound = false;

function bindSystemThemeListener() {
  if (mediaQueryBound || !process.client) {
    return;
  }

  mediaQueryBound = true;
  const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
  isDarkModePreferred.value = mediaQuery.matches;
  mediaQuery.addEventListener("change", (event) => {
    isDarkModePreferred.value = event.matches;
  });
}

export function useTheme() {
  const themeCookie = useCookie<Theme>("app_theme", {
    default: () => "light",
    sameSite: "lax",
    path: "/",
  });
  const themeState = useState<Theme>("app-theme", () => themeCookie.value || "light");

  bindSystemThemeListener();

  const theme = computed({
    get: () => themeState.value,
    set: (value: Theme) => {
      themeState.value = value;
      themeCookie.value = value;
    },
  });

  const effectiveTheme = computed<"light" | "dark">(() => {
    if (theme.value === "system") {
      return isDarkModePreferred.value ? "dark" : "light";
    }

    return theme.value;
  });

  useHead(() => ({
    htmlAttrs: {
      "data-theme": effectiveTheme.value,
    },
  }));

  return {
    theme,
    isDark: computed(() => effectiveTheme.value === "dark"),
    setTheme: (newTheme: Theme) => {
      theme.value = newTheme;
    },
    toggleTheme: () => {
      theme.value = effectiveTheme.value === "light" ? "dark" : "light";
    },
  };
}
