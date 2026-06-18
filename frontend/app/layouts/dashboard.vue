<script setup lang="ts">
import { onClickOutside } from "@vueuse/core";
import { computed, onMounted, onUnmounted, ref, watch } from "vue";

import { buildDashboardNavigation } from "../features/shell/navigation";
import { useDashboardHeaderTabs } from "~/features/shell/composables/useDashboardHeaderTabs";
import AppTabs from "~/shared/ui/AppTabs.vue";
import AppSearch from "~/shared/ui/AppSearch.vue";


const config = useRuntimeConfig();
const router = useRouter();
const route = useRoute();
const appName = config.public.appName as string;
const { t, locale, setLocale } = useI18n();
const authStore = useAuthStore();
const { theme, toggleTheme } = useTheme();

const pageTitle = useState<string>("page-title", () => "");
const isGlobalLoading = useGlobalProgress();

const { hasHeaderTabs, activeTabItems, activeTabValue, activeTabAriaLabel } = useDashboardHeaderTabs();

watch(() => route.path, () => { pageTitle.value = ""; });

const isComplianceRoute = computed(() => route.path.startsWith("/compliance"));

const breadcrumbOwner = computed(() => authStore.user?.username || "neo-kanta");
const breadcrumbRepo = computed(() => {
  if (isComplianceRoute.value) return "compliance";
  if (route.path.startsWith("/investment")) return "investment";
  return appName.toLowerCase().replace(/\s+/g, "-");
});

function handleOwnerClick() {
  void router.push("/");
}

function handleRepoClick() {
  if (breadcrumbRepo.value === "compliance") {
    void router.push("/compliance");
  } else if (breadcrumbRepo.value === "investment") {
    void router.push("/investment/funds");
  } else {
    void router.push("/");
  }
}

const breadcrumbIcon = computed(() => {
  if (isComplianceRoute.value) return "compliance";
  if (route.path.startsWith("/investment")) return "portfolio";
  return "system";
});


// Default to collapsed (closed) so the app loads with content visible,
// matching GitHub's overlay-sidebar UX. Restored from localStorage if set.
const isSidebarCollapsed = ref(true);
const isMobileSidebarOpen = ref(false);
const isMobileViewport = ref(false);
const isSidebarResizing = ref(false);
const showUserMenu = ref(false);
const showLanguageMenu = ref(false);
const isSearchOpen = ref(false);
const sectionOpenState = ref<Record<string, boolean>>({});
const languageMenuRef = ref<HTMLElement | null>(null);
const userMenuRef = ref<HTMLElement | null>(null);

const desktopSidebarWidth = ref(280);
const SIDEBAR_COLLAPSED_WIDTH = 84;
const SIDEBAR_MIN_WIDTH = 248;
const SIDEBAR_MAX_WIDTH = 360;

const navigationSections = computed(() =>
  buildDashboardNavigation(t, authStore.hasPermission),
);

const currentNavTitle = computed(() => {
  const navItem = navigationSections.value
    .flatMap((s) => s.items)
    .find((item) => item.to === route.path);
  return navItem ? navItem.label : "";
});

const languages = computed<Array<{ code: "en" | "th" | "zh"; name: string }>>(
  () => [
    { code: "en", name: t("language.en") },
    { code: "th", name: t("language.th") },
    { code: "zh", name: t("language.zh") },
  ],
);

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

const sidebarInlineWidth = computed(() => `${desktopSidebarWidth.value}px`);

