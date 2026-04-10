<script setup lang="ts">
import { onClickOutside } from "@vueuse/core";
import { computed, onMounted, onUnmounted, ref, watch } from "vue";

import { buildDashboardNavigation } from "../features/shell/navigation";

const config = useRuntimeConfig();
const router = useRouter();
const route = useRoute();
const appName = config.public.appName as string;
const { t, locale, setLocale } = useI18n();
const authStore = useAuthStore();
const { theme, toggleTheme } = useTheme();

const isSidebarCollapsed = ref(false);
const isMobileSidebarOpen = ref(false);
const isMobileViewport = ref(false);
const isSidebarResizing = ref(false);
const showUserMenu = ref(false);
const showLanguageMenu = ref(false);
const languageMenuRef = ref<HTMLElement | null>(null);
const userMenuRef = ref<HTMLElement | null>(null);

const desktopSidebarWidth = ref(280);
const SIDEBAR_COLLAPSED_WIDTH = 84;
const SIDEBAR_MIN_WIDTH = 248;
const SIDEBAR_MAX_WIDTH = 360;

const navigationSections = computed(() =>
  buildDashboardNavigation(t, authStore.hasPermission),
);
const languages = computed(() => [
  { code: "en", name: t("language.en") },
  { code: "th", name: t("language.th") },
  { code: "zh", name: t("language.zh") },
]);

const userDisplayName = computed(
  () => authStore.user?.displayName || t("auth.welcome"),
);

const userSubtitle = computed(
  () => authStore.user?.username || t("shell.activeSession"),
);

const userInitials = computed(() => {
  const source =
    authStore.user?.displayName || authStore.user?.username || "IMS";
  const parts = source.split(/\s+/).filter(Boolean).slice(0, 2);

  if (parts.length === 0) {
    return "IM";
  }

  return parts.map((part) => part[0]?.toUpperCase() || "").join("");
});

const themeToggleLabel = computed(() =>
  theme.value === "dark"
    ? t("theme.switchToLight", "Switch to light mode")
    : t("theme.switchToDark", "Switch to dark mode"),
);

const sidebarToggleLabel = computed(() => {
  if (isMobileViewport.value) {
    return isMobileSidebarOpen.value
      ? t("shell.closeNavigation")
      : t("shell.openNavigation");
  }

  return isSidebarCollapsed.value
    ? t("shell.expandNavigation")
    : t("shell.collapseNavigation");
});

const sidebarToggleIcon = computed(() => {
  if (isMobileViewport.value) {
    return isMobileSidebarOpen.value ? "close" : "menu";
  }

  return isSidebarCollapsed.value ? "panel-open" : "panel-close";
});

const sidebarInlineWidth = computed(() => {
  if (isMobileViewport.value) {
    return `${desktopSidebarWidth.value}px`;
  }

  return isSidebarCollapsed.value
    ? `${SIDEBAR_COLLAPSED_WIDTH}px`
    : `${desktopSidebarWidth.value}px`;
});

function syncViewport() {
  if (!import.meta.client) {
    return;
  }

  isMobileViewport.value = window.innerWidth < 1024;

  if (!isMobileViewport.value) {
    isMobileSidebarOpen.value = false;
  }
}

function closeSidebar() {
  isMobileSidebarOpen.value = false;
}

function startSidebarResize(event: PointerEvent) {
  if (isMobileViewport.value || isSidebarCollapsed.value) {
    return;
  }

  isSidebarResizing.value = true;
  document.body.style.cursor = "col-resize";
  document.body.style.userSelect = "none";
  window.addEventListener("pointermove", onSidebarResize);
  window.addEventListener("pointerup", stopSidebarResize);
  event.preventDefault();
}

function onSidebarResize(event: PointerEvent) {
  if (!isSidebarResizing.value) {
    return;
  }

  desktopSidebarWidth.value = Math.min(
    SIDEBAR_MAX_WIDTH,
    Math.max(SIDEBAR_MIN_WIDTH, event.clientX),
  );
}

function stopSidebarResize() {
  if (!isSidebarResizing.value) {
    return;
  }

  isSidebarResizing.value = false;
  document.body.style.cursor = "";
  document.body.style.userSelect = "";
  window.removeEventListener("pointermove", onSidebarResize);
  window.removeEventListener("pointerup", stopSidebarResize);

  if (import.meta.client) {
    localStorage.setItem(
      "app_sidebar_width",
      String(desktopSidebarWidth.value),
    );
  }
}

