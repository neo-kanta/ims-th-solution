<script setup lang="ts">
import DecisionStatusBadge from "./DecisionStatusBadge.vue";
import type { ApiDecision } from "../services/decisionApi";

defineProps<{ decision: ApiDecision }>();
</script>

<template>
  <div class="panel-header">Decision Detail</div>
  <div class="legacy-form">
    <div class="form-row">
      <div class="form-field">
        <label>Decision No</label>
        <input :value="decision.decision_number ?? '—'" disabled class="highlighted-text" />
      </div>
      <div class="form-field">
        <label>Status</label>
        <div class="field-value">
          <DecisionStatusBadge :status="decision.status" />
        </div>
      </div>
    </div>

    <div class="form-row">
      <div class="form-field">
        <label>Business Date</label>
        <input :value="(decision.business_date ?? '').slice(0, 10)" disabled />
      </div>
      <div class="form-field">
        <label>Submitted At</label>
        <input
          :value="decision.submitted_at ? decision.submitted_at.replace('T', ' ').slice(0, 16) : '—'"
          disabled
        />
      </div>
    </div>

    <div class="form-row">
      <div class="form-field">
        <label>Instrument Code</label>
        <input :value="decision.instrument_code ?? '—'" disabled class="cell-mono" />
      </div>
      <div class="form-field">
        <label>Side</label>
        <span class="side-badge" :data-side="decision.side">
          {{ decision.side ?? "—" }}
        </span>
      </div>
    </div>

    <div class="form-row">
      <div class="form-field">
        <label>Quantity</label>
        <input :value="decision.quantity || '—'" disabled />
      </div>
      <div class="form-field">
        <label>Amount</label>
        <input :value="decision.amount || '—'" disabled />
      </div>
      <div class="form-field">
        <label>Limit Price</label>
        <input :value="decision.limit_price || '—'" disabled />
      </div>
    </div>

    <div class="form-row">
      <div class="form-field">
        <label>Currency</label>
        <input :value="decision.currency ?? '—'" disabled />
      </div>
      <div class="form-field form-field--double">
        <label>Exchange</label>
        <input :value="decision.exchange || '—'" disabled />
      </div>
    </div>

    <div class="form-row">
      <div class="form-field form-field--double">
        <label>Research Report No</label>
        <input :value="decision.research_report_no || '—'" disabled />
      </div>
    </div>

    <div class="form-row">
      <div class="form-field form-field--double">
        <label>Rationale</label>
        <textarea :value="decision.rationale || '—'" disabled rows="3" />
      </div>
    </div>

    <div v-if="decision.cancellation_reason" class="form-row">
      <div class="form-field form-field--double">
        <label>Cancellation Reason</label>
        <textarea :value="decision.cancellation_reason" disabled rows="2" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel-header {
  background: var(--bg-card-hover);
  border-bottom: 1px solid var(--border-subtle);
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.legacy-form {
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow-y: auto;
  flex: 1;
}

.form-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
}

.form-field--double {
  flex: 2;
}

.form-field label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
}

.form-field input,
.form-field textarea {
  padding: 5px 8px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card-muted);
  color: var(--text-primary);
  width: 100%;
  box-sizing: border-box;
  resize: none;
}

.form-field input:disabled,
.form-field textarea:disabled {
  opacity: 0.85;
  cursor: default;
}

.field-value {
  padding: 5px 0;
}

.highlighted-text {
  color: var(--state-danger, #cf222e) !important;
  font-weight: 700;
  font-family: monospace;
}

.cell-mono {
  font-family: monospace;
}

.side-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
  margin-top: 2px;
}

.side-badge[data-side="BUY"] {
  background: rgba(26, 127, 55, 0.15);
  color: var(--state-success, #1a7f37);
}

.side-badge[data-side="SELL"] {
  background: rgba(207, 34, 46, 0.15);
  color: var(--state-danger, #cf222e);
}
</style>
