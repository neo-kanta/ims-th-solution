<script setup lang="ts">
definePageMeta({
  layout: 'dashboard',
  middleware: ['auth', 'permission'],
  meta: { permission: 'AUDIT_VIEW' },
})

const activeTab    = ref('workers')
const selectedJob  = ref<string | null>(null)
const now          = '2026-03-13 11:08:42'

const statusMap: Record<string, { label: string; badge: string; dot: string }> = {
  RUNNING:  { label: 'Running',  badge: 'badge-success', dot: '#059669' },
  IDLE:     { label: 'Idle',     badge: 'badge-neutral', dot: '#94a3b8' },
  ERROR:    { label: 'Error',    badge: 'badge-error',   dot: '#dc2626' },
  DEGRADED: { label: 'Degraded', badge: 'badge-warning', dot: '#d97706' },
  STOPPED:  { label: 'Stopped',  badge: 'badge-neutral', dot: '#64748b' },
}

const workers = [
  {
    id: 'WRK-NAV',      name: 'NAV Refresh Worker',          status: 'RUNNING',
    schedule: '*/15 * * * *', lastRun: '2026-03-13 11:00:00', nextRun: '2026-03-13 11:15:00',
    lastDuration: '4.2s', avgDuration: '3.8s', successRate: '99.4%', runsToday: 44,
    desc: 'Fetches latest NAV data from market data providers and updates portfolio valuations for all active contracts.',
    recentRuns: [
      { ts: '2026-03-13 11:00:00', status: 'SUCCESS', duration: '4.2s', records: 5 },
      { ts: '2026-03-13 10:45:00', status: 'SUCCESS', duration: '3.9s', records: 5 },
      { ts: '2026-03-13 10:30:00', status: 'SUCCESS', duration: '3.7s', records: 5 },
      { ts: '2026-03-13 10:15:00', status: 'SUCCESS', duration: '4.1s', records: 5 },
      { ts: '2026-03-13 10:00:00', status: 'TIMEOUT', duration: '30s',  records: 0 },
    ],
  },
  {
    id: 'WRK-ALERT',    name: 'Watchlist Alert Worker',       status: 'RUNNING',
    schedule: '*/5 * * * *',  lastRun: '2026-03-13 11:05:00', nextRun: '2026-03-13 11:10:00',
    lastDuration: '0.8s', avgDuration: '0.7s', successRate: '100%', runsToday: 133,
    desc: 'Evaluates all active watchlist rules against latest market prices. Fires notifications when thresholds are crossed.',
    recentRuns: [
      { ts: '2026-03-13 11:05:00', status: 'SUCCESS', duration: '0.8s', records: 12 },
      { ts: '2026-03-13 11:00:00', status: 'SUCCESS', duration: '0.7s', records: 12 },
      { ts: '2026-03-13 10:55:00', status: 'SUCCESS', duration: '0.7s', records: 12 },
      { ts: '2026-03-13 10:50:00', status: 'SUCCESS', duration: '0.6s', records: 12 },
      { ts: '2026-03-13 10:45:00', status: 'SUCCESS', duration: '0.8s', records: 12 },
    ],
  },
  {
    id: 'WRK-EXPIRE',   name: 'Research Expiry Checker',      status: 'RUNNING',
    schedule: '0 8 * * *',    lastRun: '2026-03-13 08:00:00', nextRun: '2026-03-14 08:00:00',
    lastDuration: '1.1s', avgDuration: '1.0s', successRate: '100%', runsToday: 1,
    desc: 'Marks research reports as EXPIRED when past their validity date. Invalidates linked investment decisions.',
    recentRuns: [
      { ts: '2026-03-13 08:00:00', status: 'SUCCESS', duration: '1.1s', records: 0 },
      { ts: '2026-03-12 08:00:00', status: 'SUCCESS', duration: '0.9s', records: 1 },
      { ts: '2026-03-11 08:00:00', status: 'SUCCESS', duration: '1.0s', records: 0 },
    ],
  },
  {
    id: 'WRK-LEAVE',    name: 'Leave Delegation Activator',   status: 'RUNNING',
    schedule: '0 * * * *',    lastRun: '2026-03-13 11:00:00', nextRun: '2026-03-13 12:00:00',
    lastDuration: '0.3s', avgDuration: '0.3s', successRate: '100%', runsToday: 11,
    desc: 'Activates and deactivates permission delegations based on approved leave windows. Updates active agent chains.',
    recentRuns: [
      { ts: '2026-03-13 11:00:00', status: 'SUCCESS', duration: '0.3s', records: 2 },
      { ts: '2026-03-13 10:00:00', status: 'SUCCESS', duration: '0.3s', records: 1 },
      { ts: '2026-03-13 09:00:00', status: 'SUCCESS', duration: '0.2s', records: 1 },
    ],
  },
  {
    id: 'WRK-ETL-PAM',  name: 'PAM ETL Integration',          status: 'DEGRADED',
    schedule: '0 18 * * 1-5', lastRun: '2026-03-12 18:00:00', nextRun: '2026-03-13 18:00:00',
    lastDuration: '45s', avgDuration: '22s', successRate: '87.5%', runsToday: 0,
    desc: 'Exports approved transaction data to PAM (Portfolio Accounting & Management) system via SFTP. Phase 2 full integration pending.',
    recentRuns: [
      { ts: '2026-03-12 18:00:00', status: 'SUCCESS', duration: '45s',  records: 6 },
      { ts: '2026-03-11 18:00:00', status: 'FAILED',  duration: '30s',  records: 0 },
      { ts: '2026-03-10 18:00:00', status: 'SUCCESS', duration: '21s',  records: 4 },
    ],
  },
  {
    id: 'WRK-AUDIT',    name: 'Audit Log Archiver',            status: 'IDLE',
    schedule: '0 2 * * *',    lastRun: '2026-03-13 02:00:00', nextRun: '2026-03-14 02:00:00',
    lastDuration: '8.4s', avgDuration: '7.1s', successRate: '100%', runsToday: 1,
    desc: 'Archives audit log entries older than 90 days to cold storage. Maintains hot store at ≤ 10,000 recent records.',
    recentRuns: [
      { ts: '2026-03-13 02:00:00', status: 'SUCCESS', duration: '8.4s', records: 142 },
      { ts: '2026-03-12 02:00:00', status: 'SUCCESS', duration: '7.0s', records: 88 },
    ],
  },
]