function toggleSidebar() {
  if (isMobileViewport.value) {
    isMobileSidebarOpen.value = !isMobileSidebarOpen.value;
    return;
  }

  isSidebarCollapsed.value = !isSidebarCollapsed.value;

  if (import.meta.client) {
    localStorage.setItem(
      "app_sidebar_collapsed",
      String(isSidebarCollapsed.value),
    );
  }
}

function toggleLanguageMenu() {
  showLanguageMenu.value = !showLanguageMenu.value;

  if (showLanguageMenu.value) {
    showUserMenu.value = false;
  }
}

function toggleUserMenu() {
  showUserMenu.value = !showUserMenu.value;

  if (showUserMenu.value) {
    showLanguageMenu.value = false;
  }
}

function setAppLocale(code: "en" | "th" | "zh") {
  setLocale(code);
  showLanguageMenu.value = false;
}

function isNavActive(to: string) {
  return to === "/"
    ? route.path === "/"
    : route.path === to || route.path.startsWith(`${to}/`);
}

async function handleLogout() {
  showUserMenu.value = false;
  await authStore.logout();
  closeSidebar();
  await router.push("/auth/login");
}

function handleEscape(event: KeyboardEvent) {
  if (event.key !== "Escape") {
    return;
  }

  showLanguageMenu.value = false;
  showUserMenu.value = false;
  closeSidebar();
}

onClickOutside(languageMenuRef, () => {
  showLanguageMenu.value = false;
});

onClickOutside(userMenuRef, () => {
  showUserMenu.value = false;
});

onMounted(() => {
  if (!import.meta.client) {
    return;
  }

  const savedWidthRaw = localStorage.getItem("app_sidebar_width");
  const savedWidth = savedWidthRaw ? Number(savedWidthRaw) : NaN;
  if (!Number.isNaN(savedWidth)) {
    desktopSidebarWidth.value = Math.min(
      SIDEBAR_MAX_WIDTH,
      Math.max(SIDEBAR_MIN_WIDTH, savedWidth),
    );
  }

  isSidebarCollapsed.value =
    localStorage.getItem("app_sidebar_collapsed") === "true";

  syncViewport();
  window.addEventListener("resize", syncViewport);
  document.addEventListener("keydown", handleEscape);
});

onUnmounted(() => {
  if (!import.meta.client) {
    return;
  }

  stopSidebarResize();
  window.removeEventListener("resize", syncViewport);
  document.removeEventListener("keydown", handleEscape);
});

watch(
  () => route.fullPath,
  () => {
    showLanguageMenu.value = false;
    showUserMenu.value = false;

    if (isMobileViewport.value) {
      closeSidebar();
    }
  },
);
</script>

