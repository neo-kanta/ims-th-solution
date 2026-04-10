<script setup lang="ts">
// Auth layout — used for login/register pages
const config = useRuntimeConfig();
const appEnv = config.public.appEnv || "PROD";
const { t } = useI18n();
</script>

<template>
  <div class="auth-layout">
    <!-- Environment Badge -->
    <div
      v-if="appEnv && appEnv !== 'PROD'"
      class="env-badge"
      :class="`env-badge-${appEnv.toLowerCase()}`"
    >
      {{ appEnv }}
    </div>

    <div class="auth-container">
      <div class="auth-header">
        <h1 class="auth-title">
          {{ t("app.name", config.public.appName || "IMS Thailand") }}
        </h1>
        <p class="auth-subtitle">
          {{ t("app.subtitle", "Investment Management System") }}
        </p>
      </div>
      <div class="card">
        <slot />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* =========================================================
   AUTH LAYOUT — LIGHT & DARK THEME STYLES
   ========================================================= */

.auth-layout {
  display: flex;
  min-height: 100vh;
  align-items: center;
  justify-content: center;
  background: var(--bg-app);
  padding: var(--space-6);
  position: relative;
  transition: background-color var(--transition-base);
}

/* Light & Dark theme background with subtle gradient */
.auth-layout {
  background: linear-gradient(
    135deg,
    var(--color-neutral-50) 0%,
    var(--color-neutral-100) 100%
  );
}

/* =========================================================
   ENVIRONMENT BADGE
   ========================================================= */

.env-badge {
  position: fixed;
  top: var(--space-4);
  right: var(--space-4);
  z-index: var(--z-sticky);
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-pill);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.env-badge-dev {
  background-color: var(--color-neutral-200);
  color: var(--color-neutral-700);
}

.env-badge-uat {
  background-color: var(--color-warning-100);
  color: var(--color-warning-800);
  border: 1px solid var(--color-warning-300);
}

/* =========================================================
   CONTAINER & HEADER
   ========================================================= */

.auth-container {
  width: 100%;
  max-width: 420px;
  animation: fadeInUp 0.6s ease-out;
}

.auth-header {
  text-align: center;
  margin-bottom: var(--space-8);
  padding: var(--space-6);
  border-radius: var(--radius-lg);
}

.auth-title {
  margin: 0;
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  letter-spacing: -0.02em;
  background: linear-gradient(
    135deg,
    var(--color-primary-800),
    var(--color-primary-500)
  );
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.auth-subtitle {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-sm);
  color: var(--color-neutral-500);
  font-weight: var(--font-weight-medium);
  letter-spacing: 0.5px;
}

/* =========================================================
   CARD STYLING
   ========================================================= */

:deep(.card) {
  background-color: var(--color-neutral-0);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  box-shadow:
    0 10px 25px -5px rgba(0, 0, 0, 0.05),
    0 8px 10px -6px rgba(0, 0, 0, 0.01);
  padding: var(--space-8);
  transition: box-shadow var(--transition-base);
}

:deep(.card):hover {
  box-shadow:
    0 20px 25px -5px rgba(0, 0, 0, 0.1),
    0 8px 10px -6px rgba(0, 0, 0, 0.01);
}

/* =========================================================
   ANIMATIONS
   ========================================================= */

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* =========================================================
   RESPONSIVE ADJUSTMENTS
   ========================================================= */

@media (max-width: 480px) {
  .auth-layout {
    padding: var(--space-4);
  }

  .auth-container {
    max-width: 100%;
  }

  .auth-header {
    margin-bottom: var(--space-6);
  }

  .auth-title {
    font-size: var(--font-size-2xl);
  }

  .auth-subtitle {
    font-size: var(--font-size-xs);
  }

  :deep(.card) {
    padding: var(--space-6);
    border-radius: var(--radius-md);
  }

  .env-badge {
    top: var(--space-2);
    right: var(--space-2);
  }
}

/* =========================================================
   REDUCED MOTION SUPPORT
   ========================================================= */

@media (prefers-reduced-motion: reduce) {
  .auth-container,
  .auth-title,
  .auth-subtitle,
  :deep(.card) {
    animation: none;
    transition: none;
  }
}

/* =========================================================
   PRINT STYLES
   ========================================================= */

@media print {
  .env-badge {
    display: none;
  }

  .auth-layout {
    background: white;
  }
}
</style>