const providers = [
  { id: 'PRV-SET',  name: 'SET Market Data',        type: 'REST/WebSocket', status: 'OPERATIONAL', latency: '12ms',  lastPing: '2026-03-13 11:08:30', endpoint: 'api.set.or.th/v1', phase: '1' },
  { id: 'PRV-BBG',  name: 'Bloomberg B-PIPE',        type: 'TCP/BLPAPI',    status: 'NOT_CONFIGURED', latency: '—', lastPing: '—',                   endpoint: 'localhost:8194',   phase: '2' },
  { id: 'PRV-SFTP', name: 'PAM SFTP Export',         type: 'SFTP',          status: 'DEGRADED',   latency: '—',     lastPing: '2026-03-12 18:00:00', endpoint: 'pam.abc-am.th:22', phase: '1' },
  { id: 'PRV-SMTP', name: 'SMTP Mail Relay',         type: 'SMTP/TLS',      status: 'OPERATIONAL', latency: '3ms',  lastPing: '2026-03-13 11:06:01', endpoint: 'smtp.ims-internal.th:587', phase: '1' },
  { id: 'PRV-DB',   name: 'PostgreSQL Primary',      type: 'pgx/pool',      status: 'OPERATIONAL', latency: '0.4ms',lastPing: '2026-03-13 11:08:42', endpoint: 'db:5432/ims_dev',  phase: '1' },
]

const selectedWorker = computed(() => workers.find(w => w.id === selectedJob))