<template>
  <div
    class="app-shell"
    :class="{
      'is-collapsed': isSidebarCollapsed && !isMobileViewport,
      'is-mobile-sidebar-open': isMobileSidebarOpen,
      'is-resizing': isSidebarResizing,
    }"
    :style="{ '--sidebar-width': sidebarInlineWidth }"
  >
    <button
      class="sidebar-backdrop"
      type="button"
      :aria-label="t('shell.closeNavigation')"
      @click="closeSidebar"
    />

    <aside
      class="app-sidebar"
      :class="{ 'is-open': isMobileSidebarOpen }"
      :aria-hidden="isMobileViewport ? String(!isMobileSidebarOpen) : undefined"
    >
      <div class="sidebar-header">
        <NuxtLink class="sidebar-brand" to="/" @click="closeSidebar">
          <span class="sidebar-brand-mark">IM</span>
          <span class="sidebar-brand-copy">
            <span class="sidebar-brand-name">{{ appName }}</span>
            <span class="sidebar-brand-meta">
              {{ t("shell.workspaceLabel") }}
            </span>
          </span>
        </NuxtLink>

        <button
          class="sidebar-mobile-close header-icon-btn"
          type="button"
          :aria-label="t('shell.closeNavigation')"
          @click="closeSidebar"
        >
          <AppIcon name="close" />
        </button>
      </div>

      <nav
        class="sidebar-nav-container"
        :aria-label="t('shell.primaryNavigation')"
      >
        <section
          v-for="section in navigationSections"
          :key="section.label"
          class="nav-section"
        >
          <div class="nav-section-label">{{ section.label }}</div>
          <div class="nav-list">
            <NuxtLink
              v-for="item in section.items"
              :key="item.to"
              :to="item.to"
              class="nav-link"
              :class="{ 'is-active': isNavActive(item.to) }"
              :title="isSidebarCollapsed ? item.label : undefined"
              :aria-label="isSidebarCollapsed ? item.label : undefined"
              @click="closeSidebar"
            >
              <span class="nav-icon">
                <AppIcon :name="item.icon" />
              </span>
              <span class="nav-text">{{ item.label }}</span>
            </NuxtLink>
          </div>
        </section>
      </nav>

      <button
        v-if="!isMobileViewport && !isSidebarCollapsed"
        class="sidebar-resizer"
        type="button"
        :aria-label="t('shell.resizeNavigation')"
        @pointerdown="startSidebarResize"
      />
    </aside>

    <div class="app-main">
      <header class="app-header">
        <div class="header-leading">
          <button
            class="header-icon-btn"
            type="button"
            :aria-expanded="
              isMobileViewport ? String(isMobileSidebarOpen) : undefined
            "
            :aria-label="sidebarToggleLabel"
            @click="toggleSidebar"
          >
            <AppIcon :name="sidebarToggleIcon" />
          </button>

          <div class="header-context">
            <div class="header-context-label">
              {{ t("shell.workspaceLabel") }}
            </div>
            <div class="header-context-title">{{ appName }}</div>
          </div>
        </div>

        <div class="header-controls">
          <div ref="languageMenuRef" class="header-menu">
            <button
              class="header-control-btn"
              type="button"
              aria-haspopup="menu"
              :aria-expanded="String(showLanguageMenu)"
              @click="toggleLanguageMenu"
            >
              <AppIcon name="globe" />
              <span class="header-control-label">{{
                locale.toUpperCase()
              }}</span>
              <AppIcon
                class="header-control-chevron"
                name="chevron-down"
                size="xs"
              />
            </button>

            <div v-if="showLanguageMenu" class="header-dropdown" role="menu">
              <button
                v-for="language in languages"
                :key="language.code"
                class="header-dropdown-item"
                :class="{ 'is-active': locale === language.code }"
                type="button"
                role="menuitemradio"
                :aria-checked="locale === language.code"
                @click="setAppLocale(language.code)"
              >
                <span>{{ language.name }}</span>
                <AppIcon
                  v-if="locale === language.code"
                  name="check"
                  size="xs"
                />
              </button>
            </div>
          </div>

          <button
            class="header-icon-btn"
            type="button"
            :title="themeToggleLabel"
            :aria-label="themeToggleLabel"
            @click="toggleTheme"
          >
            <AppIcon :name="theme === 'dark' ? 'sun' : 'moon'" />
          </button>

          <div ref="userMenuRef" class="header-menu">
            <button
              class="user-menu-btn"
              type="button"
              aria-haspopup="menu"
              :aria-expanded="String(showUserMenu)"
              @click="toggleUserMenu"
            >
              <span class="user-avatar">{{ userInitials }}</span>
              <span class="user-menu-copy">
                <span class="user-menu-name">{{ userDisplayName }}</span>
                <span class="user-menu-role">{{ userSubtitle }}</span>
              </span>
              <AppIcon
                class="header-control-chevron"
                name="chevron-down"
                size="xs"
              />
            </button>

            <div
              v-if="showUserMenu"
              class="header-dropdown user-menu-dropdown"
              role="menu"
            >
              <div class="user-menu-info">
                <div class="user-menu-info-label">{{ userDisplayName }}</div>
                <div class="user-menu-info-value">{{ userSubtitle }}</div>
              </div>

              <button
                class="header-dropdown-item user-menu-logout"
                type="button"
                role="menuitem"
                @click="handleLogout"
              >
                {{ t("auth.logout") }}
              </button>
            </div>
          </div>
        </div>
      </header>

      <main class="app-content">
        <div class="page-container">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>
