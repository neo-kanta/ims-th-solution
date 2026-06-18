<script setup lang="ts">
// Leave & Delegation is Phase 2 and OUT OF SCOPE for the current IMS demo.
//
// This component used to render a placeholder leave/delegation form. The form
// could be mistaken for a working feature, so it was removed. The component is
// kept in the source tree only because legacy imports may exist; it now
// renders the same Phase 2 notice as /pages/leave/index.vue.
//
// Do NOT add a working leave/delegation form here without first:
//   1. shipping the backend leave-request entity + REST routes,
//   2. wiring a non-Nop LeaveChecker / DelegateResolver in main.go,
//   3. adding the IAM login-time leave gate.
</script>

<template>
  <section class="phase2-shell">
    <header class="phase2-header">
      <h2 class="phase2-title">Leave & Delegation</h2>
      <span class="phase2-badge">Phase 2 — Not implemented</span>
    </header>
    <p class="phase2-body">
      Self-service leave, temporary leave, cancellation, and delegate
      assignment are not part of the current IMS demo scope. Approval tasks are
      routed directly to the configured approvers without any delegate skip or
      leave-aware reassignment.
    </p>
  </section>
</template>

<style scoped>
.phase2-shell {
  max-width: 720px;
  margin: var(--space-4) auto;
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
  margin-bottom: var(--space-3);
}
.phase2-title {
  margin: 0;
  font-size: var(--font-size-md);
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