const overallHealth = computed(() => {
  const errors = workers.filter(w => w.status === 'ERROR').length
  const degraded = workers.filter(w => w.status === 'DEGRADED').length
  if (errors > 0) return { label: 'DEGRADED', badge: 'badge-error' }
  if (degraded > 0) return { label: 'PARTIAL', badge: 'badge-warning' }
  return { label: 'HEALTHY', badge: 'badge-success' }
})

const runBadge = (s: string) => s === 'SUCCESS' ? 'badge-success' : s === 'FAILED' ? 'badge-error' : 'badge-warning'
const providerBadge = (s: string) => s === 'OPERATIONAL' ? 'badge-success' : s === 'DEGRADED' ? 'badge-warning' : 'badge-neutral'
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <div class="breadcrumb" style="margin-bottom:4px;"><span>System</span><span class="breadcrumb-sep">/</span><span class="breadcrumb-current">Jobs & Integration</span></div>
        <h1 class="page-title">Jobs & Integration Health</h1>
        <p class="page-desc">Scheduled worker status, run history, and external provider connectivity</p>
      </div>
      <div style="display:flex;gap:8px;align-items:center;">
        <span class="badge" :class="overallHealth.badge" style="font-size:11px;">System {{ overallHealth.label }}</span>
        <button class="btn btn-secondary btn-sm">Refresh</button>
      </div>
    </div>

    <div class="grid-4" style="margin-bottom:16px;">
      <div class="stat-card"><div class="stat-card-label">Active Workers</div><div class="stat-card-value" style="color:#059669;">{{ workers.filter(w=>w.status==='RUNNING').length }}</div><div class="stat-card-meta">of {{ workers.length }} total</div></div>
      <div class="stat-card"><div class="stat-card-label">Degraded</div><div class="stat-card-value" style="color:#d97706;">{{ workers.filter(w=>w.status==='DEGRADED').length }}</div><div class="stat-card-meta">Needs attention</div></div>
      <div class="stat-card"><div class="stat-card-label">Providers Online</div><div class="stat-card-value" style="color:#059669;">{{ providers.filter(p=>p.status==='OPERATIONAL').length }}</div><div class="stat-card-meta">of {{ providers.length }} configured</div></div>
      <div class="stat-card"><div class="stat-card-label">Last Checked</div><div class="stat-card-value" style="font-size:14px;color:#475569;">11:08</div><div class="stat-card-meta">{{ now }}</div></div>
    </div>

    <div class="card" style="margin-bottom:0;">
      <div class="card-header" style="padding-bottom:0;">
        <div class="tab-bar" style="border-bottom:none;flex:1;">
          <div class="tab-item" :class="activeTab==='workers'?'active':''" @click="activeTab='workers'">Workers <span class="tab-count">{{ workers.length }}</span></div>
          <div class="tab-item" :class="activeTab==='providers'?'active':''" @click="activeTab='providers'">Providers <span class="tab-count">{{ providers.length }}</span></div>
        </div>
      </div>

      <!-- WORKERS TAB -->
      <template v-if="activeTab==='workers'">
        <div style="display:grid;grid-template-columns:1fr 400px;gap:0;align-items:start;">
          <div>
            <table class="data-table" style="border-right:1px solid #f1f5f9;">
              <thead><tr>
                <th>Worker</th><th>Schedule</th><th>Last Run</th><th>Duration</th><th>Success Rate</th><th>Status</th><th class="col-action"></th>
              </tr></thead>
              <tbody>
                <tr v-for="w in workers" :key="w.id"
                  @click="selectedJob = selectedJob===w.id ? null : w.id"
                  style="cursor:pointer;"
                  :class="selectedJob===w.id?'selected':''"
                >
                  <td>
                    <div style="display:flex;align-items:center;gap:8px;">
                      <span style="width:7px;height:7px;border-radius:50%;flex-shrink:0;" :style="`background:${statusMap[w.status]?.dot};`"></span>
                      <div>
                        <div style="font-size:12.5px;font-weight:600;">{{ w.name }}</div>
                        <div class="col-mono" style="font-size:10.5px;color:#94a3b8;">{{ w.id }}</div>
                      </div>
                    </div>
                  </td>
                  <td class="col-mono" style="font-size:11.5px;color:#64748b;">{{ w.schedule }}</td>
                  <td class="col-mono" style="font-size:11.5px;color:#475569;">{{ w.lastRun.slice(11) }}</td>
                  <td style="font-size:12.5px;text-align:right;font-family:var(--font-family-mono);">{{ w.lastDuration }}</td>
                  <td style="font-size:12.5px;text-align:right;font-family:var(--font-family-mono);" :style="parseFloat(w.successRate)<95?'color:#dc2626;font-weight:600;':''">{{ w.successRate }}</td>
                  <td><span class="badge" :class="statusMap[w.status]?.badge" style="font-size:10px;">{{ statusMap[w.status]?.label }}</span></td>
                  <td class="col-action">
                    <button class="btn btn-ghost btn-icon-sm" title="Trigger now"><svg width="11" height="11" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5,3 13,8 5,13"/></svg></button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="selectedWorker" style="padding:16px;display:flex;flex-direction:column;gap:12px;">
            <div>
              <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px;">
                <span style="width:8px;height:8px;border-radius:50%;" :style="`background:${statusMap[selectedWorker.status]?.dot};`"></span>
                <span style="font-size:14px;font-weight:700;">{{ selectedWorker.name }}</span>
              </div>
              <div class="col-mono" style="font-size:11.5px;color:#94a3b8;margin-bottom:8px;">{{ selectedWorker.id }}</div>
              <div style="font-size:12.5px;color:#475569;line-height:1.6;">{{ selectedWorker.desc }}</div>
            </div>

            <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;">
              <div style="padding:8px 10px;background:#f8fafc;border:1px solid #e2e8f0;border-radius:5px;">
                <div class="text-label" style="margin-bottom:1px;">Schedule</div>
                <div class="col-mono" style="font-size:12px;">{{ selectedWorker.schedule }}</div>
              </div>
              <div style="padding:8px 10px;background:#f8fafc;border:1px solid #e2e8f0;border-radius:5px;">
                <div class="text-label" style="margin-bottom:1px;">Next Run</div>
                <div class="col-mono" style="font-size:12px;color:#2563eb;">{{ selectedWorker.nextRun.slice(11) }}</div>
              </div>
              <div style="padding:8px 10px;background:#f8fafc;border:1px solid #e2e8f0;border-radius:5px;">
                <div class="text-label" style="margin-bottom:1px;">Avg Duration</div>
                <div class="col-mono" style="font-size:12px;">{{ selectedWorker.avgDuration }}</div>
              </div>
              <div style="padding:8px 10px;background:#f8fafc;border:1px solid #e2e8f0;border-radius:5px;">
                <div class="text-label" style="margin-bottom:1px;">Runs Today</div>
                <div style="font-size:13px;font-weight:700;">{{ selectedWorker.runsToday }}</div>
              </div>
            </div>

            <div>
              <div style="font-size:12.5px;font-weight:700;margin-bottom:8px;color:#374151;">Recent Runs</div>
              <div style="display:flex;flex-direction:column;gap:4px;">
                <div v-for="run in selectedWorker.recentRuns" :key="run.ts"
                  style="display:grid;grid-template-columns:1fr 70px 40px 60px;gap:6px;align-items:center;padding:6px 10px;background:#f8fafc;border-radius:4px;border:1px solid #f1f5f9;"
                >
                  <div class="col-mono" style="font-size:11.5px;color:#475569;">{{ run.ts.slice(11) }}</div>
                  <span class="badge" :class="runBadge(run.status)" style="font-size:9px;justify-self:start;">{{ run.status }}</span>
                  <div class="col-mono" style="font-size:11.5px;text-align:right;">{{ run.duration }}</div>
                  <div style="font-size:11.5px;color:#64748b;text-align:right;">{{ run.records > 0 ? run.records+'rec' : '—' }}</div>
                </div>
              </div>
            </div>

            <div style="display:flex;gap:6px;">
              <button class="btn btn-secondary btn-sm" style="flex:1;">View Full Log</button>
              <button class="btn btn-primary btn-sm" style="flex:1;">
                <svg width="11" height="11" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:4px;"><polygon points="5,3 13,8 5,13"/></svg>
                Trigger Now
              </button>
            </div>
          </div>
          <div v-else style="padding:48px 24px;">
            <div class="empty-state">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#cbd5e1" stroke-width="1.5"><circle cx="12" cy="12" r="3"/><path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83"/></svg>
              <div class="empty-title">Select a worker</div>
              <div class="empty-desc">Click any worker row to view run history and controls</div>
            </div>
          </div>
        </div>
      </template>

      <!-- PROVIDERS TAB -->
      <template v-if="activeTab==='providers'">
        <div class="card-body">
          <div class="alert alert-info" style="margin-bottom:14px;">
            <svg width="13" height="13" viewBox="0 0 16 16" fill="currentColor" style="flex-shrink:0;"><circle cx="8" cy="8" r="6"/></svg>
            <div style="font-size:12px;">Provider health is checked every 30 seconds. Latency is measured as round-trip ping to the endpoint. Phase 2 providers are placeholders — not yet integrated.</div>
          </div>
          <div style="display:flex;flex-direction:column;gap:8px;">
            <div v-for="p in providers" :key="p.id"
              style="display:grid;grid-template-columns:100px 200px 120px 1fr 80px 160px 60px;gap:12px;align-items:center;padding:12px 14px;border:1px solid #e2e8f0;border-radius:6px;"
              :style="p.status==='DEGRADED'?'background:#fffbeb;border-color:#fde68a;':p.status==='NOT_CONFIGURED'?'background:#f8fafc;':'background:#f0fdf4;border-color:#d1fae5;'"
            >
              <span class="col-mono" style="font-size:11px;font-weight:700;color:#475569;">{{ p.id }}</span>
              <div>
                <div style="font-size:12.5px;font-weight:600;">{{ p.name }}</div>
                <div style="font-size:11px;color:#94a3b8;">{{ p.type }}</div>
              </div>
              <div style="display:flex;align-items:center;gap:6px;">
                <span style="width:7px;height:7px;border-radius:50%;" :style="p.status==='OPERATIONAL'?'background:#059669;':p.status==='DEGRADED'?'background:#d97706;':'background:#94a3b8;'"></span>
                <span class="badge" :class="providerBadge(p.status)" style="font-size:10px;">{{ p.status }}</span>
              </div>
              <div class="col-mono" style="font-size:11.5px;color:#64748b;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ p.endpoint }}</div>
              <div style="text-align:right;">
                <div class="text-label" style="margin-bottom:1px;">Latency</div>
                <div class="col-mono" style="font-size:12px;font-weight:600;" :style="p.status==='OPERATIONAL'?'color:#059669;':''">{{ p.latency }}</div>
              </div>
              <div>
                <div class="text-label" style="margin-bottom:1px;">Last Ping</div>
                <div class="col-mono" style="font-size:11px;color:#94a3b8;">{{ p.lastPing !== '—' ? p.lastPing.slice(11) : '—' }}</div>
              </div>
              <div style="text-align:center;">
                <span class="badge" :class="p.phase==='2'?'badge-neutral':'badge-info'" style="font-size:9px;">Ph {{ p.phase }}</span>
              </div>
            </div>
          </div>

          <div style="margin-top:16px;padding:14px;background:#f8fafc;border:1px solid #e2e8f0;border-radius:6px;">
            <div style="font-size:12.5px;font-weight:700;margin-bottom:6px;color:#374151;">PAM SFTP — Degraded Detail</div>
            <div class="alert alert-warning" style="margin-bottom:0;">
              <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8" style="flex-shrink:0;"><path d="M8 1l7 13H1L8 1z"/><path d="M8 6v4M8 11v1"/></svg>
              <div style="font-size:12px;">Last export (2026-03-11) failed with SFTP connection timeout. Retry succeeded next business day. Root cause: network firewall rule change on PAM server side. Manual verification recommended before next scheduled export (2026-03-13 18:00).</div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
