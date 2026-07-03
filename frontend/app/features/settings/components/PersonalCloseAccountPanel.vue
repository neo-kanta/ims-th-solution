<template>
  <section class="settings-panel" style="border-color: var(--alert-danger-border);">
    <header class="settings-panel__header">
      <h2 class="settings-panel__title" style="color: var(--alert-danger-text);">Close account</h2>
    </header>
    <div class="settings-panel__body" style="padding: var(--space-5);">
      <p class="text-secondary" style="margin-bottom: var(--space-4);">
        Closing your account will permanently remove your access to IMS. This action is not reversible.
      </p>
      <AppButton variant="danger" disabled title="Account closure requires administrator approval">
        Close my account
      </AppButton>
    </div>
  </section>
</template>
