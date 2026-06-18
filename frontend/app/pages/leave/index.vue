<script setup lang="ts">
// Leave & Delegation is Phase 2 and intentionally OUT OF SCOPE for the
// current IMS demo. The backend has no leave / delegation endpoints wired,
// and the approval engine uses no-op leave/delegate stubs. This page exists
// so direct navigation does not 404, but it does not pretend the feature
// works.
definePageMeta({
  layout: "dashboard",
  middleware: ["auth"],
});
</script>

<template>
  <section class="phase2-shell">
    <header class="phase2-header">
      <h1 class="phase2-title">Leave & Delegation</h1>
      <span class="phase2-badge">Phase 2 — Not implemented</span>
    </header>
    <p class="phase2-body">
      The Leave &amp; Delegation module is not part of the current IMS demo
      scope. Approval routing currently uses the configured approvers
      directly and does not consult leave or delegate status.
    </p>
  </section>
</template>

<style scoped>
.phase2-shell {
  max-width: 720px;
  margin: var(--space-6) auto;
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-elevated);
}
.phase2-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}
.phase2-title {
  margin: 0;
  font-size: var(--font-size-lg);
  color: var(--text-primary);
}
.phase2-badge {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-pill);
  background: var(--bg-warning-muted, #fef3c7);
  color: var(--text-warning, #92400e);
}
.phase2-body {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.55;
}
</style>