const isSidebarOpen = computed(() => {
  if (isMobileViewport.value) {
    return isMobileSidebarOpen.value;
  }
  return !isSidebarCollapsed.value;
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

  if (!isMobileViewport.value) {
    isSidebarCollapsed.value = true;
    if (import.meta.client) {
      localStorage.setItem("app_sidebar_collapsed", "true");
    }
  }
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

function isSectionOpen(sectionIcon: string): boolean {
  if (sectionIcon in sectionOpenState.value) {
    return sectionOpenState.value[sectionIcon] ?? true;
  }
  return true;
}

function toggleSection(sectionIcon: string): void {
  const wasOpen = isSectionOpen(sectionIcon);
  sectionOpenState.value[sectionIcon] = !wasOpen;
  if (import.meta.client) {
    localStorage.setItem(`app_nav_section_${sectionIcon}`, String(!wasOpen));
  }
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

  // Default is closed; only override if the user explicitly opened it before.
  const storedCollapsed = localStorage.getItem("app_sidebar_collapsed");
  if (storedCollapsed !== null) {
    isSidebarCollapsed.value = storedCollapsed === "true";
  }

  for (const section of navigationSections.value) {
    const saved = localStorage.getItem(`app_nav_section_${section.icon}`);
    if (saved !== null) {
      sectionOpenState.value[section.icon] = saved === "true";
    }
  }

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
      'is-sidebar-open': isSidebarOpen,
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
      :aria-hidden="isMobileViewport ? !isMobileSidebarOpen : undefined"
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
        <ClientOnly>
          <section
            v-for="section in navigationSections"
            :key="section.label"
            class="nav-section"
          >
            <button
              class="nav-section-toggle"
              :class="{ 'is-closed': !isSectionOpen(section.icon) }"
              type="button"
              :aria-expanded="isSectionOpen(section.icon)"
              @click="toggleSection(section.icon)"
            >
              <span class="nav-section-toggle__label">{{ section.label }}</span>
              <AppIcon
                class="nav-section-toggle__chevron"
                name="chevron-down"
                size="xs"
              />
            </button>

            <div
              v-show="
                (isSidebarCollapsed && !isMobileViewport) ||
                isSectionOpen(section.icon)
              "
              class="nav-list"
            >
              <NuxtLink
                v-for="item in section.items"
                :key="item.to"
                :to="item.to"
                class="nav-link"
                :class="{ 'is-active': isNavActive(item.to) }"
                :title="
                  isSidebarCollapsed && !isMobileViewport
                    ? item.label
                    : undefined
                "
                :aria-label="
                  isSidebarCollapsed && !isMobileViewport
                    ? item.label
                    : undefined
                "
                @click="closeSidebar"
              >
                <span class="nav-icon">
                  <AppIcon :name="item.icon" size="sm" />
                </span>
                <span class="nav-text">{{ item.label }}</span>
              </NuxtLink>
            </div>
          </section>
        </ClientOnly>
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
      <header class="app-header" :class="{ 'app-header--with-tabs': hasHeaderTabs }">
        <div class="header-progress" :class="{ 'is-active': isGlobalLoading }"></div>
        
        <!-- Sidebar Toggle Button (Column 1, Row 1) -->
        <button
          class="header-icon-btn header-menu-toggle"
          type="button"
          :aria-expanded="isMobileViewport ? isMobileSidebarOpen : undefined"
          :aria-label="sidebarToggleLabel"
          @click="toggleSidebar"
        >
          <AppIcon name="menu" size="sm" />
        </button>

        <!-- Row 1: Header Top Row (Column 2, Row 1) -->
        <div class="app-header__top-row">
          <!-- Breadcrumbs in GitHub style -->
          <div class="header-breadcrumbs">
            <span class="header-breadcrumbs__icon-wrap">
              <AppIcon :name="breadcrumbIcon" size="sm" class="header-breadcrumbs__icon" />
            </span>
            <span class="header-breadcrumbs__owner" @click="handleOwnerClick">{{ breadcrumbOwner }}</span>
            <span class="header-breadcrumbs__separator">/</span>
            <span
              class="header-breadcrumbs__repo"
              :class="{ 'header-breadcrumbs__repo--link': !!(pageTitle || currentNavTitle) }"
              @click="handleRepoClick"
            >
              {{ breadcrumbRepo }}
            </span>
            <template v-if="pageTitle || currentNavTitle">
              <span class="header-breadcrumbs__separator">/</span>
              <span class="header-breadcrumbs__current-title">{{ pageTitle || currentNavTitle }}</span>
            </template>
          </div>

          <div class="header-controls">
            <button
              class="header-search"
              type="button"
              @click="isSearchOpen = true"
              aria-label="Search or jump to…"
            >
              <AppIcon class="header-search-icon" name="search" size="sm" />
              <span class="header-search-placeholder">
                Type <kbd class="header-search-kbd">/</kbd> to search
              </span>
            </button>

            <!-- Shortcuts -->
            <div class="github-header-items">
              <!-- Issue -->
              <div class="github-header-item">
                <NuxtLink to="/" class="github-header-btn" :aria-label="t('header.issue')">
                  <AppIcon name="issue" size="sm" />
                </NuxtLink>
                <div class="github-tooltip">
                  <AppIcon name="issue" size="xs" />
                  <span>{{ t('header.issue') }}</span>
                </div>
              </div>

              <!-- Pull Request -->
              <div class="github-header-item">
                <NuxtLink to="/" class="github-header-btn" :aria-label="t('header.pullRequest')">
                  <AppIcon name="pull-request" size="sm" />
                </NuxtLink>
                <div class="github-tooltip">
                  <AppIcon name="pull-request" size="xs" />
                  <span>{{ t('header.pullRequest') }}</span>
                </div>
              </div>

              <!-- Compliance -->
              <div class="github-header-item">
                <NuxtLink to="/compliance" class="github-header-btn" :aria-label="t('header.compliance')">
                  <AppIcon name="compliance" size="sm" />
                </NuxtLink>
                <div class="github-tooltip">
                  <AppIcon name="compliance" size="xs" />
                  <span>{{ t('header.compliance') }}</span>
                </div>
              </div>

              <!-- Audit -->
              <div class="github-header-item">
                <NuxtLink to="/administration/settings" class="github-header-btn" :aria-label="t('header.audit')">
                  <AppIcon name="audit" size="sm" />
                </NuxtLink>
                <div class="github-tooltip">
                  <AppIcon name="audit" size="xs" />
                  <span>{{ t('header.audit') }}</span>
                </div>
              </div>
            </div>

            <span class="github-header-separator"></span>

            <div ref="userMenuRef" class="header-menu">
              <ClientOnly>
                <button
                  class="github-user-btn"
                  type="button"
                  aria-haspopup="menu"
                  :aria-expanded="showUserMenu"
                  @click="toggleUserMenu"
                  :aria-label="userDisplayName"
                >
                  <span class="github-user-avatar">{{ userInitials }}</span>
                </button>
                <template #fallback>
                  <button class="github-user-btn" type="button" disabled>
                    <span class="github-user-avatar">·</span>
                  </button>
                </template>
              </ClientOnly>

              <div
                v-if="showUserMenu"
                class="header-dropdown user-menu-dropdown"
                role="menu"
              >
                <div class="user-menu-info">
                  <div class="user-menu-info-label">{{ userDisplayName }}</div>
                  <div class="user-menu-info-value">{{ userSubtitle }}</div>
                </div>

                <NuxtLink
                  class="header-dropdown-item"
                  :class="{ 'is-active': isNavActive('/settings') }"
                  to="/settings"
                  role="menuitem"
                  :aria-current="isNavActive('/settings') ? 'page' : undefined"
                  @click="showUserMenu = false"
                >
                  <span>{{ t("navigation.personalSettings") }}</span>
                </NuxtLink>

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
        </div>

        <!-- Row 2: Tabs Row (rendered dynamically for Compliance and active Fund contexts) -->
        <div v-if="hasHeaderTabs" class="app-header__tabs-row">
          <AppTabs
            v-slot:default
            :items="activeTabItems"
            :model-value="activeTabValue"
            :aria-label="activeTabAriaLabel"
          />
        </div>
      </header>

      <main class="app-content">
        <div class="page-container">
          <slot />
        </div>
      </main>
    </div>
    <!-- Advanced Search Component -->
    <AppSearch v-model="isSearchOpen" />
  </div>
</template>

<style scoped>
.app-header {
  height: auto !important;
  min-height: var(--header-height);
  display: grid;
  grid-template-columns: auto 1fr;
  grid-template-rows: var(--header-height); /* Default: 1 row mode */
  align-items: center;
  padding: 0 var(--space-7);
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-header);
}

.header-menu-toggle {
  grid-column: 1;
  grid-row: 1;
  margin-right: var(--space-4);
}

.app-header__top-row {
  grid-column: 2;
  grid-row: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.header-breadcrumbs {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  min-width: 0;
  text-transform: lowercase;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.app-header__tabs-row {
  grid-column: 2;
  grid-row: 2;
  display: flex;
  align-items: flex-end;
  width: 100%;
  padding: 0;
}

.header-breadcrumbs__icon-wrap {
  display: flex;
  align-items: center;
  color: var(--text-secondary);
}

.header-breadcrumbs__owner {
  color: var(--text-link);
  cursor: pointer;
}

.header-breadcrumbs__owner:hover {
  text-decoration: underline;
}

.header-breadcrumbs__separator {
  color: var(--text-tertiary);
  font-weight: var(--font-weight-regular);
}

.header-breadcrumbs__repo {
  font-weight: var(--font-weight-bold);
}

.header-breadcrumbs__repo--link {
  color: var(--text-link) !important;
  font-weight: var(--font-weight-semibold) !important;
  cursor: pointer;
}

.header-breadcrumbs__repo--link:hover {
  text-decoration: underline;
}

.header-breadcrumbs__current-title {
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
}

.app-header--with-tabs {
  grid-template-rows: var(--header-height) auto; /* 2 rows mode */
  padding-bottom: 0;
}
</style>
